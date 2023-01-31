package util

import (
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"testing"
)

func TestHashPassword(t *testing.T) {
	password := RandomString(6)
	hashedPassword1, err := HashPassword(password)
	require.NoError(t, err)
	require.NotNil(t, hashedPassword1)

	err = CheckHashedPassword(hashedPassword1, password)
	require.NoError(t, err)

	wrongPassword := RandomString(7)
	err = CheckHashedPassword(hashedPassword1, wrongPassword)
	require.EqualError(t, err, bcrypt.ErrMismatchedHashAndPassword.Error())

	hashedPassword2, err := HashPassword(password)
	require.NotEqual(t, hashedPassword1, hashedPassword2)
}
