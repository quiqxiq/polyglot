package auth

import (
	"fmt"
	"testing"

	"github.com/casbin/casbin/v3"
	"github.com/casbin/casbin/v3/model"

	"github.com/quixiq/polyglot/internal/config"
)

// newTestEnforcer builds a CasbinEnforcer with an in-memory model (no gorm).
func newTestEnforcer(t *testing.T, policies [][]string, roles [][]string) *CasbinEnforcer {
	t.Helper()
	m, err := model.NewModelFromString(config.RBACModelConf)
	if err != nil {
		t.Fatalf("load model: %v", err)
	}
	e, err := casbin.NewEnforcer(m)
	if err != nil {
		t.Fatalf("new enforcer: %v", err)
	}
	for _, p := range policies {
		if _, err := e.AddPolicy(p[0], p[1], p[2]); err != nil {
			t.Fatalf("add policy %v: %v", p, err)
		}
	}
	for _, g := range roles {
		if _, err := e.AddRoleForUser(g[0], g[1]); err != nil {
			t.Fatalf("add role %v: %v", g, err)
		}
	}
	return &CasbinEnforcer{enforcer: e}
}

func TestEnforceRoleMatrix(t *testing.T) {
	seeder := newTestEnforcer(t, nil, nil)
	SeedSystemPolicies(seeder)

	// SeedSystemPolicies assigns admin/agent/teknisi/owner.
	roles := [][]string{
		{"1001", "owner"},
		{"1002", "admin"},
		{"1003", "agent"},
		{"1004", "teknisi"},
	}
	for _, g := range roles {
		if _, err := seeder.AddRoleForUser(g[0], g[1]); err != nil {
			t.Fatalf("assign role %v: %v", g, err)
		}
	}

	tests := []struct {
		name   string
		userID string
		obj    string
		want   bool
	}{
		{"owner can manage rbac", "1001", "rbac:manage", true},
		{"owner can write skill", "1001", "skill:manage", true},
		{"owner can device command", "1001", "device:command", true},

		{"admin can manage skill", "1002", "skill:manage", true},
		{"admin can manage devices", "1002", "device:manage", true},
		{"admin can manage llmconfig", "1002", "llmconfig:manage", true},
		{"admin CANNOT manage rbac", "1002", "rbac:manage", false},
		{"admin can list technicians", "1002", "technician:read", true},
		{"admin can read logs", "1002", "log:read", true},

		{"agent can read conversation", "1003", "conversation:read", true},
		{"agent can write conversation", "1003", "conversation:write", true},
		{"agent can read skill", "1003", "skill:read", true},
		{"agent CANNOT write skill", "1003", "skill:manage", false},
		{"agent CANNOT device read", "1003", "device:read", false},
		{"agent CANNOT read logs", "1003", "log:read", false},
		{"agent CANNOT manage rbac", "1003", "rbac:manage", false},

		{"teknisi can read device", "1004", "device:read", true},
		{"teknisi can device command", "1004", "device:command", true},
		{"teknisi CANNOT manage device", "1004", "device:manage", false},
		{"teknisi can read conversation", "1004", "conversation:read", true},
		{"teknisi can read skill", "1004", "skill:read", true},
		{"teknisi can read technician", "1004", "technician:read", true},
		{"teknisi can read logs", "1004", "log:read", true},
		{"teknisi CANNOT manage rbac", "1004", "rbac:manage", false},

		{"unknown user denied", "9999", "skill:read", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := seeder.Enforce(tt.userID, tt.obj, "*")
			if err != nil {
				t.Fatalf("Enforce: %v", err)
			}
			if got != tt.want {
				t.Fatalf("Enforce(%s, %q) = %v, want %v", tt.userID, tt.obj, got, tt.want)
			}
		})
	}
}

func TestWildcardRegexMatching(t *testing.T) {
	ce := newTestEnforcer(t, [][]string{
		{"admin", "device:.*", "*"},
		{"owner", ".*", "*"},
	}, nil)

	tests := []struct {
		role string
		obj  string
		want bool
	}{
		{"admin", "device:read", true},
		{"admin", "device:command", true},
		{"admin", "device:anything-else", true},
		{"admin", "customer:read", false},
		{"owner", "customer:read", true},
		{"owner", "rbac:manage", true},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s/%s", tt.role, tt.obj), func(t *testing.T) {
			got, err := ce.Enforce(tt.role, tt.obj, "*")
			if err != nil {
				t.Fatalf("Enforce: %v", err)
			}
			if got != tt.want {
				t.Fatalf("Enforce(%s, %q) = %v, want %v", tt.role, tt.obj, got, tt.want)
			}
		})
	}
}

func TestSyncRolePermissions(t *testing.T) {
	ce := newTestEnforcer(t, [][]string{
		{"billing_staff", "customer:read", "*"},
		{"billing_staff", "old_permission:read", "*"},
	}, nil)

	// Sync new permissions for billing_staff
	newPerms := []string{"billing:read", "billing:write", "customer:read"}
	if err := ce.SyncRolePermissions("billing_staff", newPerms); err != nil {
		t.Fatalf("SyncRolePermissions failed: %v", err)
	}

	// Verify new permissions are allowed
	for _, p := range newPerms {
		allowed, err := ce.Enforce("billing_staff", p, "*")
		if err != nil || !allowed {
			t.Errorf("expected billing_staff to have %s, allowed=%v, err=%v", p, allowed, err)
		}
	}

	// Verify old removed permission is now denied
	allowed, err := ce.Enforce("billing_staff", "old_permission:read", "*")
	if err != nil || allowed {
		t.Errorf("expected old_permission:read to be denied for billing_staff, allowed=%v", allowed)
	}
}

func TestDeleteRole(t *testing.T) {
	ce := newTestEnforcer(t, [][]string{
		{"custom_role", "device:read", "*"},
	}, [][]string{
		{"2001", "custom_role"},
	})

	// Deleting owner should fail
	if err := ce.DeleteRole("owner"); err == nil {
		t.Errorf("expected error when deleting owner role, got nil")
	}

	// Delete custom_role
	if err := ce.DeleteRole("custom_role"); err != nil {
		t.Fatalf("DeleteRole failed: %v", err)
	}

	// Verify permissions and user assignment are removed
	allowed, _ := ce.Enforce("custom_role", "device:read", "*")
	if allowed {
		t.Errorf("expected custom_role to have no permissions after deletion")
	}

	userAllowed, _ := ce.Enforce("2001", "device:read", "*")
	if userAllowed {
		t.Errorf("expected user 2001 to lose custom_role permissions after deletion")
	}
}
