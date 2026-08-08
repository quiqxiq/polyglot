package connectadapter

import (
	"errors"

	"connectrpc.com/connect"
	"github.com/quixiq/polyglot/internal/domain/device"
)

// MapDomainError converts internal domain and infrastructure errors into
// standard ConnectRPC typed errors with proper status codes.
func MapDomainError(err error) error {
	if err == nil {
		return nil
	}

	var connectErr *connect.Error
	if errors.As(err, &connectErr) {
		return err
	}

	switch {
	case errors.Is(err, device.ErrNotFound):
		return connect.NewError(connect.CodeNotFound, err)
	default:
		return connect.NewError(connect.CodeInternal, err)
	}
}
