package registration

import "github.com/quixiq/polyglot/pkg/fault"

// Sentinel errors for the registration domain.
var (
	// ErrNotFound indicates the registration request does not exist.
	ErrNotFound = fault.New(fault.KindNotFound, "registration: not found")
	// ErrInvalidInput indicates the registration payload fails validation.
	ErrInvalidInput = fault.New(fault.KindInvalidInput, "registration: validation failed")
	// ErrInvalidTransition indicates an illegal registration status transition.
	ErrInvalidTransition = fault.New(fault.KindFailedPrecondition, "registration: invalid status transition")
)
