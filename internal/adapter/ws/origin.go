package ws

import (
	"net/url"
	"strings"
)

// OriginPatterns converts CORS origin URLs (e.g. http://localhost:5173) into
// websocket origin host patterns. Wildcards are ignored deliberately: the
// WebSocket library rejects "*" to avoid cross-site hijacking.
func OriginPatterns(origins []string) []string {
	patterns := make([]string, 0, len(origins))
	seen := make(map[string]struct{}, len(origins))
	for _, raw := range origins {
		raw = strings.TrimSpace(raw)
		if raw == "" || raw == "*" {
			continue
		}
		host := raw
		if u, err := url.Parse(raw); err == nil && u.Host != "" {
			host = u.Host
		}
		host = strings.ToLower(strings.TrimSpace(host))
		if host == "" || strings.Contains(host, "*") {
			continue
		}
		if _, ok := seen[host]; ok {
			continue
		}
		seen[host] = struct{}{}
		patterns = append(patterns, host)
	}
	return patterns
}
