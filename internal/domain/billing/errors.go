package billing

import "errors"

// ErrNotFound indicates the requested billing entity does not exist.
var ErrNotFound = errors.New("billing: not found")
