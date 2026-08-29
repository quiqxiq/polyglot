package model

import (
	"time"

	"github.com/quixiq/polyglot/internal/domain/billing"
)

// InvoiceItemModel is the GORM model for invoice line items
// (DATABASE-SCHEMA-ISP.md §2.6 — invoice_items).
type InvoiceItemModel struct {
	ID        string `gorm:"primaryKey"`
	InvoiceID string `gorm:"type:text;not null;index"`

	Description string  `gorm:"type:varchar(255);not null"` // "Paket 100-RB-100 (Agustus 2026)"
	Quantity    float64 `gorm:"not null;default:1.00"`
	UnitPrice   float64
	Amount      float64

	ItemType  string `gorm:"type:varchar(30);not null;default:SUBSCRIPTION_FEE"`
	CreatedAt time.Time
}

// TableName returns the database table name for invoice items.
func (InvoiceItemModel) TableName() string {
	return "invoice_items"
}

// ToDomain converts an invoice item database model to its domain representation.
func (m *InvoiceItemModel) ToDomain() billing.InvoiceItem {
	if m == nil {
		return billing.InvoiceItem{}
	}
	return billing.InvoiceItem{
		ID:          m.ID,
		InvoiceID:   m.InvoiceID,
		Description: m.Description,
		Quantity:    m.Quantity,
		UnitPrice:   m.UnitPrice,
		Amount:      m.Amount,
		ItemType:    m.ItemType,
		CreatedAt:   m.CreatedAt,
	}
}

// InvoiceItemModelFromDomain converts an invoice item domain entity to a database model.
func InvoiceItemModelFromDomain(i billing.InvoiceItem) *InvoiceItemModel {
	return &InvoiceItemModel{
		ID:          i.ID,
		InvoiceID:   i.InvoiceID,
		Description: i.Description,
		Quantity:    i.Quantity,
		UnitPrice:   i.UnitPrice,
		Amount:      i.Amount,
		ItemType:    i.ItemType,
		CreatedAt:   i.CreatedAt,
	}
}
