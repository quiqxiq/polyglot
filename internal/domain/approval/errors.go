package approval

import "github.com/quixiq/polyglot/pkg/fault"

var (
	ErrNotFound                = fault.New(fault.KindNotFound, "approval: not found")
	ErrInvalidTransition       = fault.New(fault.KindFailedPrecondition, "approval: invalid status transition")
	ErrExpired                 = fault.New(fault.KindFailedPrecondition, "approval: expired")
	ErrRevisionConflict        = fault.New(fault.KindFailedPrecondition, "approval: revision conflict")
	ErrRestorePointUnavailable = fault.New(fault.KindFailedPrecondition, "approval: restore point unavailable")
	ErrSelfApprove             = fault.New(fault.KindPermissionDenied, "approval: AI/agent cannot approve its own request")
)
