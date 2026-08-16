package hotspot

import (
	"context"
	"fmt"

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
func (g *Gateway) ListUsers(ctx context.Context, driver port.DeviceDriver, profileFilter string) ([]port.HotspotUser, error) {
	res, err := g.exec(ctx, driver, mikrotik.NewPrintHotspotUsersCommand(profileFilter))
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

// SetupExpireMonitor implements port.HotspotGateway.
func (g *Gateway) SetupExpireMonitor(ctx context.Context, driver port.DeviceDriver, interval string) (command.Result, error) {
	if interval == "" {
		interval = "00:05:00"
	}
	scriptContent := BuildExpireMonitorScript()

	scriptCmd := command.Command{
		Raw:  "/system/script/add",
		Args: map[string]string{"name": "mikhmon-expire-monitor", "source": scriptContent},
	}
	if _, err := g.exec(ctx, driver, scriptCmd); err != nil {
		return command.Result{}, fmt.Errorf("failed to create expire-monitor script: %w", err)
	}

	schedCmd := command.Command{
		Raw:  "/system/scheduler/add",
		Args: map[string]string{"name": "mikhmon-expire-scheduler", "interval": interval, "on-event": "/system script run mikhmon-expire-monitor"},
	}
	return g.exec(ctx, driver, schedCmd)
}

// ListReports implements port.HotspotGateway.
func (g *Gateway) ListReports(ctx context.Context, driver port.DeviceDriver) ([]port.MikhmonTransaction, error) {
	cmd := command.Command{
		Raw:  "/system/script/print",
		Args: map[string]string{"?owner": "mikhmon-report"},
	}
	res, err := g.exec(ctx, driver, cmd)
	if err != nil {
		return nil, err
	}
	return ParseMikhmonTransactions(res), nil
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
