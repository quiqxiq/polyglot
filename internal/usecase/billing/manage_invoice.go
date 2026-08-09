package billing

import (
	"context"
	"fmt"

	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	"github.com/quixiq/polyglot/internal/port"
)

type InvoiceUsecase struct {
	repo port.InvoiceRepository
}

func NewInvoiceUsecase(repo port.InvoiceRepository) *InvoiceUsecase {
	return &InvoiceUsecase{repo: repo}
}

func (u *InvoiceUsecase) ListInvoices(ctx context.Context, customerID string) ([]domainBilling.Invoice, error) {
	if u.repo == nil {
		return nil, fmt.Errorf("invoice repository unavailable")
	}
	if customerID != "" {
		return u.repo.FindByCustomerID(ctx, customerID)
	}
	return u.repo.FindAll(ctx)
}

func (u *InvoiceUsecase) GetInvoice(ctx context.Context, id string) (domainBilling.Invoice, error) {
	if u.repo == nil {
		return domainBilling.Invoice{}, fmt.Errorf("invoice repository unavailable")
	}
	return u.repo.FindByID(ctx, id)
}

func (u *InvoiceUsecase) CreateInvoice(ctx context.Context, inv domainBilling.Invoice) (domainBilling.Invoice, error) {
	if u.repo == nil {
		return domainBilling.Invoice{}, fmt.Errorf("invoice repository unavailable")
	}
	if inv.CustomerID == "" || inv.Amount <= 0 {
		return domainBilling.Invoice{}, fmt.Errorf("customer_id and valid amount are required")
	}
	if inv.Status == "" {
		inv.Status = domainBilling.StatusUnpaid
	}
	if err := u.repo.Save(ctx, inv); err != nil {
		return domainBilling.Invoice{}, err
	}
	return inv, nil
}

func (u *InvoiceUsecase) PayInvoice(ctx context.Context, id string) (domainBilling.Invoice, error) {
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
