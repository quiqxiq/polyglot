package iface

import (
	"context"
	"fmt"

	"github.com/quixiq/polyglot/internal/port"
)

// ListInterfaces queries RouterOS for interface details.
func ListInterfaces(ctx context.Context, driver port.DeviceDriver, exec port.CommandExecutor, typeFilter, nameFilter string) ([]port.Interface, error) {
	cmd := NewPrintInterfacesCommand(typeFilter, nameFilter)
	res, err := exec(ctx, driver, cmd)
	if err != nil {
		return nil, fmt.Errorf("mikrotik list interfaces: %w", err)
	}
	return ParseInterfaces(res), nil
}

// MonitorTrafficOnce queries RouterOS for a single snapshot of interface traffic rates.
func MonitorTrafficOnce(ctx context.Context, driver port.DeviceDriver, exec port.CommandExecutor, ifaceName string) (port.InterfaceTrafficStats, error) {
	cmd := NewMonitorTrafficOnceCommand(ifaceName)
	res, err := exec(ctx, driver, cmd)
	if err != nil {
		return port.InterfaceTrafficStats{}, fmt.Errorf("mikrotik monitor traffic %q: %w", ifaceName, err)
	}
	return ParseInterfaceTrafficStats(res), nil
}

