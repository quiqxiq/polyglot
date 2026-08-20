package middleware

import (
	"net/http"
	"strings"
)

// CORS returns standard net/http middleware for Cross-Origin Resource Sharing.
func CORS(allowedOrigins []string, appEnv string) Middleware {
	isDev := strings.ToLower(appEnv) == "development"

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			allow := false

			if origin != "" {
				for _, o := range allowedOrigins {
					if (o == "*" || o == origin) && (isDev || (!strings.Contains(origin, "localhost") && !strings.Contains(origin, "127.0.0.1"))) {
						allow = true
						break
					}
				}
			}

			if allow {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			} else if origin != "" && isDev {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			} else if len(allowedOrigins) > 0 {
				w.Header().Set("Access-Control-Allow-Origin", allowedOrigins[0])
			}

			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, Connect-Protocol-Version, Connect-Content-Type, Connect-Timeout-Ms")
			w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
