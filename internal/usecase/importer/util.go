package importer

import (
	"fmt"

	domainAudit "github.com/quixiq/polyglot/internal/domain/audit"
	domainTenant "github.com/quixiq/polyglot/internal/domain/tenant"
)

func fmtLine(n int) string          { return fmt.Sprintf("baris %d:", n) }
func errf(f string, a ...any) error { return fmt.Errorf(f, a...) }

func domainAuditEntry(action, entityID string) domainAudit.AuditLog {
	return domainAudit.AuditLog{
		TenantID:   domainTenant.DefaultID,
		ActorType:  domainAudit.ActorSystem,
		Action:     action,
		EntityType: "import",
		EntityID:   entityID,
	}
}
