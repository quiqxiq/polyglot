package plan

import (
	domainPlan "github.com/quixiq/polyglot/internal/domain/plan"
	domainSubscription "github.com/quixiq/polyglot/internal/domain/subscription"
	"github.com/quixiq/polyglot/internal/port"
)

// BuildPPPoEProvisionSpec membangun spesifikasi terisolasi untuk provisi PPPoE.
func BuildPPPoEProvisionSpec(sub domainSubscription.Subscription, pl domainPlan.ServicePlan) domainSubscription.PPPoEProvisionSpec {
	rate := pl.RateLimitWithBurst()
	if sub.RateLimit != "" {
		rate = sub.RateLimit
	}
	remotePool := pl.RemoteAddressPool
	if pl.PPPoE != nil && pl.PPPoE.RemoteAddressPool != "" {
		remotePool = pl.PPPoE.RemoteAddressPool
	}
	addrList := pl.AddressList
	if pl.PPPoE != nil && pl.PPPoE.AddressList != "" {
		addrList = pl.PPPoE.AddressList
	}
	sessionTimeout := pl.SessionTimeout
	if pl.PPPoE != nil && pl.PPPoE.SessionTimeout != "" {
		sessionTimeout = pl.PPPoE.SessionTimeout
	}
	idleTimeout := pl.IdleTimeout
	if pl.PPPoE != nil && pl.PPPoE.IdleTimeout != "" {
		idleTimeout = pl.PPPoE.IdleTimeout
	}

	localAddr := sub.LocalAddress
	remoteAddr := sub.RemoteAddress
	callerID := ""
	routes := ""
	if sub.PPPoE != nil {
		if sub.PPPoE.LocalAddress != "" {
			localAddr = sub.PPPoE.LocalAddress
		}
		if sub.PPPoE.RemoteAddress != "" {
			remoteAddr = sub.PPPoE.RemoteAddress
		}
		callerID = sub.PPPoE.CallerID
		routes = sub.PPPoE.Routes
	}

	return domainSubscription.PPPoEProvisionSpec{
		Secret: domainSubscription.PPPoESecretSpec{
			Username:      sub.RemoteUsername,
			Password:      sub.RemotePassword,
			Profile:       pl.Name,
			Service:       "pppoe",
			LocalAddress:  localAddr,
			RemoteAddress: remoteAddr,
			CallerID:      callerID,
			Routes:        routes,
			Comment:       "polyglot:" + sub.ID,
			Disabled:      sub.Status == domainSubscription.StatusSuspended,
		},
		Profile: domainSubscription.PPPoEProfileSpec{
			Name:              pl.Name,
			RateLimit:         rate,
			RemoteAddressPool: remotePool,
			ParentQueue:       pl.ParentQueue,
			AddressList:       addrList,
			SessionTimeout:    sessionTimeout,
			IdleTimeout:       idleTimeout,
			Comment:           "AUTO plan profile",
		},
	}
}

