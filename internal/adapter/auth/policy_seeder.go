package auth

import (
	"github.com/quixiq/polyglot/pkg/logger"
	"strings"
)

// SeedSystemPolicies seeds the default RBAC policies using the "resource:action"
// object format (e.g. "knowledge:write"). Enforcement matches the request object
// against policy objects with regexMatch, so wildcards are expressed as regex
// patterns ("device:.*" for all device actions, ".*" for full access).
//
// Role matrix (final, agreed with product owner):
//   - owner  : full access to everything, including rbac:manage
//   - admin  : every resource EXCEPT rbac:manage (explicit per-resource list,
//     no deny rule — admin simply has no rbac policy)
//   - agent  : conversations (read/write), customer:read, knowledge:read,
//     billing:read, whatsapp:read
//   - teknisi: device:read + device:command (non-destructive enforced in the
//     ExecuteCommand usecase, not here), conversation:read, customer:read,
//     knowledge:read, technician:read, probe:read, hotspot:read
func SeedSystemPolicies(ce *CasbinEnforcer) {
	if ce == nil {
		return
	}

	removeLegacyPathPolicies(ce)

	systemPolicies := [][]string{
		// Owner — wildcard penuh (termasuk rbac:manage).
		{"owner", ".*", "*"},

		// Admin — semua resource kecuali rbac (daftar eksplisit, tanpa deny).
		{"admin", "device:.*", "*"},
		{"admin", "probe:.*", "*"},
		{"admin", "customer:.*", "*"},
		{"admin", "billing:.*", "*"},
		{"admin", "whatsapp:.*", "*"},
		{"admin", "conversation:.*", "*"},
		{"admin", "skill:.*", "*"},
		{"admin", "llmconfig:.*", "*"},
		{"admin", "technician:.*", "*"},
		{"admin", "hotspot:.*", "*"},
		{"admin", "ppp:.*", "*"},
		{"admin", "log:.*", "*"},
		{"admin", "user:.*", "*"},
		{"admin", "setting:.*", "*"},
		{"admin", "profile:.*", "*"},
		{"admin", "report:.*", "*"},
		{"admin", "ispadmin:.*", "*"},
		{"admin", "registration:.*", "*"},
		{"admin", "cashbook:.*", "*"},
		{"admin", "notification:.*", "*"},

		// Agent / CS — customer service, voucher generator, & knowledge lookup.
		{"agent", "conversation:.*", "*"},
		{"agent", "customer:read", "*"},
		{"agent", "skill:read", "*"},
		{"agent", "billing:read", "*"},
		{"agent", "whatsapp:read", "*"},
		{"agent", "profile:.*", "*"},
		{"agent", "report:read", "*"},
		{"agent", "registration:read", "*"},
		{"agent", "notification:read", "*"},
		{"agent", "hotspot:voucher:generate", "*"},
		{"agent", "hotspot:user:read", "*"},
		{"agent", "hotspot:profile:read", "*"},
		{"agent", "ppp:secret:read", "*"},

		// Teknisi — monitoring, field installation & diagnostic.
		{"teknisi", "device:read", "*"},
		{"teknisi", "device:diagnostic:exec", "*"},
		{"teknisi", "conversation:read", "*"},
		{"teknisi", "customer:read", "*"},
		{"teknisi", "skill:read", "*"},
		{"teknisi", "technician:read", "*"},
		{"teknisi", "probe:read", "*"},
		{"teknisi", "hotspot:profile:read", "*"},
		{"teknisi", "hotspot:user:read", "*"},
		{"teknisi", "hotspot:active:read", "*"},
		{"teknisi", "hotspot:dhcp:read", "*"},
		{"teknisi", "hotspot:host:read", "*"},
		{"teknisi", "hotspot:binding:read", "*"},
		{"teknisi", "hotspot:read", "*"},
		{"teknisi", "hotspot:report:read", "*"},
		{"teknisi", "hotspot:expire:read", "*"},
		{"teknisi", "hotspot:template:read", "*"},
		{"teknisi", "ppp:secret:read", "*"},
		{"teknisi", "ppp:profile:read", "*"},
		{"teknisi", "ppp:active:read", "*"},
		{"teknisi", "log:read", "*"},
		{"teknisi", "profile:.*", "*"},
		{"teknisi", "registration:read", "*"},
		{"teknisi", "registration:install", "*"},
	}

	added := 0
	for _, p := range systemPolicies {
		ok, err := ce.AddPolicy(p[0], p[1], p[2])
		if err == nil && ok {
			added++
		}
	}

	if added > 0 {
		logger.WithComponent("PolicySeeder").Infof("Seeded %d new RBAC system policies (resource:action format)", added)
	}
}

// removeLegacyPathPolicies deletes policies from the old model whose object
// was an HTTP path (e.g. "/api/v1/*") — they never match the new
// "resource:action" enforcement and would only confuse the RBAC admin UI.
func removeLegacyPathPolicies(ce *CasbinEnforcer) {
	existing, err := ce.GetPolicies()
	if err != nil {
		return
	}
	removed := 0
	for _, p := range existing {
		if len(p) < 2 || !strings.HasPrefix(p[1], "/") {
			continue
		}
		if ok, err := ce.RemovePolicy(p[0], p[1], p[2]); err == nil && ok {
			removed++
		}
	}
	if removed > 0 {
		logger.WithComponent("PolicySeeder").Infof("Removed %d legacy path-based RBAC policies", removed)
	}
}

// EnsureUserRoleAssignments syncs Casbin role assignments (g table) with the
// users table: every registered user gets the role stored in their users.role column.
// Stale roles that no longer match are removed.
func EnsureUserRoleAssignments(ce *CasbinEnforcer, users []*UserRef) {
	if ce == nil {
		return
	}
	synced := 0
	for _, u := range users {
		if u == nil || u.Role == "" {
			continue
		}
		currentRoles, _ := ce.GetRolesForUser(u.ID)
		for _, r := range currentRoles {
			if r != u.Role {
				_, _ = ce.DeleteRoleForUser(u.ID, r)
			}
		}
		if ok, err := ce.AddRoleForUser(u.ID, u.Role); err == nil && ok {
			synced++
		}
	}
	if synced > 0 {
		logger.WithComponent("PolicySeeder").Infof("Synced role assignments for %d user(s) in Casbin", synced)
	}
}

// UserRef is the minimal user data the seeder needs for role assignment.
type UserRef struct {
	ID   string
	Role string
}
