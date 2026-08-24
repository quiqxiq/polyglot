package billing

import (
	"strconv"
	"strings"

	domainPlan "github.com/quixiq/polyglot/internal/domain/plan"
	domainSubscription "github.com/quixiq/polyglot/internal/domain/subscription"
	"github.com/quixiq/polyglot/internal/port"
)

// subscriberAccountFromPlan adalah SATU-SATUNYA titik mapping
// ServicePlan + Subscription -> port.SubscriberAccount.
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
		Username:      sub.RemoteUsername,
		Password:      sub.RemotePassword,
		Profile:       pl.Name,
		RateLimit:     rate,
		AddressPool:   pl.IPPoolName,
		ParentQueue:   pl.ParentQueue,
		AddressList:   pl.AddressList,
		SharedUsers:   pl.SharedUsers,
		Price:         hargaJual,
		SellingPrice:  hargaModal,
		Validity:      validity,
		ExpireMode:    pl.ExpireMode,
		LockUser:      pl.LockUser,
		LockServer:    pl.LockServer,
		BaseRateLimit: pl.RateLimit(),
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
