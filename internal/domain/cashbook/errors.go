package cashbook

import "github.com/quixiq/polyglot/pkg/fault"

var (
	// ErrNotFound indicates the requested cashbook entity was not found.
	ErrNotFound = fault.New(fault.KindNotFound, "cashbook: not found")
	// ErrInvalidInput indicates the cashbook input failed validation.
	ErrInvalidInput = fault.New(fault.KindInvalidInput, "cashbook: validation failed")
)
