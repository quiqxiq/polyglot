package billing

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	domainPlan "github.com/quixiq/polyglot/internal/domain/plan"
	domainSubscription "github.com/quixiq/polyglot/internal/domain/subscription"
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

// Run creates invoices for the given 'YYYY-MM' period.
func (u *RunBillingUseCase) Run(ctx context.Context, tenantID, period string) (BillingRunResult, error) {
	if !periodRe.MatchString(period) {
		return BillingRunResult{}, fmt.Errorf("%w: period must be YYYY-MM", domainBilling.ErrInvalidInput)
	}
	active, err := u.subs.ListActive(ctx)
	if err != nil {
		return BillingRunResult{}, fmt.Errorf("list active subscriptions: %w", err)
	}

	res := BillingRunResult{}
	for _, sub := range active {
		if _, err := u.invoices.FindBySubscriptionPeriod(ctx, sub.ID, period); err == nil {
			res.Skipped++
			continue // sudah ditagih
		}
		pl, err := u.plans.FindByID(ctx, sub.PlanID)
		if err != nil {
			res.Skipped++
			logger.WithComponent("BillingUC").WithFields(map[string]any{
				"subscription_id": sub.ID, "plan_id": sub.PlanID,
			}).Warn("plan missing, subscription skipped")
			continue
		}

		dueDays := 0
		if u.reader != nil {
			dueDays = atoiSafe(u.reader.GetValue(ctx, "isp.billing_due_days", "0"))
		}
		inv, items := buildMonthlyInvoice(sub, pl, period, u.now(), dueDays)
		if err := u.invoices.SaveWithItems(ctx, inv, items); err != nil {
			return res, fmt.Errorf("save invoice %s: %w", inv.InvoiceNumber, err)
		}
		res.Created++
	}
	return res, nil
}

// buildMonthlyInvoice menyusun faktur + item dari harga paket atau override
// custom_price langganan; pajak mengikuti tax_percent paket.
func buildMonthlyInvoice(sub domainSubscription.Subscription, pl domainPlan.ServicePlan, period string, now time.Time, dueDays int) (domainBilling.Invoice, []domainBilling.InvoiceItem) {
	base := pl.Price
	if sub.CustomPrice != nil && *sub.CustomPrice > 0 {
		base = *sub.CustomPrice
	}
	tax := base * pl.TaxPercent / 100

	year, month := parsePeriod(period)
	var due time.Time
	if dueDays > 0 {
		// Jatuh tempo = terbit + N hari kalender (settings).
		due = time.Date(year, month, 1, 23, 59, 59, 0, time.UTC).AddDate(0, 0, dueDays)
	} else {
		due = endOfMonthPeriod(year, month)
	}

	invID := idgen.New("inv")
	subID := sub.ID
	inv := domainBilling.Invoice{
		ID:                invID,
		TenantID:          orTenantID(sub.TenantID),
		InvoiceNumber:     fmt.Sprintf("INV-%s-%04d", strings.ReplaceAll(period, "-", ""), now.UnixNano()%10000),
		CustomerID:        sub.CustomerID,
		SubscriptionID:    &subID,
		Period:            period,
		Subtotal:          base,
		TaxAmount:         tax,
		Total:             base + tax,
		DueDate:           due,
		Status:            domainBilling.StatusUnpaid,
		QRPayload:         "polyglot://invoice/" + invID,
		ManualPaymentCode: "PAY-" + idgen.Digits(6),
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	items := []domainBilling.InvoiceItem{{
		ID:          idgen.New("itm"),
		InvoiceID:   invID,
		Description: fmt.Sprintf("Paket %s (%s)", pl.Name, period),
		Quantity:    1,
		UnitPrice:   base,
		Amount:      base,
		ItemType:    domainBilling.ItemTypeSubscriptionFee,
		CreatedAt:   now,
	}}
	return inv, items
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
	fmt.Sscanf(p, "%d-%d", &y, &m)
	return y, time.Month(m)
}

func endOfMonthPeriod(year int, month time.Month) time.Time {
	firstNext := time.Date(year, month+1, 1, 23, 59, 59, 0, time.UTC)
	return firstNext.AddDate(0, 0, -1)
}

func orTenantID(t string) string {
	if t == "" {
		return "tenant-default"
	}
	return t
}
