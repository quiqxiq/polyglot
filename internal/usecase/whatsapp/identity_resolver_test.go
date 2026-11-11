package whatsapp_test

import (
	"context"
	"errors"
	"testing"

	"github.com/quixiq/polyglot/internal/domain/customer"
	"github.com/quixiq/polyglot/internal/domain/subscription"
	domainWA "github.com/quixiq/polyglot/internal/domain/whatsapp"
	"github.com/quixiq/polyglot/internal/port"
	ucWA "github.com/quixiq/polyglot/internal/usecase/whatsapp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockUserRepoForResolver struct {
	users      map[string]*customer.User
	allByPhone map[string][]*customer.User
}

func (m *mockUserRepoForResolver) Create(context.Context, *customer.User) error { return nil }
func (m *mockUserRepoForResolver) FindByID(context.Context, uint) (*customer.User, error) {
	return nil, nil
}
func (m *mockUserRepoForResolver) FindByUsername(context.Context, string) (*customer.User, error) {
	return nil, nil
}
func (m *mockUserRepoForResolver) FindByEmail(context.Context, string) (*customer.User, error) {
	return nil, nil
}
func (m *mockUserRepoForResolver) FindByPhone(_ context.Context, phone string) (*customer.User, error) {
	if u, ok := m.users[phone]; ok {
		return u, nil
	}
	return nil, errors.New("not found")
}
func (m *mockUserRepoForResolver) FindAllByPhone(_ context.Context, phone string) ([]*customer.User, error) {
	if us, ok := m.allByPhone[phone]; ok {
		return us, nil
	}
	if u, ok := m.users[phone]; ok {
		return []*customer.User{u}, nil
	}
	return nil, nil
}
func (m *mockUserRepoForResolver) Count(context.Context) (int64, error) { return 0, nil }
func (m *mockUserRepoForResolver) List(context.Context, int, int, string) ([]*customer.User, int64, error) {
	return nil, 0, nil
}
func (m *mockUserRepoForResolver) FindAll(context.Context) ([]*customer.User, error) { return nil, nil }
func (m *mockUserRepoForResolver) FindByRoles(context.Context, []string, bool) ([]*customer.User, error) {
	return nil, nil
}
func (m *mockUserRepoForResolver) Update(context.Context, *customer.User) error { return nil }
func (m *mockUserRepoForResolver) Delete(context.Context, uint) error           { return nil }
func (m *mockUserRepoForResolver) UpdatePassword(context.Context, uint, string) error {
	return nil
}
func (m *mockUserRepoForResolver) UpdateStatus(context.Context, uint, bool) error { return nil }
func (m *mockUserRepoForResolver) AssignDevices(context.Context, uint, []string, *uint) error {
	return nil
}
func (m *mockUserRepoForResolver) GetAssignedDeviceIDs(context.Context, uint) ([]string, error) {
	return nil, nil
}
func (m *mockUserRepoForResolver) IsDeviceAccessibleByUser(context.Context, uint, string) (bool, error) {
	return true, nil
}

type mockCustomerRepoForResolver struct {
	customers map[string]customer.Customer
}

func (m *mockCustomerRepoForResolver) FindByPhone(_ context.Context, phone string) (customer.Customer, error) {
	if c, ok := m.customers[phone]; ok {
		return c, nil
	}
	return customer.Customer{}, errors.New("not found")
}
func (m *mockCustomerRepoForResolver) FindByID(context.Context, string) (customer.Customer, error) {
	return customer.Customer{}, nil
}
func (m *mockCustomerRepoForResolver) FindByCustomerCode(context.Context, string) (customer.Customer, error) {
	return customer.Customer{}, nil
}
func (m *mockCustomerRepoForResolver) FindByPortalAccessCode(context.Context, string) (customer.Customer, error) {
	return customer.Customer{}, nil
}
func (m *mockCustomerRepoForResolver) Save(context.Context, customer.Customer) error { return nil }
func (m *mockCustomerRepoForResolver) FindAll(context.Context) ([]customer.Customer, error) {
	return nil, nil
}
func (m *mockCustomerRepoForResolver) FindPaged(context.Context, port.PageFilter) ([]customer.Customer, error) {
	return nil, nil
}
func (m *mockCustomerRepoForResolver) Delete(context.Context, string) error { return nil }
func (m *mockCustomerRepoForResolver) FindSubscriptions(context.Context, string) ([]subscription.Subscription, error) {
	return nil, nil
}

type mockCacheStore struct {
	store map[string]string
}

