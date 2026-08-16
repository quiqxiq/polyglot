package port

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/command"
)

// SessionGateway abstracts vendor-specific session monitoring operations
// (PPPoE, Hotspot, DHCP) on a network device. It is implemented by the
// mikrotik driver package and executed through the injected
// CommandExecutor so the policy gate is never bypassed.
type SessionGateway interface {
	// ListPPPActive fetches all active PPPoE sessions from the router.
	ListPPPActive(ctx context.Context, driver DeviceDriver) ([]PPPActiveSession, error)
	// ListPPPInactive fetches all offline PPPoE subscriber secrets.
	ListPPPInactive(ctx context.Context, driver DeviceDriver) ([]PPPoESecret, error)
	// KickPPPSession forcibly disconnects an active PPPoE session by RouterOS .id.
	KickPPPSession(ctx context.Context, driver DeviceDriver, rosID string) (command.Result, error)
	// ListHotspotActive fetches all active Hotspot sessions from the router.
	ListHotspotActive(ctx context.Context, driver DeviceDriver) ([]HotspotActiveSession, error)
	// ListHotspotInactiveUsers fetches all offline Hotspot users.
	ListHotspotInactiveUsers(ctx context.Context, driver DeviceDriver) ([]HotspotUser, error)
	// KickHotspotSession forcibly disconnects an active Hotspot session by RouterOS .id.
	KickHotspotSession(ctx context.Context, driver DeviceDriver, rosID string) (command.Result, error)
	// ListDHCPLeases fetches all DHCP server leases from the router.
	ListDHCPLeases(ctx context.Context, driver DeviceDriver, macFilter string) ([]DHCPLease, error)
	// SetDHCPLeaseBlock blocks or unblocks a DHCP lease by RouterOS .id.
	SetDHCPLeaseBlock(ctx context.Context, driver DeviceDriver, rosID string, blocked bool, comment string) (command.Result, error)
}
