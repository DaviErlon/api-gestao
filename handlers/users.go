package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/DaviErlon/api-gestao/entities"
	"github.com/DaviErlon/api-gestao/repository"
	"golang.org/x/crypto/bcrypt"
)

func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/users")
	id = strings.Trim(id, "/")

	switch {
	case r.Method == http.MethodGet && id == "":
		listUsers(w, r)
	case r.Method == http.MethodGet && id != "":
		getUser(w, r, id)
	case r.Method == http.MethodPut:
		updateOwnUser(w, r)
	default:
		http.Error(w, "método não permitido", http.StatusMethodNotAllowed)
	}
}

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

func updateOwnUser(w http.ResponseWriter, r *http.Request) {
	callerID, ok := callerID(r)
	if !ok {
		http.Error(w, "Não autorizado", http.StatusForbidden)
		return
	}

	var body struct {
		Name      *string `json:"name"`
		Password  *string `json:"password"`
		Login     *string `json:"login"`
		CurrSenha string  `json:"curr_password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "corpo inválido", http.StatusBadRequest)
		return
	}

	// 🔒 pegar hash atual
	var hash string
	err := repository.DB.QueryRow(
		`SELECT password FROM users WHERE id=$1`,
		callerID,
	).Scan(&hash)

	if err != nil {
		http.Error(w, "usuário não encontrado", http.StatusNotFound)
		return
	}

	// 🔐 validar senha
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(body.CurrSenha)) != nil {
		http.Error(w, "senha atual inválida", http.StatusUnauthorized)
		return
	}

	// 🧠 montar campos dinamicamente (mais limpo)
	fields := []string{}
	args := []any{}
	i := 1

	add := func(field string, value any) {
		fields = append(fields, field+"=$"+strconv.Itoa(i))
		args = append(args, value)
		i++
	}

	if body.Name != nil {
		add("name", *body.Name)
	}

	if body.Login != nil {
		add("login", *body.Login)
	}

	if body.Password != nil {
		hashed, err := bcrypt.GenerateFromPassword([]byte(*body.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "erro ao gerar hash", http.StatusInternalServerError)
			return
		}
		add("password", string(hashed))
	}

	if len(fields) == 0 {
		http.Error(w, "nenhum campo para atualizar", http.StatusBadRequest)
		return
	}

	query := "UPDATE users SET " + strings.Join(fields, ", ") + " WHERE id=$" + strconv.Itoa(i)
	args = append(args, callerID)

	_, err = repository.DB.Exec(query, args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"message": "perfil atualizado",
	})
}

func listUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := repository.DB.Query(
		`SELECT id, name, login, COALESCE(empresa_id, -1) FROM users`,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	list := []entities.User{}
	for rows.Next() {
		var u entities.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Login, &u.EmpresaID); err != nil {
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
