package mikrotik

import (
	"context"
	"fmt"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/port"
)

// Gateway implements port.SessionGateway and port.DeviceDiagnostics for
// MikroTik RouterOS. It holds all vendor-native command construction and
// result parsing knowledge and executes commands through the injected
// CommandExecutor so the policy gate is applied exactly once per command.
// Stateless: the per-request device connection is passed in as
// port.DeviceDriver.
type Gateway struct {
	exec port.CommandExecutor
}

// NewGateway creates a Gateway bound to exec. exec must be the
// policy-gated executor (usecase/network.ExecuteCommand) — a bare
// driver.Execute here would silently bypass destructive-command approval.
func NewGateway(exec port.CommandExecutor) *Gateway {
	return &Gateway{exec: exec}
}

var _ port.SessionGateway = (*Gateway)(nil)
var _ port.DeviceDiagnostics = (*Gateway)(nil)
var _ port.PPPGateway = (*Gateway)(nil)

// ListSecrets implements port.PPPGateway.
func (g *Gateway) ListSecrets(ctx context.Context, driver port.DeviceDriver, nameFilter string) ([]port.PPPoESecret, error) {
	res, err := g.exec(ctx, driver, NewPrintPPPoESecretsCommand(nameFilter))
	if err != nil {
		return nil, err
	}
	return ParsePPPoESecrets(res), nil
}

// GetSecret implements port.PPPGateway.
func (g *Gateway) GetSecret(ctx context.Context, driver port.DeviceDriver, rosID string) (port.PPPoESecret, error) {
	cmd := command.Command{
		Raw:  "/ppp/secret/print",
		Args: map[string]string{"?.id": rosID},
	}
	res, err := g.exec(ctx, driver, cmd)
	if err != nil {
		return port.PPPoESecret{}, err
	}
	secrets := ParsePPPoESecrets(res)
	if len(secrets) == 0 {
		return port.PPPoESecret{}, fmt.Errorf("ppp secret %q not found", rosID)
	}
	return secrets[0], nil
}

// AddSecret implements port.PPPGateway.
func (g *Gateway) AddSecret(ctx context.Context, driver port.DeviceDriver, p port.PPPoESecretParams) (command.Result, error) {
	return g.exec(ctx, driver, NewAddPPPoESecretCommand(p))
}

// UpdateSecret implements port.PPPGateway.
func (g *Gateway) UpdateSecret(ctx context.Context, driver port.DeviceDriver, rosID string, p port.PPPoESecretParams) (command.Result, error) {
	return g.exec(ctx, driver, NewSetPPPoESecretCommand(rosID, p))
}

// RemoveSecret implements port.PPPGateway.
func (g *Gateway) RemoveSecret(ctx context.Context, driver port.DeviceDriver, rosID string) (command.Result, error) {
	return g.exec(ctx, driver, NewRemovePPPoESecretCommand(rosID))
}

// SetSecretDisabled implements port.PPPGateway.
func (g *Gateway) SetSecretDisabled(ctx context.Context, driver port.DeviceDriver, rosID string, disabled bool) (command.Result, error) {
	val := "no"
	if disabled {
		val = "yes"
	}
	cmd := command.Command{
		Raw:  "/ppp/secret/set",
		Args: map[string]string{"numbers": rosID, "disabled": val},
	}
	return g.exec(ctx, driver, cmd)
}

// ListProfiles implements port.PPPGateway.
func (g *Gateway) ListProfiles(ctx context.Context, driver port.DeviceDriver, nameFilter string) ([]port.PPPProfile, error) {
	res, err := g.exec(ctx, driver, NewPrintPPPProfilesCommand(nameFilter))
	if err != nil {
		return nil, err
	}
	return ParsePPPProfiles(res), nil
}

// GetProfile implements port.PPPGateway.
func (g *Gateway) GetProfile(ctx context.Context, driver port.DeviceDriver, rosID string) (port.PPPProfile, error) {
	cmd := command.Command{
		Raw:  "/ppp/profile/print",
		Args: map[string]string{"?.id": rosID},
	}
	res, err := g.exec(ctx, driver, cmd)
	if err != nil {
		return port.PPPProfile{}, err
	}
	profiles := ParsePPPProfiles(res)
	if len(profiles) == 0 {
		return port.PPPProfile{}, fmt.Errorf("ppp profile %q not found", rosID)
	}
	return profiles[0], nil
}

// AddProfile implements port.PPPGateway.
func (g *Gateway) AddProfile(ctx context.Context, driver port.DeviceDriver, p port.PPPProfileParams) (command.Result, error) {
	return g.exec(ctx, driver, NewAddPPPProfileCommand(p))
}

// UpdateProfile implements port.PPPGateway.
func (g *Gateway) UpdateProfile(ctx context.Context, driver port.DeviceDriver, rosID string, p port.PPPProfileParams) (command.Result, error) {
	return g.exec(ctx, driver, NewSetPPPProfileCommand(rosID, p))
}

// RemoveProfile implements port.PPPGateway.
func (g *Gateway) RemoveProfile(ctx context.Context, driver port.DeviceDriver, rosID string) (command.Result, error) {
	return g.exec(ctx, driver, NewRemovePPPProfileCommand(rosID))
}

// ListActive implements port.PPPGateway.
func (g *Gateway) ListActive(ctx context.Context, driver port.DeviceDriver, nameFilter string) ([]port.PPPActiveSession, error) {
	res, err := g.exec(ctx, driver, NewPrintPPPActiveCommand(nameFilter))
	if err != nil {
		return nil, err
	}
	return ParsePPPActiveSessions(res), nil
}

