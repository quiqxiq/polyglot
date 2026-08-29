package hotspot

import (
	"context"
	"fmt"
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

// GetSystemIdentity retrieves the router's configured identity (name).
func (u *UseCase) GetSystemIdentity(ctx context.Context, driver port.DeviceDriver) (string, error) {
	return u.gateway.GetSystemIdentity(ctx, driver)
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

// GetExpireMonitorStatus reports the install/enabled state of the expire
// monitor scheduler.
func (u *UseCase) GetExpireMonitorStatus(ctx context.Context, driver port.DeviceDriver) (port.ExpireMonitorStatus, error) {
	return u.gateway.GetExpireMonitorStatus(ctx, driver)
}

// DisableExpireMonitor disables the found expire-monitor scheduler. It fails
// when the monitor is not installed.
func (u *UseCase) DisableExpireMonitor(ctx context.Context, driver port.DeviceDriver, disabled bool) (command.Result, error) {
	status, err := u.gateway.GetExpireMonitorStatus(ctx, driver)
	if err != nil {
		return command.Result{}, err
	}
	if !status.IsInstalled {
		return command.Result{}, fmt.Errorf("expire monitor is not installed")
	}
	return u.gateway.SetExpireMonitorDisabled(ctx, driver, status.SchedulerID, disabled)
}

// RemoveExpireMonitor deletes the found expire-monitor scheduler (the script
// is left behind). It fails when the monitor is not installed.
func (u *UseCase) RemoveExpireMonitor(ctx context.Context, driver port.DeviceDriver) (command.Result, error) {
	status, err := u.gateway.GetExpireMonitorStatus(ctx, driver)
	if err != nil {
		return command.Result{}, err
	}
	if !status.IsInstalled {
		return command.Result{}, fmt.Errorf("expire monitor is not installed")
	}
	return u.gateway.RemoveExpireMonitor(ctx, driver, status.SchedulerID)
}

// GetReports fetches all Mikhmon report records from /system/script.
func (u *UseCase) GetReports(ctx context.Context, driver port.DeviceDriver) ([]port.MikhmonTransaction, error) {
	return u.gateway.ListReports(ctx, driver, port.ReportFilter{})
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

// GetReportsByFilter delegates report filtering to the gateway using legacy
// Mikhmon semantics: date → ?source=, month (owner "aug2026") → ?owner=,
// year → date-suffix post-filter. All three empty returns every mikhmon record.
func (u *UseCase) GetReportsByFilter(ctx context.Context, driver port.DeviceDriver, date, month, year string) ([]port.MikhmonTransaction, error) {
	return u.gateway.ListReports(ctx, driver, port.ReportFilter{
		Day:   date,
		Month: month,
		Year:  year,
	})
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
	if users, err := u.gateway.ListUsers(ctx, driver, port.ListUsersFilter{}); err == nil {
		summary.TotalUsers = len(users)
	}

	// Today's Income
	if inc, err := u.GetTodayIncome(ctx, driver); err == nil {
		summary.TodayIncome = inc
	}

	return summary, nil
}

// GetIPBindings fetches all /ip/hotspot/ip-binding entries.
func (u *UseCase) GetIPBindings(ctx context.Context, driver port.DeviceDriver) ([]port.HotspotIPBinding, error) {
	return u.gateway.ListIPBindings(ctx, driver)
}

// CreateIPBinding creates a new /ip/hotspot/ip-binding entry.
func (u *UseCase) CreateIPBinding(ctx context.Context, driver port.DeviceDriver, p port.HotspotIPBindingParams) (command.Result, error) {
	return u.gateway.CreateIPBinding(ctx, driver, p)
}

// UpdateIPBinding updates an existing /ip/hotspot/ip-binding entry.
func (u *UseCase) UpdateIPBinding(ctx context.Context, driver port.DeviceDriver, rosID string, p port.HotspotIPBindingParams) (command.Result, error) {
	return u.gateway.UpdateIPBinding(ctx, driver, rosID, p)
}

// DeleteIPBinding removes an /ip/hotspot/ip-binding entry by RouterOS .id.
func (u *UseCase) DeleteIPBinding(ctx context.Context, driver port.DeviceDriver, rosID string) (command.Result, error) {
	return u.gateway.DeleteIPBinding(ctx, driver, rosID)
}

// GetCookies fetches all /ip/hotspot/cookie entries.
func (u *UseCase) GetCookies(ctx context.Context, driver port.DeviceDriver) ([]port.HotspotCookie, error) {
	return u.gateway.ListCookies(ctx, driver)
}

// DeleteCookie removes a cookie by RouterOS .id (or all cookies if rosID is empty or "all").
func (u *UseCase) DeleteCookie(ctx context.Context, driver port.DeviceDriver, rosID string) (command.Result, error) {
	return u.gateway.DeleteCookie(ctx, driver, rosID)
}
