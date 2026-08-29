package command

import "github.com/quixiq/polyglot/pkg/fault"

// Sentinel errors for the command domain: policy enforcement for device
// command execution.
var (
	// ErrApprovalRequired indicates the command must be approved before
	// execution.
	ErrApprovalRequired = fault.New(fault.KindFailedPrecondition, "command: approval required")
	// ErrDenied indicates the command policy rejected execution.
	ErrDenied = fault.New(fault.KindPermissionDenied, "command: denied by policy")
	// ErrDriverNotStreaming indicates the device driver does not support
	// streaming output.
	ErrDriverNotStreaming = fault.New(fault.KindFailedPrecondition, "command: device driver does not support streaming")
)
