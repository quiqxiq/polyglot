package billing

import "errors"

// Kesalahan umum usecase billing.
var (
	ErrValidation = errors.New("validation failed")
)

// ErrNotFoundBilling dipakai usecase billing untuk entitas tak ditemukan.
var ErrNotFoundBilling = errors.New("not found")

// ErrInvalidTransitionBilling menandai transisi status langganan ilegal.
var ErrInvalidTransitionBilling = errors.New("invalid subscription transition")
