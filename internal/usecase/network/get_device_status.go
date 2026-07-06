package network

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/port"
)

// GetDeviceStatus retrieves the current status of a device via its driver's
// curated "get_status" operation. Full worked example of the curated
// (non-raw) pipeline: Translate -> ExecuteCommand (which itself does
// Classify -> Decide -> Execute). See NetOps-Architecture.md §5.3.
func GetDeviceStatus(ctx context.Context, driver port.DeviceDriver) (command.Result, error) {
	cmd, err := driver.Translate(command.OpGetStatus)
	if err != nil {
		return command.Result{}, err
	}
	return ExecuteCommand(ctx, driver, cmd)
}
