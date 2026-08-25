package mikrotik

import (
	"context"
	"fmt"
	"strings"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/port"
)

// IsolirConfig is the vendor-neutral isolation configuration.
// Canonical definition lives in internal/port.
type IsolirConfig = port.IsolirConfig

// IsolirSetupResult reports which RouterOS objects exist or were created.
// Canonical definition lives in internal/port.
type IsolirSetupResult = port.IsolirSetupResult

// natCommentPrefix marks rules owned by the isolation feature so cleanup
// and dedup stay precise even on shared routers.
const natCommentPrefix = "ISOLIR_REDIRECT"

// EnsureIsolirInfrastructure is IDEMPOTENT: safe to run before every
// suspension. It guarantees, in order:
//  1. /ip pool <PoolName> exists (created from PoolRange when missing).
//  2. /ppp profile <ProfileName> exists (rate-limit kecil + remote-address
//     dari pool isolir sehingga pelanggan suspensi mendarat di pool itu).
//  3. One dst-nat rule per redirect port exists, matching source addresses
//     from the isolation pool range and redirecting web traffic to the
//     payment portal (dst-nat to portal IP when set, plain redirect
//     otherwise).
func (g *Gateway) EnsureIsolirInfrastructure(ctx context.Context, driver port.DeviceDriver, cfg IsolirConfig) (IsolirSetupResult, error) {
	if err := validateIsolirConfig(cfg); err != nil {
		return IsolirSetupResult{}, err
	}
	res := IsolirSetupResult{}

	// 1. Pool
	poolRes, err := g.exec(ctx, driver, NewPrintIPPoolsCommand(cfg.PoolName))
	if err != nil {
		return res, fmt.Errorf("isolir: list pools: %w", err)
	}
	if len(ParseIPPools(poolRes)) > 0 {
		res.PoolExisted = true
	} else if _, err := g.exec(ctx, driver, NewAddIPPoolCommand(cfg.PoolName, cfg.PoolRange, "Isolated subscribers (auto)")); err != nil {
		return res, fmt.Errorf("isolir: create pool %q: %w", cfg.PoolName, err)
	}

	// 2. Profile
	profileExisted, err := g.EnsureIsolirProfile(ctx, driver, cfg)
	if err != nil {
		return res, err
	}
	res.ProfileExisted = profileExisted

	// 3. NAT rules per redirect port
	natRes, err := g.exec(ctx, driver, NewPrintFirewallNATRulesCommand())
	if err != nil {
		return res, fmt.Errorf("isolir: list nat rules: %w", err)
	}
	existing := indexIsolirNATByPort(ParseFirewallNATRules(natRes))
	for _, portNum := range splitPorts(cfg.RedirectPorts) {
		portNum = strings.TrimSpace(portNum)
		if portNum == "" {
			continue
		}
		if id, ok := existing[portNum]; ok {
			res.NATRuleIDs = append(res.NATRuleIDs, id)
			continue
		}
		addRes, err := g.AddNATRule(ctx, driver, isolirRedirectRule(cfg, portNum))
		if err != nil {
			return res, fmt.Errorf("isolir: add nat rule for port %s: %w", portNum, err)
		}
		id := firstNewID(addRes)
		res.CreatedNATIDs = append(res.CreatedNATIDs, id)
		res.NATRuleIDs = append(res.NATRuleIDs, id)
	}
	return res, nil
}

// EnsureIsolirProfile guarantees the suspended-subscriber PPP profile
// exists with the low rate-limit needed for the portal redirect.
func (g *Gateway) EnsureIsolirProfile(ctx context.Context, driver port.DeviceDriver, cfg IsolirConfig) (existed bool, err error) {
	profiles, err := g.ListProfiles(ctx, driver, cfg.ProfileName)
	if err != nil {
		return false, fmt.Errorf("isolir: list profiles: %w", err)
	}
	if len(profiles) > 0 {
		return true, nil
	}
	params := IsolirProfileParams(cfg.PoolName)
	params.Name = cfg.ProfileName // config wins over the default constant
	if _, err := g.AddProfile(ctx, driver, params); err != nil {
		return false, fmt.Errorf("isolir: create profile %q: %w", cfg.ProfileName, err)
	}
	return false, nil
}

