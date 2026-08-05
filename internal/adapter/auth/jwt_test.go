package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJWTService(t *testing.T) {
	jwtSvc := NewJWTService("test-secret-key-32-bytes-long!", 24)

	t.Run("Generate and Validate Token", func(t *testing.T) {
		token, err := jwtSvc.GenerateToken(1, "admin@gnet.co.id", "admin")
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		claims, err := jwtSvc.ValidateToken(token)
		assert.NoError(t, err)
		assert.Equal(t, uint(1), claims.UserID)
		assert.Equal(t, "admin@gnet.co.id", claims.Email)
		assert.Equal(t, "admin", claims.Role)
	})

	t.Run("Reject Invalid Token", func(t *testing.T) {
		_, err := jwtSvc.ValidateToken("invalid.token.string")
		assert.Error(t, err)
	})
}
