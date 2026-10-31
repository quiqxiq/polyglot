package ont

import "github.com/quixiq/polyglot/pkg/fault"

// Sentinel errors domain ONT.
var (
	// ErrNotFound indicates the ONT does not exist.
	ErrNotFound = fault.New(fault.KindNotFound, "ont: not found")
	// ErrInvalidIdentity indicates an invalid MAC/serial.
	ErrInvalidIdentity = fault.New(fault.KindInvalidInput, "ont: invalid identity")
	// ErrAlreadyLinked indicates the ONT is already linked to another service.
	ErrAlreadyLinked = fault.New(fault.KindConflict, "ont: already linked")
	// ErrNotLinked indicates the ONT has no linked customer/subscription.
	ErrNotLinked = fault.New(fault.KindFailedPrecondition, "ont: not linked")
	// ErrInvalidInput indicates invalid ONT operation input.
	ErrInvalidInput = fault.New(fault.KindInvalidInput, "ont: invalid input")
	// ErrACSNotConfigured indicates the GenieACS server is not configured yet.
	ErrACSNotConfigured = fault.New(fault.KindFailedPrecondition, "ont: acs not configured")
	// ErrACSTaskFailed indicates a GenieACS task faulted or never completed.
	ErrACSTaskFailed = fault.New(fault.KindUnavailable, "ont: acs task failed")
)
