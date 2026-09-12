package billing

import (
	"context"
	"fmt"
	"time"

	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	"github.com/quixiq/polyglot/internal/port"
)

// InvoiceUseCase manages invoice queries and lifecycle operations.
type InvoiceUseCase struct {
	repo      port.InvoiceRepository
	canceller port.InvoiceCanceller
	settings  port.SettingReader
}

func NewInvoiceUseCase(repo port.InvoiceRepository) *InvoiceUseCase {
	return &InvoiceUseCase{repo: repo}
}

// WithCanceller menautkan pembatalan atomik + koreksi kas (F3-6).
func (u *InvoiceUseCase) WithCanceller(c port.InvoiceCanceller) *InvoiceUseCase {
	u.canceller = c
	return u
}

// WithSettings menautkan sumber konfigurasi akun/kategori koreksi kas.
func (u *InvoiceUseCase) WithSettings(s port.SettingReader) *InvoiceUseCase {
	u.settings = s
	return u
}

func (u *InvoiceUseCase) ListInvoices(ctx context.Context, customerID string) ([]domainBilling.Invoice, error) {
	if u.repo == nil {
		return nil, domainBilling.ErrRepositoryUnavailable
	}
	if customerID != "" {
		return u.repo.FindByCustomerID(ctx, customerID)
	}
	invoices, err := u.repo.FindPaged(ctx, port.PageFilter{TenantID: "tenant-default"})
	if err != nil {
		return nil, fmt.Errorf("list invoices: %w", err)
	}
	return invoices, nil
}

func (u *InvoiceUseCase) GetInvoice(ctx context.Context, id string) (domainBilling.Invoice, error) {
	if u.repo == nil {
		return domainBilling.Invoice{}, domainBilling.ErrRepositoryUnavailable
	}
	return u.repo.FindByID(ctx, id)
}

// CancelInvoice membatalkan faktur UNPAID/OVERDUE/PARTIAL secara resmi.
// Mengembalikan ErrInvoiceAlreadyPaid bila faktur sudah lunas.
func (u *InvoiceUseCase) CancelInvoice(ctx context.Context, id, reason string) (domainBilling.Invoice, error) {
	if u.repo == nil {
		return domainBilling.Invoice{}, domainBilling.ErrRepositoryUnavailable
	}
	if id == "" {
		return domainBilling.Invoice{}, domainBilling.ErrInvalidInput
	}
	if u.canceller != nil {
		accountID, categoryID := "ca-1001-kas-kantor", "cc-refund"
		if u.settings != nil {
			accountID = u.settings.GetValue(ctx, "isp.cancel_cash_account_id", accountID)
			categoryID = u.settings.GetValue(ctx, "isp.cancel_expense_category_id", categoryID)
		}
		inv, cerr := u.canceller.Cancel(ctx, port.CancelInvoiceCommand{
			InvoiceID: id, Reason: reason,
			CashAccountID: accountID, ExpenseCategoryID: categoryID,
		})
		if cerr != nil {
			return inv, fmt.Errorf("cancel invoice %s: %w", id, cerr)
		}
		return inv, nil
	}
	inv, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return domainBilling.Invoice{}, fmt.Errorf("find invoice %s: %w", id, err)
	}
	if inv.Status == domainBilling.StatusPaid {
		return inv, domainBilling.ErrInvoiceAlreadyPaid
	}
	if inv.Status == domainBilling.StatusCancelled {
		return inv, nil
	}
	now := time.Now()
	inv.Status = domainBilling.StatusCancelled
	inv.CancelledAt = &now
	inv.CancelReason = reason
	inv.UpdatedAt = now
	if err := u.repo.Save(ctx, inv); err != nil {
		return domainBilling.Invoice{}, fmt.Errorf("save invoice %s: %w", id, err)
	}
	return inv, nil
}
