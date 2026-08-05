package network

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/quixiq/polyglot/internal/domain/command"
	// DEVIASI: Use case jaringan mengimpor driver/mikrotik dan driver/mikrotik/mikhmon
	// khusus untuk mengkonstruksi command native MikroTik/Mikhmon yang kemudian
	// dieksekusi via interface port.DeviceDriver.
	"github.com/quixiq/polyglot/internal/driver/mikrotik"
	"github.com/quixiq/polyglot/internal/driver/mikrotik/mikhmon"
	"github.com/quixiq/polyglot/internal/port"
)

// MikhmonDashboardSummary holds aggregated statistics for the Mikhmon dashboard.
type MikhmonDashboardSummary struct {
	Uptime      string `json:"uptime"`
	Version     string `json:"version"`
	CPULoad     int    `json:"cpu_load"`
	BoardName   string `json:"board_name"`
	Identity    string `json:"identity"`
	TotalUsers  int    `json:"total_users"`
	ActiveUsers int    `json:"active_users"`
	TodayIncome int64  `json:"today_income"`
}

// MikhmonUseCase orchestrates all Mikhmon and Hotspot operations.
type MikhmonUseCase struct{}

// NewMikhmonUseCase creates a new MikhmonUseCase instance.
func NewMikhmonUseCase() *MikhmonUseCase {
	return &MikhmonUseCase{}
}

// CreateProfile builds and executes the /ip/hotspot/user/profile/add command
// with full Mikhmon metadata and on-login RouterOS script.
func (u *MikhmonUseCase) CreateProfile(ctx context.Context, driver port.DeviceDriver, p mikhmon.MikhmonProfileParams) (command.Result, error) {
	cmd := mikhmon.NewAddMikhmonProfileCommand(p)
	return ExecuteCommand(ctx, driver, cmd)
}

// GetProfiles fetches all Hotspot User Profiles.
func (u *MikhmonUseCase) GetProfiles(ctx context.Context, driver port.DeviceDriver) ([]mikrotik.HotspotUserProfile, error) {
	cmd := mikrotik.NewPrintHotspotUserProfilesCommand("")
	res, err := ExecuteCommand(ctx, driver, cmd)
	if err != nil {
		return nil, err
	}
	return mikrotik.ParseHotspotUserProfiles(res), nil
}

// GenerateVouchers generates a batch of count vouchers and executes their creation commands.
func (u *MikhmonUseCase) GenerateVouchers(ctx context.Context, driver port.DeviceDriver, p mikhmon.VoucherGenerateParams, count int) (mikhmon.VoucherBatch, error) {
	batch := mikhmon.NewGenerateVoucherBatchCommands(p, count)
	for _, cmd := range batch.Commands {
		if _, err := ExecuteCommand(ctx, driver, cmd); err != nil {
			return batch, fmt.Errorf("failed to create voucher: %w", err)
		}
	}
	return batch, nil
}

// GetUsers lists hotspot users, optionally filtered by profile.
func (u *MikhmonUseCase) GetUsers(ctx context.Context, driver port.DeviceDriver, profileFilter string) ([]mikrotik.HotspotUser, error) {
	cmd := mikrotik.NewPrintHotspotUsersCommand(profileFilter)
	res, err := ExecuteCommand(ctx, driver, cmd)
	if err != nil {
		return nil, err
	}
	return mikrotik.ParseHotspotUsers(res), nil
}