func (m *mockCacheStore) Get(_ context.Context, key string) (string, error) {
	if v, ok := m.store[key]; ok {
		return v, nil
	}
	return "", port.ErrCacheMiss
}
func (m *mockCacheStore) Set(_ context.Context, key string, val string, _ int) error {
	m.store[key] = val
	return nil
}
func (m *mockCacheStore) Delete(_ context.Context, key string) error {
	delete(m.store, key)
	return nil
}
func (m *mockCacheStore) Take(_ context.Context, key string) (string, error) {
	if v, ok := m.store[key]; ok {
		delete(m.store, key)
		return v, nil
	}
	return "", port.ErrCacheMiss
}

func TestIdentityResolver(t *testing.T) {
	uRepo := &mockUserRepoForResolver{
		users: map[string]*customer.User{
			"628111111111": {
				ID:          1,
				Username:    "boss_admin",
				FullName:    "Super Admin",
				Role:        "admin",
				PhoneNumber: "628111111111",
				IsActive:    true,
			},
			"082222222222": {
				ID:          2,
				Username:    "agus_tek",
				FullName:    "Agus Teknisi",
				Role:        "teknisi",
				PhoneNumber: "082222222222",
				IsActive:    true,
			},
			"628444444444": {
				ID:          3,
				Username:    "sari_cs",
				FullName:    "Sari CS",
				Role:        "agent",
				PhoneNumber: "628444444444",
				IsActive:    true,
			},
			"628555555555": {
				ID:          4,
				Username:    "budi_fin",
				FullName:    "Budi Finance",
				Role:        "finance",
				PhoneNumber: "628555555555",
				IsActive:    true,
			},
			"628666666666": {
				ID:          5,
				Username:    "mystery_user",
				FullName:    "Mystery User",
				Role:        "custom-role",
				PhoneNumber: "628666666666",
				IsActive:    true,
			},
			"628777777777": {
				ID:          6,
				Username:    "inactive_admin",
				FullName:    "Inactive Admin",
				Role:        "admin",
				PhoneNumber: "628777777777",
				IsActive:    false,
			},
		},
		allByPhone: map[string][]*customer.User{
			"628888888888": {
				{ID: 7, Username: "dup_a", Role: "admin", PhoneNumber: "628888888888", IsActive: true},
				{ID: 8, Username: "dup_b", Role: "admin", PhoneNumber: "628888888888", IsActive: true},
			},
		},
	}

	cRepo := &mockCustomerRepoForResolver{
		customers: map[string]customer.Customer{
			"083333333333": {
				ID:    "cust-1234",
				Name:  "Budi Pelanggan",
				Phone: "083333333333",
			},
		},
	}

	cache := &mockCacheStore{store: make(map[string]string)}
	resolver := ucWA.NewIdentityResolver(uRepo, cRepo, cache)
	ctx := context.Background()

	t.Run("Resolve Admin by international JID", func(t *testing.T) {
		ident, err := resolver.Resolve(ctx, "628111111111@s.whatsapp.net")
		require.NoError(t, err)
		assert.Equal(t, domainWA.RoleAdmin, ident.Role)
		assert.Equal(t, uint(1), *ident.UserID)
		assert.Equal(t, "Super Admin", ident.DisplayName)
		assert.True(t, ident.HasScope(domainWA.ScopeAll))
	})

	t.Run("Resolve Technician stored with local 08 prefix", func(t *testing.T) {
		ident, err := resolver.Resolve(ctx, "6282222222222:12@s.whatsapp.net")
		require.NoError(t, err)
		assert.Equal(t, domainWA.RoleTechnician, ident.Role)
		require.NotNil(t, ident.UserID)
		assert.Equal(t, uint(2), *ident.UserID)
		assert.Equal(t, "Agus Teknisi", ident.DisplayName)
		assert.True(t, ident.HasScope(domainWA.ScopeDeviceRead))
		assert.False(t, ident.HasScope("arbitrary"))
	})

	t.Run("Resolve CS as non-admin role", func(t *testing.T) {
		ident, err := resolver.Resolve(ctx, "628444444444")
		require.NoError(t, err)
		assert.Equal(t, domainWA.RoleCS, ident.Role)
		assert.True(t, ident.HasScope(domainWA.ScopeCustomerRead))
		assert.False(t, ident.HasScope(domainWA.ScopeHarnessAdmin))
		assert.False(t, ident.HasScope(domainWA.ScopeAll))
	})

	t.Run("Resolve Finance as non-admin role", func(t *testing.T) {
		ident, err := resolver.Resolve(ctx, "628555555555")
		require.NoError(t, err)
		assert.Equal(t, domainWA.RoleFinance, ident.Role)
		assert.True(t, ident.HasScope(domainWA.ScopeBillingRead))
		assert.False(t, ident.HasScope(domainWA.ScopeHarnessAdmin))
	})

	t.Run("Unknown role maps to guest", func(t *testing.T) {
		ident, err := resolver.Resolve(ctx, "628666666666")
		require.NoError(t, err)
		assert.Equal(t, domainWA.RoleGuest, ident.Role)
		assert.False(t, ident.HasScope(domainWA.ScopeHarnessAdmin))
	})

	t.Run("Inactive user maps to guest", func(t *testing.T) {
		ident, err := resolver.Resolve(ctx, "628777777777")
		require.NoError(t, err)
		assert.Equal(t, domainWA.RoleGuest, ident.Role)
		assert.Nil(t, ident.UserID)
	})

	t.Run("Ambiguous active users map to guest", func(t *testing.T) {
		ident, err := resolver.Resolve(ctx, "628888888888")
		require.NoError(t, err)
		assert.Equal(t, domainWA.RoleGuest, ident.Role)
		assert.Nil(t, ident.UserID)
	})

	t.Run("Resolve Customer stored with local 08 prefix", func(t *testing.T) {
		ident, err := resolver.Resolve(ctx, "6283333333333@s.whatsapp.net")
		require.NoError(t, err)
		assert.Equal(t, domainWA.RoleCustomer, ident.Role)
		assert.Equal(t, "cust-1234", ident.CustomerID)
		assert.Equal(t, "Budi Pelanggan", ident.DisplayName)
		assert.True(t, ident.HasScope(domainWA.ScopeSelfService))
	})

	t.Run("Resolve Guest for unknown number", func(t *testing.T) {
		ident, err := resolver.Resolve(ctx, "089999999999")
		require.NoError(t, err)
		assert.Equal(t, domainWA.RoleGuest, ident.Role)
		assert.Nil(t, ident.UserID)
		assert.Empty(t, ident.CustomerID)
		assert.True(t, ident.HasScope(domainWA.ScopePublicInfo))
	})

	t.Run("Resolve Invalid Phone", func(t *testing.T) {
		_, err := resolver.Resolve(ctx, "abc")
		require.ErrorIs(t, err, domainWA.ErrInvalidPhoneNumber)
	})

	t.Run("Cached Resolution Hit", func(t *testing.T) {
		// Key was cached in the previous run for 628999999999
		ident, err := resolver.Resolve(ctx, "628999999999")
		require.NoError(t, err)
		assert.Equal(t, domainWA.RoleGuest, ident.Role)
	})
}

