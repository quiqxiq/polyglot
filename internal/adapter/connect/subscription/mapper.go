package subscription

import (
	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	domainPlan "github.com/quixiq/polyglot/internal/domain/plan"
	domainSub "github.com/quixiq/polyglot/internal/domain/subscription"
	subUC "github.com/quixiq/polyglot/internal/usecase/subscription"
)

// ToProtoSubscription maps a domain Subscription model to a protobuf Subscription message.
func ToProtoSubscription(sub *domainSub.Subscription) *devicepb.Subscription {
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
		}
		pb.HotspotConfig = cfg
	} else {
		cfg := &devicepb.PPPoESubscriptionConfig{
			RateLimit:     sub.RateLimit,
			RouterProfile: sub.RouterProfile,
		}
		if sub.PPPoE != nil {
			cfg.LocalAddress = sub.PPPoE.LocalAddress
			cfg.RemoteAddress = sub.PPPoE.RemoteAddress
			cfg.CallerId = sub.PPPoE.CallerID
			cfg.Routes = sub.PPPoE.Routes
		}
		pb.PppoeConfig = cfg
	}
	return pb
}

// ToProtoSubscriptionDetail maps an enriched subscription detail to a protobuf Subscription message.
func ToProtoSubscriptionDetail(detail *subUC.Detail) *devicepb.Subscription {
	if detail == nil {
		return nil
	}
	pb := ToProtoSubscription(&detail.Subscription)
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

// ToProtoSubscriptionDetailList converts a slice of subscription details into a protobuf slice.
func ToProtoSubscriptionDetailList(details []subUC.Detail) []*devicepb.Subscription {
	out := make([]*devicepb.Subscription, len(details))
	for i := range details {
		out[i] = ToProtoSubscriptionDetail(&details[i])
	}
	return out
}
