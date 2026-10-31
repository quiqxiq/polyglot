package billing_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quixiq/polyglot/internal/domain/billing"
)

func baseFirstInvoiceInput(now time.Time) billing.FirstInvoiceInput {
	return billing.FirstInvoiceInput{
		InvoiceID:         "inv-1",
		InvoiceNumber:     "INV-202608-000001",
		ManualPaymentCode: "PAY-000001",
		ItemIDFee:         "itm-fee",
		ItemIDInstall:     "itm-install",
		SubscriptionID:    "sub-1",
		PlanName:          "Home 10Mbps",
		PlanPrice:         100000,
		InstallationFee:   50000,
		TaxPercent:        11,
		ProrateMode:       billing.ProrateModeCalendarDay,
		DiscountType:      billing.DiscountTypeAmount,
		DiscountValue:     10000,
		InstallDate:       time.Date(2026, time.August, 20, 9, 0, 0, 0, time.UTC),
		AnchorDay:         1,
		FirstDueDays:      7,
		Now:               time.Date(2026, time.August, 20, 10, 0, 0, 0, time.UTC),
	}
}

func TestBuildFirstInvoice(t *testing.T) {
	t.Parallel()
	in := baseFirstInvoiceInput(time.Date(2026, time.August, 20, 10, 0, 0, 0, time.UTC))
	inv, items := billing.BuildFirstInvoice(in)

	// Prorata 12/31 dari 100000 = 38709.68.
	assert.InDelta(t, 38709.68, inv.Subtotal-50000, 0.005, "komponen prorata")
	assert.InDelta(t, 88709.68, inv.Subtotal, 0.005, "subtotal = prorata + biaya pasang")
	assert.InDelta(t, 10000, inv.Discount, 0.005, "discount")
	assert.InDelta(t, 10000, inv.DiscountAmount, 0.005, "discount amount")
	assert.InDelta(t, 8658.06, inv.TaxAmount, 0.005, "PPN 11% dari (subtotal - diskon)")
	assert.InDelta(t, 87367.74, inv.Total, 0.005, "total akhir")
	assert.Equal(t, 12, inv.ProratedDays)
	assert.Equal(t, 31, inv.PeriodDays)
	assert.Equal(t, "2026-08", inv.Period)
	assert.Equal(t, billing.InvoiceTypeFirstInstallation, inv.InvoiceType)
	assert.Equal(t, billing.StatusUnpaid, inv.Status)
	assert.Equal(t, "tenant-default", inv.TenantID)
	require.NotNil(t, inv.SubscriptionID)
	assert.Equal(t, "sub-1", *inv.SubscriptionID)
	require.NotNil(t, inv.PeriodStart)
	require.NotNil(t, inv.PeriodEnd)
	assert.Equal(t, date(2026, time.August, 20), *inv.PeriodStart)
	assert.Equal(t, date(2026, time.August, 31), *inv.PeriodEnd)
	assert.Equal(t, time.Date(2026, time.August, 27, 23, 59, 59, 0, time.UTC), inv.DueDate)
	assert.Equal(t, "polyglot://invoice/inv-1", inv.QRPayload)

	require.Len(t, items, 2)
	assert.Equal(t, billing.ItemTypeSubscriptionFee, items[0].ItemType)
	assert.InDelta(t, 38709.68, items[0].Amount, 0.005)
	assert.Equal(t, billing.ItemTypeInstallationFee, items[1].ItemType)
	assert.InDelta(t, 50000, items[1].Amount, 0.005)
}

func TestBuildFirstInvoice_InstallOnAnchorDayFullPeriod(t *testing.T) {
	t.Parallel()
	in := baseFirstInvoiceInput(time.Date(2026, time.September, 1, 8, 0, 0, 0, time.UTC))
	in.InstallDate = time.Date(2026, time.September, 1, 8, 0, 0, 0, time.UTC)
	in.InstallationFee = 0
	in.DiscountType = billing.DiscountTypeNone
	in.DiscountValue = 0
	in.TaxPercent = 0

	inv, items := billing.BuildFirstInvoice(in)
	assert.Equal(t, 30, inv.ProratedDays)
	assert.Equal(t, 30, inv.PeriodDays)
	assert.InDelta(t, 100000, inv.Subtotal, 0.005)
	assert.InDelta(t, 0, inv.Discount, 0.005)
	assert.InDelta(t, 100000, inv.Total, 0.005)
	require.Len(t, items, 1, "tanpa biaya pasang hanya satu item")
	require.NotNil(t, inv.PeriodEnd)
	assert.Equal(t, date(2026, time.September, 30), *inv.PeriodEnd)
}

func TestBuildFirstInvoice_AnchorAfterInstallDay(t *testing.T) {
	t.Parallel()
	// Anchor 15, pasang 5 Agustus → periode 5..14 Agustus (10/31 hari).
	in := baseFirstInvoiceInput(time.Date(2026, time.August, 5, 9, 0, 0, 0, time.UTC))
	in.InstallDate = time.Date(2026, time.August, 5, 9, 0, 0, 0, time.UTC)
	in.AnchorDay = 15
	in.InstallationFee = 0
	in.DiscountValue = 0

	inv, _ := billing.BuildFirstInvoice(in)
	assert.Equal(t, 10, inv.ProratedDays)
	assert.Equal(t, 31, inv.PeriodDays)
	require.NotNil(t, inv.PeriodEnd)
	assert.Equal(t, date(2026, time.August, 14), *inv.PeriodEnd)
	assert.InDelta(t, 32258.06, inv.Subtotal, 0.005)
}

func TestBuildFirstInvoice_DefaultTenantAndDueDays(t *testing.T) {
	t.Parallel()
	in := baseFirstInvoiceInput(time.Date(2026, time.August, 20, 10, 0, 0, 0, time.UTC))
	in.TenantID = ""
	in.FirstDueDays = 0
	inv, _ := billing.BuildFirstInvoice(in)
	assert.Equal(t, "tenant-default", inv.TenantID)
	assert.Equal(t, time.Date(2026, time.August, 20, 23, 59, 59, 0, time.UTC), inv.DueDate)
}
