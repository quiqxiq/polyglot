package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quixiq/polyglot/internal/adapter/auth"
)

func TestJWTHandler(t *testing.T) {
	secret := []byte("super-secret-key")
	expiry := time.Hour

	jwtHandler := auth.NewJWT(secret, expiry)
	ctx := context.Background()

	t.Run("Issue and Validate Success", func(t *testing.T) {
		token, err := jwtHandler.Issue(ctx, "user-123", "johndoe", "superadmin")
		require.NoError(t, err)
		require.NotEmpty(t, token)

		claims, err := jwtHandler.Validate(ctx, token)
		require.NoError(t, err)
		require.NotNil(t, claims)

		assert.Equal(t, "user-123", claims.UserID)
		assert.Equal(t, "johndoe", claims.Username)
		assert.Equal(t, "superadmin", claims.Role)
	})

	t.Run("Validate Expired Token", func(t *testing.T) {
		shortExpiryHandler := auth.NewJWT(secret, -time.Hour) // explicitly expired
		token, err := shortExpiryHandler.Issue(ctx, "user-123", "johndoe", "superadmin")
		require.NoError(t, err)

		claims, err := jwtHandler.Validate(ctx, token)
		require.Error(t, err)
		require.Nil(t, claims)
		assert.ErrorIs(t, err, auth.ErrInvalidToken)
	})

	t.Run("Validate Invalid Signature", func(t *testing.T) {
		token, err := jwtHandler.Issue(ctx, "user-123", "johndoe", "superadmin")
		require.NoError(t, err)

		// Modify token signature (last characters)
		invalidToken := token[:len(token)-5] + "12345"

		claims, err := jwtHandler.Validate(ctx, invalidToken)
		require.Error(t, err)
		require.Nil(t, claims)
		assert.ErrorIs(t, err, auth.ErrInvalidToken)
	})

	t.Run("Validate Malformed Token", func(t *testing.T) {
		claims, err := jwtHandler.Validate(ctx, "not.a.valid.token")
		require.Error(t, err)
		require.Nil(t, claims)
		assert.ErrorIs(t, err, auth.ErrInvalidToken)
	})
}
