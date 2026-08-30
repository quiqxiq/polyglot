package billing

import (
	"strconv"
	"strings"

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
	return domainSubscription.PPPoEProvisionSpec{
		Secret: domainSubscription.PPPoESecretSpec{
			Username:      sub.RemoteUsername,
			Password:      sub.RemotePassword,
			Profile:       pl.Name,
			Service:       "pppoe",
			LocalAddress:  sub.LocalAddress,
			RemoteAddress: sub.RemoteAddress,
			Comment:       "polyglot:" + sub.ID,
			Disabled:      sub.Status == domainSubscription.StatusSuspended,
		},
		Profile: domainSubscription.PPPoEProfileSpec{
			Name:              pl.Name,
			RateLimit:         rate,
			RemoteAddressPool: pl.RemoteAddressPool,
			ParentQueue:       pl.ParentQueue,
			AddressList:       pl.AddressList,
			Comment:           "AUTO plan profile",
		},
	}
}

// BuildHotspotProvisionSpec membangun spesifikasi terisolasi untuk provisi Hotspot.
func BuildHotspotProvisionSpec(sub domainSubscription.Subscription, pl domainPlan.ServicePlan) domainSubscription.HotspotProvisionSpec {
	rate := pl.RateLimitWithBurst()
	if sub.RateLimit != "" {
		rate = sub.RateLimit
	}
	hargaJual, hargaModal := formatMoney(pl.SellingPrice, pl.Price)
	validity := pl.Validity
	if strings.EqualFold(pl.ExpireMode, domainPlan.ExpireNone) {
		validity = ""
	}
	return domainSubscription.HotspotProvisionSpec{
		User: domainSubscription.HotspotUserSpec{
			Username:    sub.RemoteUsername,
			Password:    sub.RemotePassword,
			Profile:     pl.Name,
			Server:      "all",
			LimitUptime: pl.LimitUptime,
			LimitBytes:  pl.LimitBytes,
			Comment:     "polyglot:" + sub.ID,
			Disabled:    sub.Status == domainSubscription.StatusSuspended,
		},
		Profile: domainSubscription.HotspotProfileSpec{
			Name:         pl.Name,
			RateLimit:    rate,
			AddressPool:  pl.IPPoolName,
			SharedUsers:  pl.SharedUsers,
			ParentQueue:  pl.ParentQueue,
			Price:        hargaJual,
			SellingPrice: hargaModal,
			Validity:     validity,
			ExpireMode:   pl.ExpireMode,
			LockUser:     pl.LockUser,
			LockServer:   pl.LockServer,
			Comment:      "AUTO plan profile",
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

// subscriberAccountFromPlan adalah SATU-SATUNYA titik mapping
// ServicePlan + Subscription -> port.SubscriberAccount (kompatibilitas backward).
// Rate: custom sub.RateLimit menang atas nilai plan; burst dipakai penuh
// bila seluruh komponen burst terdefinisi.
func subscriberAccountFromPlan(
	sub domainSubscription.Subscription, pl domainPlan.ServicePlan,
) port.SubscriberAccount {
	rate := pl.RateLimitWithBurst()
	if sub.RateLimit != "" {
		rate = sub.RateLimit
	}
	hargaJual, hargaModal := formatMoney(pl.SellingPrice, pl.Price)
	validity := pl.Validity
	if strings.EqualFold(pl.ExpireMode, domainPlan.ExpireNone) {
		// Mode 0 = tanpa auto-expire: profil hotspot tidak membawa validity
		// agar script expire monitor Mikhmon tidak pernah menandai user.
		validity = ""
	}
	return port.SubscriberAccount{
		Username:          sub.RemoteUsername,
		Password:          sub.RemotePassword,
		Profile:           pl.Name,
		RateLimit:         rate,
		AddressPool:       pl.IPPoolName,
		ParentQueue:       pl.ParentQueue,
		AddressList:       pl.AddressList,
		SharedUsers:       pl.SharedUsers,
		Price:             hargaJual,
		SellingPrice:      hargaModal,
		Validity:          validity,
		ExpireMode:        pl.ExpireMode,
		LockUser:          pl.LockUser,
		LockServer:        pl.LockServer,
		BaseRateLimit:     pl.RateLimit(),
		RemoteAddressPool: pl.RemoteAddressPool,
	}
}

// formatMoney memetakan harga ke kolom profil hotspot mengikuti konvensi
// Mikhmon: Price = harga jual ke pelanggan (SellingPrice bila terisi, else
// harga dasar plan); SellingPrice = harga modal (harga dasar plan).
// Keduanya "" bila nol agar parameter diabaikan RouterOS.
func formatMoney(selling, base float64) (string, string) {
	jual := selling
	if jual <= 0 {
		jual = base
	}
	if jual <= 0 {
		return "", ""
	}
	return strconv.FormatFloat(jual, 'f', -1, 64),
		strconv.FormatFloat(base, 'f', -1, 64)
}
