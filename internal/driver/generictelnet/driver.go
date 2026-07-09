// Package generictelnet implements port.DeviceDriver for any Telnet-CLI
// vendor that doesn't have (or doesn't yet have) its own dedicated driver
// package. It shares its engine with internal/driver/genericssh — see
// internal/driver/genericcli and docs/adr/0004-generic-cli-driver-scrapligo.md.
//
// This is deliberately NOT used by internal/driver/zteolt: ZTE OLT's own
// raw Telnet client is a separate, earlier decision (see
// TECH-STACK-DAN-PERSIAPAN.md §8 — "tidak ada platform definition scrapligo
// untuk ZTE") kept as-is here rather than silently replaced. If ZTE OLT
// turns out to work fine through this generic engine with a custom
// platform YAML, that's a deliberate follow-up decision (its own ADR), not
// an assumption baked into this package.
package generictelnet

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/driver/genericcli"
	"github.com/quixiq/polyglot/internal/port"
)

// defaultPort is Telnet's conventional port, used when target.Port is
// zero. Unlike SSH, scrapligo's own default port is always 22 regardless
// of transport, so this MUST be set explicitly for Telnet targets — see
// options.WithPort in scrapligo's driver/options package.
const defaultPort = 23

// transportType selects scrapligo's built-in Telnet transport, including
// its IAC (Interpret As Command) option negotiation — see
// docs/adr/0004-generic-cli-driver-scrapligo.md for confirmation this is a
// real, complete transport in scrapligo v1.3.3 and not something this
// project had to hand-roll.
const transportType = "telnet"

// Driver implements port.DeviceDriver for any Telnet-CLI vendor without
// its own dedicated driver package — the vendor difference is entirely
// data: a scrapligo platform definition (built-in name, or a custom YAML
// path under internal/platformdef/) for prompt/paging/login handling, plus
// a genericcli.Catalog for this project's own risk classification and
// operation translation. See
// docs/adr/0004-generic-cli-driver-scrapligo.md.
//
// DEVIASI: like genericssh, this package has no commands.go — there is no
// fixed catalog to hold, since Catalog is supplied by the caller per
// device, not hardcoded per vendor.
type Driver struct {
	session *genericcli.Session
}

// Compile-time proof that *Driver actually satisfies port.DeviceDriver —
// required per CLAUDE.md §1.2 for every vendor driver.
var _ port.DeviceDriver = (*Driver)(nil)

// NewDriver connects to target over Telnet using platformDef (anything
// scrapligo's platform.NewPlatform accepts) and catalog (this vendor's
// risk classification + operation translation).
//
// Unlike other vendor drivers' NewDriver(ctx, target), this one necessarily
// takes two more parameters — see genericssh.NewDriver's doc comment for
// the identical rationale.
func NewDriver(ctx context.Context, target device.Target, platformDef any, catalog genericcli.Catalog) (*Driver, error) {
	session, err := genericcli.NewSession(ctx, target, transportType, defaultPort, platformDef, catalog)
	if err != nil {
		return nil, err
	}
	return &Driver{session: session}, nil
}

// Execute runs cmd against the connected device over the persistent Telnet
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

// Close releases the underlying Telnet connection.
func (d *Driver) Close() error {
	return d.session.Close()
}
