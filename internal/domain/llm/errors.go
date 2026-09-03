package llm

import "github.com/quixiq/polyglot/pkg/fault"

var (
	// ErrNotFound indicates the requested LLM configuration was not found.
	ErrNotFound = fault.New(fault.KindNotFound, "llm: config not found")
	// ErrInvalidInput indicates the LLM configuration input is invalid.
	ErrInvalidInput = fault.New(fault.KindInvalidInput, "llm: invalid configuration")
	// ErrProviderUnavailable indicates the external LLM provider is currently unavailable.
	ErrProviderUnavailable = fault.New(fault.KindUnavailable, "llm: provider unavailable")
)
