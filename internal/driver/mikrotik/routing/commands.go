package routing

import (
	"github.com/quixiq/polyglot/internal/domain/command"
)

// NewPrintBGPSessionsCommand builds the command.Command for /routing/bgp/session/print.
func NewPrintBGPSessionsCommand(nameFilter string) command.Command {
	args := map[string]string{
		".proplist": ".id,name,remote.address,remote.as,local.address,local.as,state,uptime,prefix-count,disabled,comment",
	}
	if nameFilter != "" {
		args["?name"] = nameFilter
	}
	return command.Command{
		Raw:  "/routing/bgp/session/print",
		Args: args,
	}
}

// NewStreamBGPSessionsCommand builds the command.Command for streaming /routing/bgp/session/print follow.
func NewStreamBGPSessionsCommand(nameFilter string) command.Command {
	args := map[string]string{
		"follow":    "",
		".proplist": ".id,name,remote.address,remote.as,local.address,local.as,state,uptime,prefix-count,disabled,comment",
	}
	if nameFilter != "" {
		args["?name"] = nameFilter
	}
	return command.Command{
		Raw:  "/routing/bgp/session/print",
		Args: args,
	}
}

// NewPrintOSPFNeighborsCommand builds the command.Command for /routing/ospf/neighbor/print.
func NewPrintOSPFNeighborsCommand() command.Command {
	return command.Command{
		Raw: "/routing/ospf/neighbor/print",
		Args: map[string]string{
			".proplist": ".id,instance,area,address,router-id,state,priority,adjacency",
		},
	}
}

// NewPrintRoutingTablesCommand builds the command.Command for /routing/table/print.
func NewPrintRoutingTablesCommand() command.Command {
	return command.Command{
		Raw: "/routing/table/print",
		Args: map[string]string{
			".proplist": ".id,name,fib,disabled,comment",
		},
	}
}
