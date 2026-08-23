package billing

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	domainNotification "github.com/quixiq/polyglot/internal/domain/notification"
	domainSubscription "github.com/quixiq/polyglot/internal/domain/subscription"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/idgen"
	"github.com/quixiq/polyglot/pkg/logger"
)

// IsolationResult rekap satu siklus worker lifecycle ISP.
type IsolationResult struct {
	Isolated        int
	Suspended       int
	Restored        int // disiapkan untuk hook pasca-lunas (fase berikut)
	Provisioned     int // retry provisioning sukses
	ProvisionFailed int
	SkippedNoRouter int
	RouterFailures  int
}

// IsolateWorker menjalankan lifecycle otomatis berbasis system_settings:
//   - retry provisi akun ke router (provision_status PENDING/FAILED)
//   - isolir otomatis bila isp.auto_isolate aktif dan tagihan lewat
//     jatuh tempo + grace
//   - auto-suspend ISOLATED lama bila isp.suspend_after_days > 0
//
// Tidak ada konstanta perilaku di kode — semua dari settings.
type IsolateWorker struct {
	subs      port.SubscriptionRepository
	invoices  port.InvoiceRepository
	customers port.CustomerRepository
	plans     port.ServicePlanRepository
	manager   port.RouterAccountManager
	notif     port.NotificationRepository
	settings  port.SettingReader

	now func() time.Time
}

// NewIsolateWorker wires dependencies.
func NewIsolateWorker(
	subs port.SubscriptionRepository,
	invoices port.InvoiceRepository,
	customers port.CustomerRepository,
	plans port.ServicePlanRepository,
	manager port.RouterAccountManager,
	notif port.NotificationRepository,
	settings port.SettingReader,
) *IsolateWorker {
	return &IsolateWorker{
		subs: subs, invoices: invoices, customers: customers,
		plans: plans, manager: manager, notif: notif, settings: settings,
		now: time.Now,
	}
}

// Run executes one lifecycle pass over all ACTIVE subscriptions.
func (w *IsolateWorker) Run(ctx context.Context) (IsolationResult, error) {
	res := IsolationResult{}
	cfg := port.LoadISPSettings(ctx, w.settings)

	lifecycle, err := w.subs.ListLifecycle(ctx)
	if err != nil {
		return res, err
	}
	now := w.now()

	for _, sub := range lifecycle {
		if err := w.retryProvisioning(ctx, sub, cfg, &res); err != nil {
			return res, err
		}
		if err := w.processIsolation(ctx, sub, cfg, now, &res); err != nil {
			return res, err
		}
	}
	return res, nil
}

// retryProvisioning mencoba push akun ke router untuk langganan yang belum
// tersinkron (PENDING/FAILED/NONE tapi sudah punya device).
func (w *IsolateWorker) retryProvisioning(ctx context.Context, sub domainSubscription.Subscription, cfg port.ISPSettings, res *IsolationResult) error {
	if sub.DeviceID == nil || *sub.DeviceID == "" || sub.ProvisionStatus == domainSubscription.ProvisionOK {
		return nil
	}
	if sub.RemoteUsername == "" || sub.RemotePassword == "" {
		res.SkippedNoRouter++
		return nil
	}
	acct := port.SubscriberAccount{
		Username:  sub.RemoteUsername,
		Password:  sub.RemotePassword,
		Profile:   w.normalProfile(ctx, sub),
		RateLimit: w.planRateLimit(ctx, sub),
		Comment:   "polyglot:" + sub.ID,
	}
	if err := w.manager.Provision(ctx, *sub.DeviceID, sub.ServiceType, acct); err != nil {
		res.ProvisionFailed++
		logger.WithComponent("IsolateWorker").WithFields(map[string]any{
			"subscription_id": sub.ID,
		}).WithError(err).Warn("provisioning failed")
		sub.ProvisionStatus = domainSubscription.ProvisionFailed
	} else {
		res.Provisioned++
		sub.ProvisionStatus = domainSubscription.ProvisionOK
		sub.RouterProfile = acct.Profile
	}
	return w.subs.Save(ctx, sub)
}

