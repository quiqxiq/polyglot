package billing

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	domainCashbook "github.com/quixiq/polyglot/internal/domain/cashbook"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/idgen"
	"github.com/quixiq/polyglot/pkg/logger"
)

// GatewayChargeUseCase menangani pembuatan tagihan online dan pemrosesan
// webhook provider → pelunasan atomik (reuse PaymentProcessor sehingga
// un-isolir otomatis ikut terpicu lewat OnPaid). Menerima satu atau banyak
// gateway lewat GatewayRegistry (F4-1).
type GatewayChargeUseCase struct {
	invoices  port.InvoiceRepository
	customs   port.CustomerRepository
	gwt       port.GatewayTransactionRepository
	registry  *GatewayRegistry
	processor port.PaymentProcessor
	reader    port.SettingReader
	cashbook  port.CashbookRepository
}

// WithCashbook sets the cashbook repository for gateway fee expenses.
func (u *GatewayChargeUseCase) WithCashbook(cb port.CashbookRepository) *GatewayChargeUseCase {
	u.cashbook = cb
	return u
}

// NewGatewayChargeUseCase membungkus satu gateway (jalur kompatibel lama).
func NewGatewayChargeUseCase(
	invoices port.InvoiceRepository,
	customs port.CustomerRepository,
	gwt port.GatewayTransactionRepository,
	gateway port.PaymentGateway,
	processor port.PaymentProcessor,
	reader port.SettingReader,
) *GatewayChargeUseCase {
	return NewGatewayChargeUseCaseWithRegistry(invoices, customs, gwt,
		NewGatewayRegistry(reader, gateway), processor, reader)
}

// NewGatewayChargeUseCaseWithRegistry wires the multi-gateway registry.
func NewGatewayChargeUseCaseWithRegistry(
	invoices port.InvoiceRepository,
	customs port.CustomerRepository,
	gwt port.GatewayTransactionRepository,
	registry *GatewayRegistry,
	processor port.PaymentProcessor,
	reader port.SettingReader,
) *GatewayChargeUseCase {
	return &GatewayChargeUseCase{
		invoices: invoices, customs: customs, gwt: gwt,
		registry: registry, processor: processor, reader: reader,
	}
}

// Registry mengekspos registry gateway (untuk endpoint daftar gateway).
func (u *GatewayChargeUseCase) Registry() *GatewayRegistry { return u.registry }

// CreateForInvoice membuat transaksi online via gateway default.
func (u *GatewayChargeUseCase) CreateForInvoice(ctx context.Context, invoiceID, channel string, expireMinutes int) (port.ChargeResult, domainBilling.GatewayTransaction, error) {
	return u.CreateForInvoiceVia(ctx, invoiceID, "", channel, expireMinutes)
}

