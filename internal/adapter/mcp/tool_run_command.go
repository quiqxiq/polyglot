package mcp

import "context"

// RunCommandHandler implements the "run_command" MCP tool.
// Approval requirement depends on command policy (see internal/domain/command)
// — set ToolAnnotations per-invocation risk is not possible statically, so
// this tool's annotation stays conservative; the real gate is
// usecase/network.ExecuteCommand's Classify/Decide check.
func RunCommandHandler(ctx context.Context) error {
	return nil
}
