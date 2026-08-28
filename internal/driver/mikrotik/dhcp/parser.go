package dhcp

import (
	"strings"

	"github.com/quixiq/polyglot/internal/domain/command"
)

// ParseLeases converts command.Result rows from /ip/dhcp-server/lease/print
// into typed DHCPLease values. Rows missing ".id" are skipped.
func ParseLeases(result command.Result) []DHCPLease {
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

// ParseServers converts command.Result rows from /ip/dhcp-server/print
// into typed DHCPServer values. Rows missing ".id" or "name" are skipped.
func ParseServers(result command.Result) []DHCPServer {
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

