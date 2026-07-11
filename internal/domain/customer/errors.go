package customer

import "errors"

// ErrNotFound indicates the requested customer does not exist.
var ErrNotFound = errors.New("customer: not found")
