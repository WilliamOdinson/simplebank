package db

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
)

func createRandomAccount(t *testing.T) (Account, CreateAccountParams) {
	t.Helper()

	arg := CreateAccountParams{
		Owner:    gofakeit.Name(),
		Balance:  int64(gofakeit.Price(0, 10000)),
		Currency: gofakeit.CurrencyShort(),
	}

	acc, err := testQueries.CreateAccount(context.Background(), arg)

	if err != nil {
		t.Fatal("Cannot create account:", err)
	}

	return acc, arg
}
