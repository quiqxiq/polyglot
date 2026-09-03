package system

import (
	"context"
	"fmt"

	"github.com/quixiq/polyglot/internal/port"
)

// GetSystemResource fetches system resources from RouterOS.
func GetSystemResource(ctx context.Context, driver port.DeviceDriver, exec port.CommandExecutor) (port.SystemResource, error) {
	cmd := NewPrintResourceCommand()
	res, err := exec(ctx, driver, cmd)
	if err != nil {
		return port.SystemResource{}, fmt.Errorf("mikrotik get system resource: %w", err)
	}
	return ParseResource(res), nil
}

// GetSystemIdentity fetches the identity of the router.
func GetSystemIdentity(ctx context.Context, driver port.DeviceDriver, exec port.CommandExecutor) (Identity, error) {
	cmd := NewPrintIdentityCommand()
	res, err := exec(ctx, driver, cmd)
	if err != nil {
		return Identity{}, fmt.Errorf("mikrotik get identity: %w", err)
	}
	return ParseIdentity(res), nil
}

// GetSystemHealth fetches hardware sensor readings.
func GetSystemHealth(ctx context.Context, driver port.DeviceDriver, exec port.CommandExecutor) (Health, error) {
	cmd := NewPrintHealthCommand()
	res, err := exec(ctx, driver, cmd)
	if err != nil {
		return Health{}, fmt.Errorf("mikrotik get health: %w", err)
	}
	return ParseHealth(res), nil
}

// GetSystemRouterboard fetches RouterBOARD hardware/firmware metadata.
func GetSystemRouterboard(ctx context.Context, driver port.DeviceDriver, exec port.CommandExecutor) (Routerboard, error) {
	cmd := NewPrintRouterboardCommand()
	res, err := exec(ctx, driver, cmd)
	if err != nil {
		return Routerboard{}, fmt.Errorf("mikrotik get routerboard: %w", err)
	}
	return ParseRouterboard(res), nil
}