// GetUser fetches a single hotspot user by RouterOS .id.
func (u *MikhmonUseCase) GetUser(ctx context.Context, driver port.DeviceDriver, rosID string) (mikrotik.HotspotUser, error) {
	cmd := command.Command{
		Raw:  "/ip/hotspot/user/print",
		Args: map[string]string{"?.id": rosID},
	}
	res, err := ExecuteCommand(ctx, driver, cmd)
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
func (u *MikhmonUseCase) RemoveUser(ctx context.Context, driver port.DeviceDriver, rosID string) (command.Result, error) {
	cmd := mikrotik.NewRemoveHotspotUserCommand(rosID)
	return ExecuteCommand(ctx, driver, cmd)
}

// ResetUserCounters resets byte/time counters for a hotspot user.
func (u *MikhmonUseCase) ResetUserCounters(ctx context.Context, driver port.DeviceDriver, rosID string) (command.Result, error) {
	cmd := mikrotik.NewResetHotspotUserCountersCommand(rosID)
	return ExecuteCommand(ctx, driver, cmd)
}

// GetActiveSessions fetches all active hotspot sessions.
func (u *MikhmonUseCase) GetActiveSessions(ctx context.Context, driver port.DeviceDriver) ([]mikrotik.HotspotActiveSession, error) {
	cmd := mikrotik.NewPrintHotspotActiveCommand("")
	res, err := ExecuteCommand(ctx, driver, cmd)
	if err != nil {
		return nil, err
	}
	return mikrotik.ParseHotspotActiveSessions(res), nil
}

// RemoveActiveSession kicks an active hotspot session.
func (u *MikhmonUseCase) RemoveActiveSession(ctx context.Context, driver port.DeviceDriver, rosID string) (command.Result, error) {
	cmd := mikrotik.NewDisconnectHotspotActiveCommand(rosID)
	return ExecuteCommand(ctx, driver, cmd)
}

// GetHosts fetches all hotspot hosts.
func (u *MikhmonUseCase) GetHosts(ctx context.Context, driver port.DeviceDriver) ([]map[string]string, error) {
	cmd := command.Command{Raw: "/ip/hotspot/host/print", Args: map[string]string{}}
	res, err := ExecuteCommand(ctx, driver, cmd)
	if err != nil {
		return nil, err
	}
	return res.Rows, nil
}

// RemoveHost deletes a hotspot host entry.
func (u *MikhmonUseCase) RemoveHost(ctx context.Context, driver port.DeviceDriver, rosID string) (command.Result, error) {
	cmd := command.Command{Raw: "/ip/hotspot/host/remove", Args: map[string]string{".id": rosID}}
	return ExecuteCommand(ctx, driver, cmd)
}

// GetHotspotServers fetches all hotspot server configurations.
func (u *MikhmonUseCase) GetHotspotServers(ctx context.Context, driver port.DeviceDriver) ([]map[string]string, error) {
	cmd := command.Command{Raw: "/ip/hotspot/print", Args: map[string]string{}}
	res, err := ExecuteCommand(ctx, driver, cmd)
	if err != nil {
		return nil, err
	}
	return res.Rows, nil
}

// GetIPPools fetches all IP address pools.
func (u *MikhmonUseCase) GetIPPools(ctx context.Context, driver port.DeviceDriver) ([]mikrotik.IPPool, error) {
	cmd := mikrotik.NewPrintIPPoolsCommand("")
	res, err := ExecuteCommand(ctx, driver, cmd)
	if err != nil {
		return nil, err
	}
	return mikrotik.ParseIPPools(res), nil
}

// GetParentQueues fetches static simple queues available as parent queues.
func (u *MikhmonUseCase) GetParentQueues(ctx context.Context, driver port.DeviceDriver) ([]mikrotik.SimpleQueue, error) {
	cmd := mikrotik.NewPrintParentQueuesCommand()
	res, err := ExecuteCommand(ctx, driver, cmd)
	if err != nil {
		return nil, err
	}
	return mikrotik.ParseSimpleQueues(res), nil
}

// GetNATRules fetches firewall NAT rules.
func (u *MikhmonUseCase) GetNATRules(ctx context.Context, driver port.DeviceDriver) ([]map[string]string, error) {
	cmd := mikrotik.NewPrintFirewallNATRulesCommand()
	res, err := ExecuteCommand(ctx, driver, cmd)
	if err != nil {
		return nil, err
	}
	return res.Rows, nil
}

// SetupExpireMonitor installs or updates the Mikhmon-Expire-Monitor scheduler on the router.
func (u *MikhmonUseCase) SetupExpireMonitor(ctx context.Context, driver port.DeviceDriver, interval string) (command.Result, error) {
	checkCmd := command.Command{
		Raw:  "/system/scheduler/print",
		Args: map[string]string{"?name": mikhmon.MikhmonExpireMonitorName},
	}
	if res, err := ExecuteCommand(ctx, driver, checkCmd); err == nil && len(res.Rows) > 0 {
		rosID := res.Rows[0][".id"]
		updateCmd := mikhmon.NewUpdateMikhmonExpireMonitorCommand(rosID, interval)
		return ExecuteCommand(ctx, driver, updateCmd)
	}

	cmd := mikhmon.NewSetupMikhmonExpireMonitorCommand(interval)
	return ExecuteCommand(ctx, driver, cmd)
}

// GetReports fetches sales transaction logs recorded by Mikhmon scripts.
func (u *MikhmonUseCase) GetReports(ctx context.Context, driver port.DeviceDriver) ([]mikhmon.MikhmonTransaction, error) {
	cmd := mikhmon.NewPrintMikhmonReportsCommand()
	res, err := ExecuteCommand(ctx, driver, cmd)
	if err != nil {
		return nil, err
	}
	return mikhmon.ParseMikhmonTransactions(res), nil
}

// DeleteReport deletes a transaction log script by RouterOS .id.
func (u *MikhmonUseCase) DeleteReport(ctx context.Context, driver port.DeviceDriver, rosID string) (command.Result, error) {
	cmd := command.Command{Raw: "/system/script/remove", Args: map[string]string{".id": rosID}}
	return ExecuteCommand(ctx, driver, cmd)
}

// GetTodayIncome calculates total sales revenue recorded today by Mikhmon scripts.
func (u *MikhmonUseCase) GetTodayIncome(ctx context.Context, driver port.DeviceDriver) (int64, error) {
	todayStr := time.Now().Format("02.01.06")
	reports, err := u.GetReports(ctx, driver)
	if err != nil {
		return 0, err
	}
	var todayIncome int64
	for _, r := range reports {
		if r.Date == todayStr {
			if val, e := strconv.ParseInt(r.Price, 10, 64); e == nil {
				todayIncome += val
			}
		}
	}
	return todayIncome, nil
}

// GetDashboardSummary aggregates system info, active users, total users, and today's income.
func (u *MikhmonUseCase) GetDashboardSummary(ctx context.Context, driver port.DeviceDriver) (MikhmonDashboardSummary, error) {
	summary := MikhmonDashboardSummary{}

	// System Resource
	resCmd := mikrotik.NewPrintSystemResourceCommand()
	if res, err := ExecuteCommand(ctx, driver, resCmd); err == nil {
		sysRes := mikrotik.ParseSystemResource(res)
		summary.Uptime = sysRes.Uptime
		summary.Version = sysRes.Version
		summary.CPULoad = sysRes.CPULoad
		summary.BoardName = sysRes.BoardName
	}

	// System Identity
	identCmd := mikrotik.NewPrintSystemIdentityCommand()
	if res, err := ExecuteCommand(ctx, driver, identCmd); err == nil && len(res.Rows) > 0 {
		summary.Identity = res.Rows[0]["name"]
	}

	// Total Users Count
	usersCmd := mikrotik.NewPrintHotspotUsersCommand("")
	if res, err := ExecuteCommand(ctx, driver, usersCmd); err == nil {
		summary.TotalUsers = len(res.Rows)
	}

	// Active Users Count
	activeCmd := mikrotik.NewPrintHotspotActiveCommand("")
	if res, err := ExecuteCommand(ctx, driver, activeCmd); err == nil {
		summary.ActiveUsers = len(res.Rows)
	}

	// Today Income
	if todayIncome, err := u.GetTodayIncome(ctx, driver); err == nil {
		summary.TodayIncome = todayIncome
	}

	return summary, nil
}
