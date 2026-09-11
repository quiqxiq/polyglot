package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres/model"
	"github.com/quixiq/polyglot/internal/domain/billing"
	"github.com/quixiq/polyglot/internal/domain/cashbook"
	"github.com/quixiq/polyglot/internal/port"
)

// InvoiceCanceller membatalkan invoice secara atomik: update status +
// jurnal kas koreksi OUT untuk pembayaran parsial yang sudah diterima (F3-6).
type InvoiceCanceller struct {
	db *gorm.DB
}

var _ port.InvoiceCanceller = (*InvoiceCanceller)(nil)

// NewInvoiceCanceller constructs an atomic invoice canceller.
func NewInvoiceCanceller(db *gorm.DB) *InvoiceCanceller {
	return &InvoiceCanceller{db: db}
}

// Cancel menjalankan pembatalan + koreksi kas dalam satu transaksi.
func (c *InvoiceCanceller) Cancel(ctx context.Context, cmd port.CancelInvoiceCommand) (billing.Invoice, error) {
	if cmd.InvoiceID == "" {
		return billing.Invoice{}, billing.ErrInvalidInput
	}
	var out billing.Invoice
	err := c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		q := tx
		if tx.Name() == "postgres" {
			q = q.Clauses(lockingClause())
		}
		var inv model.InvoiceModel
		if err := q.First(&inv, "id = ?", cmd.InvoiceID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotFound
			}
			return err
		}
		if inv.Status == billing.StatusPaid {
			return billing.ErrInvoiceAlreadyPaid
		}
		if inv.Status == billing.StatusCancelled {
			out = inv.ToDomain()
			return nil // idempoten
		}

		now := cmd.Now
		if now.IsZero() {
			now = time.Now()
		}
		updates := map[string]any{
			"status":        billing.StatusCancelled,
			"cancelled_at":  now,
			"cancel_reason": cmd.Reason,
			"updated_at":    now,
		}
		if err := tx.Model(&model.InvoiceModel{}).Where("id = ?", inv.ID).Updates(updates).Error; err != nil {
			return err
		}

		// Koreksi kas: uang yang sudah masuk dikembalikan (OUT) agar kas
		// tidak menyimpan pembayaran untuk tagihan yang dibatalkan.
		if inv.PaidAmount > 0 && cmd.CashAccountID != "" && cmd.ExpenseCategoryID != "" {
			reversal := &model.CashTransactionModel{
				ID:            newID("trx"),
				TenantID:      inv.TenantID,
				TransactionNo: fmt.Sprintf("TRX-%s-%06d", now.Format("200601"), now.UnixNano()%1000000),
				AccountID:     cmd.CashAccountID,
				CategoryID:    cmd.ExpenseCategoryID,
				Direction:     cashbook.DirectionOut,
				Amount:        inv.PaidAmount,
				TrxDate:       now,
				SourceType:    cashbook.SourceExpense,
				Description:   fmt.Sprintf("Koreksi pembatalan tagihan %s", inv.InvoiceNumber),
				RecordedBy:    cmd.RecordedBy,
			}
			if err := tx.Create(reversal).Error; err != nil {
				return err
			}
		}

		var updated model.InvoiceModel
		if err := tx.First(&updated, "id = ?", inv.ID).Error; err != nil {
			return err
		}
		out = updated.ToDomain()
		return nil
	})
	return out, err
}
