package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres/model"
	"github.com/quixiq/polyglot/internal/domain/customer"
)

func (s *Store) CreateTechnician(ctx context.Context, tech *customer.Technician) error {
	m := model.TechnicianModelFromDomain(tech)
	if err := s.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	tech.ID = m.ID
	return nil
}

func (s *Store) FindAllTechnicians(ctx context.Context) ([]customer.Technician, error) {
	var mList []model.TechnicianModel
	if err := s.db.WithContext(ctx).Order("created_at DESC").Find(&mList).Error; err != nil {
		return nil, err
	}
	var res []customer.Technician
	for _, m := range mList {
		if d := m.ToDomain(); d != nil {
			res = append(res, *d)
		}
	}
	return res, nil
}

func (s *Store) FindActiveTechnicians(ctx context.Context) ([]customer.Technician, error) {
	var mList []model.TechnicianModel
	if err := s.db.WithContext(ctx).Where("is_active = ?", true).Order("full_name ASC").Find(&mList).Error; err != nil {
		return nil, err
	}
	var res []customer.Technician
	for _, m := range mList {
		if d := m.ToDomain(); d != nil {
			res = append(res, *d)
		}
	}
	return res, nil
}

func (s *Store) FindTechnicianByID(ctx context.Context, id uint) (*customer.Technician, error) {
	var m model.TechnicianModel
	if err := s.db.WithContext(ctx).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return m.ToDomain(), nil
}

func (s *Store) UpdateTechnician(ctx context.Context, tech *customer.Technician) error {
	m := model.TechnicianModelFromDomain(tech)
	return s.db.WithContext(ctx).Save(m).Error
}

func (s *Store) DeleteTechnician(ctx context.Context, id uint) error {
	return s.db.WithContext(ctx).Delete(&model.TechnicianModel{}, id).Error
}
