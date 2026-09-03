package portal

import (
	"time"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	domainCustomer "github.com/quixiq/polyglot/internal/domain/customer"
	portalUC "github.com/quixiq/polyglot/internal/usecase/portal"
)

func toProtoCustomerPortal(c *domainCustomer.Customer) *devicepb.Customer {
	if c == nil {
		return nil
	}
	return &devicepb.Customer{
		Id:           c.ID,
		TenantId:     c.TenantID,
		CustomerCode: c.CustomerCode,
		Name:         c.Name,
		Phone:        c.Phone,
		Email:        c.Email,
		Address:      c.Address,
		Status:       c.Status,
	}
}

func toProtoSubscriptionSummary(sub *portalUC.SubscriptionView) *devicepb.PortalSubscriptionSummary {
	if sub == nil {
		return nil
	}
	var endUnix int64
	if sub.EndDate != nil {
		if t, err := time.Parse("2006-01-02", *sub.EndDate); err == nil {
			endUnix = t.Unix()
		}
	}
	return &devicepb.PortalSubscriptionSummary{
		Id:          sub.ID,
		PlanId:      sub.PlanID,
		ServiceType: sub.ServiceType,
		Status:      sub.Status,
		RateLimit:   sub.RateLimit,
		BillingDay:  int32(sub.BillingDay),
		EndDateUnix: endUnix,
	}
}

func toProtoUnpaidInvoiceSummary(inv portalUC.InvoiceView) *devicepb.UnpaidInvoiceSummary {
	return &devicepb.UnpaidInvoiceSummary{
		Id:                inv.ID,
		InvoiceNumber:     inv.InvoiceNumber,
		Period:            inv.Period,
		Total:             inv.Total,
		PaidAmount:        inv.PaidAmount,
		Outstanding:       inv.Outstanding,
		DueDate:           inv.DueDate,
		Status:            inv.Status,
		ManualPaymentCode: inv.ManualPaymentCode,
	}
}

func toProtoInvoicePortalList(list []domainBilling.Invoice) []*devicepb.Invoice {
	out := make([]*devicepb.Invoice, len(list))
	for i, inv := range list {
		out[i] = &devicepb.Invoice{
			Id:                inv.ID,
			InvoiceNumber:     inv.InvoiceNumber,
			Period:            inv.Period,
			Total:             inv.Total,
			PaidAmount:        inv.PaidAmount,
			Status:            inv.Status,
			DueDateUnix:       inv.DueDate.Unix(),
			ManualPaymentCode: inv.ManualPaymentCode,
		}
	}
	return out
}

func toProtoPaymentEntries(payments []domainBilling.Payment) []*devicepb.PaymentEntry {
	entries := make([]*devicepb.PaymentEntry, len(payments))
	for i, p := range payments {
		entries[i] = &devicepb.PaymentEntry{
			Id:              p.ID,
			PaymentNo:       p.PaymentNo,
			Amount:          p.Amount,
			PaymentDateUnix: p.PaymentDate.Unix(),
			ScanMethod:      p.ScanMethod,
			Reference:       p.Reference,
		}
	}
	return entries
}
