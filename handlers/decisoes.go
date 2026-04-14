package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/DaviErlon/api-gestao/auth"
	"github.com/DaviErlon/api-gestao/entities"
	"github.com/DaviErlon/api-gestao/repository"
)

// DecisoesHandler — uma decisão por ciclo por empresa
// GET    /decisoes       → lista todas as decisões da empresa do usuário logado
// GET    /decisoes/{id}  → busca uma decisão da empresa do usuário logado
// POST   /decisoes       → cria (empresa_id vem do contexto; rejeita se já existe para o ciclo)
// PUT    /decisoes/{id}  → atualiza a decisão da própria empresa
// DELETE /decisoes/{id}  → deleta a decisão da própria empresa
func DecisoesHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/decisoes")
	id = strings.Trim(id, "/")

	switch {
	case r.Method == http.MethodGet && id == "":
		listDecisoes(w, r)
	case r.Method == http.MethodGet && id != "":
		getDecisao(w, r, id)
	case r.Method == http.MethodPost:
		createDecisao(w, r)
	case r.Method == http.MethodPut && id != "":
		updateDecisao(w, r, id)
	case r.Method == http.MethodDelete && id != "":
		deleteDecisao(w, r, id)
	default:
		http.Error(w, "método não permitido", http.StatusMethodNotAllowed)
	}
}

// empresaIDFromContext busca o empresa_id do usuário logado no banco
func empresaIDFromContext(r *http.Request) (int, error) {
	userID, ok := r.Context().Value(auth.UserIDKey).(int)
	if !ok {
		return 0, sql.ErrNoRows
	}
	var empresaID int
	err := repository.DB.QueryRow(`SELECT empresa_id FROM users WHERE id=$1`, userID).Scan(&empresaID)
	return empresaID, err
}

func listDecisoes(w http.ResponseWriter, r *http.Request) {
	empresaID, err := empresaIDFromContext(r)
	if err != nil {
		http.Error(w, "Não autenticado", http.StatusUnauthorized)
		return
	}

	rows, err := repository.DB.Query(
		`SELECT id, marketing, ped, suporte, seguranca, expansao, empresa_id, ciclo_id
		 FROM decisoes WHERE empresa_id=$1`,
		empresaID,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var list []entities.Decisao
	for rows.Next() {
		var d entities.Decisao
		if err := rows.Scan(&d.ID, &d.Marketing, &d.PeD, &d.Suporte, &d.Seguranca, &d.Expansao, &d.EmpresaID, &d.CicloID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		list = append(list, d)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func getDecisao(w http.ResponseWriter, r *http.Request, rawID string) {
	id, err := strconv.Atoi(rawID)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	empresaID, err := empresaIDFromContext(r)
	if err != nil {
		http.Error(w, "Não autenticado", http.StatusUnauthorized)
		return
	}

	var d entities.Decisao
	err = repository.DB.QueryRow(
		`SELECT id, marketing, ped, suporte, seguranca, expansao, empresa_id, ciclo_id
		 FROM decisoes WHERE id=$1 AND empresa_id=$2`, id, empresaID,
	).Scan(&d.ID, &d.Marketing, &d.PeD, &d.Suporte, &d.Seguranca, &d.Expansao, &d.EmpresaID, &d.CicloID)
	if err == sql.ErrNoRows {
		http.Error(w, "decisão não encontrada", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(d)
}

// createDecisao — empresa_id vem do contexto; rejeita se já existe decisão para esse ciclo+empresa
func createDecisao(w http.ResponseWriter, r *http.Request) {
	empresaID, err := empresaIDFromContext(r)
	if err != nil {
		http.Error(w, "Não autenticado", http.StatusUnauthorized)
		return
	}

	var d entities.Decisao
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		http.Error(w, "corpo inválido", http.StatusBadRequest)
		return
	}
	d.EmpresaID = empresaID

	// Verifica unicidade antes de inserir para retornar erro claro
	var exists bool
	err = repository.DB.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM decisoes WHERE empresa_id=$1 AND ciclo_id=$2)`,
		d.EmpresaID, d.CicloID,
	).Scan(&exists)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if exists {
		http.Error(w, "já existe uma decisão para esse ciclo", http.StatusConflict)
		return
	}

	err = repository.DB.QueryRow(
		`INSERT INTO decisoes (marketing, ped, suporte, seguranca, expansao, empresa_id, ciclo_id)
		 VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id`,
		d.Marketing, d.PeD, d.Suporte, d.Seguranca, d.Expansao, d.EmpresaID, d.CicloID,
	).Scan(&d.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(d)
}

// updateDecisao — só atualiza se a decisão pertence à empresa do usuário logado
func updateDecisao(w http.ResponseWriter, r *http.Request, rawID string) {
	id, err := strconv.Atoi(rawID)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	empresaID, err := empresaIDFromContext(r)
	if err != nil {
		http.Error(w, "Não autenticado", http.StatusUnauthorized)
		return
	}

	var d entities.Decisao
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		http.Error(w, "corpo inválido", http.StatusBadRequest)
		return
	}
	d.ID = id
	d.EmpresaID = empresaID

	res, err := repository.DB.Exec(
		`UPDATE decisoes SET marketing=$1, ped=$2, suporte=$3, seguranca=$4, expansao=$5
		 WHERE id=$6 AND empresa_id=$7`,
		d.Marketing, d.PeD, d.Suporte, d.Seguranca, d.Expansao, d.ID, d.EmpresaID,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		http.Error(w, "decisão não encontrada", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(d)
}

// deleteDecisao — só deleta se a decisão pertence à empresa do usuário logado
func deleteDecisao(w http.ResponseWriter, r *http.Request, rawID string) {
	id, err := strconv.Atoi(rawID)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	empresaID, err := empresaIDFromContext(r)
	if err != nil {
		http.Error(w, "Não autenticado", http.StatusUnauthorized)
		return
	}

	res, err := repository.DB.Exec(
		`DELETE FROM decisoes WHERE id=$1 AND empresa_id=$2`, id, empresaID,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		http.Error(w, "decisão não encontrada", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}