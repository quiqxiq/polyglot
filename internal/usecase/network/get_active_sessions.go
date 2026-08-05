package network

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/command"
	// DEVIASI: Use case jaringan mengimpor driver/mikrotik khusus untuk
	// mengkonstruksi command native MikroTik yang kemudian dieksekusi via
	// interface port.DeviceDriver.
	"github.com/quixiq/polyglot/internal/driver/mikrotik"
	"github.com/quixiq/polyglot/internal/port"
)

// ActiveSessionsUseCase handles monitoring of active/inactive sessions (PPPoE, Hotspot, DHCP).
type ActiveSessionsUseCase struct{}

// NewActiveSessionsUseCase creates a new ActiveSessionsUseCase instance.
func NewActiveSessionsUseCase() *ActiveSessionsUseCase {
	return &ActiveSessionsUseCase{}
}

// GetPPPActiveSessions fetches all active PPPoE sessions from the router.
func (u *ActiveSessionsUseCase) GetPPPActiveSessions(ctx context.Context, driver port.DeviceDriver) ([]mikrotik.PPPActiveSession, error) {
	cmd := mikrotik.NewPrintPPPActiveCommand("")
	res, err := ExecuteCommand(ctx, driver, cmd)
	if err != nil {
		return nil, err
	}
	return mikrotik.ParsePPPActiveSessions(res), nil
}

// GetPPPInactiveSessions fetches all offline PPPoE subscriber secrets from the router.
func (u *ActiveSessionsUseCase) GetPPPInactiveSessions(ctx context.Context, driver port.DeviceDriver) ([]mikrotik.PPPoESecret, error) {
	secCmd := mikrotik.NewPrintPPPoESecretsCommand("")
	secRes, err := ExecuteCommand(ctx, driver, secCmd)
	if err != nil {
		return nil, err
	}
	secrets := mikrotik.ParsePPPoESecrets(secRes)

	actCmd := mikrotik.NewPrintPPPActiveCommand("")
	actRes, err := ExecuteCommand(ctx, driver, actCmd)
	if err != nil {
		return nil, err
	}
	active := mikrotik.ParsePPPActiveSessions(actRes)

	return mikrotik.FilterInactivePPPoESecrets(secrets, active), nil
}

// KickPPPSession forcibly disconnects an active PPPoE session by RouterOS .id.
func (u *ActiveSessionsUseCase) KickPPPSession(ctx context.Context, driver port.DeviceDriver, rosID string) (command.Result, error) {
	cmd := mikrotik.NewKickPPPActiveCommand(rosID)
	return ExecuteCommand(ctx, driver, cmd)
}

// GetHotspotActiveSessions fetches all active Hotspot sessions from the router.
func (u *ActiveSessionsUseCase) GetHotspotActiveSessions(ctx context.Context, driver port.DeviceDriver) ([]mikrotik.HotspotActiveSession, error) {
	cmd := mikrotik.NewPrintHotspotActiveCommand("")
	res, err := ExecuteCommand(ctx, driver, cmd)
	if err != nil {
		return nil, err
	}
	return mikrotik.ParseHotspotActiveSessions(res), nil
}

// GetHotspotInactiveUsers fetches all offline Hotspot users from the router.
func (u *ActiveSessionsUseCase) GetHotspotInactiveUsers(ctx context.Context, driver port.DeviceDriver) ([]mikrotik.HotspotUser, error) {
	uCmd := mikrotik.NewPrintHotspotUsersCommand("")
	uRes, err := ExecuteCommand(ctx, driver, uCmd)
	if err != nil {
		return nil, err
	}
	users := mikrotik.ParseHotspotUsers(uRes)

	actCmd := mikrotik.NewPrintHotspotActiveCommand("")
	actRes, err := ExecuteCommand(ctx, driver, actCmd)
	if err != nil {
		return nil, err
	}
	active := mikrotik.ParseHotspotActiveSessions(actRes)

	return mikrotik.FilterInactiveHotspotUsers(users, active), nil
}

// KickHotspotSession forcibly disconnects an active Hotspot session by RouterOS .id.
func (u *ActiveSessionsUseCase) KickHotspotSession(ctx context.Context, driver port.DeviceDriver, rosID string) (command.Result, error) {
	cmd := mikrotik.NewDisconnectHotspotActiveCommand(rosID)
	return ExecuteCommand(ctx, driver, cmd)
}

// GetDHCPLeases fetches all DHCP server leases from the router.
func (u *ActiveSessionsUseCase) GetDHCPLeases(ctx context.Context, driver port.DeviceDriver, macFilter string) ([]mikrotik.DHCPLease, error) {
	cmd := mikrotik.NewPrintDHCPLeasesCommand(macFilter)
	res, err := ExecuteCommand(ctx, driver, cmd)
	if err != nil {
		return nil, err
	}
	return mikrotik.ParseDHCPLeases(res), nil
}

// SetDHCPLeaseBlock blocks or unblocks a DHCP lease by RouterOS .id.
func (u *ActiveSessionsUseCase) SetDHCPLeaseBlock(ctx context.Context, driver port.DeviceDriver, rosID string, blocked bool, comment string) (command.Result, error) {
	cmd := mikrotik.NewSetDHCPLeaseBlockCommand(rosID, mikrotik.DHCPLeaseBlockParams{
		Blocked: blocked,
		Comment: comment,
	})
	return ExecuteCommand(ctx, driver, cmd)
}
