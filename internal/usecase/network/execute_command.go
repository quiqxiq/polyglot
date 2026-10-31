package network

import (
	"context"
	"fmt"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/port"
)

// Error sentinels (command.ErrApprovalRequired, command.ErrDenied) live in
// internal/domain/command/errors.go per DEVELOPMENT-GUIDELINES.md §6.

// ExecuteCommand runs cmd against a device via driver, after checking policy.
// usecase/ never inspects vendor command syntax — it only asks driver to
// Classify, then applies the generic domain/command policy decision. This is
// the single READ execution pipeline; every MUTATION must instead go through
// the AI Router Ops proposal/approval pipeline (PLAN-AI-ROUTER-OPS v2).
func ExecuteCommand(ctx context.Context, driver port.DeviceDriver, cmd command.Command) (command.Result, error) {
	if driver == nil {
		return command.Result{}, command.ErrDriverUnavailable
	}
	if err := validateIfSupported(ctx, driver, cmd); err != nil {
		return command.Result{}, err
	}
	class := driver.Classify(cmd)
	switch command.Decide(class) {
	case command.DecisionAutoApprove:
		// Continue with execution for explicitly safe (READ) commands.
	case command.DecisionDeny:
		if class.DenyReason != "" {
			return command.Result{}, fmt.Errorf("%w: %s", command.ErrDenied, class.DenyReason)
		}
		return command.Result{}, command.ErrDenied
	case command.DecisionRequireApproval:
		return command.Result{}, command.ErrApprovalRequired
	}
	return driver.Execute(ctx, cmd)
}

// validateIfSupported pre-flights cmd via the driver's optional validation
// capability when it has one (see port.ValidatingDeviceDriver). Drivers
// without it are skipped — validation is strictly additive, never a
// required capability. Validation runs BEFORE policy: an invalid command
// (unknown path/attribute, syntax error) is a client bug to surface
// immediately, not a policy question to escalate for HITL.
func validateIfSupported(ctx context.Context, driver port.DeviceDriver, cmd command.Command) error {
	if v, ok := driver.(port.ValidatingDeviceDriver); ok {
		return v.Validate(ctx, cmd)
	}
	return nil
}
