package main

import (
	"context"
	"log"

	"github.com/WilliamOdinson/simplebank/api"
	db "github.com/WilliamOdinson/simplebank/db/sqlc"
	"github.com/WilliamOdinson/simplebank/util"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()
	config, err := util.LoadConfig(".")

	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	pool, err := pgxpool.New(ctx, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	defer pool.Close()
	store := db.NewStore(pool)
	server := api.NewServer(store)

	log.Printf("Starting server at %s", config.ServerAddress)
	err = server.Start(config.ServerAddress)
	if err != nil {
		log.Fatal("cannot start server:", err)
	}
}
