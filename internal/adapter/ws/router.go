package ws

import (
	"io"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/quixiq/polyglot/internal/driver/mikrotik"
)

// RegisterStreamingRoutes registers all SSE realtime streaming endpoints for Mikhmon & Hotspot monitoring.
func RegisterStreamingRoutes(r *gin.Engine, h *MikhmonStreamHandler) {
	// 1. Ethernet Traffic Stream
	r.GET("/ws/devices/:deviceId/mikhmon/traffic", func(c *gin.Context) {
		deviceID := c.Param("deviceId")
		iface := envOr("INTERFACE", c.Query("interface"))
		if iface == "" {
			iface = "ether1"
		}

		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")

		outChan := make(chan []byte, 10)
		go func() {
			_ = h.StreamTraffic(c.Request.Context(), deviceID, iface, outChan)
			close(outChan)
		}()

		c.Stream(func(w io.Writer) bool {
			msg, ok := <-outChan
			if !ok {
				return false
			}
			c.SSEvent("traffic", string(msg))
			return true
		})
	})

	// 2. Simple Queues Stream
	r.GET("/ws/devices/:deviceId/mikhmon/queues", func(c *gin.Context) {
		deviceID := c.Param("deviceId")
		p := mikrotik.QueueStreamParams{
			NameFilter:   c.Query("name"),
			ParentFilter: c.Query("parent"),
			ParentsOnly:  c.Query("parents_only") == "true" || c.Query("parents_only") == "1",
			Interval:     c.Query("interval"),
		}

		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")

		outChan := make(chan []byte, 10)
		go func() {
			_ = h.StreamQueueStats(c.Request.Context(), deviceID, p, outChan)
			close(outChan)
		}()

		c.Stream(func(w io.Writer) bool {
			msg, ok := <-outChan
			if !ok {
				return false
			}
			c.SSEvent("queue_stats", string(msg))
			return true
		})
	})

	// 3. System Resource Stream (CPU/RAM/Uptime)
	r.GET("/ws/devices/:deviceId/mikhmon/resource", func(c *gin.Context) {
		deviceID := c.Param("deviceId")
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")

		outChan := make(chan []byte, 10)
		go func() {
			_ = h.StreamResource(c.Request.Context(), deviceID, outChan)
			close(outChan)
		}()

		c.Stream(func(w io.Writer) bool {
			msg, ok := <-outChan
			if !ok {
				return false
			}
			c.SSEvent("resource", string(msg))
			return true
		})
	})

	// 4. Hotspot Active Sessions Stream
	r.GET("/ws/devices/:deviceId/mikhmon/hotspot-active", func(c *gin.Context) {
		deviceID := c.Param("deviceId")
		userFilter := c.Query("user")
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")

		outChan := make(chan []byte, 10)
		go func() {
			_ = h.StreamActiveSessions(c.Request.Context(), deviceID, userFilter, outChan)
			close(outChan)
		}()

		c.Stream(func(w io.Writer) bool {
			msg, ok := <-outChan
			if !ok {
				return false
			}
			c.SSEvent("hotspot_active", string(msg))
			return true
		})
	})

	// 5. Hotspot Users Stream (All Vouchers)
	r.GET("/ws/devices/:deviceId/mikhmon/hotspot-users", func(c *gin.Context) {
		deviceID := c.Param("deviceId")
		profileFilter := c.Query("profile")
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")

		outChan := make(chan []byte, 10)
		go func() {
			_ = h.StreamHotspotUsers(c.Request.Context(), deviceID, profileFilter, outChan)
			close(outChan)
		}()

		c.Stream(func(w io.Writer) bool {
			msg, ok := <-outChan
			if !ok {
				return false
			}
			c.SSEvent("hotspot_users", string(msg))
			return true
		})
	})

	// 6. PPPoE Active Sessions Stream
	r.GET("/ws/devices/:deviceId/mikhmon/ppp-active", func(c *gin.Context) {
		deviceID := c.Param("deviceId")
		userFilter := c.Query("user")
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")

		outChan := make(chan []byte, 10)
		go func() {
			_ = h.StreamPPPActive(c.Request.Context(), deviceID, userFilter, outChan)
			close(outChan)
		}()

		c.Stream(func(w io.Writer) bool {
			msg, ok := <-outChan
			if !ok {
				return false
			}
			c.SSEvent("ppp_active", string(msg))
			return true
		})
	})

	// 7. PPPoE Secrets Stream
	r.GET("/ws/devices/:deviceId/mikhmon/ppp-secrets", func(c *gin.Context) {
		deviceID := c.Param("deviceId")
		nameFilter := c.Query("name")
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")

		outChan := make(chan []byte, 10)
		go func() {
			_ = h.StreamPPPOESecrets(c.Request.Context(), deviceID, nameFilter, outChan)
			close(outChan)
		}()

		c.Stream(func(w io.Writer) bool {
			msg, ok := <-outChan
			if !ok {
				return false
			}
			c.SSEvent("ppp_secrets", string(msg))
			return true
		})
	})

	// 8. Hotspot Inactive Stream
	r.GET("/ws/devices/:deviceId/mikhmon/hotspot-inactive", func(c *gin.Context) {
		deviceID := c.Param("deviceId")
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")

		outChan := make(chan []byte, 10)
		go func() {
			_ = h.StreamHotspotInactive(c.Request.Context(), deviceID, outChan)
			close(outChan)
		}()

		c.Stream(func(w io.Writer) bool {
			msg, ok := <-outChan
			if !ok {
				return false
			}
			c.SSEvent("hotspot_inactive", string(msg))
			return true
		})
	})

	// 9. PPPoE Inactive Stream
	r.GET("/ws/devices/:deviceId/mikhmon/ppp-inactive", func(c *gin.Context) {
		deviceID := c.Param("deviceId")
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")

		outChan := make(chan []byte, 10)
		go func() {
			_ = h.StreamPPPOEInactive(c.Request.Context(), deviceID, outChan)
			close(outChan)
		}()

		c.Stream(func(w io.Writer) bool {
			msg, ok := <-outChan
			if !ok {
				return false
			}
			c.SSEvent("ppp_inactive", string(msg))
			return true
		})
	})
}

// RegisterDeviceStreamingRoutes registers SSE/WebSocket realtime streaming endpoints for device inventory status.
func RegisterDeviceStreamingRoutes(r *gin.Engine, dh *DeviceStreamHandler) {
	// Stream status for all registered devices using native wire streaming
	r.GET("/ws/devices/stream", func(c *gin.Context) {
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")

		outChan := make(chan []byte, 10)
		go func() {
			_ = dh.StreamDevicesStatus(c.Request.Context(), outChan)
			close(outChan)
		}()

		c.Stream(func(w io.Writer) bool {
			msg, ok := <-outChan
			if !ok {
				return false
			}
			c.SSEvent("devices_status", string(msg))
			return true
		})
	})

	// Stream status for a single device using native wire streaming
	r.GET("/ws/devices/:deviceId/status", func(c *gin.Context) {
		deviceID := c.Param("deviceId")
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")

		outChan := make(chan []byte, 10)
		go func() {
			_ = dh.StreamSingleDeviceStatus(c.Request.Context(), deviceID, outChan)
			close(outChan)
		}()

		c.Stream(func(w io.Writer) bool {
			msg, ok := <-outChan
			if !ok {
				return false
			}
			c.SSEvent("device_status", string(msg))
			return true
		})
	})
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
