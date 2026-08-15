package port

import (
	"context"
	"io"

	"github.com/quixiq/polyglot/internal/domain/command"
)

// DeviceDriver is the contract every vendor-specific driver must implement.
type DeviceDriver interface {
	// Execute runs cmd against the already-connected device.
	Execute(ctx context.Context, cmd command.Command) (command.Result, error)
	// Classify reports the risk class of cmd according to this vendor's own conventions.
	Classify(cmd command.Command) command.Class
	// Translate maps an abstract Operation to this vendor's native Command.
	Translate(op command.Operation) (command.Command, error)
	// Close releases the underlying connection.
	Close() error
}

// DriverProvider defines a function that provides an active DeviceDriver by device ID.
type DriverProvider func(ctx context.Context, deviceID string) (DeviceDriver, error)

// TerminalSession represents an interactive PTY session (SSH/Telnet) connected to a device.
type TerminalSession interface {
	Stdin() io.Writer
	Stdout() io.Reader
	Resize(cols, rows int) error
	Close() error
}

// TerminalDeviceDriver is an optional interface implemented by drivers that support interactive PTY SSH/Telnet streaming.
type TerminalDeviceDriver interface {
	OpenTerminalSession(ctx context.Context, cols, rows int) (TerminalSession, error)
}
