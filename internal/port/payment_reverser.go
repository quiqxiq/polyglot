package port

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/billing"
)

// PaymentReversal adalah input pembatalan (void) pembayaran.
type PaymentReversal struct {
	PaymentID string
	Reason    string
	ActorID   *uint // users.id yang melakukan reversal
}

// PaymentReverser membalik pembayaran secara atomik: tandai void, catat
// jurnal koreksi kas keluar, dan sesuaikan paid_amount/status invoice.
type PaymentReverser interface {
	ReversePayment(ctx context.Context, in PaymentReversal) (billing.Payment, error)
}
