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

// Usuário comum
func CiclosHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/ciclos")
	id = strings.Trim(id, "/")

	switch {
	case r.Method == http.MethodGet && id == "":
		listCiclosOwn(w, r)
	case r.Method == http.MethodGet && id != "":
		getCicloOwn(w, r, id)
	case r.Method == http.MethodPost:
		createCicloOwn(w, r)
	case r.Method == http.MethodPut && id != "":
		updateCicloOwn(w, r, id)
	case r.Method == http.MethodDelete && id != "":
		deleteCicloOwn(w, r, id)
	default:
		http.Error(w, "método não permitido", http.StatusMethodNotAllowed)
	}
}

// Admin
func AdminCiclosHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/ciclos")
	id = strings.Trim(id, "/")

	switch {
	case r.Method == http.MethodGet && id == "":
		listCiclos(w, r)
	case r.Method == http.MethodGet && id != "":
		getCiclo(w, r, id)
	case r.Method == http.MethodPut && id != "":
		updateCiclo(w, r, id)
	case r.Method == http.MethodDelete && id != "":
		deleteCiclo(w, r, id)
	default:
		http.Error(w, "método não permitido", http.StatusMethodNotAllowed)
	}
}

// ---------------------------------------------------------------------------
// com filtro de empresa
// ---------------------------------------------------------------------------

