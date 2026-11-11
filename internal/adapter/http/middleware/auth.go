package middleware

import (
	"crypto/subtle"
	"net"
	"net/http"
	"strings"

	"github.com/quixiq/polyglot/internal/adapter/auth"
	"github.com/quixiq/polyglot/pkg/response"
)

const BearerScheme = "Bearer"

// AuthenticateJWTOrServiceToken authenticates a loopback service caller with a
// static bearer token on the listed paths, and falls back to JWT authentication
// everywhere else. The service path list keeps the token from reaching any
// protected route beyond the ones the harness needs.
func AuthenticateJWTOrServiceToken(jwtService *auth.JWTService, serviceToken string, servicePaths []string) Middleware {
	token := strings.TrimSpace(serviceToken)
	paths := make(map[string]struct{}, len(servicePaths))
	for _, path := range servicePaths {
		paths[path] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if token != "" {
				if _, ok := paths[r.URL.Path]; ok && isLoopbackRequest(r) {
					if bearer, ok := bearerToken(r); ok && subtle.ConstantTimeCompare([]byte(bearer), []byte(token)) == 1 {
						ctx := auth.WithIdentity(r.Context(), auth.ServiceUserID, []string{auth.RoleService})
						next.ServeHTTP(w, r.WithContext(ctx))
						return
					}
				}
			}
			authenticateJWT(jwtService)(next).ServeHTTP(w, r)
		})
	}
}

func authenticateJWT(jwtService *auth.JWTService) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.WriteHTTPStatusError(w, http.StatusUnauthorized, "Authorization header missing")
				return
			}

			tokenStr, ok := bearerToken(r)
			if !ok {
				response.WriteHTTPStatusError(w, http.StatusUnauthorized, "Invalid Authorization header format. Expected 'Bearer <token>'")
				return
			}

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

// bearerToken extracts the credential from a `Bearer <token>` header.
func bearerToken(r *http.Request) (string, bool) {
	parts := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], BearerScheme) {
		return "", false
	}
	return parts[1], true
}

// isLoopbackRequest reports whether the direct peer is the local host. It reads
// RemoteAddr only: a forwarded address would be caller-controlled.
func isLoopbackRequest(r *http.Request) bool {
	host := r.RemoteAddr
	if splitHost, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		host = splitHost
	}
	ip := net.ParseIP(strings.Trim(host, "[]"))
	return ip != nil && ip.IsLoopback()
}
