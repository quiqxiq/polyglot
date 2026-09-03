package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/quixiq/polyglot/pkg/logger"
	"github.com/quixiq/polyglot/pkg/response"
)

// Recovery returns standard net/http middleware that recovers from panics and logs with Logrus.
func Recovery() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					stack := string(debug.Stack())
					logger.WithComponent("HTTP").WithFields(map[string]any{"stack": stack, "panic": rec}).Error("panic recovered")
					response.WriteHTTPStatusError(w, http.StatusInternalServerError, "internal server error")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
