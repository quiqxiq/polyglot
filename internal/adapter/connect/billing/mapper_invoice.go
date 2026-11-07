package billing

import (
	"time"

	polyglotv1 "github.com/quixiq/polyglot/api/gen/v1"
	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
)

func toProtoInvoice(inv *domainBilling.Invoice) *polyglotv1.Invoice {
	if inv == nil {
		return nil
	}
	var subID string
	if inv.SubscriptionID != nil {
		subID = *inv.SubscriptionID
	}
	var paidAtUnix int64
	if inv.PaidAt != nil {
		paidAtUnix = inv.PaidAt.Unix()
	}
	var periodStartUnix, periodEndUnix int64
	if inv.PeriodStart != nil {
		periodStartUnix = inv.PeriodStart.Unix()
	}
	if inv.PeriodEnd != nil {
		periodEndUnix = inv.PeriodEnd.Unix()
	}
	var issuedAtUnix int64
	if !inv.IssuedAt.IsZero() {
		issuedAtUnix = inv.IssuedAt.Unix()
	}
	var daysOverdue int32
	if inv.DueDate.Before(inv.IssuedAt) || (inv.Status != domainBilling.StatusPaid && inv.Status != domainBilling.StatusCancelled) {
		diff := time.Since(inv.DueDate)
		if diff > 0 {
			daysOverdue = int32(diff.Hours() / 24)
		}
	}
	return &polyglotv1.Invoice{
		Id:                inv.ID,
		TenantId:          inv.TenantID,
		InvoiceNumber:     inv.InvoiceNumber,
		CustomerId:        inv.CustomerID,
		SubscriptionId:    subID,
		Period:            inv.Period,
		Subtotal:          inv.Subtotal,
		Discount:          inv.Discount,
		TaxPercent:        inv.TaxPercent,
		TaxAmount:         inv.TaxAmount,
		Total:             inv.Total,
		PaidAmount:        inv.PaidAmount,
		Outstanding:       inv.Outstanding(),
		DaysOverdue:       daysOverdue,
		DueDateUnix:       inv.DueDate.Unix(),
		PaidAtUnix:        paidAtUnix,
		Status:            inv.Status,
		QrPayload:         inv.QRPayload,
		ManualPaymentCode: inv.ManualPaymentCode,
		Notes:             inv.Notes,
		CreatedAtUnix:     inv.CreatedAt.Unix(),
		InvoiceType:       inv.InvoiceType,
		PeriodStartUnix:   periodStartUnix,
		PeriodEndUnix:     periodEndUnix,
		ProratedDays:      int32(inv.ProratedDays),
		PeriodDays:        int32(inv.PeriodDays),
		DiscountType:      inv.DiscountType,
		DiscountValue:     inv.DiscountValue,
		DiscountAmount:    inv.DiscountAmount,
		PromoCode:         inv.PromoCode,
		IssuedAtUnix:      issuedAtUnix,
	}
}

func toProtoInvoiceList(invoices []domainBilling.Invoice) []*polyglotv1.Invoice {
	out := make([]*polyglotv1.Invoice, len(invoices))
	for i := range invoices {
		out[i] = toProtoInvoice(&invoices[i])
	}
	return out
}
