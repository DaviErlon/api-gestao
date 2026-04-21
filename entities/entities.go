package entities

import (
	"time"
)

// Empresa representa a tabela "empresas"
type Empresa struct {
	ID   int    `db:"id"   json:"id"`
	Name string `db:"name" json:"name"`
}

// User representa a tabela "users"
type User struct {
	ID        int    `db:"id"         json:"id"`
	Name      string `db:"name"       json:"name"`
	Login     string `db:"login"      json:"login"`
	Password  string `db:"password"   json:"password,omitempty"`
	EmpresaID int    `db:"empresa_id" json:"empresa_id"`
}

// Ciclo representa a tabela "ciclos"
type Ciclo struct {
	ID          int     `db:"id"           json:"id"`
	Rodada      int     `db:"rodada"       json:"rodada"`
	Saldo       float64 `db:"saldo"        json:"saldo"`
	Divida      float64 `db:"divida"       json:"divida"`
	Juros       float64 `db:"juros"        json:"juros"`
	Clientes    int     `db:"clientes"     json:"clientes"`
	MarketShare float64 `db:"market_share" json:"market_share"`
	Tech        int     `db:"tech"         json:"tech"`
	Reputacao   int     `db:"reputacao"    json:"reputacao"`
	Seguranca   int     `db:"seguranca"    json:"seguranca"`
	NPS         int     `db:"nps"          json:"nps"`
	Valuation   float64 `db:"valuation"    json:"valuation"`
	EmpresaID   int     `db:"empresa_id"   json:"empresa_id"`
}

// Decisao representa a tabela "decisoes"
type Decisao struct {
	ID        int     `db:"id"         json:"id"`
	Marketing float64 `db:"marketing"  json:"marketing"`
	PeD       float64 `db:"ped"        json:"ped"`
	Suporte   float64 `db:"suporte"    json:"suporte"`
	Seguranca float64 `db:"seguranca"  json:"seguranca"`
	Expansao  int     `db:"expansao"   json:"expansao"`
	EmpresaID int     `db:"empresa_id" json:"empresa_id"`
	CicloID   int     `db:"ciclo_id"   json:"ciclo_id"`
}

// Post representa a tabela "posts"
type Post struct {
	ID         int       `db:"id"        json:"id"`
	Content    string    `db:"content"   json:"content"`
	AuthorID   int       `db:"author_id" json:"author_id"`
	AuthorName string    `json:"author_name"`
	Pin        bool      `db:"pin"       json:"pin"`
	PostedAt   time.Time `db:"posted_at" json:"posted_at"`
	Likes      int       `db:"likes"     json:"likes"`
	Curtido    bool      `json:"curtido"`
}

// Reuniao representa a tabela "reunioes"
// O campo Duracao usa time.Duration para mapear o tipo INTERVAL do PostgreSQL.
// Ao usar pgx/sqlx, converta para/de time.Duration conforme necessário.
type Reuniao struct {
	ID        int           `db:"id"        json:"id"`
	CicloID   int           `db:"ciclo_id"  json:"ciclo_id"`
	AuthorID  int           `db:"author_id" json:"author_id"`
	Titulo    string        `db:"titulo"    json:"titulo"`
	Descricao *string       `db:"descricao" json:"descricao,omitempty"`
	Inicio    time.Time     `db:"inicio"    json:"inicio"`
	Duracao   time.Duration `db:"duracao"   json:"duracao"`
	Aberta    bool          `db:"aberta"    json:"aberta"`
}

type Pagination struct {
	Page     int  `json:"page"`
	Limit    int  `json:"limit"`
	HasNext  bool `json:"has_next"`
}

type PaginatedResponse[T any] struct {
	Data       []T        `json:"data"`
	Pagination Pagination `json:"pagination"`
}