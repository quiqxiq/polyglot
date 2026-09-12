package mocktest

import (
	"context"
	"sort"
	"sync"
	"time"

	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	"github.com/quixiq/polyglot/internal/port"
)

// FakeInvoiceRepo is an in-memory invoice repository for tests.
type FakeInvoiceRepo struct {
	mu    sync.Mutex
	byID  map[string]domainBilling.Invoice
	items map[string][]domainBilling.InvoiceItem // invoice_id → items
}

// NewFakeInvoiceRepo constructs an empty fake invoice repository.
func NewFakeInvoiceRepo() *FakeInvoiceRepo {
	return &FakeInvoiceRepo{
		byID:  map[string]domainBilling.Invoice{},
		items: map[string][]domainBilling.InvoiceItem{},
	}
}

// Save stores an invoice in memory.
func (f *FakeInvoiceRepo) Save(_ context.Context, inv domainBilling.Invoice) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.byID[inv.ID] = inv
	return nil
}

// SaveWithItems stores an invoice and its line items.
func (f *FakeInvoiceRepo) SaveWithItems(_ context.Context, inv domainBilling.Invoice, items []domainBilling.InvoiceItem) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.byID[inv.ID] = inv
	f.items[inv.ID] = items
	return nil
}

// FindByID finds an invoice by ID.
func (f *FakeInvoiceRepo) FindByID(_ context.Context, id string) (domainBilling.Invoice, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	inv, ok := f.byID[id]
	if !ok {
		return domainBilling.Invoice{}, ErrFakeNotFound
	}
	return inv, nil
}

// FindByCustomerID lists invoices for a customer ID.
func (f *FakeInvoiceRepo) FindByCustomerID(_ context.Context, cid string) ([]domainBilling.Invoice, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []domainBilling.Invoice
	for _, inv := range f.byID {
		if inv.CustomerID == cid {
			out = append(out, inv)
		}
	}
	return out, nil
}

// FindAll returns all stored invoices.
func (f *FakeInvoiceRepo) FindAll(_ context.Context) ([]domainBilling.Invoice, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]domainBilling.Invoice, 0, len(f.byID))
	for _, inv := range f.byID {
		out = append(out, inv)
	}
	return out, nil
}

// FindPaged implements tenant + limit/offset filtering (F6-8).
func (f *FakeInvoiceRepo) FindPaged(_ context.Context, filter port.PageFilter) ([]domainBilling.Invoice, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]domainBilling.Invoice, 0, len(f.byID))
	for _, inv := range f.byID {
		if filter.TenantID != "" && inv.TenantID != filter.TenantID {
			continue
		}
		out = append(out, inv)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	if filter.Offset > 0 {
		if filter.Offset >= len(out) {
			return []domainBilling.Invoice{}, nil
		}
		out = out[filter.Offset:]
	}
	if filter.Limit > 0 && filter.Limit < len(out) {
		out = out[:filter.Limit]
	}
	return out, nil
}

// UpdateStatus updates the status of an invoice.
func (f *FakeInvoiceRepo) UpdateStatus(_ context.Context, id, status string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	inv, ok := f.byID[id]
	if !ok {
		return ErrFakeNotFound
	}
	inv.Status = status
	f.byID[id] = inv
	return nil
}

// FindByPaymentCode finds an invoice by manual payment code.
func (f *FakeInvoiceRepo) FindByPaymentCode(_ context.Context, code string) (domainBilling.Invoice, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, inv := range f.byID {
		if inv.ManualPaymentCode == code {
			return inv, nil
		}
	}
	return domainBilling.Invoice{}, ErrFakeNotFound
}

// FindByQRPayload finds an invoice by QR payload string.
func (f *FakeInvoiceRepo) FindByQRPayload(_ context.Context, qr string) (domainBilling.Invoice, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, inv := range f.byID {
		if inv.QRPayload == qr {
			return inv, nil
		}
	}
	return domainBilling.Invoice{}, ErrFakeNotFound
}