// KickActive implements port.PPPGateway.
func (g *Gateway) KickActive(ctx context.Context, driver port.DeviceDriver, rosID string) (command.Result, error) {
	return g.exec(ctx, driver, NewKickPPPActiveCommand(rosID))
}

// ListInactive implements port.PPPGateway.
func (g *Gateway) ListInactive(ctx context.Context, driver port.DeviceDriver) ([]port.PPPoESecret, error) {
	return g.ListPPPInactive(ctx, driver)
}

// ListPPPActive implements port.SessionGateway.
func (g *Gateway) ListPPPActive(ctx context.Context, driver port.DeviceDriver) ([]port.PPPActiveSession, error) {
	res, err := g.exec(ctx, driver, NewPrintPPPActiveCommand(""))
	if err != nil {
		return nil, err
	}
	return ParsePPPActiveSessions(res), nil
}

// ListPPPInactive implements port.SessionGateway.
func (g *Gateway) ListPPPInactive(ctx context.Context, driver port.DeviceDriver) ([]port.PPPoESecret, error) {
	secRes, err := g.exec(ctx, driver, NewPrintPPPoESecretsCommand(""))
	if err != nil {
		return nil, err
	}
	secrets := ParsePPPoESecrets(secRes)

	actRes, err := g.exec(ctx, driver, NewPrintPPPActiveCommand(""))
	if err != nil {
		return nil, err
	}
	active := ParsePPPActiveSessions(actRes)

	return FilterInactivePPPoESecrets(secrets, active), nil
}

// KickPPPSession implements port.SessionGateway.
func (g *Gateway) KickPPPSession(ctx context.Context, driver port.DeviceDriver, rosID string) (command.Result, error) {
	return g.exec(ctx, driver, NewKickPPPActiveCommand(rosID))
}

// ListHotspotActive implements port.SessionGateway.
func (g *Gateway) ListHotspotActive(ctx context.Context, driver port.DeviceDriver) ([]port.HotspotActiveSession, error) {
	res, err := g.exec(ctx, driver, NewPrintHotspotActiveCommand(""))
	if err != nil {
		return nil, err
	}
	return ParseHotspotActiveSessions(res), nil
}

// ListHotspotInactiveUsers implements port.SessionGateway.
func (g *Gateway) ListHotspotInactiveUsers(ctx context.Context, driver port.DeviceDriver) ([]port.HotspotUser, error) {
	uRes, err := g.exec(ctx, driver, NewPrintHotspotUsersCommand(""))
	if err != nil {
		return nil, err
	}
	users := ParseHotspotUsers(uRes)

	actRes, err := g.exec(ctx, driver, NewPrintHotspotActiveCommand(""))
	if err != nil {
		return nil, err
	}
	active := ParseHotspotActiveSessions(actRes)

	return FilterInactiveHotspotUsers(users, active), nil
}

// KickHotspotSession implements port.SessionGateway.
func (g *Gateway) KickHotspotSession(ctx context.Context, driver port.DeviceDriver, rosID string) (command.Result, error) {
	return g.exec(ctx, driver, NewDisconnectHotspotActiveCommand(rosID))
}

// ListDHCPLeases implements port.SessionGateway.
func (g *Gateway) ListDHCPLeases(ctx context.Context, driver port.DeviceDriver, macFilter string) ([]port.DHCPLease, error) {
	res, err := g.exec(ctx, driver, NewPrintDHCPLeasesCommand(macFilter))
	if err != nil {
		return nil, err
	}
	return ParseDHCPLeases(res), nil
}

// SetDHCPLeaseBlock implements port.SessionGateway.
func (g *Gateway) SetDHCPLeaseBlock(ctx context.Context, driver port.DeviceDriver, rosID string, blocked bool, comment string) (command.Result, error) {
	cmd := NewSetDHCPLeaseBlockCommand(rosID, DHCPLeaseBlockParams{
		Blocked: blocked,
		Comment: comment,
	})
	return g.exec(ctx, driver, cmd)
}

// GetSystemResource implements port.DeviceDiagnostics.
func (g *Gateway) GetSystemResource(ctx context.Context, driver port.DeviceDriver) (port.SystemResource, error) {
	res, err := g.exec(ctx, driver, NewPrintSystemResourceCommand())
	if err != nil {
		return port.SystemResource{}, err
	}
	return ParseSystemResource(res), nil
}

// GetSystemIdentity implements port.DeviceDiagnostics.
func (g *Gateway) GetSystemIdentity(ctx context.Context, driver port.DeviceDriver) (string, error) {
	res, err := g.exec(ctx, driver, NewPrintSystemIdentityCommand())
	if err != nil {
		return "", err
	}
	return ParseSystemIdentity(res).Name, nil
}

// ListInterfaces implements port.DeviceDiagnostics.
func (g *Gateway) ListInterfaces(ctx context.Context, driver port.DeviceDriver) ([]port.Interface, error) {
	res, err := g.exec(ctx, driver, NewPrintInterfacesCommand(""))
	if err != nil {
		return nil, err
	}
	return ParseInterfaces(res), nil
}

// MonitorTrafficOnce implements port.DeviceDiagnostics.
func (g *Gateway) MonitorTrafficOnce(ctx context.Context, driver port.DeviceDriver, ifaceName string) (port.InterfaceTrafficStats, error) {
	res, err := g.exec(ctx, driver, NewMonitorTrafficOnceCommand(ifaceName))
	if err != nil {
		return port.InterfaceTrafficStats{}, err
	}
	return ParseInterfaceTrafficStats(res), nil
}
