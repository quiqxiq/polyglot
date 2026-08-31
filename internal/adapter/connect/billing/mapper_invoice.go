package billing

import (
	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
)

func toProtoInvoice(inv *domainBilling.Invoice) *devicepb.Invoice {
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
	return &devicepb.Invoice{
		Id:                inv.ID,
		TenantId:          inv.TenantID,
		InvoiceNumber:     inv.InvoiceNumber,
		CustomerId:        inv.CustomerID,
		SubscriptionId:    subID,
		Period:            inv.Period,
		Subtotal:          inv.Subtotal,
		Discount:          inv.Discount,
		TaxAmount:         inv.TaxAmount,
		Total:             inv.Total,
		PaidAmount:        inv.PaidAmount,
		DueDateUnix:       inv.DueDate.Unix(),
		PaidAtUnix:        paidAtUnix,
		Status:            inv.Status,
		QrPayload:         inv.QRPayload,
		ManualPaymentCode: inv.ManualPaymentCode,
		Notes:             inv.Notes,
		CreatedAtUnix:     inv.CreatedAt.Unix(),
	}
}

func toProtoInvoiceList(invoices []domainBilling.Invoice) []*devicepb.Invoice {
	out := make([]*devicepb.Invoice, len(invoices))
	for i := range invoices {
		out[i] = toProtoInvoice(&invoices[i])
	}
	return out
}
