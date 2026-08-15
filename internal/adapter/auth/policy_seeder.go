package auth

import (
	"github.com/quixiq/polyglot/pkg/logger"
)

// SeedSystemPolicies seeds default RBAC policies for all Polyglot NetOps Engine modules:
// - Network Devices & Mikhmon Hotspot Administration
// - Realtime WebSockets & SSE Streaming Monitoring
// - Billing, Subscriptions, Plans & Customers
// - WhatsApp Bot, Conversations & Knowledge Base (RAG)
// - Technicians & Dynamic RBAC Management
func SeedSystemPolicies(ce *CasbinEnforcer) {
	if ce == nil {
		return
	}

	systemPolicies := [][]string{
		// 1. Owner & Admin: Full unrestricted access to all /api/v1/* and /ws/* endpoints
		{"owner", "/api/v1/*", ".*"},
		{"owner", "/ws/*", ".*"},
		{"admin", "/api/v1/*", ".*"},
		{"admin", "/ws/*", ".*"},

		// 2. Agent / CS: Customer Service, Live Chat, Conversations, Knowledge Lookup & Customer Info
		{"agent", "/api/v1/conversations*", "(GET|POST|PUT)"},
		{"agent", "/api/v1/sessions*", "GET"},
		{"agent", "/api/v1/knowledge*", "GET"},
		{"agent", "/api/v1/customers*", "(GET|POST)"},
		{"agent", "/api/v1/subscriptions*", "GET"},
		{"agent", "/api/v1/plans*", "GET"},

		// 3. Teknisi: Field Tech, Network Device Monitoring, Mikhmon Read & Streams, Escalations
		{"teknisi", "/api/v1/devices*", "GET"},
		{"teknisi", "/api/v1/devices/*/mikhmon/*", "GET"},
		{"teknisi", "/ws/devices/*/mikhmon/*", "GET"},
		{"teknisi", "/api/v1/conversations*", "(GET|POST)"},
		{"teknisi", "/api/v1/technicians*", "GET"},
	}

	added := 0
	for _, p := range systemPolicies {
		ok, err := ce.AddPolicy(p[0], p[1], p[2])
		if err == nil && ok {
			added++
		}
	}

	if added > 0 {
		logger.WithField("count", added).Info("[PolicySeeder] Seeded new RBAC system policies into PostgreSQL store")
	}
}
