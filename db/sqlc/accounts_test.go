package db

import (
	"context"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func TestCreateAccount(t *testing.T) {
	acc, arg := createRandomAccount(t)
	require.NotEmpty(t, acc)

	require.NotZero(t, acc.ID)
	require.NotZero(t, acc.CreatedAt)

	require.Equal(t, arg.Owner, acc.Owner)
	require.Equal(t, arg.Balance, acc.Balance)
	require.Equal(t, arg.Currency, acc.Currency)

	t.Cleanup(func() {
		_ = testQueries.DeleteAccount(context.Background(), acc.ID)
	})
}

func TestGetAccount(t *testing.T) {
	acc1, _ := createRandomAccount(t)

	acc2, err := testQueries.GetAccount(context.Background(), acc1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, acc2)

	require.Equal(t, acc1.ID, acc2.ID)
	require.Equal(t, acc1.Owner, acc2.Owner)
	require.Equal(t, acc1.Balance, acc2.Balance)
	require.Equal(t, acc1.Currency, acc2.Currency)

	require.True(t, acc1.CreatedAt.Valid)
	require.True(t, acc2.CreatedAt.Valid)

	t1 := acc1.CreatedAt.Time.UTC()
	t2 := acc2.CreatedAt.Time.UTC()
	require.WithinDuration(t, t1, t2, time.Second)

	t.Cleanup(func() {
		_ = testQueries.DeleteAccount(context.Background(), acc2.ID)
	})
}

func TestUpdateAccount(t *testing.T) {
	acc1, arg1 := createRandomAccount(t)

	arg2 := UpdateAccountParams{
		ID:      acc1.ID,
		Balance: arg1.Balance + int64(gofakeit.Price(1, 10000)),
	}

	acc2, err := testQueries.UpdateAccount(context.Background(), arg2)

	require.NoError(t, err)
	require.NotEmpty(t, acc2)
	require.NotEqual(t, acc1.Balance, acc2.Balance)
	require.Equal(t, arg2.Balance, acc2.Balance)

	t.Cleanup(func() {
		_ = testQueries.DeleteAccount(context.Background(), acc2.ID)
	})
}

func TestDeleteAccount(t *testing.T) {
	acc, _ := createRandomAccount(t)

	err := testQueries.DeleteAccount(context.Background(), acc.ID)
	require.NoError(t, err)

	_, err = testQueries.GetAccount(context.Background(), acc.ID)
	require.EqualError(t, err, pgx.ErrNoRows.Error())
}

func TestListAccounts(t *testing.T) {
	for i := 0; i < 10; i++ {
		createRandomAccount(t)
	}

	arg := ListAccountsParams{
		Limit:  5,
		Offset: 5,
	}

	accounts, err := testQueries.ListAccounts(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, accounts, 5)

	for _, acc := range accounts {
		require.NotEmpty(t, acc)
	}
}
