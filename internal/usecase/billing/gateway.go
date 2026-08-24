package billing

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/idgen"
	"github.com/quixiq/polyglot/pkg/logger"
)

// GatewayChargeUseCase menangani pembuatan tagihan online dan pemrosesan
// webhook provider → pelunasan atomik (reuse PaymentProcessor sehingga
// un-isolir otomatis ikut terpicu lewat OnPaid).
type GatewayChargeUseCase struct {
	invoices  port.InvoiceRepository
	customs   port.CustomerRepository
	gwt       port.GatewayTransactionRepository
	gateway   port.PaymentGateway
	processor port.PaymentProcessor
	reader    port.SettingReader
}

func NewGatewayChargeUseCase(
	invoices port.InvoiceRepository,
	customs port.CustomerRepository,
	gwt port.GatewayTransactionRepository,
	gateway port.PaymentGateway,
	processor port.PaymentProcessor,
	reader port.SettingReader,
) *GatewayChargeUseCase {
	return &GatewayChargeUseCase{
		invoices: invoices, customs: customs, gwt: gwt,
		gateway: gateway, processor: processor, reader: reader,
	}
}

// CreateForInvoice membuat transaksi online untuk satu invoice UNPAID.
func (u *GatewayChargeUseCase) CreateForInvoice(ctx context.Context, invoiceID, channel string, expireMinutes int) (port.ChargeResult, domainBilling.GatewayTransaction, error) {
	if !u.gateway.Enabled(ctx) {
		return port.ChargeResult{}, domainBilling.GatewayTransaction{}, port.ErrGatewayDisabled
	}
	inv, err := u.invoices.FindByID(ctx, invoiceID)
	if err != nil {
		return port.ChargeResult{}, domainBilling.GatewayTransaction{}, ErrNotFoundBilling
	}
	switch inv.Status {
	case domainBilling.StatusPaid:
		return port.ChargeResult{}, domainBilling.GatewayTransaction{}, port.ErrInvoiceAlreadyPaid
	case domainBilling.StatusCancelled:
		return port.ChargeResult{}, domainBilling.GatewayTransaction{}, port.ErrInvoiceCancelled
	}
	outstanding := inv.Total - inv.PaidAmount
	if outstanding <= 0 {
		return port.ChargeResult{}, domainBilling.GatewayTransaction{}, port.ErrInvoiceAlreadyPaid
	}
	cust, err := u.customs.FindByID(ctx, inv.CustomerID)
	if err != nil {
		return port.ChargeResult{}, domainBilling.GatewayTransaction{}, fmt.Errorf("customer: %w", err)
	}

	res, cerr := u.gateway.CreateCharge(ctx, port.ChargeRequest{
		TenantID: inv.TenantID, InvoiceID: inv.ID, InvoiceNumber: inv.InvoiceNumber,
		Amount: outstanding, Channel: channel,
		CustomerName: cust.Name, CustomerPhone: cust.Phone, CustomerEmail: cust.Email,
		ExpireMinutes: expireMinutes,
	})
	if cerr != nil {
		return port.ChargeResult{}, domainBilling.GatewayTransaction{}, cerr
	}

	tx := domainBilling.GatewayTransaction{
		ID: idgen.New("gtx"), TenantID: inv.TenantID,
		Gateway: u.gateway.Name(), ExternalID: res.ExternalID,
		InvoiceID: &inv.ID, Amount: outstanding, FeeAmount: res.FeeAmount,
		Status:     domainBilling.GatewayStatusPending,
		PaymentURL: res.PaymentURL, QRString: firstNonEmpty(res.QRString, res.VANumber),
		RawCallback: res.RawResponse,
		CreatedAt:   time.Now(), UpdatedAt: time.Now(),
	}
	if err := u.gwt.Save(ctx, tx); err != nil {
		return port.ChargeResult{}, domainBilling.GatewayTransaction{}, err
	}
	return res, tx, nil
}

// HandleWebhook memvalidasi callback provider lalu menyelesaikan invoice
// bila status SETTLED. Mengembalikan invoiceID dan flag settled.
func (u *GatewayChargeUseCase) HandleWebhook(ctx context.Context, body []byte, signature string) (invoiceID string, settled bool, err error) {
	ev, err := u.gateway.ParseWebhook(ctx, body, signature)
	if err != nil {
		return "", false, err
	}
	tx, err := u.gwt.FindByExternalID(ctx, u.gateway.Name(), ev.ExternalID)
	if err != nil {
		return "", false, fmt.Errorf("%w: %s/%s", port.ErrGatewayUnknownRef, u.gateway.Name(), ev.ExternalID)
	}
	tx.RawCallback = json.RawMessage(ev.Raw)
	tx.UpdatedAt = time.Now()

	switch ev.Status {
	case domainBilling.GatewayStatusSettled:
		if tx.InvoiceID == nil {
			return "", false, fmt.Errorf("transaksi %s tanpa invoice", tx.ID)
		}
		pay, perr := u.processor.ProcessCashPayment(ctx, port.CashPaymentCommand{
			TenantID:         tx.TenantID,
			InvoiceID:        *tx.InvoiceID,
			Amount:           ev.PaidAmount,
			CashAccountID:    u.reader.GetValue(ctx, "gw.tripay.cash_account_id", "ca-1001-kas-kantor"),
			IncomeCategoryID: u.reader.GetValue(ctx, "gw.tripay.income_category_id", "cc-tagihan"),
			ScanMethod:       domainBilling.ScanPaymentGateway,
			Reference:        tx.ExternalID,
		})
		if perr != nil && !errors.Is(perr, port.ErrInvoiceAlreadyPaid) {
			return "", false, perr
		}
		if perr == nil {
			if lerr := u.gwt.LinkPayment(ctx, tx.ID, pay.ID, tx.FeeAmount); lerr != nil {
				logger.WithComponent("GatewayCharge").WithError(lerr).Warn("link payment gagal")
			}
		}
		tx.Status = domainBilling.GatewayStatusSettled
		tx.PaidAt = ptrTime(time.Now())
	default:
		tx.Status = ev.Status
	}
	if err := u.gwt.UpdateStatus(ctx, tx.ID, tx.Status); err != nil {
		return "", false, err
	}
	if tx.InvoiceID != nil && tx.Status == domainBilling.GatewayStatusSettled {
		return *tx.InvoiceID, true, nil
	}
	return "", false, nil
}

func ptrTime(t time.Time) *time.Time { return &t }

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
