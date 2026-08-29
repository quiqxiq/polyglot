// Package response translates errors from the domain and usecase layers into
// ConnectRPC typed errors. It depends only on pkg/fault — never on
// internal/* — so adding a new domain error never requires editing this
// package: declare the sentinel with the right fault.Kind and the mapping
// follows automatically.
package response

import (
	"errors"

	"connectrpc.com/connect"

	"github.com/quixiq/polyglot/pkg/fault"
)

// MapDomainError converts domain and usecase errors into standard ConnectRPC
// typed errors with proper status codes, based on the fault.Kind carried by
// the error (see pkg/fault). Every ConnectRPC handler in
// internal/adapter/connect/ MUST use this single function for error mapping
// (DEVELOPMENT-GUIDELINES.md §6).
//
// Mapping table:
//
//	KindNotFound           → CodeNotFound
//	KindInvalidInput       → CodeInvalidArgument
//	KindAlreadyExists      → CodeAlreadyExists
//	KindPermissionDenied   → CodePermissionDenied
//	KindUnauthenticated    → CodeUnauthenticated
//	KindFailedPrecondition → CodeFailedPrecondition
//	KindConflict           → CodeAborted
//	KindUnavailable        → CodeUnavailable
//	KindResourceExhausted  → CodeResourceExhausted
//	KindUnknown / none     → CodeInternal
func MapDomainError(err error) error {
	if err == nil {
		return nil
	}

	// Already a connect.Error — pass through.
	var connErr *connect.Error
	if errors.As(err, &connErr) {
		return connErr
	}

	switch fault.KindOf(err) {
	case fault.KindNotFound:
		return connect.NewError(connect.CodeNotFound, err)
	case fault.KindInvalidInput:
		return connect.NewError(connect.CodeInvalidArgument, err)
	case fault.KindAlreadyExists:
		return connect.NewError(connect.CodeAlreadyExists, err)
	case fault.KindPermissionDenied:
		return connect.NewError(connect.CodePermissionDenied, err)
	case fault.KindUnauthenticated:
		return connect.NewError(connect.CodeUnauthenticated, err)
	case fault.KindFailedPrecondition:
		return connect.NewError(connect.CodeFailedPrecondition, err)
	case fault.KindConflict:
		return connect.NewError(connect.CodeAborted, err)
	case fault.KindUnavailable:
		return connect.NewError(connect.CodeUnavailable, err)
	case fault.KindResourceExhausted:
		return connect.NewError(connect.CodeResourceExhausted, err)
	default:
		return connect.NewError(connect.CodeInternal, err)
	}
}
