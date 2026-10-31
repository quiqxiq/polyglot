package restorepoint

import "github.com/quixiq/polyglot/pkg/fault"

var (
	ErrNotFound       = fault.New(fault.KindNotFound, "restorepoint: not found")
	ErrNotAvailable   = fault.New(fault.KindFailedPrecondition, "restorepoint: not available")
	ErrExpired        = fault.New(fault.KindFailedPrecondition, "restorepoint: expired")
	ErrCreationFailed = fault.New(fault.KindUnavailable, "restorepoint: creation failed")
	ErrSizeLimit      = fault.New(fault.KindResourceExhausted, "restorepoint: artifact exceeds size limit")
	ErrStorage        = fault.New(fault.KindUnavailable, "restorepoint: storage error")
)
