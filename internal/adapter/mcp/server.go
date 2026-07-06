package mcp

import "context"

// NewServer builds and returns the MCP server exposing network tools.
// Keep MCP surface intentionally narrow per NetOps-Architecture.md §2
// prinsip 3 — only operations that genuinely need AI judgment
// (get_device_status, run_command, push_config). Admin CRUD stays in
// internal/adapter/http, never exposed as an MCP tool.
// TODO: wire github.com/modelcontextprotocol/go-sdk/mcp (current STABLE
// v1.x release, NOT the @v1.7.0-pre.1 pre-release) per
// TECH-STACK-DAN-PERSIAPAN.md §5.
//
// ⚠ Re-review after 28 Juli 2026: the MCP spec has a major stateless
// revision landing then (see TECH-STACK-DAN-PERSIAPAN.md, bagian pembuka).
// Don't assume a persistent session-per-connection model is the only
// shape this adapter will ever need.
func NewServer(ctx context.Context) error {
	return nil
}
