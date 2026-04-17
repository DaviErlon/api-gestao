package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/DaviErlon/api-gestao/auth"
	"github.com/DaviErlon/api-gestao/email"
	"github.com/DaviErlon/api-gestao/entities"
	"github.com/DaviErlon/api-gestao/repository"
)

// ---------------------------------------------------------------------------
// Scheduler de e-mails
// ---------------------------------------------------------------------------

type emailJob struct {
	cancel context.CancelFunc
}

var (
	emailJobs   = make(map[int]*emailJob)
	emailJobsMu sync.Mutex
)

func InitReunioesScheduler() {
	rows, err := repository.DB.Query(`
		SELECT r.id, r.titulo, r.inicio, c.empresa_id
		FROM reunioes r
		JOIN ciclos c ON c.id = r.ciclo_id
	`)
	if err != nil {
		log.Printf("[init] erro ao carregar reuniões: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id, empresaID int
		var titulo string
		var inicio time.Time

		if err := rows.Scan(&id, &titulo, &inicio, &empresaID); err != nil {
			continue
		}

		scheduleEmailJob(id, empresaID, titulo, inicio)
	}
}

func scheduleEmailJob(reuniaoID, empresaID int, titulo string, inicio time.Time) {
	cancelEmailJob(reuniaoID)

	delay := time.Until(inicio.Add(-20 * time.Minute))
	ctx, cancel := context.WithCancel(context.Background())

	emailJobsMu.Lock()
	emailJobs[reuniaoID] = &emailJob{cancel: cancel}
	emailJobsMu.Unlock()

	go func() {
		defer func() {
			emailJobsMu.Lock()
			delete(emailJobs, reuniaoID)
			emailJobsMu.Unlock()
		}()

		if delay <= 0 {
			sendEmailToEmpresa(empresaID, titulo, inicio)
			return
		}

		timer := time.NewTimer(delay)
		defer timer.Stop()

		select {
		case <-timer.C:
			sendEmailToEmpresa(empresaID, titulo, inicio)
		case <-ctx.Done():
			return
		}
	}()
}

func cancelEmailJob(reuniaoID int) {
	emailJobsMu.Lock()
	defer emailJobsMu.Unlock()

	if job, ok := emailJobs[reuniaoID]; ok {
		job.cancel()
		delete(emailJobs, reuniaoID)
	}
}

func sendEmailToEmpresa(empresaID int, titulo string, inicio time.Time) {
	rows, err := repository.DB.Query(
		`SELECT name, login FROM users WHERE empresa_id=$1`, empresaID,
	)
	if err != nil {
		return
	}
	defer rows.Close()

	dataHora := inicio.Format("02/01/2006 15:04")

	for rows.Next() {
		var name, login string
		rows.Scan(&name, &login)

		subject := fmt.Sprintf("Reunião: %s", titulo)
		body := fmt.Sprintf("Olá %s, reunião %s às %s", name, titulo, dataHora)

		email.EMAIL.Send(login, subject, body)
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func ultimoCicloDaEmpresa(empresaID int) (int, error) {
	var cicloID int

	err := repository.DB.QueryRow(`
		SELECT id
		FROM ciclos
		WHERE empresa_id = $1
		ORDER BY rodada DESC
		LIMIT 1
	`, empresaID).Scan(&cicloID)

	return cicloID, err
}

func callerID(r *http.Request) (int, bool) {
	id, ok := r.Context().Value(auth.UserIDKey).(int)
	return id, ok
}

func empresaDaReuniao(reuniaoID int) (int, error) {
	var empresaID int
	err := repository.DB.QueryRow(
		`SELECT c.empresa_id
		 FROM reunioes r
		 JOIN ciclos c ON c.id = r.ciclo_id
		 WHERE r.id=$1`, reuniaoID,
	).Scan(&empresaID)
	return empresaID, err
}

func empresaDoCiclo(cicloID int) (int, error) {
	var empresaID int
	err := repository.DB.QueryRow(
		`SELECT empresa_id FROM ciclos WHERE id=$1`, cicloID,
	).Scan(&empresaID)
	return empresaID, err
}

// ---------------------------------------------------------------------------
// ROUTERS
// ---------------------------------------------------------------------------

func ReunioesHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.Trim(strings.TrimPrefix(r.URL.Path, "/reunioes"), "/")

	switch {
	case r.Method == http.MethodGet && id == "":
		listReunioesOwn(w, r)
	case r.Method == http.MethodGet && id != "":
		getReuniaoOwn(w, r, id)
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

func AdminReunioesHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.Trim(strings.TrimPrefix(r.URL.Path, "/reunioes"), "/")

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

// ---------------------------------------------------------------------------
// Com filtro
// ---------------------------------------------------------------------------

func listReunioesOwn(w http.ResponseWriter, r *http.Request) {
	empresaID, _ := empresaIDFromContext(r)

	rows, _ := repository.DB.Query(`
		SELECT r.id, r.ciclo_id, r.author_id, r.titulo, r.descricao, r.inicio, r.duracao
		FROM reunioes r
		JOIN ciclos c ON c.id=r.ciclo_id
		WHERE c.empresa_id=$1`, empresaID)

	defer rows.Close()

	writeList(w, rows)
}

func getReuniaoOwn(w http.ResponseWriter, r *http.Request, rawID string) {
	id, _ := strconv.Atoi(rawID)

	callerEmpresa, _ := empresaIDFromContext(r)
	reuniaoEmpresa, _ := empresaDaReuniao(id)

	if callerEmpresa != reuniaoEmpresa {
		http.Error(w, "acesso negado", http.StatusForbidden)
		return
	}

	getReuniao(w, r, rawID)
}

func createReuniao(w http.ResponseWriter, r *http.Request) {
	var re entities.Reuniao
	if err := json.NewDecoder(r.Body).Decode(&re); err != nil {
		http.Error(w, "corpo inválido", http.StatusBadRequest)
		return
	}

	uid, ok := callerID(r)
	if !ok {
		http.Error(w, "não autenticado", http.StatusUnauthorized)
		return
	}

	empresaID, err := empresaIDFromContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	cicloID, err := ultimoCicloDaEmpresa(empresaID)
	if err == sql.ErrNoRows {
		http.Error(w, "empresa não possui ciclos", http.StatusBadRequest)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	re.CicloID = cicloID
	re.AuthorID = uid

	err = repository.DB.QueryRow(
		`INSERT INTO reunioes (ciclo_id, author_id, titulo, descricao, inicio, duracao)
		 VALUES ($1,$2,$3,$4,$5,$6) RETURNING id`,
		re.CicloID, re.AuthorID, re.Titulo, re.Descricao, re.Inicio, re.Duracao,
	).Scan(&re.ID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	scheduleEmailJob(re.ID, empresaID, re.Titulo, re.Inicio)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(re)
}

func updateOwnReuniao(w http.ResponseWriter, r *http.Request, rawID string) {
	id, _ := strconv.Atoi(rawID)

	uid, _ := callerID(r)

	var authorID int
	repository.DB.QueryRow(`SELECT author_id FROM reunioes WHERE id=$1`, id).Scan(&authorID)

	if uid != authorID {
		http.Error(w, "acesso negado", http.StatusForbidden)
		return
	}

	updateReuniao(w, r, rawID)
}

func deleteOwnReuniao(w http.ResponseWriter, r *http.Request, rawID string) {
	id, _ := strconv.Atoi(rawID)

	uid, _ := callerID(r)

	var authorID int
	repository.DB.QueryRow(`SELECT author_id FROM reunioes WHERE id=$1`, id).Scan(&authorID)

	if uid != authorID {
		http.Error(w, "acesso negado", http.StatusForbidden)
		return
	}

	deleteReuniao(w, r, rawID)
}

// ---------------------------------------------------------------------------
// Sem filtro
// ---------------------------------------------------------------------------

func listReunioes(w http.ResponseWriter, r *http.Request) {
	rows, _ := repository.DB.Query(
		`SELECT id, ciclo_id, author_id, titulo, descricao, inicio, duracao FROM reunioes`,
	)
	defer rows.Close()

	writeList(w, rows)
}

func getReuniao(w http.ResponseWriter, r *http.Request, rawID string) {
	id, _ := strconv.Atoi(rawID)

	var re entities.Reuniao
	err := repository.DB.QueryRow(
		`SELECT id, ciclo_id, author_id, titulo, descricao, inicio, duracao
		 FROM reunioes WHERE id=$1`, id,
	).Scan(&re.ID, &re.CicloID, &re.AuthorID, &re.Titulo, &re.Descricao, &re.Inicio, &re.Duracao)

	if err == sql.ErrNoRows {
		http.Error(w, "não encontrado", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(re)
}

func updateReuniao(w http.ResponseWriter, r *http.Request, rawID string) {
	id, _ := strconv.Atoi(rawID)

	var re entities.Reuniao
	json.NewDecoder(r.Body).Decode(&re)
	re.ID = id

	repository.DB.Exec(
		`UPDATE reunioes SET ciclo_id=$1, titulo=$2, descricao=$3, inicio=$4, duracao=$5 WHERE id=$6`,
		re.CicloID, re.Titulo, re.Descricao, re.Inicio, re.Duracao, re.ID,
	)

	empresaID, _ := empresaDoCiclo(re.CicloID)
	scheduleEmailJob(re.ID, empresaID, re.Titulo, re.Inicio)

	json.NewEncoder(w).Encode(re)
}

func deleteReuniao(w http.ResponseWriter, r *http.Request, rawID string) {
	id, _ := strconv.Atoi(rawID)

	cancelEmailJob(id)
	repository.DB.Exec(`DELETE FROM reunioes WHERE id=$1`, id)

	w.WriteHeader(http.StatusNoContent)
}

// ---------------------------------------------------------------------------
// UTIL
// ---------------------------------------------------------------------------

func writeList(w http.ResponseWriter, rows *sql.Rows) {
	var list []entities.Reuniao

	for rows.Next() {
		var re entities.Reuniao
		rows.Scan(
			&re.ID, &re.CicloID, &re.AuthorID,
			&re.Titulo, &re.Descricao,
			&re.Inicio, &re.Duracao,
		)
		list = append(list, re)
	}

	json.NewEncoder(w).Encode(list)
}