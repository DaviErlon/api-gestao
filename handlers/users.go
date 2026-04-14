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

func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/users")
	id = strings.Trim(id, "/")

	switch {
	case r.Method == http.MethodGet && id == "":
		listUsers(w, r)
	case r.Method == http.MethodGet && id != "":
		getUser(w, r, id)
	case r.Method == http.MethodPut && id != "":
		updateOwnUser(w, r, id)
	default:
		http.Error(w, "método não permitido", http.StatusMethodNotAllowed)
	}
}

// AdminProfileHandler - rotas exclusivas para admin
// GET    /users       - lista todos
// GET    /users/{id}  - busca um
// POST   /users       - cria usuário
// PUT    /users/{id}  - atualiza qualquer usuário (todos os campos)
// DELETE /users/{id}  - remove qualquer usuário
func AdminProfileHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/users")
	id = strings.Trim(id, "/")

	switch {
	case r.Method == http.MethodGet && id == "":
		listUsers(w, r)
	case r.Method == http.MethodGet && id != "":
		getUser(w, r, id)
	case r.Method == http.MethodPost:
		createUser(w, r)
	case r.Method == http.MethodPut && id != "":
		updateUser(w, r, id)
	case r.Method == http.MethodDelete && id != "":
		deleteUser(w, r, id)
	default:
		http.Error(w, "método não permitido", http.StatusMethodNotAllowed)
	}
}

// updateOwnUser - usuário só pode alterar o próprio registro.
// Campos permitidos: name, password. login e empresa_id são bloqueados.
func updateOwnUser(w http.ResponseWriter, r *http.Request, rawID string) {
	targetID, err := strconv.Atoi(rawID)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	callerID, ok := r.Context().Value(auth.UserIDKey).(int)
	if !ok {
		http.Error(w, "Não autenticado", http.StatusUnauthorized)
		return
	}

	if callerID != targetID {
		http.Error(w, "Acesso negado: você só pode alterar o próprio perfil", http.StatusForbidden)
		return
	}

	var body struct {
		Name     string `json:"name"`
		Password string `json:"password"`
		Login    string `json:"login"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "corpo inválido", http.StatusBadRequest)
		return
	}

	res, err := repository.DB.Exec(
		`UPDATE users SET name=$1, password=$2, login=$3 WHERE id=$4`,
		body.Name, body.Password, body.Login, targetID,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		http.Error(w, "usuário não encontrado", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "perfil atualizado"})
}

func listUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := repository.DB.Query(
		`SELECT id, name, login, password, empresa_id FROM users`,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var list []entities.User
	for rows.Next() {
		var u entities.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Login, &u.Password, &u.EmpresaID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		list = append(list, u)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func getUser(w http.ResponseWriter, r *http.Request, rawID string) {
	id, err := strconv.Atoi(rawID)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}
	var u entities.User
	err = repository.DB.QueryRow(
		`SELECT id, name, login, password, empresa_id FROM users WHERE id=$1`, id,
	).Scan(&u.ID, &u.Name, &u.Login, &u.Password, &u.EmpresaID)
	if err == sql.ErrNoRows {
		http.Error(w, "usuário não encontrado", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(u)
}

func createUser(w http.ResponseWriter, r *http.Request) {
	var u entities.User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, "corpo inválido", http.StatusBadRequest)
		return
	}
	err := repository.DB.QueryRow(
		`INSERT INTO users (name, login, password, empresa_id) VALUES ($1,$2,$3,$4) RETURNING id`,
		u.Name, u.Login, u.Password, u.EmpresaID,
	).Scan(&u.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(u)
}

// updateUser — versão admin: pode alterar qualquer campo de qualquer usuário
func updateUser(w http.ResponseWriter, r *http.Request, rawID string) {
	id, err := strconv.Atoi(rawID)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}
	var u entities.User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, "corpo inválido", http.StatusBadRequest)
		return
	}
	u.ID = id
	res, err := repository.DB.Exec(
		`UPDATE users SET name=$1, login=$2, password=$3, empresa_id=$4 WHERE id=$5`,
		u.Name, u.Login, u.Password, u.EmpresaID, u.ID,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		http.Error(w, "usuário não encontrado", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(u)
}

func deleteUser(w http.ResponseWriter, r *http.Request, rawID string) {
	id, err := strconv.Atoi(rawID)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}
	res, err := repository.DB.Exec(`DELETE FROM users WHERE id=$1`, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		http.Error(w, "usuário não encontrado", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
