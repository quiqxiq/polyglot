package port

import "github.com/quixiq/polyglot/internal/domain/device"

// Interface represents one row returned by /interface/print.
// Aliased to domain model per DEVELOPMENT-GUIDELINES.md §4.2.
type Interface = device.Interface

// InterfaceTrafficStats holds the real-time traffic statistics returned by
// /interface/monitor-traffic.
// Aliased to domain model per DEVELOPMENT-GUIDELINES.md §4.2.
type InterfaceTrafficStats = device.InterfaceTrafficStats
