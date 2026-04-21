package repository

import (
	"context"
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

	// ⏱ timeout 
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	for {
		if ctx.Err() != nil {
			log.Fatal("[db] tempo limite de conexão excedido (8s)")
		}

		err = DB.PingContext(ctx)
		if err == nil {
			log.Println("Conectado ao PostgreSQL com sucesso")
			return
		}

		log.Printf("[db] tentando conectar... erro: %v", err)

		time.Sleep(2 * time.Second)
	}
}