package routing

import (
	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/driver/mikrotik/internal/rosutil"
)

// RoutingRule represents a policy routing rule in RouterOS v7 (/routing/rule).
type RoutingRule struct {
	ID          string `json:"id"`
	Action      string `json:"action"`
	DstAddress  string `json:"dst_address"`
	SrcAddress  string `json:"src_address"`
	Table       string `json:"table"`
	RoutingMark string `json:"routing_mark"`
	Disabled    bool   `json:"disabled"`
	Comment     string `json:"comment"`
}

// NewPrintRoutingRulesCommand queries routing rules with strict proplist.
func NewPrintRoutingRulesCommand() command.Command {
	args := make(map[string]string)
	rosutil.SetProplist(args, ".id", "action", "dst-address", "src-address", "table", "routing-mark", "disabled", "comment")
	return command.Command{
		Raw:  "/routing/rule/print",
		Args: args,
	}
}

// NewAddRoutingRuleCommand creates a policy routing rule.
func NewAddRoutingRuleCommand(rule RoutingRule) command.Command {
	args := make(map[string]string)
	if rule.Action != "" {
		args["action"] = rule.Action
	}
	if rule.DstAddress != "" {
		args["dst-address"] = rule.DstAddress
	}
	if rule.SrcAddress != "" {
		args["src-address"] = rule.SrcAddress
	}
	if rule.Table != "" {
		args["table"] = rule.Table
	}
	if rule.RoutingMark != "" {
		args["routing-mark"] = rule.RoutingMark
	}
	if rule.Comment != "" {
		args["comment"] = rule.Comment
	}
	if rule.Disabled {
		args["disabled"] = "yes"
	}
	return command.Command{
		Raw:  "/routing/rule/add",
		Args: args,
	}
}

// NewSetRoutingRuleCommand updates a policy routing rule.
func NewSetRoutingRuleCommand(rule RoutingRule) command.Command {
	args := map[string]string{
		".id":          rule.ID,
		"action":       rule.Action,
		"dst-address":  rule.DstAddress,
		"src-address":  rule.SrcAddress,
		"table":        rule.Table,
		"routing-mark": rule.RoutingMark,
		"comment":      rule.Comment,
	}
	if rule.Disabled {
		args["disabled"] = "yes"
	} else {
		args["disabled"] = "no"
	}
	return command.Command{
		Raw:  "/routing/rule/set",
		Args: args,
	}
}

// NewRemoveRoutingRuleCommand deletes a policy routing rule by ID.
func NewRemoveRoutingRuleCommand(id string) command.Command {
	return command.Command{
		Raw: "/routing/rule/remove",
		Args: map[string]string{
			".id": id,
		},
	}
}

// NewMoveRoutingRuleCommand reorders a routing rule before destination.
func NewMoveRoutingRuleCommand(id, destination string) command.Command {
	args := map[string]string{
		"numbers": id,
	}
	if destination != "" {
		args["destination"] = destination
	}
	return command.Command{
		Raw:  "/routing/rule/move",
		Args: args,
	}
}

// ParseRoutingRules converts reply rows into RoutingRule slice.
func ParseRoutingRules(rows []map[string]string) []RoutingRule {
	out := make([]RoutingRule, 0, len(rows))
	for _, r := range rows {
		out = append(out, RoutingRule{
			ID:          r[".id"],
			Action:      r["action"],
			DstAddress:  r["dst-address"],
			SrcAddress:  r["src-address"],
			Table:       r["table"],
			RoutingMark: r["routing-mark"],
			Disabled:    r["disabled"] == "true",
			Comment:     r["comment"],
		})
	}
	return out
}
