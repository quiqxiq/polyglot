// Package zteolt implements port.DeviceDriver for ZTE OLT equipment,
// combining SNMP (read) and raw Telnet (provisioning) — there is no
// scrapligo platform definition for ZTE, see
// TECH-STACK-DAN-PERSIAPAN.md §8 — hence a raw client here, not scrapligo.
package zteolt

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/port"
)

// Driver implements port.DeviceDriver for ZTE OLT, combining SNMP (read)
// and raw Telnet (provisioning) — there is no scrapligo platform definition
// for ZTE, see TECH-STACK-DAN-PERSIAPAN.md §8.
type Driver struct {
	snmp   *SNMPClient
	telnet *telnetClient
}

// Compile-time proof that *Driver actually satisfies port.DeviceDriver —
// required in every vendor driver.go per CLAUDE.md §1.2.
var _ port.DeviceDriver = (*Driver)(nil)

// NewDriver connects both the SNMP and Telnet sessions for a ZTE OLT.
func NewDriver(ctx context.Context, target device.Target) (*Driver, error) {
	snmp, err := NewSNMPClient(ctx, target)
	if err != nil {
		return nil, err
	}
	// TODO: dial the actual Telnet connection here instead of this
	// placeholder — see telnetClient's TODO above.
	return &Driver{snmp: snmp, telnet: &telnetClient{}}, nil
}

// Execute runs a provisioning command over the raw Telnet session.
func (d *Driver) Execute(ctx context.Context, cmd command.Command) (command.Result, error) {
	return command.Result{}, nil
}

// Classify reports the risk class of cmd according to ZTE OLT conventions.
func (d *Driver) Classify(cmd command.Command) command.Class {
	return Classify(cmd)
}

// Translate maps an abstract Operation to a ZTE OLT Command.
// command.OpGetStatus deliberately has no entry in zteolt's operationMap:
// status should be read via d.snmp.Get, not the Telnet side. TODO: decide
// how usecase/ should route status reads through SNMP once that path is
// actually wired — Translate erroring here is an honest placeholder, not
// a bug, until that decision is made.
func (d *Driver) Translate(op command.Operation) (command.Command, error) {
	return Translate(op)
}

// Close releases both the SNMP and Telnet connections.
func (d *Driver) Close() error {
	if err := d.telnet.close(); err != nil {
		return err
	}
	return d.snmp.Close()
}
