package dhcp

import (
	"github.com/quixiq/polyglot/internal/domain/command"
)

// NewPrintLeasesCommand builds the command.Command for
// /ip/dhcp-server/lease/print. Pass a non-empty macFilter to find the lease
// for a specific MAC address — the primary lookup key for DHCP suspension.
func NewPrintLeasesCommand(macFilter string) command.Command {
	args := map[string]string{}
	if macFilter != "" {
		args["?mac-address"] = macFilter
	}
	return command.Command{
		Raw:  "/ip/dhcp-server/lease/print",
		Args: args,
	}
}

// NewStreamLeasesCommand builds the command.Command for
// /ip/dhcp-server/lease/print follow, which causes RouterOS to push a row
// every time a DHCP lease changes state (bound, expired, offered, blocked).
//
// Parse each command.Result from StreamHandle.Chan() with ParseLeases.
// Useful for real-time monitoring of DHCP binding status on a dashboard.
//
// macFilter: limit to one MAC address (empty = all leases).
func NewStreamLeasesCommand(macFilter string) command.Command {
	args := map[string]string{"follow": ""}
	if macFilter != "" {
		args["?mac-address"] = macFilter
	}
	return command.Command{
		Raw:  "/ip/dhcp-server/lease/print",
		Args: args,
	}
}

// NewSetLeaseBlockCommand builds the command.Command for
// /ip/dhcp-server/lease/set to block or unblock a lease. This is the only
// mutation the application performs on DHCP leases.
//
// rosID must come from a prior /ip/dhcp-server/lease/print result.
func NewSetLeaseBlockCommand(rosID string, p DHCPLeaseBlockParams) command.Command {
	blocked := "no"
	if p.Blocked {
		blocked = "yes"
	}
	args := map[string]string{
		".id":     rosID,
		"blocked": blocked,
	}
	if p.Comment != "" {
		args["comment"] = p.Comment
	}
	return command.Command{Raw: "/ip/dhcp-server/lease/set", Args: args}
}

// NewPrintServersCommand builds the command.Command for
// /ip/dhcp-server/print (list all DHCP server configurations).
func NewPrintServersCommand() command.Command {
	return command.Command{
		Raw:  "/ip/dhcp-server/print",
		Args: map[string]string{},
	}
}