// CreateForInvoiceVia membuat transaksi online via gateway terpilih
// (kosong = default settings). Idempoten: transaksi PENDING untuk invoice +
// gateway yang sama dipakai ulang; PENDING gateway lain ditandai EXPIRED agar
// tidak ada dua tagihan aktif sekaligus (F4-6).
func (u *GatewayChargeUseCase) CreateForInvoiceVia(ctx context.Context, invoiceID, gatewayName, channel string, expireMinutes int) (port.ChargeResult, domainBilling.GatewayTransaction, error) {
	gw, err := u.registry.Get(ctx, gatewayName)
	if err != nil {
		return port.ChargeResult{}, domainBilling.GatewayTransaction{}, err
	}
	if !gw.Enabled(ctx) {
		return port.ChargeResult{}, domainBilling.GatewayTransaction{}, domainBilling.ErrGatewayDisabled
	}
	inv, err := u.invoices.FindByID(ctx, invoiceID)
	if err != nil {
		return port.ChargeResult{}, domainBilling.GatewayTransaction{}, domainBilling.ErrNotFound
	}
	switch inv.Status {
	case domainBilling.StatusPaid:
		return port.ChargeResult{}, domainBilling.GatewayTransaction{}, domainBilling.ErrInvoiceAlreadyPaid
	case domainBilling.StatusCancelled:
		return port.ChargeResult{}, domainBilling.GatewayTransaction{}, domainBilling.ErrInvoiceCancelled
	}
	outstanding := inv.Total - inv.PaidAmount
	if outstanding <= 0 {
		return port.ChargeResult{}, domainBilling.GatewayTransaction{}, domainBilling.ErrInvoiceAlreadyPaid
	}
	if tx, ok := u.reusePendingCharge(ctx, inv.ID, gw.Name()); ok {
		return chargeResultFromTx(tx), tx, nil
	}
	// Beban biaya gateway (planning v2 D11): MERCHANT ditanggung mitra
	// (tagihan tidak berubah), CUSTOMER/SPLIT ditambahkan ke invoice sekali
	// (idempoten) + markup admin mitra.
	if applied, ferr := u.applyGatewayFee(ctx, inv, outstanding); ferr == nil {
		inv = applied
		outstanding = inv.Total - inv.PaidAmount
	}
	cust, err := u.customs.FindByID(ctx, inv.CustomerID)
	if err != nil {
		return port.ChargeResult{}, domainBilling.GatewayTransaction{}, fmt.Errorf("customer: %w", err)
	}

	res, cerr := gw.CreateCharge(ctx, port.ChargeRequest{
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
		Gateway: gw.Name(), ExternalID: res.ExternalID,
		InvoiceID: &inv.ID, Amount: outstanding, FeeAmount: res.FeeAmount,
		Status:         domainBilling.GatewayStatusPending,
		PaymentChannel: res.Channel,
		PaymentURL:     res.PaymentURL, QRString: firstNonEmpty(res.QRString, res.VANumber),
		RawCallback: res.RawResponse,
		CreatedAt:   time.Now(), UpdatedAt: time.Now(),
	}
	if !res.ExpiresAt.IsZero() {
		expires := res.ExpiresAt
		tx.ExpiresAt = &expires
	}
	if err := u.gwt.Save(ctx, tx); err != nil {
		return port.ChargeResult{}, domainBilling.GatewayTransaction{}, err
	}
	return res, tx, nil
}

// applyGatewayFee menambahkan porsi biaya gateway yang dibebankan pelanggan
// (MERCHANT = 0, CUSTOMER = penuh, SPLIT = separuh) + markup ke invoice.
// Idempoten: item fee hanya ditambahkan sekali per invoice.
func (u *GatewayChargeUseCase) applyGatewayFee(ctx context.Context, inv domainBilling.Invoice, outstanding float64) (domainBilling.Invoice, error) {
	if u.reader == nil || outstanding <= 0 {
		return inv, nil
	}
	bearer := strings.TrimSpace(u.reader.GetValue(ctx, "isp.gateway_fee_bearer", domainBilling.GatewayFeeBearerMerchant))
	percent := atoiSafe(u.reader.GetValue(ctx, "isp.gateway_fee_percent", "0"))
	markup := atoiSafe(u.reader.GetValue(ctx, "isp.gateway_markup_amount", "0"))
	feeEstimate := outstanding * float64(percent) / 100
	customerFee := domainBilling.GatewayCustomerFee(feeEstimate, bearer) + float64(markup)
	if customerFee <= 0 {
		return inv, nil
	}
	updated, err := u.invoices.AddGatewayFeeItem(ctx, inv.ID, customerFee)
	if err != nil {
		logger.WithComponent("GatewayCharge").WithFields(map[string]any{
			"invoice_id": inv.ID, "fee": customerFee,
		}).WithError(err).Warn("gagal menambahkan biaya gateway ke invoice")
		return inv, err
	}
	return updated, nil
}

// reusePendingCharge mengembalikan transaksi PENDING invoice untuk gateway
// yang sama; transaksi PENDING gateway lain ditandai EXPIRED (F4-6).
func (u *GatewayChargeUseCase) reusePendingCharge(ctx context.Context, invoiceID, gatewayName string) (domainBilling.GatewayTransaction, bool) {
	txs, err := u.gwt.FindByInvoice(ctx, invoiceID)
	if err != nil {
		return domainBilling.GatewayTransaction{}, false
	}
	var reusable *domainBilling.GatewayTransaction
	for i := range txs {
		if txs[i].Status != domainBilling.GatewayStatusPending {
			continue
		}
		if strings.EqualFold(txs[i].Gateway, gatewayName) {
			reusable = &txs[i]
			continue
		}
		stale := txs[i]
		stale.Status = domainBilling.GatewayStatusExpired
		stale.UpdatedAt = time.Now()
		if serr := u.gwt.Save(ctx, stale); serr != nil {
			logger.WithComponent("GatewayCharge").WithError(serr).Warn("gagal expire transaksi pending lama")
		}
	}
	if reusable == nil {
		return domainBilling.GatewayTransaction{}, false
	}
	return *reusable, true
}

// HandleWebhook memproses callback via gateway default (kompatibel lama).
func (u *GatewayChargeUseCase) HandleWebhook(ctx context.Context, body []byte, signature string) (invoiceID string, settled bool, err error) {
	return u.HandleWebhookFor(ctx, "", body, signature)
}

// HandleWebhookFor memvalidasi callback provider terpilih lalu menyelesaikan
// invoice bila status SETTLED. Lookup transaksi memakai ExternalID lalu
// fallback MerchantRef.
func (u *GatewayChargeUseCase) HandleWebhookFor(ctx context.Context, gatewayName string, body []byte, signature string) (invoiceID string, settled bool, err error) {
	gw, err := u.registry.Get(ctx, gatewayName)
	if err != nil {
		return "", false, err
	}
	ev, err := gw.ParseWebhook(ctx, body, signature)
	if err != nil {
		return "", false, fmt.Errorf("parse webhook %s: %w", gw.Name(), err)
	}
	tx, err := u.gwt.FindByExternalID(ctx, gw.Name(), ev.ExternalID)
	if err != nil && ev.MerchantRef != "" && ev.MerchantRef != ev.ExternalID {
		tx, err = u.gwt.FindByExternalID(ctx, gw.Name(), ev.MerchantRef)
	}
	if err != nil {
		return "", false, fmt.Errorf("%w: %s/%s", domainBilling.ErrGatewayUnknownRef, gw.Name(), ev.ExternalID)
	}
	tx.RawCallback = json.RawMessage(ev.Raw)
	tx.CallbackCount++
	tx.UpdatedAt = time.Now()

	// Terminal status: jangan turunkan transaksi yang sudah SETTLED.
	if tx.Status == domainBilling.GatewayStatusSettled {
		if tx.InvoiceID != nil {
			return *tx.InvoiceID, true, nil
		}
		return "", true, nil
	}

	var paymentID string
	switch ev.Status {
	case domainBilling.GatewayStatusSettled:
		if tx.InvoiceID == nil {
			return "", false, fmt.Errorf("%w: %s", domainBilling.ErrGatewayTxMissingInvoice, tx.ID)
		}
		cashAccountID, incomeCategoryID := u.cashSettings(ctx, gw.Name())
		pay, perr := u.processor.ProcessCashPayment(ctx, port.CashPaymentCommand{
			TenantID:         tx.TenantID,
			InvoiceID:        *tx.InvoiceID,
			Amount:           ev.PaidAmount,
			CashAccountID:    cashAccountID,
			IncomeCategoryID: incomeCategoryID,
			ScanMethod:       domainBilling.ScanPaymentGateway,
			Reference:        tx.ExternalID,
		})
		if perr != nil && !errors.Is(perr, domainBilling.ErrInvoiceAlreadyPaid) {
			return "", false, perr
		}
		if perr == nil {
			paymentID = pay.ID
		}
		tx.Status = domainBilling.GatewayStatusSettled
		tx.PaidAt = ptrTime(time.Now())
		u.postFeeJournal(ctx, tx, gw.Name())
	default:
		tx.Status = ev.Status
	}
	if err := u.gwt.Save(ctx, tx); err != nil {
		return "", false, err
	}
	if paymentID != "" {
		if lerr := u.gwt.LinkPayment(ctx, tx.ID, paymentID, tx.FeeAmount); lerr != nil {
			logger.WithComponent("GatewayCharge").WithError(lerr).Warn("link payment gagal")
		}
	}
	if tx.InvoiceID != nil && tx.Status == domainBilling.GatewayStatusSettled {
		return *tx.InvoiceID, true, nil
	}
	return "", false, nil
}

// CheckPaymentStatus memeriksa status via gateway default (kompatibel lama).
func (u *GatewayChargeUseCase) CheckPaymentStatus(ctx context.Context, externalID string) (invoiceID string, settled bool, status string, err error) {
	return u.CheckPaymentStatusVia(ctx, "", externalID)
}

// CheckPaymentStatusVia memeriksa status transaksi terkini ke provider
// terpilih dan melunasi tagihan bila status sudah settled.
func (u *GatewayChargeUseCase) CheckPaymentStatusVia(ctx context.Context, gatewayName, externalID string) (invoiceID string, settled bool, status string, err error) {
	gw, err := u.registry.Get(ctx, gatewayName)
	if err != nil {
		return "", false, "", err
	}
	ev, err := gw.CheckStatus(ctx, externalID)
	if err != nil {
		return "", false, "", fmt.Errorf("check status: %w", err)
	}
	tx, err := u.gwt.FindByExternalID(ctx, gw.Name(), externalID)
	if err != nil {
		return "", false, ev.Status, fmt.Errorf("%w: %s/%s", domainBilling.ErrGatewayUnknownRef, gw.Name(), externalID)
	}

	// Terminal status: transaksi yang sudah SETTLED tidak boleh diturunkan statusnya.
	if tx.Status == domainBilling.GatewayStatusSettled {
		if tx.InvoiceID != nil {
			return *tx.InvoiceID, true, tx.Status, nil
		}
		return "", true, tx.Status, nil
	}

	var paymentID string
	if ev.Status == domainBilling.GatewayStatusSettled {
		if tx.InvoiceID == nil {
			return "", false, ev.Status, fmt.Errorf("%w: %s", domainBilling.ErrGatewayTxMissingInvoice, tx.ID)
		}
		cashAccountID, incomeCategoryID := u.cashSettings(ctx, gw.Name())
		pay, perr := u.processor.ProcessCashPayment(ctx, port.CashPaymentCommand{
			TenantID:         tx.TenantID,
			InvoiceID:        *tx.InvoiceID,
			Amount:           ev.PaidAmount,
			CashAccountID:    cashAccountID,
			IncomeCategoryID: incomeCategoryID,
			ScanMethod:       domainBilling.ScanPaymentGateway,
			Reference:        tx.ExternalID,
		})
		if perr != nil && !errors.Is(perr, domainBilling.ErrInvoiceAlreadyPaid) {
			return "", false, ev.Status, fmt.Errorf("process cash payment: %w", perr)
		}
		if perr == nil {
			paymentID = pay.ID
		}
		tx.Status = domainBilling.GatewayStatusSettled
		tx.PaidAt = ptrTime(time.Now())
		u.postFeeJournal(ctx, tx, gw.Name())
	} else if ev.Status != "" {
		tx.Status = ev.Status
	}
	tx.UpdatedAt = time.Now()

	if err := u.gwt.Save(ctx, tx); err != nil {
		return "", false, tx.Status, fmt.Errorf("update status: %w", err)
	}
	if paymentID != "" {
		if lerr := u.gwt.LinkPayment(ctx, tx.ID, paymentID, tx.FeeAmount); lerr != nil {
			logger.WithComponent("GatewayCharge").WithError(lerr).Warn("link payment gagal")
		}
	}
	if tx.InvoiceID != nil {
		return *tx.InvoiceID, tx.Status == domainBilling.GatewayStatusSettled, tx.Status, nil
	}
	return "", tx.Status == domainBilling.GatewayStatusSettled, tx.Status, nil
}

func (u *GatewayChargeUseCase) postFeeJournal(ctx context.Context, tx domainBilling.GatewayTransaction, gwName string) {
	if u.cashbook == nil || tx.FeeAmount <= 0 {
		return
	}
	expCat := "cc-gateway-fee"
	if u.reader != nil {
		if v := u.reader.GetValue(ctx, "isp.gateway_fee_expense_category_id", "cc-gateway-fee"); v != "" {
			expCat = v
		}
	}
	cashAccountID, _ := u.cashSettings(ctx, gwName)
	_ = u.cashbook.SaveTransaction(ctx, domainCashbook.CashTransaction{
		ID:            idgen.New("trx"),
		TenantID:      tx.TenantID,
		TransactionNo: fmt.Sprintf("TRX-%s-%s", time.Now().Format("200601"), idgen.Digits(6)),
		AccountID:     cashAccountID,
		CategoryID:    expCat,
		Direction:     domainCashbook.DirectionOut,
		Amount:        tx.FeeAmount,
		TrxDate:       time.Now().UTC(),
		SourceType:    domainCashbook.SourceExpense,
		SourceID:      tx.ID,
		Description:   fmt.Sprintf("Biaya payment gateway %s %s", tx.Gateway, tx.ExternalID),
		CreatedAt:     time.Now().UTC(),
	})
}

// SyncPendingTransactions memeriksa transaksi PENDING yang lebih tua dari batas stale.
func (u *GatewayChargeUseCase) SyncPendingTransactions(ctx context.Context) (synced int, expired int, err error) {
	fRepo, ok := u.gwt.(port.GatewayTransactionFilteredRepository)
	if !ok {
		return 0, 0, nil
	}
	staleMin := 15
	if u.reader != nil {
		if v, err := strconv.Atoi(u.reader.GetValue(ctx, "isp.gateway_sync_stale_minutes", "15")); err == nil && v > 0 {
			staleMin = v
		}
	}
	before := time.Now().UTC().Add(-time.Duration(staleMin) * time.Minute)
	pendingList, err := fRepo.ListPendingBefore(ctx, "tenant-default", before, 50)
	if err != nil {
		return 0, 0, err
	}
	now := time.Now().UTC()
	for _, tx := range pendingList {
		if tx.ExpiresAt != nil && tx.ExpiresAt.Before(now) {
			tx.Status = domainBilling.GatewayStatusExpired
			tx.UpdatedAt = now
			_ = u.gwt.Save(ctx, tx)
			expired++
			continue
		}
		_, _, _, _ = u.CheckPaymentStatusVia(ctx, tx.Gateway, tx.ExternalID)
		synced++
	}
	return synced, expired, nil
}

// cashSettings membaca akun/kategori kas per gateway, dengan fallback key
// global lalu key legacy Tripay (kompatibel deployment lama).
func (u *GatewayChargeUseCase) cashSettings(ctx context.Context, gatewayName string) (string, string) {
	name := strings.ToLower(gatewayName)
	cash := firstNonEmpty(
		u.reader.GetValue(ctx, "gw."+name+".cash_account_id", ""),
		firstNonEmpty(
			u.reader.GetValue(ctx, "gw.cash_account_id", ""),
			u.reader.GetValue(ctx, "gw.tripay.cash_account_id", "ca-1001-kas-kantor"),
		),
	)
	income := firstNonEmpty(
		u.reader.GetValue(ctx, "gw."+name+".income_category_id", ""),
		firstNonEmpty(
			u.reader.GetValue(ctx, "gw.income_category_id", ""),
			u.reader.GetValue(ctx, "gw.tripay.income_category_id", "cc-tagihan"),
		),
	)
	return cash, income
}

// chargeResultFromTx mengembalikan ChargeResult dari transaksi tersimpan
// (dipakai saat reuse transaksi PENDING).
func chargeResultFromTx(tx domainBilling.GatewayTransaction) port.ChargeResult {
	res := port.ChargeResult{
		ExternalID: tx.ExternalID, PaymentURL: tx.PaymentURL,
		QRString: tx.QRString, VANumber: tx.QRString,
		Channel: tx.PaymentChannel, FeeAmount: tx.FeeAmount,
		Status: tx.Status, RawResponse: tx.RawCallback,
	}
	if tx.ExpiresAt != nil {
		res.ExpiresAt = *tx.ExpiresAt
	}
	return res
}

func ptrTime(t time.Time) *time.Time { return &t }

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
