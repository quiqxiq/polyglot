package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres/model"
	"github.com/quixiq/polyglot/internal/domain/billing"
	"github.com/quixiq/polyglot/internal/port"
)

// PaymentSubmissionRepository menyimpan bukti transfer manual
// (tabel payment_submissions, planning v2 Fase 3).
type PaymentSubmissionRepository struct {
	db *gorm.DB
}

var _ port.PaymentSubmissionRepository = (*PaymentSubmissionRepository)(nil)

// NewPaymentSubmissionRepository returns a port.PaymentSubmissionRepository
// backed by GORM/Postgres.
func NewPaymentSubmissionRepository(db *gorm.DB) *PaymentSubmissionRepository {
	return &PaymentSubmissionRepository{db: db}
}

// Save membuat atau memperbarui satu submisi bukti transfer.
func (r *PaymentSubmissionRepository) Save(ctx context.Context, s billing.PaymentSubmission) error {
	return r.db.WithContext(ctx).Save(model.PaymentSubmissionModelFromDomain(s)).Error
}

// FindByID mengambil submisi berdasarkan id.
func (r *PaymentSubmissionRepository) FindByID(ctx context.Context, id string) (billing.PaymentSubmission, error) {
	var m model.PaymentSubmissionModel
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return billing.PaymentSubmission{}, ErrNotFound
		}
		return billing.PaymentSubmission{}, err
	}
	return m.ToDomain(), nil
}

// List mengembalikan submisi terbaru, opsional difilter status.
func (r *PaymentSubmissionRepository) List(ctx context.Context, f port.PaymentSubmissionFilter) ([]billing.PaymentSubmission, error) {
	q := r.db.WithContext(ctx).Model(&model.PaymentSubmissionModel{})
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	q = q.Order("created_at desc")
	if f.Limit > 0 {
		q = q.Limit(f.Limit)
	}
	var rows []model.PaymentSubmissionModel
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]billing.PaymentSubmission, len(rows))
	for i := range rows {
		out[i] = rows[i].ToDomain()
	}
	return out, nil
}
