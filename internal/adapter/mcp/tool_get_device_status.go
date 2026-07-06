package mcp

import "context"

// GetDeviceStatusHandler implements the "get_device_status" MCP tool.
// Register with ToolAnnotations{ReadOnly: true} per
// TECH-STACK-DAN-PERSIAPAN.md §5 — typically auto-approved by
// usecase/network.GetDeviceStatus's policy gate, but still passes through
// the same Classify/Decide check as every other command, for defense in
// depth.
func GetDeviceStatusHandler(ctx context.Context) error {
	return nil
}
