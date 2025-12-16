package db

import (
	"context"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/require"
)

func TestCreateEntry(t *testing.T) {
	ctx := context.Background()

	acc, _ := createRandomAccount(t)

	arg := CreateEntryParams{
		AccountID: acc.ID,
		Amount:    int64(gofakeit.Price(-10000, 10000)),
	}

	entry, err := testQueries.CreateEntry(ctx, arg)
	require.NoError(t, err)
	require.NotEmpty(t, entry)

	require.NotZero(t, entry.ID)
	require.True(t, entry.CreatedAt.Valid)

	require.Equal(t, arg.AccountID, entry.AccountID)
	require.Equal(t, arg.Amount, entry.Amount)

	t.Cleanup(func() {
		deleteEntry(t, entry.ID)
		_ = testQueries.DeleteAccount(ctx, acc.ID)
	})
}

func TestGetEntry(t *testing.T) {
	ctx := context.Background()

	acc, _ := createRandomAccount(t)

	ent1, err := testQueries.CreateEntry(ctx, CreateEntryParams{
		AccountID: acc.ID,
		Amount:    int64(gofakeit.Price(-10000, 10000)),
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		deleteEntry(t, ent1.ID)
		_ = testQueries.DeleteAccount(ctx, acc.ID)
	})

	ent2, err := testQueries.GetEntry(ctx, ent1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, ent2)

	require.Equal(t, ent1.ID, ent2.ID)
	require.Equal(t, ent1.AccountID, ent2.AccountID)
	require.Equal(t, ent1.Amount, ent2.Amount)

	require.True(t, ent1.CreatedAt.Valid)
	require.True(t, ent2.CreatedAt.Valid)

	t1 := ent1.CreatedAt.Time.UTC()
	t2 := ent2.CreatedAt.Time.UTC()
	require.WithinDuration(t, t1, t2, time.Second)
}

func TestListEntries(t *testing.T) {
	ctx := context.Background()

	acc, _ := createRandomAccount(t)

	var createdIDs []int64
	for i := 0; i < 10; i++ {
		ent, err := testQueries.CreateEntry(ctx, CreateEntryParams{
			AccountID: acc.ID,
			Amount:    int64(gofakeit.Price(-10000, 10000)),
		})
		require.NoError(t, err)
		createdIDs = append(createdIDs, ent.ID)
	}

	t.Cleanup(func() {
		for _, id := range createdIDs {
			deleteEntry(t, id)
		}
		_ = testQueries.DeleteAccount(ctx, acc.ID)
	})

	arg := ListEntriesParams{
		AccountID: acc.ID,
		Limit:     5,
		Offset:    5,
	}

	entries, err := testQueries.ListEntries(ctx, arg)
	require.NoError(t, err)
	require.Len(t, entries, 5)

	for _, e := range entries {
		require.NotEmpty(t, e)
		require.Equal(t, acc.ID, e.AccountID)
	}
}
