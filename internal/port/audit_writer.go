package port

import "context"

// AuditWriter records command executions and business-critical changes.
type AuditWriter interface {
	Write(ctx context.Context) error
}
