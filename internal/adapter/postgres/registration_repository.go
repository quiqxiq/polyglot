package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres/model"
	"github.com/quixiq/polyglot/internal/domain/registration"
	"github.com/quixiq/polyglot/internal/port"
)

var _ port.RegistrationRepository = (*RegistrationRepository)(nil)

// RegistrationRepository persists new-subscriber application forms.
type RegistrationRepository struct {
	db *gorm.DB
}

func NewRegistrationRepository(db *gorm.DB) *RegistrationRepository {
	return &RegistrationRepository{db: db}
}

func (r *RegistrationRepository) Save(ctx context.Context, reg registration.Registration) error {
	m := model.RegistrationModelFromDomain(reg)
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *RegistrationRepository) FindByID(ctx context.Context, id string) (registration.Registration, error) {
	var m model.RegistrationModel
	err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return registration.Registration{}, registration.ErrNotFound
	}
	if err != nil {
		return registration.Registration{}, err
	}
	return m.ToDomain(), nil
}

func (r *RegistrationRepository) List(ctx context.Context, status string, limit int) ([]registration.Registration, error) {
	q := r.db.WithContext(ctx).Model(&model.RegistrationModel{})
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if limit > 0 {
		q = q.Limit(limit)
	}
	var mList []model.RegistrationModel
	if err := q.Order("created_at desc").Find(&mList).Error; err != nil {
		return nil, err
	}
	regs := make([]registration.Registration, len(mList))
	for i, m := range mList {
		regs[i] = m.ToDomain()
	}
	return regs, nil
}

func (r *RegistrationRepository) HasActiveByPhone(ctx context.Context, phone string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.RegistrationModel{}).
		Where("phone = ? AND status IN ('PENDING','APPROVED')", phone).
		Count(&count).Error
	return count > 0, err
}
