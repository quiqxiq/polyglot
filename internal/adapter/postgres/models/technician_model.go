package models

import (
	"time"

	"github.com/quixiq/polyglot/internal/domain/customer"
)

// TechnicianModel is the GORM database model for Technicians.
type TechnicianModel struct {
	ID             uint   `gorm:"primaryKey"`
	FullName       string `gorm:"not null"`
	Username       string `gorm:"uniqueIndex;not null"`
	PhoneNumber    string `gorm:"not null"`
	Specialization string
	IsActive       bool `gorm:"default:true"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (m *TechnicianModel) ToDomain() *customer.Technician {
	if m == nil {
		return nil
	}
	return &customer.Technician{
		ID:             m.ID,
		FullName:       m.FullName,
		Username:       m.Username,
		PhoneNumber:    m.PhoneNumber,
		Specialization: m.Specialization,
		IsActive:       m.IsActive,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
}

func TechnicianModelFromDomain(t *customer.Technician) *TechnicianModel {
	if t == nil {
		return nil
	}
	return &TechnicianModel{
		ID:             t.ID,
		FullName:       t.FullName,
		Username:       t.Username,
		PhoneNumber:    t.PhoneNumber,
		Specialization: t.Specialization,
		IsActive:       t.IsActive,
		CreatedAt:      t.CreatedAt,
		UpdatedAt:      t.UpdatedAt,
	}
}
