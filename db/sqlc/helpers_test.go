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

func createRandomUser(t *testing.T) (User, CreateUserParams) {
	t.Helper()
	hashedPassword, err := util.HashPassword(gofakeit.Password(true, true, true, true, false, 16))
	if err != nil {
		t.Fatal("Cannot hash password:", err)
	}
	arg := CreateUserParams{
		Username:       gofakeit.Username(),
		HashedPassword: hashedPassword,
		FullName:       gofakeit.Name(),
		Email:          gofakeit.Email(),
	}
	user, err := testQueries.CreateUser(context.Background(), arg)
	if err != nil {
		t.Fatal("Cannot create user:", err)
	}
	return user, arg
}

func createRandomAccount(t *testing.T) (Account, CreateAccountParams) {
	t.Helper()
	user, _ := createRandomUser(t)
	arg := CreateAccountParams{
		Owner:    user.Username,
		Balance:  int64(gofakeit.Price(0, 10000)),
		Currency: randomCurrency(),
	}
	acc, err := testQueries.CreateAccount(context.Background(), arg)
	if err != nil {
		t.Fatal("Cannot create account:", err)
	}
	return acc, arg
}

func createRandomAccountForUser(t *testing.T, username string, currency string) (Account, CreateAccountParams) {
	t.Helper()
	arg := CreateAccountParams{
		Owner:    username,
		Balance:  int64(gofakeit.Price(0, 10000)),
		Currency: currency,
	}
	acc, err := testQueries.CreateAccount(context.Background(), arg)

	if err != nil {
		t.Fatal("Cannot create account:", err)
	}

	return acc, arg
}

func deleteUser(t *testing.T, username string) {
	t.Helper()
	tag, err := testQueries.db.Exec(
		context.Background(),
		"DELETE FROM users WHERE username = $1",
		username,
	)
	if err != nil {
		t.Fatal("Cannot delete user:", err)
	}
	if tag.RowsAffected() != 1 {
		t.Fatalf("expected to delete 1 row, deleted %d", tag.RowsAffected())
	}
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
