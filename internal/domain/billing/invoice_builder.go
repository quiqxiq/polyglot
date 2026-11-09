package billing

import (
	"fmt"
	"time"

	domainTenant "github.com/quixiq/polyglot/internal/domain/tenant"
)

// FirstInvoiceInput is the input for building the first installation invoice
// (planning v2 D3): biaya prorata paket + biaya pasang − diskon + PPN.
type FirstInvoiceInput struct {
	InvoiceID         string
	InvoiceNumber     string
	ManualPaymentCode string
	ItemIDFee         string
	ItemIDInstall     string
	TenantID          string
	SubscriptionID    string
	PlanName          string
	PlanPrice         float64
	InstallationFee   float64
	TaxPercent        float64
	ProrateMode       ProrateMode
	DiscountType      DiscountType
	DiscountValue     float64
	InstallDate       time.Time // tgl pasang (bagian tanggal saja)
	AnchorDay         int       // hari anchor siklus (mis. 1)
	FirstDueDays      int       // isp.first_invoice_due_days
	Now               time.Time
}

// BuildFirstInvoice menyusun faktur pasang baru beserta baris item-nya.
// Periode: tgl pasang s/d H-1 anchor berikutnya; bila tgl pasang tepat di
// hari anchor, periode = bulan berjalan penuh. Due date = tanggal Now +
// FirstDueDays hari kalender (23:59:59 UTC).
func BuildFirstInvoice(in FirstInvoiceInput) (Invoice, []InvoiceItem) {
	anchor := clampAnchor(in.AnchorDay)
	start, end := firstCycleWindow(dateOnly(in.InstallDate), anchor)
	prorated := CalculateProrate(ProrateParams{
		PlanPrice: in.PlanPrice,
		StartDate: start,
		EndDate:   end,
		Mode:      in.ProrateMode,
	})

	subtotal := round2(prorated.Amount + in.InstallationFee)
	discount := ApplyDiscount(subtotal, in.DiscountType, in.DiscountValue)
	tax := round2((subtotal - discount) * in.TaxPercent / 100)
	total := round2(subtotal - discount + tax)

	tenant := in.TenantID
	if tenant == "" {
		tenant = domainTenant.DefaultID
	}
	discountType := in.DiscountType
	if discountType == "" {
		discountType = DiscountTypeNone
	}
	due := dateOnly(in.Now).AddDate(0, 0, maxInt(in.FirstDueDays, 0)).
		Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	invID := in.InvoiceID
	subID := in.SubscriptionID
	startDate, endDate := start, end
	inv := Invoice{
		ID:                invID,
		TenantID:          tenant,
		InvoiceNumber:     in.InvoiceNumber,
		SubscriptionID:    &subID,
		Period:            in.Now.Format("2006-01"),
		InvoiceType:       InvoiceTypeFirstInstallation,
		PeriodStart:       &startDate,
		PeriodEnd:         &endDate,
		ProratedDays:      prorated.ActiveDays,
		PeriodDays:        prorated.PeriodDays,
		Subtotal:          subtotal,
		Discount:          discount,
		DiscountType:      string(discountType),
		DiscountValue:     in.DiscountValue,
		DiscountAmount:    discount,
		TaxAmount:         tax,
		Total:             total,
		DueDate:           due,
		Status:            StatusUnpaid,
		QRPayload:         "polyglot://invoice/" + invID,
		ManualPaymentCode: in.ManualPaymentCode,
		IssuedAt:          in.Now,
		CreatedAt:         in.Now,
		UpdatedAt:         in.Now,
	}

	items := []InvoiceItem{{
		ID:          in.ItemIDFee,
		InvoiceID:   invID,
		Description: fmt.Sprintf("Paket %s (prorata %d/%d hari)", in.PlanName, prorated.ActiveDays, prorated.PeriodDays),
		Quantity:    1,
		UnitPrice:   prorated.Amount,
		Amount:      prorated.Amount,
		ItemType:    ItemTypeSubscriptionFee,
		CreatedAt:   in.Now,
	}}
	if in.InstallationFee > 0 {
		items = append(items, InvoiceItem{
			ID:          in.ItemIDInstall,
			InvoiceID:   invID,
			Description: "Biaya pemasangan awal",
			Quantity:    1,
			UnitPrice:   in.InstallationFee,
			Amount:      in.InstallationFee,
			ItemType:    ItemTypeInstallationFee,
			CreatedAt:   in.Now,
		})
	}
	return inv, items
}

// firstCycleWindow mengembalikan periode tagihan pertama: dari tanggal pasang
// s/d H-1 anchor berikutnya. Bila tanggal pasang tepat di hari anchor, periode
// adalah bulan berjalan penuh.
func firstCycleWindow(install time.Time, anchor int) (time.Time, time.Time) {
	if install.Day() == anchor {
		last := time.Date(install.Year(), install.Month()+1, 0, 0, 0, 0, 0, time.UTC)
		return install, last
	}
	nextAnchor := nextAnchorAfter(install, anchor)
	return install, nextAnchor.AddDate(0, 0, -1)
}

// nextAnchorAfter mengembalikan tanggal anchor pertama yang >= t (dijepit ke
// hari terakhir bulan pendek).
func nextAnchorAfter(t time.Time, anchor int) time.Time {
	day := anchor
	last := daysInMonth(t)
	if day > last {
		day = last
	}
	d := time.Date(t.Year(), t.Month(), day, 0, 0, 0, 0, time.UTC)
	if d.Before(t) {
		nm := time.Date(t.Year(), t.Month()+1, 1, 0, 0, 0, 0, time.UTC)
		last = daysInMonth(nm)
		if anchor > last {
			day = last
		} else {
			day = anchor
		}
		d = time.Date(nm.Year(), nm.Month(), day, 0, 0, 0, 0, time.UTC)
	}
	return d
}

// clampAnchor menjepit hari anchor ke rentang 1..28 agar aman di bulan pendek.
func clampAnchor(anchor int) int {
	if anchor <= 0 {
		return 1
	}
	if anchor > 28 {
		return 28
	}
	return anchor
}

// maxInt mengembalikan nilai maksimum dari a dan b.
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
