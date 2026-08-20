package middleware

import "net/http"

// Middleware defines standard Go HTTP middleware signature.
type Middleware func(http.Handler) http.Handler

// Chain combines multiple middlewares into a single handler.
func Chain(handler http.Handler, middlewares ...Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}
