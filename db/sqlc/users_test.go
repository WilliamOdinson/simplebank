package db

import (
	"context"
	"testing"
	"time"

	"github.com/WilliamOdinson/simplebank/util"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func TestCreateUser(t *testing.T) {
	user, arg := createRandomUser(t)
	t.Cleanup(func() {
		deleteUser(t, user.Username)
	})

	require.NotEmpty(t, user)
	require.Equal(t, arg.Username, user.Username)
	require.Equal(t, arg.HashedPassword, user.HashedPassword)
	require.Equal(t, arg.FullName, user.FullName)
	require.Equal(t, arg.Email, user.Email)
	require.True(t, user.PasswordChangedAt.Time.IsZero())
	require.NotZero(t, user.CreatedAt)
}

func TestGetUser(t *testing.T) {
	ctx := context.Background()
	user1, _ := createRandomUser(t)
	t.Cleanup(func() {
		deleteUser(t, user1.Username)
	})

	user2, err := testQueries.GetUser(ctx, user1.Username)
	require.NoError(t, err)
	require.NotEmpty(t, user2)
	require.Equal(t, user1.Username, user2.Username)
	require.Equal(t, user1.HashedPassword, user2.HashedPassword)
	require.Equal(t, user1.FullName, user2.FullName)
	require.Equal(t, user1.Email, user2.Email)
	require.WithinDuration(t, user1.PasswordChangedAt.Time, user2.PasswordChangedAt.Time, time.Second)
	require.WithinDuration(t, user1.CreatedAt.Time, user2.CreatedAt.Time, time.Second)
}

func TestGetUserNotFound(t *testing.T) {
	ctx := context.Background()
	_, err := testQueries.GetUser(ctx, "nonexistent_user")
	require.EqualError(t, err, pgx.ErrNoRows.Error())
}

func TestCreateUserDuplicateUsername(t *testing.T) {
	ctx := context.Background()
	user1, _ := createRandomUser(t)
	t.Cleanup(func() {
		deleteUser(t, user1.Username)
	})

	hashedPassword, err := util.HashPassword(gofakeit.Password(true, true, true, false, false, 8))
	require.NoError(t, err)

	arg := CreateUserParams{
		Username:       user1.Username,
		HashedPassword: hashedPassword,
		FullName:       gofakeit.Name(),
		Email:          gofakeit.Email(),
	}
	_, err = testQueries.CreateUser(ctx, arg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "duplicate key")
}

func TestCreateUserDuplicateEmail(t *testing.T) {
	ctx := context.Background()
	user1, _ := createRandomUser(t)
	t.Cleanup(func() {
		deleteUser(t, user1.Username)
	})

	hashedPassword, err := util.HashPassword(gofakeit.Password(true, true, true, false, false, 8))
	require.NoError(t, err)

	arg := CreateUserParams{
		Username:       gofakeit.LetterN(10),
		HashedPassword: hashedPassword,
		FullName:       gofakeit.Name(),
		Email:          user1.Email,
	}
	_, err = testQueries.CreateUser(ctx, arg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "duplicate key")
}
