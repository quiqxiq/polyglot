package port

import (
	"context"
	"time"
)

// AuditEvent records a single command execution or business-critical change,
// persisted to the command_audit_log table per Polyglot-Architecture.md §7.2.
// The UserID field ties the action to a real human or AI agent identity for
// accountability — never empty for MCP-originated executions (the OBO token
// identifies the user behind the LLM).
type AuditEvent struct {
	DeviceID  string
	UserID    string
	Command   string
	Result    string
	Timestamp time.Time
}

// AuditWriter records command executions and business-critical changes.
// Implementations (internal/audit) write to Postgres; the MCP adapter calls
// Write after every tool execution for a full audit trail per
// Polyglot-Architecture.md §2 prinsip 8.
type AuditWriter interface {
	Write(ctx context.Context, event AuditEvent) error
}
