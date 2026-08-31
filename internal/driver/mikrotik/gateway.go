package mikrotik

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/driver/mikrotik/dhcp"
	"github.com/quixiq/polyglot/internal/driver/mikrotik/firewall"
	"github.com/quixiq/polyglot/internal/driver/mikrotik/hotspot"
	"github.com/quixiq/polyglot/internal/driver/mikrotik/iface"
	"github.com/quixiq/polyglot/internal/driver/mikrotik/ppp"
	"github.com/quixiq/polyglot/internal/driver/mikrotik/queue"
	"github.com/quixiq/polyglot/internal/driver/mikrotik/system"
	"github.com/quixiq/polyglot/internal/port"
)

// Gateway implements port.SessionGateway, port.DeviceDiagnostics, port.PPPGateway,
// and port.FirewallGateway for MikroTik RouterOS by aggregating the dedicated sub-packages.
type Gateway struct {
	exec    port.CommandExecutor
	pppGW   *ppp.Gateway
	fireGW  *firewall.Gateway
	queueGW *queue.Gateway
	hotGW   *hotspot.Gateway
}

// NewGateway creates an aggregated Gateway bound to exec.
func NewGateway(exec port.CommandExecutor) *Gateway {
	return &Gateway{
		exec:    exec,
		pppGW:   ppp.NewGateway(exec),
		fireGW:  firewall.NewGateway(exec),
		queueGW: queue.NewGateway(exec),
		hotGW:   hotspot.NewGateway(exec),
	}
}

// NewQueueGateway creates a standalone QueueGateway bound to exec.
func NewQueueGateway(exec port.CommandExecutor) port.QueueGateway {
	return queue.NewGateway(exec)
}

var _ port.SessionGateway = (*Gateway)(nil)
var _ port.DeviceDiagnostics = (*Gateway)(nil)
var _ port.PPPGateway = (*Gateway)(nil)
var _ port.FirewallGateway = (*Gateway)(nil)

// ─── port.PPPGateway Delegation ───────────────────────────────────────────────

func (g *Gateway) ListSecrets(ctx context.Context, driver port.DeviceDriver, nameFilter string) ([]port.PPPoESecret, error) {
	return g.pppGW.ListSecrets(ctx, driver, nameFilter)
}

func (g *Gateway) GetSecret(ctx context.Context, driver port.DeviceDriver, rosID string) (port.PPPoESecret, error) {
	return g.pppGW.GetSecret(ctx, driver, rosID)
}

func (g *Gateway) AddSecret(ctx context.Context, driver port.DeviceDriver, p port.PPPoESecretParams) (command.Result, error) {
	return g.pppGW.AddSecret(ctx, driver, p)
}

func (g *Gateway) UpdateSecret(ctx context.Context, driver port.DeviceDriver, rosID string, p port.PPPoESecretParams) (command.Result, error) {
	return g.pppGW.UpdateSecret(ctx, driver, rosID, p)
}

func (g *Gateway) RemoveSecret(ctx context.Context, driver port.DeviceDriver, rosID string) (command.Result, error) {
	return g.pppGW.RemoveSecret(ctx, driver, rosID)
}

func (g *Gateway) SetSecretDisabled(ctx context.Context, driver port.DeviceDriver, rosID string, disabled bool) (command.Result, error) {
	return g.pppGW.SetSecretDisabled(ctx, driver, rosID, disabled)
}

func (g *Gateway) ListProfiles(ctx context.Context, driver port.DeviceDriver, nameFilter string) ([]port.PPPProfile, error) {
	return g.pppGW.ListProfiles(ctx, driver, nameFilter)
}

func (g *Gateway) GetProfile(ctx context.Context, driver port.DeviceDriver, rosID string) (port.PPPProfile, error) {
	return g.pppGW.GetProfile(ctx, driver, rosID)
}

func (g *Gateway) AddProfile(ctx context.Context, driver port.DeviceDriver, p port.PPPProfileParams) (command.Result, error) {
	return g.pppGW.AddProfile(ctx, driver, p)
}

func (g *Gateway) UpdateProfile(ctx context.Context, driver port.DeviceDriver, rosID string, p port.PPPProfileParams) (command.Result, error) {
	return g.pppGW.UpdateProfile(ctx, driver, rosID, p)
}

func (g *Gateway) RemoveProfile(ctx context.Context, driver port.DeviceDriver, rosID string) (command.Result, error) {
	return g.pppGW.RemoveProfile(ctx, driver, rosID)
}

func (g *Gateway) ListActive(ctx context.Context, driver port.DeviceDriver, nameFilter string) ([]port.PPPActiveSession, error) {
	return g.pppGW.ListActive(ctx, driver, nameFilter)
}

func (g *Gateway) KickActive(ctx context.Context, driver port.DeviceDriver, rosID string) (command.Result, error) {
	return g.pppGW.KickActive(ctx, driver, rosID)
}

func (g *Gateway) ListInactive(ctx context.Context, driver port.DeviceDriver) ([]port.PPPoESecret, error) {
	return g.pppGW.ListInactive(ctx, driver)
}

// ─── port.SessionGateway Delegation ──────────────────────────────────────────

func (g *Gateway) ListPPPActive(ctx context.Context, driver port.DeviceDriver) ([]port.PPPActiveSession, error) {
	return g.pppGW.ListActive(ctx, driver, "")
}

func (g *Gateway) ListPPPInactive(ctx context.Context, driver port.DeviceDriver) ([]port.PPPoESecret, error) {
	return g.pppGW.ListInactive(ctx, driver)
}

func (g *Gateway) KickPPPSession(ctx context.Context, driver port.DeviceDriver, rosID string) (command.Result, error) {
	return g.pppGW.KickActive(ctx, driver, rosID)
}

