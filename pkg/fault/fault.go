// Package fault provides typed error kinds shared by the domain and usecase
// layers. It is dependency-free so every layer (domain, usecase, port,
// adapter) may use it without violating the architectural boundary rules.
//
// Sentinel errors are declared with New and carry a Kind; usecase code wraps
// them with fmt.Errorf("...: %w", err) as usual, and KindOf traverses the
// wrap chain to recover the kind. The transport layer (pkg/response) maps
// kinds to ConnectRPC codes without importing any internal package.
package fault

import (
	"errors"
)

// Kind classifies an error for transport-level mapping. The zero value is
// KindUnknown, which maps to an internal error.
type Kind int

const (
	// KindUnknown is the zero value: the error carries no classification.
	KindUnknown Kind = iota
	// KindNotFound indicates the requested entity does not exist.
	KindNotFound
	// KindInvalidInput indicates malformed or missing input.
	KindInvalidInput
	// KindAlreadyExists indicates a uniqueness conflict on create.
	KindAlreadyExists
	// KindPermissionDenied indicates the caller may not perform the action.
	KindPermissionDenied
	// KindUnauthenticated indicates missing, invalid, or expired credentials.
	KindUnauthenticated
	// KindFailedPrecondition indicates a prerequisite is not yet satisfied.
	KindFailedPrecondition
	// KindConflict indicates the request conflicts with current state
	// (e.g. an illegal status transition or an already-paid invoice).
	KindConflict
	// KindUnavailable indicates a dependency (device, gateway) is unreachable.
	KindUnavailable
	// KindResourceExhausted indicates a rate limit or quota was exceeded.
	KindResourceExhausted
)

// String returns a stable, lowercase name for the kind.
func (k Kind) String() string {
	switch k {
	case KindUnknown:
		return "unknown"
	case KindNotFound:
		return "not_found"
	case KindInvalidInput:
		return "invalid_input"
	case KindAlreadyExists:
		return "already_exists"
	case KindPermissionDenied:
		return "permission_denied"
	case KindUnauthenticated:
		return "unauthenticated"
	case KindFailedPrecondition:
		return "failed_precondition"
	case KindConflict:
		return "conflict"
	case KindUnavailable:
		return "unavailable"
	case KindResourceExhausted:
		return "resource_exhausted"
	default:
		return "unknown"
	}
}

// Error is an error carrying a Kind. Declared sentinel errors are *Error
// values; KindOf finds them through fmt.Errorf %w chains via errors.As.
type Error struct {
	kind Kind
	msg  string
	err  error
}

// New returns a sentinel error of the given kind with the given message.
// Message convention: "<domain>: <description>", lowercase English, no
// trailing punctuation (e.g. "billing: invoice not found").
func New(kind Kind, msg string) *Error {
	return &Error{kind: kind, msg: msg}
}

// Wrap attaches kind to err while preserving its message and wrap chain.
// Use it when propagating an error that carries no kind of its own:
//
//	return fault.Wrap(fault.KindUnavailable, err)
//
// It returns nil when err is nil.
func Wrap(kind Kind, err error) error {
	if err == nil {
		return nil
	}
	return &Error{kind: kind, msg: err.Error(), err: err}
}

// Error returns the error message.
func (e *Error) Error() string {
	return e.msg
}

// Unwrap returns the wrapped error, or nil for sentinel roots.
func (e *Error) Unwrap() error {
	return e.err
}

// Kind returns the classification carried by this error.
func (e *Error) Kind() Kind {
	return e.kind
}

// KindOf reports the Kind carried by err or by the nearest *Error in its
// wrap chain. It returns KindUnknown when err carries no kind.
func KindOf(err error) Kind {
	var ferr *Error
	if errors.As(err, &ferr) {
		return ferr.kind
	}
	return KindUnknown
}
