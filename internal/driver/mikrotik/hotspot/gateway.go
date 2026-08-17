package hotspot

import (
	"context"
	"fmt"
	"strings"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/driver/mikrotik"
	"github.com/quixiq/polyglot/internal/port"
)

// Gateway implements port.HotspotGateway for MikroTik RouterOS. It holds
// all vendor-native command construction and result parsing knowledge
// (RouterOS + Mikhmon conventions) and executes commands through the
// injected CommandExecutor so the policy gate (classify → decide) is
// applied exactly once per command. Stateless: the per-request device
// connection is passed in as port.DeviceDriver.
type Gateway struct {
	exec port.CommandExecutor
}

// NewGateway creates a HotspotGateway bound to exec. exec must be the
// policy-gated executor (usecase/network.ExecuteCommand) — a bare
// driver.Execute here would silently bypass destructive-command approval.
func NewGateway(exec port.CommandExecutor) *Gateway {
	return &Gateway{exec: exec}
}

var _ port.HotspotGateway = (*Gateway)(nil)

// GetSystemResource implements port.HotspotGateway.
func (g *Gateway) GetSystemResource(ctx context.Context, driver port.DeviceDriver) (port.SystemResource, error) {
	res, err := g.exec(ctx, driver, mikrotik.NewPrintSystemResourceCommand())
	if err != nil {
		return port.SystemResource{}, err
	}
	return mikrotik.ParseSystemResource(res), nil
}

// GetSystemIdentity implements port.HotspotGateway.
func (g *Gateway) GetSystemIdentity(ctx context.Context, driver port.DeviceDriver) (string, error) {
	res, err := g.exec(ctx, driver, mikrotik.NewPrintSystemIdentityCommand())
	if err != nil {
		return "", err
	}
	return mikrotik.ParseSystemIdentity(res).Name, nil
}

// GetUserProfiles implements port.HotspotGateway.
func (g *Gateway) GetUserProfiles(ctx context.Context, driver port.DeviceDriver) ([]port.HotspotUserProfile, error) {
	res, err := g.exec(ctx, driver, mikrotik.NewPrintHotspotUserProfilesCommand(""))
	if err != nil {
		return nil, err
	}
	return mikrotik.ParseHotspotUserProfiles(res), nil
}

// CreateUserProfile implements port.HotspotGateway.
func (g *Gateway) CreateUserProfile(ctx context.Context, driver port.DeviceDriver, p port.MikhmonProfileParams) (command.Result, error) {
	return g.exec(ctx, driver, NewAddMikhmonProfileCommand(p))
}

// UpdateUserProfile implements port.HotspotGateway.
func (g *Gateway) UpdateUserProfile(ctx context.Context, driver port.DeviceDriver, rosID string, p port.MikhmonProfileParams) (command.Result, error) {
	return g.exec(ctx, driver, NewSetMikhmonProfileCommand(rosID, p))
}

// DeleteUserProfile implements port.HotspotGateway.
func (g *Gateway) DeleteUserProfile(ctx context.Context, driver port.DeviceDriver, rosID string) (command.Result, error) {
	return g.exec(ctx, driver, mikrotik.NewRemoveHotspotUserProfileCommand(rosID))
}

// GenerateVouchers implements port.HotspotGateway.
func (g *Gateway) GenerateVouchers(ctx context.Context, driver port.DeviceDriver, p port.VoucherGenerateParams, count int) (port.VoucherBatch, error) {
	batch := NewGenerateVoucherBatchCommands(p, count)
	for _, cmd := range batch.Commands {
		if _, err := g.exec(ctx, driver, cmd); err != nil {
			return port.VoucherBatch{}, fmt.Errorf("failed to create voucher: %w", err)
		}
	}
	out := port.VoucherBatch{Vouchers: make([]port.GeneratedVoucher, 0, len(batch.Vouchers))}
	for _, v := range batch.Vouchers {
		out.Vouchers = append(out.Vouchers, port.GeneratedVoucher{
			Username: v.Username,
			Password: v.Password,
			Comment:  v.Comment,
		})
	}
	return out, nil
}

