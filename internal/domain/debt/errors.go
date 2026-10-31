package debt

import "github.com/quixiq/polyglot/pkg/fault"

// Sentinel errors domain hutang-piutang.
var (
	ErrNotFound     = fault.New(fault.KindNotFound, "debt: not found")
	ErrInvalidInput = fault.New(fault.KindInvalidInput, "debt: invalid input")
	ErrAlreadyPaid  = fault.New(fault.KindFailedPrecondition, "debt: already paid")
	ErrCancelled    = fault.New(fault.KindFailedPrecondition, "debt: cancelled")
	ErrOverpayment  = fault.New(fault.KindInvalidInput, "debt: payment exceeds remaining amount")
	ErrHasPayments  = fault.New(fault.KindFailedPrecondition, "debt: has payments, cancel not allowed")
)
