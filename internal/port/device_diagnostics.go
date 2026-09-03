package port

import (
	"context"

	domainDevice "github.com/quixiq/polyglot/internal/domain/device"
)

// SystemIdentity alias to domain model.
type SystemIdentity = domainDevice.SystemIdentity

// LogEntry alias to domain model.
type LogEntry = domainDevice.LogEntry

// SystemHealth alias to domain model.
type SystemHealth = domainDevice.SystemHealth

// SystemClock alias to domain model.
type SystemClock = domainDevice.SystemClock

// SystemRouterboard alias to domain model.
type SystemRouterboard = domainDevice.SystemRouterboard

// DeviceDiagnostics abstracts vendor-specific live diagnostics operations
// used for connectivity testing and monitoring. It is implemented by the
// mikrotik driver package and executed through the injected
// CommandExecutor.
type DeviceDiagnostics interface {
	// GetSystemResource fetches router system resource metrics.
	GetSystemResource(ctx context.Context, driver DeviceDriver) (SystemResource, error)
	// GetSystemIdentity fetches the router's configured identity (name).
	GetSystemIdentity(ctx context.Context, driver DeviceDriver) (string, error)
	// ListInterfaces fetches /interface/print entries with optional type and name filters.
	ListInterfaces(ctx context.Context, driver DeviceDriver, typeFilter, nameFilter string) ([]Interface, error)
	// MonitorTrafficOnce returns a single snapshot of current traffic rates
	// for the given interface.
	MonitorTrafficOnce(ctx context.Context, driver DeviceDriver, ifaceName string) (InterfaceTrafficStats, error)
}
