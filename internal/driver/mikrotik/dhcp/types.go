package dhcp

import (
	domainDevice "github.com/quixiq/polyglot/internal/domain/device"
)

// Lease is the vendor-neutral DHCP lease row.
// Canonical definition lives in internal/domain/device.
type Lease = domainDevice.DHCPLease

// LeaseBlockParams is the vendor-neutral DHCP lease block/unblock parameter set.
// Canonical definition lives in internal/domain/device.
type LeaseBlockParams = domainDevice.DHCPLeaseBlockParams

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
