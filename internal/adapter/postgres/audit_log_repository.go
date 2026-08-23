package postgres

import (
	"context"

	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres/model"
	"github.com/quixiq/polyglot/internal/domain/audit"
	"github.com/quixiq/polyglot/internal/port"
)

// AuditLogRepository implements port.AuditLogWriter: menulis entri audit
// alur bisnis ke tabel audit_logs (migrasi 000017).
type AuditLogRepository struct {
	db *gorm.DB
}

var _ port.AuditLogWriter = (*AuditLogRepository)(nil)

// NewAuditLogRepository returns a port.AuditLogWriter backed by GORM/Postgres.
func NewAuditLogRepository(db *gorm.DB) *AuditLogRepository {
	return &AuditLogRepository{db: db}
}

func (r *AuditLogRepository) Write(ctx context.Context, entry audit.AuditLog) error {
	m := model.AuditLogModelFromDomain(entry)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	entry.ID = m.ID
	entry.CreatedAt = m.CreatedAt
	return nil
}
