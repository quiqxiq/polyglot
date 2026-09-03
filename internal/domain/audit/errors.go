package audit

import "github.com/quixiq/polyglot/pkg/fault"

var (
	// ErrNotFound indicates the requested audit record was not found.
	ErrNotFound = fault.New(fault.KindNotFound, "audit: not found")
	// ErrInvalidInput indicates invalid audit query parameters.
	ErrInvalidInput = fault.New(fault.KindInvalidInput, "audit: invalid parameters")
)
