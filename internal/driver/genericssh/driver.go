package genericssh

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/port"
)

// Driver implements port.DeviceDriver for vendors without built-in scrapligo
// support, using a custom platform definition from internal/platformdef/.
type Driver struct{}

// Compile-time proof that *Driver actually satisfies port.DeviceDriver —
// required in every vendor driver.go per CLAUDE.md §1.2.
var _ port.DeviceDriver = (*Driver)(nil)

// NewDriver connects using a custom scrapligo platform YAML definition.
// TODO: load definition via platform.NewPlatform(pathToYAML, ...).
func NewDriver(ctx context.Context, target device.Target) (*Driver, error) {
	return &Driver{}, nil
}

// Execute runs cmd against the connected device. cmd.Raw is a literal CLI string.
func (d *Driver) Execute(ctx context.Context, cmd command.Command) (command.Result, error) {
	return command.Result{}, nil
}

// Classify reports the risk class of cmd. Generic SSH vendors have no
// curated pattern list by default — see commands.go.
func (d *Driver) Classify(cmd command.Command) command.Class {
	return Classify(cmd)
}

// Translate maps an abstract Operation to a Command for this vendor.
// Always errors until a specific platformdef/catalog is curated for it.
func (d *Driver) Translate(op command.Operation) (command.Command, error) {
	return Translate(op)
}

// Close releases the underlying SSH connection.
func (d *Driver) Close() error {
	return nil
}
