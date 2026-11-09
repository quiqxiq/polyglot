package iface

import (
	domainDevice "github.com/quixiq/polyglot/internal/domain/device"
)

// Interface is the vendor-neutral network interface row.
// Canonical definition lives in internal/domain/device.
type Interface = domainDevice.Interface

// InterfaceTrafficStats is the vendor-neutral interface traffic rate snapshot.
// Canonical definition lives in internal/domain/device.
type InterfaceTrafficStats = domainDevice.InterfaceTrafficStats
