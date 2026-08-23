package postgres

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres/model"
	"github.com/quixiq/polyglot/internal/domain/reporting"
	"github.com/quixiq/polyglot/internal/port"
)

type ReportingRepository struct {
	db *gorm.DB
}

var _ port.ReportingRepository = (*ReportingRepository)(nil)

// NewReportingRepository returns a port.ReportingRepository backed by GORM/Postgres.
func NewReportingRepository(db *gorm.DB) *ReportingRepository {
	return &ReportingRepository{db: db}
}

func (r *ReportingRepository) UpsertSnapshot(ctx context.Context, s reporting.DailyFinancialSnapshot) error {
	m := model.DailySnapshotModelFromDomain(s)
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND snapshot_date = ?", s.TenantID, s.SnapshotDate).
		Assign(*m).
		FirstOrCreate(m).Error
	if err != nil {
		return err
	}
	s.ID = m.ID
	return nil
}

func (r *ReportingRepository) GetByDate(ctx context.Context, tenantID string, date time.Time) (reporting.DailyFinancialSnapshot, error) {
	var m model.DailySnapshotModel
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND snapshot_date = ?", tenantID, date).
		First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return reporting.DailyFinancialSnapshot{}, ErrNotFound
	}
	return m.ToDomain(), err
}

func (r *ReportingRepository) ListRange(ctx context.Context, tenantID string, from, to time.Time) ([]reporting.DailyFinancialSnapshot, error) {
	var mList []model.DailySnapshotModel
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND snapshot_date >= ? AND snapshot_date <= ?", tenantID, from, to).
		Order("snapshot_date asc").
		Find(&mList).Error
	if err != nil {
		return nil, err
	}
	out := make([]reporting.DailyFinancialSnapshot, len(mList))
	for i := range mList {
		out[i] = mList[i].ToDomain()
	}
	return out, nil
}
