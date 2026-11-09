package routerjob

import "github.com/quixiq/polyglot/pkg/fault"

// Sentinel errors domain antrean job router.
var (
	// ErrNotFound indicates the router job does not exist.
	ErrNotFound = fault.New(fault.KindNotFound, "routerjob: not found")
	// ErrInvalidTransition indicates an illegal job status transition.
	ErrInvalidTransition = fault.New(fault.KindFailedPrecondition, "routerjob: invalid status transition")
	// ErrUnsupportedType indicates an unknown job type.
	ErrUnsupportedType = fault.New(fault.KindInvalidInput, "routerjob: unsupported job type")
	// ErrInvalidInput indicates invalid job payload or parameters.
	ErrInvalidInput = fault.New(fault.KindInvalidInput, "routerjob: invalid input")
	// ErrHandlerNotFound indicates no handler is registered for the job type.
	ErrHandlerNotFound = fault.New(fault.KindFailedPrecondition, "routerjob: no handler registered")
	// ErrQueueNotConfigured indicates the router job queue or action executor is not configured.
	ErrQueueNotConfigured = fault.New(fault.KindFailedPrecondition, "routerjob: queue not configured")
)
