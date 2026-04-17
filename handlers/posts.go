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

// ---------------------------------------------------------------------------
// Função auxiliar para obter empresa do autor do post
// ---------------------------------------------------------------------------

// empresaDoPost retorna o empresa_id do autor do post.
func empresaDoPost(postID int) (int, error) {
    var empresaID int
    err := repository.DB.QueryRow(
        `SELECT u.empresa_id
           FROM posts p
           JOIN users u ON u.id = p.author_id
          WHERE p.id = $1`, postID,
    ).Scan(&empresaID)
    return empresaID, err
}

// ---------------------------------------------------------------------------
// PostsHandler — usuário autenticado comum (com bolha de empresa)
// ---------------------------------------------------------------------------

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
    case r.Method == http.MethodPut && id != "":
        likeOnPost(w, r, id)
    case r.Method == http.MethodDelete && id != "":
        deleteOwnPost(w, r, id)
    default:
        http.Error(w, "método não permitido", http.StatusMethodNotAllowed)
    }
}

// ---------------------------------------------------------------------------
// AdminPostsHandler — admin (sem filtro de empresa)
// ---------------------------------------------------------------------------

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

// ---------------------------------------------------------------------------
// Handlers para usuário comum (com filtro por empresa)
// ---------------------------------------------------------------------------

// listPostsOwn lista apenas posts de usuários da mesma empresa do caller.
func listPostsOwn(w http.ResponseWriter, r *http.Request) {
    empresaID, err := empresaIDFromContext(r)
    if err != nil {
        http.Error(w, err.Error(), http.StatusForbidden)
        return
    }

    rows, err := repository.DB.Query(
        `SELECT p.id, p.content, p.author_id, p.pin, p.posted_at, p.likes
           FROM posts p
           JOIN users u ON u.id = p.author_id
          WHERE u.empresa_id = $1`, empresaID,
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

// getPostOwn retorna um post somente se ele pertencer à empresa do caller.
func getPostOwn(w http.ResponseWriter, r *http.Request, rawID string) {
    postID, err := strconv.Atoi(rawID)
    if err != nil {
        http.Error(w, "ID inválido", http.StatusBadRequest)
        return
    }

    callerEmpresa, err := empresaIDFromContext(r)
    if err != nil {
        http.Error(w, err.Error(), http.StatusForbidden)
        return
    }

    postEmpresa, err := empresaDoPost(postID)
    if err == sql.ErrNoRows {
        http.Error(w, "post não encontrado", http.StatusNotFound)
        return
    } else if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    if callerEmpresa != postEmpresa {
        http.Error(w, "Acesso negado: post de outra empresa", http.StatusForbidden)
        return
    }

    // Se passou, busca os dados do post
    var p entities.Post
    err = repository.DB.QueryRow(
        `SELECT id, content, author_id, pin, posted_at, likes FROM posts WHERE id=$1`, postID,
    ).Scan(&p.ID, &p.Content, &p.AuthorID, &p.Pin, &p.PostedAt, &p.Likes)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(p)
}

// createPost — cria post para o usuário logado (verifica se ele tem empresa).
func createPost(w http.ResponseWriter, r *http.Request) {
    // Verifica se o usuário tem empresa (e já obtém o callerID implicitamente)
    _, err := empresaIDFromContext(r)
    if err != nil {
        http.Error(w, err.Error(), http.StatusForbidden)
        return
    }
    // Extrai o userID do contexto
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

    err = repository.DB.QueryRow(
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

// likeOnPost(w http.ResponseWritter)
func likeOnPost(w http.ResponseWriter, r *http.Request, rawID string) {
	postID, err := strconv.Atoi(rawID)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	callerEmpresa, err := empresaIDFromContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	postEmpresa, err := empresaDoPost(postID)
	if err == sql.ErrNoRows {
		http.Error(w, "post não encontrado", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if callerEmpresa != postEmpresa {
		http.Error(w, "Acesso negado: post de outra empresa", http.StatusForbidden)
		return
	}

	userID, ok := callerID(r)
	if !ok {
		http.Error(w, "Não autenticado", http.StatusUnauthorized)
		return
	}

	// 🔁 tenta remover (deslike)
	res, err := repository.DB.Exec(
		`DELETE FROM likes WHERE usuario_id=$1 AND post_id=$2`,
		userID, postID,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rows, _ := res.RowsAffected()

	if rows > 0 {
		// 👍 deu deslike
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// ❤️ não tinha → adiciona like
	_, err = repository.DB.Exec(
		`INSERT INTO likes (usuario_id, post_id) VALUES ($1, $2)`,
		userID, postID,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// deleteOwnPost — deleta apenas o próprio post e da mesma empresa.
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

    callerEmpresa, err := empresaIDFromContext(r)
    if err != nil {
        http.Error(w, err.Error(), http.StatusForbidden)
        return
    }

    var authorID, postEmpresa int
    err = repository.DB.QueryRow(
        `SELECT p.author_id, u.empresa_id
           FROM posts p
           JOIN users u ON u.id = p.author_id
          WHERE p.id = $1`, postID,
    ).Scan(&authorID, &postEmpresa)
    if err == sql.ErrNoRows {
        http.Error(w, "post não encontrado", http.StatusNotFound)
        return
    } else if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    if callerID != authorID || callerEmpresa != postEmpresa {
        http.Error(w, "Acesso negado: você só pode deletar seus próprios posts", http.StatusForbidden)
        return
    }

    deletePost(w, r, rawID)
}

// ---------------------------------------------------------------------------
// Handlers para admin (sem filtros)
// ---------------------------------------------------------------------------

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