package models

import (
	"time"

	"github.com/quixiq/polyglot/internal/domain/customer"
)

// TechnicianModel is the GORM database model for Technicians.
// TableName eksplisit ke `technicians` (migrasi 000002) — tanpa ini GORM
// memakai `technician_models`, divergen dari migrasi (prod).
type TechnicianModel struct {
	ID       uint   `gorm:"primaryKey"`
	FullName string `gorm:"not null"`
	// Tipe eksplisit = migrasi 000002; `unique` (bukan `uniqueIndex`) karena
	// migrasi memakai UNIQUE constraint per-kolom — dengan `uniqueIndex` GORM
	// membaca columnType.Unique()=true tapi field.Unique=false → AutoMigrate
	// crash saat DROP CONSTRAINT `uni_technicians_username` yang tidak ada.
	Username       string `gorm:"type:varchar(100);unique;not null"`
	PhoneNumber    string `gorm:"type:varchar(50);not null"`
	Specialization string `gorm:"type:varchar(255)"`
	IsActive       bool   `gorm:"default:true"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// TableName maps TechnicianModel ke tabel migrasi `technicians`.
func (TechnicianModel) TableName() string { return "technicians" }

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
