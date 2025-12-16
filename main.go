package main

import (
	"context"
	"log"

	"github.com/WilliamOdinson/simplebank/api"
	db "github.com/WilliamOdinson/simplebank/db/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	dbSource      = "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable"
	serverAddress = "0.0.0.0:8080"
)

func main() {
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, dbSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	defer pool.Close()
	store := db.NewStore(pool)
	server := api.NewServer(store)

	log.Printf("Starting server at %s", serverAddress)
	err = server.Start(serverAddress)
	if err != nil {
		log.Fatal("cannot start server:", err)
	}
}
