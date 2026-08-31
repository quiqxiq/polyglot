package billing

import (
	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	domainPlan "github.com/quixiq/polyglot/internal/domain/plan"
	domainSub "github.com/quixiq/polyglot/internal/domain/subscription"
	billingUC "github.com/quixiq/polyglot/internal/usecase/billing"
)

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
