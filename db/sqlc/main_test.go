package db

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/WilliamOdinson/simplebank/util"
	"github.com/jackc/pgx/v5/pgxpool"
)

var testQueries *Queries
var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	ctx := context.Background()
	config, err := util.LoadConfig("../..")

	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	pool, err := pgxpool.New(ctx, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}
	testPool = pool
	defer testPool.Close()

	testQueries = New(testPool)

	code := m.Run()
	os.Exit(code)
}
