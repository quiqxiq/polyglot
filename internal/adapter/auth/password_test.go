package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashPassword_RoundTrip(t *testing.T) {
	hash, err := HashPassword("mysecretpassword")
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, "mysecretpassword", hash)

	require.NoError(t, CheckPassword("mysecretpassword", hash))
}

func TestCheckPassword_WrongPassword(t *testing.T) {
	hash, err := HashPassword("correctpassword")
	require.NoError(t, err)

	err = CheckPassword("wrongpassword", hash)
	require.Error(t, err)
}

func TestHashPassword_DifferentHashes(t *testing.T) {
	hash1, err := HashPassword("samepassword")
	require.NoError(t, err)

	hash2, err := HashPassword("samepassword")
	require.NoError(t, err)

	// bcrypt generates different salts each time.
	assert.NotEqual(t, hash1, hash2)

	// Both should still verify correctly.
	require.NoError(t, CheckPassword("samepassword", hash1))
	require.NoError(t, CheckPassword("samepassword", hash2))
}
