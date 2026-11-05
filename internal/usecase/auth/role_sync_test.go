package auth

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeRoleAuthorizer is a minimal in-memory port.RoleAuthorizer used to test
// role reconciliation without Casbin or a database.
type fakeRoleAuthorizer struct {
	roles      map[string][]string
	getErr     error
	addFailure bool
}

func newFakeRoleAuthorizer(roles map[string][]string) *fakeRoleAuthorizer {
	return &fakeRoleAuthorizer{roles: roles}
}

func (f *fakeRoleAuthorizer) GetRolesForUser(user string) ([]string, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	return append([]string(nil), f.roles[user]...), nil
}

func (f *fakeRoleAuthorizer) GetImplicitPermissionsForUser(_ string) ([]string, error) {
	return nil, nil
}

func (f *fakeRoleAuthorizer) AddRoleForUser(user, role string) (bool, error) {
	if f.addFailure {
		return false, errors.New("add failed")
	}
	for _, existing := range f.roles[user] {
		if existing == role {
			return false, nil
		}
	}
	f.roles[user] = append(f.roles[user], role)
	return true, nil
}

func (f *fakeRoleAuthorizer) DeleteRoleForUser(user, role string) (bool, error) {
	current := f.roles[user]
	for i, existing := range current {
		if existing == role {
			f.roles[user] = append(current[:i], current[i+1:]...)
			return true, nil
		}
	}
	return false, nil
}

func (f *fakeRoleAuthorizer) DeleteRolesForUser(user string) (bool, error) {
	hadRoles := len(f.roles[user]) > 0
	delete(f.roles, user)
	return hadRoles, nil
}

func TestSyncEffectiveRoles_NilAuthorizer(t *testing.T) {
	assert.Equal(t, []string{"teknisi"}, syncEffectiveRoles(nil, 1, "teknisi", nil))
	assert.Equal(t, []string{"agent"}, syncEffectiveRoles(nil, 1, "", []string{"agent"}))
	assert.Nil(t, syncEffectiveRoles(nil, 1, "", nil))
}

func TestSyncEffectiveRoles_PreservesCustomRoles(t *testing.T) {
	authorizer := newFakeRoleAuthorizer(map[string][]string{
		"42": {"billing_staff", "teknisi", "auditor"},
	})

	got := syncEffectiveRoles(authorizer, 42, "teknisi", nil)

	assert.ElementsMatch(t, []string{"billing_staff", "teknisi", "auditor"}, got)
	assert.ElementsMatch(t, []string{"billing_staff", "teknisi", "auditor"}, authorizer.roles["42"])
}

func TestSyncEffectiveRoles_RemovesStaleBuiltInRoles(t *testing.T) {
	authorizer := newFakeRoleAuthorizer(map[string][]string{
		"42": {"admin", "technician", "custom_role"},
	})

	got := syncEffectiveRoles(authorizer, 42, "teknisi", nil)

	assert.ElementsMatch(t, []string{"teknisi", "custom_role"}, got)
	assert.ElementsMatch(t, []string{"teknisi", "custom_role"}, authorizer.roles["42"])
}

func TestSyncEffectiveRoles_UsesFallbackWhenLookupFails(t *testing.T) {
	authorizer := newFakeRoleAuthorizer(map[string][]string{
		"42": {"admin"},
	})
	authorizer.getErr = errors.New("store unavailable")

	got := syncEffectiveRoles(authorizer, 42, "teknisi", []string{"admin", "custom_role"})

	assert.ElementsMatch(t, []string{"teknisi", "custom_role"}, got)
}

func TestSyncEffectiveRoles_AddsMissingPrimaryRole(t *testing.T) {
	authorizer := newFakeRoleAuthorizer(map[string][]string{})

	got := syncEffectiveRoles(authorizer, 7, "agent", nil)

	require.Equal(t, []string{"agent"}, got)
	require.Equal(t, []string{"agent"}, authorizer.roles["7"])
}

func TestSyncEffectiveRoles_CaseInsensitiveDeduplication(t *testing.T) {
	authorizer := newFakeRoleAuthorizer(map[string][]string{
		"42": {"TEKNISI"},
	})

	got := syncEffectiveRoles(authorizer, 42, "teknisi", nil)

	assert.Equal(t, []string{"TEKNISI"}, got)
	assert.Equal(t, []string{"TEKNISI"}, authorizer.roles["42"])
}

func TestSyncEffectiveRoles_FallsBackToPrimaryWhenAddFails(t *testing.T) {
	authorizer := newFakeRoleAuthorizer(map[string][]string{})
	authorizer.addFailure = true

	got := syncEffectiveRoles(authorizer, 7, "agent", nil)

	assert.Equal(t, []string{"agent"}, got)
}
