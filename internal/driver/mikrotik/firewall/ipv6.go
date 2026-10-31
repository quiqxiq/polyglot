package firewall

import (
	"strings"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/driver/mikrotik/internal/rosutil"
)

// IPv6FirewallRule represents a rule in /ipv6/firewall (filter, nat, mangle, or raw).
type IPv6FirewallRule struct {
	ID               string `json:"id"`
	Chain            string `json:"chain"`
	Action           string `json:"action"`
	SrcAddress       string `json:"src_address"`
	DstAddress       string `json:"dst_address"`
	SrcAddressList   string `json:"src_address_list"`
	DstAddressList   string `json:"dst_address_list"`
	Protocol         string `json:"protocol"`
	SrcPort          string `json:"src_port"`
	DstPort          string `json:"dst_port"`
	InInterface      string `json:"in_interface"`
	OutInterface     string `json:"out_interface"`
	InInterfaceList  string `json:"in_interface_list"`
	OutInterfaceList string `json:"out_interface_list"`
	ConnectionState  string `json:"connection_state"`
	PlaceBefore      string `json:"place_before,omitempty"`
	Disabled         bool   `json:"disabled"`
	Comment          string `json:"comment"`
	Bytes            string `json:"bytes,omitempty"`
	Packets          string `json:"packets,omitempty"`
}

func buildIPv6RuleArgs(r IPv6FirewallRule) map[string]string {
	args := map[string]string{
		"chain":  r.Chain,
		"action": r.Action,
	}
	rosutil.SetIfNonEmpty(args, "src-address", r.SrcAddress)
	rosutil.SetIfNonEmpty(args, "dst-address", r.DstAddress)
	rosutil.SetIfNonEmpty(args, "src-address-list", r.SrcAddressList)
	rosutil.SetIfNonEmpty(args, "dst-address-list", r.DstAddressList)
	rosutil.SetIfNonEmpty(args, "protocol", r.Protocol)
	rosutil.SetIfNonEmpty(args, "src-port", r.SrcPort)
	rosutil.SetIfNonEmpty(args, "dst-port", r.DstPort)
	rosutil.SetIfNonEmpty(args, "in-interface", r.InInterface)
	rosutil.SetIfNonEmpty(args, "out-interface", r.OutInterface)
	rosutil.SetIfNonEmpty(args, "in-interface-list", r.InInterfaceList)
	rosutil.SetIfNonEmpty(args, "out-interface-list", r.OutInterfaceList)
	rosutil.SetIfNonEmpty(args, "connection-state", r.ConnectionState)
	rosutil.SetIfNonEmpty(args, "place-before", r.PlaceBefore)
	rosutil.SetIfNonEmpty(args, "comment", r.Comment)
	if r.Disabled {
		args["disabled"] = "yes"
	}
	return args
}

// NewPrintIPv6FilterRulesCommand queries /ipv6/firewall/filter.
func NewPrintIPv6FilterRulesCommand(chain string) command.Command {
	args := make(map[string]string)
	rosutil.SetProplist(args, ".id", "chain", "action", "src-address", "dst-address", "protocol", "dst-port", "in-interface-list", "out-interface-list", "connection-state", "disabled", "comment", "bytes", "packets")
	if chain != "" {
		args["?chain"] = chain
	}
	return command.Command{
		Raw:  "/ipv6/firewall/filter/print",
		Args: args,
	}
}

// NewAddIPv6FilterRuleCommand adds a rule to /ipv6/firewall/filter.
func NewAddIPv6FilterRuleCommand(r IPv6FirewallRule) command.Command {
	return command.Command{
		Raw:  "/ipv6/firewall/filter/add",
		Args: buildIPv6RuleArgs(r),
	}
}

// NewSetIPv6FilterRuleCommand updates a rule in /ipv6/firewall/filter.
func NewSetIPv6FilterRuleCommand(rosID string, r IPv6FirewallRule) command.Command {
	args := buildIPv6RuleArgs(r)
	args[".id"] = rosID
	delete(args, "place-before")
	if r.Disabled {
		args["disabled"] = "yes"
	} else {
		args["disabled"] = "no"
	}
	return command.Command{
		Raw:  "/ipv6/firewall/filter/set",
		Args: args,
	}
}

// NewRemoveIPv6FilterRuleCommand removes a rule from /ipv6/firewall/filter.
func NewRemoveIPv6FilterRuleCommand(rosID string) command.Command {
	return command.Command{
		Raw: "/ipv6/firewall/filter/remove",
		Args: map[string]string{
			".id": rosID,
		},
	}
}

// NewMoveIPv6FilterRuleCommand moves a rule in /ipv6/firewall/filter before destination.
func NewMoveIPv6FilterRuleCommand(id, destination string) command.Command {
	args := map[string]string{
		".id": id,
	}
	if destination != "" {
		args["destination"] = destination
	}
	return command.Command{
		Raw:  "/ipv6/firewall/filter/move",
		Args: args,
	}
}

