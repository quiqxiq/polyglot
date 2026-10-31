package postgres

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres/model"
	"github.com/quixiq/polyglot/internal/domain/audit"
	"github.com/quixiq/polyglot/internal/port"
)

// CommandLogRepository adalah implementasi port.CommandLogRepository berbasis
// GORM/Postgres untuk jejak perintah perangkat (PLAN-AI-ROUTER-OPS v2).
type CommandLogRepository struct {
	db *gorm.DB
}

var _ port.CommandLogRepository = (*CommandLogRepository)(nil)

// NewCommandLogRepository returns a port.CommandLogRepository backed by GORM.
func NewCommandLogRepository(db *gorm.DB) *CommandLogRepository {
	return &CommandLogRepository{db: db}
}

// Write menambahkan satu entri command log append-only. ID diisi otomatis
// oleh kolom IDENTITY.
func (r *CommandLogRepository) Write(ctx context.Context, entry audit.CommandLog) error {
	if entry.Command == "" || entry.Mode == "" || entry.Status == "" {
		return fmt.Errorf("%w: command, mode and status are required", audit.ErrInvalidInput)
	}
	m := model.CommandLogModelFromDomain(entry)
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now().UTC()
	}
	return r.db.WithContext(ctx).Create(m).Error
}

// List mengembalikan command log sesuai filter beserta total baris (sebelum
// limit/offset), terbaru lebih dulu.
func (r *CommandLogRepository) List(ctx context.Context, f port.CommandLogFilter) ([]audit.CommandLog, int, error) {
	q := r.db.WithContext(ctx).Model(&model.CommandLogModel{}).
		Where("tenant_id = ?", orDefault(f.TenantID, defaultTenantID))
	if f.DeviceID != "" {
		q = q.Where("device_id = ?", f.DeviceID)
	}
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	q = q.Order("created_at desc, id desc")
	if f.Limit > 0 {
		q = q.Limit(f.Limit)
	}
	if f.Offset > 0 {
		q = q.Offset(f.Offset)
	}
	var rows []model.CommandLogModel
	if err := q.Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	out := make([]audit.CommandLog, len(rows))
	for i := range rows {
		out[i] = rows[i].ToDomain()
	}
	return out, int(total), nil
}

// DeleteBefore removes command logs older than cutoff and returns the number
// of deleted rows.
func (r *CommandLogRepository) DeleteBefore(ctx context.Context, cutoff time.Time) (int, error) {
	res := r.db.WithContext(ctx).Where("created_at < ?", cutoff).Delete(&model.CommandLogModel{})
	if res.Error != nil {
		return 0, res.Error
	}
	return int(res.RowsAffected), nil
}
