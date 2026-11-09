package port

import "github.com/quixiq/polyglot/internal/domain/device"

// DHCPLease represents one row returned by /ip/dhcp-server/lease/print.
// Aliased to domain model per DEVELOPMENT-GUIDELINES.md §4.2.
type DHCPLease = device.DHCPLease
