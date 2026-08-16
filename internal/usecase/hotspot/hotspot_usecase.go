package hotspot

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/port"
)

// DashboardSummary holds aggregated statistics for the Hotspot dashboard.
type DashboardSummary struct {
	Uptime      string `json:"uptime"`
	Version     string `json:"version"`
	CPULoad     int    `json:"cpu_load"`
	BoardName   string `json:"board_name"`
	Identity    string `json:"identity"`
	TotalUsers  int    `json:"total_users"`
	ActiveUsers int    `json:"active_users"`
	TodayIncome int64  `json:"today_income"`
}

// UseCase orchestrates all Hotspot and Voucher operations. It depends only
// on port.HotspotGateway — all vendor-native command construction and
// parsing knowledge lives in the driver implementation behind that seam.
type UseCase struct {
	TemplateDir string
	gateway     port.HotspotGateway
}

// New creates a new UseCase instance. gateway must be non-nil.
func New(templateDir string, gateway port.HotspotGateway) *UseCase {
	if templateDir == "" {
		templateDir = "internal/template"
	}
	return &UseCase{TemplateDir: templateDir, gateway: gateway}
}

// GetSystemResource retrieves system resource metrics from the device.
func (u *UseCase) GetSystemResource(ctx context.Context, driver port.DeviceDriver) (port.SystemResource, error) {
	return u.gateway.GetSystemResource(ctx, driver)
}

// CreateProfile builds and executes the /ip/hotspot/user/profile/add command
func (u *UseCase) CreateProfile(ctx context.Context, driver port.DeviceDriver, p port.MikhmonProfileParams) (command.Result, error) {
	return u.gateway.CreateUserProfile(ctx, driver, p)
}

// GetProfiles fetches all Hotspot User Profiles.
func (u *UseCase) GetProfiles(ctx context.Context, driver port.DeviceDriver) ([]port.HotspotUserProfile, error) {
	return u.gateway.GetUserProfiles(ctx, driver)
}

// GenerateVouchers generates a batch of count vouchers and executes their creation commands.
func (u *UseCase) GenerateVouchers(ctx context.Context, driver port.DeviceDriver, p port.VoucherGenerateParams, count int) (port.VoucherBatch, error) {
	return u.gateway.GenerateVouchers(ctx, driver, p, count)
}

// GetUsers lists hotspot users, optionally filtered by profile.
func (u *UseCase) GetUsers(ctx context.Context, driver port.DeviceDriver, profileFilter string) ([]port.HotspotUser, error) {
	return u.gateway.ListUsers(ctx, driver, profileFilter)
}

// GetUser fetches a single hotspot user by RouterOS .id.
func (u *UseCase) GetUser(ctx context.Context, driver port.DeviceDriver, rosID string) (port.HotspotUser, error) {
	return u.gateway.GetUser(ctx, driver, rosID)
}

// RemoveUser deletes a hotspot user by RouterOS .id.
func (u *UseCase) RemoveUser(ctx context.Context, driver port.DeviceDriver, rosID string) (command.Result, error) {
	return u.gateway.RemoveUser(ctx, driver, rosID)
}

// ResetUserCounters resets byte/time counters for a hotspot user by RouterOS .id.
func (u *UseCase) ResetUserCounters(ctx context.Context, driver port.DeviceDriver, rosID string) (command.Result, error) {
	return u.gateway.ResetUserCounters(ctx, driver, rosID)
}

// GetActiveSessions fetches all currently connected hotspot active sessions.
func (u *UseCase) GetActiveSessions(ctx context.Context, driver port.DeviceDriver) ([]port.HotspotActiveSession, error) {
	return u.gateway.ListActiveSessions(ctx, driver)
}

// RemoveActiveSession kicks an active session by its RouterOS .id.
func (u *UseCase) RemoveActiveSession(ctx context.Context, driver port.DeviceDriver, rosID string) (command.Result, error) {
	return u.gateway.RemoveActiveSession(ctx, driver, rosID)
}

// GetHosts fetches all /ip/hotspot/host entries.
func (u *UseCase) GetHosts(ctx context.Context, driver port.DeviceDriver) ([]map[string]string, error) {
	return u.gateway.ListHosts(ctx, driver)
}

// RemoveHost deletes a hotspot host entry by RouterOS .id.
func (u *UseCase) RemoveHost(ctx context.Context, driver port.DeviceDriver, rosID string) (command.Result, error) {
	return u.gateway.RemoveHost(ctx, driver, rosID)
}

// GetHotspotServers fetches all /ip/hotspot/print entries.
func (u *UseCase) GetHotspotServers(ctx context.Context, driver port.DeviceDriver) ([]map[string]string, error) {
	return u.gateway.ListHotspotServers(ctx, driver)
}

// GetIPPools fetches all /ip/pool entries.
func (u *UseCase) GetIPPools(ctx context.Context, driver port.DeviceDriver) ([]port.IPPool, error) {
	return u.gateway.ListIPPools(ctx, driver)
}

// GetParentQueues fetches all parent /queue/simple entries.
func (u *UseCase) GetParentQueues(ctx context.Context, driver port.DeviceDriver) ([]port.SimpleQueue, error) {
	return u.gateway.ListParentQueues(ctx, driver)
}

// GetNATRules fetches all /ip/firewall/nat entries.
func (u *UseCase) GetNATRules(ctx context.Context, driver port.DeviceDriver) ([]map[string]string, error) {
	return u.gateway.ListNATRules(ctx, driver)
}

