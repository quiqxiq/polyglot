package postgres

import (
	"context"
	"time"

	"gorm.io/gorm"

	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	"github.com/quixiq/polyglot/internal/port"
)

// PaymentReader implements port.PaymentReader via join payments→invoices.
type PaymentReader struct {
	db *gorm.DB
}

var _ port.PaymentReader = (*PaymentReader)(nil)

// NewPaymentReader returns a port.PaymentReader backed by GORM/Postgres.
func NewPaymentReader(db *gorm.DB) *PaymentReader {
	return &PaymentReader{db: db}
}

func (r *PaymentReader) ListByCustomer(ctx context.Context, customerID string, limit int) ([]domainBilling.Payment, error) {
	q := r.db.WithContext(ctx).Table("payments p").
		Select("p.id, p.tenant_id, p.payment_no, p.invoice_id, p.payment_method_id, "+
			"p.amount, p.payment_date, p.received_by, p.scan_method, p.reference, p.notes, p.created_at").
		Joins("JOIN invoices i ON i.id = p.invoice_id").
		Where("i.customer_id = ?", customerID).
		Order("p.payment_date desc")
	if limit > 0 {
		q = q.Limit(limit)
	}
	var models []struct {
		ID              string    `gorm:"column:id"`
		TenantID        string    `gorm:"column:tenant_id"`
		PaymentNo       string    `gorm:"column:payment_no"`
		InvoiceID       string    `gorm:"column:invoice_id"`
		PaymentMethodID *string   `gorm:"column:payment_method_id"`
		Amount          float64   `gorm:"column:amount"`
		PaymentDate     time.Time `gorm:"column:payment_date"`
		ReceivedBy      *uint     `gorm:"column:received_by"`
		ScanMethod      string    `gorm:"column:scan_method"`
		Reference       string    `gorm:"column:reference"`
		Notes           string    `gorm:"column:notes"`
		CreatedAt       time.Time `gorm:"column:created_at"`
	}
	if err := q.Scan(&models).Error; err != nil {
		return nil, err
	}
	out := make([]domainBilling.Payment, len(models))
	for i, m := range models {
		mm := m
		out[i] = domainBilling.Payment{
			ID: mm.ID, TenantID: mm.TenantID, PaymentNo: mm.PaymentNo,
			InvoiceID: mm.InvoiceID, PaymentMethodID: mm.PaymentMethodID,
			Amount: mm.Amount, PaymentDate: mm.PaymentDate, ReceivedBy: mm.ReceivedBy,
			ScanMethod: mm.ScanMethod, Reference: mm.Reference, Notes: mm.Notes,
			CreatedAt: mm.CreatedAt,
		}
	}
	return out, nil
}
