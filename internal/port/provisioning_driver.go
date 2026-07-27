package port

import (
	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/domain/provision"
)

// ProvisioningDriver is an OPTIONAL capability a DeviceDriver may also
// implement, for vendors that support object provisioning — creating, listing,
// or reconfiguring device-side objects such as Mikrotik /ppp secret and
// /ppp profile.
//
// This is deliberately a separate interface, not an addition to DeviceDriver
// itself — the same reasoning as StreamingDeviceDriver (see
// streaming_driver.go): most vendors have no such objects (an OLT or a CPE
// managed via GenieACS has no /ppp secret), so folding provisioning into
// DeviceDriver would force every other driver to grow stub methods it can't
// honor, and their `var _ port.DeviceDriver = (*Driver)(nil)` assertions would
// keep passing vacuously — the assertion would stop signaling real support.
//
// A separate interface + type assertion at the call site
// (`pd, ok := driver.(port.ProvisioningDriver)`) keeps provisioning support
// genuinely opt-in and visible per vendor.
type ProvisioningDriver interface {
	DeviceDriver

	// TranslateProvision maps an abstract, vendor-neutral provisioning
	// Operation to this vendor's native command SEQUENCE. It returns a slice
	// because one abstract operation may translate to several ordered commands
	// (e.g. Mikrotik's "change profile" is a profile update followed by an
	// active-session kill); each command is executed and audited independently
	// through the usual usecase/network pipeline. Returns an error if this
	// vendor has no definition for op.
	TranslateProvision(op provision.Operation) ([]command.Command, error)
}