// processIsolation handles isolate + auto-suspend transitions.
func (w *IsolateWorker) processIsolation(ctx context.Context, sub domainSubscription.Subscription, cfg port.ISPSettings, now time.Time, res *IsolationResult) error {
	switch sub.Status {
	case domainSubscription.StatusIsolated:
		return w.maybeAutoSuspend(ctx, sub, cfg, now, res)
	case domainSubscription.StatusActive:
	default:
		return nil // SUSPENDED/PENDING dsb. bukan ranah worker
	}
	if !cfg.AutoIsolate {
		return nil // isolir otomatis dimatikan via settings
	}

	graceCutoff := now.AddDate(0, 0, -cfg.IsolateGraceDays)
	unpaid, found := w.overdueInvoice(ctx, sub, graceCutoff)
	if !found {
		return nil
	}
	if sub.DeviceID == nil || *sub.DeviceID == "" {
		res.SkippedNoRouter++
		return nil
	}

	opt := port.IsolationOptions{
		AddressList: cfg.IsolirAddressList,
		Redirect:    redirectConfig(cfg),
	}
	opt.IsolirProfile = cfg.HotspotIsolirProfile
	if !isHotspotService(sub.ServiceType) {
		opt.IsolirProfile = cfg.PPPoEIsolirProfile
	}

	if err := w.manager.Isolate(ctx, *sub.DeviceID, sub.ServiceType, sub.RemoteUsername, opt); err != nil {
		res.RouterFailures++ // dicoba ulang siklus berikutnya
		logger.WithComponent("IsolateWorker").WithFields(map[string]any{
			"subscription_id": sub.ID,
		}).WithError(err).Warn("router isolate failed")
		return nil
	}
	if err := w.subs.UpdateStatus(ctx, sub.ID, domainSubscription.StatusIsolated); err != nil {
		return err
	}
	w.queueNotice(ctx, sub, unpaid)
	res.Isolated++
	return nil
}

// maybeAutoSuspend menonaktifkan akun ISOLATED yang sudah terlalu lama —
// hanya bila isp.suspend_after_days > 0 (selain itu manual oleh admin).
func (w *IsolateWorker) maybeAutoSuspend(ctx context.Context, sub domainSubscription.Subscription, cfg port.ISPSettings, now time.Time, res *IsolationResult) error {
	if cfg.SuspendAfterDays <= 0 {
		return nil
	}
	suspendCutoff := now.AddDate(0, 0, -(cfg.IsolateGraceDays + cfg.SuspendAfterDays))
	if _, found := w.overdueInvoice(ctx, sub, suspendCutoff); !found {
		return nil
	}
	if sub.DeviceID == nil || *sub.DeviceID == "" {
		return nil
	}
	if err := w.manager.Suspend(ctx, *sub.DeviceID, sub.ServiceType, sub.RemoteUsername); err != nil {
		res.RouterFailures++
		return nil
	}
	if err := w.subs.UpdateStatus(ctx, sub.ID, domainSubscription.StatusSuspended); err != nil {
		return err
	}
	res.Suspended++
	return nil
}

// overdueInvoice returns the first UNPAID invoice of the subscription whose
// due date passed before cutoff.
func (w *IsolateWorker) overdueInvoice(ctx context.Context, sub domainSubscription.Subscription, cutoff time.Time) (domainBilling.Invoice, bool) {
	invoices, err := w.invoices.FindByCustomerID(ctx, sub.CustomerID)
	if err != nil {
		return domainBilling.Invoice{}, false
	}
	for _, inv := range invoices {
		if inv.SubscriptionID == nil || *inv.SubscriptionID != sub.ID {
			continue
		}
		if inv.Status == domainBilling.StatusUnpaid && inv.DueDate.Before(cutoff) {
			return inv, true
		}
	}
	return domainBilling.Invoice{}, false
}

