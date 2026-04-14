package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/DaviErlon/api-gestao/auth"
	"github.com/DaviErlon/api-gestao/entities"
	"github.com/DaviErlon/api-gestao/repository"
)

// PostsHandler — rotas para usuário autenticado comum
// GET    /posts       → lista todos
// GET    /posts/{id}  → busca um
// POST   /posts       → cria (author_id e posted_at preenchidos pelo servidor)
// PUT    /posts/{id}  → atualiza apenas o próprio post
// DELETE /posts/{id}  → deleta apenas o próprio post
func PostsHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/posts")
	id = strings.Trim(id, "/")

	switch {
	case r.Method == http.MethodGet && id == "":
		listPosts(w, r)
	case r.Method == http.MethodGet && id != "":
		getPost(w, r, id)
	case r.Method == http.MethodPost:
		createPost(w, r)
	case r.Method == http.MethodPut && id != "":
		updateOwnPost(w, r, id)
	case r.Method == http.MethodDelete && id != "":
		deleteOwnPost(w, r, id)
	default:
		http.Error(w, "método não permitido", http.StatusMethodNotAllowed)
	}
}

// AdminPostsHandler — rotas exclusivas para admin
// GET    /posts       → lista todos
// GET    /posts/{id}  → busca um
// PUT    /posts/{id}  → atualiza qualquer post (inclui pin)
// DELETE /posts/{id}  → deleta qualquer post
func AdminPostsHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/posts")
	id = strings.Trim(id, "/")

	switch {
	case r.Method == http.MethodGet && id == "":
		listPosts(w, r)
	case r.Method == http.MethodGet && id != "":
		getPost(w, r, id)
	case r.Method == http.MethodPut && id != "":
		updatePost(w, r, id)
	case r.Method == http.MethodDelete && id != "":
		deletePost(w, r, id)
	default:
		http.Error(w, "método não permitido", http.StatusMethodNotAllowed)
	}
}

func listPosts(w http.ResponseWriter, r *http.Request) {
	rows, err := repository.DB.Query(
		`SELECT id, content, author_id, pin, posted_at, likes FROM posts`,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var list []entities.Post
	for rows.Next() {
		var p entities.Post
		if err := rows.Scan(&p.ID, &p.Content, &p.AuthorID, &p.Pin, &p.PostedAt, &p.Likes); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		list = append(list, p)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func getPost(w http.ResponseWriter, r *http.Request, rawID string) {
	id, err := strconv.Atoi(rawID)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}
	var p entities.Post
	err = repository.DB.QueryRow(
		`SELECT id, content, author_id, pin, posted_at, likes FROM posts WHERE id=$1`, id,
	).Scan(&p.ID, &p.Content, &p.AuthorID, &p.Pin, &p.PostedAt, &p.Likes)
	if err == sql.ErrNoRows {
		http.Error(w, "post não encontrado", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

// createPost — author_id é sempre o usuário logado, posted_at é sempre now()
func createPost(w http.ResponseWriter, r *http.Request) {
	callerID, ok := r.Context().Value(auth.UserIDKey).(int)
	if !ok {
		http.Error(w, "Não autenticado", http.StatusUnauthorized)
		return
	}

	var p entities.Post
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "corpo inválido", http.StatusBadRequest)
		return
	}
	p.AuthorID = callerID
	p.PostedAt = time.Now()

	err := repository.DB.QueryRow(
		`INSERT INTO posts (content, author_id, pin, posted_at, likes) VALUES ($1,$2,$3,$4,$5) RETURNING id`,
		p.Content, p.AuthorID, p.Pin, p.PostedAt, p.Likes,
	).Scan(&p.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(p)
}

// updateOwnPost — usuário só pode editar o content do próprio post (não altera pin nem likes)
func updateOwnPost(w http.ResponseWriter, r *http.Request, rawID string) {
	postID, err := strconv.Atoi(rawID)
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
	err = repository.DB.QueryRow(`SELECT author_id FROM posts WHERE id=$1`, postID).Scan(&authorID)
	if err == sql.ErrNoRows {
		http.Error(w, "post não encontrado", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if callerID != authorID {
		http.Error(w, "Acesso negado: você só pode editar seus próprios posts", http.StatusForbidden)
		return
	}

	var body struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "corpo inválido", http.StatusBadRequest)
		return
	}

	res, err := repository.DB.Exec(
		`UPDATE posts SET content=$1 WHERE id=$2`,
		body.Content, postID,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		http.Error(w, "post não encontrado", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "post atualizado"})
}

// deleteOwnPost — usuário só pode deletar o próprio post
func deleteOwnPost(w http.ResponseWriter, r *http.Request, rawID string) {
	postID, err := strconv.Atoi(rawID)
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
	err = repository.DB.QueryRow(`SELECT author_id FROM posts WHERE id=$1`, postID).Scan(&authorID)
	if err == sql.ErrNoRows {
		http.Error(w, "post não encontrado", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if callerID != authorID {
		http.Error(w, "Acesso negado: você só pode deletar seus próprios posts", http.StatusForbidden)
		return
	}

	deletePost(w, r, rawID)
}

// updatePost — versão admin: atualiza qualquer campo de qualquer post
func updatePost(w http.ResponseWriter, r *http.Request, rawID string) {
	id, err := strconv.Atoi(rawID)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}
	var p entities.Post
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "corpo inválido", http.StatusBadRequest)
		return
	}
	p.ID = id
	res, err := repository.DB.Exec(
		`UPDATE posts SET content=$1, pin=$2, likes=$3 WHERE id=$4`,
		p.Content, p.Pin, p.Likes, p.ID,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		http.Error(w, "post não encontrado", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

// deletePost — versão admin: deleta qualquer post
func deletePost(w http.ResponseWriter, r *http.Request, rawID string) {
	id, err := strconv.Atoi(rawID)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}
	res, err := repository.DB.Exec(`DELETE FROM posts WHERE id=$1`, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		http.Error(w, "post não encontrado", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}