// ListUsers implements port.HotspotGateway.
func (g *Gateway) ListUsers(ctx context.Context, driver port.DeviceDriver, f port.ListUsersFilter) ([]port.HotspotUser, error) {
	cmd := command.Command{Raw: "/ip/hotspot/user/print", Args: map[string]string{}}
	if f.Name != "" {
		cmd.Args["?name"] = f.Name
	}
	if f.Profile != "" {
		cmd.Args["?profile"] = f.Profile
	}
	if f.Comment != "" {
		cmd.Args["?comment"] = f.Comment
	}
	if f.OnlyUnused {
		cmd.Args["?uptime"] = "0s"
	}
	res, err := g.exec(ctx, driver, cmd)
	if err != nil {
		return nil, err
	}
	return mikrotik.ParseHotspotUsers(res), nil
}

// GetUser implements port.HotspotGateway.
func (g *Gateway) GetUser(ctx context.Context, driver port.DeviceDriver, rosID string) (port.HotspotUser, error) {
	cmd := command.Command{
		Raw:  "/ip/hotspot/user/print",
		Args: map[string]string{"?.id": rosID},
	}
	res, err := g.exec(ctx, driver, cmd)
	if err != nil {
		return port.HotspotUser{}, err
	}
	users := mikrotik.ParseHotspotUsers(res)
	if len(users) == 0 {
		return port.HotspotUser{}, fmt.Errorf("user %q not found", rosID)
	}
	return users[0], nil
}

// AddUser implements port.HotspotGateway.
func (g *Gateway) AddUser(ctx context.Context, driver port.DeviceDriver, p port.HotspotUserParams) (command.Result, error) {
	return g.exec(ctx, driver, mikrotik.NewAddHotspotUserCommand(p))
}

// UpdateUser implements port.HotspotGateway.
func (g *Gateway) UpdateUser(ctx context.Context, driver port.DeviceDriver, rosID string, p port.HotspotUserParams) (command.Result, error) {
	return g.exec(ctx, driver, mikrotik.NewSetHotspotUserCommand(rosID, p))
}

// RemoveUser implements port.HotspotGateway.
func (g *Gateway) RemoveUser(ctx context.Context, driver port.DeviceDriver, rosID string) (command.Result, error) {
	return g.exec(ctx, driver, mikrotik.NewRemoveHotspotUserCommand(rosID))
}

// ResetUserCounters implements port.HotspotGateway.
func (g *Gateway) ResetUserCounters(ctx context.Context, driver port.DeviceDriver, rosID string) (command.Result, error) {
	return g.exec(ctx, driver, mikrotik.NewResetHotspotUserCountersCommand(rosID))
}

// ListActiveSessions implements port.HotspotGateway.
func (g *Gateway) ListActiveSessions(ctx context.Context, driver port.DeviceDriver) ([]port.HotspotActiveSession, error) {
	res, err := g.exec(ctx, driver, mikrotik.NewPrintHotspotActiveCommand(""))
	if err != nil {
		return nil, err
	}
	return mikrotik.ParseHotspotActiveSessions(res), nil
}

// RemoveActiveSession implements port.HotspotGateway.
func (g *Gateway) RemoveActiveSession(ctx context.Context, driver port.DeviceDriver, rosID string) (command.Result, error) {
	return g.exec(ctx, driver, mikrotik.NewDisconnectHotspotActiveCommand(rosID))
}

// ListHosts implements port.HotspotGateway.
func (g *Gateway) ListHosts(ctx context.Context, driver port.DeviceDriver) ([]map[string]string, error) {
	cmd := command.Command{Raw: "/ip/hotspot/host/print"}
	res, err := g.exec(ctx, driver, cmd)
	if err != nil {
		return nil, err
	}
	return res.Rows, nil
}

