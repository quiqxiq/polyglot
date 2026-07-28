package subscription

import "errors"

// ErrNotFound indicates the requested subscription does not exist.
var ErrNotFound = errors.New("subscription: not found")
