package model

import (
	"time"

	"github.com/quixiq/polyglot/internal/domain/billing"
)

// BulkPaymentBatchModel is the GORM model for bulk_payment_batches table.
type BulkPaymentBatchModel struct {
	ID              string  `gorm:"primaryKey"`
	TenantID        string  `gorm:"type:text;not null;default:tenant-default"`
	CreatedBy       *uint   `gorm:"column:created_by"`
	PaymentMethodID *string `gorm:"column:payment_method_id;type:text"`
	InvoiceCount    int     `gorm:"not null;default:0"`
	Status          string  `gorm:"type:varchar(20);not null;default:QUEUED"`
	JobID           string  `gorm:"column:job_id;type:text"`
	Notes           string  `gorm:"type:text"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// TableName returns the table name for bulk_payment_batches.
func (BulkPaymentBatchModel) TableName() string {
	return "bulk_payment_batches"
}

// ToDomain converts BulkPaymentBatchModel to domain billing.BulkPaymentBatch.
func (m *BulkPaymentBatchModel) ToDomain() billing.BulkPaymentBatch {
	if m == nil {
		return billing.BulkPaymentBatch{}
	}
	return billing.BulkPaymentBatch{
		ID:              m.ID,
		TenantID:        m.TenantID,
		CreatedBy:       m.CreatedBy,
		PaymentMethodID: m.PaymentMethodID,
		InvoiceCount:    m.InvoiceCount,
		Status:          m.Status,
		JobID:           m.JobID,
		Notes:           m.Notes,
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
	}
}

// BulkPaymentBatchModelFromDomain converts domain BulkPaymentBatch to BulkPaymentBatchModel.
func BulkPaymentBatchModelFromDomain(b billing.BulkPaymentBatch) *BulkPaymentBatchModel {
	return &BulkPaymentBatchModel{
		ID:              b.ID,
		TenantID:        b.TenantID,
		CreatedBy:       b.CreatedBy,
		PaymentMethodID: b.PaymentMethodID,
		InvoiceCount:    b.InvoiceCount,
		Status:          b.Status,
		JobID:           b.JobID,
		Notes:           b.Notes,
		CreatedAt:       b.CreatedAt,
		UpdatedAt:       b.UpdatedAt,
	}
}

// BulkPaymentItemModel is the GORM model for bulk_payment_items table.
type BulkPaymentItemModel struct {
	ID          string  `gorm:"primaryKey"`
	BatchID     string  `gorm:"type:text;not null;index"`
	InvoiceID   string  `gorm:"type:text;not null;index"`
	Amount      float64 `gorm:"type:numeric(15,2);not null"`
	Status      string  `gorm:"type:varchar(20);not null;default:QUEUED;index"`
	Error       string  `gorm:"type:text"`
	PaymentID   *string `gorm:"column:payment_id;type:text"`
	ProcessedAt *time.Time
	CreatedAt   time.Time
}

// TableName returns the table name for bulk_payment_items.
func (BulkPaymentItemModel) TableName() string {
	return "bulk_payment_items"
}

// ToDomain converts BulkPaymentItemModel to domain billing.BulkPaymentItem.
func (m *BulkPaymentItemModel) ToDomain() billing.BulkPaymentItem {
	if m == nil {
		return billing.BulkPaymentItem{}
	}
	return billing.BulkPaymentItem{
		ID:          m.ID,
		BatchID:     m.BatchID,
		InvoiceID:   m.InvoiceID,
		Amount:      m.Amount,
		Status:      m.Status,
		Error:       m.Error,
		PaymentID:   m.PaymentID,
		ProcessedAt: m.ProcessedAt,
		CreatedAt:   m.CreatedAt,
	}
}

// BulkPaymentItemModelFromDomain converts domain BulkPaymentItem to BulkPaymentItemModel.
func BulkPaymentItemModelFromDomain(i billing.BulkPaymentItem) *BulkPaymentItemModel {
	return &BulkPaymentItemModel{
		ID:          i.ID,
		BatchID:     i.BatchID,
		InvoiceID:   i.InvoiceID,
		Amount:      i.Amount,
		Status:      i.Status,
		Error:       i.Error,
		PaymentID:   i.PaymentID,
		ProcessedAt: i.ProcessedAt,
		CreatedAt:   i.CreatedAt,
	}
}
