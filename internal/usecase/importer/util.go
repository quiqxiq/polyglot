package importer

import (
	"fmt"

	domainAudit "github.com/quixiq/polyglot/internal/domain/audit"
)

func fmtLine(n int) string          { return fmt.Sprintf("baris %d:", n) }
func errf(f string, a ...any) error { return fmt.Errorf(f, a...) }

func domainAuditEntry(action, entityID string) domainAudit.AuditLog {
	return domainAudit.AuditLog{
		TenantID:   "tenant-default",
		ActorType:  domainAudit.ActorSystem,
		Action:     action,
		EntityType: "import",
		EntityID:   entityID,
	}
}
