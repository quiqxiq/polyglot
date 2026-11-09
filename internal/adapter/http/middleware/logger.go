package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/quixiq/polyglot/pkg/logger"
)

type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
	bytesCount int
}

func (w *responseWriterWrapper) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *responseWriterWrapper) Write(b []byte) (int, error) {
	if w.statusCode == 0 {
		w.statusCode = http.StatusOK
	}
	n, err := w.ResponseWriter.Write(b)
	w.bytesCount += n
	return n, err
}

// Flush forwards to the underlying writer so streaming handlers (SSE)
// keep seeing an http.Flusher through this wrapper.
func (w *responseWriterWrapper) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Unwrap exposes the original ResponseWriter so http.ResponseController and
// libraries that follow Unwrap (e.g. coder/websocket for Hijack) can reach
// the underlying optional interfaces.
func (w *responseWriterWrapper) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

var _ http.Flusher = (*responseWriterWrapper)(nil)

// RequestLogger returns a standard net/http middleware logging all HTTP requests via pkg/logger.
func RequestLogger() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			reqID := RequestIDFromContext(r.Context())
			isStream := isStreamingRequest(r)

			if isStream {
				logger.WithComponent("HTTP").WithFields(logger.Fields{
					"request_id": reqID,
					"method":     r.Method,
					"path":       r.URL.Path,
					"remote":     r.RemoteAddr,
				}).Info("streaming started")
			}

			wrapper := &responseWriterWrapper{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			next.ServeHTTP(wrapper, r)

			duration := time.Since(start)

			msg := "request completed"
			if isStream {
				msg = "streaming completed"
			}

			logger.WithComponent("HTTP").WithFields(logger.Fields{
				"request_id":  reqID,
				"method":      r.Method,
				"path":        r.URL.Path,
				"status":      wrapper.statusCode,
				"duration_ms": duration.Milliseconds(),
				"remote":      r.RemoteAddr,
				"size":        wrapper.bytesCount,
			}).Info(msg)
		})
	}
}

func isStreamingRequest(r *http.Request) bool {
	if strings.Contains(r.Header.Get("Accept"), "text/event-stream") {
		return true
	}
	if strings.EqualFold(r.Header.Get("Upgrade"), "websocket") {
		return true
	}
	if strings.Contains(r.URL.Path, "/Stream") {
		return true
	}
	return false
}
