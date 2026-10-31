package ont

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/quixiq/polyglot/internal/domain/notification"
	"github.com/quixiq/polyglot/internal/domain/ont"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/idgen"
	"github.com/quixiq/polyglot/pkg/logger"
	"github.com/quixiq/polyglot/pkg/phone"
)

// AlertEvaluator mengimplementasikan hook OLT (state & Rx ONT) untuk
// notifikasi WA transisi — PLAN-ACS-OLT-HIOSO.md D14.
type AlertEvaluator struct {
	onts      port.ONTRepository
	customers port.CustomerRepository
	notif     port.NotificationRepository
	settings  port.SettingReader
}

// NewAlertEvaluator membangun evaluator alert ONT.
func NewAlertEvaluator(
	onts port.ONTRepository,
	customers port.CustomerRepository,
	notif port.NotificationRepository,
	settings port.SettingReader,
) *AlertEvaluator {
	return &AlertEvaluator{onts: onts, customers: customers, notif: notif, settings: settings}
}

// ONTStateChanged mengirim notifikasi offline/pulih pada transisi state.
func (a *AlertEvaluator) ONTStateChanged(ctx context.Context, before, after ont.ONT) {
	o := a.fullONT(ctx, after)
	if o.CustomerID == nil || !a.customerNotifyEnabled(ctx) {
		return
	}
	switch after.State {
	case ont.StateOffline:
		a.notify(ctx, o, "ONT_OFFLINE",
			"Halo {{customer_name}}, koneksi internet di alamat {{address}} terdeteksi terputus (LOS/offline). Tim teknis kami sedang memeriksa.",
			map[string]string{"customer_name": a.customerName(ctx, o), "address": a.customerAddress(ctx, o)})
	case ont.StateOnline:
		if before.State == ont.StateOffline || before.State == ont.StateUnknown {
			a.notify(ctx, o, "ONT_ONLINE",
				"Halo {{customer_name}}, koneksi internet Anda sudah kembali normal.",
				map[string]string{"customer_name": a.customerName(ctx, o)})
		}
	}
}

// ONTRxChanged mengirim notifikasi saat kualitas Rx optik menurun.
func (a *AlertEvaluator) ONTRxChanged(ctx context.Context, before, after ont.ONT) {
	if after.RxPowerDBM == nil {
		return
	}
	o := a.fullONT(ctx, after)
	if o.CustomerID == nil || !a.customerNotifyEnabled(ctx) {
		return
	}
	warning := a.floatSetting(ctx, "olt.rx_warning_dbm", -27)
	critical := a.floatSetting(ctx, "olt.rx_critical_dbm", -30)
	beforeClass := ont.ClassifyRx(before.RxPowerDBM, warning, critical)
	afterClass := ont.ClassifyRx(after.RxPowerDBM, warning, critical)
	if beforeClass == afterClass || afterClass == ont.RxNormal || afterClass == ont.RxUnavailable {
		return
	}
	a.notify(ctx, o, "ONT_LOW_RX",
		"Halo {{customer_name}}, kualitas sinyal optik perangkat Anda menurun ({{rx_power}} dBm). Tim teknis akan melakukan pengecekan.",
		map[string]string{
			"customer_name": a.customerName(ctx, o),
			"rx_power":      fmt.Sprintf("%.1f", *after.RxPowerDBM),
		})
}

func (a *AlertEvaluator) fullONT(ctx context.Context, fallback ont.ONT) ont.ONT {
	if o, err := a.onts.FindByID(ctx, fallback.ID); err == nil {
		return o
	}
	return fallback
}

func (a *AlertEvaluator) customerNotifyEnabled(ctx context.Context) bool {
	return a.settings.GetValue(ctx, "alert.ont_customer_notify", "true") == "true"
}

func (a *AlertEvaluator) customerName(ctx context.Context, o ont.ONT) string {
	if o.CustomerID == nil {
		return "Pelanggan"
	}
	c, err := a.customers.FindByID(ctx, *o.CustomerID)
	if err != nil {
		return "Pelanggan"
	}
	return c.Name
}

func (a *AlertEvaluator) customerAddress(ctx context.Context, o ont.ONT) string {
	if o.CustomerID == nil {
		return "-"
	}
	c, err := a.customers.FindByID(ctx, *o.CustomerID)
	if err != nil {
		return "-"
	}
	return c.Address
}

// coolingDown mencegah notifikasi berulang dalam jendela cooldown; event
// alert dicatat di timeline agar dedupe tahan restart.
func (a *AlertEvaluator) coolingDown(ctx context.Context, ontID, messageType string) bool {
	cooldownMin := 60
	if v, err := strconv.Atoi(a.settings.GetValue(ctx, "alert.ont_cooldown_minutes", "60")); err == nil && v > 0 {
		cooldownMin = v
	}
	events, err := a.onts.ListEvents(ctx, ontID, 100)
	if err != nil {
		return false
	}
	cutoff := time.Now().UTC().Add(-time.Duration(cooldownMin) * time.Minute)
	for _, e := range events {
		if e.EventType != "ALERT_"+messageType {
			continue
		}
		if e.CreatedAt.After(cutoff) {
			return true
		}
	}
	return false
}

func (a *AlertEvaluator) notify(ctx context.Context, o ont.ONT, messageType, fallback string, vars map[string]string) {
	if a.coolingDown(ctx, o.ID, messageType) {
		return
	}
	if o.CustomerID == nil {
		return
	}
	cust, err := a.customers.FindByID(ctx, *o.CustomerID)
	if err != nil || cust.Phone == "" {
		return
	}
	content := a.renderTemplate(ctx, cust.TenantID, messageType, fallback, vars)
	n := notification.WANotification{
		ID:             idgen.New("wa"),
		TenantID:       cust.TenantID,
		CustomerID:     &cust.ID,
		RecipientPhone: phone.Normalize(cust.Phone),
		MessageType:    messageType,
		MessageContent: content,
		Status:         notification.StatusQueued,
		CreatedAt:      time.Now().UTC(),
	}
	if err := a.notif.Queue(ctx, n); err != nil {
		logger.WithComponent("ONTAlert").WithError(err).Warn("queue ont alert failed")
		return
	}
	_ = a.onts.RecordEvent(ctx, ont.Event{
		ID:        idgen.New("ontev"),
		TenantID:  cust.TenantID,
		ONTID:     o.ID,
		EventType: "ALERT_" + messageType,
		Message:   "notifikasi " + messageType + " dikirim",
		CreatedAt: time.Now().UTC(),
	})
}

func (a *AlertEvaluator) renderTemplate(ctx context.Context, tenantID, key, fallback string, vars map[string]string) string {
	tpl, err := a.notif.FindTemplateByKey(ctx, tenantID, key)
	if err != nil {
		return fallback
	}
	rep := make([]string, 0, len(vars)*2)
	for k, v := range vars {
		rep = append(rep, "{{"+k+"}}", v)
	}
	return strings.NewReplacer(rep...).Replace(tpl.Content)
}

func (a *AlertEvaluator) floatSetting(ctx context.Context, key string, fallback float64) float64 {
	raw := a.settings.GetValue(ctx, key, "")
	if raw == "" {
		return fallback
	}
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return fallback
	}
	return v
}
