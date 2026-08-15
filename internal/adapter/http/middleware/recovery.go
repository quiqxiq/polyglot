package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/quixiq/polyglot/pkg/logger"
	"github.com/quixiq/polyglot/pkg/response"
)

// Recoverer recovers from panics and logs the stack trace using Logrus.
func Recoverer() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rvr := recover(); rvr != nil {
					if rvr == http.ErrAbortHandler {
						panic(rvr)
					}

					stack := string(debug.Stack())
					logger.FromContext(r.Context()).WithFields(logger.Fields{
						"panic":  rvr,
						"stack":  stack,
						"path":   r.URL.Path,
						"method": r.Method,
					}).Error("Panic recovered in HTTP handler")

					response.Fail(w, http.StatusInternalServerError, "INTERNAL_ERROR",
						"Internal server error", fmt.Sprintf("%v", rvr))
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
