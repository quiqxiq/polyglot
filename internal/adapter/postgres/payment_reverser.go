package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres/model"
	"github.com/quixiq/polyglot/internal/domain/billing"
	domainCashbook "github.com/quixiq/polyglot/internal/domain/cashbook"
	"github.com/quixiq/polyglot/internal/port"
)

// PaymentReverser membalik pembayaran secara atomik (planning v2 Fase 3):
// tandai payment void, catat jurnal koreksi kas keluar (REFUND), dan
// sesuaikan paid_amount/status invoice. Invoice tidak pernah menjadi negatif.
type PaymentReverser struct {
	db *gorm.DB
}

var _ port.PaymentReverser = (*PaymentReverser)(nil)

// NewPaymentReverser returns a port.PaymentReverser backed by GORM/Postgres.
func NewPaymentReverser(db *gorm.DB) *PaymentReverser {
	return &PaymentReverser{db: db}
}

// ReversePayment mengeksekusi reversal satu pembayaran.
func (r *PaymentReverser) ReversePayment(ctx context.Context, in port.PaymentReversal) (billing.Payment, error) {
	if strings.TrimSpace(in.PaymentID) == "" || strings.TrimSpace(in.Reason) == "" {
		return billing.Payment{}, fmt.Errorf("%w: payment_id and reason are required", billing.ErrInvalidInput)
	}
	var out billing.Payment
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Kunci baris payment agar tidak ada double-reversal konkuren.
		var pay model.PaymentModel
		q := tx
		if r.db.Dialector.Name() == "postgres" {
			q = q.Clauses(lockingClause())
		}
		if err := q.First(&pay, "id = ?", in.PaymentID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotFound
			}
			return err
		}
		if pay.VoidedAt != nil {
			return billing.ErrPaymentAlreadyVoided
		}

		// 2. Kunci invoice target.
		var inv model.InvoiceModel
		qInv := tx
		if r.db.Dialector.Name() == "postgres" {
			qInv = qInv.Clauses(lockingClause())
		}
		if err := qInv.First(&inv, "id = ?", pay.InvoiceID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotFound
			}
			return err
		}

		now := time.Now()
		// 3. Jurnal koreksi kas keluar; rekening mengikuti transaksi asal.
		accountID := "ca-1001-kas-kantor"
		var origTx model.CashTransactionModel
		if err := tx.Where("source_type = ? AND source_id = ?", domainCashbook.SourcePayment, pay.ID).
			Order("created_at desc").First(&origTx).Error; err == nil && origTx.AccountID != "" {
			accountID = origTx.AccountID
		}
		correction := &model.CashTransactionModel{
			ID:            newID("trx"),
			TenantID:      pay.TenantID,
			TransactionNo: fmt.Sprintf("TRX-%s-%06d", now.Format("200601"), nextDocumentSeq(tx, "cash_transactions_no_seq")),
			AccountID:     accountID,
			CategoryID:    "cc-refund",
			Direction:     domainCashbook.DirectionOut,
			Amount:        pay.Amount,
			TrxDate:       now,
			SourceType:    domainCashbook.SourceRefund,
			SourceID:      pay.ID,
			Description:   fmt.Sprintf("Koreksi reversal pembayaran %s", pay.PaymentNo),
			RecordedBy:    in.ActorID,
		}
		if err := tx.Create(correction).Error; err != nil {
			return err
		}

		// 4. Tandai payment void (baris tetap ada untuk audit).
		if err := tx.Model(&model.PaymentModel{}).Where("id = ?", pay.ID).
			Updates(map[string]any{
				"voided_at":   now,
				"void_reason": in.Reason,
				"voided_by":   in.ActorID,
			}).Error; err != nil {
			return err
		}

		// 5. Sesuaikan invoice: paid_amount turun, status dihitung ulang.
		newPaid := inv.PaidAmount - pay.Amount
		if newPaid < 0 {
			newPaid = 0
		}
		status := inv.Status
		switch {
		case newPaid <= 0:
			status = billing.StatusUnpaid
		case newPaid >= inv.Total-1e-9:
			status = billing.StatusPaid
		default:
			status = billing.StatusPartial
		}
		updates := map[string]any{"paid_amount": newPaid, "status": status}
		if newPaid <= 0 {
			updates["paid_at"] = nil
		}
		if err := tx.Model(&model.InvoiceModel{}).Where("id = ?", inv.ID).Updates(updates).Error; err != nil {
			return err
		}

		// 5b. Rollback masa aktif langganan jika invoice tadinya PAID dan sekarang tidak lagi PAID
		if inv.Status == billing.StatusPaid && status != billing.StatusPaid {
			if err := rollbackSubscriptionOnReversal(tx, inv); err != nil {
				return err
			}
		}

		pay.VoidedAt = &now
		pay.VoidReason = in.Reason
		pay.VoidedBy = in.ActorID
		out = pay.ToDomain()
		return nil
	})
	if err != nil {
		return billing.Payment{}, err
	}
	return out, nil
}

func rollbackSubscriptionOnReversal(tx *gorm.DB, inv model.InvoiceModel) error {
	if inv.SubscriptionID == nil || *inv.SubscriptionID == "" {
		return nil
	}
	subID := *inv.SubscriptionID
	var latestPaid model.InvoiceModel
	err := tx.Where("subscription_id = ? AND status = ? AND id != ? AND deleted_at IS NULL", subID, billing.StatusPaid, inv.ID).
		Order("period_end DESC, period DESC").
		First(&latestPaid).Error

	if err == nil {
		end := latestPaid.PeriodEnd
		if end == nil {
			if parsed, parseErr := time.Parse("2006-01", latestPaid.Period); parseErr == nil {
				e := time.Date(parsed.Year(), parsed.Month()+1, 0, 0, 0, 0, 0, time.UTC)
				end = &e
			}
		}
		if end != nil {
			next := end.AddDate(0, 0, 1)
			return tx.Model(&model.SubscriptionModel{}).Where("id = ?", subID).
				Updates(map[string]any{
					"paid_through":  end,
					"end_date":      end,
					"next_due_date": next,
				}).Error
		}
	}

	// Tidak ada tagihan lunas lainnya yang tersisa untuk langganan ini
	return tx.Model(&model.SubscriptionModel{}).Where("id = ?", subID).
		Updates(map[string]any{
			"paid_through":  nil,
			"next_due_date": inv.PeriodStart,
		}).Error
}
