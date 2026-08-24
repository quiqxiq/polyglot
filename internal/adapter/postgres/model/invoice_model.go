package model

import (
	"time"

	"github.com/quixiq/polyglot/internal/domain/billing"
)

// InvoiceModel is the GORM model for monthly billing invoices
// (DATABASE-SCHEMA-ISP.md §2.6). Nominal dipecah dari `amount` tunggal
// menjadi subtotal/discount/tax/total/paid_amount.
type InvoiceModel struct {
	ID string `gorm:"primaryKey"`
	// invoice_number VARCHAR(50) UNIQUE — "INV-202608-00042".
	TenantID       string  `gorm:"type:text;not null;default:tenant-default"`
	InvoiceNumber  string  `gorm:"type:varchar(50);unique;not null"`
	CustomerID     string  `gorm:"type:text;not null;index"`
	SubscriptionID *string `gorm:"column:subscription_id;type:text"`

	Period     string  `gorm:"type:varchar(10);not null;index"` // 'YYYY-MM'
	Subtotal   float64 `gorm:"type:numeric(15,2)"`
	Discount   float64 `gorm:"type:numeric(15,2)"`
	TaxAmount  float64 `gorm:"type:numeric(15,2)"`
	Total      float64 `gorm:"type:numeric(15,2)"`
	PaidAmount float64 `gorm:"type:numeric(15,2)"`

	DueDate time.Time `gorm:"type:date;not null;index"`
	PaidAt  *time.Time
	Status  string `gorm:"type:varchar(20);not null;default:UNPAID;index"`

	QRPayload         string `gorm:"type:varchar(255);unique;not null"`
	ManualPaymentCode string `gorm:"type:varchar(30);unique;not null"`

	Notes        string `gorm:"type:text"`
	CancelledAt  *time.Time
	CancelReason string `gorm:"type:text"`
	DeletedAt    *time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (InvoiceModel) TableName() string {
	return "invoices"
}

func (m *InvoiceModel) ToDomain() billing.Invoice {
	if m == nil {
		return billing.Invoice{}
	}
	return billing.Invoice{
		ID:                m.ID,
		TenantID:          m.TenantID,
		InvoiceNumber:     m.InvoiceNumber,
		CustomerID:        m.CustomerID,
		SubscriptionID:    m.SubscriptionID,
		Period:            m.Period,
		Subtotal:          m.Subtotal,
		Discount:          m.Discount,
		TaxAmount:         m.TaxAmount,
		Total:             m.Total,
		PaidAmount:        m.PaidAmount,
		DueDate:           m.DueDate,
		PaidAt:            m.PaidAt,
		Status:            m.Status,
		QRPayload:         m.QRPayload,
		ManualPaymentCode: m.ManualPaymentCode,
		Notes:             m.Notes,
		CancelledAt:       m.CancelledAt,
		CancelReason:      m.CancelReason,
		DeletedAt:         m.DeletedAt,
		CreatedAt:         m.CreatedAt,
		UpdatedAt:         m.UpdatedAt,
	}
}

func InvoiceModelFromDomain(inv billing.Invoice) *InvoiceModel {
	return &InvoiceModel{
		ID:                inv.ID,
		TenantID:          inv.TenantID,
		InvoiceNumber:     inv.InvoiceNumber,
		CustomerID:        inv.CustomerID,
		SubscriptionID:    inv.SubscriptionID,
		Period:            inv.Period,
		Subtotal:          inv.Subtotal,
		Discount:          inv.Discount,
		TaxAmount:         inv.TaxAmount,
		Total:             inv.Total,
		PaidAmount:        inv.PaidAmount,
		DueDate:           inv.DueDate,
		PaidAt:            inv.PaidAt,
		Status:            inv.Status,
		QRPayload:         inv.QRPayload,
		ManualPaymentCode: inv.ManualPaymentCode,
		Notes:             inv.Notes,
		CancelledAt:       inv.CancelledAt,
		CancelReason:      inv.CancelReason,
		DeletedAt:         inv.DeletedAt,
		CreatedAt:         inv.CreatedAt,
		UpdatedAt:         inv.UpdatedAt,
	}
}
