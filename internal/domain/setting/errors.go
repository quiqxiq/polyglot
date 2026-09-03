package setting

import "github.com/quixiq/polyglot/pkg/fault"

var (
	// ErrNotFound indicates the requested setting was not found.
	ErrNotFound = fault.New(fault.KindNotFound, "setting: not found")
	// ErrInvalidInput indicates invalid setting input.
	ErrInvalidInput = fault.New(fault.KindInvalidInput, "setting: invalid input")
	// ErrRepositoryNotConfigured indicates the usecase was wired without a
	// setting repository (composition-root error).
	ErrRepositoryNotConfigured = fault.New(fault.KindFailedPrecondition, "setting: repository is not configured")
)
