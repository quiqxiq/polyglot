package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnsureUserRoleAssignmentsPreservesCustomRoles(t *testing.T) {
	ce := newTestEnforcer(t, nil, [][]string{
		{"10", "admin"},
		{"10", "custom_role"},
		{"20", "technician"},
	})

	EnsureUserRoleAssignments(ce, []*UserRef{
		{ID: "10", Role: "teknisi"},
		{ID: "20", Role: "teknisi"},
		{ID: "30", Role: "agent"},
		{ID: "40", Role: ""},
	})

	tests := []struct {
		name  string
		user  string
		roles []string
	}{
		{"stale built-in removed, custom preserved", "10", []string{"custom_role", "teknisi"}},
		{"legacy technician replaced", "20", []string{"teknisi"}},
		{"missing assignment added", "30", []string{"agent"}},
		{"empty role skipped", "40", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ce.GetRolesForUser(tt.user)
			require.NoError(t, err)
			if len(tt.roles) == 0 {
				assert.Empty(t, got)
				return
			}
			assert.ElementsMatch(t, tt.roles, got)
		})
	}
}

func TestEnsureUserRoleAssignmentsNilEnforcer(t *testing.T) {
	// Must not panic on nil enforcer.
	EnsureUserRoleAssignments(nil, []*UserRef{{ID: "1", Role: "owner"}})
}
