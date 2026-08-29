package hotspot

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/port"
)

// CreateProfile builds and executes the /ip/hotspot/user/profile/add command
func (u *UseCase) CreateProfile(ctx context.Context, driver port.DeviceDriver, p port.MikhmonProfileParams) (command.Result, error) {
	return u.gateway.CreateUserProfile(ctx, driver, p)
}

// GetProfiles fetches all Hotspot User Profiles.
func (u *UseCase) GetProfiles(ctx context.Context, driver port.DeviceDriver) ([]port.HotspotUserProfile, error) {
	return u.gateway.GetUserProfiles(ctx, driver)
}

// UpdateProfile updates an existing profile by RouterOS .id.
func (u *UseCase) UpdateProfile(ctx context.Context, driver port.DeviceDriver, rosID string, p port.MikhmonProfileParams) (command.Result, error) {
	return u.gateway.UpdateUserProfile(ctx, driver, rosID, p)
}

// DeleteProfile removes a user profile by RouterOS .id.
func (u *UseCase) DeleteProfile(ctx context.Context, driver port.DeviceDriver, rosID string) (command.Result, error) {
	return u.gateway.DeleteUserProfile(ctx, driver, rosID)
}
