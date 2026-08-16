package port

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/command"
)

// CommandExecutor runs a vendor-native command against a device through the
// policy gate (classify → decide → execute). It is implemented by
// usecase/network.ExecuteCommand so that gateway implementations in the
// driver layer never bypass the human-in-the-loop policy for destructive
// commands. Drivers build the native command; the policy decision stays in
// the usecase layer.
type CommandExecutor func(ctx context.Context, driver DeviceDriver, cmd command.Command) (command.Result, error)
