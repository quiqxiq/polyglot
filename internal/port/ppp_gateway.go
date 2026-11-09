package port

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/command"
	domainPPP "github.com/quixiq/polyglot/internal/domain/ppp"
)

// PPPGateway abstracts all vendor-specific PPP operations (Secrets, Profiles,
// Active Sessions, and Inactive calculations) on network devices.
type PPPGateway interface {
	// Secret Operations
	ListSecrets(ctx context.Context, driver DeviceDriver, nameFilter string) ([]domainPPP.PPPoESecret, error)
	GetSecret(ctx context.Context, driver DeviceDriver, rosID string) (domainPPP.PPPoESecret, error)
	AddSecret(ctx context.Context, driver DeviceDriver, p PPPoESecretParams) (command.Result, error)
	UpdateSecret(ctx context.Context, driver DeviceDriver, rosID string, p PPPoESecretParams) (command.Result, error)
	RemoveSecret(ctx context.Context, driver DeviceDriver, rosID string) (command.Result, error)
	SetSecretDisabled(ctx context.Context, driver DeviceDriver, rosID string, disabled bool) (command.Result, error)

	// Profile Operations
	ListProfiles(ctx context.Context, driver DeviceDriver, nameFilter string) ([]domainPPP.Profile, error)
	GetProfile(ctx context.Context, driver DeviceDriver, rosID string) (domainPPP.Profile, error)
	AddProfile(ctx context.Context, driver DeviceDriver, p PPPProfileParams) (command.Result, error)
	UpdateProfile(ctx context.Context, driver DeviceDriver, rosID string, p PPPProfileParams) (command.Result, error)
	RemoveProfile(ctx context.Context, driver DeviceDriver, rosID string) (command.Result, error)

	// Active / Inactive Operations
	ListActive(ctx context.Context, driver DeviceDriver, nameFilter string) ([]domainPPP.ActiveSession, error)
	KickActive(ctx context.Context, driver DeviceDriver, rosID string) (command.Result, error)
	ListInactive(ctx context.Context, driver DeviceDriver) ([]domainPPP.PPPoESecret, error)
}
