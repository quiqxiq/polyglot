package port

import "context"

// DeviceAuthorizer defines the contract for validating whether a user can access a specific device.
type DeviceAuthorizer interface {
	CanAccessDevice(ctx context.Context, userID uint, userRoles []string, deviceID string) (bool, error)
}
