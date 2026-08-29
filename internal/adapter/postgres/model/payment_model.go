package model

import (
	"time"

	"github.com/quixiq/polyglot/internal/domain/billing"
)

// PaymentMethodModel is the GORM model for payment channel master
// (DATABASE-SCHEMA-ISP.md §2.7 — payment_methods).
type PaymentMethodModel struct {
	ID        string `gorm:"primaryKey"`
	TenantID  string `gorm:"type:text;not null;default:tenant-default"`
	Name      string `gorm:"type:varchar(50);not null"` // 'TUNAI', 'TRANSFER_BCA', 'QRIS', 'TRIPAY_VA'
	Type      string `gorm:"type:varchar(30);not null"` // CASH | BANK | QRIS | GATEWAY
	IsActive  bool   `gorm:"not null;default:true"`
	CreatedAt time.Time
}

// TableName returns the database table name for payment methods.
func (PaymentMethodModel) TableName() string {
	return "payment_methods"
}

// ToDomain converts a payment method database model to its domain representation.
func (m *PaymentMethodModel) ToDomain() billing.PaymentMethod {
	if m == nil {
		return billing.PaymentMethod{}
	}
	return billing.PaymentMethod{
		ID:        m.ID,
		TenantID:  m.TenantID,
		Name:      m.Name,
		Type:      m.Type,
		IsActive:  m.IsActive,
		CreatedAt: m.CreatedAt,
	}
}

// PaymentMethodModelFromDomain converts a payment method domain entity to a database model.
func PaymentMethodModelFromDomain(p billing.PaymentMethod) *PaymentMethodModel {
	return &PaymentMethodModel{
		ID:        p.ID,
		TenantID:  p.TenantID,
		Name:      p.Name,
		Type:      p.Type,
		IsActive:  p.IsActive,
		CreatedAt: p.CreatedAt,
	}
}

// PaymentModel is the GORM model for payment receipts settling invoices
// (DATABASE-SCHEMA-ISP.md §2.7 — payments).
// received_by BIGINT → users.id (uint, mengikuti UserModel eksisting).
type PaymentModel struct {
	ID              string  `gorm:"primaryKey"`
	TenantID        string  `gorm:"type:text;not null;default:tenant-default"`
	PaymentNo       string  `gorm:"type:varchar(50);unique;not null"`
	InvoiceID       string  `gorm:"type:text;not null;index"`
	PaymentMethodID *string `gorm:"column:payment_method_id;type:text"`

	Amount      float64   `gorm:"type:numeric(15,2);not null"`
	PaymentDate time.Time `gorm:"not null;index"`
	ReceivedBy  *uint     `gorm:"column:received_by"`
	ScanMethod  string    `gorm:"type:varchar(30);not null;default:MANUAL"`
	Reference   string    `gorm:"type:varchar(100)"`
	Notes       string    `gorm:"type:text"`
	CreatedAt   time.Time
}

// TableName returns the database table name for payments.
func (PaymentModel) TableName() string {
	return "payments"
}

// ToDomain converts a payment database model to its domain representation.
func (m *PaymentModel) ToDomain() billing.Payment {
	if m == nil {
		return billing.Payment{}
	}
	return billing.Payment{
		ID:              m.ID,
		TenantID:        m.TenantID,
		PaymentNo:       m.PaymentNo,
		InvoiceID:       m.InvoiceID,
		PaymentMethodID: m.PaymentMethodID,
		Amount:          m.Amount,
		PaymentDate:     m.PaymentDate,
		ReceivedBy:      m.ReceivedBy,
		ScanMethod:      m.ScanMethod,
		Reference:       m.Reference,
		Notes:           m.Notes,
		CreatedAt:       m.CreatedAt,
	}
}

// PaymentModelFromDomain converts a payment domain entity to a database model.
func PaymentModelFromDomain(p billing.Payment) *PaymentModel {
	return &PaymentModel{
		ID:              p.ID,
		TenantID:        p.TenantID,
		PaymentNo:       p.PaymentNo,
		InvoiceID:       p.InvoiceID,
		PaymentMethodID: p.PaymentMethodID,
		Amount:          p.Amount,
		PaymentDate:     p.PaymentDate,
		ReceivedBy:      p.ReceivedBy,
		ScanMethod:      p.ScanMethod,
		Reference:       p.Reference,
		Notes:           p.Notes,
		CreatedAt:       p.CreatedAt,
	}
}
