package plan

import "github.com/quixiq/polyglot/pkg/fault"

// Sentinel errors for the plan domain.
var (
	ErrNotFound     = fault.New(fault.KindNotFound, "plan: not found")
	ErrInvalidInput = fault.New(fault.KindInvalidInput, "plan: validation failed")
	ErrPlanInUse    = fault.New(fault.KindFailedPrecondition, "plan: in use by active subscriptions")
)
