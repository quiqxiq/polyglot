package netwatch

import (
	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/driver/mikrotik/internal/rosutil"
)

// NewPrintNetwatchCommand queries monitored hosts with strict proplist.
func NewPrintNetwatchCommand() command.Command {
	args := make(map[string]string)
	rosutil.SetProplist(args, ".id", "host", "interval", "timeout", "status", "since", "up-script", "down-script", "disabled", "comment")
	return command.Command{
		Raw:  "/tool/netwatch/print",
		Args: args,
	}
}

// NewStreamNetwatchCommand builds a Mode A follow command for real-time netwatch status updates.
func NewStreamNetwatchCommand() command.Command {
	args := map[string]string{
		"follow": "",
	}
	rosutil.SetProplist(args, ".id", "host", "interval", "timeout", "status", "since", "up-script", "down-script", "disabled", "comment")
	return command.Command{
		Raw:  "/tool/netwatch/print",
		Args: args,
	}
}

// NewAddNetwatchCommand adds a host to netwatch.
func NewAddNetwatchCommand(n NetwatchHost) command.Command {
	args := map[string]string{
		"host": n.Host,
	}
	if n.Interval != "" {
		args["interval"] = n.Interval
	}
	if n.Timeout != "" {
		args["timeout"] = n.Timeout
	}
	if n.UpScript != "" {
		args["up-script"] = n.UpScript
	}
	if n.DownScript != "" {
		args["down-script"] = n.DownScript
	}
	if n.Comment != "" {
		args["comment"] = n.Comment
	}
	if n.Disabled {
		args["disabled"] = "yes"
	}
	return command.Command{
		Raw:  "/tool/netwatch/add",
		Args: args,
	}
}

// NewSetNetwatchCommand updates an existing netwatch entry.
func NewSetNetwatchCommand(n NetwatchHost) command.Command {
	args := map[string]string{
		".id":         n.ID,
		"host":        n.Host,
		"interval":    n.Interval,
		"timeout":     n.Timeout,
		"up-script":   n.UpScript,
		"down-script": n.DownScript,
		"comment":     n.Comment,
	}
	if n.Disabled {
		args["disabled"] = "yes"
	} else {
		args["disabled"] = "no"
	}
	return command.Command{
		Raw:  "/tool/netwatch/set",
		Args: args,
	}
}

// NewRemoveNetwatchCommand deletes a netwatch entry by ID.
func NewRemoveNetwatchCommand(id string) command.Command {
	return command.Command{
		Raw: "/tool/netwatch/remove",
		Args: map[string]string{
			".id": id,
		},
	}
}
