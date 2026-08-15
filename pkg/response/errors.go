package response

import (
	"errors"

	"connectrpc.com/connect"
)

// Common standard domain errors.
var (
	ErrNotFound       = errors.New("record not found")
	ErrAlreadyExists  = errors.New("record already exists")
	ErrUnauthorized   = errors.New("unauthorized access")
	ErrForbidden      = errors.New("permission denied")
	ErrInvalidInput   = errors.New("invalid request parameters")
	ErrInternal       = errors.New("internal server error")
	ErrUnavailable    = errors.New("service temporarily unavailable")
	ErrDeviceOffline  = errors.New("target device is offline or unreachable")
)

// ToConnectError maps standard domain and standard Go errors into appropriate ConnectRPC errors.
func ToConnectError(err error) *connect.Error {
	if err == nil {
		return nil
	}

	var connErr *connect.Error
	if errors.As(err, &connErr) {
		return connErr
	}

	switch {
	case errors.Is(err, ErrNotFound):
		return connect.NewError(connect.CodeNotFound, err)
	case errors.Is(err, ErrUnauthorized):
		return connect.NewError(connect.CodeUnauthenticated, err)
	case errors.Is(err, ErrForbidden):
		return connect.NewError(connect.CodePermissionDenied, err)
	case errors.Is(err, ErrInvalidInput):
		return connect.NewError(connect.CodeInvalidArgument, err)
	case errors.Is(err, ErrAlreadyExists):
		return connect.NewError(connect.CodeAlreadyExists, err)
	case errors.Is(err, ErrUnavailable), errors.Is(err, ErrDeviceOffline):
		return connect.NewError(connect.CodeUnavailable, err)
	default:
		return connect.NewError(connect.CodeInternal, err)
	}
}
