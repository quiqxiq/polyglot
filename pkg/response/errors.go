package response

import (
	"errors"

	"connectrpc.com/connect"
	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/domain/knowledge"
	knowledgeuc "github.com/quixiq/polyglot/internal/usecase/knowledge"
)

// MapDomainError converts domain and usecase errors into standard ConnectRPC
// typed errors with proper status codes.  Every ConnectRPC handler in
// internal/adapter/connect/ MUST use this single function for error mapping
// (DEVELOPMENT-GUIDELINES.md §6).
func MapDomainError(err error) error {
	if err == nil {
		return nil
	}

	// Already a connect.Error — pass through.
	var connErr *connect.Error
	if errors.As(err, &connErr) {
		return connErr
	}

	// InvalidArgument (validation errors from usecase layer).
	switch {
	case errors.Is(err, knowledgeuc.ErrInvalidTitle),
		errors.Is(err, knowledgeuc.ErrEmptyContent):
		return connect.NewError(connect.CodeInvalidArgument, err)

	// FailedPrecondition (infrastructure not ready).
	case errors.Is(err, knowledgeuc.ErrEmbedNotConfigured):
		return connect.NewError(connect.CodeFailedPrecondition, err)

	// NotFound.
	case errors.Is(err, device.ErrNotFound),
		errors.Is(err, knowledge.ErrNotFound):
		return connect.NewError(connect.CodeNotFound, err)
	}

	return connect.NewError(connect.CodeInternal, err)
}
