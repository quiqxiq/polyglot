package response

import (
	"errors"

	"connectrpc.com/connect"
	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/domain/skill"
	customerErr "github.com/quixiq/polyglot/internal/domain/customer"
	planErr "github.com/quixiq/polyglot/internal/domain/plan"
	regErr "github.com/quixiq/polyglot/internal/domain/registration"
	subErr "github.com/quixiq/polyglot/internal/domain/subscription"
	authuc "github.com/quixiq/polyglot/internal/usecase/auth"
	chatuc "github.com/quixiq/polyglot/internal/usecase/chat"
	convuc "github.com/quixiq/polyglot/internal/usecase/conversation"
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
	// (approval, streaming capability) belum terpenuhi.
	if errors.Is(err, networkuc.ErrApprovalRequired) ||
		errors.Is(err, networkuc.ErrDriverNotStreaming) {
		return connect.NewError(connect.CodeFailedPrecondition, err)
	}

	// Unauthenticated: kredensial salah / token tidak sah.
	if errors.Is(err, authuc.ErrInvalidCredentials) {
		return connect.NewError(connect.CodeUnauthenticated, err)
	}

	// AlreadyExists: entitas duplikat saat create.
	if errors.Is(err, useruc.ErrUserAlreadyExists) ||
		errors.Is(err, skill.ErrSkillAlreadyExists) {
		return connect.NewError(connect.CodeAlreadyExists, err)
	}

	// InvalidArgument (validation errors dari usecase layer).
	if errors.Is(err, authuc.ErrRefreshTokenRequired) ||
		errors.Is(err, chatuc.ErrEmptyChatJID) ||
		errors.Is(err, skill.ErrInvalidSlug) ||
		errors.Is(err, planErr.ErrNameRequired) ||
		errors.Is(err, planErr.ErrInvalidServiceType) ||
		errors.Is(err, planErr.ErrInvalidRate) ||
		errors.Is(err, planErr.ErrInvalidPrice) ||
		errors.Is(err, customerErr.ErrNameRequired) ||
		errors.Is(err, customerErr.ErrPhoneRequired) ||
		errors.Is(err, subErr.ErrCustomerRequired) ||
		errors.Is(err, subErr.ErrPlanRequired) ||
		errors.Is(err, subErr.ErrInvalidServiceType) ||
		errors.Is(err, subErr.ErrInvalidBillingDay) ||
		errors.Is(err, regErr.ErrPlanRequired) ||
		errors.Is(err, regErr.ErrNameRequired) ||
		errors.Is(err, regErr.ErrPhoneRequired) ||
		errors.Is(err, regErr.ErrAddressRequired) ||
		errors.Is(err, regErr.ErrDeviceRequired) {
		return connect.NewError(connect.CodeInvalidArgument, err)
	}

	// AlreadyExists: entitas duplikat saat create.
	if errors.Is(err, useruc.ErrUserAlreadyExists) ||
		errors.Is(err, skill.ErrSkillAlreadyExists) ||
		errors.Is(err, planErr.ErrAlreadyExists) ||
		errors.Is(err, customerErr.ErrAlreadyExists) ||
		errors.Is(err, regErr.ErrAlreadyPending) {
		return connect.NewError(connect.CodeAlreadyExists, err)
	}

	// FailedPrecondition: aksi belum bisa dijalankan karena prasyarat
	// (approval, streaming capability) belum terpenuhi.
	if errors.Is(err, networkuc.ErrApprovalRequired) ||
		errors.Is(err, networkuc.ErrDriverNotStreaming) ||
		errors.Is(err, regErr.ErrInvalidTransition) ||
		errors.Is(err, networkuc.ErrDeviceRequired) {
		return connect.NewError(connect.CodeFailedPrecondition, err)
	}

	// NotFound.
	if errors.Is(err, convuc.ErrNotFound) ||
		errors.Is(err, device.ErrNotFound) ||
		errors.Is(err, skill.ErrSkillNotFound) ||
		errors.Is(err, planErr.ErrNotFound) ||
		errors.Is(err, customerErr.ErrNotFound) ||
		errors.Is(err, subErr.ErrNotFound) ||
		errors.Is(err, regErr.ErrNotFound) {
		return connect.NewError(connect.CodeNotFound, err)
	}

	return connect.NewError(connect.CodeInternal, err)
}
