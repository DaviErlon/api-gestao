package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/DaviErlon/api-gestao/auth"
	"github.com/DaviErlon/api-gestao/entities"
	"github.com/DaviErlon/api-gestao/repository"
)

// Usuário comum
func DecisoesHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/decisoes")
	id = strings.Trim(id, "/")

	switch {
	case r.Method == http.MethodGet && id == "":
		listDecisoesOwn(w, r)
	case r.Method == http.MethodGet && id != "":
		getDecisaoOwn(w, r, id)
	case r.Method == http.MethodPost:
		createDecisaoOwn(w, r)
	case r.Method == http.MethodPut && id != "":
		updateDecisaoOwn(w, r, id)
	case r.Method == http.MethodDelete && id != "":
		deleteDecisaoOwn(w, r, id)
	default:
		http.Error(w, "método não permitido", http.StatusMethodNotAllowed)
	}
}

// Admin
func AdminDecisoesHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/decisoes")
	id = strings.Trim(id, "/")

	switch {
	case r.Method == http.MethodGet && id == "":
		listDecisoes(w, r)
	case r.Method == http.MethodGet && id != "":
		getDecisao(w, r, id)
	case r.Method == http.MethodPut && id != "":
		updateDecisao(w, r, id)
	case r.Method == http.MethodDelete && id != "":
		deleteDecisao(w, r, id)
	default:
		http.Error(w, "método não permitido", http.StatusMethodNotAllowed)
	}
}

// ---------------------------------------------------------------------------
// Helper
// ---------------------------------------------------------------------------

func empresaIDFromContext(r *http.Request) (int, error) {
	userID, ok := r.Context().Value(auth.UserIDKey).(int)
	if !ok {
		return 0, fmt.Errorf("usuário não autenticado")
	}

	var empresaID int
	err := repository.DB.QueryRow(
		`SELECT empresa_id FROM users WHERE id=$1`,
		userID,
	).Scan(&empresaID)

	if err != nil {
		return 0, err
	}

	return empresaID, nil
}

// ---------------------------------------------------------------------------
// com filtro por empresa
// ---------------------------------------------------------------------------

