package setting

import "github.com/quixiq/polyglot/pkg/fault"

var (
	// ErrNotFound indicates the requested setting was not found.
	ErrNotFound = fault.New(fault.KindNotFound, "setting: not found")
	// ErrInvalidInput indicates invalid setting input.
	ErrInvalidInput = fault.New(fault.KindInvalidInput, "setting: invalid input")
)
