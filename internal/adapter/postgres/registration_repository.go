package postgres

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres/model"
	"github.com/quixiq/polyglot/internal/domain/registration"
	"github.com/quixiq/polyglot/internal/port"
)

type RegistrationRepository struct {
	db *gorm.DB
}

var _ port.RegistrationRepository = (*RegistrationRepository)(nil)

// NewRegistrationRepository returns a port.RegistrationRepository backed by GORM/Postgres.
func NewRegistrationRepository(db *gorm.DB) *RegistrationRepository {
	return &RegistrationRepository{db: db}
}

func (r *RegistrationRepository) Save(ctx context.Context, reg registration.Registration) error {
	m := model.RegistrationModelFromDomain(reg)
	if err := r.db.WithContext(ctx).Save(m).Error; err != nil {
		return err
	}
	reg.CreatedAt = m.CreatedAt
	reg.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *RegistrationRepository) FindByID(ctx context.Context, id string) (registration.Registration, error) {
	var m model.RegistrationModel
	err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return registration.Registration{}, ErrNotFound
	}
	return m.ToDomain(), err
}

func (r *RegistrationRepository) FindByNo(ctx context.Context, regNo string) (registration.Registration, error) {
	var m model.RegistrationModel
	err := r.db.WithContext(ctx).First(&m, "registration_no = ?", regNo).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return registration.Registration{}, ErrNotFound
	}
	return m.ToDomain(), err
}

func (r *RegistrationRepository) List(ctx context.Context, f port.RegistrationFilter) ([]registration.Registration, error) {
	q := r.db.WithContext(ctx).Model(&model.RegistrationModel{})
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	if f.Phone != "" {
		q = q.Where("phone = ?", f.Phone)
	}
	if f.TechID != nil {
		q = q.Where("assigned_technician_id = ?", *f.TechID)
	}
	if f.Scheduled {
		q = q.Where("scheduled_install_date IS NOT NULL")
	}
	if f.Limit > 0 {
		q = q.Limit(f.Limit)
	}
	var mList []model.RegistrationModel
	if err := q.Order("created_at desc").Find(&mList).Error; err != nil {
		return nil, err
	}
	out := make([]registration.Registration, len(mList))
	for i := range mList {
		out[i] = mList[i].ToDomain()
	}
	return out, nil
}

func (r *RegistrationRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	res := r.db.WithContext(ctx).Model(&model.RegistrationModel{}).
		Where("id = ?", id).Update("status", status)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *RegistrationRepository) CountPendingSince(ctx context.Context, since time.Time) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&model.RegistrationModel{}).
		Where("status = ? AND created_at >= ?", registration.StatusPending, since).
		Count(&n).Error
	return n, err
}
