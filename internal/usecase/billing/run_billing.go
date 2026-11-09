package billing

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	domainNotification "github.com/quixiq/polyglot/internal/domain/notification"
	domainPlan "github.com/quixiq/polyglot/internal/domain/plan"
	domainSubscription "github.com/quixiq/polyglot/internal/domain/subscription"
	domainTenant "github.com/quixiq/polyglot/internal/domain/tenant"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/idgen"
	"github.com/quixiq/polyglot/pkg/logger"
)

var periodRe = regexp.MustCompile(`^\d{4}-(0[1-9]|1[0-2])$`)

// BillingRunResult rekap satu siklus generator tagihan.
type BillingRunResult struct {
	Created int
	Skipped int // sudah tertagih atau data tidak lengkap
}

// RunBillingUseCase generates monthly invoices for all ACTIVE subscriptions,
// idempoten per (subscription_id, period) (DATABASE-SCHEMA-ISP.md §3 poin 3).
type RunBillingUseCase struct {
	subs     port.SubscriptionRepository
	plans    port.ServicePlanRepository
	invoices port.InvoiceRepository
	// settings opsional — bila diisi, jatuh tempo mengikuti
	// isp.billing_due_days (janji desain fase 3.5); bila nil, fallback
	// akhir bulan periode.
	reader port.SettingReader

	// notif/customers opsional — bila diisi, faktur baru mengantrekan
	// notifikasi WhatsApp INVOICE_CREATED (planning v2 Fase 1).
	notif     port.NotificationRepository
	customers port.CustomerRepository

	now func() time.Time
}

// NewRunBillingUseCase wires dependencies.
func NewRunBillingUseCase(subs port.SubscriptionRepository, plans port.ServicePlanRepository, inv port.InvoiceRepository) *RunBillingUseCase {
	return &RunBillingUseCase{subs: subs, plans: plans, invoices: inv, now: time.Now}
}

// WithSettings menautkan sumber konfigurasi dinamis.
func (u *RunBillingUseCase) WithSettings(r port.SettingReader) *RunBillingUseCase {
	u.reader = r
	return u
}

// WithNotifications menautkan antrean notifikasi WhatsApp faktur baru.
func (u *RunBillingUseCase) WithNotifications(notif port.NotificationRepository, customers port.CustomerRepository) *RunBillingUseCase {
	u.notif = notif
	u.customers = customers
	return u
}