// FindBySubscriptionPeriod finds an invoice for a specific subscription and period.
func (f *FakeInvoiceRepo) FindBySubscriptionPeriod(_ context.Context, subID, period string) (domainBilling.Invoice, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, inv := range f.byID {
		if inv.SubscriptionID != nil && *inv.SubscriptionID == subID && inv.Period == period {
			return inv, nil
		}
	}
	return domainBilling.Invoice{}, ErrFakeNotFound
}

// ItemsOf exposes stored line items for assertions.
func (f *FakeInvoiceRepo) ItemsOf(invoiceID string) []domainBilling.InvoiceItem {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.items[invoiceID]
}

// HasForSubscription implements the invoice delete-guard lookup.
func (f *FakeInvoiceRepo) HasForSubscription(_ context.Context, subID string) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, inv := range f.byID {
		if inv.SubscriptionID != nil && *inv.SubscriptionID == subID {
			return true, nil
		}
	}
	return false, nil
}

// Delete removes an invoice and its items by ID.
func (f *FakeInvoiceRepo) Delete(_ context.Context, id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.byID, id)
	delete(f.items, id)
	return nil
}

// DeleteByCustomerID removes all invoices and items for a customer.
func (f *FakeInvoiceRepo) DeleteByCustomerID(_ context.Context, customerID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	for id, inv := range f.byID {
		if inv.CustomerID == customerID {
			delete(f.byID, id)
			delete(f.items, id)
		}
	}
	return nil
}

// ─── InvoiceCanceller ───────────────────────────────────────────────────

// FakeInvoiceCanceller membatalkan invoice di FakeInvoiceRepo dan merekam
// command (termasuk parameter jurnal koreksi kas).
type FakeInvoiceCanceller struct {
	Invoices *FakeInvoiceRepo
	Commands []port.CancelInvoiceCommand
	Fail     error
}

// NewFakeInvoiceCanceller constructs a canceller backed by the invoice fake.
func NewFakeInvoiceCanceller(invoices *FakeInvoiceRepo) *FakeInvoiceCanceller {
	return &FakeInvoiceCanceller{Invoices: invoices}
}

// Cancel records the command and performs a fake atomic cancellation.
func (f *FakeInvoiceCanceller) Cancel(ctx context.Context, cmd port.CancelInvoiceCommand) (domainBilling.Invoice, error) {
	f.Commands = append(f.Commands, cmd)
	if f.Fail != nil {
		return domainBilling.Invoice{}, f.Fail
	}
	inv, err := f.Invoices.FindByID(ctx, cmd.InvoiceID)
	if err != nil {
		return domainBilling.Invoice{}, err
	}
	if inv.Status == domainBilling.StatusPaid {
		return inv, domainBilling.ErrInvoiceAlreadyPaid
	}
	if inv.Status == domainBilling.StatusCancelled {
		return inv, nil
	}
	now := cmd.Now
	if now.IsZero() {
		now = time.Now()
	}
	inv.Status = domainBilling.StatusCancelled
	inv.CancelReason = cmd.Reason
	inv.CancelledAt = &now
	inv.UpdatedAt = now
	if err := f.Invoices.Save(ctx, inv); err != nil {
		return domainBilling.Invoice{}, err
	}
	return inv, nil
}

// ─── PaymentProcessor ───────────────────────────────────────────────────

// FakePaymentProcessor is a mock implementation of port.PaymentProcessor.
type FakePaymentProcessor struct {
	mu   sync.Mutex
	Cmds []port.CashPaymentCommand
	Err  error
	Pay  domainBilling.Payment
}

// ProcessCashPayment records the cash payment command and returns mock payment.
func (f *FakePaymentProcessor) ProcessCashPayment(_ context.Context, cmd port.CashPaymentCommand) (domainBilling.Payment, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Cmds = append(f.Cmds, cmd)
	if f.Err != nil {
		return domainBilling.Payment{}, f.Err
	}
	return f.Pay, nil
}
