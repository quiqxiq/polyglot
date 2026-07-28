package network

import (
	"context"
	"errors"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/port"
)

// ErrApprovalRequired indicates cmd was classified as destructive and needs
// human approval (HITL) before ExecuteCommand may proceed.
var ErrApprovalRequired = errors.New("execute_command: approval required")

// ErrDenied indicates cmd was denied by policy outright.
var ErrDenied = errors.New("execute_command: denied by policy")

// ErrUnknownDecision indicates command.Decide returned a Decision value this
// usecase does not recognize. It exists so the policy switch can fail closed
// (refuse to execute) instead of falling through to Execute — a new Decision
// added without updating this gate must never silently auto-execute.
var ErrUnknownDecision = errors.New("execute_command: unknown policy decision")

// ExecuteCommand runs cmd against a device via driver, after checking policy.
// usecase/ never inspects vendor command syntax — it only asks driver to
// Classify, then applies the generic domain/command policy decision. This
// is the single execution pipeline every other usecase in this package
// (GetDeviceStatus, PushConfig) reuses, so the gate is applied exactly once.
// TODO: wire actual HITL approval flow (pause/resume) per
// Polyglot-Architecture.md §6.1 (sequence: Policy -> pending_approval ->
// LibreChat HITL prompt -> approve/reject -> resume Execute).
func ExecuteCommand(ctx context.Context, driver port.DeviceDriver, cmd command.Command) (command.Result, error) {
	switch command.Decide(driver.Classify(cmd)) {
	case command.DecisionAutoApprove:
		return driver.Execute(ctx, cmd)
	case command.DecisionRequireApproval:
		return command.Result{}, ErrApprovalRequired
	case command.DecisionDeny:
		return command.Result{}, ErrDenied
	default:
		// Fail closed: an unrecognized decision must never fall through to
		// Execute. This is the single HITL gate, so a silently unhandled
		// case is a policy hole, not a no-op.
		return command.Result{}, ErrUnknownDecision
	}
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
		return command.Result{}, ErrDenied
	}
	return driver.Execute(ctx, cmd)
}
