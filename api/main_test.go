package api

import (
	"os"
	"testing"
	"time"

	db "github.com/WilliamOdinson/simplebank/db/sqlc"
	"github.com/WilliamOdinson/simplebank/util"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/gin-gonic/gin"
)

func newTestServer(t *testing.T, store db.Store) *Server {
	config := util.Config{
		DBSource:            "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable",
		ServerAddress:       "0.0.0.0:8080",
		TokenSymmetricKey:   gofakeit.LetterN(32),
		AccessTokenDuration: 15 * time.Minute,
	}

	server, err := NewServer(config, store)
	if err != nil {
		t.Fatalf("cannot create server: %v", err)
	}

	return server
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}
