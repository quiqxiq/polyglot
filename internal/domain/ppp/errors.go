package ppp

import "github.com/quixiq/polyglot/pkg/fault"

var (
	// ErrNotFound indicates the requested PPP entity was not found.
	ErrNotFound = fault.New(fault.KindNotFound, "ppp: resource not found")
	// ErrSecretNotFound indicates the PPPoE secret was not found.
	ErrSecretNotFound = fault.New(fault.KindNotFound, "ppp: secret not found")
	// ErrProfileNotFound indicates the PPP profile was not found.
	ErrProfileNotFound = fault.New(fault.KindNotFound, "ppp: profile not found")
	// ErrInvalidInput indicates PPP input parameters were invalid.
	ErrInvalidInput = fault.New(fault.KindInvalidInput, "ppp: invalid input")
)
