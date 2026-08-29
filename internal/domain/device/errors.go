package device

import (
	"errors"

	"github.com/quixiq/polyglot/pkg/fault"
)

// ErrNotFound indicates the requested device does not exist.
var ErrNotFound = fault.New(fault.KindNotFound, "device: not found")

// ErrUnauthorized indicates the caller does not have permission to access the device.
var ErrUnauthorized = fault.New(fault.KindPermissionDenied, "device: unauthorized access")

// ErrDiagnosticsUnconfigured indicates a diagnostics operation (ping, traceroute)
// was requested on a device whose driver has no diagnostics gateway configured.
var ErrDiagnosticsUnconfigured = fault.New(fault.KindFailedPrecondition, "device: diagnostics gateway not configured")

// ErrInvalidInput indicates invalid device inventory input.
var ErrInvalidInput = fault.New(fault.KindInvalidInput, "device: invalid input")

// IsNotFound reports whether err is (or wraps) ErrNotFound.
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}