// RemoveIsolirInfrastructure deletes only artifacts this package created:
// NAT rules tagged ISOLIR_REDIRECT plus — optionally — the pool and
// profile when no other subscriber still references them.
func (g *Gateway) RemoveIsolirInfrastructure(ctx context.Context, driver port.DeviceDriver, cfg IsolirConfig) error {
	natRes, err := g.exec(ctx, driver, NewPrintFirewallNATRulesCommand())
	if err != nil {
		return fmt.Errorf("isolir teardown: list nat rules: %w", err)
	}
	for _, rule := range ParseFirewallNATRules(natRes) {
		if !strings.HasPrefix(rule.Comment, natCommentPrefix) {
			continue
		}
		if _, err := g.RemoveNATRule(ctx, driver, rule.RosID); err != nil {
			return fmt.Errorf("isolir teardown: remove nat rule %s: %w", rule.RosID, err)
		}
	}

	if profiles, err := g.ListProfiles(ctx, driver, cfg.ProfileName); err == nil {
		for _, p := range profiles {
			_, _ = g.RemoveProfile(ctx, driver, p.RosID)
		}
	}
	if pools, err := func() ([]IPPool, error) {
		res, err := g.exec(ctx, driver, NewPrintIPPoolsCommand(cfg.PoolName))
		if err != nil {
			return nil, err
		}
		return ParseIPPools(res), nil
	}(); err == nil {
		for _, p := range pools {
			_, _ = g.exec(ctx, driver, NewRemoveIPPoolCommand(p.RosID))
		}
	}
	return nil
}

// isolirRedirectRule builds one dst-nat/redirect rule for a single port.
func isolirRedirectRule(cfg IsolirConfig, portNum string) FirewallNATRuleParams {
	action := "redirect"
	toAddresses := ""
	if cfg.PortalIP != "" {
		action = "dst-nat"
		toAddresses = cfg.PortalIP
	}
	return FirewallNATRuleParams{
		Chain:       "dstnat",
		Action:      action,
		SrcAddress:  cfg.PoolRange,
		Protocol:    "tcp",
		DstPort:     portNum,
		ToAddresses: toAddresses,
		ToPorts:     cfg.PortalHTTPPort,
		Comment:     fmt.Sprintf("%s %s", natCommentPrefix, portNum),
	}
}

// indexIsolirNATByPort maps redirect-port → rosID for existing ISOLIR rules.
func indexIsolirNATByPort(rules []FirewallNATRule) map[string]string {
	out := make(map[string]string)
	for _, r := range rules {
		if r.Chain != "dstnat" || !strings.HasPrefix(r.Comment, natCommentPrefix) {
			continue
		}
		fields := strings.Fields(r.Comment)
		if len(fields) >= 2 {
			out[fields[1]] = r.RosID
		}
	}
	return out
}

func splitPorts(s string) []string {
	return strings.Split(s, ",")
}

func firstNewID(res command.Result) string {
	for _, row := range res.Rows {
		if id := row[".id"]; id != "" {
			return id
		}
	}
	return ""
}

// validateIsolirConfig enforces the required fields before touching the router.
func validateIsolirConfig(c IsolirConfig) error {
	switch {
	case c.ProfileName == "":
		return fmt.Errorf("isolir config: profile name is required")
	case c.PoolName == "":
		return fmt.Errorf("isolir config: pool name is required")
	case c.PoolRange == "":
		return fmt.Errorf("isolir config: pool range is required")
	case c.PortalHTTPPort == "":
		return fmt.Errorf("isolir config: portal http port is required")
	case c.RedirectPorts == "":
		return fmt.Errorf("isolir config: redirect ports are required")
	default:
		return nil
	}
}