func (w *IsolateWorker) queueNotice(ctx context.Context, sub domainSubscription.Subscription, inv domainBilling.Invoice) {
	if w.notif == nil {
		return
	}
	cust, err := w.customers.FindByID(ctx, sub.CustomerID)
	if err != nil {
		return
	}
	content := renderTemplateOrDefault(ctx, w.notif, cust.TenantID, "ISOLATION_NOTICE",
		"Yth {{customer_name}}, layanan Anda dibatasi karena tagihan melewati jatuh tempo. Silakan lakukan pembayaran untuk pemulihan otomatis.",
		map[string]string{"customer_name": cust.Name})
	n := domainNotification.WANotification{
		ID:             idgen.New("wa"),
		TenantID:       cust.TenantID,
		CustomerID:     &cust.ID,
		InvoiceID:      &inv.ID,
		RecipientPhone: cust.Phone,
		MessageType:    "ISOLATION_NOTICE",
		MessageContent: content,
		Status:         domainNotification.StatusQueued,
		CreatedAt:      time.Now(),
	}
	if err := w.notif.Queue(ctx, n); err != nil {
		logger.WithComponent("IsolateWorker").WithError(err).Warn("queue isolation notice failed")
	}
}

// normalProfile adalah nama profil paket di router: nilai tersimpan atau
// nama paket itu sendiri.
func (w *IsolateWorker) normalProfile(ctx context.Context, sub domainSubscription.Subscription) string {
	if sub.RouterProfile != "" {
		return sub.RouterProfile
	}
	if pl, err := w.plans.FindByID(ctx, sub.PlanID); err == nil && pl.Name != "" {
		return pl.Name
	}
	return sub.PlanID
}

func (w *IsolateWorker) planRateLimit(ctx context.Context, sub domainSubscription.Subscription) string {
	if sub.RateLimit != "" {
		return sub.RateLimit
	}
	pl, err := w.plans.FindByID(ctx, sub.PlanID)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%s/%s",
		kbpsToRate(pl.BandwidthDownloadKbps),
		kbpsToRate(pl.BandwidthUploadKbps))
}

// redirectConfig membangun konfigurasi dst-nat dari settings; URL kosong
// berarti tanpa rule global (isolir cukup profil 0/0).
func redirectConfig(cfg port.ISPSettings) *port.IsolationRedirectConfig {
	host, pport, ok := splitHostPort(cfg.PaymentRedirectURL)
	if !ok || host == "" {
		return nil
	}
	portNum := pport
	if portNum == "" {
		portNum = "80"
	}
	return &port.IsolationRedirectConfig{
		SrcAddressList: cfg.IsolirAddressList,
		PaymentHost:    host,
		PaymentPort:    portNum,
		Protocols:      []string{"tcp"},
		DstPorts:       []string{"80", "443"},
		Disabled:       false,
	}
}

func splitHostPort(s string) (host, portStr string, ok bool) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "http://")
	s = strings.TrimPrefix(s, "https://")
	if i := strings.LastIndex(s, ":"); i >= 0 {
		if _, err := strconv.Atoi(strings.TrimSpace(s[i+1:])); err == nil {
			return s[:i], strings.TrimSpace(s[i+1:]), true
		}
	}
	if s == "" {
		return "", "", false
	}
	return s, "", true
}

func isHotspotService(serviceType string) bool {
	return strings.EqualFold(serviceType, "HOTSPOT")
}

// kbpsToRate converts kbps → RouterOS rate token ("5M", "512k").
func kbpsToRate(kbps int) string {
	if kbps <= 0 {
		return ""
	}
	if kbps >= 1000 {
		return strconv.Itoa((kbps+500)/1000) + "M"
	}
	return strconv.Itoa(kbps) + "k"
}
