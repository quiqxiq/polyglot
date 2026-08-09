package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func CORS(allowedOrigins []string, appEnv string) gin.HandlerFunc {
	isDev := strings.ToLower(appEnv) == "development"

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		allow := false

		if origin != "" {
			for _, o := range allowedOrigins {
				if (o == "*" || o == origin) && (isDev || (!strings.Contains(origin, "localhost") && !strings.Contains(origin, "127.0.0.1"))) {
					allow = true
					break
				}
			}
		}

		if allow {
			c.Header("Access-Control-Allow-Origin", origin)
		} else if origin != "" && isDev {
			c.Header("Access-Control-Allow-Origin", origin)
		} else if len(allowedOrigins) > 0 {
			c.Header("Access-Control-Allow-Origin", allowedOrigins[0])
		}

		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, Connect-Protocol-Version, Connect-Content-Type, Connect-Timeout-Ms")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
