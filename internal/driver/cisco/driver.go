package cisco

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/port"
)

// Driver implements port.DeviceDriver for Cisco IOS-XE/XR/NX-OS via
// scrapligo v1.3.3 (github.com/scrapli/scrapligo — the v1 mainline, not
// /v2; see TECH-STACK-DAN-PERSIAPAN.md §7 for why v1.3.3 is the pinned
// choice).
type Driver struct{}

// Compile-time proof that *Driver actually satisfies port.DeviceDriver —
// required in every vendor driver.go per CLAUDE.md §1.2.
var _ port.DeviceDriver = (*Driver)(nil)

// NewDriver connects to target via scrapligo's built-in Cisco platform.
// TODO: use platform.NewPlatform("cisco_iosxe", ...) per TECH-STACK-DAN-PERSIAPAN.md §7.
func NewDriver(ctx context.Context, target device.Target) (*Driver, error) {
	return &Driver{}, nil
}

// Execute runs cmd against the connected device. cmd.Raw is a literal CLI
// string (e.g. "show ip interface brief").
func (d *Driver) Execute(ctx context.Context, cmd command.Command) (command.Result, error) {
	return command.Result{}, nil
}

// Classify reports the risk class of cmd according to Cisco CLI conventions.
func (d *Driver) Classify(cmd command.Command) command.Class {
	return Classify(cmd)
}

// Translate maps an abstract Operation to a Cisco CLI Command.
func (d *Driver) Translate(op command.Operation) (command.Command, error) {
	return Translate(op)
}

// Close releases the underlying SSH connection.
func (d *Driver) Close() error {
	return nil
}
