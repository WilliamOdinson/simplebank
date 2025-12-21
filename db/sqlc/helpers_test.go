package db

import (
	"context"
	"testing"

	"github.com/WilliamOdinson/simplebank/util"
	"github.com/brianvoe/gofakeit/v7"
)

var supportedCurrencies = []string{util.USD, util.EUR, util.CAD}

func randomCurrency() string {
	return supportedCurrencies[gofakeit.Number(0, len(supportedCurrencies)-1)]
}

func createRandomAccount(t *testing.T) (Account, CreateAccountParams) {
	t.Helper()

	arg := CreateAccountParams{
		Owner:    gofakeit.Name(),
		Balance:  int64(gofakeit.Price(0, 10000)),
		Currency: randomCurrency(),
	}

	acc, err := testQueries.CreateAccount(context.Background(), arg)

	if err != nil {
		t.Fatal("Cannot create account:", err)
	}

	return acc, arg
}

func deleteEntry(t *testing.T, entry_id int64) {
	t.Helper()

	tag, err := testQueries.db.Exec(
		context.Background(),
		"DELETE FROM entries WHERE id = $1",
		entry_id,
	)

	if err != nil {
		t.Fatal("Cannot delete entry:", err)
	}
	if tag.RowsAffected() != 1 {
		t.Fatalf("expected to delete 1 row, deleted %d", tag.RowsAffected())
	}
}

func deleteTransfer(t *testing.T, transfer_id int64) {
	t.Helper()

	tag, err := testQueries.db.Exec(
		context.Background(),
		"DELETE FROM transfers WHERE id = $1",
		transfer_id,
	)

	if err != nil {
		t.Fatal("Cannot delete transfer:", err)
	}
	if tag.RowsAffected() != 1 {
		t.Fatalf("expected to delete 1 row, deleted %d", tag.RowsAffected())
	}
}
