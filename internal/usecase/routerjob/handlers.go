package routerjob

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	domainRouterJob "github.com/quixiq/polyglot/internal/domain/routerjob"
	domainSetting "github.com/quixiq/polyglot/internal/domain/setting"
	domainSubscription "github.com/quixiq/polyglot/internal/domain/subscription"
	"github.com/quixiq/polyglot/internal/port"
	oltUC "github.com/quixiq/polyglot/internal/usecase/olt"
	ontUC "github.com/quixiq/polyglot/internal/usecase/ont"
	planUC "github.com/quixiq/polyglot/internal/usecase/plan"
)

// OLTActions adalah subset usecase OLT yang dieksekusi worker.
type OLTActions interface {
	SyncONUs(ctx context.Context, oltDeviceID string) (oltUC.SyncResult, error)
	RenameONU(ctx context.Context, oltDeviceID string, ponIndex, onuIndex int, name string) error
	RebootONU(ctx context.Context, oltDeviceID string, ponIndex, onuIndex int) error
	SetONUAdmin(ctx context.Context, oltDeviceID string, ponIndex, onuIndex int, enabled bool) error
	DeleteONU(ctx context.Context, oltDeviceID string, ponIndex, onuIndex int) error
	SaveOLTConfig(ctx context.Context, oltDeviceID string) error
}

// ACSActions adalah subset usecase ONT untuk aksi TR-069.
type ACSActions interface {
	SyncDevices(ctx context.Context) (ontUC.SyncStats, error)
	RefreshTelemetry(ctx context.Context, ontID string) error
	RebootDevice(ctx context.Context, ontID string) error
	SetWiFi(ctx context.Context, ontID, ssid, password, band string) error
	PushPPPoE(ctx context.Context, ontID, username, password string) error
	BootstrapPreset(ctx context.Context) error
}

// MassPaymentAction adalah aksi pemrosesan batch pembayaran massal.
type MassPaymentAction interface {
	ProcessBatch(ctx context.Context, batchID string) error
}

// HandlerDeps adalah dependensi untuk mengeksekusi job router.
type HandlerDeps struct {
	Subs     port.SubscriptionRepository
	Plans    port.ServicePlanRepository
	Manager  port.RouterAccountManager
	Settings port.SettingReader
	OLT      OLTActions
	ACS      ACSActions
	MassPay  MassPaymentAction
}

// NewHandlers membangun map handler per tipe job.
func NewHandlers(deps HandlerDeps) map[domainRouterJob.Type]Handler {
	return map[domainRouterJob.Type]Handler{
		domainRouterJob.TypeProvisionPPPoE:      provisionHandler(deps),
		domainRouterJob.TypeProvisionHotspot:    provisionHandler(deps),
		domainRouterJob.TypeProvisionDedicated:  provisionHandler(deps),
		domainRouterJob.TypeIsolate:             isolateHandler(deps),
		domainRouterJob.TypeRestore:             restoreHandler(deps),
		domainRouterJob.TypeDeleteAccount:       deleteHandler(deps),
		domainRouterJob.TypeSyncProfile:         syncProfileHandler(deps),
		domainRouterJob.TypeSyncOLT:             syncOLTHandler(deps),
		domainRouterJob.TypeONTRename:           ontRenameHandler(deps),
		domainRouterJob.TypeONTReboot:           ontRebootHandler(deps),
		domainRouterJob.TypeONTSetAdmin:         ontSetAdminHandler(deps),
		domainRouterJob.TypeONTDelete:           ontDeleteHandler(deps),
		domainRouterJob.TypeOLTSaveConfig:       oltSaveConfigHandler(deps),
		domainRouterJob.TypeACSSyncMatch:        acsSyncMatchHandler(deps),
		domainRouterJob.TypeACSRefreshTelemetry: acsRefreshTelemetryHandler(deps),
		domainRouterJob.TypeACSReboot:           acsRebootHandler(deps),
		domainRouterJob.TypeACSSetWiFi:          acsSetWiFiHandler(deps),
		domainRouterJob.TypeACSPushPPPoE:        acsPushPPPoEHandler(deps),
		domainRouterJob.TypeACSBootstrapPreset:  acsBootstrapPresetHandler(deps),
		domainRouterJob.TypeONTBind:             ontBindHandler(deps),
		domainRouterJob.TypeMassPayment:         massPaymentHandler(deps),
	}
}