// RemoveHost implements port.HotspotGateway.
func (g *Gateway) RemoveHost(ctx context.Context, driver port.DeviceDriver, rosID string) (command.Result, error) {
	cmd := command.Command{
		Raw:  "/ip/hotspot/host/remove",
		Args: map[string]string{".id": rosID},
	}
	return g.exec(ctx, driver, cmd)
}

// ListHotspotServers implements port.HotspotGateway.
func (g *Gateway) ListHotspotServers(ctx context.Context, driver port.DeviceDriver) ([]map[string]string, error) {
	cmd := command.Command{Raw: "/ip/hotspot/print"}
	res, err := g.exec(ctx, driver, cmd)
	if err != nil {
		return nil, err
	}
	return res.Rows, nil
}

// ListIPPools implements port.HotspotGateway.
func (g *Gateway) ListIPPools(ctx context.Context, driver port.DeviceDriver) ([]port.IPPool, error) {
	res, err := g.exec(ctx, driver, mikrotik.NewPrintIPPoolsCommand(""))
	if err != nil {
		return nil, err
	}
	return mikrotik.ParseIPPools(res), nil
}

// ListParentQueues implements port.HotspotGateway.
func (g *Gateway) ListParentQueues(ctx context.Context, driver port.DeviceDriver) ([]port.SimpleQueue, error) {
	res, err := g.exec(ctx, driver, mikrotik.NewPrintSimpleQueuesCommand(""))
	if err != nil {
		return nil, err
	}
	queues := mikrotik.ParseSimpleQueues(res)
	parents := make([]port.SimpleQueue, 0)
	for _, q := range queues {
		if q.Parent == "none" || q.Parent == "" {
			parents = append(parents, q)
		}
	}
	return parents, nil
}

// ListNATRules implements port.HotspotGateway.
func (g *Gateway) ListNATRules(ctx context.Context, driver port.DeviceDriver) ([]map[string]string, error) {
	cmd := command.Command{Raw: "/ip/firewall/nat/print"}
	res, err := g.exec(ctx, driver, cmd)
	if err != nil {
		return nil, err
	}
	return res.Rows, nil
}

// SetupExpireMonitor implements port.HotspotGateway. It is idempotent and
// recognises both scheduler forms (Fase 6 decision A):
//   - legacy single-step scheduler "Mikhmon-Expire-Monitor" exists → update it
//     in place (interval + on-event script source);
//   - gateway two-step form (script "mikhmon-expire-monitor" + scheduler
//     "mikhmon-expire-scheduler") exists → re-arm it;
//   - neither exists → create the gateway two-step form.
// Default interval is "00:01:00" (legacy Mikhmon).
func (g *Gateway) SetupExpireMonitor(ctx context.Context, driver port.DeviceDriver, interval string) (command.Result, error) {
	if interval == "" {
		interval = "00:01:00"
	}

	status, err := g.GetExpireMonitorStatus(ctx, driver)
	if err != nil {
		return command.Result{}, fmt.Errorf("check expire monitor: %w", err)
	}

	scriptContent := BuildExpireMonitorScript()

	// Legacy single-step scheduler: update in place.
	if status.IsInstalled && status.SchedulerName == MikhmonExpireMonitorName {
		return g.exec(ctx, driver, NewUpdateMikhmonExpireMonitorCommand(status.SchedulerID, interval))
	}

	// Gateway two-step scheduler already exists: re-arm (enable + interval).
	if status.IsInstalled && status.SchedulerName == mikhmonExpireSchedulerName {
		schedCmd := command.Command{
			Raw: "/system/scheduler/set",
			Args: map[string]string{
				".id":      status.SchedulerID,
				"interval": interval,
				"on-event": "/system script run " + mikhmonExpireScriptName,
				"disabled": "no",
			},
		}
		return g.exec(ctx, driver, schedCmd)
	}

	// Not installed: create the gateway two-step form (script + scheduler).
	scriptCmd := command.Command{
		Raw:  "/system/script/add",
		Args: map[string]string{"name": mikhmonExpireScriptName, "source": scriptContent},
	}
	if _, err := g.exec(ctx, driver, scriptCmd); err != nil {
		return command.Result{}, fmt.Errorf("failed to create expire-monitor script: %w", err)
	}

	schedCmd := command.Command{
		Raw:  "/system/scheduler/add",
		Args: map[string]string{"name": mikhmonExpireSchedulerName, "interval": interval, "on-event": "/system script run " + mikhmonExpireScriptName},
	}
	return g.exec(ctx, driver, schedCmd)
}

