package ws

import (
	"net/http"

	"github.com/quixiq/polyglot/internal/usecase/network"
)

// RegisterEventRoutes registers the realtime SSE & WebSocket terminal endpoints onto a standard http.ServeMux.
func RegisterEventRoutes(
	mux *http.ServeMux,
	sseHub *SSEHub,
	openTermUC *network.OpenTerminalUseCase,
) {
	mux.HandleFunc("GET /events", sseHub.ServeHTTP)

	termHandler := NewTerminalHandler(openTermUC)
	mux.HandleFunc("GET /ws/devices/{id}/terminal", termHandler.ServeHTTP)
}
