package db

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

var testQueries *Queries
var testPool *pgxpool.Pool

const dbSource = "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable"

func TestMain(m *testing.M) {
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, dbSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}
	testPool = pool
	defer testPool.Close()

	testQueries = New(testPool)

	code := m.Run()
	os.Exit(code)
}
