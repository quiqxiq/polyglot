package ppp

import (
	"context"
	"fmt"

	"github.com/quixiq/polyglot/internal/domain/command"
	domainPPP "github.com/quixiq/polyglot/internal/domain/ppp"
	"github.com/quixiq/polyglot/internal/port"
)

// UseCase orchestrates all PPPoE / PPP operations (Secrets, Profiles, Active Sessions,
// and Inactive subscribers). It depends only on port.PPPGateway — all vendor-native
// command construction lives behind that seam.
type UseCase struct {
	gateway port.PPPGateway
}

// New creates a new UseCase instance.
func New(gateway port.PPPGateway) *UseCase {
	return &UseCase{gateway: gateway}
}

// ListSecrets retrieves all PPPoE secrets, optionally filtered by username.
func (u *UseCase) ListSecrets(ctx context.Context, driver port.DeviceDriver, nameFilter string) ([]port.PPPoESecret, error) {
	return u.gateway.ListSecrets(ctx, driver, nameFilter)
}

// GetSecret fetches a single PPPoE secret by its RouterOS ID.
func (u *UseCase) GetSecret(ctx context.Context, driver port.DeviceDriver, rosID string) (port.PPPoESecret, error) {
	if rosID == "" {
		return port.PPPoESecret{}, fmt.Errorf("%w: ros_id is required", domainPPP.ErrInvalidInput)
	}
	return u.gateway.GetSecret(ctx, driver, rosID)
}

// AddSecret creates a new PPPoE secret.
func (u *UseCase) AddSecret(ctx context.Context, driver port.DeviceDriver, p port.PPPoESecretParams) (command.Result, error) {
	if p.Name == "" {
		return command.Result{}, fmt.Errorf("%w: secret name (username) is required", domainPPP.ErrInvalidInput)
	}
	return u.gateway.AddSecret(ctx, driver, p)
}

// UpdateSecret modifies an existing PPPoE secret.
func (u *UseCase) UpdateSecret(ctx context.Context, driver port.DeviceDriver, rosID string, p port.PPPoESecretParams) (command.Result, error) {
	if rosID == "" {
		return command.Result{}, fmt.Errorf("%w: ros_id is required", domainPPP.ErrInvalidInput)
	}
	return u.gateway.UpdateSecret(ctx, driver, rosID, p)
}

// RemoveSecret deletes a PPPoE secret by its RouterOS ID.
func (u *UseCase) RemoveSecret(ctx context.Context, driver port.DeviceDriver, rosID string) (command.Result, error) {
	if rosID == "" {
		return command.Result{}, fmt.Errorf("%w: ros_id is required", domainPPP.ErrInvalidInput)
	}
	return u.gateway.RemoveSecret(ctx, driver, rosID)
}

// SetSecretDisabled enables or disables a PPPoE secret.
func (u *UseCase) SetSecretDisabled(ctx context.Context, driver port.DeviceDriver, rosID string, disabled bool) (command.Result, error) {
	if rosID == "" {
		return command.Result{}, fmt.Errorf("%w: ros_id is required", domainPPP.ErrInvalidInput)
	}
	return u.gateway.SetSecretDisabled(ctx, driver, rosID, disabled)
}

// ListProfiles retrieves all PPP profiles, optionally filtered by name.
func (u *UseCase) ListProfiles(ctx context.Context, driver port.DeviceDriver, nameFilter string) ([]port.PPPProfile, error) {
	return u.gateway.ListProfiles(ctx, driver, nameFilter)
}

// GetProfile fetches a single PPP profile by its RouterOS ID.
func (u *UseCase) GetProfile(ctx context.Context, driver port.DeviceDriver, rosID string) (port.PPPProfile, error) {
	if rosID == "" {
		return port.PPPProfile{}, fmt.Errorf("%w: ros_id is required", domainPPP.ErrInvalidInput)
	}
	return u.gateway.GetProfile(ctx, driver, rosID)
}

// AddProfile creates a new PPP profile.
func (u *UseCase) AddProfile(ctx context.Context, driver port.DeviceDriver, p port.PPPProfileParams) (command.Result, error) {
	if p.Name == "" {
		return command.Result{}, fmt.Errorf("%w: profile name is required", domainPPP.ErrInvalidInput)
	}
	return u.gateway.AddProfile(ctx, driver, p)
}

// UpdateProfile modifies an existing PPP profile.
func (u *UseCase) UpdateProfile(ctx context.Context, driver port.DeviceDriver, rosID string, p port.PPPProfileParams) (command.Result, error) {
	if rosID == "" {
		return command.Result{}, fmt.Errorf("%w: ros_id is required", domainPPP.ErrInvalidInput)
	}
	return u.gateway.UpdateProfile(ctx, driver, rosID, p)
}

// RemoveProfile deletes a PPP profile by its RouterOS ID.
func (u *UseCase) RemoveProfile(ctx context.Context, driver port.DeviceDriver, rosID string) (command.Result, error) {
	if rosID == "" {
		return command.Result{}, fmt.Errorf("%w: ros_id is required", domainPPP.ErrInvalidInput)
	}
	return u.gateway.RemoveProfile(ctx, driver, rosID)
}

// ListActive retrieves currently active PPP sessions.
func (u *UseCase) ListActive(ctx context.Context, driver port.DeviceDriver, nameFilter string) ([]port.PPPActiveSession, error) {
	return u.gateway.ListActive(ctx, driver, nameFilter)
}

// KickActive forcibly disconnects an active PPP session by its RouterOS ID.
func (u *UseCase) KickActive(ctx context.Context, driver port.DeviceDriver, rosID string) (command.Result, error) {
	if rosID == "" {
		return command.Result{}, fmt.Errorf("%w: ros_id is required", domainPPP.ErrInvalidInput)
	}
	return u.gateway.KickActive(ctx, driver, rosID)
}

// KickActiveBatch forcibly disconnects multiple active PPP sessions by their RouterOS IDs.
func (u *UseCase) KickActiveBatch(ctx context.Context, driver port.DeviceDriver, rosIDs []string) (int, error) {
	count := 0
	for _, id := range rosIDs {
		if id == "" {
			continue
		}
		if _, err := u.gateway.KickActive(ctx, driver, id); err == nil {
			count++
		}
	}
	return count, nil
}

// ListInactive calculates and returns offline subscribers (secrets without active sessions).
func (u *UseCase) ListInactive(ctx context.Context, driver port.DeviceDriver) ([]port.PPPoESecret, error) {
	return u.gateway.ListInactive(ctx, driver)
}
