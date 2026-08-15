package middleware

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/quixiq/polyglot/pkg/logger"
)

type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int64
}

func (w *responseWriterWrapper) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *responseWriterWrapper) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.bytesWritten += int64(n)
	return n, err
}

// RequestLogger logs incoming HTTP requests with structured context via Logrus.
func RequestLogger() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			reqID := r.Header.Get("X-Request-ID")
			if reqID == "" {
				reqID = uuid.NewString()
			}

			w.Header().Set("X-Request-ID", reqID)

			ctx := logger.WithRequestID(r.Context(), reqID)
			r = r.WithContext(ctx)

			wrapper := &responseWriterWrapper{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			next.ServeHTTP(wrapper, r)

			duration := time.Since(start)
			status := wrapper.statusCode

			fields := logger.Fields{
				"method":      r.Method,
				"path":        r.URL.Path,
				"status":      status,
				"duration_ms": duration.Milliseconds(),
				"bytes":       wrapper.bytesWritten,
				"remote_addr": r.RemoteAddr,
				"user_agent":  r.UserAgent(),
			}

			logEntry := logger.FromContext(ctx).WithFields(fields)
			if status >= 500 {
				logEntry.Error("HTTP request server error")
			} else if status >= 400 {
				logEntry.Warn("HTTP request client error")
			} else {
				logEntry.Info("HTTP request completed")
			}
		})
	}
}