// Run creates invoices for the given 'YYYY-MM' period.
func (u *RunBillingUseCase) Run(ctx context.Context, tenantID, period string) (BillingRunResult, error) {
	if !periodRe.MatchString(period) {
		return BillingRunResult{}, fmt.Errorf("%w: period must be YYYY-MM", domainBilling.ErrInvalidInput)
	}
	// Status yang ditagih dapat dikonfigurasi (planning v2 D4):
	// ACTIVE + ISOLATED selalu; SUSPENDED hanya bila isp.bill_suspended=true;
	// ISOLATED dapat dinonaktifkan via isp.bill_isolated=false.
	statuses := []string{domainSubscription.StatusActive}
	if u.reader != nil {
		if !strings.EqualFold(u.reader.GetValue(ctx, "isp.bill_isolated", "true"), "false") {
			statuses = append(statuses, domainSubscription.StatusIsolated)
		}
		if strings.EqualFold(u.reader.GetValue(ctx, "isp.bill_suspended", "false"), "true") {
			statuses = append(statuses, domainSubscription.StatusSuspended)
		}
	} else {
		statuses = append(statuses, domainSubscription.StatusIsolated)
	}
	billable, err := u.subs.ListByStatuses(ctx, statuses)
	if err != nil {
		return BillingRunResult{}, fmt.Errorf("list billable subscriptions: %w", err)
	}

	// Konfigurasi Billing Ops v3: mode siklus, lead prabayar, tutup buku,
	// dan PPN default mitra. Semua fallback ke perilaku lama bila setting kosong.
	mode := "PASCABAYAR"
	leadDays := 0
	closingDate := 1
	taxDefault := 0.0
	dueDays := 0
	if u.reader != nil {
		mode = strings.ToUpper(strings.TrimSpace(u.reader.GetValue(ctx, "isp.billing_mode", "PASCABAYAR")))
		leadDays = atoiSafe(u.reader.GetValue(ctx, "isp.prepay_invoice_lead_days", "7"))
		closingDate = atoiSafe(u.reader.GetValue(ctx, "isp.closing_date", "1"))
		if closingDate < 1 || closingDate > 28 {
			closingDate = 1
		}
		taxDefault, _ = strconv.ParseFloat(strings.TrimSpace(u.reader.GetValue(ctx, "isp.tax_percent", "0")), 64)
		dueDays = atoiSafe(u.reader.GetValue(ctx, "isp.billing_due_days", "0"))
	}
	if leadDays < 0 {
		leadDays = 0
	}

	res := BillingRunResult{}
	now := u.now().UTC()
	for _, sub := range billable {
		if _, err := u.invoices.FindBySubscriptionPeriod(ctx, sub.ID, period); err == nil {
			res.Skipped++
			continue // sudah ditagih (termasuk faktur pertama di bulan pasang)
		}
		// Terbitkan hanya saat siklus tagihan langganan sudah tiba (F3-3).
		// Mode PRABAYAR boleh menerbitkan periode berikutnya lebih awal;
		// periode lampau (backfill) tetap diproses penuh.
		if !invoiceIssueDue(period, anchorDayOf(sub, closingDate), now, mode, leadDays) {
			res.Skipped++
			continue
		}
		pl, err := u.plans.FindByID(ctx, sub.PlanID)
		if err != nil {
			res.Skipped++
			logger.WithComponent("BillingUC").WithFields(map[string]any{
				"subscription_id": sub.ID, "plan_id": sub.PlanID,
			}).Warn("plan missing, subscription skipped")
			continue
		}

		inv, items, periodEnd := buildMonthlyInvoice(sub, pl, period, now, dueDays, mode, leadDays, closingDate, taxDefault)
		if err := u.invoices.SaveWithItems(ctx, inv, items); err != nil {
			return res, fmt.Errorf("save invoice %s: %w", inv.InvoiceNumber, err)
		}
		u.queueInvoiceCreated(ctx, sub, inv)
		// Majukan penanda jatuh tempo berikutnya (planning v2 §3.2).
		nextDue := periodEnd.AddDate(0, 0, 1)
		sub.NextDueDate = &nextDue
		if err := u.subs.Save(ctx, sub); err != nil {
			logger.WithComponent("BillingUC").WithFields(map[string]any{
				"subscription_id": sub.ID,
			}).WithError(err).Warn("gagal simpan next_due_date subscription")
		}
		res.Created++
	}
	return res, nil
}

// queueInvoiceCreated mengantrekan notifikasi WhatsApp INVOICE_CREATED untuk
// faktur yang baru terbit (planning v2 Fase 1). Best-effort: kegagalan
// notifikasi tidak menggagalkan billing run.
func (u *RunBillingUseCase) queueInvoiceCreated(ctx context.Context, sub domainSubscription.Subscription, inv domainBilling.Invoice) {
	if u.notif == nil || u.customers == nil {
		return
	}
	cust, err := u.customers.FindByID(ctx, sub.CustomerID)
	if err != nil {
		logger.WithComponent("BillingUC").WithFields(map[string]any{
			"customer_id": sub.CustomerID, "invoice_id": inv.ID,
		}).WithError(err).Warn("customer not found; INVOICE_CREATED skipped")
		return
	}
	content := renderTemplateOrDefault(ctx, u.notif, cust.TenantID, "INVOICE_CREATED",
		"Yth {{customer_name}}, tagihan periode {{period}} sebesar Rp{{total}} telah terbit. Jatuh tempo {{due_date}}. Kode bayar: {{payment_code}}.",
		map[string]string{
			"customer_name": cust.Name,
			"period":        inv.Period,
			"total":         fmt.Sprintf("%.0f", inv.Total),
			"due_date":      inv.DueDate.Format("02 Jan 2006"),
			"payment_code":  inv.ManualPaymentCode,
		})
	n := domainNotification.WANotification{
		ID:             idgen.New("wa"),
		TenantID:       cust.TenantID,
		CustomerID:     &cust.ID,
		InvoiceID:      &inv.ID,
		RecipientPhone: cust.Phone,
		MessageType:    "INVOICE_CREATED",
		MessageContent: content,
		Status:         domainNotification.StatusQueued,
		CreatedAt:      u.now().UTC(),
	}
	if err := u.notif.Queue(ctx, n); err != nil {
		logger.WithComponent("BillingUC").WithFields(map[string]any{
			"invoice_id": inv.ID,
		}).WithError(err).Warn("queue INVOICE_CREATED failed")
	}
}

