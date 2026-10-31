package model

import (
	"time"

	"github.com/quixiq/polyglot/internal/domain/billing"
)

// PaymentSubmissionModel is the GORM model for manual transfer proof
// submissions awaiting admin review
// (integrasi-router-billing-psb-global-smartapp.md §3.4 — payment_submissions).
type PaymentSubmissionModel struct {
	ID        string `gorm:"column:id;primaryKey"`
	TenantID  string `gorm:"column:tenant_id;type:text;not null;default:tenant-default"`
	InvoiceID string `gorm:"column:invoice_id;type:text;not null;index"`

	Amount        float64 `gorm:"column:amount;type:numeric(15,2);not null"`
	BankAccountID *string `gorm:"column:bank_account_id;type:text"`
	SenderBank    string  `gorm:"column:sender_bank;type:varchar(50)"`
	SenderName    string  `gorm:"column:sender_name;type:varchar(100)"`
	ProofImageURL string  `gorm:"column:proof_image_url;type:text;not null"`
	Notes         string  `gorm:"column:notes;type:text"`

	Status          string     `gorm:"column:status;type:varchar(20);not null;default:PENDING;index"`
	ReviewedBy      *uint      `gorm:"column:reviewed_by"`
	ReviewedAt      *time.Time `gorm:"column:reviewed_at"`
	RejectionReason string     `gorm:"column:rejection_reason;type:text"`

	PaymentID *string `gorm:"column:payment_id;type:text"`

	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

// TableName returns the database table name for payment submissions.
func (PaymentSubmissionModel) TableName() string {
	return "payment_submissions"
}

// ToDomain converts a payment submission database model to its domain representation.
func (m *PaymentSubmissionModel) ToDomain() billing.PaymentSubmission {
	if m == nil {
		return billing.PaymentSubmission{}
	}
	return billing.PaymentSubmission{
		ID:              m.ID,
		TenantID:        m.TenantID,
		InvoiceID:       m.InvoiceID,
		Amount:          m.Amount,
		BankAccountID:   m.BankAccountID,
		SenderBank:      m.SenderBank,
		SenderName:      m.SenderName,
		ProofImageURL:   m.ProofImageURL,
		Notes:           m.Notes,
		Status:          m.Status,
		ReviewedBy:      m.ReviewedBy,
		ReviewedAt:      m.ReviewedAt,
		RejectionReason: m.RejectionReason,
		PaymentID:       m.PaymentID,
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
	}
}

// PaymentSubmissionModelFromDomain converts a payment submission domain entity to a database model.
func PaymentSubmissionModelFromDomain(s billing.PaymentSubmission) *PaymentSubmissionModel {
	return &PaymentSubmissionModel{
		ID:              s.ID,
		TenantID:        s.TenantID,
		InvoiceID:       s.InvoiceID,
		Amount:          s.Amount,
		BankAccountID:   s.BankAccountID,
		SenderBank:      s.SenderBank,
		SenderName:      s.SenderName,
		ProofImageURL:   s.ProofImageURL,
		Notes:           s.Notes,
		Status:          s.Status,
		ReviewedBy:      s.ReviewedBy,
		ReviewedAt:      s.ReviewedAt,
		RejectionReason: s.RejectionReason,
		PaymentID:       s.PaymentID,
		CreatedAt:       s.CreatedAt,
		UpdatedAt:       s.UpdatedAt,
	}
}
