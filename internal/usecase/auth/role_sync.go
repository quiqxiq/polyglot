package auth

import (
	"fmt"
	"strings"

	"github.com/quixiq/polyglot/internal/domain/customer"
	"github.com/quixiq/polyglot/internal/port"
)

// syncEffectiveRoles reconciles Casbin's grouping table with users.role during
// authentication. users.role is the source of truth, so stale built-in roles
// (including the legacy "technician" alias) are removed, the persisted role is
// added when missing, and custom roles granted through the RBAC service are
// preserved.
func syncEffectiveRoles(roles port.RoleAuthorizer, userID uint, primaryRole string, fallback []string) []string {
	if roles == nil {
		if len(fallback) == 0 && primaryRole != "" {
			return []string{primaryRole}
		}
		return fallback
	}

	id := fmt.Sprintf("%d", userID)
	current, err := roles.GetRolesForUser(id)
	if err != nil || len(current) == 0 {
		current = append([]string(nil), fallback...)
	}

	synced := make([]string, 0, len(current)+1)
	seen := make(map[string]struct{}, len(current)+1)

	for _, role := range current {
		role = strings.TrimSpace(role)
		if role == "" {
			continue
		}

		key := strings.ToLower(role)
		if _, exists := seen[key]; exists {
			continue
		}

		if primaryRole != "" && !strings.EqualFold(role, primaryRole) && customer.IsBuiltInRole(role) {
			_, _ = roles.DeleteRoleForUser(id, role)
			continue
		}

		seen[key] = struct{}{}
		synced = append(synced, role)
	}

	if primaryRole != "" {
		key := strings.ToLower(primaryRole)
		if _, exists := seen[key]; !exists {
			if ok, addErr := roles.AddRoleForUser(id, primaryRole); addErr == nil && ok {
				synced = append(synced, primaryRole)
			}
		}
	}

	if len(synced) == 0 && primaryRole != "" {
		synced = []string{primaryRole}
	}

	return synced
}
