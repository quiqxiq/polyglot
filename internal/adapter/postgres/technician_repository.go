package postgres

import (
	"errors"

	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres/models"
	"github.com/quixiq/polyglot/internal/domain/customer"
)

func (s *Store) CreateTechnician(tech *customer.Technician) error {
	m := models.TechnicianModelFromDomain(tech)
	if err := s.db.Create(m).Error; err != nil {
		return err
	}
	tech.ID = m.ID
	return nil
}

func (s *Store) FindAllTechnicians() ([]customer.Technician, error) {
	var mList []models.TechnicianModel
	if err := s.db.Order("created_at DESC").Find(&mList).Error; err != nil {
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

func (s *Store) FindActiveTechnicians() ([]customer.Technician, error) {
	var mList []models.TechnicianModel
	if err := s.db.Where("is_active = ?", true).Order("full_name ASC").Find(&mList).Error; err != nil {
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

func (s *Store) FindTechnicianByID(id uint) (*customer.Technician, error) {
	var m models.TechnicianModel
	if err := s.db.First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return m.ToDomain(), nil
}

func (s *Store) UpdateTechnician(tech *customer.Technician) error {
	m := models.TechnicianModelFromDomain(tech)
	return s.db.Save(m).Error
}

func (s *Store) DeleteTechnician(id uint) error {
	return s.db.Delete(&models.TechnicianModel{}, id).Error
}
