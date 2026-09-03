package middleware

import (
	"net/http"
	"strings"

	"github.com/quixiq/polyglot/internal/adapter/auth"
	"github.com/quixiq/polyglot/pkg/response"
)

const BearerScheme = "Bearer"

// AuthenticateJWT returns a standard HTTP middleware that verifies JWT bearer tokens.
func AuthenticateJWT(jwtService *auth.JWTService) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.WriteHTTPStatusError(w, http.StatusUnauthorized, "Authorization header missing")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], BearerScheme) {
				response.WriteHTTPStatusError(w, http.StatusUnauthorized, "Invalid Authorization header format. Expected 'Bearer <token>'")
				return
			}

			tokenStr := parts[1]
			claims, err := jwtService.ValidateToken(tokenStr)
			if err != nil {
				response.WriteHTTPStatusError(w, http.StatusUnauthorized, "Invalid or expired authentication token")
				return
			}

			roles := claims.Roles
			if len(roles) == 0 && claims.Role != "" {
				roles = []string{claims.Role}
			}

			ctx := auth.WithIdentity(r.Context(), claims.UserID, roles)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