func (g *Gateway) ListHotspotActive(ctx context.Context, driver port.DeviceDriver) ([]port.HotspotActiveSession, error) {
	return g.hotGW.ListActiveSessions(ctx, driver)
}

func (g *Gateway) ListHotspotInactiveUsers(ctx context.Context, driver port.DeviceDriver) ([]port.HotspotUser, error) {
	uRes, err := g.exec(ctx, driver, hotspot.NewPrintUsersCommand(""))
	if err != nil {
		return nil, err
	}
	users := hotspot.ParseUsers(uRes)

	actRes, err := g.exec(ctx, driver, hotspot.NewPrintActiveCommand(""))
	if err != nil {
		return nil, err
	}
	active := hotspot.ParseActiveSessions(actRes)

	return hotspot.FilterInactiveUsers(users, active), nil
}

func (g *Gateway) KickHotspotSession(ctx context.Context, driver port.DeviceDriver, rosID string) (command.Result, error) {
	return g.hotGW.RemoveActiveSession(ctx, driver, rosID)
}

func (g *Gateway) ListDHCPLeases(ctx context.Context, driver port.DeviceDriver, macFilter string) ([]port.DHCPLease, error) {
	return dhcp.ListLeases(ctx, driver, g.exec, macFilter)
}

func (g *Gateway) SetDHCPLeaseBlock(ctx context.Context, driver port.DeviceDriver, rosID string, blocked bool, comment string) (command.Result, error) {
	return dhcp.SetLeaseBlock(ctx, driver, g.exec, rosID, dhcp.DHCPLeaseBlockParams{
		Blocked: blocked,
		Comment: comment,
	})
}

// ─── port.DeviceDiagnostics Delegation ────────────────────────────────────────

func (g *Gateway) GetSystemResource(ctx context.Context, driver port.DeviceDriver) (port.SystemResource, error) {
	return system.GetSystemResource(ctx, driver, g.exec)
}

func (g *Gateway) GetSystemIdentity(ctx context.Context, driver port.DeviceDriver) (string, error) {
	ident, err := system.GetSystemIdentity(ctx, driver, g.exec)
	if err != nil {
		return "", err
	}
	return ident.Name, nil
}

func (g *Gateway) ListInterfaces(ctx context.Context, driver port.DeviceDriver, typeFilter, nameFilter string) ([]port.Interface, error) {
	return iface.ListInterfaces(ctx, driver, g.exec, typeFilter, nameFilter)
}

func (g *Gateway) MonitorTrafficOnce(ctx context.Context, driver port.DeviceDriver, ifaceName string) (port.InterfaceTrafficStats, error) {
	return iface.MonitorTrafficOnce(ctx, driver, g.exec, ifaceName)
}

// ─── port.FirewallGateway Delegation ──────────────────────────────────────────

// EnsureIsolationRedirect ensures isolation redirect rule is in place.
func (g *Gateway) EnsureIsolationRedirect(ctx context.Context, driver port.DeviceDriver, cfg port.IsolationRedirectConfig) error {
	return g.fireGW.EnsureIsolationRedirect(ctx, driver, cfg)
}

// EnsureIsolationFilter ensures isolation filter rule is in place.
func (g *Gateway) EnsureIsolationFilter(ctx context.Context, driver port.DeviceDriver, srcAddressList, paymentHost string) error {
	return g.fireGW.EnsureIsolationFilter(ctx, driver, srcAddressList, paymentHost)
}

// DisableIsolationRedirect disables isolation redirect rule.
func (g *Gateway) DisableIsolationRedirect(ctx context.Context, driver port.DeviceDriver) error {
	return g.fireGW.DisableIsolationRedirect(ctx, driver)
}

// HasIsolationRedirect checks if isolation redirect rule exists.
func (g *Gateway) HasIsolationRedirect(ctx context.Context, driver port.DeviceDriver, srcAddressList string) (bool, error) {
	return g.fireGW.HasIsolationRedirect(ctx, driver, srcAddressList)
}

// CountAddressListEntries counts entries in address list.
func (g *Gateway) CountAddressListEntries(ctx context.Context, driver port.DeviceDriver, listName string) (int, error) {
	return g.fireGW.CountAddressListEntries(ctx, driver, listName)
}

func (g *Gateway) AddToAddressList(ctx context.Context, driver port.DeviceDriver, listName, address, comment string) error {
	return g.fireGW.AddToAddressList(ctx, driver, listName, address, comment)
}

func (g *Gateway) RemoveFromAddressList(ctx context.Context, driver port.DeviceDriver, listName, address string) error {
	return g.fireGW.RemoveFromAddressList(ctx, driver, listName, address)
}

func (g *Gateway) RemoveFromAddressListByComment(ctx context.Context, driver port.DeviceDriver, listName, commentContains string) error {
	return g.fireGW.RemoveFromAddressListByComment(ctx, driver, listName, commentContains)
}

type NATRule = firewall.NATRule

func (g *Gateway) ListFirewallNATRules(ctx context.Context, driver port.DeviceDriver, chain, comment, srcAddressList string) ([]firewall.NATRule, error) {
	return g.fireGW.ListFirewallNATRules(ctx, driver, chain, comment, srcAddressList)
}

// IsolationRedirectComment is the comment tag used to identify app-managed redirect NAT rules.
const IsolationRedirectComment = firewall.IsolationRedirectComment

// FindIsolationRedirectRules filters NAT rules that contain the IsolationRedirectComment.
func FindIsolationRedirectRules(rules []firewall.NATRule) []firewall.NATRule {
	return firewall.FindIsolationRedirectRules(rules)
}
