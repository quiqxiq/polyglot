package mikrotik

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/port"
)

// Driver implements port.DeviceDriver for Mikrotik RouterOS via
// go-routeros/v3 (github.com/go-routeros/routeros/v3).
// A Driver instance represents one already-open connection to one device.
type Driver struct{}

// Compile-time proof that *Driver actually satisfies port.DeviceDriver —
// without this, a signature mismatch could still pass `go build` as long as
// nothing else assigns *Driver to an interface variable. Required in every
// vendor driver.go per CLAUDE.md §1.2.
var _ port.DeviceDriver = (*Driver)(nil)

// NewDriver connects to target and returns a ready, connected Driver.
// TODO: use routeros.DialContext (the *Context variant — always use it,
// per CLAUDE.md §5 "context.Context selalu ada") per
// TECH-STACK-DAN-PERSIAPAN.md §6.
func NewDriver(ctx context.Context, target device.Target) (*Driver, error) {
	return &Driver{}, nil
}

// Execute runs cmd against the connected RouterOS device.
// cmd.Raw is a RouterOS API path (e.g. "/interface/print"), cmd.Args are
// sentence attributes.
func (d *Driver) Execute(ctx context.Context, cmd command.Command) (command.Result, error) {
	return command.Result{}, nil
}

// Classify reports the risk class of cmd according to RouterOS API conventions.
func (d *Driver) Classify(cmd command.Command) command.Class {
	return Classify(cmd)
}

// Translate maps an abstract Operation to a RouterOS-native Command.
func (d *Driver) Translate(op command.Operation) (command.Command, error) {
	return Translate(op)
}

// Close releases the underlying RouterOS API connection.
func (d *Driver) Close() error {
	return nil
}
