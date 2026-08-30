package iface

import (
	"github.com/quixiq/polyglot/internal/port"
)

// Interface is the vendor-neutral network interface row.
// Canonical definition lives in internal/port.
type Interface = port.Interface

// InterfaceTrafficStats is the vendor-neutral interface traffic rate snapshot.
// Canonical definition lives in internal/port.
type InterfaceTrafficStats = port.InterfaceTrafficStats
