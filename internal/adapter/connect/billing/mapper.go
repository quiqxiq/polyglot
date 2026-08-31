package billing

import (
	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	domainPlan "github.com/quixiq/polyglot/internal/domain/plan"
	domainSub "github.com/quixiq/polyglot/internal/domain/subscription"
	billingUC "github.com/quixiq/polyglot/internal/usecase/billing"
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
	pb := &devicepb.Subscription{
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

	if sub.ServiceType == domainPlan.TypeHotspot {
		cfg := &devicepb.HotspotSubscriptionConfig{
			Server:        "all",
			RateLimit:     sub.RateLimit,
			RouterProfile: sub.RouterProfile,
		}
		if sub.Hotspot != nil {
			if sub.Hotspot.Server != "" {
				cfg.Server = sub.Hotspot.Server
			}
			cfg.MacAddress = sub.Hotspot.MacAddress
			cfg.IpAddress = sub.Hotspot.IPAddress
			cfg.LimitUptime = sub.Hotspot.LimitUptime
			cfg.LimitBytes = sub.Hotspot.LimitBytes
			if sub.Hotspot.RateLimit != "" {
				cfg.RateLimit = sub.Hotspot.RateLimit
			}
			if sub.Hotspot.RouterProfile != "" {
				cfg.RouterProfile = sub.Hotspot.RouterProfile
			}
		}
		pb.HotspotConfig = cfg
	} else {
		cfg := &devicepb.PPPoESubscriptionConfig{
			LocalAddress:  sub.LocalAddress,
			RemoteAddress: sub.RemoteAddress,
			RateLimit:     sub.RateLimit,
			RouterProfile: sub.RouterProfile,
		}
		if sub.PPPoE != nil {
			if sub.PPPoE.LocalAddress != "" {
				cfg.LocalAddress = sub.PPPoE.LocalAddress
			}
			if sub.PPPoE.RemoteAddress != "" {
				cfg.RemoteAddress = sub.PPPoE.RemoteAddress
			}
			cfg.CallerId = sub.PPPoE.CallerID
			cfg.Routes = sub.PPPoE.Routes
			if sub.PPPoE.RateLimit != "" {
				cfg.RateLimit = sub.PPPoE.RateLimit
			}
			if sub.PPPoE.RouterProfile != "" {
				cfg.RouterProfile = sub.PPPoE.RouterProfile
			}
		}
		pb.PppoeConfig = cfg
	}
	return pb
}

func toProtoSubscriptionList(subs []domainSub.Subscription) []*devicepb.Subscription {
	out := make([]*devicepb.Subscription, len(subs))
	for i := range subs {
		out[i] = toProtoSubscription(&subs[i])
	}
	return out
}

func toProtoSubscriptionDetail(detail *billingUC.SubscriptionDetail) *devicepb.Subscription {
	if detail == nil {
		return nil
	}
	pb := toProtoSubscription(&detail.Subscription)
	if pb == nil {
		return nil
	}
	if detail.Plan != nil {
		pb.PlanName = detail.Plan.Name
		pb.PlanPrice = detail.Plan.Price
		rateLimit := detail.Plan.RateLimitWithBurst()
		if rateLimit == "" {
			rateLimit = detail.Plan.RateLimit()
		}
		if pb.RateLimit == "" {
			pb.RateLimit = rateLimit
		}
		if pb.RouterProfile == "" {
			pb.RouterProfile = detail.Plan.Name
		}
		if pb.PppoeConfig != nil {
			if pb.PppoeConfig.RateLimit == "" {
				pb.PppoeConfig.RateLimit = rateLimit
			}
			if pb.PppoeConfig.RouterProfile == "" {
				pb.PppoeConfig.RouterProfile = detail.Plan.Name
			}
		}
		if pb.HotspotConfig != nil {
			if pb.HotspotConfig.RateLimit == "" {
				pb.HotspotConfig.RateLimit = rateLimit
			}
			if pb.HotspotConfig.RouterProfile == "" {
				pb.HotspotConfig.RouterProfile = detail.Plan.Name
			}
		}
	}
	if detail.Customer != nil {
		pb.CustomerName = detail.Customer.Name
		pb.CustomerPhone = detail.Customer.Phone
		pb.CustomerCode = detail.Customer.CustomerCode
	}
	if detail.Device != nil {
		pb.DeviceName = detail.Device.Name
		pb.DeviceHost = detail.Device.Host
	}
	return pb
}

func toProtoSubscriptionDetailList(details []billingUC.SubscriptionDetail) []*devicepb.Subscription {
	out := make([]*devicepb.Subscription, len(details))
	for i := range details {
		out[i] = toProtoSubscriptionDetail(&details[i])
	}
	return out
}

// ─── Plan ───────────────────────────────────────────────────────────────

func toProtoPlan(p *domainPlan.ServicePlan) *devicepb.Plan {
	if p == nil {
		return nil
	}
	sharedUsers := int32(p.SharedUsers)
	if sharedUsers <= 0 {
		sharedUsers = 1
	}
	pb := &devicepb.Plan{
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
		InstallationFee:       p.InstallationFee,
		TaxPercent:            p.TaxPercent,
		IpPoolName:            p.IPPoolName,
		RemoteAddressPool:     p.RemoteAddressPool,
		ParentQueue:           p.ParentQueue,
		AddressList:           p.AddressList,
		SharedUsers:           sharedUsers,
		IsActive:              p.IsActive,
		Description:           p.Description,
		SessionTimeout:        p.SessionTimeout,
		IdleTimeout:           p.IdleTimeout,
	}

	if p.ServiceType == domainPlan.TypeHotspot {
		cfg := &devicepb.HotspotPlanConfig{
			IpPoolName:     p.IPPoolName,
			AddressList:    p.AddressList,
			SharedUsers:    sharedUsers,
			SessionTimeout: p.SessionTimeout,
			IdleTimeout:    p.IdleTimeout,
		}
		if p.Hotspot != nil {
			if p.Hotspot.IPPoolName != "" {
				cfg.IpPoolName = p.Hotspot.IPPoolName
			}
			if p.Hotspot.AddressList != "" {
				cfg.AddressList = p.Hotspot.AddressList
			}
			if p.Hotspot.SharedUsers > 0 {
				cfg.SharedUsers = int32(p.Hotspot.SharedUsers)
			}
			if p.Hotspot.SessionTimeout != "" {
				cfg.SessionTimeout = p.Hotspot.SessionTimeout
			}
			if p.Hotspot.IdleTimeout != "" {
				cfg.IdleTimeout = p.Hotspot.IdleTimeout
			}
		}
		pb.HotspotConfig = cfg
	} else {
		cfg := &devicepb.PPPoEPlanConfig{
			RemoteAddressPool: p.RemoteAddressPool,
			AddressList:       p.AddressList,
			SessionTimeout:    p.SessionTimeout,
			IdleTimeout:       p.IdleTimeout,
		}
		if p.PPPoE != nil {
			if p.PPPoE.RemoteAddressPool != "" {
				cfg.RemoteAddressPool = p.PPPoE.RemoteAddressPool
			}
			if p.PPPoE.AddressList != "" {
				cfg.AddressList = p.PPPoE.AddressList
			}
			if p.PPPoE.SessionTimeout != "" {
				cfg.SessionTimeout = p.PPPoE.SessionTimeout
			}
			if p.PPPoE.IdleTimeout != "" {
				cfg.IdleTimeout = p.PPPoE.IdleTimeout
			}
		}
		pb.PppoeConfig = cfg
	}
	return pb
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
	sharedUsers := int(pb.SharedUsers)
	if sharedUsers <= 0 {
		sharedUsers = 1
	}
	p := domainPlan.ServicePlan{
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
		InstallationFee:       pb.InstallationFee,
		TaxPercent:            pb.TaxPercent,
		IPPoolName:            pb.IpPoolName,
		RemoteAddressPool:     pb.RemoteAddressPool,
		ParentQueue:           pb.ParentQueue,
		AddressList:           pb.AddressList,
		SharedUsers:           sharedUsers,
		IsActive:              pb.IsActive,
		Description:           pb.Description,
		SessionTimeout:        pb.SessionTimeout,
		IdleTimeout:           pb.IdleTimeout,
	}

	if pb.PppoeConfig != nil {
		p.PPPoE = &domainPlan.PPPoEPlanConfig{
			RemoteAddressPool: pb.PppoeConfig.RemoteAddressPool,
			AddressList:       pb.PppoeConfig.AddressList,
			SessionTimeout:    pb.PppoeConfig.SessionTimeout,
			IdleTimeout:       pb.PppoeConfig.IdleTimeout,
		}
		if pb.PppoeConfig.RemoteAddressPool != "" {
			p.RemoteAddressPool = pb.PppoeConfig.RemoteAddressPool
		}
		if pb.PppoeConfig.AddressList != "" {
			p.AddressList = pb.PppoeConfig.AddressList
		}
		if pb.PppoeConfig.SessionTimeout != "" {
			p.SessionTimeout = pb.PppoeConfig.SessionTimeout
		}
		if pb.PppoeConfig.IdleTimeout != "" {
			p.IdleTimeout = pb.PppoeConfig.IdleTimeout
		}
	}
	if pb.HotspotConfig != nil {
		hsSharedUsers := int(pb.HotspotConfig.SharedUsers)
		if hsSharedUsers <= 0 {
			hsSharedUsers = sharedUsers
		}
		p.Hotspot = &domainPlan.HotspotPlanConfig{
			IPPoolName:     pb.HotspotConfig.IpPoolName,
			AddressList:    pb.HotspotConfig.AddressList,
			SharedUsers:    hsSharedUsers,
			SessionTimeout: pb.HotspotConfig.SessionTimeout,
			IdleTimeout:    pb.HotspotConfig.IdleTimeout,
		}
		if pb.HotspotConfig.IpPoolName != "" {
			p.IPPoolName = pb.HotspotConfig.IpPoolName
		}
		if pb.HotspotConfig.AddressList != "" {
			p.AddressList = pb.HotspotConfig.AddressList
		}
		if pb.HotspotConfig.SharedUsers > 0 {
			p.SharedUsers = int(pb.HotspotConfig.SharedUsers)
		}
		if pb.HotspotConfig.SessionTimeout != "" {
			p.SessionTimeout = pb.HotspotConfig.SessionTimeout
		}
		if pb.HotspotConfig.IdleTimeout != "" {
			p.IdleTimeout = pb.HotspotConfig.IdleTimeout
		}
	}
	return p
}
