package ws

import (
	"net/http"

	"github.com/quixiq/polyglot/internal/usecase/network"
)

// RegisterEventRoutes registers the realtime SSE & WebSocket terminal endpoints on standard http.ServeMux.
func RegisterEventRoutes(
	mux *http.ServeMux,
	sseHub *SSEHub,
	openTermUC *network.OpenTerminalUseCase,
) {
	mux.Handle("GET /events", sseHub)

	termHandler := NewTerminalHandler(openTermUC)
	mux.Handle("GET /ws/devices/{id}/terminal", termHandler)
}
