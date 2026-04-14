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

// ReunioesHandler — rotas para usuário autenticado comum
// GET  /reunioes       → lista todas
// GET  /reunioes/{id}  → busca uma
// POST /reunioes       → cria (author_id preenchido pelo contexto)
// PUT  /reunioes/{id}  → atualiza apenas a própria reunião
// DELETE /reunioes/{id}→ deleta apenas a própria reunião
func ReunioesHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/reunioes")
	id = strings.Trim(id, "/")

	switch {
	case r.Method == http.MethodGet && id == "":
		listReunioes(w, r)
	case r.Method == http.MethodGet && id != "":
		getReuniao(w, r, id)
	case r.Method == http.MethodPost:
		createReuniao(w, r)
	case r.Method == http.MethodPut && id != "":
		updateOwnReuniao(w, r, id)
	case r.Method == http.MethodDelete && id != "":
		deleteOwnReuniao(w, r, id)
	default:
		http.Error(w, "método não permitido", http.StatusMethodNotAllowed)
	}
}

// AdminReunioesHandler — rotas exclusivas para admin
// GET    /reunioes       → lista todas
// GET    /reunioes/{id}  → busca uma
// PUT    /reunioes/{id}  → atualiza qualquer reunião
// DELETE /reunioes/{id}  → deleta qualquer reunião
func AdminReunioesHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/reunioes")
	id = strings.Trim(id, "/")

	switch {
	case r.Method == http.MethodGet && id == "":
		listReunioes(w, r)
	case r.Method == http.MethodGet && id != "":
		getReuniao(w, r, id)
	case r.Method == http.MethodPut && id != "":
		updateReuniao(w, r, id)
	case r.Method == http.MethodDelete && id != "":
		deleteReuniao(w, r, id)
	default:
		http.Error(w, "método não permitido", http.StatusMethodNotAllowed)
	}
}

func listReunioes(w http.ResponseWriter, r *http.Request) {
	rows, err := repository.DB.Query(
		`SELECT id, ciclo_id, author_id, titulo, descricao, inicio, duracao, aberta FROM reunioes`,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var list []entities.Reuniao
	for rows.Next() {
		var re entities.Reuniao
		if err := rows.Scan(&re.ID, &re.CicloID, &re.AuthorID, &re.Titulo, &re.Descricao, &re.Inicio, &re.Duracao, &re.Aberta); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		list = append(list, re)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func getReuniao(w http.ResponseWriter, r *http.Request, rawID string) {
	id, err := strconv.Atoi(rawID)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}
	var re entities.Reuniao
	err = repository.DB.QueryRow(
		`SELECT id, ciclo_id, author_id, titulo, descricao, inicio, duracao, aberta FROM reunioes WHERE id=$1`, id,
	).Scan(&re.ID, &re.CicloID, &re.AuthorID, &re.Titulo, &re.Descricao, &re.Inicio, &re.Duracao, &re.Aberta)
	if err == sql.ErrNoRows {
		http.Error(w, "reunião não encontrada", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(re)
}

// createReuniao — author_id é sempre o usuário logado, ignorando qualquer valor enviado no body
func createReuniao(w http.ResponseWriter, r *http.Request) {
	callerID, ok := r.Context().Value(auth.UserIDKey).(int)
	if !ok {
		http.Error(w, "Não autenticado", http.StatusUnauthorized)
		return
	}

	var re entities.Reuniao
	if err := json.NewDecoder(r.Body).Decode(&re); err != nil {
		http.Error(w, "corpo inválido", http.StatusBadRequest)
		return
	}
	re.AuthorID = callerID

	err := repository.DB.QueryRow(
		`INSERT INTO reunioes (ciclo_id, author_id, titulo, descricao, inicio, duracao, aberta)
		 VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id`,
		re.CicloID, re.AuthorID, re.Titulo, re.Descricao, re.Inicio, re.Duracao, re.Aberta,
	).Scan(&re.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(re)
}

// updateOwnReuniao — usuário só pode alterar reuniões que ele criou
func updateOwnReuniao(w http.ResponseWriter, r *http.Request, rawID string) {
	reuniaoID, err := strconv.Atoi(rawID)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	callerID, ok := r.Context().Value(auth.UserIDKey).(int)
	if !ok {
		http.Error(w, "Não autenticado", http.StatusUnauthorized)
		return
	}

	var authorID int
	err = repository.DB.QueryRow(`SELECT author_id FROM reunioes WHERE id=$1`, reuniaoID).Scan(&authorID)
	if err == sql.ErrNoRows {
		http.Error(w, "reunião não encontrada", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if callerID != authorID {
		http.Error(w, "Acesso negado: você só pode alterar reuniões que criou", http.StatusForbidden)
		return
	}

	updateReuniao(w, r, rawID)
}

// deleteOwnReuniao — usuário só pode deletar reuniões que ele criou
func deleteOwnReuniao(w http.ResponseWriter, r *http.Request, rawID string) {
	reuniaoID, err := strconv.Atoi(rawID)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	callerID, ok := r.Context().Value(auth.UserIDKey).(int)
	if !ok {
		http.Error(w, "Não autenticado", http.StatusUnauthorized)
		return
	}

	var authorID int
	err = repository.DB.QueryRow(`SELECT author_id FROM reunioes WHERE id=$1`, reuniaoID).Scan(&authorID)
	if err == sql.ErrNoRows {
		http.Error(w, "reunião não encontrada", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if callerID != authorID {
		http.Error(w, "Acesso negado: você só pode deletar reuniões que criou", http.StatusForbidden)
		return
	}

	deleteReuniao(w, r, rawID)
}

// updateReuniao — versão admin: atualiza qualquer reunião sem verificar authorship
func updateReuniao(w http.ResponseWriter, r *http.Request, rawID string) {
	id, err := strconv.Atoi(rawID)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}
	var re entities.Reuniao
	if err := json.NewDecoder(r.Body).Decode(&re); err != nil {
		http.Error(w, "corpo inválido", http.StatusBadRequest)
		return
	}
	re.ID = id
	res, err := repository.DB.Exec(
		`UPDATE reunioes SET ciclo_id=$1, titulo=$2, descricao=$3, inicio=$4, duracao=$5, aberta=$6 WHERE id=$7`,
		re.CicloID, re.Titulo, re.Descricao, re.Inicio, re.Duracao, re.Aberta, re.ID,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		http.Error(w, "reunião não encontrada", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(re)
}

// deleteReuniao — versão admin: deleta qualquer reunião sem verificar authorship
func deleteReuniao(w http.ResponseWriter, r *http.Request, rawID string) {
	id, err := strconv.Atoi(rawID)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}
	res, err := repository.DB.Exec(`DELETE FROM reunioes WHERE id=$1`, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		http.Error(w, "reunião não encontrada", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}