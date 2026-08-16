package models

import (
	"time"

	"github.com/quixiq/polyglot/internal/domain/billing"
)

// InvoiceModel represents the GORM model for billing invoices.
type InvoiceModel struct {
	ID         string     `gorm:"primaryKey"`
	CustomerID string     `gorm:"not null;index"`
	Amount     float64    `gorm:"not null"`
	Status     string     `gorm:"not null;default:UNPAID"`
	DueDate    time.Time  `gorm:"not null"`
	PaidAt     *time.Time `gorm:"column:paid_at"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (InvoiceModel) TableName() string {
	return "invoices"
}

func (m *InvoiceModel) ToDomain() billing.Invoice {
	if m == nil {
		return billing.Invoice{}
	}
	return billing.Invoice{
		ID:         m.ID,
		CustomerID: m.CustomerID,
		Amount:     m.Amount,
		Status:     m.Status,
		DueDate:    m.DueDate,
		PaidAt:     m.PaidAt,
		CreatedAt:  m.CreatedAt,
		UpdatedAt:  m.UpdatedAt,
	}
}

func InvoiceModelFromDomain(inv billing.Invoice) *InvoiceModel {
	return &InvoiceModel{
		ID:         inv.ID,
		CustomerID: inv.CustomerID,
		Amount:     inv.Amount,
		Status:     inv.Status,
		DueDate:    inv.DueDate,
		PaidAt:     inv.PaidAt,
		CreatedAt:  inv.CreatedAt,
		UpdatedAt:  inv.UpdatedAt,
	}
}