// anchorDayOf returns the effective billing anchor day of a subscription
// (BillingAnchorDay > BillingDay > isp.closing_date > default 1), clamped 1..28.
func anchorDayOf(sub domainSubscription.Subscription, closingDate int) int {
	day := sub.BillingAnchorDay
	if day <= 0 {
		day = sub.BillingDay
	}
	if day <= 0 {
		day = closingDate
	}
	if day <= 0 {
		day = 1
	}
	if day > 28 {
		day = 28
	}
	return day
}

// anchorWindow returns the billing cycle window (start..end inclusive) for
// the 'YYYY-MM' period of an anchor-day billing cycle, plus its day count.
// Mis. anchor 1 → 1 s/d akhir bulan; anchor 15 → 15 s/d 14 bulan berikutnya.
func anchorWindow(year int, month time.Month, anchorDay int) (start, end time.Time, days int) {
	last := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
	a := anchorDay
	if a > last {
		a = last
	}
	start = time.Date(year, month, a, 0, 0, 0, 0, time.UTC)
	end = start.AddDate(0, 1, -1)
	days = int(end.Sub(start).Hours()/24) + 1
	return start, end, days
}

// buildMonthlyInvoice menyusun faktur + item dari harga paket atau override
// custom_price langganan; pajak mengikuti tax_percent paket. Kembalian ketiga
// adalah akhir periode siklus (untuk komputasi next_due_date pemanggil).
func buildMonthlyInvoice(sub domainSubscription.Subscription, pl domainPlan.ServicePlan, period string, now time.Time, dueDays int, mode string, leadDays, closingDate int, taxDefault float64) (domainBilling.Invoice, []domainBilling.InvoiceItem, time.Time) {
	base := pl.Price
	if sub.CustomPrice != nil && *sub.CustomPrice > 0 {
		base = *sub.CustomPrice
	}
	// PPN: tarif paket menang; fallback PPN default mitra. Mode INCLUSIVE
	// memperlakukan harga paket sebagai harga akhir (PPN di dalamnya).
	taxPercent := pl.TaxPercent
	if taxPercent <= 0 {
		taxPercent = taxDefault
	}
	if taxPercent < 0 {
		taxPercent = 0
	}
	tax := 0.0
	subtotal := base
	total := base
	if strings.EqualFold(strings.TrimSpace(pl.TaxMode), "INCLUSIVE") && taxPercent > 0 {
		tax = roundMoney2(base * taxPercent / (100 + taxPercent))
		subtotal = roundMoney2(base - tax)
	} else {
		tax = roundMoney2(base * taxPercent / 100)
		total = roundMoney2(base + tax)
	}

	year, month := parsePeriod(period)
	anchorDay := anchorDayOf(sub, closingDate)
	start, end, days := anchorWindow(year, month, anchorDay)

	var due time.Time
	if strings.EqualFold(mode, "PRABAYAR") {
		// Prabayar: jatuh tempo = hari mulai siklus (bayar sebelum memakai).
		due = time.Date(start.Year(), start.Month(), start.Day(), 23, 59, 59, 0, time.UTC)
	} else if dueDays > 0 {
		// Pascabayar: jatuh tempo = awal siklus tagihan + N hari kalender.
		due = time.Date(start.Year(), start.Month(), start.Day(), 23, 59, 59, 0, time.UTC).AddDate(0, 0, dueDays)
	} else {
		due = endOfMonthPeriod(year, month)
	}
	_ = leadDays // lead dipakai invoiceIssueDue; disimpan agar signature konsisten

	invID := idgen.New("inv")
	subID := sub.ID
	inv := domainBilling.Invoice{
		ID:                invID,
		TenantID:          orTenantID(sub.TenantID),
		InvoiceNumber:     fmt.Sprintf("INV-%s-%s", strings.ReplaceAll(period, "-", ""), idgen.Digits(6)),
		CustomerID:        sub.CustomerID,
		SubscriptionID:    &subID,
		Period:            period,
		InvoiceType:       domainBilling.InvoiceTypeRecurring,
		PeriodStart:       &start,
		PeriodEnd:         &end,
		ProratedDays:      days,
		PeriodDays:        days,
		Subtotal:          subtotal,
		TaxPercent:        taxPercent,
		TaxAmount:         tax,
		Total:             total,
		DueDate:           due,
		Status:            domainBilling.StatusUnpaid,
		QRPayload:         "polyglot://invoice/" + invID,
		ManualPaymentCode: "PAY-" + idgen.Digits(6),
		IssuedAt:          now,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	items := []domainBilling.InvoiceItem{{
		ID:          idgen.New("itm"),
		InvoiceID:   invID,
		Description: fmt.Sprintf("Paket %s (%s)", pl.Name, period),
		Quantity:    1,
		UnitPrice:   subtotal,
		Amount:      subtotal,
		ItemType:    domainBilling.ItemTypeSubscriptionFee,
		CreatedAt:   now,
	}}
	return inv, items, end
}

// WithClock mengganti sumber waktu (untuk test deterministik).
func (u *RunBillingUseCase) WithClock(now func() time.Time) *RunBillingUseCase {
	u.now = now
	return u
}

// invoiceIssueDue melaporkan apakah invoice untuk 'YYYY-MM' sudah waktunya
// diterbitkan. Pascabayar: saat awal siklus tiba. Prabayar: leadDays sebelum
// awal siklus (termasuk periode berikutnya). Periode lampau = backfill.
func invoiceIssueDue(period string, anchorDay int, now time.Time, mode string, leadDays int) bool {
	year, month := parsePeriod(period)
	if year == 0 {
		return false
	}
	start, _, _ := anchorWindow(year, month, anchorDay)
	prabayar := strings.EqualFold(mode, "PRABAYAR")
	issueAt := start
	if prabayar {
		issueAt = start.AddDate(0, 0, -leadDays)
	}
	nowMonth := now.UTC().Format("2006-01")
	switch {
	case period < nowMonth:
		return true // backfill periode lampau
	case period == nowMonth:
		return !now.UTC().Before(issueAt)
	default:
		// Periode depan hanya boleh untuk PRABAYAR dalam jendela lead.
		if !prabayar {
			return false
		}
		nextMonth := now.UTC().AddDate(0, 1, 0).Format("2006-01")
		return period == nextMonth && !now.UTC().Before(issueAt)
	}
}

func atoiSafe(s string) int {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil || n < 0 {
		return 0
	}
	return n
}

func parsePeriod(p string) (int, time.Month) {
	var y int
	var m int
	if _, err := fmt.Sscanf(p, "%d-%d", &y, &m); err != nil {
		return 0, 0
	}
	return y, time.Month(m)
}

func endOfMonthPeriod(year int, month time.Month) time.Time {
	firstNext := time.Date(year, month+1, 1, 23, 59, 59, 0, time.UTC)
	return firstNext.AddDate(0, 0, -1)
}

func orTenantID(t string) string {
	if t == "" {
		return domainTenant.DefaultID
	}
	return t
}

// roundMoney2 membulatkan nominal uang ke 2 desimal.
func roundMoney2(v float64) float64 {
	return math.Round(v*100) / 100
}
