package auth

import (
	"context"
	"strings"

	"github.com/quixiq/polyglot/internal/port"
)

// DefaultDeviceAuthorizer validates user access to specific network devices based on role and assignment.
type DefaultDeviceAuthorizer struct {
	userRepo port.UserRepository
}

var _ port.DeviceAuthorizer = (*DefaultDeviceAuthorizer)(nil)

// NewDeviceAuthorizer creates a new DefaultDeviceAuthorizer.
func NewDeviceAuthorizer(userRepo port.UserRepository) *DefaultDeviceAuthorizer {
	return &DefaultDeviceAuthorizer{
		userRepo: userRepo,
	}
}

// CanAccessDevice checks whether a user with given ID and roles has authorization to operate on deviceID.
// Users with role 'owner' have global wildcard access to all devices.
// Non-owners require an explicit record in user_devices.
func (a *DefaultDeviceAuthorizer) CanAccessDevice(ctx context.Context, userID uint, userRoles []string, deviceID string) (bool, error) {
	if deviceID == "" {
		return false, nil
	}

	// Owner role has unrestricted access to all devices in the system
	for _, r := range userRoles {
		if strings.EqualFold(r, "owner") {
			return true, nil
		}
	}

	if userID == 0 || a.userRepo == nil {
		return false, nil
	}

	return a.userRepo.IsDeviceAccessibleByUser(ctx, userID, deviceID)
}
