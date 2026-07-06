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

// ExecuteCommand runs cmd against a device via driver, after checking policy.
// usecase/ never inspects vendor command syntax — it only asks driver to
// Classify, then applies the generic domain/command policy decision. This
// is the single execution pipeline every other usecase in this package
// (GetDeviceStatus, PushConfig) reuses, so the gate is applied exactly once.
// TODO: wire actual HITL approval flow (pause/resume) per
// NetOps-Architecture.md §6.1 (sequence: Policy -> pending_approval ->
// LibreChat HITL prompt -> approve/reject -> resume Execute).
func ExecuteCommand(ctx context.Context, driver port.DeviceDriver, cmd command.Command) (command.Result, error) {
	switch command.Decide(driver.Classify(cmd)) {
	case command.DecisionDeny:
		return command.Result{}, ErrDenied
	case command.DecisionRequireApproval:
		return command.Result{}, ErrApprovalRequired
	}
	return driver.Execute(ctx, cmd)
}
