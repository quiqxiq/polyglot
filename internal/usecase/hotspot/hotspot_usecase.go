package hotspot

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/driver/mikrotik"
	mikhmon "github.com/quixiq/polyglot/internal/driver/mikrotik/hotspot"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/internal/usecase/network"
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

// UseCase orchestrates all Hotspot and Voucher operations.
type UseCase struct {
	TemplateDir string
}

// New creates a new UseCase instance.
func New(templateDir string) *UseCase {
	if templateDir == "" {
		templateDir = "internal/templates"
	}
	return &UseCase{TemplateDir: templateDir}
}

// GetSystemResource retrieves system resource metrics from the device.
func (u *UseCase) GetSystemResource(ctx context.Context, driver port.DeviceDriver) (mikrotik.SystemResource, error) {
	cmd := mikrotik.NewPrintSystemResourceCommand()
	res, err := network.ExecuteCommand(ctx, driver, cmd)
	if err != nil {
		return mikrotik.SystemResource{}, err
	}
	return mikrotik.ParseSystemResource(res), nil
}

// CreateProfile builds and executes the /ip/hotspot/user/profile/add command
func (u *UseCase) CreateProfile(ctx context.Context, driver port.DeviceDriver, p mikhmon.MikhmonProfileParams) (command.Result, error) {
	cmd := mikhmon.NewAddMikhmonProfileCommand(p)
	return network.ExecuteCommand(ctx, driver, cmd)
}

// GetProfiles fetches all Hotspot User Profiles.
func (u *UseCase) GetProfiles(ctx context.Context, driver port.DeviceDriver) ([]mikrotik.HotspotUserProfile, error) {
	cmd := mikrotik.NewPrintHotspotUserProfilesCommand("")
	res, err := network.ExecuteCommand(ctx, driver, cmd)
	if err != nil {
		return nil, err
	}
	return mikrotik.ParseHotspotUserProfiles(res), nil
}

// GenerateVouchers generates a batch of count vouchers and executes their creation commands.
func (u *UseCase) GenerateVouchers(ctx context.Context, driver port.DeviceDriver, p mikhmon.VoucherGenerateParams, count int) (mikhmon.VoucherBatch, error) {
	batch := mikhmon.NewGenerateVoucherBatchCommands(p, count)
	for _, cmd := range batch.Commands {
		if _, err := network.ExecuteCommand(ctx, driver, cmd); err != nil {
			return batch, fmt.Errorf("failed to create voucher: %w", err)
		}
	}
	return batch, nil
}

// GetUsers lists hotspot users, optionally filtered by profile.
func (u *UseCase) GetUsers(ctx context.Context, driver port.DeviceDriver, profileFilter string) ([]mikrotik.HotspotUser, error) {
	cmd := mikrotik.NewPrintHotspotUsersCommand(profileFilter)
	res, err := network.ExecuteCommand(ctx, driver, cmd)
	if err != nil {
		return nil, err
	}
	return mikrotik.ParseHotspotUsers(res), nil
}

// GetUser fetches a single hotspot user by RouterOS .id.
func (u *UseCase) GetUser(ctx context.Context, driver port.DeviceDriver, rosID string) (mikrotik.HotspotUser, error) {
	cmd := command.Command{
		Raw:  "/ip/hotspot/user/print",
		Args: map[string]string{"?.id": rosID},
	}
	res, err := network.ExecuteCommand(ctx, driver, cmd)
	if err != nil {
		return mikrotik.HotspotUser{}, err
	}
	users := mikrotik.ParseHotspotUsers(res)
	if len(users) == 0 {
		return mikrotik.HotspotUser{}, fmt.Errorf("user %q not found", rosID)
	}
	return users[0], nil
}

// RemoveUser deletes a hotspot user by RouterOS .id.
func (u *UseCase) RemoveUser(ctx context.Context, driver port.DeviceDriver, rosID string) (command.Result, error) {
	cmd := mikrotik.NewRemoveHotspotUserCommand(rosID)
	return network.ExecuteCommand(ctx, driver, cmd)
}

// ResetUserCounters resets byte/time counters for a hotspot user by RouterOS .id.
func (u *UseCase) ResetUserCounters(ctx context.Context, driver port.DeviceDriver, rosID string) (command.Result, error) {
	cmd := mikrotik.NewResetHotspotUserCountersCommand(rosID)
	return network.ExecuteCommand(ctx, driver, cmd)
}

// GetActiveSessions fetches all currently connected hotspot active sessions.
func (u *UseCase) GetActiveSessions(ctx context.Context, driver port.DeviceDriver) ([]mikrotik.HotspotActiveSession, error) {
	cmd := mikrotik.NewPrintHotspotActiveCommand("")
	res, err := network.ExecuteCommand(ctx, driver, cmd)
	if err != nil {
		return nil, err
	}
	return mikrotik.ParseHotspotActiveSessions(res), nil
}

