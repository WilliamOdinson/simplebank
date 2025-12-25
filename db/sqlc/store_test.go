package db

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
)

func TestTransferTx(t *testing.T) {
	ctx := context.Background()
	store := NewStore(testPool)

	// create users first
	user1, _ := createRandomUser(t)
	user2, _ := createRandomUser(t)

	// create accounts with sufficient balance for transfers
	acc1Arg := CreateAccountParams{
		Owner:    user1.Username,
		Balance:  10000,
		Currency: randomCurrency(),
	}
	account1, err := testQueries.CreateAccount(ctx, acc1Arg)
	require.NoError(t, err)

	acc2Arg := CreateAccountParams{
		Owner:    user2.Username,
		Balance:  10000,
		Currency: randomCurrency(),
	}
	account2, err := testQueries.CreateAccount(ctx, acc2Arg)
	require.NoError(t, err)

	n := 5

	amount := int64(gofakeit.Number(1, 100))

	errs := make(chan error)
	results := make(chan TransferTxResult)

	existed := make(map[int]bool)

	// collect all created transfer and entry IDs for cleanup.
	var transferIDs []int64
	var entryIDs []int64

	for i := 0; i < n; i++ {
		go func() {
			result, err := store.TransferTx(ctx, TransferTxParams{
				FromAccountID: account1.ID,
				ToAccountID:   account2.ID,
				Amount:        amount,
			})

			if err != nil {
				t.Errorf("TransferTx failed: %v", err)
			}

			errs <- err
			results <- result
		}()
	}

	// check results
	for i := 0; i < n; i++ {
		err := <-errs
		if err != nil {
			t.Errorf("received error from goroutine: %v", err)
		}

		require.NoError(t, err)

		result := <-results
		require.NotEmpty(t, result)

		// collect IDs for cleanup
		transferIDs = append(transferIDs, result.Transfer.ID)
		entryIDs = append(entryIDs, result.FromEntry.ID, result.ToEntry.ID)

		// check transfer
		transfer := result.Transfer
		require.NotEmpty(t, transfer)
		require.Equal(t, account1.ID, transfer.FromAccountID)
		require.Equal(t, account2.ID, transfer.ToAccountID)
		require.Equal(t, amount, transfer.Amount)
		require.NotZero(t, transfer.ID)
		require.NotZero(t, transfer.CreatedAt)

		_, err = store.GetTransfer(ctx, transfer.ID)
		require.NoError(t, err)

		// check from entry
		fromEntry := result.FromEntry
		require.NotEmpty(t, fromEntry)
		require.Equal(t, account1.ID, fromEntry.AccountID)
		require.Equal(t, -amount, fromEntry.Amount)
		require.NotZero(t, fromEntry.ID)
		require.NotZero(t, fromEntry.CreatedAt)

		_, err = store.GetEntry(ctx, fromEntry.ID)
		require.NoError(t, err)

		// check to entry
		toEntry := result.ToEntry
		require.NotEmpty(t, toEntry)
		require.Equal(t, account2.ID, toEntry.AccountID)
		require.Equal(t, amount, toEntry.Amount)
		require.NotZero(t, toEntry.ID)
		require.NotZero(t, toEntry.CreatedAt)

		_, err = store.GetEntry(ctx, toEntry.ID)
		require.NoError(t, err)

		// check accounts
		fromAccount := result.FromAccount
		require.NotEmpty(t, fromAccount)
		require.Equal(t, account1.ID, fromAccount.ID)

		toAccount := result.ToAccount
		require.NotEmpty(t, toAccount)
		require.Equal(t, account2.ID, toAccount.ID)

		// check accounts' balance
		diff1 := account1.Balance - fromAccount.Balance
		diff2 := toAccount.Balance - account2.Balance
		require.Equal(t, diff1, diff2)
		require.True(t, diff1 > 0)
		require.True(t, diff1%amount == 0) // multiple of amount

		// check how many times the amount has been transferred
		k := int(diff1 / amount)
		require.True(t, k >= 1 && k <= n)
		require.NotContains(t, existed, k)
		existed[k] = true
	}

	// check the final updated balance
	updatedAccount1, err := store.GetAccount(ctx, account1.ID)
	require.NoError(t, err)

	updatedAccount2, err := store.GetAccount(ctx, account2.ID)
	require.NoError(t, err)

	require.Equal(t, account1.Balance-int64(n)*amount, updatedAccount1.Balance)
	require.Equal(t, account2.Balance+int64(n)*amount, updatedAccount2.Balance)

	// cleanup
	t.Cleanup(func() {
		for _, id := range transferIDs {
			deleteTransfer(t, id)
		}
		for _, id := range entryIDs {
			deleteEntry(t, id)
		}
		_ = testQueries.DeleteAccount(ctx, account2.ID)
		_ = testQueries.DeleteAccount(ctx, account1.ID)
		deleteUser(t, user2.Username)
		deleteUser(t, user1.Username)
	})
}

