package auth

import (
	"context"
	"testing"

	"github.com/quixiq/polyglot/internal/domain/customer"
)

type mockUserRepoForDeviceAuth struct {
	accessibleMap map[string]bool
}

func (m *mockUserRepoForDeviceAuth) Create(ctx context.Context, user *customer.User) error { return nil }
func (m *mockUserRepoForDeviceAuth) FindByID(ctx context.Context, id uint) (*customer.User, error) {
	return nil, nil
}
func (m *mockUserRepoForDeviceAuth) FindByUsername(ctx context.Context, username string) (*customer.User, error) {
	return nil, nil
}
func (m *mockUserRepoForDeviceAuth) FindByEmail(ctx context.Context, email string) (*customer.User, error) {
	return nil, nil
}
func (m *mockUserRepoForDeviceAuth) Count(ctx context.Context) (int64, error) { return 0, nil }
func (m *mockUserRepoForDeviceAuth) List(ctx context.Context, page, pageSize int, search string) ([]*customer.User, int64, error) {
	return nil, 0, nil
}
func (m *mockUserRepoForDeviceAuth) FindAll(ctx context.Context) ([]*customer.User, error) {
	return nil, nil
}
func (m *mockUserRepoForDeviceAuth) FindByRoles(ctx context.Context, roles []string, activeOnly bool) ([]*customer.User, error) {
	return nil, nil
}
func (m *mockUserRepoForDeviceAuth) Update(ctx context.Context, user *customer.User) error { return nil }
func (m *mockUserRepoForDeviceAuth) Delete(ctx context.Context, id uint) error            { return nil }
func (m *mockUserRepoForDeviceAuth) UpdatePassword(ctx context.Context, id uint, passwordHash string) error {
	return nil
}
func (m *mockUserRepoForDeviceAuth) UpdateStatus(ctx context.Context, id uint, isActive bool) error {
	return nil
}
func (m *mockUserRepoForDeviceAuth) AssignDevices(ctx context.Context, userID uint, deviceIDs []string, assignedBy *uint) error {
	return nil
}
func (m *mockUserRepoForDeviceAuth) GetAssignedDeviceIDs(ctx context.Context, userID uint) ([]string, error) {
	return nil, nil
}
func (m *mockUserRepoForDeviceAuth) IsDeviceAccessibleByUser(ctx context.Context, userID uint, deviceID string) (bool, error) {
	key := string(rune(userID)) + ":" + deviceID
	return m.accessibleMap[key], nil
}

func TestDeviceAuthorizer_CanAccessDevice(t *testing.T) {
	mockRepo := &mockUserRepoForDeviceAuth{
		accessibleMap: map[string]bool{
			string(rune(20)) + ":dev-branch-1": true,
			string(rune(20)) + ":dev-branch-2": true,
		},
	}

	authorizer := NewDeviceAuthorizer(mockRepo)
	ctx := context.Background()

	tests := []struct {
		name      string
		userID    uint
		roles     []string
		deviceID  string
		wantAllow bool
	}{
		{
			name:      "Owner has wildcard access to any device",
			userID:    1,
			roles:     []string{"owner"},
			deviceID:  "dev-core-router",
			wantAllow: true,
		},
		{
			name:      "Admin with assigned device is allowed",
			userID:    20,
			roles:     []string{"admin"},
			deviceID:  "dev-branch-1",
			wantAllow: true,
		},
		{
			name:      "Admin with unassigned device is denied",
			userID:    20,
			roles:     []string{"admin"},
			deviceID:  "dev-branch-99",
			wantAllow: false,
		},
		{
			name:      "Empty deviceID is denied",
			userID:    1,
			roles:     []string{"owner"},
			deviceID:  "",
			wantAllow: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := authorizer.CanAccessDevice(ctx, tt.userID, tt.roles, tt.deviceID)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.wantAllow {
				t.Fatalf("CanAccessDevice(%d, %v, %q) = %v, want %v", tt.userID, tt.roles, tt.deviceID, got, tt.wantAllow)
			}
		})
	}
}
