package mikrotik

import (
	"strings"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/port"
)

// DHCPLease is the vendor-neutral DHCP lease row.
// Canonical definition lives in internal/port (see port.DHCPLease docs).
type DHCPLease = port.DHCPLease

// DHCPServer represents one row returned by /ip/dhcp-server/print.
// Servers are read for monitoring only — no create/modify/delete.
type DHCPServer struct {
	RosID       string
	Name        string
	Interface   string
	AddressPool string
	LeaseTime   string
	Disabled    bool
}

// DHCPLeaseBlockParams is the vendor-neutral DHCP lease block/unblock parameter set.
// Canonical definition lives in internal/port (see port.DHCPLeaseBlockParams docs).
type DHCPLeaseBlockParams = port.DHCPLeaseBlockParams

// NewPrintDHCPLeasesCommand builds the command.Command for
// /ip/dhcp-server/lease/print. Pass a non-empty macFilter to find the lease
// for a specific MAC address — the primary lookup key for DHCP suspension.
func NewPrintDHCPLeasesCommand(macFilter string) command.Command {
	args := map[string]string{}
	if macFilter != "" {
		args["?mac-address"] = macFilter
	}
	return command.Command{
		Raw:  "/ip/dhcp-server/lease/print",
		Args: args,
	}
}

// NewStreamDHCPLeasesCommand builds the command.Command for
// /ip/dhcp-server/lease/print follow, which causes RouterOS to push a row
// every time a DHCP lease changes state (bound, expired, offered, blocked).
// Use with Driver.Stream — isStreamingCommand returns true because
// "follow" is present.
//
// Parse each command.Result from StreamHandle.Chan() with ParseDHCPLeases.
// Useful for real-time monitoring of DHCP binding status on a dashboard.
//
// macFilter: limit to one MAC address (empty = all leases).
func NewStreamDHCPLeasesCommand(macFilter string) command.Command {
	args := map[string]string{"follow": ""}
	if macFilter != "" {
		args["?mac-address"] = macFilter
	}
	return command.Command{
		Raw:  "/ip/dhcp-server/lease/print",
		Args: args,
	}
}

// NewSetDHCPLeaseBlockCommand builds the command.Command for
// /ip/dhcp-server/lease/set to block or unblock a lease. This is the only
// mutation the application performs on DHCP leases.
//
// rosID must come from a prior /ip/dhcp-server/lease/print result.
// Blocking a lease forces RouterOS to refuse the client's next renewal,
// effectively disconnecting them from the network after their current lease
// expires (or sooner on some RouterOS versions).
func NewSetDHCPLeaseBlockCommand(rosID string, p DHCPLeaseBlockParams) command.Command {
	blocked := "no"
	if p.Blocked {
		blocked = "yes"
	}
	args := map[string]string{
		".id":     rosID,
		"blocked": blocked,
	}
	setIfNonEmpty(args, "comment", p.Comment)
	return command.Command{Raw: "/ip/dhcp-server/lease/set", Args: args}
}

// NewPrintDHCPServersCommand builds the command.Command for
// /ip/dhcp-server/print (list all DHCP server configurations).
func NewPrintDHCPServersCommand() command.Command {
	return command.Command{
		Raw:  "/ip/dhcp-server/print",
		Args: map[string]string{},
	}
}

// ParseDHCPLeases converts command.Result rows from /ip/dhcp-server/lease/print
// into typed DHCPLease values. Rows missing ".id" are skipped.
func ParseDHCPLeases(result command.Result) []DHCPLease {
	leases := make([]DHCPLease, 0, len(result.Rows))
	for _, row := range result.Rows {
		id := row[".id"]
		if id == "" {
			continue
		}
		leases = append(leases, DHCPLease{
			RosID:         id,
			Address:       row["address"],
			MACAddress:    row["mac-address"],
			ClientID:      row["client-id"],
			Server:        row["server"],
			Status:        row["status"],
			ActiveAddress: row["active-address"],
			HostName:      row["host-name"],
			Blocked:       strings.EqualFold(row["blocked"], "true"),
			Comment:       row["comment"],
		})
	}
	return leases
}

// ParseDHCPServers converts command.Result rows from /ip/dhcp-server/print
// into typed DHCPServer values. Rows missing ".id" or "name" are skipped.
func ParseDHCPServers(result command.Result) []DHCPServer {
	servers := make([]DHCPServer, 0, len(result.Rows))
	for _, row := range result.Rows {
		id := row[".id"]
		name := row["name"]
		if id == "" || name == "" {
			continue
		}
		servers = append(servers, DHCPServer{
			RosID:       id,
			Name:        name,
			Interface:   row["interface"],
			AddressPool: row["address-pool"],
			LeaseTime:   row["lease-time"],
			Disabled:    strings.EqualFold(row["disabled"], "true"),
		})
	}
	return servers
}
