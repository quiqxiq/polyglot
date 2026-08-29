package port

import (
	"context"
	"time"

	"github.com/quixiq/polyglot/internal/domain/audit"
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

// AuditWriter records command executions and device-level operations.
// Implementations write to Postgres; the MCP adapter calls Write after
// every tool execution for a full audit trail.
type AuditWriter interface {
	Write(ctx context.Context, event AuditEvent) error
}

// AuditLogWriter records business-critical application activity to the audit_logs table
// (DATABASE-SCHEMA-ISP.md §2.10). Distinct from AuditWriter: AuditWriter captures low-level
// device commands, while AuditLogWriter captures high-level business workflow events.
type AuditLogWriter interface {
	Write(ctx context.Context, entry audit.AuditLog) error
}
