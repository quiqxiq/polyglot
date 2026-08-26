package device

import "errors"

// ErrNotFound indicates the requested device does not exist.
var ErrNotFound = errors.New("device: not found")

// ErrUnauthorized indicates the caller does not have permission to access the device.
var ErrUnauthorized = errors.New("device: unauthorized access")

// IsNotFound reports whether err is (or wraps) ErrNotFound.
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}
