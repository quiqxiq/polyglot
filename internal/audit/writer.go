package audit

import "context"

// NewWriter returns a port.AuditWriter backed by Postgres
// (`command_audit_log` table per Polyglot-Architecture.md §7.2).
func NewWriter(ctx context.Context) error {
	return nil
}
