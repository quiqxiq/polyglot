package port

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/audit"
)

// AuditEvent alias to domain model per DEVELOPMENT-GUIDELINES.md §4.2.
type AuditEvent = audit.Event

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