func provisionHandler(deps HandlerDeps) Handler {
	return func(ctx context.Context, job domainRouterJob.Job) error {
		subID, err := payloadSubscriptionID(job)
		if err != nil {
			return err
		}
		sub, err := deps.Subs.FindByID(ctx, subID)
		if err != nil {
			return fmt.Errorf("subscription %s: %w", subID, err)
		}
		pl, err := deps.Plans.FindByID(ctx, sub.PlanID)
		if err != nil {
			return fmt.Errorf("plan %s: %w", sub.PlanID, err)
		}
		var provErr error
		switch strings.ToLower(sub.ServiceType) {
		case "hotspot":
			provErr = deps.Manager.ProvisionHotspot(ctx, job.DeviceID, planUC.BuildHotspotProvisionSpec(sub, pl))
		case "dedicated":
			provErr = deps.Manager.ProvisionDedicated(ctx, job.DeviceID, planUC.BuildDedicatedProvisionSpec(sub, pl))
		default:
			provErr = deps.Manager.ProvisionPPPoE(ctx, job.DeviceID, planUC.BuildPPPoEProvisionSpec(sub, pl))
		}
		if provErr != nil {
			markProvisionFailed(ctx, deps, sub, job.Attempts, provErr)
			return provErr
		}
		sub.ProvisionStatus = domainSubscription.ProvisionOK
		sub.RouterProfile = pl.RouterProfile()
		sub.ProvisionAttempts = job.Attempts
		sub.LastProvisionError = ""
		return deps.Subs.Save(ctx, sub)
	}
}

func isolateHandler(deps HandlerDeps) Handler {
	return func(ctx context.Context, job domainRouterJob.Job) error {
		subID, err := payloadSubscriptionID(job)
		if err != nil {
			return err
		}
		sub, err := deps.Subs.FindByID(ctx, subID)
		if err != nil {
			return fmt.Errorf("subscription %s: %w", subID, err)
		}
		cfg := port.LoadISPSettings(ctx, deps.Settings)
		opt := port.IsolationOptions{
			AddressList: cfg.IsolirAddressList,
			Redirect:    redirectConfig(cfg),
		}
		opt.IsolirProfile = cfg.PPPoEIsolirProfile
		if strings.EqualFold(sub.ServiceType, "HOTSPOT") {
			opt.IsolirProfile = cfg.HotspotIsolirProfile
		}
		deviceID := job.DeviceID
		if deviceID == "" && sub.DeviceID != nil {
			deviceID = *sub.DeviceID
		}
		if err := deps.Manager.Isolate(ctx, deviceID, sub.ServiceType, sub.RemoteUsername, opt); err != nil {
			return err
		}
		return deps.Subs.UpdateStatus(ctx, sub.ID, domainSubscription.StatusIsolated)
	}
}

func restoreHandler(deps HandlerDeps) Handler {
	return func(ctx context.Context, job domainRouterJob.Job) error {
		subID, err := payloadSubscriptionID(job)
		if err != nil {
			return err
		}
		sub, err := deps.Subs.FindByID(ctx, subID)
		if err != nil {
			return fmt.Errorf("subscription %s: %w", subID, err)
		}
		cfg := port.LoadISPSettings(ctx, deps.Settings)
		profile := sub.RouterProfile
		if profile == "" {
			if pl, perr := deps.Plans.FindByID(ctx, sub.PlanID); perr == nil {
				profile = pl.RouterProfile()
			}
		}
		if profile == "" {
			return fmt.Errorf("normal profile unknown for subscription %s", sub.ID)
		}
		deviceID := job.DeviceID
		if deviceID == "" && sub.DeviceID != nil {
			deviceID = *sub.DeviceID
		}
		if err := deps.Manager.Restore(ctx, deviceID, sub.ServiceType, sub.RemoteUsername, profile, cfg.IsolirAddressList); err != nil {
			return err
		}
		return deps.Subs.UpdateStatus(ctx, sub.ID, domainSubscription.StatusActive)
	}
}

