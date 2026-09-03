package plan

import (
	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	domainPlan "github.com/quixiq/polyglot/internal/domain/plan"
)

// ToProtoPlan maps a domain ServicePlan to a protobuf Plan message.
func ToProtoPlan(p *domainPlan.ServicePlan) *devicepb.Plan {
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

// ToProtoPlanList converts a slice of domain ServicePlans into a slice of protobuf Plans.
func ToProtoPlanList(plans []domainPlan.ServicePlan) []*devicepb.Plan {
	out := make([]*devicepb.Plan, len(plans))
	for i := range plans {
		out[i] = ToProtoPlan(&plans[i])
	}
	return out
}

// FromProtoPlan maps a protobuf Plan message to a domain ServicePlan.
func FromProtoPlan(pb *devicepb.Plan) domainPlan.ServicePlan {
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