// NewStreamIPv6FilterRulesCommand streams /ipv6/firewall/filter events.
func NewStreamIPv6FilterRulesCommand(chain string) command.Command {
	args := make(map[string]string)
	rosutil.SetProplist(args, ".id", "chain", "action", "src-address", "dst-address", "protocol", "dst-port", "disabled", "comment", "bytes", "packets")
	args["follow"] = ""
	if chain != "" {
		args["?chain"] = chain
	}
	return command.Command{
		Raw:  "/ipv6/firewall/filter/print",
		Args: args,
	}
}

// NewPrintIPv6NATRulesCommand queries /ipv6/firewall/nat.
func NewPrintIPv6NATRulesCommand(chain string) command.Command {
	args := make(map[string]string)
	rosutil.SetProplist(args, ".id", "chain", "action", "src-address", "dst-address", "protocol", "disabled", "comment")
	if chain != "" {
		args["?chain"] = chain
	}
	return command.Command{
		Raw:  "/ipv6/firewall/nat/print",
		Args: args,
	}
}

// NewAddIPv6NATRuleCommand adds a rule to /ipv6/firewall/nat.
func NewAddIPv6NATRuleCommand(r IPv6FirewallRule) command.Command {
	return command.Command{
		Raw:  "/ipv6/firewall/nat/add",
		Args: buildIPv6RuleArgs(r),
	}
}

// NewRemoveIPv6NATRuleCommand removes a rule from /ipv6/firewall/nat.
func NewRemoveIPv6NATRuleCommand(rosID string) command.Command {
	return command.Command{
		Raw: "/ipv6/firewall/nat/remove",
		Args: map[string]string{
			".id": rosID,
		},
	}
}

// NewPrintIPv6MangleRulesCommand queries /ipv6/firewall/mangle.
func NewPrintIPv6MangleRulesCommand(chain string) command.Command {
	args := make(map[string]string)
	rosutil.SetProplist(args, ".id", "chain", "action", "src-address", "dst-address", "protocol", "disabled", "comment")
	if chain != "" {
		args["?chain"] = chain
	}
	return command.Command{
		Raw:  "/ipv6/firewall/mangle/print",
		Args: args,
	}
}

// NewAddIPv6MangleRuleCommand adds a rule to /ipv6/firewall/mangle.
func NewAddIPv6MangleRuleCommand(r IPv6FirewallRule) command.Command {
	return command.Command{
		Raw:  "/ipv6/firewall/mangle/add",
		Args: buildIPv6RuleArgs(r),
	}
}

// NewRemoveIPv6MangleRuleCommand removes a rule from /ipv6/firewall/mangle.
func NewRemoveIPv6MangleRuleCommand(rosID string) command.Command {
	return command.Command{
		Raw: "/ipv6/firewall/mangle/remove",
		Args: map[string]string{
			".id": rosID,
		},
	}
}

// NewPrintIPv6RawRulesCommand queries /ipv6/firewall/raw.
func NewPrintIPv6RawRulesCommand(chain string) command.Command {
	args := make(map[string]string)
	rosutil.SetProplist(args, ".id", "chain", "action", "src-address", "dst-address", "protocol", "disabled", "comment")
	if chain != "" {
		args["?chain"] = chain
	}
	return command.Command{
		Raw:  "/ipv6/firewall/raw/print",
		Args: args,
	}
}

// NewAddIPv6RawRuleCommand adds a rule to /ipv6/firewall/raw.
func NewAddIPv6RawRuleCommand(r IPv6FirewallRule) command.Command {
	return command.Command{
		Raw:  "/ipv6/firewall/raw/add",
		Args: buildIPv6RuleArgs(r),
	}
}

// NewRemoveIPv6RawRuleCommand removes a rule from /ipv6/firewall/raw.
func NewRemoveIPv6RawRuleCommand(rosID string) command.Command {
	return command.Command{
		Raw: "/ipv6/firewall/raw/remove",
		Args: map[string]string{
			".id": rosID,
		},
	}
}

// ParseIPv6FirewallRules converts reply rows into IPv6FirewallRule slice.
func ParseIPv6FirewallRules(rows []map[string]string) []IPv6FirewallRule {
	out := make([]IPv6FirewallRule, 0, len(rows))
	for _, r := range rows {
		out = append(out, IPv6FirewallRule{
			ID:               r[".id"],
			Chain:            r["chain"],
			Action:           r["action"],
			SrcAddress:       r["src-address"],
			DstAddress:       r["dst-address"],
			SrcAddressList:   r["src-address-list"],
			DstAddressList:   r["dst-address-list"],
			Protocol:         r["protocol"],
			SrcPort:          r["src-port"],
			DstPort:          r["dst-port"],
			InInterface:      r["in-interface"],
			OutInterface:     r["out-interface"],
			InInterfaceList:  r["in-interface-list"],
			OutInterfaceList: r["out-interface-list"],
			ConnectionState:  r["connection-state"],
			Disabled:         strings.EqualFold(r["disabled"], "true") || strings.EqualFold(r["disabled"], "yes"),
			Comment:          r["comment"],
			Bytes:            r["bytes"],
			Packets:          r["packets"],
		})
	}
	return out
}