func deleteHandler(deps HandlerDeps) Handler {
	return func(ctx context.Context, job domainRouterJob.Job) error {
		subID, err := payloadSubscriptionID(job)
		if err != nil {
			return err
		}
		sub, err := deps.Subs.FindByID(ctx, subID)
		if err != nil {
			return fmt.Errorf("subscription %s: %w", subID, err)
		}
		deviceID := job.DeviceID
		if deviceID == "" && sub.DeviceID != nil {
			deviceID = *sub.DeviceID
		}
		if err := deps.Manager.Terminate(ctx, deviceID, sub.ServiceType, sub.RemoteUsername); err != nil {
			return err
		}
		return deps.Subs.UpdateStatus(ctx, sub.ID, domainSubscription.StatusTerminated)
	}
}

func syncProfileHandler(deps HandlerDeps) Handler {
	return func(ctx context.Context, job domainRouterJob.Job) error {
		var payload struct {
			PlanID string `json:"plan_id"`
		}
		if err := json.Unmarshal(job.Payload, &payload); err != nil || payload.PlanID == "" {
			return fmt.Errorf("%w: payload plan_id invalid", domainRouterJob.ErrInvalidInput)
		}
		pl, err := deps.Plans.FindByID(ctx, payload.PlanID)
		if err != nil {
			return fmt.Errorf("plan %s: %w", payload.PlanID, err)
		}
		return deps.Manager.SyncPlanProfile(ctx, job.DeviceID, pl)
	}
}

func markProvisionFailed(ctx context.Context, deps HandlerDeps, sub domainSubscription.Subscription, attempts int, cause error) {
	sub.ProvisionStatus = domainSubscription.ProvisionFailed
	sub.ProvisionAttempts = attempts
	sub.LastProvisionError = truncateError(cause)
	if err := deps.Subs.Save(ctx, sub); err != nil {
		_ = err // best-effort; job retry akan mencatat ulang
	}
}

func payloadSubscriptionID(job domainRouterJob.Job) (string, error) {
	var payload struct {
		SubscriptionID string `json:"subscription_id"`
	}
	if err := json.Unmarshal(job.Payload, &payload); err != nil {
		return "", fmt.Errorf("%w: payload invalid: %w", domainRouterJob.ErrInvalidInput, err)
	}
	if payload.SubscriptionID == "" {
		return "", fmt.Errorf("%w: payload subscription_id required", domainRouterJob.ErrInvalidInput)
	}
	return payload.SubscriptionID, nil
}

// redirectConfig menurunkan konfigurasi dst-nat isolir dari setting
// isp.payment_redirect_url (format host:port).
func redirectConfig(cfg domainSetting.ISPSettings) *port.IsolationRedirectConfig {
	host, portStr := splitHostPort(cfg.PaymentRedirectURL)
	if host == "" {
		return nil
	}
	if portStr == "" {
		portStr = "80"
	}
	return &port.IsolationRedirectConfig{
		SrcAddressList: cfg.IsolirAddressList,
		PaymentHost:    host,
		PaymentPort:    portStr,
		Protocols:      []string{"tcp"},
		DstPorts:       []string{"80", "443"},
	}
}

func splitHostPort(s string) (host, portStr string) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", ""
	}
	if i := strings.LastIndex(s, ":"); i > 0 && !strings.Contains(s[i+1:], "/") {
		return s[:i], s[i+1:]
	}
	return s, ""
}
