package port

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/command"
)

// FirewallNATRuleParams describes one /ip/firewall/nat rule. The isolir
// flow uses chain=dstnat rules that redirect web traffic from the isolated
// subscriber pool to the payment portal.
type FirewallNATRuleParams struct {
	Chain       string // 'dstnat' | 'srcnat'
	Action      string // 'redirect' | 'dst-nat' | 'masquerade' | ...
	SrcAddress  string // match source IP/CIDR (e.g. isolir pool range)
	DstPort     string // match destination ports ("80,443")
	Protocol    string // "tcp" typically
	ToAddresses string // dst-nat target (portal IP); ignored by redirect
	ToPorts     string // rewrite destination port
	InInterface string // optional ingress interface matcher
	Comment     string // free-text label; convention "ISOLIR_REDIRECT ..."
	Disabled    bool
}

// FirewallNATRule is one parsed row of /ip/firewall/nat/print.
type FirewallNATRule struct {
	RosID       string
	Chain       string
	Action      string
	SrcAddress  string
	DstPort     string
	Protocol    string
	ToAddresses string
	ToPorts     string
	Comment     string
	Disabled    bool
}

// FirewallGateway exposes the firewall operations needed by provisioning
// and isolation usecases. Vendor command knowledge stays in the driver.
type FirewallGateway interface {
	AddNATRule(ctx context.Context, driver DeviceDriver, p FirewallNATRuleParams) (command.Result, error)
	RemoveNATRule(ctx context.Context, driver DeviceDriver, rosID string) (command.Result, error)
	ListNATRules(ctx context.Context, driver DeviceDriver) ([]FirewallNATRule, error)
}
