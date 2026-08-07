package ws

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/port"
)

// RegisterEventRoutes registers the realtime SSE & WebSocket terminal endpoints.
func RegisterEventRoutes(
	r *gin.Engine,
	sseHub *SSEHub,
	getter func(ctx context.Context, deviceID string) (port.DeviceDriver, error),
	targetResolver func(ctx context.Context, deviceID string) (*device.Target, error),
) {
	r.GET("/events", sseHub.RegisterClient)

	termHandler := NewTerminalHandler(getter, targetResolver)
	r.GET("/ws/devices/:id/terminal", func(c *gin.Context) {
		termHandler.ServeHTTP(c)
	})
}
