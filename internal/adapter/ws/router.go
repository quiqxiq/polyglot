package ws

import (
	"github.com/gin-gonic/gin"

	"github.com/quixiq/polyglot/internal/usecase/network"
)

// RegisterEventRoutes registers the realtime SSE & WebSocket terminal endpoints.
func RegisterEventRoutes(
	r *gin.Engine,
	sseHub *SSEHub,
	openTermUC *network.OpenTerminalUseCase,
) {
	r.GET("/events", sseHub.RegisterClient)

	termHandler := NewTerminalHandler(openTermUC)
	r.GET("/ws/devices/:id/terminal", func(c *gin.Context) {
		termHandler.ServeHTTP(c)
	})
}
