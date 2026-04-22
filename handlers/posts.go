package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/DaviErlon/api-gestao/auth"
	"github.com/DaviErlon/api-gestao/entities"
	"github.com/DaviErlon/api-gestao/repository"
)

// =========================================================
// AUXILIARES
// =========================================================

func enrichPostWithLikes(p *entities.Post, userID int) error {
	if err := repository.DB.QueryRow(`
		SELECT COUNT(*) FROM likes WHERE post_id = $1
	`, p.ID).Scan(&p.Likes); err != nil {
		return err
	}

	if err := repository.DB.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM likes
			WHERE post_id = $1 AND usuario_id = $2
		)
	`, p.ID, userID).Scan(&p.Curtido); err != nil {
		return err
	}

	return nil
}

func getPagination(r *http.Request) (limit, offset, page int) {
	q := r.URL.Query()

	limit = 10
	page = 1

	if l := q.Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}

	if p := q.Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}

	offset = (page - 1) * limit
	return
}

// =========================================================
// ROUTER USER
// =========================================================

func PostsHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/posts")
	id = strings.Trim(id, "/")

	switch {
	case r.Method == http.MethodGet && id == "":
		listPostsOwn(w, r)

	case r.Method == http.MethodGet && id != "":
		getPostOwn(w, r, id)

	case r.Method == http.MethodPost:
		createPost(w, r)

	case r.Method == http.MethodPut && strings.HasSuffix(r.URL.Path, "/like"):
		toggleLikeOnPost(w, r, id)

	case r.Method == http.MethodPut && strings.HasSuffix(r.URL.Path, "/pin"):
		togglePinOnPost(w, r, id)

	case r.Method == http.MethodDelete && id != "":
		deleteOwnPost(w, r, id)

	default:
		http.Error(w, "método não permitido", http.StatusMethodNotAllowed)
	}
}

// =========================================================
// ROUTER ADMIN
// =========================================================

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

// =========================================================
// USER POSTS
// =========================================================
func listPostsOwn(w http.ResponseWriter, r *http.Request) {
	empresaID, err := empresaIDFromContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	userID, _ := callerID(r)
	limit, offset, page := getPagination(r)

	// 🔥 truque: pega +1 pra saber se tem próxima página
	rows, err := repository.DB.Query(`
		SELECT 
			p.id,
			p.content,
			p.author_id,
			u.name,
			p.pin,
			p.posted_at
		FROM posts p
		JOIN users u ON u.id = p.author_id
		WHERE u.empresa_id = $1
		ORDER BY p.pin DESC, p.posted_at DESC
		LIMIT $2 OFFSET $3
	`, empresaID, limit+1, offset)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var list []entities.Post

	for rows.Next() {
		var p entities.Post

		if err := rows.Scan(
			&p.ID,
			&p.Content,
			&p.AuthorID,
			&p.AuthorName,
			&p.Pin,
			&p.PostedAt,
		); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_ = enrichPostWithLikes(&p, userID)
		list = append(list, p)
	}

	// 🔥 detecta próxima página
	hasNext := len(list) > limit

	if hasNext {
		list = list[:limit] // remove o extra
	}

	response := entities.PaginatedResponse[entities.Post]{
		Data: list,
		Pagination: entities.Pagination{
			Page:    page,
			Limit:   limit,
			HasNext: hasNext,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func getPostOwn(w http.ResponseWriter, r *http.Request, rawID string) {
	postID, _ := strconv.Atoi(rawID)
	userID, _ := callerID(r)

	empresaID, err := empresaIDFromContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	var p entities.Post

	err = repository.DB.QueryRow(`
		SELECT 
			p.id,
			p.content,
			p.author_id,
			u.name,
			p.pin,
			p.posted_at
		FROM posts p
		JOIN users u ON u.id = p.author_id
		WHERE p.id = $1 AND u.empresa_id = $2
	`, postID, empresaID).Scan(
		&p.ID,
		&p.Content,
		&p.AuthorID,
		&p.AuthorName,
		&p.Pin,
		&p.PostedAt,
	)

	if err != nil {
		http.Error(w, "post não encontrado", http.StatusNotFound)
		return
	}

	_ = enrichPostWithLikes(&p, userID)

	json.NewEncoder(w).Encode(p)
}

func createPost(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(auth.UserIDKey).(int)
	if !ok {
		http.Error(w, "não autenticado", http.StatusUnauthorized)
		return
	}

	var p entities.Post
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "corpo inválido", http.StatusBadRequest)
		return
	}

	p.AuthorID = userID
	p.PostedAt = time.Now()

	err := repository.DB.QueryRow(`
		INSERT INTO posts (content, author_id, pin, posted_at)
		VALUES ($1,$2,$3,$4)
		RETURNING id
	`, p.Content, p.AuthorID, p.Pin, p.PostedAt).Scan(&p.ID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(p)
}

// =========================================================
// LIKE / PIN
// =========================================================

func toggleLikeOnPost(w http.ResponseWriter, r *http.Request, rawID string) {
	idStr := strings.Split(rawID, "/")[0]
	postID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	userID, ok := callerID(r)
	if !ok {
		http.Error(w, "não autenticado", http.StatusUnauthorized)
		return
	}

	empresaID, err := empresaIDFromContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	var exists bool
	err = repository.DB.QueryRow(`
		SELECT EXISTS(
			SELECT 1
			FROM posts p
			JOIN users u ON u.id = p.author_id
			WHERE p.id = $1 AND u.empresa_id = $2
		)
	`, postID, empresaID).Scan(&exists)

	if err != nil {
		http.Error(w, "erro interno", http.StatusInternalServerError)
		return
	}

	if !exists {
		http.Error(w, "post não encontrado", http.StatusNotFound)
		return
	}

	res, err := repository.DB.Exec(`
		DELETE FROM likes WHERE usuario_id=$1 AND post_id=$2
	`, userID, postID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rows, _ := res.RowsAffected()
	if rows > 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	_, err = repository.DB.Exec(`
		INSERT INTO likes (usuario_id, post_id)
		VALUES ($1,$2)
	`, userID, postID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func togglePinOnPost(w http.ResponseWriter, r *http.Request, rawID string) {
	postID, _ := strconv.Atoi(strings.Split(rawID, "/")[0])

	empresaID, err := empresaIDFromContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	_, err = repository.DB.Exec(`
		UPDATE posts
		SET pin = NOT pin
		WHERE id = $1 AND author_id IN (
			SELECT id FROM users WHERE empresa_id = $2
		)
	`, postID, empresaID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// =========================================================
// DELETE USER POST
// =========================================================

func deleteOwnPost(w http.ResponseWriter, r *http.Request, rawID string) {
	postID, _ := strconv.Atoi(rawID)

	empresaID, err := empresaIDFromContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	_, err = repository.DB.Exec(`
		DELETE FROM posts
		WHERE id = $1 AND author_id IN (
			SELECT id FROM users WHERE empresa_id = $2
		)
	`, postID, empresaID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// =========================================================
// ADMIN POSTS
// =========================================================

func listPosts(w http.ResponseWriter, r *http.Request) {
	userID, _ := callerID(r)
	limit, offset, page := getPagination(r)

	rows, err := repository.DB.Query(`
		SELECT 
			p.id,
			p.content,
			p.author_id,
			u.name,
			p.pin,
			p.posted_at
		FROM posts p
		JOIN users u ON u.id = p.author_id
		ORDER BY p.pin DESC, p.posted_at DESC
		LIMIT $1 OFFSET $2
	`, limit+1, offset)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var list []entities.Post

	for rows.Next() {
		var p entities.Post

		if err := rows.Scan(
			&p.ID,
			&p.Content,
			&p.AuthorID,
			&p.AuthorName,
			&p.Pin,
			&p.PostedAt,
		); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_ = enrichPostWithLikes(&p, userID)
		list = append(list, p)
	}

	hasNext := len(list) > limit

	if hasNext {
		list = list[:limit]
	}

	response := entities.PaginatedResponse[entities.Post]{
		Data: list,
		Pagination: entities.Pagination{
			Page:    page,
			Limit:   limit,
			HasNext: hasNext,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func getPost(w http.ResponseWriter, r *http.Request, rawID string) {
	id, _ := strconv.Atoi(rawID)
	userID, _ := callerID(r)

	var p entities.Post

	err := repository.DB.QueryRow(`
		SELECT 
			p.id,
			p.content,
			p.author_id,
			u.name,
			p.pin,
			p.posted_at
		FROM posts p
		JOIN users u ON u.id = p.author_id
		WHERE p.id=$1
	`, id).Scan(
		&p.ID,
		&p.Content,
		&p.AuthorID,
		&p.AuthorName,
		&p.Pin,
		&p.PostedAt,
	)

	if err != nil {
		http.Error(w, "post não encontrado", http.StatusNotFound)
		return
	}

	_ = enrichPostWithLikes(&p, userID)

	json.NewEncoder(w).Encode(p)
}

func updatePost(w http.ResponseWriter, r *http.Request, rawID string) {
	id, _ := strconv.Atoi(rawID)

	var p entities.Post
	json.NewDecoder(r.Body).Decode(&p)

	repository.DB.Exec(`
		UPDATE posts SET content=$1, pin=$2 WHERE id=$3
	`, p.Content, p.Pin, id)

	json.NewEncoder(w).Encode(p)
}

func deletePost(w http.ResponseWriter, r *http.Request, rawID string) {
	id, _ := strconv.Atoi(rawID)

	repository.DB.Exec(`
		DELETE FROM posts WHERE id=$1
	`, id)

	w.WriteHeader(http.StatusNoContent)
}
