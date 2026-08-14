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
		{"owner can write knowledge", "1001", "knowledge:write", true},
		{"owner can device command", "1001", "device:command", true},

		{"admin can manage knowledge", "1002", "knowledge:write", true},
		{"admin can embed knowledge", "1002", "knowledge:embed", true},
		{"admin can manage devices", "1002", "device:manage", true},
		{"admin can manage llmconfig", "1002", "llmconfig:manage", true},
		{"admin CANNOT manage rbac", "1002", "rbac:manage", false},
		{"admin can list technicians", "1002", "technician:read", true},

		{"agent can read conversation", "1003", "conversation:read", true},
		{"agent can write conversation", "1003", "conversation:write", true},
		{"agent can read knowledge", "1003", "knowledge:read", true},
		{"agent CANNOT write knowledge", "1003", "knowledge:write", false},
		{"agent CANNOT device read", "1003", "device:read", false},
		{"agent CANNOT manage rbac", "1003", "rbac:manage", false},

		{"teknisi can read device", "1004", "device:read", true},
		{"teknisi can device command", "1004", "device:command", true},
		{"teknisi CANNOT manage device", "1004", "device:manage", false},
		{"teknisi can read conversation", "1004", "conversation:read", true},
		{"teknisi can read technician", "1004", "technician:read", true},
		{"teknisi CANNOT manage rbac", "1004", "rbac:manage", false},

		{"unknown user denied", "9999", "knowledge:read", false},
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
