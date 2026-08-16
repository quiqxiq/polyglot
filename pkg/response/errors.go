package response

import (
	"errors"

	"connectrpc.com/connect"
	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/domain/knowledge"
	authuc "github.com/quixiq/polyglot/internal/usecase/auth"
	chatuc "github.com/quixiq/polyglot/internal/usecase/chat"
	convuc "github.com/quixiq/polyglot/internal/usecase/conversation"
	knowledgeuc "github.com/quixiq/polyglot/internal/usecase/knowledge"
	networkuc "github.com/quixiq/polyglot/internal/usecase/network"
	useruc "github.com/quixiq/polyglot/internal/usecase/user"
)

// MapDomainError converts domain and usecase errors into standard ConnectRPC
// typed errors with proper status codes. Every ConnectRPC handler in
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

	// PermissionDenied: kebijakan usecase menolak eksekusi.
	if errors.Is(err, networkuc.ErrDenied) {
		return connect.NewError(connect.CodePermissionDenied, err)
	}

	// FailedPrecondition: aksi belum bisa dijalankan karena prasyarat
	// (approval, streaming capability, infra embed) belum terpenuhi.
	if errors.Is(err, networkuc.ErrApprovalRequired) ||
		errors.Is(err, networkuc.ErrDriverNotStreaming) ||
		errors.Is(err, knowledgeuc.ErrEmbedNotConfigured) {
		return connect.NewError(connect.CodeFailedPrecondition, err)
	}

	// Unauthenticated: kredensial salah / token tidak sah.
	if errors.Is(err, authuc.ErrInvalidCredentials) {
		return connect.NewError(connect.CodeUnauthenticated, err)
	}

	// AlreadyExists: entitas duplikat saat create.
	if errors.Is(err, useruc.ErrUserAlreadyExists) {
		return connect.NewError(connect.CodeAlreadyExists, err)
	}

	// InvalidArgument (validation errors dari usecase layer).
	if errors.Is(err, authuc.ErrRefreshTokenRequired) ||
		errors.Is(err, chatuc.ErrEmptyChatJID) ||
		errors.Is(err, knowledgeuc.ErrInvalidTitle) ||
		errors.Is(err, knowledgeuc.ErrEmptyContent) {
		return connect.NewError(connect.CodeInvalidArgument, err)
	}

	// NotFound.
	if errors.Is(err, convuc.ErrNotFound) ||
		errors.Is(err, device.ErrNotFound) ||
		errors.Is(err, knowledge.ErrNotFound) {
		return connect.NewError(connect.CodeNotFound, err)
	}

	return connect.NewError(connect.CodeInternal, err)
}
