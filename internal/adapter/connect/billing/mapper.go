package billing

import (
	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	domainPlan "github.com/quixiq/polyglot/internal/domain/plan"
	domainSub "github.com/quixiq/polyglot/internal/domain/subscription"
)

// ─── Invoice ────────────────────────────────────────────────────────────

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

// ─── Subscription ───────────────────────────────────────────────────────

func toProtoSubscription(sub *domainSub.Subscription) *devicepb.Subscription {
	if sub == nil {
		return nil
	}
	var devID string
	if sub.DeviceID != nil {
		devID = *sub.DeviceID
	}
	var endUnix int64
	if sub.EndDate != nil {
		endUnix = sub.EndDate.Unix()
	}
	var customPrice float64
	if sub.CustomPrice != nil {
		customPrice = *sub.CustomPrice
	}
	return &devicepb.Subscription{
		Id:              sub.ID,
		TenantId:        sub.TenantID,
		CustomerId:      sub.CustomerID,
		PlanId:          sub.PlanID,
		DeviceId:        devID,
		ServiceType:     sub.ServiceType,
		RemoteUsername:  sub.RemoteUsername,
		RateLimit:       sub.RateLimit,
		RouterProfile:   sub.RouterProfile,
		ProvisionStatus: sub.ProvisionStatus,
		BillingCycle:    sub.BillingCycle,
		BillingDay:      int32(sub.BillingDay),
		Status:          sub.Status,
		StartDateUnix:   sub.StartDate.Unix(),
		EndDateUnix:     endUnix,
		CustomPrice:     customPrice,
		Notes:           sub.Notes,
	}
}

func toProtoSubscriptionList(subs []domainSub.Subscription) []*devicepb.Subscription {
	out := make([]*devicepb.Subscription, len(subs))
	for i := range subs {
		out[i] = toProtoSubscription(&subs[i])
	}
	return out
}

// ─── Plan ───────────────────────────────────────────────────────────────

func toProtoPlan(p *domainPlan.ServicePlan) *devicepb.Plan {
	if p == nil {
		return nil
	}
	return &devicepb.Plan{
		Id:                    p.ID,
		Name:                  p.Name,
		ServiceType:           p.ServiceType,
		BandwidthDownloadKbps: int32(p.BandwidthDownloadKbps),
		BandwidthUploadKbps:   int32(p.BandwidthUploadKbps),
		BurstDownloadKbps:     int32(p.BurstDownloadKbps),
		BurstUploadKbps:       int32(p.BurstUploadKbps),
		BurstThresholdKbps:    int32(p.BurstThresholdKbps),
		BurstTimeSeconds:      int32(p.BurstTimeSeconds),
		Price:                 p.Price,
		SellingPrice:          p.SellingPrice,
		InstallationFee:       p.InstallationFee,
		TaxPercent:            p.TaxPercent,
		Validity:              p.Validity,
		ValidityMode:          p.ValidityMode,
		SimultaneousUse:       int32(p.SimultaneousUse),
		IpPoolName:            p.IPPoolName,
		ParentQueue:           p.ParentQueue,
		AddressList:           p.AddressList,
		SharedUsers:           int32(p.SharedUsers),
		ExpireMode:            p.ExpireMode,
		LockUser:              p.LockUser,
		LockServer:            p.LockServer,
		IsActive:              p.IsActive,
		Description:           p.Description,
	}
}

func toProtoPlanList(plans []domainPlan.ServicePlan) []*devicepb.Plan {
	out := make([]*devicepb.Plan, len(plans))
	for i := range plans {
		out[i] = toProtoPlan(&plans[i])
	}
	return out
}

func fromProtoPlan(pb *devicepb.Plan) domainPlan.ServicePlan {
	if pb == nil {
		return domainPlan.ServicePlan{}
	}
	return domainPlan.ServicePlan{
		ID:                    pb.Id,
		Name:                  pb.Name,
		ServiceType:           pb.ServiceType,
		BandwidthDownloadKbps: int(pb.BandwidthDownloadKbps),
		BandwidthUploadKbps:   int(pb.BandwidthUploadKbps),
		BurstDownloadKbps:     int(pb.BurstDownloadKbps),
		BurstUploadKbps:       int(pb.BurstUploadKbps),
		BurstThresholdKbps:    int(pb.BurstThresholdKbps),
		BurstTimeSeconds:      int(pb.BurstTimeSeconds),
		Price:                 pb.Price,
		SellingPrice:          pb.SellingPrice,
		InstallationFee:       pb.InstallationFee,
		TaxPercent:            pb.TaxPercent,
		Validity:              pb.Validity,
		ValidityMode:          pb.ValidityMode,
		SimultaneousUse:       int(pb.SimultaneousUse),
		IPPoolName:            pb.IpPoolName,
		ParentQueue:           pb.ParentQueue,
		AddressList:           pb.AddressList,
		SharedUsers:           int(pb.SharedUsers),
		ExpireMode:            pb.ExpireMode,
		LockUser:              pb.LockUser,
		LockServer:            pb.LockServer,
		IsActive:              pb.IsActive,
		Description:           pb.Description,
	}
}
