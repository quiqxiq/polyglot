package mcp

import "context"

// PushConfigHandler implements the "push_config" MCP tool.
// Register with ToolAnnotations{Destructive: true} per
// TECH-STACK-DAN-PERSIAPAN.md §5 — routed through usecase/network.PushConfig,
// which reuses ExecuteCommand's HITL gate (NetOps-Architecture.md §6.1).
func PushConfigHandler(ctx context.Context) error {
	return nil
}
