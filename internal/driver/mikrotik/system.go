package mikrotik

import (
	"github.com/quixiq/polyglot/internal/domain/command"
)

// NewRebootCommand builds the command.Command for /system/reboot.
func NewRebootCommand() command.Command {
	return command.Command{
		Raw:  "/system/reboot",
		Args: map[string]string{},
	}
}

// NewShutdownCommand builds the command.Command for /system/shutdown.
func NewShutdownCommand() command.Command {
	return command.Command{
		Raw:  "/system/shutdown",
		Args: map[string]string{},
	}
}
