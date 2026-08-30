package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type requestIDContextKey struct{}

const requestIDHeader = "X-Request-ID"

// RequestID adds or preserves a bounded request correlation ID.
func RequestID() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := strings.TrimSpace(r.Header.Get(requestIDHeader))
			if len(requestID) == 0 || len(requestID) > 128 {
				requestID = uuid.NewString()
			}

			w.Header().Set(requestIDHeader, requestID)
			ctx := context.WithValue(r.Context(), requestIDContextKey{}, requestID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequestIDFromContext returns request correlation ID, when present.
func RequestIDFromContext(ctx context.Context) string {
	requestID, _ := ctx.Value(requestIDContextKey{}).(string)
	return requestID
}
