package billing

import (
	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	domainSub "github.com/quixiq/polyglot/internal/domain/subscription"
)

// toProtoInvoice maps a domain Invoice to Protobuf Invoice message.
func toProtoInvoice(inv *domainBilling.Invoice) *devicepb.Invoice {
	if inv == nil {
		return nil
	}
	return &devicepb.Invoice{
		Id:            inv.ID,
		CustomerId:    inv.CustomerID,
		Amount:        inv.Total,
		Status:        inv.Status,
		DueDateUnix:   inv.DueDate.Unix(),
		CreatedAtUnix: inv.CreatedAt.Unix(),
	}
}

// toProtoInvoiceList maps a slice of domain Invoices to a slice of Protobuf Invoices.
func toProtoInvoiceList(invoices []domainBilling.Invoice) []*devicepb.Invoice {
	pbInvoices := make([]*devicepb.Invoice, len(invoices))
	for i := range invoices {
		pbInvoices[i] = toProtoInvoice(&invoices[i])
	}
	return pbInvoices
}

// toProtoSubscription maps a domain Subscription to Protobuf Subscription message.
func toProtoSubscription(sub *domainSub.Subscription) *devicepb.Subscription {
	if sub == nil {
		return nil
	}
	var endDateUnix int64
	if sub.EndDate != nil {
		endDateUnix = sub.EndDate.Unix()
	}
	var price float64
	if sub.CustomPrice != nil {
		price = *sub.CustomPrice
	}
	return &devicepb.Subscription{
		Id:            sub.ID,
		CustomerId:    sub.CustomerID,
		PlanId:        sub.PlanID,
		Status:        sub.Status,
		StartDateUnix: sub.StartDate.Unix(),
		EndDateUnix:   endDateUnix,
		Price:         price,
	}
}

// toProtoSubscriptionList maps a slice of domain Subscriptions to a slice of Protobuf Subscriptions.
func toProtoSubscriptionList(subs []domainSub.Subscription) []*devicepb.Subscription {
	pbSubs := make([]*devicepb.Subscription, len(subs))
	for i := range subs {
		pbSubs[i] = toProtoSubscription(&subs[i])
	}
	return pbSubs
}