// SetupExpireMonitor adds or updates the Mikhmon expire-monitor scheduler.
func (u *UseCase) SetupExpireMonitor(ctx context.Context, driver port.DeviceDriver, interval string) (command.Result, error) {
	return u.gateway.SetupExpireMonitor(ctx, driver, interval)
}

// GetReports fetches all report records from /system/script/job or /log.
func (u *UseCase) GetReports(ctx context.Context, driver port.DeviceDriver) ([]port.MikhmonTransaction, error) {
	return u.gateway.ListReports(ctx, driver)
}

// DeleteReport deletes a report script record by RouterOS .id.
func (u *UseCase) DeleteReport(ctx context.Context, driver port.DeviceDriver, rosID string) (command.Result, error) {
	return u.gateway.DeleteReport(ctx, driver, rosID)
}

// GetTodayIncome calculates total income from report entries matching today's date.
func (u *UseCase) GetTodayIncome(ctx context.Context, driver port.DeviceDriver) (int64, error) {
	reports, err := u.GetReports(ctx, driver)
	if err != nil {
		return 0, err
	}

	today := time.Now().Format("jan/02/2006")
	todayLower := strings.ToLower(today)

	var total int64
	for _, r := range reports {
		if strings.ToLower(r.Date) == todayLower {
			priceInt, _ := strconv.ParseInt(r.Price, 10, 64)
			total += priceInt
		}
	}
	return total, nil
}

// UpdateProfile updates an existing profile by RouterOS .id.
func (u *UseCase) UpdateProfile(ctx context.Context, driver port.DeviceDriver, rosID string, p port.MikhmonProfileParams) (command.Result, error) {
	return u.gateway.UpdateUserProfile(ctx, driver, rosID, p)
}

// DeleteProfile removes a user profile by RouterOS .id.
func (u *UseCase) DeleteProfile(ctx context.Context, driver port.DeviceDriver, rosID string) (command.Result, error) {
	return u.gateway.DeleteUserProfile(ctx, driver, rosID)
}

// AddUser creates a new hotspot user directly (non-batch).
func (u *UseCase) AddUser(ctx context.Context, driver port.DeviceDriver, p port.HotspotUserParams) (command.Result, error) {
	res, err := u.gateway.AddUser(ctx, driver, p)
	if err != nil {
		return res, err
	}
	if p.Comment != "" {
		_, _ = u.gateway.ParseUserComment(p.Comment) // validasi komentar best-effort, hasil tidak dipakai
	}
	return res, nil
}

// UpdateUser modifies an existing hotspot user by RouterOS .id.
func (u *UseCase) UpdateUser(ctx context.Context, driver port.DeviceDriver, rosID string, p port.HotspotUserParams) (command.Result, error) {
	return u.gateway.UpdateUser(ctx, driver, rosID, p)
}

// GetUsersByTag fetches all hotspot users whose comment contains tag.
func (u *UseCase) GetUsersByTag(ctx context.Context, driver port.DeviceDriver, tag string) ([]port.HotspotUser, error) {
	users, err := u.GetUsers(ctx, driver, "")
	if err != nil {
		return nil, err
	}

	filtered := make([]port.HotspotUser, 0)
	for _, usr := range users {
		if usr.Comment == "" {
			continue
		}
		parsed, parseErr := u.gateway.ParseUserComment(usr.Comment)
		if parseErr != nil {
			continue
		}
		if tag == "" || strings.EqualFold(parsed.Tag, tag) {
			filtered = append(filtered, usr)
		}
	}
	return filtered, nil
}

// GetReportsByFilter filters transaction reports by date (e.g. "jan/15/2025"), month ("jan"), or year ("2025").
func (u *UseCase) GetReportsByFilter(ctx context.Context, driver port.DeviceDriver, date, month, year string) ([]port.MikhmonTransaction, error) {
	all, err := u.GetReports(ctx, driver)
	if err != nil {
		return nil, err
	}
	if date == "" && month == "" && year == "" {
		return all, nil
	}
	filtered := make([]port.MikhmonTransaction, 0)
	for _, r := range all {
		if date != "" && !strings.HasPrefix(r.Date, date) {
			continue
		}
		if month != "" && !strings.Contains(strings.ToLower(r.Date), strings.ToLower(month)) {
			continue
		}
		if year != "" && !strings.HasSuffix(r.Date, year) {
			continue
		}
		filtered = append(filtered, r)
	}
	return filtered, nil
}

// GetDashboardSummary aggregates system info, active users, total users, and today's income.
func (u *UseCase) GetDashboardSummary(ctx context.Context, driver port.DeviceDriver) (DashboardSummary, error) {
	summary := DashboardSummary{}

	// System Resource
	if sysRes, err := u.gateway.GetSystemResource(ctx, driver); err == nil {
		summary.Uptime = sysRes.Uptime
		summary.Version = sysRes.Version
		summary.CPULoad = sysRes.CPULoad
		summary.BoardName = sysRes.BoardName
	}

	// System Identity
	if ident, err := u.gateway.GetSystemIdentity(ctx, driver); err == nil {
		summary.Identity = ident
	}

	// Active Hotspot Users
	if active, err := u.gateway.ListActiveSessions(ctx, driver); err == nil {
		summary.ActiveUsers = len(active)
	}

	// Total Hotspot Users
	if users, err := u.gateway.ListUsers(ctx, driver, ""); err == nil {
		summary.TotalUsers = len(users)
	}

	// Today's Income
	if inc, err := u.GetTodayIncome(ctx, driver); err == nil {
		summary.TodayIncome = inc
	}

	return summary, nil
}