func TestTransferTxSameAccountConstraint(t *testing.T) {
	ctx := context.Background()
	store := NewStore(testPool)

	acc, _ := createRandomAccount(t)

	t.Cleanup(func() {
		_ = testQueries.DeleteAccount(ctx, acc.ID)
		deleteUser(t, acc.Owner)
	})

	// transfer to the same account should fail
	_, err := store.TransferTx(ctx, TransferTxParams{
		FromAccountID: acc.ID,
		ToAccountID:   acc.ID,
		Amount:        100,
	})
	require.Error(t, err)

	var pgErr *pgconn.PgError
	require.ErrorAs(t, err, &pgErr)
	require.Equal(t, "23514", pgErr.Code) // check_violation
}

func TestTransferTxNegativeAmountConstraint(t *testing.T) {
	ctx := context.Background()
	store := NewStore(testPool)

	acc1, _ := createRandomAccount(t)
	acc2, _ := createRandomAccount(t)

	t.Cleanup(func() {
		_ = testQueries.DeleteAccount(ctx, acc2.ID)
		_ = testQueries.DeleteAccount(ctx, acc1.ID)
		deleteUser(t, acc2.Owner)
		deleteUser(t, acc1.Owner)
	})

	// negative amount should fail
	_, err := store.TransferTx(ctx, TransferTxParams{
		FromAccountID: acc1.ID,
		ToAccountID:   acc2.ID,
		Amount:        -100,
	})
	require.Error(t, err)

	var pgErr *pgconn.PgError
	require.ErrorAs(t, err, &pgErr)
	require.Equal(t, "23514", pgErr.Code) // check_violation
}

func TestTransferTxZeroAmountConstraint(t *testing.T) {
	ctx := context.Background()
	store := NewStore(testPool)

	acc1, _ := createRandomAccount(t)
	acc2, _ := createRandomAccount(t)

	t.Cleanup(func() {
		_ = testQueries.DeleteAccount(ctx, acc2.ID)
		_ = testQueries.DeleteAccount(ctx, acc1.ID)
		deleteUser(t, acc2.Owner)
		deleteUser(t, acc1.Owner)
	})

	// zero amount should fail
	_, err := store.TransferTx(ctx, TransferTxParams{
		FromAccountID: acc1.ID,
		ToAccountID:   acc2.ID,
		Amount:        0,
	})
	require.Error(t, err)

	var pgErr *pgconn.PgError
	require.ErrorAs(t, err, &pgErr)
	require.Equal(t, "23514", pgErr.Code) // check_violation
}

