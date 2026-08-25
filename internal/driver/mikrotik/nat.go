package mikrotik

import (
	"strings"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/port"
)

// FirewallNATRuleParams is the vendor-neutral NAT rule parameter set.
// Canonical definition lives in internal/port.
type FirewallNATRuleParams = port.FirewallNATRuleParams

// FirewallNATRule is the vendor-neutral /ip/firewall/nat row.
// Canonical definition lives in internal/port.
type FirewallNATRule = port.FirewallNATRule

// NewAddFirewallNATCommand builds the command.Command for
// /ip/firewall/nat/add. Chain and action are mandatory; all other fields
// are sent only when non-empty (RouterOS keeps its defaults).
func NewAddFirewallNATCommand(p FirewallNATRuleParams) command.Command {
	args := map[string]string{
		"chain":  p.Chain,
		"action": p.Action,
	}
	setIfNonEmpty(args, "src-address", p.SrcAddress)
	setIfNonEmpty(args, "dst-port", p.DstPort)
	setIfNonEmpty(args, "protocol", p.Protocol)
	setIfNonEmpty(args, "to-addresses", p.ToAddresses)
	setIfNonEmpty(args, "to-ports", p.ToPorts)
	setIfNonEmpty(args, "in-interface", p.InInterface)
	setIfNonEmpty(args, "comment", p.Comment)
	if p.Disabled {
		args["disabled"] = "yes"
	}
	return command.Command{Raw: "/ip/firewall/nat/add", Args: args}
}

// NewRemoveFirewallNATCommand builds the command.Command for
// /ip/firewall/nat/remove. Classified as ClassDestructive.
func NewRemoveFirewallNATCommand(rosID string) command.Command {
	return command.Command{
		Raw:  "/ip/firewall/nat/remove",
		Args: map[string]string{".id": rosID},
	}
}

// ParseFirewallNATRules converts command.Result rows from
// /ip/firewall/nat/print into typed FirewallNATRule values. Rows missing
// ".id", "chain", or "action" are skipped.
func ParseFirewallNATRules(result command.Result) []FirewallNATRule {
	rules := make([]FirewallNATRule, 0, len(result.Rows))
	for _, row := range result.Rows {
		id := row[".id"]
		if id == "" || row["chain"] == "" || row["action"] == "" {
			continue
		}
		rules = append(rules, FirewallNATRule{
			RosID:       id,
			Chain:       row["chain"],
			Action:      row["action"],
			SrcAddress:  row["src-address"],
			DstPort:     row["dst-port"],
			Protocol:    row["protocol"],
			ToAddresses: row["to-addresses"],
			ToPorts:     row["to-ports"],
			Comment:     row["comment"],
			Disabled:    strings.EqualFold(row["disabled"], "true"),
		})
	}
	return rules
}
