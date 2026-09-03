package setting

import "context"

// ValueGetter defines the minimal interface for fetching key-value configuration.
type ValueGetter interface {
	GetValue(ctx context.Context, key, fallback string) string
}

// ISPSettings holds business settings for ISP billing and automatic isolation.
type ISPSettings struct {
	BillingCycleStartDay int    // tanggal mulai periode tagihan (1-28)
	BillingDueDays       int    // hari kalender jatuh tempo setelah periode
	AutoIsolate          bool   // isolir otomatis aktif?
	IsolateGraceDays     int    // toleransi hari setelah jatuh tempo
	PPPoEIsolirProfile   string // profil isolir PPPoE di router
	HotspotIsolirProfile string // profil isolir hotspot di router
	PaymentRedirectURL   string // host:port halaman bayar tujuan dst-nat
	SuspendAfterDays     int    // ISOLATED → SUSPENDED otomatis (0 = off)
	IsolirAddressList    string // address-list penanda pelanggan terisolir
}

// LoadISPSettings reads all isp.* keys with secure fallbacks.
func LoadISPSettings(ctx context.Context, r ValueGetter) ISPSettings {
	return ISPSettings{
		BillingCycleStartDay: atoiOr(r.GetValue(ctx, "isp.billing_cycle_start_day", "1"), 1),
		BillingDueDays:       atoiOr(r.GetValue(ctx, "isp.billing_due_days", "20"), 20),
		AutoIsolate:          r.GetValue(ctx, "isp.auto_isolate", "true") == "true",
		IsolateGraceDays:     atoiOr(r.GetValue(ctx, "isp.isolate_grace_days", "3"), 3),
		PPPoEIsolirProfile:   r.GetValue(ctx, "isp.pppoe_isolir_profile", "ISOLIR"),
		HotspotIsolirProfile: r.GetValue(ctx, "isp.hotspot_isolir_profile", "ISOLIR"),
		PaymentRedirectURL:   r.GetValue(ctx, "isp.payment_redirect_url", "192.168.233.195:5176"),
		SuspendAfterDays:     atoiOr(r.GetValue(ctx, "isp.suspend_after_days", "90"), 90),
		IsolirAddressList:    r.GetValue(ctx, "isp.isolir_address_list", "ISOLIR_USERS"),
	}
}

func atoiOr(s string, def int) int {
	n := 0
	neg := false
	for i, c := range s {
		if i == 0 && c == '-' {
			neg = true
			continue
		}
		if c < '0' || c > '9' {
			return def
		}
		n = n*10 + int(c-'0')
	}
	if neg {
		n = -n
	}
	if n == 0 && s != "0" {
		return def
	}
	return n
}
