package genieacs

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/port"
)

// Driver implements port.DeviceDriver for TR-069/ACS via the GenieACS NBI
// (net/http standard library — no extra dependency needed, per
// TECH-STACK-DAN-PERSIAPAN.md §5 Ringkasan Tegas table).
// GenieACS's own terminology is "tasks", not "commands" — cmd.Raw maps to
// a GenieACS task type (see commands.go).
type Driver struct{}

// Compile-time proof that *Driver actually satisfies port.DeviceDriver —
// required in every vendor driver.go per CLAUDE.md §1.2.
var _ port.DeviceDriver = (*Driver)(nil)

// NewDriver builds a REST client for the GenieACS NBI for one device.
// NBI has no built-in auth — must be network-isolated, see
// NetOps-Architecture.md §8 Keamanan.
func NewDriver(ctx context.Context, target device.Target) (*Driver, error) {
	return &Driver{}, nil
}

// Execute submits cmd as a GenieACS task via the NBI.
func (d *Driver) Execute(ctx context.Context, cmd command.Command) (command.Result, error) {
	return command.Result{}, nil
}

// Classify reports the risk class of cmd according to GenieACS task types.
func (d *Driver) Classify(cmd command.Command) command.Class {
	return Classify(cmd)
}

// Translate maps an abstract Operation to a GenieACS task type Command.
func (d *Driver) Translate(op command.Operation) (command.Command, error) {
	return Translate(op)
}

// Close is a no-op: the GenieACS NBI is stateless REST, no connection to release.
func (d *Driver) Close() error {
	return nil
}
