package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/DaviErlon/api-gestao/entities"
	"github.com/DaviErlon/api-gestao/repository"
)

func EmpresasHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/empresas")
	id = strings.Trim(id, "/")

	switch {
	case r.Method == http.MethodGet && id == "":
		listEmpresas(w, r)
	case r.Method == http.MethodGet && id != "":
		getEmpresa(w, r, id)
	default:
		http.Error(w, "método não permitido", http.StatusMethodNotAllowed)
	}
}

func AdminEmpresasHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/empresas")
	id = strings.Trim(id, "/")

	switch {
	case r.Method == http.MethodGet && id == "":
		listEmpresas(w, r)
	case r.Method == http.MethodGet && id != "":
		getEmpresa(w, r, id)
	case r.Method == http.MethodPost:
		createEmpresa(w, r)
	case r.Method == http.MethodPut && id != "":
		updateEmpresa(w, r, id)
	case r.Method == http.MethodDelete && id != "":
		deleteEmpresa(w, r, id)
	default:
		http.Error(w, "método não permitido", http.StatusMethodNotAllowed)
	}
}

func listEmpresas(w http.ResponseWriter, r *http.Request) {
	rows, err := repository.DB.Query(
		`SELECT id, name FROM empresas`,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	list := []entities.Empresa{}
	for rows.Next() {
		var e entities.Empresa
		if err := rows.Scan(&e.ID, &e.Name); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		list = append(list, e)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func getEmpresa(w http.ResponseWriter, r *http.Request, rawID string) {
	id, err := strconv.Atoi(rawID)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}
	var e entities.Empresa
	err = repository.DB.QueryRow(
		`SELECT id, name FROM empresas WHERE id=$1`, id,
	).Scan(&e.ID, &e.Name)
	if err == sql.ErrNoRows {
		http.Error(w, "empresa não encontrada", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(e)
}

func createEmpresa(w http.ResponseWriter, r *http.Request) {
	var e entities.Empresa
	if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
		http.Error(w, "corpo inválido", http.StatusBadRequest)
		return
	}
	err := repository.DB.QueryRow(
		`INSERT INTO empresas (name) VALUES ($1) RETURNING id`,
		e.Name,
	).Scan(&e.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(e)
}

func updateEmpresa(w http.ResponseWriter, r *http.Request, rawID string) {
	id, err := strconv.Atoi(rawID)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}
	var e entities.Empresa
	if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
		http.Error(w, "corpo inválido", http.StatusBadRequest)
		return
	}
	e.ID = id
	res, err := repository.DB.Exec(
		`UPDATE empresas SET name=$1 WHERE id=$2`,
		e.Name, e.ID,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		http.Error(w, "empresa não encontrada", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(e)
}

func deleteEmpresa(w http.ResponseWriter, r *http.Request, rawID string) {
	id, err := strconv.Atoi(rawID)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}
	res, err := repository.DB.Exec(`DELETE FROM empresas WHERE id=$1`, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		http.Error(w, "empresa não encontrada", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}