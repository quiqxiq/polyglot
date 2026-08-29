package model

import (
	"encoding/json"
	"time"

	"github.com/quixiq/polyglot/internal/domain/billing"
)

// GatewayTransactionModel is the GORM model for online payment-gateway
// transactions incl. webhook callback bookkeeping
// (DATABASE-SCHEMA-ISP.md §2.7 — gateway_transactions).
type GatewayTransactionModel struct {
	ID       string `gorm:"primaryKey"`
	TenantID string `gorm:"type:text;not null;default:tenant-default"`
	Gateway  string `gorm:"type:varchar(30);not null;uniqueIndex:uq_gateway_ext"` // 'TRIPAY','MIDTRANS','XENDIT'
	// external_id VARCHAR(100) — order ID dari payment gateway;
	// unik per (gateway, external_id).
	ExternalID string  `gorm:"type:varchar(100);not null;uniqueIndex:uq_gateway_ext"`
	InvoiceID  *string `gorm:"type:text;index"`
	PaymentID  *string `gorm:"column:payment_id;type:text"`

	Amount    float64 `gorm:"type:numeric(15,2);not null"`
	FeeAmount float64 `gorm:"type:numeric(15,2)"`

	Status         string `gorm:"type:varchar(30);not null;default:PENDING;index"`
	PaymentChannel string `gorm:"type:varchar(50)"`
	PaymentURL     string `gorm:"type:text"`
	QRString       string `gorm:"type:text"`

	RawCallback   json.RawMessage `gorm:"type:jsonb"`
	CallbackCount int             `gorm:"not null;default:0"`

	PaidAt    *time.Time
	ExpiresAt *time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
}

// TableName returns the database table name for gateway transactions.
func (GatewayTransactionModel) TableName() string {
	return "gateway_transactions"
}

// ToDomain converts a gateway transaction database model to its domain representation.
func (m *GatewayTransactionModel) ToDomain() billing.GatewayTransaction {
	if m == nil {
		return billing.GatewayTransaction{}
	}
	return billing.GatewayTransaction{
		ID:             m.ID,
		TenantID:       m.TenantID,
		Gateway:        m.Gateway,
		ExternalID:     m.ExternalID,
		InvoiceID:      m.InvoiceID,
		PaymentID:      m.PaymentID,
		Amount:         m.Amount,
		FeeAmount:      m.FeeAmount,
		Status:         m.Status,
		PaymentChannel: m.PaymentChannel,
		PaymentURL:     m.PaymentURL,
		QRString:       m.QRString,
		RawCallback:    m.RawCallback,
		CallbackCount:  m.CallbackCount,
		PaidAt:         m.PaidAt,
		ExpiresAt:      m.ExpiresAt,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
}

// GatewayTransactionModelFromDomain converts a gateway transaction domain entity to a database model.
func GatewayTransactionModelFromDomain(t billing.GatewayTransaction) *GatewayTransactionModel {
	return &GatewayTransactionModel{
		ID:             t.ID,
		TenantID:       t.TenantID,
		Gateway:        t.Gateway,
		ExternalID:     t.ExternalID,
		InvoiceID:      t.InvoiceID,
		PaymentID:      t.PaymentID,
		Amount:         t.Amount,
		FeeAmount:      t.FeeAmount,
		Status:         t.Status,
		PaymentChannel: t.PaymentChannel,
		PaymentURL:     t.PaymentURL,
		QRString:       t.QRString,
		RawCallback:    t.RawCallback,
		CallbackCount:  t.CallbackCount,
		PaidAt:         t.PaidAt,
		ExpiresAt:      t.ExpiresAt,
		CreatedAt:      t.CreatedAt,
		UpdatedAt:      t.UpdatedAt,
	}
}
