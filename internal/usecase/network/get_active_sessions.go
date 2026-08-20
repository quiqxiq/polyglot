package network

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/port"
)

// ActiveSessionsUseCase handles monitoring of active/inactive sessions (PPPoE, Hotspot, DHCP).
// It depends only on port.SessionGateway — vendor-native command construction
// and parsing lives in the driver implementation behind that seam.
type ActiveSessionsUseCase struct {
	gateway port.SessionGateway
}

// NewActiveSessionsUseCase creates a new ActiveSessionsUseCase instance.
func NewActiveSessionsUseCase(gateway port.SessionGateway) *ActiveSessionsUseCase {
	return &ActiveSessionsUseCase{gateway: gateway}
}

// GetPPPActiveSessions fetches all active PPPoE sessions from the router.
func (u *ActiveSessionsUseCase) GetPPPActiveSessions(ctx context.Context, driver port.DeviceDriver) ([]port.PPPActiveSession, error) {
	return u.gateway.ListPPPActive(ctx, driver)
}

// GetPPPInactiveSessions fetches all offline PPPoE subscriber secrets from the router.
func (u *ActiveSessionsUseCase) GetPPPInactiveSessions(ctx context.Context, driver port.DeviceDriver) ([]port.PPPoESecret, error) {
	return u.gateway.ListPPPInactive(ctx, driver)
}

// KickPPPSession forcibly disconnects an active PPPoE session by RouterOS .id.
func (u *ActiveSessionsUseCase) KickPPPSession(ctx context.Context, driver port.DeviceDriver, rosID string) (command.Result, error) {
	return u.gateway.KickPPPSession(ctx, driver, rosID)
}

// GetHotspotActiveSessions fetches all active Hotspot sessions from the router.
func (u *ActiveSessionsUseCase) GetHotspotActiveSessions(ctx context.Context, driver port.DeviceDriver) ([]port.HotspotActiveSession, error) {
	return u.gateway.ListHotspotActive(ctx, driver)
}

// GetHotspotInactiveUsers fetches all offline Hotspot users from the router.
func (u *ActiveSessionsUseCase) GetHotspotInactiveUsers(ctx context.Context, driver port.DeviceDriver) ([]port.HotspotUser, error) {
	return u.gateway.ListHotspotInactiveUsers(ctx, driver)
}

// KickHotspotSession forcibly disconnects an active Hotspot session by RouterOS .id.
func (u *ActiveSessionsUseCase) KickHotspotSession(ctx context.Context, driver port.DeviceDriver, rosID string) (command.Result, error) {
	return u.gateway.KickHotspotSession(ctx, driver, rosID)
}

// GetDHCPLeases fetches all DHCP server leases from the router.
func (u *ActiveSessionsUseCase) GetDHCPLeases(ctx context.Context, driver port.DeviceDriver, macFilter string) ([]port.DHCPLease, error) {
	return u.gateway.ListDHCPLeases(ctx, driver, macFilter)
}

// SetDHCPLeaseBlock blocks or unblocks a DHCP lease by RouterOS .id.
func (u *ActiveSessionsUseCase) SetDHCPLeaseBlock(ctx context.Context, driver port.DeviceDriver, rosID string, blocked bool, comment string) (command.Result, error) {
	return u.gateway.SetDHCPLeaseBlock(ctx, driver, rosID, blocked, comment)
}
