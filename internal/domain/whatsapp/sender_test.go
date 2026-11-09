package whatsapp_test

import (
	"testing"

	"github.com/quixiq/polyglot/internal/domain/whatsapp"
	"github.com/stretchr/testify/assert"
)

func TestDefaultScopesForRole(t *testing.T) {
	assert.Contains(t, whatsapp.DefaultScopesForRole(whatsapp.RoleAdmin), whatsapp.ScopeAll)
	assert.Contains(t, whatsapp.DefaultScopesForRole(whatsapp.RoleTechnician), whatsapp.ScopeDeviceRead)
	assert.Contains(t, whatsapp.DefaultScopesForRole(whatsapp.RoleCustomer), whatsapp.ScopeSelfService)
	assert.Contains(t, whatsapp.DefaultScopesForRole(whatsapp.RoleGuest), whatsapp.ScopePublicInfo)
}

func TestSenderIdentity_HasScope(t *testing.T) {
	admin := whatsapp.SenderIdentity{
		Role:          whatsapp.RoleAdmin,
		AllowedScopes: whatsapp.DefaultScopesForRole(whatsapp.RoleAdmin),
	}
	assert.True(t, admin.HasScope(whatsapp.ScopeDeviceRead))
	assert.True(t, admin.HasScope("any:arbitrary:scope"))

	tech := whatsapp.SenderIdentity{
		Role:          whatsapp.RoleTechnician,
		AllowedScopes: whatsapp.DefaultScopesForRole(whatsapp.RoleTechnician),
	}
	assert.True(t, tech.HasScope(whatsapp.ScopeDeviceRead))
	assert.True(t, tech.HasScope(whatsapp.ScopeOntOps))
	assert.False(t, tech.HasScope("arbitrary:scope"))

	cust := whatsapp.SenderIdentity{
		Role:          whatsapp.RoleCustomer,
		AllowedScopes: whatsapp.DefaultScopesForRole(whatsapp.RoleCustomer),
	}
	assert.True(t, cust.HasScope(whatsapp.ScopeSelfService))
	assert.False(t, cust.HasScope(whatsapp.ScopeDeviceRead))
}
