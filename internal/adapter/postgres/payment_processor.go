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
	"github.com/quixiq/polyglot/internal/domain/notification"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/logger"
)

// PaymentProcessor implements port.PaymentProcessor: seluruh efek pembayaran
// kasir dieksekusi dalam SATU transaksi database — gagal sebagian tidak
// mungkin (DATABASE-SCHEMA-ISP.md §4.2).
//
// OnPaid dipanggil SETELAH transaksi commit sukses (hook un-isolir otomatis).
// Kegagalan hook tidak membatalkan pembayaran — hanya dilog.
type PaymentProcessor struct {
	db    *gorm.DB
	OnPaid func(ctx context.Context, inv billing.Invoice, pay billing.Payment)
}

// NewPaymentProcessor returns a port.PaymentProcessor backed by GORM/Postgres.
func NewPaymentProcessor(db *gorm.DB) *PaymentProcessor {
	return &PaymentProcessor{db: db}
}

func (p *PaymentProcessor) ProcessCashPayment(ctx context.Context, cmd port.CashPaymentCommand) (billing.Payment, error) {
	if cmd.InvoiceID == "" || cmd.Amount <= 0 || cmd.CashAccountID == "" || cmd.IncomeCategoryID == "" {
		return billing.Payment{}, fmt.Errorf("%w: invoice_id, amount, cash_account_id and income_category_id are required", ErrInvalidArgument)
	}
	scanMethod := cmd.ScanMethod
	if scanMethod == "" {
		scanMethod = billing.ScanManual
	}

	var out billing.Payment
	var paidInv billing.Invoice
	err := p.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Kunci baris invoice untuk mencegah double-payment konkuren.
		// (SELECT ... FOR UPDATE hanya valid di Postgres; sqlite test skip.)
		q := tx
		if p.db.Dialector.Name() == "postgres" {
			q = q.Clauses(lockingClause())
		}
		var inv model.InvoiceModel
		if err := q.First(&inv, "id = ?", cmd.InvoiceID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotFound
			}
			return err
		}
		switch inv.Status {
		case billing.StatusPaid:
			return port.ErrInvoiceAlreadyPaid
		case billing.StatusCancelled:
			return port.ErrInvoiceCancelled
		}
		outstanding := inv.Total - inv.PaidAmount
		if cmd.Amount > outstanding+1e-9 {
			return fmt.Errorf("%w: outstanding %.2f, got %.2f", port.ErrOverpayment, outstanding, cmd.Amount)
		}

		now := time.Now()
		// 2. Update invoice.
		newPaid := inv.PaidAmount + cmd.Amount
		status := billing.StatusPartial
		if newPaid >= inv.Total-1e-9 {
			status = billing.StatusPaid
		}
		updates := map[string]any{"paid_amount": newPaid, "status": status}
		if status == billing.StatusPaid {
			updates["paid_at"] = now
		}
		if err := tx.Model(&model.InvoiceModel{}).Where("id = ?", inv.ID).Updates(updates).Error; err != nil {
			return err
		}

		// 3. Kwitansi pembayaran.
		payModel := &model.PaymentModel{
			ID:              newID("pay"),
			TenantID:        orDefault(cmd.TenantID, inv.TenantID),
			PaymentNo:       fmt.Sprintf("PAY-%s-%06d", now.Format("200601"), now.UnixNano()%1000000),
			InvoiceID:       inv.ID,
			PaymentMethodID: cmd.PaymentMethodID,
			Amount:          cmd.Amount,
			PaymentDate:     now,
			ReceivedBy:      cmd.ReceivedBy,
			ScanMethod:      scanMethod,
			Reference:       cmd.Reference,
			Notes:           cmd.Notes,
		}
		if err := tx.Create(payModel).Error; err != nil {
			return err
		}

		// 4. Jurnal kas masuk otomatis.
		cashModel := &model.CashTransactionModel{
			ID:            newID("trx"),
			TenantID:      payModel.TenantID,
			TransactionNo: fmt.Sprintf("TRX-%s-%06d", now.Format("200601"), now.UnixNano()%1000000),
			AccountID:     cmd.CashAccountID,
			CategoryID:    cmd.IncomeCategoryID,
			Direction:     cashbook.DirectionIn,
			Amount:        cmd.Amount,
			TrxDate:       now,
			SourceType:    cashbook.SourcePayment,
			SourceID:      payModel.ID,
			Description:   fmt.Sprintf("Pembayaran tagihan %s", inv.InvoiceNumber),
			RecordedBy:    cmd.ReceivedBy,
		}
		if err := tx.Create(cashModel).Error; err != nil {
			return err
		}

		// 5. Antrean WA bukti lunas — dikirim worker bot belakangan.
		waModel := &model.WANotificationModel{
			ID:             newID("wa"),
			TenantID:       payModel.TenantID,
			InvoiceID:      strPtr(inv.ID),
			CustomerID:     nil, // diisi ulang bila kolom customer diketahui
			RecipientPhone: "",  // worker mengisi dari customer saat kirim bila kosong
			MessageType:    "PAYMENT_RECEIPT",
			MessageContent: receiptContent(inv.InvoiceNumber, cmd.Amount, now),
			Status:         notification.StatusQueued,
		}
		if err := p.fillCustomerPhone(tx, inv.CustomerID, waModel); err != nil {
			return err
		}
		if err := tx.Create(waModel).Error; err != nil {
			return err
		}

		out = payModel.ToDomain()
		paidInv = inv.ToDomain()
		return nil
	})
	if err != nil {
		return billing.Payment{}, err
	}
	if p.OnPaid != nil {
		p.invokeOnPaid(ctx, paidInv, out)
	}
	return out, nil
}

// invokeOnPaid memanggil hook post-commit dengan proteksi panic.
func (p *PaymentProcessor) invokeOnPaid(ctx context.Context, inv billing.Invoice, pay billing.Payment) {
	defer func() {
		if r := recover(); r != nil {
			logger.WithComponent("PaymentProcessor").WithFields(map[string]any{
				"invoice_id": inv.ID, "panic": r,
			}).Error("OnPaid hook panicked")
		}
	}()
	p.OnPaid(ctx, inv, pay)
}

// fillCustomerPhone melengkapi nomor tujuan + customer_id pada antrean WA.
func (p *PaymentProcessor) fillCustomerPhone(tx *gorm.DB, customerID string, wa *model.WANotificationModel) error {
	var cust model.CustomerModel
	if err := tx.Select("id, phone").First(&cust, "id = ?", customerID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil // invoice tanpa customer valid tetap boleh tercatat
		}
		return err
	}
	wa.CustomerID = &cust.ID
	wa.RecipientPhone = cust.Phone
	return nil
}

func receiptContent(invoiceNo string, amount float64, paidAt time.Time) string {
	return fmt.Sprintf("Terima kasih, pembayaran Rp%.2f untuk tagihan %s telah kami terima pada %s.",
		amount, invoiceNo, paidAt.Format("02 Jan 2006 15:04"))
}
