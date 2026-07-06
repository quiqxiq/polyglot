package network

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/port"
)

// PushConfig pushes a configuration change to a device. Destructive by
// nature — reuses the exact same gate as ExecuteCommand, so it is HITL-gated
// the moment a vendor's commands.go lists the relevant Raw pattern as
// destructive (see e.g. internal/driver/mikrotik/commands.go).
func PushConfig(ctx context.Context, driver port.DeviceDriver, cmd command.Command) (command.Result, error) {
	return ExecuteCommand(ctx, driver, cmd)
}
