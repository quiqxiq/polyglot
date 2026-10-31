package olt

import "github.com/quixiq/polyglot/pkg/fault"

// Sentinel errors domain OLT/ONU.
var (
	// ErrNotFound indicates the OLT inventory record does not exist.
	ErrNotFound = fault.New(fault.KindNotFound, "olt: not found")
	// ErrUnsupportedVariant indicates an OLT variant this release cannot drive.
	ErrUnsupportedVariant = fault.New(fault.KindFailedPrecondition, "olt: unsupported variant")
	// ErrUnreachable indicates the OLT did not answer SNMP/CLI within budget.
	ErrUnreachable = fault.New(fault.KindUnavailable, "olt: device unreachable")
	// ErrCommandRejected indicates the OLT CLI answered with an error pattern.
	ErrCommandRejected = fault.New(fault.KindFailedPrecondition, "olt: command rejected by device")
	// ErrONUNotFound indicates the ONU position is not present on the OLT.
	ErrONUNotFound = fault.New(fault.KindNotFound, "olt: onu not found")
	// ErrInvalidInput indicates invalid OLT/ONU parameters.
	ErrInvalidInput = fault.New(fault.KindInvalidInput, "olt: invalid input")
)
