package auth

import (
	"strings"

	"github.com/quixiq/polyglot/internal/domain/customer"
	"github.com/quixiq/polyglot/pkg/logger"
)

// SeedSystemPolicies seeds the default RBAC policies using the "resource:action"
// object format (e.g. "knowledge:write"). Enforcement matches the request object
// against policy objects with regexMatch, so wildcards are expressed as regex
// patterns ("device:.*" for all device actions, ".*" for full access).
//
// Role matrix (final, agreed with product owner):
//   - owner  : full access to everything, including rbac:manage
//   - superadmin: legacy alias, full access to everything
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

		// Superadmin — role legacy; tetap diberi wildcard agar akun lama tidak
		// kehilangan akses sebelum dimigrasikan ke owner.
		{"superadmin", ".*", "*"},

		// Admin — semua resource kecuali rbac (daftar eksplisit, tanpa deny).
		{"admin", "device:.*", "*"},
		{"admin", "probe:.*", "*"},
		{"admin", "customer:.*", "*"},
		{"admin", "billing:.*", "*"},
		{"admin", "debt:.*", "*"},
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
		{"admin", "monitor:.*", "*"},
		{"admin", "network:.*", "*"},
		{"admin", "olt:.*", "*"},
		{"admin", "ont:.*", "*"},
		{"admin", "dial:.*", "*"},

		// Agent / CS — customer service, voucher generator, & knowledge lookup.
		{"agent", "conversation:.*", "*"},
		{"agent", "customer:read", "*"},
		{"agent", "skill:read", "*"},
		{"agent", "billing:read", "*"},
		{"agent", "debt:read", "*"},
		{"agent", "whatsapp:read", "*"},
		{"agent", "profile:.*", "*"},
		{"agent", "report:read", "*"},
		{"agent", "registration:read", "*"},
		{"agent", "notification:read", "*"},
		{"agent", "hotspot:voucher:generate", "*"},
		{"agent", "hotspot:user:read", "*"},
		{"agent", "hotspot:profile:read", "*"},
		{"agent", "ppp:secret:read", "*"},

		// Finance — kasir/keuangan: tagihan, pembayaran, hutang, kas, laporan.
		{"finance", "billing:.*", "*"},
		{"finance", "debt:.*", "*"},
		{"finance", "cashbook:.*", "*"},
		{"finance", "report:.*", "*"},
		{"finance", "customer:read", "*"},
		{"finance", "subscription:read", "*"},
		{"finance", "notification:read", "*"},

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
		{"teknisi", "log:read", "*"},
		{"teknisi", "profile:.*", "*"},
		{"teknisi", "registration:read", "*"},
		{"teknisi", "registration:install", "*"},
		{"teknisi", "monitor:read", "*"},
		{"teknisi", "network:read", "*"},
		{"teknisi", "olt:read", "*"},
		{"teknisi", "ont:read", "*"},

		// Teknisi — Web Dial dedicated portal permissions (dial:*):
		{"teknisi", "dial:dashboard:read", "*"},
		{"teknisi", "dial:dashboard:ping", "*"},
		{"teknisi", "dial:device:read", "*"},
		{"teknisi", "dial:ppp-active:read", "*"},
		{"teknisi", "dial:ppp-active:ping", "*"},
		{"teknisi", "dial:ppp-non-active:read", "*"},
		{"teknisi", "dial:ppp-secrets:read", "*"},
		{"teknisi", "dial:ppp-profiles:read", "*"},
		{"teknisi", "dial:logs:read", "*"},
		{"teknisi", "dial:monitor:read", "*"},
		{"teknisi", "dial:monitor:diagnostic", "*"},
	}

	added := 0
	for _, p := range systemPolicies {
		ok, err := ce.AddPolicy(p[0], p[1], p[2])
		if err == nil && ok {
			added++
		}
	}

	if added > 0 {
		logger.WithComponent("PolicySeeder").WithField("policy_count", added).Info("seeded RBAC system policies")
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
		if len(p) < 2 {
			continue
		}
		// Hapus legacy HTTP path policies ("/api/v1/*")
		if strings.HasPrefix(p[1], "/") {
			if ok, err := ce.RemovePolicy(p[0], p[1], p[2]); err == nil && ok {
				removed++
			}
			continue
		}
		// Hapus legacy PPP main permissions dari teknisi (dial:* dipakai untuk Web Dial)
		if p[0] == "teknisi" && (p[1] == "ppp:secret:read" || p[1] == "ppp:profile:read" || p[1] == "ppp:active:read") {
			if ok, err := ce.RemovePolicy(p[0], p[1], p[2]); err == nil && ok {
				removed++
			}
		}
	}
	if removed > 0 {
		logger.WithComponent("PolicySeeder").WithField("policy_count", removed).Info("removed legacy RBAC policies")
	}
}

// EnsureUserRoleAssignments syncs Casbin role assignments (g table) with the
// users table: every registered user gets the role stored in their users.role
// column. Only stale built-in roles (including the legacy "technician" alias)
// are removed; custom roles created through the RBAC service are preserved.
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
			if strings.EqualFold(r, u.Role) {
				continue
			}
			if customer.IsBuiltInRole(r) {
				_, _ = ce.DeleteRoleForUser(u.ID, r)
			}
		}
		if ok, err := ce.AddRoleForUser(u.ID, u.Role); err == nil && ok {
			synced++
		}
	}
	if synced > 0 {
		logger.WithComponent("PolicySeeder").WithField("assignment_count", synced).Info("synced Casbin role assignments")
	}
}

// UserRef is the minimal user data the seeder needs for role assignment.
type UserRef struct {
	ID   string
	Role string
}
