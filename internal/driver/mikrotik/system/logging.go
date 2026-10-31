package system

import (
	"strconv"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/driver/mikrotik/internal/rosutil"
)

// LogRule represents a RouterOS system logging rule.
type LogRule struct {
	ID       string `json:"id"`
	Topics   string `json:"topics"`
	Action   string `json:"action"`
	Prefix   string `json:"prefix"`
	Disabled bool   `json:"disabled"`
}

// LogAction represents a logging action destination (memory, disk, echo, remote).
type LogAction struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Target     string `json:"target"`
	Remote     string `json:"remote"`
	RemotePort int    `json:"remote_port"`
}

// NewPrintLogRulesCommand queries logging rules with strict proplist.
func NewPrintLogRulesCommand() command.Command {
	args := make(map[string]string)
	rosutil.SetProplist(args, ".id", "topics", "action", "prefix", "disabled")
	return command.Command{
		Raw:  "/system/logging/print",
		Args: args,
	}
}

// NewAddLogRuleCommand creates a logging rule.
func NewAddLogRuleCommand(rule LogRule) command.Command {
	args := map[string]string{
		"topics": rule.Topics,
		"action": rule.Action,
	}
	if rule.Prefix != "" {
		args["prefix"] = rule.Prefix
	}
	if rule.Disabled {
		args["disabled"] = "yes"
	}
	return command.Command{
		Raw:  "/system/logging/add",
		Args: args,
	}
}

// NewSetLogRuleCommand updates an existing logging rule.
func NewSetLogRuleCommand(rule LogRule) command.Command {
	args := map[string]string{
		".id":    rule.ID,
		"topics": rule.Topics,
		"action": rule.Action,
		"prefix": rule.Prefix,
	}
	if rule.Disabled {
		args["disabled"] = "yes"
	} else {
		args["disabled"] = "no"
	}
	return command.Command{
		Raw:  "/system/logging/set",
		Args: args,
	}
}

// NewRemoveLogRuleCommand deletes a logging rule by ID.
func NewRemoveLogRuleCommand(id string) command.Command {
	return command.Command{
		Raw: "/system/logging/remove",
		Args: map[string]string{
			".id": id,
		},
	}
}

// NewPrintLogActionsCommand queries logging targets with strict proplist.
func NewPrintLogActionsCommand() command.Command {
	args := make(map[string]string)
	rosutil.SetProplist(args, ".id", "name", "target", "remote", "remote-port")
	return command.Command{
		Raw:  "/system/logging/action/print",
		Args: args,
	}
}

// NewAddLogActionCommand creates a logging action destination.
func NewAddLogActionCommand(act LogAction) command.Command {
	args := map[string]string{
		"name":   act.Name,
		"target": act.Target,
	}
	if act.Remote != "" {
		args["remote"] = act.Remote
	}
	if act.RemotePort > 0 {
		args["remote-port"] = strconv.Itoa(act.RemotePort)
	}
	return command.Command{
		Raw:  "/system/logging/action/add",
		Args: args,
	}
}

// NewSetLogActionCommand updates a logging action.
func NewSetLogActionCommand(act LogAction) command.Command {
	args := map[string]string{
		".id":    act.ID,
		"name":   act.Name,
		"target": act.Target,
		"remote": act.Remote,
	}
	if act.RemotePort > 0 {
		args["remote-port"] = strconv.Itoa(act.RemotePort)
	}
	return command.Command{
		Raw:  "/system/logging/action/set",
		Args: args,
	}
}

// NewRemoveLogActionCommand deletes a logging action by ID.
func NewRemoveLogActionCommand(id string) command.Command {
	return command.Command{
		Raw: "/system/logging/action/remove",
		Args: map[string]string{
			".id": id,
		},
	}
}

// ParseLogRules converts reply rows into LogRule slice.
func ParseLogRules(rows []map[string]string) []LogRule {
	out := make([]LogRule, 0, len(rows))
	for _, s := range rows {
		out = append(out, LogRule{
			ID:       s[".id"],
			Topics:   s["topics"],
			Action:   s["action"],
			Prefix:   s["prefix"],
			Disabled: s["disabled"] == "true",
		})
	}
	return out
}

// ParseLogActions converts reply rows into LogAction slice.
func ParseLogActions(rows []map[string]string) []LogAction {
	out := make([]LogAction, 0, len(rows))
	for _, s := range rows {
		port, _ := strconv.Atoi(s["remote-port"])
		out = append(out, LogAction{
			ID:         s[".id"],
			Name:       s["name"],
			Target:     s["target"],
			Remote:     s["remote"],
			RemotePort: port,
		})
	}
	return out
}
