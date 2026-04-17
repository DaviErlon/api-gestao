package repository

import (
	"database/sql"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func Connect() {
	dsn := os.Getenv("DSN")

	var err error
	DB, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}

	ticker := time.NewTicker(4 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		err := DB.Ping()
		if err == nil {
			log.Println("Conectado ao PostgreSQL com sucesso")
			return
		}

		log.Printf("[db] aguardando conexão... erro: %v", err)
	}
}
