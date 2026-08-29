package auth_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/quixiq/polyglot/internal/domain/customer"
	authUC "github.com/quixiq/polyglot/internal/usecase/auth"
)

type mockUserRepo struct {
	users map[uint]*customer.User
}

func newMockUserRepo() *mockUserRepo {
	pwHash, _ := bcrypt.GenerateFromPassword([]byte("oldSecret123"), bcrypt.DefaultCost)
	return &mockUserRepo{
		users: map[uint]*customer.User{
			1: {
				ID:             1,
				Username:       "admin_user",
				Email:          "admin@example.com",
				PasswordHash:   string(pwHash),
				Role:           "admin",
				FullName:       "Admin Original",
				PhoneNumber:    "628123456789",
				Specialization: "Network Admin",
				IsActive:       true,
			},
		},
	}
}

func (m *mockUserRepo) FindByID(ctx context.Context, id uint) (*customer.User, error) {
	if u, ok := m.users[id]; ok {
		return u, nil
	}
	return nil, customer.ErrUserNotFound
}

func (m *mockUserRepo) FindByUsername(ctx context.Context, username string) (*customer.User, error) {
	for _, u := range m.users {
		if u.Username == username {
			return u, nil
		}
	}
	return nil, customer.ErrUserNotFound
}

func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*customer.User, error) {
	for _, u := range m.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, nil
}

func (m *mockUserRepo) FindByRoles(ctx context.Context, roles []string, activeOnly bool) ([]*customer.User, error) {
	return nil, nil
}

func (m *mockUserRepo) Create(ctx context.Context, u *customer.User) error {
	m.users[u.ID] = u
	return nil
}

func (m *mockUserRepo) Update(ctx context.Context, u *customer.User) error {
	m.users[u.ID] = u
	return nil
}

func (m *mockUserRepo) FindAll(ctx context.Context) ([]*customer.User, error) {
	var list []*customer.User
	for _, u := range m.users {
		list = append(list, u)
	}
	return list, nil
}

func (m *mockUserRepo) UpdateStatus(ctx context.Context, id uint, isActive bool) error {
	if u, ok := m.users[id]; ok {
		u.IsActive = isActive
		return nil
	}
	return customer.ErrUserNotFound
}

func (m *mockUserRepo) UpdatePassword(ctx context.Context, id uint, hash string) error {
	if u, ok := m.users[id]; ok {
		u.PasswordHash = hash
		return nil
	}
	return customer.ErrUserNotFound
}

func (m *mockUserRepo) SetActive(ctx context.Context, id uint, active bool) error {
	return nil
}

func (m *mockUserRepo) Delete(ctx context.Context, id uint) error {
	delete(m.users, id)
	return nil
}

func (m *mockUserRepo) Count(ctx context.Context) (int64, error) {
	return int64(len(m.users)), nil
}

func (m *mockUserRepo) List(ctx context.Context, offset, limit int, search string) ([]*customer.User, int64, error) {
	var list []*customer.User
	for _, u := range m.users {
		list = append(list, u)
	}
	return list, int64(len(list)), nil
}

func (m *mockUserRepo) AssignDevices(ctx context.Context, userID uint, deviceIDs []string, assignedBy *uint) error {
	if u, ok := m.users[userID]; ok {
		u.AssignedDeviceIDs = deviceIDs
		return nil
	}
	return customer.ErrUserNotFound
}

func (m *mockUserRepo) GetAssignedDeviceIDs(ctx context.Context, userID uint) ([]string, error) {
	if u, ok := m.users[userID]; ok {
		return u.AssignedDeviceIDs, nil
	}
	return nil, nil
}

func (m *mockUserRepo) IsDeviceAccessibleByUser(ctx context.Context, userID uint, deviceID string) (bool, error) {
	if u, ok := m.users[userID]; ok {
		for _, d := range u.AssignedDeviceIDs {
			if d == deviceID {
				return true, nil
			}
		}
	}
	return false, nil
}

func TestAuthUseCase_UpdateProfileAndChangePassword(t *testing.T) {
	ctx := context.Background()
	userRepo := newMockUserRepo()
	uc := authUC.NewAuthUseCase(userRepo, nil, nil, nil, nil)

	// Test UpdateProfile
	updated, err := uc.UpdateProfile(ctx, 1, "Budi Teknisi Baru", "628999888777", "budi.baru@example.com", "Wireless & OLT")
	require.NoError(t, err)
	assert.Equal(t, "Budi Teknisi Baru", updated.FullName)
	assert.Equal(t, "628999888777", updated.PhoneNumber)
	assert.Equal(t, "budi.baru@example.com", updated.Email)
	assert.Equal(t, "Wireless & OLT", updated.Specialization)

	// Test ChangePassword - Incorrect Old Password
	err = uc.ChangePassword(ctx, 1, "wrongPassword", "newPassword1234")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "current password is incorrect")

	// Test ChangePassword - Too Short
	err = uc.ChangePassword(ctx, 1, "oldSecret123", "short")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "at least 8 characters")

	// Test ChangePassword - Success
	err = uc.ChangePassword(ctx, 1, "oldSecret123", "newSecret5678")
	require.NoError(t, err)

	// Verify new password works
	user, err := userRepo.FindByID(ctx, 1)
	require.NoError(t, err)
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("newSecret5678"))
	assert.NoError(t, err)
}
