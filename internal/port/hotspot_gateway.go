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
	// ListUsers lists hotspot users, optionally filtered (profile/comment/name/only_unused).
	ListUsers(ctx context.Context, driver DeviceDriver, f ListUsersFilter) ([]HotspotUser, error)
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
	// SetUserDisabled toggles the disabled flag of a hotspot user directly
	// (isolation flow). UpdateUser cannot express boolean clears because
	// empty fields are skipped by design.
	SetUserDisabled(ctx context.Context, driver DeviceDriver, rosID string, disabled bool) (command.Result, error)
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
	// DeleteUsersByFilter deletes hotspot users matching filter mode (by_profile, by_comment, expired).
	DeleteUsersByFilter(ctx context.Context, driver DeviceDriver, mode, value string) (int, error)
	// ListIPBindings fetches all /ip/hotspot/ip-binding entries.
	ListIPBindings(ctx context.Context, driver DeviceDriver) ([]HotspotIPBinding, error)
	// CreateIPBinding creates a new /ip/hotspot/ip-binding entry.
	CreateIPBinding(ctx context.Context, driver DeviceDriver, p HotspotIPBindingParams) (command.Result, error)
	// UpdateIPBinding updates an existing /ip/hotspot/ip-binding entry.
	UpdateIPBinding(ctx context.Context, driver DeviceDriver, rosID string, p HotspotIPBindingParams) (command.Result, error)
	// DeleteIPBinding removes an /ip/hotspot/ip-binding entry by RouterOS .id.
	DeleteIPBinding(ctx context.Context, driver DeviceDriver, rosID string) (command.Result, error)
	// ListCookies fetches all /ip/hotspot/cookie entries.
	ListCookies(ctx context.Context, driver DeviceDriver) ([]HotspotCookie, error)
	// DeleteCookie removes a cookie by RouterOS .id (or all cookies if rosID is empty or "all").
	DeleteCookie(ctx context.Context, driver DeviceDriver, rosID string) (command.Result, error)
	// SetupExpireMonitor adds or updates the Mikhmon expire-monitor scheduler.
	// It is idempotent and recognises both the legacy single-step scheduler
	// (Mikhmon-Expire-Monitor) and the gateway two-step form.
	SetupExpireMonitor(ctx context.Context, driver DeviceDriver, interval string) (command.Result, error)
	// GetExpireMonitorStatus reports the state of the expire-monitor scheduler
	// (recognises both the legacy "Mikhmon-Expire-Monitor" and the gateway
	// "mikhmon-expire-scheduler" names).
	GetExpireMonitorStatus(ctx context.Context, driver DeviceDriver) (ExpireMonitorStatus, error)
	// SetExpireMonitorDisabled toggles the scheduler disabled flag by RouterOS .id.
	SetExpireMonitorDisabled(ctx context.Context, driver DeviceDriver, rosID string, disabled bool) (command.Result, error)
	// RemoveExpireMonitor deletes the expire-monitor scheduler by RouterOS .id.
	RemoveExpireMonitor(ctx context.Context, driver DeviceDriver, rosID string) (command.Result, error)
	// ListReports fetches report records from /system/script filtered by f.
	ListReports(ctx context.Context, driver DeviceDriver, f ReportFilter) ([]MikhmonTransaction, error)
	// DeleteReport deletes a report script record by RouterOS .id.
	DeleteReport(ctx context.Context, driver DeviceDriver, rosID string) (command.Result, error)
	// ParseUserComment parses a Mikhmon-formatted user comment. Pure parse —
	// driver is unused and kept only for a uniform interface shape.
	ParseUserComment(comment string) (MikhmonComment, error)
}