// BuildHotspotProvisionSpec membangun spesifikasi terisolasi untuk provisi Hotspot Permanent.
func BuildHotspotProvisionSpec(sub domainSubscription.Subscription, pl domainPlan.ServicePlan) domainSubscription.HotspotProvisionSpec {
	rate := pl.RateLimitWithBurst()
	if sub.RateLimit != "" {
		rate = sub.RateLimit
	}
	ipPool := pl.IPPoolName
	if pl.Hotspot != nil && pl.Hotspot.IPPoolName != "" {
		ipPool = pl.Hotspot.IPPoolName
	}
	addrList := pl.AddressList
	if pl.Hotspot != nil && pl.Hotspot.AddressList != "" {
		addrList = pl.Hotspot.AddressList
	}
	sharedUsers := 1
	if pl.Hotspot != nil && pl.Hotspot.SharedUsers > 0 {
		sharedUsers = pl.Hotspot.SharedUsers
	} else if pl.SharedUsers > 0 {
		sharedUsers = pl.SharedUsers
	}
	sessionTimeout := pl.SessionTimeout
	if pl.Hotspot != nil && pl.Hotspot.SessionTimeout != "" {
		sessionTimeout = pl.Hotspot.SessionTimeout
	}
	idleTimeout := pl.IdleTimeout
	if pl.Hotspot != nil && pl.Hotspot.IdleTimeout != "" {
		idleTimeout = pl.Hotspot.IdleTimeout
	}

	server := "all"
	macAddr := ""
	ipAddr := ""
	limitUptime := ""
	limitBytes := ""
	if sub.Hotspot != nil {
		if sub.Hotspot.Server != "" {
			server = sub.Hotspot.Server
		}
		macAddr = sub.Hotspot.MacAddress
		ipAddr = sub.Hotspot.IPAddress
		limitUptime = sub.Hotspot.LimitUptime
		limitBytes = sub.Hotspot.LimitBytes
	}

	return domainSubscription.HotspotProvisionSpec{
		User: domainSubscription.HotspotUserSpec{
			Username:    sub.RemoteUsername,
			Password:    sub.RemotePassword,
			Profile:     pl.Name,
			Server:      server,
			MacAddress:  macAddr,
			IPAddress:   ipAddr,
			LimitUptime: limitUptime,
			LimitBytes:  limitBytes,
			Comment:     "polyglot:" + sub.ID,
			Disabled:    sub.Status == domainSubscription.StatusSuspended,
		},
		Profile: domainSubscription.HotspotProfileSpec{
			Name:           pl.Name,
			RateLimit:      rate,
			AddressPool:    ipPool,
			AddressList:    addrList,
			SharedUsers:    sharedUsers,
			ParentQueue:    pl.ParentQueue,
			SessionTimeout: sessionTimeout,
			IdleTimeout:    idleTimeout,
			Comment:        "AUTO plan profile",
		},
	}
}

// BuildDedicatedProvisionSpec membangun spesifikasi untuk sambungan Dedicated CIR/MIR.
func BuildDedicatedProvisionSpec(sub domainSubscription.Subscription, pl domainPlan.ServicePlan) domainSubscription.DedicatedProvisionSpec {
	pppoeSpec := BuildPPPoEProvisionSpec(sub, pl)
	target := sub.RemoteAddress
	if target == "" {
		target = sub.RemoteUsername
	}
	return domainSubscription.DedicatedProvisionSpec{
		PPPoE: pppoeSpec,
		Queue: domainSubscription.DedicatedQueueSpec{
			QueueName:   sub.RemoteUsername,
			Target:      target,
			MaxLimit:    pl.RateLimitWithBurst(),
			LimitAt:     pl.RateLimit(),
			Priority:    "1/1",
			ParentQueue: pl.ParentQueue,
			Comment:     "AUTO dedicated queue polyglot:" + sub.ID,
		},
	}
}

// SubscriberAccountFromPlan adalah titik mapping ServicePlan + Subscription -> port.SubscriberAccount.
func SubscriberAccountFromPlan(
	sub domainSubscription.Subscription, pl domainPlan.ServicePlan,
) port.SubscriberAccount {
	rate := pl.RateLimitWithBurst()
	if sub.RateLimit != "" {
		rate = sub.RateLimit
	}
	ipPool := pl.IPPoolName
	if pl.Hotspot != nil && pl.Hotspot.IPPoolName != "" {
		ipPool = pl.Hotspot.IPPoolName
	}
	remotePool := pl.RemoteAddressPool
	if pl.PPPoE != nil && pl.PPPoE.RemoteAddressPool != "" {
		remotePool = pl.PPPoE.RemoteAddressPool
	}
	addrList := pl.AddressList
	if pl.Hotspot != nil && pl.Hotspot.AddressList != "" {
		addrList = pl.Hotspot.AddressList
	} else if pl.PPPoE != nil && pl.PPPoE.AddressList != "" {
		addrList = pl.PPPoE.AddressList
	}
	sharedUsers := 1
	if pl.Hotspot != nil && pl.Hotspot.SharedUsers > 0 {
		sharedUsers = pl.Hotspot.SharedUsers
	} else if pl.SharedUsers > 0 {
		sharedUsers = pl.SharedUsers
	}

	return port.SubscriberAccount{
		Username:          sub.RemoteUsername,
		Password:          sub.RemotePassword,
		Profile:           pl.Name,
		RateLimit:         rate,
		AddressPool:       ipPool,
		ParentQueue:       pl.ParentQueue,
		AddressList:       addrList,
		SharedUsers:       sharedUsers,
		BaseRateLimit:     pl.RateLimit(),
		RemoteAddressPool: remotePool,
	}
}
