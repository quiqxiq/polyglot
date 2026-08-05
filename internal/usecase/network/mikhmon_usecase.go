package network

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/quixiq/polyglot/internal/domain/command"
	// DEVIASI: Use case jaringan mengimpor driver/mikrotik dan driver/mikrotik/mikhmon
	// khusus untuk mengkonstruksi command native MikroTik/Mikhmon yang kemudian
	// dieksekusi via interface port.DeviceDriver.
	"github.com/quixiq/polyglot/internal/driver/mikrotik"
	"github.com/quixiq/polyglot/internal/driver/mikrotik/mikhmon"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/voucher"
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
	// TemplateDir is the path to the directory containing voucher template files
	// (header/row/footer for each layout). Defaults to "internal/templates".
	TemplateDir string
}

// NewMikhmonUseCase creates a new MikhmonUseCase instance.
// templateDir is the path to the internal/templates directory.
// Pass an empty string to use the default relative path "internal/templates".
func NewMikhmonUseCase(templateDir string) *MikhmonUseCase {
	if templateDir == "" {
		templateDir = "internal/templates"
	}
	return &MikhmonUseCase{TemplateDir: templateDir}
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

// UpdateProfile updates an existing Hotspot User Profile by RouterOS .id.
func (u *MikhmonUseCase) UpdateProfile(ctx context.Context, driver port.DeviceDriver, rosID string, p mikhmon.MikhmonProfileParams) (command.Result, error) {
	cmd := mikhmon.NewSetMikhmonProfileCommand(rosID, p)
	return ExecuteCommand(ctx, driver, cmd)
}

// DeleteProfile removes a Hotspot User Profile by RouterOS .id.
func (u *MikhmonUseCase) DeleteProfile(ctx context.Context, driver port.DeviceDriver, rosID string) (command.Result, error) {
	cmd := mikrotik.NewRemoveHotspotUserProfileCommand(rosID)
	return ExecuteCommand(ctx, driver, cmd)
}

// AddUser creates a new hotspot user (non-voucher type "up") with a pre-login Mikhmon comment.
// The comment is formatted as "up-<code>-<date>-<tag>" per Mikhmon convention.
func (u *MikhmonUseCase) AddUser(ctx context.Context, driver port.DeviceDriver, p mikrotik.HotspotUserParams) (command.Result, error) {
	// Wrap comment in Mikhmon "up" format only if comment is a plain tag (no existing Mikhmon format).
	if p.Comment != "" && !mikhmon.IsMikhmonComment(p.Comment) {
		code := mikhmon.GenerateVoucherCode(3, mikhmon.CharSetUpperNum)
		p.Comment = mikhmon.FormatPreLoginComment("up", code, p.Comment, time.Now())
	}
	cmd := mikrotik.NewAddHotspotUserCommand(p)
	return ExecuteCommand(ctx, driver, cmd)
}

// UpdateUser updates an existing hotspot user by RouterOS .id.
func (u *MikhmonUseCase) UpdateUser(ctx context.Context, driver port.DeviceDriver, rosID string, p mikrotik.HotspotUserParams) (command.Result, error) {
	cmd := mikrotik.NewSetHotspotUserCommand(rosID, p)
	return ExecuteCommand(ctx, driver, cmd)
}

// GetUsersByTag retrieves hotspot users whose comment contains the given tag prefix.
// tag is matched against the pre-login comment label field (e.g. "Voucher_1_Hari").
// An empty tag returns all users.
func (u *MikhmonUseCase) GetUsersByTag(ctx context.Context, driver port.DeviceDriver, tag string) ([]mikrotik.HotspotUser, error) {
	all, err := u.GetUsers(ctx, driver, "")
	if err != nil {
		return nil, err
	}
	if tag == "" {
		return all, nil
	}
	filtered := make([]mikrotik.HotspotUser, 0)
	for _, user := range all {
		parsed, parseErr := mikhmon.ParseMikhmonComment(user.Comment)
		if parseErr != nil {
			continue
		}
		if strings.Contains(parsed.Tag, tag) {
			filtered = append(filtered, user)
		}
	}
	return filtered, nil
}

// GetReportsByFilter fetches sales transaction logs filtered by optional date, month, or year.
// All filter parameters are optional — empty string = no filter for that field.
// date format: "DD" (e.g. "05"), month format: "MMM" (e.g. "jan"), year: "YYYY" (e.g. "2026").
func (u *MikhmonUseCase) GetReportsByFilter(ctx context.Context, driver port.DeviceDriver, date, month, year string) ([]mikhmon.MikhmonTransaction, error) {
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

// RenderVoucherHTML converts a generated VoucherBatch into a printable HTML page
// using the template files at u.TemplateDir.
//
// layout: "default" | "small" | "thermal".
// hotspotName, dnsName, logo are injected into every voucher card.
// Returns the complete HTML string ready to be served as text/html.
func (u *MikhmonUseCase) RenderVoucherHTML(batch mikhmon.VoucherBatch, layout, hotspotName, dnsName, logo string) (string, error) {
	vouchers := make([]voucher.VoucherData, 0, len(batch.Vouchers))
	for i, v := range batch.Vouchers {
		// Extract tag from Mikhmon comment (used as validity label in card).
		cardValidity := ""
		if parsed, err := mikhmon.ParseMikhmonComment(v.Comment); err == nil {
			cardValidity = parsed.Tag
		}
		vouchers = append(vouchers, voucher.VoucherData{
			Username:    v.Username,
			Password:    v.Password,
			Comment:     v.Comment,
			Validity:    cardValidity,
			HotspotName: hotspotName,
			DNSName:     dnsName,
			Logo:        logo,
			Number:      i + 1,
		})
	}
	return voucher.Render(vouchers, voucher.Layout(layout), u.TemplateDir)
}