// GetExpireMonitorStatus implements port.HotspotGateway.
func (g *Gateway) GetExpireMonitorStatus(ctx context.Context, driver port.DeviceDriver) (port.ExpireMonitorStatus, error) {
	res, err := g.exec(ctx, driver, mikrotik.NewPrintSystemSchedulersCommand(""))
	if err != nil {
		return port.ExpireMonitorStatus{}, err
	}
	return classifyExpireMonitorSchedulers(mikrotik.ParseSystemSchedulers(res)), nil
}

// SetExpireMonitorDisabled implements port.HotspotGateway.
func (g *Gateway) SetExpireMonitorDisabled(ctx context.Context, driver port.DeviceDriver, rosID string, disabled bool) (command.Result, error) {
	flag := "no"
	if disabled {
		flag = "yes"
	}
	cmd := command.Command{
		Raw:  "/system/scheduler/set",
		Args: map[string]string{".id": rosID, "disabled": flag},
	}
	return g.exec(ctx, driver, cmd)
}

// RemoveExpireMonitor implements port.HotspotGateway.
func (g *Gateway) RemoveExpireMonitor(ctx context.Context, driver port.DeviceDriver, rosID string) (command.Result, error) {
	cmd := command.Command{
		Raw:  "/system/scheduler/remove",
		Args: map[string]string{".id": rosID},
	}
	return g.exec(ctx, driver, cmd)
}

// ListReports implements port.HotspotGateway. Filters follow legacy Mikhmon:
// day → ?source= (get_report), month → ?owner= (get_livereport); when neither
// is given the record set is scoped with ?comment=mikhmon. The year filter is
// applied as a date-suffix match over the fetched records.
func (g *Gateway) ListReports(ctx context.Context, driver port.DeviceDriver, f port.ReportFilter) ([]port.MikhmonTransaction, error) {
	args := map[string]string{}
	if f.Day != "" {
		args["?source"] = f.Day
	}
	if f.Month != "" {
		args["?owner"] = f.Month
	}
	if len(args) == 0 {
		args["?comment"] = "mikhmon"
	}

	cmd := command.Command{Raw: "/system/script/print", Args: args}
	res, err := g.exec(ctx, driver, cmd)
	if err != nil {
		return nil, err
	}

	records := ParseMikhmonTransactions(res)
	if f.Year == "" {
		return records, nil
	}
	filtered := make([]port.MikhmonTransaction, 0, len(records))
	for _, r := range records {
		if strings.HasSuffix(r.Date, f.Year) {
			filtered = append(filtered, r)
		}
	}
	return filtered, nil
}

// DeleteReport implements port.HotspotGateway.
func (g *Gateway) DeleteReport(ctx context.Context, driver port.DeviceDriver, rosID string) (command.Result, error) {
	cmd := command.Command{
		Raw:  "/system/script/remove",
		Args: map[string]string{".id": rosID},
	}
	return g.exec(ctx, driver, cmd)
}

// ParseUserComment implements port.HotspotGateway.
func (g *Gateway) ParseUserComment(comment string) (port.MikhmonComment, error) {
	return ParseMikhmonComment(comment)
}