// RemoveActiveSession kicks an active session by its RouterOS .id.
func (u *UseCase) RemoveActiveSession(ctx context.Context, driver port.DeviceDriver, rosID string) (command.Result, error) {
	cmd := command.Command{
		Raw:  "/ip/hotspot/active/remove",
		Args: map[string]string{".id": rosID},
	}
	return network.ExecuteCommand(ctx, driver, cmd)
}

// GetHosts fetches all /ip/hotspot/host entries.
func (u *UseCase) GetHosts(ctx context.Context, driver port.DeviceDriver) ([]map[string]string, error) {
	cmd := command.Command{Raw: "/ip/hotspot/host/print"}
	res, err := network.ExecuteCommand(ctx, driver, cmd)
	if err != nil {
		return nil, err
	}
	return res.Rows, nil
}

// RemoveHost deletes a hotspot host entry by RouterOS .id.
func (u *UseCase) RemoveHost(ctx context.Context, driver port.DeviceDriver, rosID string) (command.Result, error) {
	cmd := command.Command{
		Raw:  "/ip/hotspot/host/remove",
		Args: map[string]string{".id": rosID},
	}
	return network.ExecuteCommand(ctx, driver, cmd)
}

// GetHotspotServers fetches all /ip/hotspot/print entries.
func (u *UseCase) GetHotspotServers(ctx context.Context, driver port.DeviceDriver) ([]map[string]string, error) {
	cmd := command.Command{Raw: "/ip/hotspot/print"}
	res, err := network.ExecuteCommand(ctx, driver, cmd)
	if err != nil {
		return nil, err
	}
	return res.Rows, nil
}

// GetIPPools fetches all /ip/pool entries.
func (u *UseCase) GetIPPools(ctx context.Context, driver port.DeviceDriver) ([]mikrotik.IPPool, error) {
	cmd := mikrotik.NewPrintIPPoolsCommand("")
	res, err := network.ExecuteCommand(ctx, driver, cmd)
	if err != nil {
		return nil, err
	}
	return mikrotik.ParseIPPools(res), nil
}

// GetParentQueues fetches all parent /queue/simple entries.
func (u *UseCase) GetParentQueues(ctx context.Context, driver port.DeviceDriver) ([]mikrotik.SimpleQueue, error) {
	cmd := mikrotik.NewPrintSimpleQueuesCommand("")
	res, err := network.ExecuteCommand(ctx, driver, cmd)
	if err != nil {
		return nil, err
	}
	queues := mikrotik.ParseSimpleQueues(res)
	parents := make([]mikrotik.SimpleQueue, 0)
	for _, q := range queues {
		if q.Parent == "none" || q.Parent == "" {
			parents = append(parents, q)
		}
	}
	return parents, nil
}

// GetNATRules fetches all /ip/firewall/nat entries.
func (u *UseCase) GetNATRules(ctx context.Context, driver port.DeviceDriver) ([]map[string]string, error) {
	cmd := command.Command{Raw: "/ip/firewall/nat/print"}
	res, err := network.ExecuteCommand(ctx, driver, cmd)
	if err != nil {
		return nil, err
	}
	return res.Rows, nil
}

// SetupExpireMonitor adds or updates the Mikhmon expire-monitor scheduler.
func (u *UseCase) SetupExpireMonitor(ctx context.Context, driver port.DeviceDriver, interval string) (command.Result, error) {
	if interval == "" {
		interval = "00:05:00"
	}
	scriptContent := mikhmon.BuildExpireMonitorScript()

	scriptCmd := command.Command{
		Raw:  "/system/script/add",
		Args: map[string]string{"name": "mikhmon-expire-monitor", "source": scriptContent},
	}
	if _, err := network.ExecuteCommand(ctx, driver, scriptCmd); err != nil {
		return command.Result{}, fmt.Errorf("failed to create expire-monitor script: %w", err)
	}

	schedCmd := command.Command{
		Raw:  "/system/scheduler/add",
		Args: map[string]string{"name": "mikhmon-expire-scheduler", "interval": interval, "on-event": "/system script run mikhmon-expire-monitor"},
	}
	return network.ExecuteCommand(ctx, driver, schedCmd)
}

// GetReports fetches all report records from /system/script/job or /log.
func (u *UseCase) GetReports(ctx context.Context, driver port.DeviceDriver) ([]mikhmon.MikhmonTransaction, error) {
	cmd := command.Command{
		Raw:  "/system/script/print",
		Args: map[string]string{"?owner": "mikhmon-report"},
	}
	res, err := network.ExecuteCommand(ctx, driver, cmd)
	if err != nil {
		return nil, err
	}
	return mikhmon.ParseMikhmonTransactions(res), nil
}

