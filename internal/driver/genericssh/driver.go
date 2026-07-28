// Package genericssh implements port.DeviceDriver for any SSH-CLI vendor
// that doesn't have (or doesn't yet have) its own dedicated driver package.
// It shares its engine with internal/driver/generictelnet — see
// internal/driver/genericcli and docs/adr/0004-generic-cli-driver-scrapligo.md.
package genericssh

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/driver/genericcli"
	"github.com/quixiq/polyglot/internal/port"
)

// defaultPort is SSH's conventional port, used when target.Port is zero.
const defaultPort = 22

// transportType selects scrapligo's pure-Go crypto/ssh transport
// ("standard") — deliberately NOT "system" (scrapligo's own default),
// which shells out to /bin/ssh and would break this project's
// single-binary deployment model. See
// docs/adr/0004-generic-cli-driver-scrapligo.md.
const transportType = "standard"

// Driver implements port.DeviceDriver for any SSH-CLI vendor without its
// own dedicated driver package — the vendor difference is entirely data: a
// scrapligo platform definition (built-in name, or a custom YAML path
// under internal/platformdef/) for prompt/paging/login handling, plus a
// genericcli.Catalog for this project's own risk classification and
// operation translation. See
// docs/adr/0004-generic-cli-driver-scrapligo.md.
//
// DEVIASI: unlike every other vendor driver package (mikrotik, cisco, ...),
// this package has no commands.go — there is no fixed catalog to hold,
// since the whole point of "generic" is that Catalog is supplied by the
// caller per device, not hardcoded per vendor. Classify/Translate simply
// delegate to whatever Catalog NewDriver was given.
type Driver struct {
	session *genericcli.Session
}

// Compile-time proof that *Driver actually satisfies port.DeviceDriver —
// required per CLAUDE.md §1.2 for every vendor driver.
var _ port.DeviceDriver = (*Driver)(nil)

// NewDriver connects to target over SSH using platformDef (anything
// scrapligo's platform.NewPlatform accepts: a built-in platform name, a
// custom YAML file path, or raw YAML/JSON bytes) and catalog (this
// vendor's risk classification + operation translation).
//
// Unlike other vendor drivers' NewDriver(ctx, target), this one necessarily
// takes two more parameters — a generic driver has no vendor of its own to
// hardcode these into. Whatever constructs Driver (eventually
// internal/registry) must supply both per device.
func NewDriver(ctx context.Context, target device.Target, platformDef any, catalog genericcli.Catalog) (*Driver, error) {
	session, err := genericcli.NewSession(ctx, target, transportType, defaultPort, platformDef, catalog)
	if err != nil {
		return nil, err
	}
	return &Driver{session: session}, nil
}

// Execute runs cmd against the connected device over the persistent SSH
// session and waits for its result.
func (d *Driver) Execute(ctx context.Context, cmd command.Command) (command.Result, error) {
	return d.session.Execute(ctx, cmd)
}

// Classify reports the risk class of cmd according to the Catalog this
// Driver was constructed with.
func (d *Driver) Classify(cmd command.Command) command.Class {
	return d.session.Classify(cmd)
}

// Translate maps an abstract Operation to this vendor's native Command,
// according to the Catalog this Driver was constructed with.
func (d *Driver) Translate(op command.Operation) (command.Command, error) {
	return d.session.Translate(op)
}

// Close releases the underlying SSH connection.
func (d *Driver) Close() error {
	return d.session.Close()
}