func listCiclosOwn(w http.ResponseWriter, r *http.Request) {
	empresaID, err := empresaIDFromContext(r)
	if err != nil {
		http.Error(w, "Não autenticado", http.StatusUnauthorized)
		return
	}

	rows, err := repository.DB.Query(`
		SELECT id, rodada, saldo, divida, juros, clientes, market_share,
		       tech, reputacao, seguranca, nps, valuation, empresa_id
		FROM ciclos
		WHERE empresa_id=$1
		ORDER BY rodada ASC
	`, empresaID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	list := []entities.Ciclo{}
	for rows.Next() {
		var c entities.Ciclo
		if err := rows.Scan(
			&c.ID, &c.Rodada, &c.Saldo, &c.Divida, &c.Juros, &c.Clientes,
			&c.MarketShare, &c.Tech, &c.Reputacao, &c.Seguranca, &c.NPS,
			&c.Valuation, &c.EmpresaID,
		); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		list = append(list, c)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func getCicloOwn(w http.ResponseWriter, r *http.Request, rawID string) {
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

	var c entities.Ciclo
	err = repository.DB.QueryRow(`
		SELECT id, rodada, saldo, divida, juros, clientes, market_share,
		       tech, reputacao, seguranca, nps, valuation, empresa_id
		FROM ciclos
		WHERE id=$1 AND empresa_id=$2
	`, id, empresaID).Scan(
		&c.ID, &c.Rodada, &c.Saldo, &c.Divida, &c.Juros, &c.Clientes,
		&c.MarketShare, &c.Tech, &c.Reputacao, &c.Seguranca, &c.NPS,
		&c.Valuation, &c.EmpresaID,
	)

	if err == sql.ErrNoRows {
		http.Error(w, "ciclo não encontrado", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(c)
}

func createCicloOwn(w http.ResponseWriter, r *http.Request) {
	empresaID, err := empresaIDFromContext(r)
	if err != nil {
		http.Error(w, "Não autenticado", http.StatusUnauthorized)
		return
	}

	var c entities.Ciclo
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, "corpo inválido", http.StatusBadRequest)
		return
	}

	c.EmpresaID = empresaID

	err = repository.DB.QueryRow(`
		INSERT INTO ciclos
		(rodada, saldo, divida, juros, clientes, market_share, tech, reputacao, seguranca, nps, valuation, empresa_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		RETURNING id
	`, c.Rodada, c.Saldo, c.Divida, c.Juros, c.Clientes, c.MarketShare,
		c.Tech, c.Reputacao, c.Seguranca, c.NPS, c.Valuation, c.EmpresaID,
	).Scan(&c.ID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(c)
}

func updateCicloOwn(w http.ResponseWriter, r *http.Request, rawID string) {
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

	var c entities.Ciclo
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, "corpo inválido", http.StatusBadRequest)
		return
	}

	res, err := repository.DB.Exec(`
		UPDATE ciclos SET
			rodada=$1, saldo=$2, divida=$3, juros=$4, clientes=$5,
			market_share=$6, tech=$7, reputacao=$8, seguranca=$9,
			nps=$10, valuation=$11
		WHERE id=$12 AND empresa_id=$13
	`, c.Rodada, c.Saldo, c.Divida, c.Juros, c.Clientes, c.MarketShare,
		c.Tech, c.Reputacao, c.Seguranca, c.NPS, c.Valuation, id, empresaID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if n, _ := res.RowsAffected(); n == 0 {
		http.Error(w, "ciclo não encontrado", http.StatusNotFound)
		return
	}

	c.ID = id
	c.EmpresaID = empresaID

	json.NewEncoder(w).Encode(c)
}

func deleteCicloOwn(w http.ResponseWriter, r *http.Request, rawID string) {
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
		`DELETE FROM ciclos WHERE id=$1 AND empresa_id=$2`,
		id, empresaID,
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if n, _ := res.RowsAffected(); n == 0 {
		http.Error(w, "ciclo não encontrado", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ---------------------------------------------------------------------------
// sem filtro
// ---------------------------------------------------------------------------

func listCiclos(w http.ResponseWriter, r *http.Request) {
	rows, err := repository.DB.Query(`SELECT * FROM ciclos ORDER BY rodada ASC`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var list []entities.Ciclo
	for rows.Next() {
		var c entities.Ciclo
		if err := rows.Scan(
			&c.ID, &c.Rodada, &c.Saldo, &c.Divida, &c.Juros, &c.Clientes,
			&c.MarketShare, &c.Tech, &c.Reputacao, &c.Seguranca,
			&c.NPS, &c.Valuation, &c.EmpresaID,
		); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		list = append(list, c)
	}

	json.NewEncoder(w).Encode(list)
}

func getCiclo(w http.ResponseWriter, r *http.Request, rawID string) {
	id, _ := strconv.Atoi(rawID)

	var c entities.Ciclo
	err := repository.DB.QueryRow(`SELECT * FROM ciclos WHERE id=$1`, id).Scan(
		&c.ID, &c.Rodada, &c.Saldo, &c.Divida, &c.Juros, &c.Clientes,
		&c.MarketShare, &c.Tech, &c.Reputacao, &c.Seguranca,
		&c.NPS, &c.Valuation, &c.EmpresaID,
	)

	if err == sql.ErrNoRows {
		http.Error(w, "ciclo não encontrado", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(c)
}

func updateCiclo(w http.ResponseWriter, r *http.Request, rawID string) {
	id, _ := strconv.Atoi(rawID)

	var c entities.Ciclo
	json.NewDecoder(r.Body).Decode(&c)

	res, _ := repository.DB.Exec(`
		UPDATE ciclos SET
			rodada=$1, saldo=$2, divida=$3, juros=$4, clientes=$5,
			market_share=$6, tech=$7, reputacao=$8, seguranca=$9,
			nps=$10, valuation=$11, empresa_id=$12
		WHERE id=$13
	`, c.Rodada, c.Saldo, c.Divida, c.Juros, c.Clientes, c.MarketShare,
		c.Tech, c.Reputacao, c.Seguranca, c.NPS, c.Valuation, c.EmpresaID, id)

	if n, _ := res.RowsAffected(); n == 0 {
		http.Error(w, "ciclo não encontrado", http.StatusNotFound)
		return
	}

	c.ID = id
	json.NewEncoder(w).Encode(c)
}

func deleteCiclo(w http.ResponseWriter, r *http.Request, rawID string) {
	id, _ := strconv.Atoi(rawID)

	res, _ := repository.DB.Exec(`DELETE FROM ciclos WHERE id=$1`, id)

	if n, _ := res.RowsAffected(); n == 0 {
		http.Error(w, "ciclo não encontrado", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}