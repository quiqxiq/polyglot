package model

import (
	"time"

	"github.com/quixiq/polyglot/internal/domain/customer"
)

// CustomerModel represents the GORM model for ISP subscribers
// (DATABASE-SCHEMA-ISP.md §2.4). Soft delete manual (*time.Time),
// bukan gorm.DeletedAt — keputusan desain fase model.
type CustomerModel struct {
	ID string `gorm:"primaryKey"`
	// customer_code VARCHAR(30) UNIQUE; unique penuh (bukan parsial) agar
	// AutoMigrate dev konvergen dengan migrasi prod.
	CustomerCode     string  `gorm:"type:varchar(30);unique;not null"`
	TenantID         string  `gorm:"type:text;not null;default:tenant-default"`
	Name             string  `gorm:"type:varchar(100);not null"`
	Phone            string  `gorm:"type:varchar(20);not null;index"`
	Email            string  `gorm:"type:varchar(100)"`
	Address          string  `gorm:"type:text;not null"`
	Latitude         *float64 `gorm:"type:double precision"`
	Longitude        *float64 `gorm:"type:double precision"`
	PortalAccessCode string  `gorm:"type:varchar(16);unique;not null"`
	Status           string  `gorm:"type:varchar(20);not null;default:ACTIVE;index"`
	Notes            string  `gorm:"type:text"`
	RegisteredAt     time.Time `gorm:"type:date;not null;default:CURRENT_DATE"`
	DeletedAt        *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func (CustomerModel) TableName() string {
	return "customers"
}

func (m *CustomerModel) ToDomain() customer.Customer {
	if m == nil {
		return customer.Customer{}
	}
	return customer.Customer{
		ID:               m.ID,
		TenantID:         m.TenantID,
		CustomerCode:     m.CustomerCode,
		Name:             m.Name,
		Phone:            m.Phone,
		Email:            m.Email,
		Address:          m.Address,
		Latitude:         m.Latitude,
		Longitude:        m.Longitude,
		PortalAccessCode: m.PortalAccessCode,
		Status:           m.Status,
		Notes:            m.Notes,
		RegisteredAt:     m.RegisteredAt,
		DeletedAt:        m.DeletedAt,
		CreatedAt:        m.CreatedAt,
		UpdatedAt:        m.UpdatedAt,
	}
}

func CustomerModelFromDomain(c customer.Customer) *CustomerModel {
	return &CustomerModel{
		ID:               c.ID,
		TenantID:         c.TenantID,
		CustomerCode:     c.CustomerCode,
		Name:             c.Name,
		Phone:            c.Phone,
		Email:            c.Email,
		Address:          c.Address,
		Latitude:         c.Latitude,
		Longitude:        c.Longitude,
		PortalAccessCode: c.PortalAccessCode,
		Status:           c.Status,
		Notes:            c.Notes,
		RegisteredAt:     c.RegisteredAt,
		DeletedAt:        c.DeletedAt,
		CreatedAt:        c.CreatedAt,
		UpdatedAt:        c.UpdatedAt,
	}
}
