package port

import (
	"context"
	"time"

	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
)

// CancelInvoiceCommand membawa parameter pembatalan invoice + koreksi kas.
// CashAccountID/ExpenseCategoryID dipakai untuk jurnal koreksi OUT bila
// invoice sudah memiliki pembayaran; kosong = lewati jurnal.
type CancelInvoiceCommand struct {
	InvoiceID         string
	Reason            string
	CashAccountID     string
	ExpenseCategoryID string
	RecordedBy        *uint
	Now               time.Time
}

// InvoiceCanceller membatalkan invoice secara ATOMIK: update status +
// jurnal kas koreksi (OUT) untuk pembayaran yang sudah diterima (F3-6).
type InvoiceCanceller interface {
	Cancel(ctx context.Context, cmd CancelInvoiceCommand) (domainBilling.Invoice, error)
}
