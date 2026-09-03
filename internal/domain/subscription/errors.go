package subscription

import "github.com/quixiq/polyglot/pkg/fault"

// Sentinel errors for the subscription domain.
var (
	ErrNotFound          = fault.New(fault.KindNotFound, "subscription: not found")
	ErrInvalidInput      = fault.New(fault.KindInvalidInput, "subscription: validation failed")
	ErrInvalidTransition = fault.New(fault.KindFailedPrecondition, "subscription: invalid transition")
)
