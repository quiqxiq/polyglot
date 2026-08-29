package dhcp

import (
	"context"
	"fmt"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/port"
)

// ListLeases queries RouterOS for DHCP leases with optional MAC filtering.
func ListLeases(ctx context.Context, driver port.DeviceDriver, exec port.CommandExecutor, macFilter string) ([]port.DHCPLease, error) {
	cmd := NewPrintLeasesCommand(macFilter)
	res, err := exec(ctx, driver, cmd)
	if err != nil {
		return nil, fmt.Errorf("mikrotik dhcp print leases: %w", err)
	}
	return ParseLeases(res), nil
}

// SetLeaseBlock updates the block status and comment of a DHCP lease in RouterOS.
func SetLeaseBlock(ctx context.Context, driver port.DeviceDriver, exec port.CommandExecutor, rosID string, p DHCPLeaseBlockParams) (command.Result, error) {
	cmd := NewSetLeaseBlockCommand(rosID, p)
	return exec(ctx, driver, cmd)
}

// ListServers queries RouterOS for configured DHCP servers.
func ListServers(ctx context.Context, driver port.DeviceDriver, exec port.CommandExecutor) ([]DHCPServer, error) {
	cmd := NewPrintServersCommand()
	res, err := exec(ctx, driver, cmd)
	if err != nil {
		return nil, fmt.Errorf("mikrotik dhcp print servers: %w", err)
	}
	return ParseServers(res), nil
}
