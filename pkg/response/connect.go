package response

import (
	"errors"

	"connectrpc.com/connect"

	"github.com/quixiq/polyglot/pkg/fault"
)

// Unavailable returns a sanitized dependency-unavailable ConnectRPC error.
func Unavailable(message string) error {
	return connect.NewError(connect.CodeUnavailable, errors.New(message))
}

// InvalidArgument returns a ConnectRPC error for transport-level input errors.
func InvalidArgument(message string) error {
	return connect.NewError(connect.CodeInvalidArgument, errors.New(message))
}

// Unauthenticated returns a ConnectRPC error for invalid transport credentials.
func Unauthenticated(message string) error {
	return connect.NewError(connect.CodeUnauthenticated, errors.New(message))
}

// PermissionDenied returns a ConnectRPC error for transport-level authorization errors.
func PermissionDenied(message string) error {
	return connect.NewError(connect.CodePermissionDenied, errors.New(message))
}

// Unimplemented returns a ConnectRPC error for an unsupported capability.
func Unimplemented(message string) error {
	return connect.NewError(connect.CodeUnimplemented, errors.New(message))
}

// Internal returns a sanitized ConnectRPC internal error.
func Internal(message string) error {
	return connect.NewError(connect.CodeInternal, errors.New(message))
}

// MapTransportError maps a non-domain transport error using the shared fault taxonomy.
// Use this when a transport-only condition must retain a stable public status.
func MapTransportError(kind fault.Kind, message string) error {
	return MapDomainError(fault.New(kind, message))
}
