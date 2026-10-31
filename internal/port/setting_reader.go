package port

import (
	"context"
	"strconv"
	"strings"

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
		PPPoEIsolirProfile:   r.GetValue(ctx, "isp.pppoe_isolir_profile", "isolir"),
		HotspotIsolirProfile: r.GetValue(ctx, "isp.hotspot_isolir_profile", "isolir"),
		PaymentRedirectURL:   r.GetValue(ctx, "isp.payment_redirect_url", "192.168.233.195:5176"),
		SuspendAfterDays:     atoiOrDefault(r.GetValue(ctx, "isp.suspend_after_days", "90"), 90),
		IsolirAddressList:    r.GetValue(ctx, "isp.isolir_address_list", "ISOLIR_USERS"),

		IsolirWebProxyEnabled: r.GetValue(ctx, "isp.isolir_webproxy_enabled", "true") == "true",
		IsolirWebProxyPort:    atoiOrDefault(r.GetValue(ctx, "isp.isolir_webproxy_port", "3125"), 3125),
		IsolirBillingHosts:    splitCSV(r.GetValue(ctx, "isp.isolir_billing_hosts", "")),
		IsolirErrorPageURL:    strings.TrimSpace(r.GetValue(ctx, "isp.isolir_error_page_url", "")),

		BillingMode:                 strings.ToUpper(strings.TrimSpace(r.GetValue(ctx, "isp.billing_mode", "PASCABAYAR"))),
		ClosingDate:                 clampInt(atoiOrDefault(r.GetValue(ctx, "isp.closing_date", "1"), 1), 1, 28),
		UnpaidDaysDisplay:           clampInt(atoiOrDefault(r.GetValue(ctx, "isp.unpaid_days_display", "7"), 7), 1, 90),
		PrepayInvoiceLeadDays:       clampInt(atoiOrDefault(r.GetValue(ctx, "isp.prepay_invoice_lead_days", "7"), 7), 0, 30),
		TaxPercent:                  atofOrDefault(r.GetValue(ctx, "isp.tax_percent", "0"), 0),
		ReminderPreOffsets:          splitCSVInts(r.GetValue(ctx, "isp.reminder_pre_offsets", "7,3,1,0")),
		ReminderPostOffsets:         splitCSVInts(r.GetValue(ctx, "isp.reminder_post_offsets", "1,3,7")),
		GatewaySyncStaleMinutes:     clampInt(atoiOrDefault(r.GetValue(ctx, "isp.gateway_sync_stale_minutes", "15"), 15), 1, 1440),
		DebtIncomeCategoryID:        strings.TrimSpace(r.GetValue(ctx, "isp.debt_income_category_id", "cc-debt-income")),
		DebtReminderEnabled:         r.GetValue(ctx, "isp.debt_reminder_enabled", "true") == "true",
		DebtReminderDays:            splitCSVInts(r.GetValue(ctx, "isp.debt_reminder_days", "3,1,0")),
		GatewayFeeExpenseCategoryID: strings.TrimSpace(r.GetValue(ctx, "isp.gateway_fee_expense_category_id", "cc-gateway-fee")),
		MassPaymentEnabled:          r.GetValue(ctx, "isp.mass_payment_enabled", "true") == "true",
	}
}

func splitCSV(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if v := strings.TrimSpace(p); v != "" {
			out = append(out, v)
		}
	}
	return out
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func atofOrDefault(s string, def float64) float64 {
	v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return def
	}
	return v
}

func splitCSVInts(s string) []int {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]int, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if v, err := strconv.Atoi(p); err == nil {
			out = append(out, v)
		}
	}
	return out
}

func atoiOrDefault(s string, def int) int {
	v, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return v
}
