package billing

import (
	"context"
	"fmt"
	"strings"

	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	"github.com/quixiq/polyglot/internal/port"
)

// PaymentUseCase menangani operasi pembayaran pasca-transaksi: reversal
// (void) pembayaran yang salah dengan jurnal koreksi kas (planning v2 Fase 3).
type PaymentUseCase struct {
	reverser port.PaymentReverser
}

// NewPaymentUseCase wires dependencies.
func NewPaymentUseCase(reverser port.PaymentReverser) *PaymentUseCase {
	return &PaymentUseCase{reverser: reverser}
}

// Reverse membatalkan satu pembayaran.
func (u *PaymentUseCase) Reverse(ctx context.Context, paymentID, reason string, actorID *uint) (domainBilling.Payment, error) {
	if strings.TrimSpace(reason) == "" {
		return domainBilling.Payment{}, fmt.Errorf("%w: reason is required", domainBilling.ErrInvalidInput)
	}
	if u.reverser == nil {
		return domainBilling.Payment{}, domainBilling.ErrRepositoryUnavailable
	}
	return u.reverser.ReversePayment(ctx, port.PaymentReversal{
		PaymentID: paymentID,
		Reason:    reason,
		ActorID:   actorID,
	})
}
