package customer

import (
	"context"
	"fmt"
	"sort"

	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	domainCustomer "github.com/quixiq/polyglot/internal/domain/customer"
	domainRegistration "github.com/quixiq/polyglot/internal/domain/registration"
	domainSubscription "github.com/quixiq/polyglot/internal/domain/subscription"
	"github.com/quixiq/polyglot/internal/port"
)

// Customer360 merangkum seluruh data pelanggan dalam satu tampilan
// (planning v2 Fase 5): profil, langganan, tagihan, pembayaran, riwayat PSB.
type Customer360 struct {
	Customer      domainCustomer.Customer
	Subscriptions []domainSubscription.Subscription
	Invoices      []domainBilling.Invoice
	Payments      []domainBilling.Payment
	Registrations []domainRegistration.Registration
}

// With360 menautkan sumber data tambahan untuk Customer 360.
func (uc *ManageCustomerUseCase) With360(payments port.PaymentReader, regs port.RegistrationRepository) *ManageCustomerUseCase {
	uc.payments = payments
	uc.regs = regs
	return uc
}

// GetCustomer360 mengumpulkan data pelanggan dari seluruh domain terkait.
// Daftar dibatasi 20 item terbaru agar respons tetap ringan.
func (uc *ManageCustomerUseCase) GetCustomer360(ctx context.Context, customerID string) (Customer360, error) {
	cust, err := uc.repo.FindByID(ctx, customerID)
	if err != nil {
		return Customer360{}, fmt.Errorf("customer %s: %w", customerID, domainCustomer.ErrCustomerNotFound)
	}
	out := Customer360{Customer: cust}
	if uc.subRepo != nil {
		if subs, serr := uc.subRepo.FindByCustomerID(ctx, customerID); serr == nil {
			out.Subscriptions = subs
		}
	}
	if uc.invRepo != nil {
		if invs, ierr := uc.invRepo.FindByCustomerID(ctx, customerID); ierr == nil {
			sort.Slice(invs, func(i, j int) bool { return invs[i].CreatedAt.After(invs[j].CreatedAt) })
			if len(invs) > 20 {
				invs = invs[:20]
			}
			out.Invoices = invs
		}
	}
	if uc.payments != nil {
		if pays, perr := uc.payments.ListByCustomer(ctx, customerID, 20); perr == nil {
			out.Payments = pays
		}
	}
	if uc.regs != nil {
		if regs, rerr := uc.regs.List(ctx, port.RegistrationFilter{Phone: cust.Phone, Limit: 20}); rerr == nil {
			out.Registrations = regs
		}
	}
	return out, nil
}

// aggregateStatus merangkum status layanan pelanggan untuk header 360.
func aggregateStatus(rows Customer360) string {
	for _, s := range rows.Subscriptions {
		if s.Status == domainSubscription.StatusIsolated {
			return domainCustomer.StatusIsolated
		}
	}
	return rows.Customer.Status
}

var _ = fmt.Sprintf
