package middleware

import (
	"net/http"
	"time"

	"github.com/quixiq/polyglot/pkg/logger"
	"github.com/sirupsen/logrus"
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

// RequestLogger returns a standard net/http middleware logging all HTTP requests via Logrus.
func RequestLogger() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			wrapper := &responseWriterWrapper{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			next.ServeHTTP(wrapper, r)

			duration := time.Since(start)

			logger.WithComponent("HTTP").WithFields(logrus.Fields{
				"method":   r.Method,
				"path":     r.URL.Path,
				"status":   wrapper.statusCode,
				"duration": duration.String(),
				"remote":   r.RemoteAddr,
				"size":     wrapper.bytesCount,
			}).Infof("%s %s %d (%s)", r.Method, r.URL.Path, wrapper.statusCode, duration)
		})
	}
}
