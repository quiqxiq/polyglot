package mikrotik

import (
	"context"

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
