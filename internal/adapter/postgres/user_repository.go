package postgres

import (
	"errors"

	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres/models"
	"github.com/quixiq/polyglot/internal/domain/customer"
)

func (s *Store) CreateUser(user *customer.User) error {
	m := models.UserModelFromDomain(user)
	if err := s.db.Create(m).Error; err != nil {
		return err
	}
	user.ID = m.ID
	return nil
}

func (s *Store) FindUserByID(id uint) (*customer.User, error) {
	var m models.UserModel
	if err := s.db.First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return m.ToDomain(), nil
}

func (s *Store) FindUserByEmail(email string) (*customer.User, error) {
	var m models.UserModel
	if err := s.db.Where("email = ?", email).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return m.ToDomain(), nil
}

func (s *Store) CountUsers() (int64, error) {
	var count int64
	err := s.db.Model(&models.UserModel{}).Count(&count).Error
	return count, err
}
