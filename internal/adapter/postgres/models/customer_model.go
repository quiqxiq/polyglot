package models

import (
	"time"

	"github.com/quixiq/polyglot/internal/domain/customer"
)

// CustomerModel is the GORM database model for customer subscribers.
type CustomerModel struct {
	ID        string `gorm:"primaryKey"`
	TenantID  string `gorm:"not null;index;default:tenant-default"`
	Name      string `gorm:"not null"`
	Email     string `gorm:"index"`
	Phone     string `gorm:"index"`
	Address   string `gorm:"type:text"`
	Status    string `gorm:"not null;default:active"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (m *CustomerModel) ToDomain() customer.Customer {
	if m == nil {
		return customer.Customer{}
	}
	return customer.Customer{
		ID:        m.ID,
		TenantID:  m.TenantID,
		Name:      m.Name,
		Email:     m.Email,
		Phone:     m.Phone,
		Address:   m.Address,
		Status:    m.Status,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func CustomerModelFromDomain(c customer.Customer) *CustomerModel {
	tenantID := c.TenantID
	if tenantID == "" {
		tenantID = "tenant-default"
	}
	return &CustomerModel{
		ID:        c.ID,
		TenantID:  tenantID,
		Name:      c.Name,
		Email:     c.Email,
		Phone:     c.Phone,
		Address:   c.Address,
		Status:    c.Status,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}