// DeleteReport deletes a report script record by RouterOS .id.
func (u *UseCase) DeleteReport(ctx context.Context, driver port.DeviceDriver, rosID string) (command.Result, error) {
	cmd := command.Command{
		Raw:  "/system/script/remove",
		Args: map[string]string{".id": rosID},
	}
	return network.ExecuteCommand(ctx, driver, cmd)
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
func (u *UseCase) UpdateProfile(ctx context.Context, driver port.DeviceDriver, rosID string, p mikhmon.MikhmonProfileParams) (command.Result, error) {
	cmd := mikhmon.NewSetMikhmonProfileCommand(rosID, p)
	return network.ExecuteCommand(ctx, driver, cmd)
}

// DeleteProfile removes a user profile by RouterOS .id.
func (u *UseCase) DeleteProfile(ctx context.Context, driver port.DeviceDriver, rosID string) (command.Result, error) {
	cmd := command.Command{
		Raw:  "/ip/hotspot/user/profile/remove",
		Args: map[string]string{".id": rosID},
	}
	return network.ExecuteCommand(ctx, driver, cmd)
}

// AddUser creates a new hotspot user directly (non-batch).
func (u *UseCase) AddUser(ctx context.Context, driver port.DeviceDriver, p mikrotik.HotspotUserParams) (command.Result, error) {
	cmd := mikrotik.NewAddHotspotUserCommand(p)
	res, err := network.ExecuteCommand(ctx, driver, cmd)
	if err != nil {
		return res, err
	}
	if p.Comment != "" {
		_, _ = mikhmon.ParseMikhmonComment(p.Comment)
	}
	return res, nil
}

// UpdateUser modifies an existing hotspot user by RouterOS .id.
func (u *UseCase) UpdateUser(ctx context.Context, driver port.DeviceDriver, rosID string, p mikrotik.HotspotUserParams) (command.Result, error) {
	cmd := mikrotik.NewSetHotspotUserCommand(rosID, p)
	return network.ExecuteCommand(ctx, driver, cmd)
}

// GetUsersByTag fetches all hotspot users whose comment contains tag.
func (u *UseCase) GetUsersByTag(ctx context.Context, driver port.DeviceDriver, tag string) ([]mikrotik.HotspotUser, error) {
	users, err := u.GetUsers(ctx, driver, "")
	if err != nil {
		return nil, err
	}

	filtered := make([]mikrotik.HotspotUser, 0)
	for _, usr := range users {
		if usr.Comment == "" {
			continue
		}
		parsed, parseErr := mikhmon.ParseMikhmonComment(usr.Comment)
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
func (u *UseCase) GetReportsByFilter(ctx context.Context, driver port.DeviceDriver, date, month, year string) ([]mikhmon.MikhmonTransaction, error) {
	all, err := u.GetReports(ctx, driver)
	if err != nil {
		return nil, err
	}
	if date == "" && month == "" && year == "" {
		return all, nil
	}
	filtered := make([]mikhmon.MikhmonTransaction, 0)
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
	resCmd := mikrotik.NewPrintSystemResourceCommand()
	if res, err := network.ExecuteCommand(ctx, driver, resCmd); err == nil {
		sysRes := mikrotik.ParseSystemResource(res)
		summary.Uptime = sysRes.Uptime
		summary.Version = sysRes.Version
		summary.CPULoad = sysRes.CPULoad
		summary.BoardName = sysRes.BoardName
	}

	// System Identity
	identCmd := mikrotik.NewPrintSystemIdentityCommand()
	if res, err := network.ExecuteCommand(ctx, driver, identCmd); err == nil && len(res.Rows) > 0 {
		summary.Identity = res.Rows[0]["name"]
	}

	// Active Hotspot Users
	actCmd := mikrotik.NewPrintHotspotActiveCommand("")
	if res, err := network.ExecuteCommand(ctx, driver, actCmd); err == nil {
		activeUsers := mikrotik.ParseHotspotActiveSessions(res)
		summary.ActiveUsers = len(activeUsers)
	}

	// Total Hotspot Users
	usrCmd := mikrotik.NewPrintHotspotUsersCommand("")
	if res, err := network.ExecuteCommand(ctx, driver, usrCmd); err == nil {
		users := mikrotik.ParseHotspotUsers(res)
		summary.TotalUsers = len(users)
	}

	// Today's Income
	if inc, err := u.GetTodayIncome(ctx, driver); err == nil {
		summary.TodayIncome = inc
	}

	return summary, nil
}
