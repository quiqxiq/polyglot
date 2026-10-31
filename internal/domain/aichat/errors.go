package aichat

import "github.com/quixiq/polyglot/pkg/fault"

var (
	ErrNotFound       = fault.New(fault.KindNotFound, "aichat: conversation not found")
	ErrInvalidInput   = fault.New(fault.KindInvalidInput, "aichat: invalid input")
	ErrProviderFailed = fault.New(fault.KindUnavailable, "aichat: provider failed")
)