func TestTransferTxInsufficientBalanceConstraint(t *testing.T) {
	ctx := context.Background()
	store := NewStore(testPool)

	// create user and account with small balance
	user1, _ := createRandomUser(t)
	acc1Arg := CreateAccountParams{
		Owner:    user1.Username,
		Balance:  100,
		Currency: randomCurrency(),
	}
	acc1, err := testQueries.CreateAccount(ctx, acc1Arg)
	require.NoError(t, err)

	acc2, _ := createRandomAccount(t)

	t.Cleanup(func() {
		_ = testQueries.DeleteAccount(ctx, acc2.ID)
		_ = testQueries.DeleteAccount(ctx, acc1.ID)
		deleteUser(t, acc2.Owner)
		deleteUser(t, user1.Username)
	})

	// transfer more than balance should fail due to balance_non_negative constraint
	_, err = store.TransferTx(ctx, TransferTxParams{
		FromAccountID: acc1.ID,
		ToAccountID:   acc2.ID,
		Amount:        200, // more than acc1's balance
	})
	require.Error(t, err)

	var pgErr *pgconn.PgError
	require.ErrorAs(t, err, &pgErr)
	require.Equal(t, "23514", pgErr.Code) // check_violation

	// verify acc1 balance is unchanged (transaction should have rolled back)
	acc1After, err := store.GetAccount(ctx, acc1.ID)
	require.NoError(t, err)
	require.Equal(t, acc1.Balance, acc1After.Balance)
}

func TestBilateralTransferTxDeadlock(t *testing.T) {
	ctx := context.Background()
	store := NewStore(testPool)

	// create users first
	user1, _ := createRandomUser(t)
	user2, _ := createRandomUser(t)

	// create accounts with sufficient balance
	acc1Arg := CreateAccountParams{
		Owner:    user1.Username,
		Balance:  10000,
		Currency: randomCurrency(),
	}
	account1, err := testQueries.CreateAccount(ctx, acc1Arg)
	require.NoError(t, err)

	acc2Arg := CreateAccountParams{
		Owner:    user2.Username,
		Balance:  10000,
		Currency: randomCurrency(),
	}
	account2, err := testQueries.CreateAccount(ctx, acc2Arg)
	require.NoError(t, err)

	n := 10

	amount := int64(gofakeit.Number(1, 100))

	errs := make(chan error)
	results := make(chan TransferTxResult)

	// collect all created transfer and entry IDs for cleanup.
	var transferIDs []int64
	var entryIDs []int64

	for i := 0; i < n; i++ {
		fromAccountID := account1.ID
		toAccountID := account2.ID

		if i%2 == 0 {
			fromAccountID = account2.ID
			toAccountID = account1.ID
		}

		go func() {
			result, err := store.TransferTx(ctx, TransferTxParams{
				FromAccountID: fromAccountID,
				ToAccountID:   toAccountID,
				Amount:        amount,
			})

			if err != nil {
				t.Errorf("TransferTx failed: %v", err)
			}

			errs <- err
			results <- result
		}()
	}

	// check results
	for i := 0; i < n; i++ {
		err := <-errs
		if err != nil {
			t.Errorf("received error from goroutine: %v", err)
		}

		require.NoError(t, err)

		result := <-results

		// collect IDs for cleanup
		transferIDs = append(transferIDs, result.Transfer.ID)
		entryIDs = append(entryIDs, result.FromEntry.ID, result.ToEntry.ID)
	}

	// check the final updated balance
	updatedAccount1, err := store.GetAccount(ctx, account1.ID)
	require.NoError(t, err)

	updatedAccount2, err := store.GetAccount(ctx, account2.ID)
	require.NoError(t, err)

	require.Equal(t, account1.Balance, updatedAccount1.Balance)
	require.Equal(t, account2.Balance, updatedAccount2.Balance)

	// cleanup
	t.Cleanup(func() {
		for _, id := range transferIDs {
			deleteTransfer(t, id)
		}
		for _, id := range entryIDs {
			deleteEntry(t, id)
		}
		_ = testQueries.DeleteAccount(ctx, account2.ID)
		_ = testQueries.DeleteAccount(ctx, account1.ID)
		deleteUser(t, user2.Username)
		deleteUser(t, user1.Username)
	})
}
