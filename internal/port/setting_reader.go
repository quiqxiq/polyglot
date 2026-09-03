package port

import (
	"context"
	"strconv"

	"github.com/quixiq/polyglot/internal/domain/setting"
)

// SettingReader membaca konfigurasi dinamis dari system_settings.
// Interface kecil agar usecase/worker dapat dites tanpa DB.
type SettingReader interface {
	GetValue(ctx context.Context, key, fallback string) string
}

// ISPSettings adalah alias ke domain model per DEVELOPMENT-GUIDELINES.md §4.2.
type ISPSettings = setting.ISPSettings

// LoadISPSettings membaca seluruh key isp.* dengan fallback default.
func LoadISPSettings(ctx context.Context, r SettingReader) ISPSettings {
	return ISPSettings{
		BillingCycleStartDay: atoiOrDefault(r.GetValue(ctx, "isp.billing_cycle_start_day", "1"), 1),
		BillingDueDays:       atoiOrDefault(r.GetValue(ctx, "isp.billing_due_days", "20"), 20),
		AutoIsolate:          r.GetValue(ctx, "isp.auto_isolate", "true") == "true",
		IsolateGraceDays:     atoiOrDefault(r.GetValue(ctx, "isp.isolate_grace_days", "3"), 3),
		PPPoEIsolirProfile:   r.GetValue(ctx, "isp.pppoe_isolir_profile", "ISOLIR"),
		HotspotIsolirProfile: r.GetValue(ctx, "isp.hotspot_isolir_profile", "ISOLIR"),
		PaymentRedirectURL:   r.GetValue(ctx, "isp.payment_redirect_url", "192.168.233.195:5176"),
		SuspendAfterDays:     atoiOrDefault(r.GetValue(ctx, "isp.suspend_after_days", "90"), 90),
		IsolirAddressList:    r.GetValue(ctx, "isp.isolir_address_list", "ISOLIR_USERS"),
	}
}

func atoiOrDefault(s string, def int) int {
	v, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return v
}
