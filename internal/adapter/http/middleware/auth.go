package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/quixiq/polyglot/internal/adapter/auth"
	"github.com/quixiq/polyglot/pkg/response"
)

type contextKey string

const (
	UserIDKey    contextKey = "user_id"
	UserEmailKey contextKey = "user_email"
	UserRoleKey  contextKey = "user_role"
	TenantIDKey  contextKey = "tenant_id"
)

// AuthenticateJWT validates the JWT bearer token and injects user claims into request context.
func AuthenticateJWT(jwtService *auth.JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.Fail(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authorization header missing", "")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				response.Fail(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid Authorization header format. Expected 'Bearer <token>'", "")
				return
			}

			tokenStr := parts[1]
			claims, err := jwtService.ValidateToken(tokenStr)
			if err != nil {
				response.Fail(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid or expired authentication token", "")
				return
			}

			ctx := r.Context()
			ctx = context.WithValue(ctx, UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, UserEmailKey, claims.Email)
			ctx = context.WithValue(ctx, UserRoleKey, claims.Role)
			ctx = context.WithValue(ctx, TenantIDKey, claims.TenantID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserRoleFromContext extracts user role from context.
func UserRoleFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(UserRoleKey).(string); ok {
		return v
	}
	return ""
}

// TenantIDFromContext extracts tenant ID from context.
func TenantIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(TenantIDKey).(string); ok {
		return v
	}
	return ""
}
