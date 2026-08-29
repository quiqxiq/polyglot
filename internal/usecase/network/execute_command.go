package network

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/port"
)

// Error sentinels (command.ErrApprovalRequired, command.ErrDenied) live in
// internal/domain/command/errors.go per DEVELOPMENT-GUIDELINES.md §6.

// ExecuteCommand runs cmd against a device via driver, after checking policy.
// usecase/ never inspects vendor command syntax — it only asks driver to
// Classify, then applies the generic domain/command policy decision. This
// is the single execution pipeline every other usecase in this package
// (GetDeviceStatus, PushConfig) reuses, so the gate is applied exactly once.
// TODO: wire actual HITL approval flow (pause/resume) per
// Polyglot-Architecture.md §6.1 (sequence: Policy -> pending_approval ->
// LibreChat HITL prompt -> approve/reject -> resume Execute).
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

func ExecuteCommand(ctx context.Context, driver port.DeviceDriver, cmd command.Command) (command.Result, error) {
	if err := validateIfSupported(ctx, driver, cmd); err != nil {
		return command.Result{}, err
	}
	switch command.Decide(driver.Classify(cmd)) {
	case command.DecisionAutoApprove:
		// Continue with execution for explicitly safe commands.
	case command.DecisionDeny:
		return command.Result{}, command.ErrDenied
	case command.DecisionRequireApproval:
		return command.Result{}, command.ErrApprovalRequired
	}
	return driver.Execute(ctx, cmd)
}

// ExecuteCommandPreApproved runs cmd against a device via driver, skipping
// the DecisionRequireApproval gate — the caller (e.g. the MCP adapter)
// asserts that the client (LibreChat) already obtained human approval via
// ToolAnnotations.DestructiveHint before calling the tool. This is the
// idiomatic MCP flow: HITL is client-side, the server executes when called.
//
// DecisionDeny still blocks — no amount of client-side approval overrides a
// hard policy deny. This preserves defense-in-depth: even if a buggy or
// malicious client calls a denied tool, the server refuses.
func ExecuteCommandPreApproved(ctx context.Context, driver port.DeviceDriver, cmd command.Command) (command.Result, error) {
	if command.Decide(driver.Classify(cmd)) == command.DecisionDeny {
		return command.Result{}, command.ErrDenied
	}
	return driver.Execute(ctx, cmd)
}
