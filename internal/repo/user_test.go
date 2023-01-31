package repo

import (
	"context"
	"github.com/max-rodziyevsky/go-simple-bank/util"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestQueries_CreateUser(t *testing.T) {
	createRandomUser(t)
}

func TestQueries_GetUser(t *testing.T) {
	randomUser := createRandomUser(t)

	user, err := testQueries.GetUser(context.Background(), randomUser.Username)
	require.NoError(t, err)
	require.NotNil(t, user)

	require.Equal(t, randomUser.Username, user.Username)
	require.Equal(t, randomUser.FullName, user.FullName)
	require.Equal(t, randomUser.Email, user.Email)
	require.Equal(t, randomUser.HashPassword, user.HashPassword)

	require.WithinDuration(t, randomUser.ChangePasswordAt, user.ChangePasswordAt, time.Second)
	require.WithinDuration(t, randomUser.CreatedAt, user.CreatedAt, time.Second)
}

func createRandomUser(t *testing.T) User {
	hashedPassword, err := util.HashPassword(util.RandomString(6))
	require.NoError(t, err)

	arg := CreateUserParams{
		Username:     util.RandomOwner(),
		FullName:     util.RandomOwner(),
		Email:        util.RandomEmail(),
		HashPassword: hashedPassword,
	}

	user, err := testQueries.CreateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotNil(t, user)

	require.Equal(t, arg.Username, user.Username)
	require.Equal(t, arg.FullName, user.FullName)
	require.Equal(t, arg.Email, user.Email)
	require.Equal(t, arg.HashPassword, user.HashPassword)

	// when user created user must zero value on change password at time
	require.True(t, user.ChangePasswordAt.IsZero())
	require.NotZero(t, user.CreatedAt)

	return user
}
