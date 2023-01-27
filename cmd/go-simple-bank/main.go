package main

import (
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/max-rodziyevsky/go-simple-bank/api"
	"github.com/max-rodziyevsky/go-simple-bank/configs"
	"github.com/max-rodziyevsky/go-simple-bank/internal/repo"
	"log"
)

func main() {
	run()
}

func run() {
	config, err := configs.LoadConfig(".")
	if err != nil {
		log.Fatal("can't load config file", err)
	}
	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatalf("can't connect to %s database", config.DBDriver)
	}

	db := repo.NewStore(conn)
	server := api.NewServer(db)

	err = server.Start(config.ServerAddress)
	if err != nil {
		log.Fatalf("can't start the server")
	}
}
