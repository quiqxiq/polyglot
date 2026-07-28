package plan

import "errors"

// ErrNotFound indicates the requested plan does not exist.
var ErrNotFound = errors.New("plan: not found")
