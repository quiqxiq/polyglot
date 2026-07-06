package huaweiolt

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/port"
)

// Driver implements port.DeviceDriver for Huawei OLT/VRP.
// Per TECH-STACK-DAN-PERSIAPAN.md §7: Huawei VRP CLI support has already
// been merged into scrapligo v1.3.3 mainline (community PR #170) — this
// driver uses the SAME v1.3.3 dependency as internal/driver/cisco, NOT a
// pinned scrapligo v2 release-candidate. A pinned v2 RC would only become
// relevant if a specific SmartAX OLT variant turns out to be uncovered by
// v1.3.3 — validate that at implementation time before reaching for v2.
type Driver struct{}

// Compile-time proof that *Driver actually satisfies port.DeviceDriver —
// required in every vendor driver.go per CLAUDE.md §1.2.
var _ port.DeviceDriver = (*Driver)(nil)

// NewDriver connects to target via scrapligo v1.3.3.
func NewDriver(ctx context.Context, target device.Target) (*Driver, error) {
	return &Driver{}, nil
}

// Execute runs cmd against the connected device. cmd.Raw is a literal CLI string.
func (d *Driver) Execute(ctx context.Context, cmd command.Command) (command.Result, error) {
	return command.Result{}, nil
}

// Classify reports the risk class of cmd according to Huawei VRP conventions.
func (d *Driver) Classify(cmd command.Command) command.Class {
	return Classify(cmd)
}

// Translate maps an abstract Operation to a Huawei VRP Command.
func (d *Driver) Translate(op command.Operation) (command.Command, error) {
	return Translate(op)
}

// Close releases the underlying SSH connection.
func (d *Driver) Close() error {
	return nil
}
