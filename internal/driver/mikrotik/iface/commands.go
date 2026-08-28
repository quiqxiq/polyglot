package iface

import (
	"github.com/quixiq/polyglot/internal/domain/command"
)

// NewPrintInterfacesCommand builds the command.Command for /interface/print.
func NewPrintInterfacesCommand(typeFilter, nameFilter string) command.Command {
	args := map[string]string{
		".proplist": ".id,name,type,mac-address,running,disabled",
	}
	if typeFilter != "" {
		args["?type"] = typeFilter
	}
	if nameFilter != "" {
		args["?name"] = nameFilter
	}
	return command.Command{
		Raw:  "/interface/print",
		Args: args,
	}
}

// NewStreamInterfacesCommand builds the command.Command for streaming /interface/print interval=<n>.
func NewStreamInterfacesCommand(typeFilter, nameFilter, interval string) command.Command {
	if interval == "" {
		interval = "1s"
	}
	args := map[string]string{
		"interval":  interval,
		".proplist": ".id,name,type,mac-address,running,disabled",
	}
	if typeFilter != "" {
		args["?type"] = typeFilter
	}
	if nameFilter != "" {
		args["?name"] = nameFilter
	}
	return command.Command{
		Raw:  "/interface/print",
		Args: args,
	}
}

// NewMonitorTrafficOnceCommand builds the command.Command for /interface/monitor-traffic once.
func NewMonitorTrafficOnceCommand(ifaceName string) command.Command {
	return command.Command{
		Raw: "/interface/monitor-traffic",
		Args: map[string]string{
			"interface": ifaceName,
			"once":      "",
			".proplist": "name,rx-bits-per-second,tx-bits-per-second,rx-packets-per-second,tx-packets-per-second",
		},
	}
}

// NewMonitorTrafficStreamCommand builds the command.Command for /interface/monitor-traffic streaming.
func NewMonitorTrafficStreamCommand(ifaceName string) command.Command {
	return command.Command{
		Raw: "/interface/monitor-traffic",
		Args: map[string]string{
			"interface": ifaceName,
			".proplist": "name,rx-bits-per-second,tx-bits-per-second,rx-packets-per-second,tx-packets-per-second",
		},
	}
}

// NewEnableInterfaceCommand builds the command.Command for /interface/enable.
func NewEnableInterfaceCommand(rosID string) command.Command {
	return command.Command{
		Raw:  "/interface/enable",
		Args: map[string]string{".id": rosID},
	}
}

// NewDisableInterfaceCommand builds the command.Command for /interface/disable.
func NewDisableInterfaceCommand(rosID string) command.Command {
	return command.Command{
		Raw:  "/interface/disable",
		Args: map[string]string{".id": rosID},
	}
}

