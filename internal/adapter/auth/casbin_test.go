package auth_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quixiq/polyglot/internal/adapter/auth"
)

func TestCasbinEnforcer(t *testing.T) {
	// Adjust paths assuming test is run from internal/adapter/auth
	modelPath := filepath.Join("..", "..", "..", "configs", "rbac_model.conf")
	policyPath := filepath.Join("..", "..", "..", "configs", "rbac_policy.csv")

	enforcer, err := auth.NewRBAC(modelPath, policyPath)
	require.NoError(t, err, "failed to initialize casbin enforcer")

	ctx := context.Background()

	tests := []struct {
		name     string
		role     string
		resource string
		action   string
		allowed  bool
	}{
		// Superadmin can access everything
		{"Superadmin any resource", "superadmin", "/api/v1/anything", "POST", true},

		// Owner rules
		{"Owner any resource", "owner", "/api/v1/anything", "DELETE", true},

		// Admin rules
		{"Admin devices", "admin", "/api/v1/devices/123", "POST", true},
		{"Admin users GET", "admin", "/api/v1/users/456", "GET", true},
		{"Admin users POST (denied)", "admin", "/api/v1/users/456", "POST", false},

		// Staff rules
		{"Staff devices GET", "staff", "/api/v1/devices", "GET", true},
		{"Staff devices POST (denied)", "staff", "/api/v1/devices", "POST", false},
		{"Staff customers PUT", "staff", "/api/v1/customers/789", "PUT", true},
		{"Staff customers DELETE (denied)", "staff", "/api/v1/customers/789", "DELETE", false},

		// Teknisi rules
		{"Teknisi devices PUT", "teknisi", "/api/v1/devices/123", "PUT", true},
		{"Teknisi devices DELETE (denied)", "teknisi", "/api/v1/devices/123", "DELETE", false},
		{"Teknisi customers GET", "teknisi", "/api/v1/customers", "GET", true},
		{"Teknisi customers POST (denied)", "teknisi", "/api/v1/customers", "POST", false},

		// Unknown role
		{"Unknown role", "guest", "/api/v1/devices", "GET", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed, err := enforcer.Enforce(ctx, tt.role, tt.resource, tt.action)
			require.NoError(t, err)
			assert.Equal(t, tt.allowed, allowed)
		})
	}
}
