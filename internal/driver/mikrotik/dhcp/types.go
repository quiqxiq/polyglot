package dhcp

import (
	"github.com/quixiq/polyglot/internal/port"
)

// DHCPLease is the vendor-neutral DHCP lease row.
// Canonical definition lives in internal/port.
type DHCPLease = port.DHCPLease

// DHCPLeaseBlockParams is the vendor-neutral DHCP lease block/unblock parameter set.
// Canonical definition lives in internal/port.
type DHCPLeaseBlockParams = port.DHCPLeaseBlockParams

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