func listDecisoesOwn(w http.ResponseWriter, r *http.Request) {
	empresaID, err := empresaIDFromContext(r)
	if err != nil {
		http.Error(w, "Não autenticado", http.StatusUnauthorized)
		return
	}

	rows, err := repository.DB.Query(
		`SELECT id, marketing, ped, suporte, seguranca, expansao, empresa_id, ciclo_id
	 	FROM decisoes
	 	WHERE empresa_id=$1
	 	ORDER BY ciclo_id ASC`,
		empresaID,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	list := []entities.Decisao{}
	for rows.Next() {
		var d entities.Decisao
		rows.Scan(&d.ID, &d.Marketing, &d.PeD, &d.Suporte, &d.Seguranca, &d.Expansao, &d.EmpresaID, &d.CicloID)
		list = append(list, d)
	}

	json.NewEncoder(w).Encode(list)
}

func getDecisaoOwn(w http.ResponseWriter, r *http.Request, rawID string) {
	id, _ := strconv.Atoi(rawID)

	empresaID, err := empresaIDFromContext(r)
	if err != nil {
		http.Error(w, "Não autenticado", http.StatusUnauthorized)
		return
	}

	var d entities.Decisao
	err = repository.DB.QueryRow(
		`SELECT id, marketing, ped, suporte, seguranca, expansao, empresa_id, ciclo_id
		 FROM decisoes WHERE id=$1 AND empresa_id=$2`,
		id, empresaID,
	).Scan(&d.ID, &d.Marketing, &d.PeD, &d.Suporte, &d.Seguranca, &d.Expansao, &d.EmpresaID, &d.CicloID)

	if err == sql.ErrNoRows {
		http.Error(w, "decisão não encontrada", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(d)
}

func createDecisaoOwn(w http.ResponseWriter, r *http.Request) {
	empresaID, err := empresaIDFromContext(r)
	if err != nil {
		http.Error(w, "Não autenticado", http.StatusUnauthorized)
		return
	}

	var d entities.Decisao
	json.NewDecoder(r.Body).Decode(&d)

	d.EmpresaID = empresaID

	var exists bool
	repository.DB.QueryRow(
		`SELECT EXISTS(
			SELECT 1 FROM decisoes WHERE empresa_id=$1 AND ciclo_id=$2
		)`,
		d.EmpresaID, d.CicloID,
	).Scan(&exists)

	if exists {
		http.Error(w, "já existe decisão para esse ciclo", http.StatusConflict)
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

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(d)
}

func updateDecisaoOwn(w http.ResponseWriter, r *http.Request, rawID string) {
	id, _ := strconv.Atoi(rawID)

	empresaID, err := empresaIDFromContext(r)
	if err != nil {
		http.Error(w, "Não autenticado", http.StatusUnauthorized)
		return
	}

	var d entities.Decisao
	json.NewDecoder(r.Body).Decode(&d)

	res, _ := repository.DB.Exec(
		`UPDATE decisoes SET marketing=$1, ped=$2, suporte=$3, seguranca=$4, expansao=$5
		 WHERE id=$6 AND empresa_id=$7`,
		d.Marketing, d.PeD, d.Suporte, d.Seguranca, d.Expansao, id, empresaID,
	)

	if n, _ := res.RowsAffected(); n == 0 {
		http.Error(w, "decisão não encontrada", http.StatusNotFound)
		return
	}

	d.ID = id
	d.EmpresaID = empresaID

	json.NewEncoder(w).Encode(d)
}

func deleteDecisaoOwn(w http.ResponseWriter, r *http.Request, rawID string) {
	id, _ := strconv.Atoi(rawID)

	empresaID, err := empresaIDFromContext(r)
	if err != nil {
		http.Error(w, "Não autenticado", http.StatusUnauthorized)
		return
	}

	res, _ := repository.DB.Exec(
		`DELETE FROM decisoes WHERE id=$1 AND empresa_id=$2`,
		id, empresaID,
	)

	if n, _ := res.RowsAffected(); n == 0 {
		http.Error(w, "decisão não encontrada", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ---------------------------------------------------------------------------
// sem filtro
// ---------------------------------------------------------------------------

func listDecisoes(w http.ResponseWriter, r *http.Request) {
	rows, _ := repository.DB.Query(`
		SELECT id, marketing, ped, suporte, seguranca, expansao, empresa_id, ciclo_id
		FROM decisoes
		ORDER BY ciclo_id ASC`)
	defer rows.Close()

	list := []entities.Decisao{}
	for rows.Next() {
		var d entities.Decisao
		rows.Scan(&d.ID, &d.Marketing, &d.PeD, &d.Suporte, &d.Seguranca, &d.Expansao, &d.EmpresaID, &d.CicloID)
		list = append(list, d)
	}

	json.NewEncoder(w).Encode(list)
}

func getDecisao(w http.ResponseWriter, r *http.Request, rawID string) {
	id, _ := strconv.Atoi(rawID)

	var d entities.Decisao
	err := repository.DB.QueryRow(
		`SELECT id, marketing, ped, suporte, seguranca, expansao, empresa_id, ciclo_id
		 FROM decisoes WHERE id=$1`,
		id,
	).Scan(&d.ID, &d.Marketing, &d.PeD, &d.Suporte, &d.Seguranca, &d.Expansao, &d.EmpresaID, &d.CicloID)

	if err == sql.ErrNoRows {
		http.Error(w, "decisão não encontrada", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(d)
}

func updateDecisao(w http.ResponseWriter, r *http.Request, rawID string) {
	id, _ := strconv.Atoi(rawID)

	var d entities.Decisao
	json.NewDecoder(r.Body).Decode(&d)

	res, _ := repository.DB.Exec(
		`UPDATE decisoes SET marketing=$1, ped=$2, suporte=$3, seguranca=$4, expansao=$5,
		 empresa_id=$6, ciclo_id=$7
		 WHERE id=$8`,
		d.Marketing, d.PeD, d.Suporte, d.Seguranca, d.Expansao, d.EmpresaID, d.CicloID, id,
	)

	if n, _ := res.RowsAffected(); n == 0 {
		http.Error(w, "decisão não encontrada", http.StatusNotFound)
		return
	}

	d.ID = id
	json.NewEncoder(w).Encode(d)
}

func deleteDecisao(w http.ResponseWriter, r *http.Request, rawID string) {
	id, _ := strconv.Atoi(rawID)

	res, _ := repository.DB.Exec(`DELETE FROM decisoes WHERE id=$1`, id)

	if n, _ := res.RowsAffected(); n == 0 {
		http.Error(w, "decisão não encontrada", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
