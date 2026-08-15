package network

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/command"
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
type MikhmonUseCase struct {
	TemplateDir string
}

// NewMikhmonUseCase creates a new MikhmonUseCase instance.
func NewMikhmonUseCase(templateDir string) *MikhmonUseCase {
	if templateDir == "" {
		templateDir = "internal/templates"
	}
	return &MikhmonUseCase{TemplateDir: templateDir}
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
