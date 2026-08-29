package billing

import (
	"context"
	"fmt"

	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	"github.com/quixiq/polyglot/internal/port"
)

// InvoiceUseCase manages invoice queries and lifecycle operations.
type InvoiceUseCase struct {
	repo port.InvoiceRepository
}

func NewInvoiceUseCase(repo port.InvoiceRepository) *InvoiceUseCase {
	return &InvoiceUseCase{repo: repo}
}

func (u *InvoiceUseCase) ListInvoices(ctx context.Context, customerID string) ([]domainBilling.Invoice, error) {
	if u.repo == nil {
		return nil, fmt.Errorf("invoice repository unavailable")
	}
	if customerID != "" {
		return u.repo.FindByCustomerID(ctx, customerID)
	}
	return u.repo.FindAll(ctx)
}

func (u *InvoiceUseCase) GetInvoice(ctx context.Context, id string) (domainBilling.Invoice, error) {
	if u.repo == nil {
		return domainBilling.Invoice{}, fmt.Errorf("invoice repository unavailable")
	}
	return u.repo.FindByID(ctx, id)
}

func (u *InvoiceUseCase) CreateInvoice(ctx context.Context, inv domainBilling.Invoice) (domainBilling.Invoice, error) {
	if u.repo == nil {
		return domainBilling.Invoice{}, fmt.Errorf("invoice repository unavailable")
	}
	if inv.CustomerID == "" || inv.Total <= 0 {
		return domainBilling.Invoice{}, fmt.Errorf("customer_id and valid total amount are required")
	}
	if inv.Status == "" {
		inv.Status = domainBilling.StatusUnpaid
	}
	if err := u.repo.Save(ctx, inv); err != nil {
		return domainBilling.Invoice{}, err
	}
	return inv, nil
}

func (u *InvoiceUseCase) PayInvoice(ctx context.Context, id string) (domainBilling.Invoice, error) {
	if u.repo == nil {
		return domainBilling.Invoice{}, fmt.Errorf("invoice repository unavailable")
	}
	inv, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return domainBilling.Invoice{}, err
	}
	inv.Status = domainBilling.StatusPaid
	if err := u.repo.Save(ctx, inv); err != nil {
		return domainBilling.Invoice{}, err
	}
	return inv, nil
}
