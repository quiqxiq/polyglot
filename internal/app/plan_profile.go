package app

import (
	"strconv"

	"github.com/quixiq/polyglot/internal/port"
)

// planProfileComment menandai profil yang dibuat otomatis dari service_plans.
const planProfileComment = "AUTO plan profile"

// hotspotExpireMode meneruskan mode expire konvensi Mikhmon
// (ntf|ntfc|rem|remc|0) apa adanya — script expire monitor Mikhmon
// membaca nilai ini dari profil.
func hotspotExpireMode(m string) port.ExpireMode {
	return port.ExpireMode(m)
}

// hotspotProfileParams membangun parameter lengkap Hotspot User Profile
// dari akun yang sudah dipetakan dari ServicePlan.
func hotspotProfileParams(a port.SubscriberAccount) port.MikhmonProfileParams {
	return port.MikhmonProfileParams{
		Name:         a.Profile,
		AddressPool:  a.AddressPool,
		SharedUsers:  intToStrOr(a.SharedUsers, "1"),
		RateLimit:    a.RateLimit,
		ParentQueue:  a.ParentQueue,
		Price:        a.Price,
		SellingPrice: a.SellingPrice,
		Validity:     a.Validity,
		ExpireMode:   hotspotExpireMode(a.ExpireMode),
		LockUser:     a.LockUser,
		LockServer:   a.LockServer,
		Comment:      planProfileComment,
	}
}

// pppProfileParams membangun parameter PPP profile dari akun terpetakan.
func pppProfileParams(a port.SubscriberAccount) port.PPPProfileParams {
	return port.PPPProfileParams{
		Name:        a.Profile,
		RateLimit:   a.RateLimit,
		ParentQueue: a.ParentQueue,
		AddressList: a.AddressList,
		Comment:     planProfileComment,
	}
}

// isolirAccount membuat akun dummy untuk auto-buat profil isolir (rate 0/0).
func isolirAccount(profileName, rateLimit string) port.SubscriberAccount {
	return port.SubscriberAccount{Profile: profileName, RateLimit: rateLimit}
}

func intToStrOr(v int, def string) string {
	if v <= 0 {
		return def
	}
	return strconv.Itoa(v)
}
