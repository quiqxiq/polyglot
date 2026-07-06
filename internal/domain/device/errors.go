package device

import "errors"

// ErrNotFound indicates the requested device does not exist.
var ErrNotFound = errors.New("device: not found")

// IsNotFound reports whether err is (or wraps) ErrNotFound.
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}
