package netconf

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/port"
)

// Driver implements port.DeviceDriver for NETCONF-capable devices via
// scrapligo's netconf package (still v1.3.3 mainline — see
// TECH-STACK-DAN-PERSIAPAN.md §7).
type Driver struct{}

// Compile-time proof that *Driver actually satisfies port.DeviceDriver —
// required in every vendor driver.go per CLAUDE.md §1.2.
var _ port.DeviceDriver = (*Driver)(nil)

// NewDriver connects to target via NETCONF.
// TODO: use scrapligo netconf driver, validated per firmware before
// production per TECH-STACK-DAN-PERSIAPAN.md §7.
func NewDriver(ctx context.Context, target device.Target) (*Driver, error) {
	return &Driver{}, nil
}

// Execute runs cmd (a NETCONF operation) against the connected device.
func (d *Driver) Execute(ctx context.Context, cmd command.Command) (command.Result, error) {
	return command.Result{}, nil
}

// Classify reports the risk class of cmd according to NETCONF conventions.
func (d *Driver) Classify(cmd command.Command) command.Class {
	return Classify(cmd)
}

// Translate maps an abstract Operation to a NETCONF Command.
func (d *Driver) Translate(op command.Operation) (command.Command, error) {
	return Translate(op)
}

// Close releases the underlying NETCONF session.
func (d *Driver) Close() error {
	return nil
}
