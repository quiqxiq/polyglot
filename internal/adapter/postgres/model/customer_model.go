package model

import (
	"time"

	"github.com/quixiq/polyglot/internal/domain/customer"
)

// CustomerModel is the GORM model for ISP subscriber identities.
type CustomerModel struct {
	ID           string  `gorm:"primaryKey"`
	TenantID     string  `gorm:"not null;index;default:tenant-default"`
	CustomerCode string  `gorm:"column:customer_code"`
	Name         string  `gorm:"not null"`
	Email        string  `gorm:"index"`
	Phone        string  `gorm:"index"`
	Address      string  `gorm:"type:text"`
	Latitude     float64 `gorm:"column:latitude"`
	Longitude    float64 `gorm:"column:longitude"`
	Status       string  `gorm:"not null;default:ACTIVE"`
	Notes        string  `gorm:"type:text"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time `gorm:"index"`
}

func (CustomerModel) TableName() string {
	return "customers"
}

func (m *CustomerModel) ToDomain() customer.Customer {
	if m == nil {
		return customer.Customer{}
	}
	return customer.Customer{
		ID:           m.ID,
		TenantID:     m.TenantID,
		CustomerCode: m.CustomerCode,
		Name:         m.Name,
		Email:        m.Email,
		Phone:        m.Phone,
		Address:      m.Address,
		Latitude:     m.Latitude,
		Longitude:    m.Longitude,
		Status:       m.Status,
		Notes:        m.Notes,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
		DeletedAt:    m.DeletedAt,
	}
}

func CustomerModelFromDomain(c customer.Customer) *CustomerModel {
	tenantID := c.TenantID
	if tenantID == "" {
		tenantID = "tenant-default"
	}
	status := c.Status
	if status == "" {
		status = "ACTIVE"
	}
	return &CustomerModel{
		ID:           c.ID,
		TenantID:     tenantID,
		CustomerCode: c.CustomerCode,
		Name:         c.Name,
		Email:        c.Email,
		Phone:        c.Phone,
		Address:      c.Address,
		Latitude:     c.Latitude,
		Longitude:    c.Longitude,
		Status:       status,
		Notes:        c.Notes,
		CreatedAt:    c.CreatedAt,
		UpdatedAt:    c.UpdatedAt,
		DeletedAt:    c.DeletedAt,
	}
}
