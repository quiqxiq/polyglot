package port

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/command"
)

// ValidatingDeviceDriver is an OPTIONAL capability a DeviceDriver may also
// implement, for vendors whose driver can pre-flight a command against the
// device's own parser and schema BEFORE it is executed (e.g. Mikrotik via
// goros' Gate 1 `:parse` and Gate 2 `/console/inspect` validation).
//
// This is deliberately a separate interface, not an addition to
// DeviceDriver itself — the same reasoning as StreamingDeviceDriver: adding
// Validate directly onto DeviceDriver would force every other vendor driver
// (cisco, genericssh, netconf, zteolt, huaweiolt, genieacs) to grow a stub
// method they don't support, and their compile-time
// `var _ port.DeviceDriver = (*Driver)(nil)` assertions would keep passing
// vacuously either way. A separate interface + type assertion at the call
// site (`v, ok := driver.(port.ValidatingDeviceDriver)`) keeps validation
// genuinely opt-in and visible per vendor.
type ValidatingDeviceDriver interface {
	DeviceDriver

	// Validate dry-runs cmd against the device without executing it. It
	// returns nil when the device accepts the command — or when the
	// session cannot validate (e.g. RouterOS v6 without /console/inspect,
	// where validation degrades silently by design). It returns a
	// descriptive error when the command is invalid (unknown path,
	// unknown attribute, syntax error), so callers can reject it before
	// it ever reaches the device.
	Validate(ctx context.Context, cmd command.Command) error
}
