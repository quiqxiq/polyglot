package hotspot

import (
	"context"
	"fmt"
	"strings"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/driver/mikrotik/firewall"
	"github.com/quixiq/polyglot/internal/driver/mikrotik/queue"
	"github.com/quixiq/polyglot/internal/driver/mikrotik/system"
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

// NewGateway creates a HotspotGateway bound to exec.
func NewGateway(exec port.CommandExecutor) *Gateway {
	return &Gateway{exec: exec}
}

var _ port.HotspotGateway = (*Gateway)(nil)

// GetSystemResource implements port.HotspotGateway.
func (g *Gateway) GetSystemResource(ctx context.Context, driver port.DeviceDriver) (port.SystemResource, error) {
	res, err := g.exec(ctx, driver, system.NewPrintResourceCommand())
	if err != nil {
		return port.SystemResource{}, err
	}
	return system.ParseResource(res), nil
}

// GetSystemIdentity implements port.HotspotGateway.
func (g *Gateway) GetSystemIdentity(ctx context.Context, driver port.DeviceDriver) (string, error) {
	res, err := g.exec(ctx, driver, system.NewPrintIdentityCommand())
	if err != nil {
		return "", err
	}
	return system.ParseIdentity(res).Name, nil
}

// GetUserProfiles implements port.HotspotGateway.
func (g *Gateway) GetUserProfiles(ctx context.Context, driver port.DeviceDriver) ([]port.HotspotUserProfile, error) {
	res, err := g.exec(ctx, driver, NewPrintUserProfilesCommand(""))
	if err != nil {
		return nil, err
	}
	return ParseUserProfiles(res), nil
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
	return g.exec(ctx, driver, NewRemoveUserProfileCommand(rosID))
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
	return ParseUsers(res), nil
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
	users := ParseUsers(res)
	if len(users) == 0 {
		return port.HotspotUser{}, fmt.Errorf("user %q not found", rosID)
	}
	return users[0], nil
}

// AddUser implements port.HotspotGateway.
func (g *Gateway) AddUser(ctx context.Context, driver port.DeviceDriver, p port.HotspotUserParams) (command.Result, error) {
	return g.exec(ctx, driver, NewAddUserCommand(HotspotUserParams{
		Server:      p.Server,
		Name:        p.Name,
		Password:    p.Password,
		Profile:     p.Profile,
		Comment:     p.Comment,
		LimitUptime: p.LimitUptime,
		LimitBytes:  p.LimitBytesOut,
		Disabled:    p.Disabled,
	}))
}

// UpdateUser implements port.HotspotGateway.
func (g *Gateway) UpdateUser(ctx context.Context, driver port.DeviceDriver, rosID string, p port.HotspotUserParams) (command.Result, error) {
	return g.exec(ctx, driver, NewSetUserCommand(rosID, HotspotUserParams{
		Server:      p.Server,
		Name:        p.Name,
		Password:    p.Password,
		Profile:     p.Profile,
		Comment:     p.Comment,
		LimitUptime: p.LimitUptime,
		LimitBytes:  p.LimitBytesOut,
		Disabled:    p.Disabled,
	}))
}

// RemoveUser implements port.HotspotGateway.
func (g *Gateway) RemoveUser(ctx context.Context, driver port.DeviceDriver, rosID string) (command.Result, error) {
	return g.exec(ctx, driver, NewRemoveUserCommand(rosID))
}

// ResetUserCounters implements port.HotspotGateway.
func (g *Gateway) ResetUserCounters(ctx context.Context, driver port.DeviceDriver, rosID string) (command.Result, error) {
	cmd := command.Command{
		Raw:  "/ip/hotspot/user/reset-counters",
		Args: map[string]string{".id": rosID},
	}
	return g.exec(ctx, driver, cmd)
}

// ListActiveSessions implements port.HotspotGateway.
func (g *Gateway) ListActiveSessions(ctx context.Context, driver port.DeviceDriver) ([]port.HotspotActiveSession, error) {
	res, err := g.exec(ctx, driver, NewPrintActiveCommand(""))
	if err != nil {
		return nil, err
	}
	return ParseActiveSessions(res), nil
}

// RemoveActiveSession implements port.HotspotGateway.
func (g *Gateway) RemoveActiveSession(ctx context.Context, driver port.DeviceDriver, rosID string) (command.Result, error) {
	return g.exec(ctx, driver, NewKickActiveCommand(rosID))
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
	res, err := g.exec(ctx, driver, firewall.NewPrintPoolsCommand(""))
	if err != nil {
		return nil, err
	}
	pools := firewall.ParsePools(res)
	portPools := make([]port.IPPool, len(pools))
	for i, p := range pools {
		portPools[i] = port.IPPool{
			RosID:    p.RosID,
			Name:     p.Name,
			Ranges:   p.Ranges,
			NextPool: p.NextPool,
			Comment:  p.Comment,
		}
	}
	return portPools, nil
}

// ListParentQueues implements port.HotspotGateway.
func (g *Gateway) ListParentQueues(ctx context.Context, driver port.DeviceDriver) ([]port.SimpleQueue, error) {
	res, err := g.exec(ctx, driver, queue.NewPrintSimpleQueuesCommand(""))
	if err != nil {
		return nil, err
	}
	queues := queue.ParseSimpleQueues(res)
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

// SetupExpireMonitor implements port.HotspotGateway.
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
	res, err := g.exec(ctx, driver, system.NewPrintSchedulerCommand(""))
	if err != nil {
		return port.ExpireMonitorStatus{}, err
	}
	return classifyExpireMonitorSchedulers(system.ParseScheduler(res)), nil
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

// ListReports implements port.HotspotGateway.
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
