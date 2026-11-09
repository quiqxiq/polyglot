package billing

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"math"
	"strings"
	"time"

	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	domainNotification "github.com/quixiq/polyglot/internal/domain/notification"
	domainTenant "github.com/quixiq/polyglot/internal/domain/tenant"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/idgen"
	"github.com/quixiq/polyglot/pkg/phone"
	"github.com/xuri/excelize/v2"
)

// InvoiceUseCase manages invoice queries and lifecycle operations.
type InvoiceUseCase struct {
	repo      port.InvoiceRepository
	canceller port.InvoiceCanceller
	settings  port.SettingReader
	notif     port.NotificationRepository
	customers port.CustomerRepository
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

// WithNotifications menautkan notifikasi dan pelanggan.
func (u *InvoiceUseCase) WithNotifications(n port.NotificationRepository, c port.CustomerRepository) *InvoiceUseCase {
	u.notif = n
	u.customers = c
	return u
}

// CreateInvoiceItemInput item baris saat membuat tagihan manual.
type CreateInvoiceItemInput struct {
	Description string
	Quantity    float64
	UnitPrice   float64
	ItemType    string
}

// CreateManualInvoiceInput input untuk membuat tagihan manual/ad-hoc.
type CreateManualInvoiceInput struct {
	TenantID       string
	CustomerID     string
	SubscriptionID *string
	Period         string
	DueDate        time.Time
	TaxPercent     float64
	Items          []CreateInvoiceItemInput
	Notes          string
}

// CreateManualInvoice menerbitkan faktur manual baru di luar generator siklus.
func (u *InvoiceUseCase) CreateManualInvoice(ctx context.Context, in CreateManualInvoiceInput) (domainBilling.Invoice, error) {
	if u.repo == nil {
		return domainBilling.Invoice{}, domainBilling.ErrRepositoryUnavailable
	}
	if in.CustomerID == "" {
		return domainBilling.Invoice{}, fmt.Errorf("%w: customer_id is required", domainBilling.ErrInvalidInput)
	}
	if len(in.Items) == 0 {
		return domainBilling.Invoice{}, fmt.Errorf("%w: at least one line item is required", domainBilling.ErrInvalidInput)
	}
	if in.DueDate.IsZero() {
		in.DueDate = time.Now().UTC().AddDate(0, 0, 7)
	}
	if in.TenantID == "" {
		in.TenantID = domainTenant.DefaultID
	}
	now := time.Now().UTC()
	period := in.Period
	if period == "" {
		period = now.Format("2006-01")
	}

	var subtotal float64
	items := make([]domainBilling.InvoiceItem, len(in.Items))
	invID := idgen.New("inv")
	for i, it := range in.Items {
		qty := it.Quantity
		if qty <= 0 {
			qty = 1
		}
		amount := roundInvoiceMoney(qty * it.UnitPrice)
		subtotal += amount
		itemType := it.ItemType
		if itemType == "" {
			itemType = domainBilling.ItemTypeAdHoc
		}
		items[i] = domainBilling.InvoiceItem{
			ID:          idgen.New("itm"),
			InvoiceID:   invID,
			Description: it.Description,
			Quantity:    qty,
			UnitPrice:   it.UnitPrice,
			Amount:      amount,
			ItemType:    itemType,
			CreatedAt:   now,
		}
	}
	subtotal = roundInvoiceMoney(subtotal)
	if subtotal <= 0 {
		return domainBilling.Invoice{}, fmt.Errorf("%w: subtotal must be greater than zero", domainBilling.ErrInvalidInput)
	}

	taxPercent := in.TaxPercent
	if taxPercent <= 0 && u.settings != nil {
		cfg := port.LoadISPSettings(ctx, u.settings)
		taxPercent = cfg.TaxPercent
	}
	var taxAmount float64
	if taxPercent > 0 {
		taxAmount = roundInvoiceMoney(subtotal * taxPercent / 100)
	}
	total := roundInvoiceMoney(subtotal + taxAmount)

	invNumber := fmt.Sprintf("INV-%s-%s", now.Format("200601"), idgen.Digits(6))
	qrPayload := fmt.Sprintf("QR-%s-%s", invNumber, idgen.Digits(4))
	paymentCode := "PAY-" + idgen.Digits(6)

	inv := domainBilling.Invoice{
		ID:                invID,
		TenantID:          in.TenantID,
		InvoiceNumber:     invNumber,
		CustomerID:        in.CustomerID,
		SubscriptionID:    in.SubscriptionID,
		Period:            period,
		InvoiceType:       domainBilling.InvoiceTypeManual,
		Subtotal:          subtotal,
		TaxPercent:        taxPercent,
		TaxAmount:         taxAmount,
		Total:             total,
		DueDate:           in.DueDate,
		Status:            domainBilling.StatusUnpaid,
		QRPayload:         qrPayload,
		ManualPaymentCode: paymentCode,
		IssuedAt:          now,
		Notes:             in.Notes,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	if err := u.repo.SaveWithItems(ctx, inv, items); err != nil {
		return domainBilling.Invoice{}, fmt.Errorf("save manual invoice: %w", err)
	}

	// Kirim notifikasi INVOICE_CREATED jika notif repo ada
	if u.notif != nil && u.customers != nil {
		if cust, err := u.customers.FindByID(ctx, in.CustomerID); err == nil && cust.Phone != "" {
			content := renderTemplateOrDefault(ctx, u.notif, cust.TenantID, "INVOICE_CREATED",
				"Yth {{customer_name}}, tagihan periode {{period}} sebesar Rp{{total}} telah terbit. Jatuh tempo {{due_date}}. Kode bayar: {{payment_code}}.",
				map[string]string{
					"customer_name": cust.Name,
					"period":        inv.Period,
					"total":         fmt.Sprintf("%.0f", inv.Total),
					"due_date":      inv.DueDate.Format("02 Jan 2006"),
					"payment_code":  inv.ManualPaymentCode,
				})
			_ = u.notif.Queue(ctx, domainNotification.WANotification{
				ID:             idgen.New("wa"),
				TenantID:       cust.TenantID,
				CustomerID:     &cust.ID,
				InvoiceID:      &inv.ID,
				RecipientPhone: phone.Normalize(cust.Phone),
				MessageType:    "INVOICE_CREATED",
				MessageContent: content,
				Status:         domainNotification.StatusQueued,
				CreatedAt:      now,
			})
		}
	}

	return inv, nil
}

// FindFiltered mengembalikan daftar invoice sesuai filter server-side.
func (u *InvoiceUseCase) FindFiltered(ctx context.Context, f port.InvoiceFilter) ([]domainBilling.Invoice, int, error) {
	if fRepo, ok := u.repo.(port.InvoiceFilteredRepository); ok {
		return fRepo.FindFiltered(ctx, f)
	}
	invoices, err := u.ListInvoices(ctx, f.CustomerID)
	if err != nil {
		return nil, 0, err
	}
	return invoices, len(invoices), nil
}

// Summary mengembalikan ringkasan total tagihan/lunas/outstanding.
func (u *InvoiceUseCase) Summary(ctx context.Context, f port.InvoiceFilter) (port.InvoiceSummary, error) {
	if fRepo, ok := u.repo.(port.InvoiceFilteredRepository); ok {
		return fRepo.Summary(ctx, f)
	}
	return port.InvoiceSummary{}, nil
}

// ExportInvoices mengekspor invoice sesuai filter ke CSV atau XLSX.
func (u *InvoiceUseCase) ExportInvoices(ctx context.Context, f port.InvoiceFilter, format string) ([]byte, string, string, error) {
	f.Limit = 50000
	invoices, _, err := u.FindFiltered(ctx, f)
	if err != nil {
		return nil, "", "", err
	}

	headers := []string{"No. Invoice", "Customer ID", "Period", "Type", "Subtotal", "Tax", "Total", "Paid", "Status", "Due Date", "Kode Bayar"}
	rows := make([][]string, len(invoices))
	for i, inv := range invoices {
		rows[i] = []string{
			inv.InvoiceNumber,
			inv.CustomerID,
			inv.Period,
			inv.InvoiceType,
			fmt.Sprintf("%.2f", inv.Subtotal),
			fmt.Sprintf("%.2f", inv.TaxAmount),
			fmt.Sprintf("%.2f", inv.Total),
			fmt.Sprintf("%.2f", inv.PaidAmount),
			inv.Status,
			inv.DueDate.Format("2006-01-02"),
			inv.ManualPaymentCode,
		}
	}

	filename := fmt.Sprintf("invoices-%s", time.Now().Format("20060102150405"))
	if strings.ToLower(format) == "xlsx" {
		file := excelize.NewFile()
		defer func() { _ = file.Close() }()
		sheet := "Invoices"
		file.SetSheetName("Sheet1", sheet)
		for colIdx, h := range headers {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, 1)
			_ = file.SetCellValue(sheet, cell, h)
		}
		for rowIdx, r := range rows {
			for colIdx, val := range r {
				cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
				_ = file.SetCellValue(sheet, cell, val)
			}
		}
		var buf bytes.Buffer
		if err := file.Write(&buf); err != nil {
			return nil, "", "", err
		}
		return buf.Bytes(), "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", filename + ".xlsx", nil
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write(headers)
	_ = w.WriteAll(rows)
	w.Flush()
	return buf.Bytes(), "text/csv", filename + ".csv", nil
}

func (u *InvoiceUseCase) ListInvoices(ctx context.Context, customerID string) ([]domainBilling.Invoice, error) {
	if u.repo == nil {
		return nil, domainBilling.ErrRepositoryUnavailable
	}
	if customerID != "" {
		return u.repo.FindByCustomerID(ctx, customerID)
	}
	invoices, err := u.repo.FindPaged(ctx, port.PageFilter{TenantID: domainTenant.DefaultID})
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

// ApplyDiscount menerapkan diskon/promo code manual pada faktur yang belum
// lunas (planning v2 §1.7/Fase 5). Total dihitung ulang:
// Total = Subtotal − Diskon + Pajak.
func (u *InvoiceUseCase) ApplyDiscount(ctx context.Context, invoiceID, discountType string, value float64, promoCode string) (domainBilling.Invoice, error) {
	if u.repo == nil {
		return domainBilling.Invoice{}, domainBilling.ErrRepositoryUnavailable
	}
	inv, err := u.repo.FindByID(ctx, invoiceID)
	if err != nil {
		return domainBilling.Invoice{}, domainBilling.ErrNotFound
	}
	switch inv.Status {
	case domainBilling.StatusPaid, domainBilling.StatusCancelled:
		return domainBilling.Invoice{}, fmt.Errorf("%w: invoice %s berstatus %s", domainBilling.ErrInvalidTransition, inv.ID, inv.Status)
	}
	dt := domainBilling.DiscountType(discountType)
	switch dt {
	case domainBilling.DiscountTypeNone:
		inv.Discount = 0
		inv.DiscountType = string(dt)
		inv.DiscountValue = 0
		inv.DiscountAmount = 0
		inv.PromoCode = ""
	case domainBilling.DiscountTypePercent, domainBilling.DiscountTypeAmount:
		if value <= 0 {
			return domainBilling.Invoice{}, fmt.Errorf("%w: discount value must be greater than zero", domainBilling.ErrInvalidInput)
		}
		amount := domainBilling.ApplyDiscount(inv.Subtotal, dt, value)
		inv.Discount = amount
		inv.DiscountType = string(dt)
		inv.DiscountValue = value
		inv.DiscountAmount = amount
		inv.PromoCode = promoCode
	default:
		return domainBilling.Invoice{}, fmt.Errorf("%w: discount_type harus NONE, PERCENT, atau AMOUNT", domainBilling.ErrInvalidInput)
	}
	inv.Total = roundInvoiceMoney(inv.Subtotal - inv.Discount + inv.TaxAmount)
	inv.UpdatedAt = time.Now().UTC()
	if err := u.repo.Save(ctx, inv); err != nil {
		return domainBilling.Invoice{}, fmt.Errorf("save invoice discount: %w", err)
	}
	return inv, nil
}

// roundInvoiceMoney membulatkan nominal uang ke 2 desimal.
func roundInvoiceMoney(v float64) float64 {
	return math.Round(v*100) / 100
}
