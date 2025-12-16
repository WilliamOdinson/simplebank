package db

import (
	"context"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/require"
)

func TestCreateTransfer(t *testing.T) {
	ctx := context.Background()
	fromAcc, _ := createRandomAccount(t)
	toAcc, _ := createRandomAccount(t)

	arg := CreateTransferParams{
		FromAccountID: fromAcc.ID,
		ToAccountID:   toAcc.ID,
		Amount:        int64(gofakeit.Price(1, 10000)),
	}

	trs, err := testQueries.CreateTransfer(ctx, arg)
	require.NoError(t, err)
	require.NotEmpty(t, trs)
	require.NotZero(t, trs.ID)
	require.True(t, trs.CreatedAt.Valid)
	require.Equal(t, arg.FromAccountID, trs.FromAccountID)
	require.Equal(t, arg.ToAccountID, trs.ToAccountID)
	require.Equal(t, arg.Amount, trs.Amount)

	t.Cleanup(func() {
		deleteTransfer(t, trs.ID)
		_ = testQueries.DeleteAccount(ctx, toAcc.ID)
		_ = testQueries.DeleteAccount(ctx, fromAcc.ID)
	})
}

func TestGetTransfer(t *testing.T) {
	ctx := context.Background()
	fromAcc, _ := createRandomAccount(t)
	toAcc, _ := createRandomAccount(t)

	trs1, err := testQueries.CreateTransfer(ctx, CreateTransferParams{
		FromAccountID: fromAcc.ID,
		ToAccountID:   toAcc.ID,
		Amount:        int64(gofakeit.Price(1, 10000)),
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		deleteTransfer(t, trs1.ID)
		_ = testQueries.DeleteAccount(ctx, toAcc.ID)
		_ = testQueries.DeleteAccount(ctx, fromAcc.ID)
	})

	trs2, err := testQueries.GetTransfer(ctx, trs1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, trs2)
	require.Equal(t, trs1.ID, trs2.ID)
	require.Equal(t, trs1.FromAccountID, trs2.FromAccountID)
	require.Equal(t, trs1.ToAccountID, trs2.ToAccountID)
	require.Equal(t, trs1.Amount, trs2.Amount)
	require.True(t, trs1.CreatedAt.Valid)
	require.True(t, trs2.CreatedAt.Valid)
	t1 := trs1.CreatedAt.Time.UTC()
	t2 := trs2.CreatedAt.Time.UTC()
	require.WithinDuration(t, t1, t2, time.Second)
}

func TestListTransfers(t *testing.T) {
	ctx := context.Background()
	acc1, _ := createRandomAccount(t)
	acc2, _ := createRandomAccount(t)

	var createdIDs []int64
	for i := 0; i < 10; i++ {
		fromID := acc1.ID
		toID := acc2.ID
		if i%2 == 1 {
			fromID, toID = acc2.ID, acc1.ID
		}
		trs, err := testQueries.CreateTransfer(ctx, CreateTransferParams{
			FromAccountID: fromID,
			ToAccountID:   toID,
			Amount:        int64(gofakeit.Price(1, 10000)),
		})
		require.NoError(t, err)
		createdIDs = append(createdIDs, trs.ID)
	}

	t.Cleanup(func() {
		for _, id := range createdIDs {
			deleteTransfer(t, id)
		}
		_ = testQueries.DeleteAccount(ctx, acc2.ID)
		_ = testQueries.DeleteAccount(ctx, acc1.ID)
	})

	arg := ListTransfersParams{
		FromAccountID: acc1.ID,
		ToAccountID:   acc1.ID,
		Limit:         5,
		Offset:        5,
	}

	transfers, err := testQueries.ListTransfers(ctx, arg)
	require.NoError(t, err)
	require.Len(t, transfers, 5)

	for _, tr := range transfers {
		require.NotEmpty(t, tr)
		require.True(t, tr.CreatedAt.Valid)
		require.True(t, tr.FromAccountID == acc1.ID || tr.ToAccountID == acc1.ID)
	}
}
