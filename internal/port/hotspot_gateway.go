package port

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/command"
)

// HotspotGateway abstracts vendor-specific hotspot and voucher operations on
// a network device. It is implemented by the mikrotik driver package: each
// method constructs the vendor-native command, executes it through the
// injected CommandExecutor (which applies the policy gate), and parses the
// result into a vendor-neutral port type. Usecases depend on this interface
// instead of importing driver packages directly.
type HotspotGateway interface {
	// GetSystemResource fetches router system resource metrics.
	GetSystemResource(ctx context.Context, driver DeviceDriver) (SystemResource, error)
	// GetSystemIdentity fetches the router's configured identity (name).
	GetSystemIdentity(ctx context.Context, driver DeviceDriver) (string, error)
	// GetUserProfiles lists all hotspot user profiles.
	GetUserProfiles(ctx context.Context, driver DeviceDriver) ([]HotspotUserProfile, error)
	// CreateUserProfile creates a hotspot user profile with Mikhmon metadata.
	CreateUserProfile(ctx context.Context, driver DeviceDriver, p MikhmonProfileParams) (command.Result, error)
	// UpdateUserProfile updates an existing profile by RouterOS .id.
	UpdateUserProfile(ctx context.Context, driver DeviceDriver, rosID string, p MikhmonProfileParams) (command.Result, error)
	// DeleteUserProfile removes a user profile by RouterOS .id.
	DeleteUserProfile(ctx context.Context, driver DeviceDriver, rosID string) (command.Result, error)
	// GenerateVouchers generates a batch of count vouchers and executes their
	// creation commands, returning the generated voucher metadata.
	GenerateVouchers(ctx context.Context, driver DeviceDriver, p VoucherGenerateParams, count int) (VoucherBatch, error)
	// ListUsers lists hotspot users, optionally filtered by profile.
	ListUsers(ctx context.Context, driver DeviceDriver, profileFilter string) ([]HotspotUser, error)
	// GetUser fetches a single hotspot user by RouterOS .id.
	GetUser(ctx context.Context, driver DeviceDriver, rosID string) (HotspotUser, error)
	// AddUser creates a new hotspot user directly (non-batch).
	AddUser(ctx context.Context, driver DeviceDriver, p HotspotUserParams) (command.Result, error)
	// UpdateUser modifies an existing hotspot user by RouterOS .id.
	UpdateUser(ctx context.Context, driver DeviceDriver, rosID string, p HotspotUserParams) (command.Result, error)
	// RemoveUser deletes a hotspot user by RouterOS .id.
	RemoveUser(ctx context.Context, driver DeviceDriver, rosID string) (command.Result, error)
	// ResetUserCounters resets byte/time counters for a hotspot user.
	ResetUserCounters(ctx context.Context, driver DeviceDriver, rosID string) (command.Result, error)
	// ListActiveSessions fetches all currently connected hotspot active sessions.
	ListActiveSessions(ctx context.Context, driver DeviceDriver) ([]HotspotActiveSession, error)
	// RemoveActiveSession kicks an active session by its RouterOS .id.
	RemoveActiveSession(ctx context.Context, driver DeviceDriver, rosID string) (command.Result, error)
	// ListHosts fetches all /ip/hotspot/host entries.
	ListHosts(ctx context.Context, driver DeviceDriver) ([]map[string]string, error)
	// RemoveHost deletes a hotspot host entry by RouterOS .id.
	RemoveHost(ctx context.Context, driver DeviceDriver, rosID string) (command.Result, error)
	// ListHotspotServers fetches all /ip/hotspot/print entries.
	ListHotspotServers(ctx context.Context, driver DeviceDriver) ([]map[string]string, error)
	// ListIPPools fetches all /ip/pool entries.
	ListIPPools(ctx context.Context, driver DeviceDriver) ([]IPPool, error)
	// ListParentQueues fetches all parent /queue/simple entries.
	ListParentQueues(ctx context.Context, driver DeviceDriver) ([]SimpleQueue, error)
	// ListNATRules fetches all /ip/firewall/nat entries.
	ListNATRules(ctx context.Context, driver DeviceDriver) ([]map[string]string, error)
	// SetupExpireMonitor adds or updates the Mikhmon expire-monitor scheduler.
	SetupExpireMonitor(ctx context.Context, driver DeviceDriver, interval string) (command.Result, error)
	// ListReports fetches all report records from /system/script.
	ListReports(ctx context.Context, driver DeviceDriver) ([]MikhmonTransaction, error)
	// DeleteReport deletes a report script record by RouterOS .id.
	DeleteReport(ctx context.Context, driver DeviceDriver, rosID string) (command.Result, error)
	// ParseUserComment parses a Mikhmon-formatted user comment. Pure parse —
	// driver is unused and kept only for a uniform interface shape.
	ParseUserComment(comment string) (MikhmonComment, error)
}
