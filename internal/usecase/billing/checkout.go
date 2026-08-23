package billing

import (
	"context"
	"fmt"
	"sort"

	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	"github.com/quixiq/polyglot/internal/port"
)

// CheckoutUseCase implements the cashier quick-pay flow:
// resolve invoice (kode bayar / QR / kode portal) lalu bayar tunai secara
// atomik via port.PaymentProcessor (DATABASE-SCHEMA-ISP.md §4.2).
type CheckoutUseCase struct {
	invoices  port.InvoiceRepository
	customers port.CustomerRepository
	processor port.PaymentProcessor
}

// NewCheckoutUseCase wires dependencies.
func NewCheckoutUseCase(inv port.InvoiceRepository, cust port.CustomerRepository, proc port.PaymentProcessor) *CheckoutUseCase {
	return &CheckoutUseCase{invoices: inv, customers: cust, processor: proc}
}

// ResolveByPaymentCode finds an invoice by its short manual payment code.
func (u *CheckoutUseCase) ResolveByPaymentCode(ctx context.Context, code string) (domainBilling.Invoice, error) {
	if code == "" {
		return domainBilling.Invoice{}, fmt.Errorf("%w: payment code is required", ErrValidation)
	}
	return u.invoices.FindByPaymentCode(ctx, code)
}

// ResolveByQR finds an invoice by its QR payload.
func (u *CheckoutUseCase) ResolveByQR(ctx context.Context, qr string) (domainBilling.Invoice, error) {
	if qr == "" {
		return domainBilling.Invoice{}, fmt.Errorf("%w: qr payload is required", ErrValidation)
	}
	return u.invoices.FindByQRPayload(ctx, qr)
}

// ResolveByPortalCode finds the oldest UNPAID invoice of the customer that
// owns the given portal access code.
func (u *CheckoutUseCase) ResolveByPortalCode(ctx context.Context, portalCode string) (domainBilling.Invoice, error) {
	if portalCode == "" {
		return domainBilling.Invoice{}, fmt.Errorf("%w: portal access code is required", ErrValidation)
	}
	cust, err := u.customers.FindByPortalAccessCode(ctx, portalCode)
	if err != nil {
		return domainBilling.Invoice{}, fmt.Errorf("customer by portal code: %w", err)
	}
	invoices, err := u.invoices.FindByCustomerID(ctx, cust.ID)
	if err != nil {
		return domainBilling.Invoice{}, err
	}
	var unpaid []domainBilling.Invoice
	for _, inv := range invoices {
		if inv.Status == domainBilling.StatusUnpaid || inv.Status == domainBilling.StatusOverdue ||
			inv.Status == domainBilling.StatusUnpaid && inv.PaidAmount < inv.Total {
			unpaid = append(unpaid, inv)
		}
	}
	if len(unpaid) == 0 {
		return domainBilling.Invoice{}, fmt.Errorf("no unpaid invoice for customer %s", cust.CustomerCode)
	}
	sort.Slice(unpaid, func(i, j int) bool { return unpaid[i].DueDate.Before(unpaid[j].DueDate) })
	return unpaid[0], nil
}

// PayCash executes the atomic cashier payment.
func (u *CheckoutUseCase) PayCash(ctx context.Context, cmd port.CashPaymentCommand) (domainBilling.Payment, error) {
	return u.processor.ProcessCashPayment(ctx, cmd)
}
