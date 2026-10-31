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
	TaxPercent float64 `gorm:"column:tax_percent;type:numeric(5,2);not null;default:0"`
	TaxAmount  float64 `gorm:"type:numeric(15,2)"`
	Total      float64 `gorm:"type:numeric(15,2)"`
	PaidAmount float64 `gorm:"type:numeric(15,2)"`

	DueDate time.Time `gorm:"type:date;not null;index"`
	PaidAt  *time.Time
	Status  string `gorm:"type:varchar(20);not null;default:UNPAID;index"`

	QRPayload         string `gorm:"type:varchar(255);unique;not null"`
	ManualPaymentCode string `gorm:"type:varchar(30);unique;not null"`

	// Kolom billing v2 (migrasi 000032).
	InvoiceType    string     `gorm:"column:invoice_type;type:varchar(30);not null;default:RECURRING"`
	PeriodStart    *time.Time `gorm:"column:period_start;type:date"`
	PeriodEnd      *time.Time `gorm:"column:period_end;type:date"`
	ProratedDays   int        `gorm:"column:prorated_days;not null;default:0"`
	PeriodDays     int        `gorm:"column:period_days;not null;default:0"`
	DiscountType   string     `gorm:"column:discount_type;type:varchar(20);not null;default:NONE"`
	DiscountValue  float64    `gorm:"column:discount_value;type:numeric(15,2);not null;default:0"`
	DiscountAmount float64    `gorm:"column:discount_amount;type:numeric(15,2);not null;default:0"`
	PromoCode      string     `gorm:"column:promo_code;type:varchar(50)"`
	CreditNoteOf   *string    `gorm:"column:credit_note_of;type:text"`
	IssuedAt       time.Time  `gorm:"column:issued_at"`

	Notes        string `gorm:"type:text"`
	CancelledAt  *time.Time
	CancelReason string `gorm:"type:text"`
	DeletedAt    *time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
}

// TableName returns the database table name for invoices.
func (InvoiceModel) TableName() string {
	return "invoices"
}

// ToDomain converts an invoice database model to its domain representation.
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
		TaxPercent:        m.TaxPercent,
		TaxAmount:         m.TaxAmount,
		Total:             m.Total,
		PaidAmount:        m.PaidAmount,
		DueDate:           m.DueDate,
		PaidAt:            m.PaidAt,
		Status:            m.Status,
		QRPayload:         m.QRPayload,
		ManualPaymentCode: m.ManualPaymentCode,
		InvoiceType:       m.InvoiceType,
		PeriodStart:       m.PeriodStart,
		PeriodEnd:         m.PeriodEnd,
		ProratedDays:      m.ProratedDays,
		PeriodDays:        m.PeriodDays,
		DiscountType:      m.DiscountType,
		DiscountValue:     m.DiscountValue,
		DiscountAmount:    m.DiscountAmount,
		PromoCode:         m.PromoCode,
		CreditNoteOf:      m.CreditNoteOf,
		IssuedAt:          m.IssuedAt,
		Notes:             m.Notes,
		CancelledAt:       m.CancelledAt,
		CancelReason:      m.CancelReason,
		DeletedAt:         m.DeletedAt,
		CreatedAt:         m.CreatedAt,
		UpdatedAt:         m.UpdatedAt,
	}
}

// InvoiceModelFromDomain converts an invoice domain entity to a database model.
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
		TaxPercent:        inv.TaxPercent,
		TaxAmount:         inv.TaxAmount,
		Total:             inv.Total,
		PaidAmount:        inv.PaidAmount,
		DueDate:           inv.DueDate,
		PaidAt:            inv.PaidAt,
		Status:            inv.Status,
		QRPayload:         inv.QRPayload,
		ManualPaymentCode: inv.ManualPaymentCode,
		InvoiceType:       inv.InvoiceType,
		PeriodStart:       inv.PeriodStart,
		PeriodEnd:         inv.PeriodEnd,
		ProratedDays:      inv.ProratedDays,
		PeriodDays:        inv.PeriodDays,
		DiscountType:      inv.DiscountType,
		DiscountValue:     inv.DiscountValue,
		DiscountAmount:    inv.DiscountAmount,
		PromoCode:         inv.PromoCode,
		CreditNoteOf:      inv.CreditNoteOf,
		IssuedAt:          inv.IssuedAt,
		Notes:             inv.Notes,
		CancelledAt:       inv.CancelledAt,
		CancelReason:      inv.CancelReason,
		DeletedAt:         inv.DeletedAt,
		CreatedAt:         inv.CreatedAt,
		UpdatedAt:         inv.UpdatedAt,
	}
}
