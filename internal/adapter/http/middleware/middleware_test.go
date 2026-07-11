package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/quixiq/polyglot/internal/adapter/auth"
	"github.com/quixiq/polyglot/internal/adapter/http/middleware"
)

// MockRBACEnforcer implements port.Enforcer for testing
type MockRBACEnforcer struct {
	Allowed bool
	Err     error
}

func (m *MockRBACEnforcer) Enforce(ctx context.Context, role, resource, action string) (bool, error) {
	return m.Allowed, m.Err
}

func TestAuthRequired(t *testing.T) {
	gin.SetMode(gin.TestMode)

	secret := []byte("test-secret")
	jwtHandler := auth.NewJWT(secret, time.Hour)

	validToken, _ := jwtHandler.Issue(context.Background(), "user123", "alice", "admin")

	tests := []struct {
		name         string
		authHeader   string
		expectedCode int
	}{
		{"No Header", "", http.StatusUnauthorized},
		{"Invalid Format", "BearerTokenOnly", http.StatusUnauthorized},
		{"Invalid Token", "Bearer invalid.token.here", http.StatusUnauthorized},
		{"Valid Token", "Bearer " + validToken, http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			_, r := gin.CreateTestContext(w)

			r.Use(middleware.AuthRequired(jwtHandler))
			r.GET("/test", func(c *gin.Context) {
				// Assert context is set correctly
				role, _ := c.Get("role")
				assert.Equal(t, "admin", role)
				c.String(http.StatusOK, "success")
			})

			req, _ := http.NewRequest(http.MethodGet, "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
		})
	}
}

func TestRBACRequired(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		role         string
		allowed      bool
		setupContext func(c *gin.Context)
		expectedCode int
	}{
		{
			name:         "Allowed",
			allowed:      true,
			setupContext: func(c *gin.Context) { c.Set("role", "admin") },
			expectedCode: http.StatusOK,
		},
		{
			name:         "Denied",
			allowed:      false,
			setupContext: func(c *gin.Context) { c.Set("role", "guest") },
			expectedCode: http.StatusForbidden,
		},
		{
			name:         "Missing Role",
			allowed:      true,
			setupContext: func(c *gin.Context) {}, // Don't set role
			expectedCode: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enforcer := &MockRBACEnforcer{Allowed: tt.allowed}
			w := httptest.NewRecorder()
			_, r := gin.CreateTestContext(w)

			// Dummy middleware to simulate AuthRequired
			r.Use(func(c *gin.Context) {
				tt.setupContext(c)
				c.Next()
			})
			r.Use(middleware.RBACRequired(enforcer))
			r.GET("/test", func(c *gin.Context) {
				c.String(http.StatusOK, "success")
			})

			req, _ := http.NewRequest(http.MethodGet, "/test", nil)
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
		})
	}
}
