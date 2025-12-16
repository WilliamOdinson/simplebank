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
	ctx := context.Background()
	acc, arg := createRandomAccount(t)

	t.Cleanup(func() {
		_ = testQueries.DeleteAccount(ctx, acc.ID)
	})

	require.NotEmpty(t, acc)

	require.NotZero(t, acc.ID)
	require.NotZero(t, acc.CreatedAt)

	require.Equal(t, arg.Owner, acc.Owner)
	require.Equal(t, arg.Balance, acc.Balance)
	require.Equal(t, arg.Currency, acc.Currency)
}

func TestGetAccount(t *testing.T) {
	ctx := context.Background()
	acc1, _ := createRandomAccount(t)

	t.Cleanup(func() {
		_ = testQueries.DeleteAccount(ctx, acc1.ID)
	})

	acc2, err := testQueries.GetAccount(ctx, acc1.ID)
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
}

func TestUpdateAccount(t *testing.T) {
	ctx := context.Background()
	acc1, arg1 := createRandomAccount(t)

	t.Cleanup(func() {
		_ = testQueries.DeleteAccount(ctx, acc1.ID)
	})

	arg2 := UpdateAccountParams{
		ID:      acc1.ID,
		Balance: arg1.Balance + int64(gofakeit.Price(1, 10000)),
	}
	acc2, err := testQueries.UpdateAccount(ctx, arg2)
	require.NoError(t, err)
	require.NotEmpty(t, acc2)
	require.NotEqual(t, acc1.Balance, acc2.Balance)
	require.Equal(t, arg2.Balance, acc2.Balance)
}

func TestDeleteAccount(t *testing.T) {
	ctx := context.Background()
	acc, _ := createRandomAccount(t)

	err := testQueries.DeleteAccount(ctx, acc.ID)
	require.NoError(t, err)

	_, err = testQueries.GetAccount(ctx, acc.ID)
	require.EqualError(t, err, pgx.ErrNoRows.Error())
}

func TestListAccounts(t *testing.T) {
	ctx := context.Background()
	var createdIDs []int64

	for i := 0; i < 10; i++ {
		acc, _ := createRandomAccount(t)
		createdIDs = append(createdIDs, acc.ID)
	}

	t.Cleanup(func() {
		for _, id := range createdIDs {
			_ = testQueries.DeleteAccount(ctx, id)
		}
	})

	arg := ListAccountsParams{
		Limit:  5,
		Offset: 5,
	}

	accounts, err := testQueries.ListAccounts(ctx, arg)
	require.NoError(t, err)
	require.Len(t, accounts, 5)

	for _, acc := range accounts {
		require.NotEmpty(t, acc)
	}
}

func TestChangeAccountBalance(t *testing.T) {
	ctx := context.Background()
	acc1, _ := createRandomAccount(t)

	t.Cleanup(func() {
		_ = testQueries.DeleteAccount(ctx, acc1.ID)
	})

	amount := int64(gofakeit.Number(1, 1000))
	arg := ChangeAccountBalanceParams{
		ID:     acc1.ID,
		Amount: amount,
	}

	acc2, err := testQueries.ChangeAccountBalance(ctx, arg)
	require.NoError(t, err)
	require.NotEmpty(t, acc2)
	require.Equal(t, acc1.Balance+amount, acc2.Balance)
}

func TestGetAccountForUpdate(t *testing.T) {
	ctx := context.Background()
	acc1, _ := createRandomAccount(t)

	t.Cleanup(func() {
		_ = testQueries.DeleteAccount(ctx, acc1.ID)
	})

	acc2, err := testQueries.GetAccountForUpdate(ctx, acc1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, acc2)
	require.Equal(t, acc1.ID, acc2.ID)
	require.Equal(t, acc1.Owner, acc2.Owner)
	require.Equal(t, acc1.Balance, acc2.Balance)
	require.Equal(t, acc1.Currency, acc2.Currency)
}

func TestCreateAccountNegativeBalanceConstraint(t *testing.T) {
	ctx := context.Background()

	arg := CreateAccountParams{
		Owner:    gofakeit.Name(),
		Balance:  -100,
		Currency: gofakeit.CurrencyShort(),
	}

	_, err := testQueries.CreateAccount(ctx, arg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "balance_non_negative")
}

func TestUpdateAccountNegativeBalanceConstraint(t *testing.T) {
	ctx := context.Background()
	acc, _ := createRandomAccount(t)

	t.Cleanup(func() {
		_ = testQueries.DeleteAccount(ctx, acc.ID)
	})

	arg := UpdateAccountParams{
		ID:      acc.ID,
		Balance: -100,
	}

	_, err := testQueries.UpdateAccount(ctx, arg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "balance_non_negative")
}

func TestChangeAccountBalanceNegativeConstraint(t *testing.T) {
	ctx := context.Background()

	arg := CreateAccountParams{
		Owner:    gofakeit.Name(),
		Balance:  100,
		Currency: "USD",
	}
	acc, err := testQueries.CreateAccount(ctx, arg)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = testQueries.DeleteAccount(ctx, acc.ID)
	})

	// attempt to decrease balance by more than current balance
	_, err = testQueries.ChangeAccountBalance(ctx, ChangeAccountBalanceParams{
		ID:     acc.ID,
		Amount: -200,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "balance_non_negative")
}
