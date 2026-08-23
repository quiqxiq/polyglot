package port

import "context"

// SettingReader membaca konfigurasi dinamis dari system_settings.
// Interface kecil agar usecase/worker dapat dites tanpa DB.
type SettingReader interface {
	GetValue(ctx context.Context, key, fallback string) string
}

// ISPSettings adalah hasil pembacaan konfigurasi billing/isolir dari
// system_settings (kategori isp_*). Semua nilai punya default aman bila key
// belum ada — tidak ada konfigurasi statis di kode.
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

// LoadISPSettings membaca seluruh key isp.* dengan fallback default.
func LoadISPSettings(ctx context.Context, r SettingReader) ISPSettings {
	return ISPSettings{
		BillingCycleStartDay: atoiOr(r.GetValue(ctx, "isp.billing_cycle_start_day", "1"), 1),
		BillingDueDays:       atoiOr(r.GetValue(ctx, "isp.billing_due_days", "20"), 20),
		AutoIsolate:          r.GetValue(ctx, "isp.auto_isolate", "true") == "true",
		IsolateGraceDays:     atoiOr(r.GetValue(ctx, "isp.isolate_grace_days", "3"), 3),
		PPPoEIsolirProfile:   r.GetValue(ctx, "isp.pppoe_isolir_profile", "isolir"),
		HotspotIsolirProfile: r.GetValue(ctx, "isp.hotspot_isolir_profile", "isolir"),
		PaymentRedirectURL:   r.GetValue(ctx, "isp.payment_redirect_url", ""),
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