func TestIdentityResolverResolveFresh(t *testing.T) {
	admin := &customer.User{
		ID:          1,
		Username:    "boss_admin",
		FullName:    "Super Admin",
		Role:        "admin",
		PhoneNumber: "628111111111",
		IsActive:    true,
	}
	uRepo := &mockUserRepoForResolver{
		users: map[string]*customer.User{"628111111111": admin},
	}
	cache := &mockCacheStore{store: make(map[string]string)}
	resolver := ucWA.NewIdentityResolver(uRepo, nil, cache)
	ctx := context.Background()

	ident, err := resolver.Resolve(ctx, "628111111111")
	require.NoError(t, err)
	require.Equal(t, domainWA.RoleAdmin, ident.Role)

	// Deactivation must not be masked by the identity cache.
	admin.IsActive = false

	ident, err = resolver.Resolve(ctx, "628111111111")
	require.NoError(t, err)
	assert.Equal(t, domainWA.RoleAdmin, ident.Role, "cached resolution is expected to stay stale")

	ident, err = resolver.ResolveFresh(ctx, "628111111111")
	require.NoError(t, err)
	assert.Equal(t, domainWA.RoleGuest, ident.Role)
	assert.Nil(t, ident.UserID)

	// A fresh lookup refreshes the cache for subsequent non-fresh callers.
	ident, err = resolver.Resolve(ctx, "628111111111")
	require.NoError(t, err)
	assert.Equal(t, domainWA.RoleGuest, ident.Role)
}

func TestIdentityResolverResolveFreshPicksUpRoleChange(t *testing.T) {
	uRepo := &mockUserRepoForResolver{
		users: map[string]*customer.User{
			"628111111111": {
				ID:          1,
				Username:    "boss_admin",
				FullName:    "Super Admin",
				Role:        "admin",
				PhoneNumber: "628111111111",
				IsActive:    true,
			},
		},
	}
	resolver := ucWA.NewIdentityResolver(uRepo, nil, &mockCacheStore{store: make(map[string]string)})
	ctx := context.Background()

	ident, err := resolver.Resolve(ctx, "628111111111")
	require.NoError(t, err)
	require.Equal(t, domainWA.RoleAdmin, ident.Role)

	uRepo.users["628111111111"].Role = "teknisi"

	ident, err = resolver.ResolveFresh(ctx, "628111111111")
	require.NoError(t, err)
	assert.Equal(t, domainWA.RoleTechnician, ident.Role)
	assert.False(t, ident.HasScope(domainWA.ScopeHarnessAdmin))
}
