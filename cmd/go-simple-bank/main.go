package main

import (
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/max-rodziyevsky/go-simple-bank/api"
	"github.com/max-rodziyevsky/go-simple-bank/internal/repo"
	"log"
)

const (
	dbDriver = "postgres"
	dbSource = "postgresql://root:secret@localhost:5433/go-simple-bank?sslmode=disable"
	address  = "0.0.0.0:8080"
)

func main() {
	run()
}

func run() {
	conn, err := sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatalf("can't connect to %s database", dbDriver)
	}

	db := repo.NewStore(conn)
	server := api.NewServer(db)

	err = server.Start(address)
	if err != nil {
		log.Fatalf("can't start the server")
	}
}
