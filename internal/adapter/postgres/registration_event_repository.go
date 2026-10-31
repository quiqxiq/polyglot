package postgres

import (
	"context"

	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres/model"
	"github.com/quixiq/polyglot/internal/domain/registration"
	"github.com/quixiq/polyglot/internal/port"
)

// RegistrationEventRepository menyimpan timeline PSB ke tabel
// registration_events (planning v2 §3.7).
type RegistrationEventRepository struct {
	db *gorm.DB
}

var _ port.RegistrationEventRepository = (*RegistrationEventRepository)(nil)

// NewRegistrationEventRepository returns a port.RegistrationEventRepository
// backed by GORM/Postgres.
func NewRegistrationEventRepository(db *gorm.DB) *RegistrationEventRepository {
	return &RegistrationEventRepository{db: db}
}

// Append menambahkan satu event ke timeline pendaftaran.
func (r *RegistrationEventRepository) Append(ctx context.Context, e registration.RegistrationEvent) error {
	m := model.RegistrationEventModelFromDomain(e)
	return r.db.WithContext(ctx).Create(m).Error
}

// ListByRegistration mengembalikan timeline pendaftaran, terurut lama→baru.
func (r *RegistrationEventRepository) ListByRegistration(ctx context.Context, registrationID string) ([]registration.RegistrationEvent, error) {
	var rows []model.RegistrationEventModel
	err := r.db.WithContext(ctx).
		Where("registration_id = ?", registrationID).
		Order("created_at asc").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]registration.RegistrationEvent, len(rows))
	for i := range rows {
		out[i] = rows[i].ToDomain()
	}
	return out, nil
}
