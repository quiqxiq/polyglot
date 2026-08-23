package port

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/audit"
)

// AuditLogWriter records business-critical activity to the audit_logs table
// (DATABASE-SCHEMA-ISP.md §2.10). Berbeda dari port.AuditWriter (audit
// eksekusi perintah device): ini audit alur bisnis ISP.
type AuditLogWriter interface {
	Write(ctx context.Context, entry audit.AuditLog) error
}
