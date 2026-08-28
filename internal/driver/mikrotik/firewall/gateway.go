package firewall

import (
	"context"
	"fmt"
	"strings"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/port"
)

// Gateway implements port.FirewallGateway for MikroTik RouterOS.
type Gateway struct {
	exec port.CommandExecutor
}

// NewGateway creates a Gateway bound to exec.
func NewGateway(exec port.CommandExecutor) *Gateway {
	return &Gateway{exec: exec}
}

// IsolationRedirectComment is the comment tag used to identify isolation redirect NAT rules.
const IsolationRedirectComment = "ISOLATION_REDIRECT"

// FindIsolationRedirectRules filters NAT rules that match the isolation redirect comment.
func FindIsolationRedirectRules(rules []NATRule) []NATRule {
	var matches []NATRule
	for _, r := range rules {
		if strings.Contains(r.Comment, IsolationRedirectComment) {
			matches = append(matches, r)
		}
	}
	return matches
}

// ListFirewallNATRules lists NAT rules filtered by chain, comment, and src-address-list.
func (g *Gateway) ListFirewallNATRules(ctx context.Context, driver port.DeviceDriver, chain, comment, srcAddressList string) ([]NATRule, error) {
	res, err := g.exec(ctx, driver, NewPrintNATCommand(chain, comment, srcAddressList))
	if err != nil {
		return nil, fmt.Errorf("list nat rules: %w", err)
	}
	return ParseNATRules(res), nil
}

// EnsureIsolationRedirect ensures the required redirect NAT rule is configured.
func (g *Gateway) EnsureIsolationRedirect(ctx context.Context, driver port.DeviceDriver, cfg port.IsolationRedirectConfig) error {
	comment := fmt.Sprintf("ISOLATION_REDIRECT_%s", cfg.SrcAddressList)
	res, err := g.exec(ctx, driver, NewPrintNATCommand("dstnat", comment, cfg.SrcAddressList))
	if err != nil {
		return fmt.Errorf("find isolation redirect nat: %w", err)
	}

	rules := ParseNATRules(res)
	portStr := cfg.PaymentPort
	if portStr == "" {
		portStr = "80"
	}

	params := NATRuleParams{
		Chain:          "dstnat",
		Action:         "dst-nat",
		ToAddresses:    cfg.PaymentHost,
		ToPorts:        portStr,
		SrcAddressList: cfg.SrcAddressList,
		Protocol:       "tcp",
		DstPort:        "80,443",
		Comment:        comment,
		Disabled:       cfg.Disabled,
	}

	if len(rules) == 0 {
		_, err := g.exec(ctx, driver, NewAddNATCommand(params))
		if err != nil {
			return fmt.Errorf("add isolation redirect nat: %w", err)
		}
		return nil
	}

	// Update existing rule if parameters differ
	_, err = g.exec(ctx, driver, NewSetNATCommand(rules[0].RosID, params))
	if err != nil {
		return fmt.Errorf("update isolation redirect nat: %w", err)
	}
	return nil
}

// DisableIsolationRedirect disables any isolation redirect rules matching the app convention.
func (g *Gateway) DisableIsolationRedirect(ctx context.Context, driver port.DeviceDriver) error {
	res, err := g.exec(ctx, driver, NewPrintNATCommand("dstnat", "", ""))
	if err != nil {
		return fmt.Errorf("list dstnat rules: %w", err)
	}
	rules := ParseNATRules(res)
	for _, r := range rules {
		if strings.Contains(r.Comment, "ISOLATION_REDIRECT") && !r.Disabled {
			cmd := command.Command{
				Raw:  "/ip/firewall/nat/set",
				Args: map[string]string{".id": r.RosID, "disabled": "yes"},
			}
			if _, err := g.exec(ctx, driver, cmd); err != nil {
				return fmt.Errorf("disable nat rule %s: %w", r.RosID, err)
			}
		}
	}
	return nil
}

// AddToAddressList adds an IP address to a firewall address list.
func (g *Gateway) AddToAddressList(ctx context.Context, driver port.DeviceDriver, listName, address, comment string) error {
	cmd := NewAddAddressListCommand(AddressListParams{
		List:    listName,
		Address: address,
		Comment: comment,
	})
	_, err := g.exec(ctx, driver, cmd)
	if err != nil {
		return fmt.Errorf("add to address-list: %w", err)
	}
	return nil
}

// RemoveFromAddressList removes an IP address from a firewall address list.
func (g *Gateway) RemoveFromAddressList(ctx context.Context, driver port.DeviceDriver, listName, address string) error {
	res, err := g.exec(ctx, driver, NewPrintAddressListCommand(AddressListPrintParams{
		List:    listName,
		Address: address,
	}))
	if err != nil {
		return fmt.Errorf("find address-list entries: %w", err)
	}
	entries := ParseAddressList(res)
	for _, e := range entries {
		if _, err := g.exec(ctx, driver, NewRemoveAddressListCommand(e.RosID)); err != nil {
			return fmt.Errorf("remove address-list entry %s: %w", e.RosID, err)
		}
	}
	return nil
}

// RemoveFromAddressListByComment removes entries matching a comment substring.
func (g *Gateway) RemoveFromAddressListByComment(ctx context.Context, driver port.DeviceDriver, listName, commentContains string) error {
	res, err := g.exec(ctx, driver, NewPrintAddressListCommand(AddressListPrintParams{
		List: listName,
	}))
	if err != nil {
		return fmt.Errorf("find address-list by comment: %w", err)
	}
	entries := ParseAddressList(res)
	for _, e := range entries {
		if strings.Contains(e.Comment, commentContains) {
			if _, err := g.exec(ctx, driver, NewRemoveAddressListCommand(e.RosID)); err != nil {
				return fmt.Errorf("remove address-list entry %s: %w", e.RosID, err)
			}
		}
	}
	return nil
}

// ListNATRules queries NAT rules matching criteria.
func (g *Gateway) ListNATRules(ctx context.Context, driver port.DeviceDriver, chain, comment, srcAddressList string) ([]NATRule, error) {
	res, err := g.exec(ctx, driver, NewPrintNATCommand(chain, comment, srcAddressList))
	if err != nil {
		return nil, err
	}
	return ParseNATRules(res), nil
}
