package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	billing "github.com/quixiq/polyglot/internal/domain/billing"
	domainSub "github.com/quixiq/polyglot/internal/domain/subscription"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/internal/registry"
	"github.com/quixiq/polyglot/pkg/logger"
)

// routerAccountManager implements port.RouterAccountManager dengan menyusun
// driver-resolver + PPPGateway + HotspotGateway + FirewallGateway + QueueGateway.
// Semua semantik mengikuti ISP nyata: isolir = pindah profil + kick +
// address-list redirect; suspend = disable; terminate = hapus.
type routerAccountManager struct {
	// resolve memetakan deviceID → DeviceDriver (di produksi: registry.Get;
	// di E2E/test: closure ke driver tetap).
	resolve func(ctx context.Context, deviceID string) (port.DeviceDriver, error)
	ppp     port.PPPGateway
	hot     port.HotspotGateway
	fw      port.FirewallGateway
	q       port.QueueGateway // khusus langganan DEDICATED (boleh nil)
}

var _ port.RouterAccountManager = (*routerAccountManager)(nil)

func newRouterAccountManager(reg *registry.Registry, ppp port.PPPGateway, hot port.HotspotGateway, fw port.FirewallGateway, q port.QueueGateway) *routerAccountManager {
	return &routerAccountManager{
		resolve: func(ctx context.Context, deviceID string) (port.DeviceDriver, error) {
			return reg.Get(ctx, deviceID)
		},
		ppp: ppp, hot: hot, fw: fw, q: q,
	}
}

const serviceHotspot = "HOTSPOT"

func isHotspot(serviceType string) bool { return strings.EqualFold(serviceType, serviceHotspot) }

// ─── Provision ──────────────────────────────────────────────────────────

func (m *routerAccountManager) ProvisionPPPoE(ctx context.Context, deviceID string, spec domainSub.PPPoEProvisionSpec) error {
	driver, err := m.resolve(ctx, deviceID)
	if err != nil {
		return fmt.Errorf("resolve driver %s: %w", deviceID, err)
	}
	if spec.Profile.Name != "" {
		if err := m.ensurePPPProfileFromSpec(ctx, driver, spec.Profile); err != nil {
			return fmt.Errorf("ensure ppp profile %s: %w", spec.Profile.Name, err)
		}
	}
	secParams := port.PPPoESecretParams{
		Name:          spec.Secret.Username,
		Password:      spec.Secret.Password,
		Profile:       spec.Secret.Profile,
		Service:       spec.Secret.Service,
		LocalAddress:  spec.Secret.LocalAddress,
		RemoteAddress: spec.Secret.RemoteAddress,
		Comment:       spec.Secret.Comment,
		Disabled:      spec.Secret.Disabled,
	}
	if secParams.Service == "" {
		secParams.Service = "pppoe"
	}
	if _, err := m.ppp.AddSecret(ctx, driver, secParams); err != nil {
		return fmt.Errorf("add ppp secret %s: %w", spec.Secret.Username, err)
	}
	return nil
}

func (m *routerAccountManager) ProvisionHotspot(ctx context.Context, deviceID string, spec domainSub.HotspotProvisionSpec) error {
	driver, err := m.resolve(ctx, deviceID)
	if err != nil {
		return fmt.Errorf("resolve driver %s: %w", deviceID, err)
	}
	if spec.Profile.Name != "" {
		if err := m.ensureHotspotProfileFromSpec(ctx, driver, spec.Profile); err != nil {
			return fmt.Errorf("ensure hotspot profile %s: %w", spec.Profile.Name, err)
		}
	}
	userParams := port.HotspotUserParams{
		Name:        spec.User.Username,
		Password:    spec.User.Password,
		Profile:     spec.User.Profile,
		Server:      spec.User.Server,
		LimitUptime: spec.User.LimitUptime,
		Comment:     spec.User.Comment,
		Disabled:    spec.User.Disabled,
	}
	if userParams.Server == "" {
		userParams.Server = "all"
	}
	if _, err := m.hot.AddUser(ctx, driver, userParams); err != nil {
		return fmt.Errorf("add hotspot user %s: %w", spec.User.Username, err)
	}
	return nil
}

func (m *routerAccountManager) ProvisionDedicated(ctx context.Context, deviceID string, spec domainSub.DedicatedProvisionSpec) error {
	if err := m.ProvisionPPPoE(ctx, deviceID, spec.PPPoE); err != nil {
		return err
	}
	driver, err := m.resolve(ctx, deviceID)
	if err != nil {
		return fmt.Errorf("resolve driver %s: %w", deviceID, err)
	}
	if m.q != nil && spec.Queue.QueueName != "" {
		qParams := port.DedicatedQueueParams{
			Name:     spec.Queue.QueueName,
			Target:   spec.Queue.Target,
			MaxLimit: spec.Queue.MaxLimit,
			LimitAt:  spec.Queue.LimitAt,
			Comment:  spec.Queue.Comment,
		}
		if _, err := m.q.AddQueue(ctx, driver, qParams); err != nil {
			return fmt.Errorf("add dedicated queue %s: %w", spec.Queue.QueueName, err)
		}
	}
	return nil
}

func (m *routerAccountManager) Provision(ctx context.Context, deviceID, serviceType string, acct port.SubscriberAccount) error {
	driver, err := m.resolve(ctx, deviceID)
	if err != nil {
		return fmt.Errorf("resolve driver %s: %w", deviceID, err)
	}
	if err := m.ensurePlanProfile(ctx, driver, serviceType, acct); err != nil {
		return err
	}
	if isHotspot(serviceType) {
		if _, err := m.hot.AddUser(ctx, driver, port.HotspotUserParams{
			Name: acct.Username, Password: acct.Password, Profile: acct.Profile,
			Comment: acct.Comment,
		}); err != nil {
			return fmt.Errorf("add hotspot user %s: %w", acct.Username, err)
		}
		return nil
	}
	if _, err := m.ppp.AddSecret(ctx, driver, port.PPPoESecretParams{
		Name: acct.Username, Password: acct.Password, Profile: acct.Profile,
		Service: "pppoe", Comment: acct.Comment,
	}); err != nil {
		return fmt.Errorf("add ppp secret %s: %w", acct.Username, err)
	}
	// DEDICATED: tegakkan CIR via simple queue per pelanggan.
	if isDedicated(serviceType) {
		if err := m.ensureDedicatedQueue(ctx, driver, acct); err != nil {
			return fmt.Errorf("dedicated queue: %w", err)
		}
	}
	return nil
}

// ensureDedicatedQueue memastikan simple queue CIR ada untuk pelanggan
// DEDICATED — idempotent: skip bila sudah sama, replace bila berubah.
func (m *routerAccountManager) ensureDedicatedQueue(ctx context.Context, driver port.DeviceDriver, acct port.SubscriberAccount) error {
	if m.q == nil || acct.Username == "" || acct.RateLimit == "" {
		return nil // queue gateway tidak tersedia / data kurang → abaikan
	}
	params := dedicatedQueueFromAccount(acct, "AUTO dedicated queue "+acct.Comment)
	name := params.Name

	existing, err := m.q.ListQueues(ctx, driver, name)
	if err != nil {
		return fmt.Errorf("list queues: %w", err)
	}
	for _, qy := range existing {
		if qy.Name != name {
			continue
		}
		if qy.MaxLimit == params.MaxLimit && qy.LimitAt == params.LimitAt {
			return nil // sudah sinkron
		}
		// Konfigurasi berubah → replace.
		if _, err := m.q.RemoveQueue(ctx, driver, qy.RosID); err != nil {
			return fmt.Errorf("remove stale queue %s: %w", name, err)
		}
		break
	}
	if _, err := m.q.AddQueue(ctx, driver, params); err != nil {
		return fmt.Errorf("add queue %s: %w", name, err)
	}
	return nil
}

// setDedicatedQueueEnabled enable/disable queue pelanggan DEDICATED — best-effort.
func (m *routerAccountManager) setDedicatedQueueEnabled(ctx context.Context, driver port.DeviceDriver, username string, enabled bool) {
	m.withDedicatedQueue(ctx, driver, username, func(qy port.SimpleQueue) error {
		_, err := m.q.SetQueueEnabled(ctx, driver, qy.RosID, enabled)
		return err
	})
}

// removeDedicatedQueue menghapus queue pelanggan DEDICATED — best-effort.
func (m *routerAccountManager) removeDedicatedQueue(ctx context.Context, driver port.DeviceDriver, username string) {
	m.withDedicatedQueue(ctx, driver, username, func(qy port.SimpleQueue) error {
		_, err := m.q.RemoveQueue(ctx, driver, qy.RosID)
		return err
	})
}

// withDedicatedQueue menjalankan aksi pada queue pelanggan bila ditemukan.
// Best-effort: error hanya di-log, tidak dikembalikan (queue hanya shaping,
// bukan auth — kegagalan tidak boleh menggagalkan isolate/terminate).
func (m *routerAccountManager) withDedicatedQueue(ctx context.Context, driver port.DeviceDriver, username string, action func(port.SimpleQueue) error) {
	if m.q == nil || username == "" {
		return
	}
	queues, err := m.q.ListQueues(ctx, driver, dedicatedQueueName(username))
	if err != nil {
		logger.WithComponent("AccountManager").WithFields(map[string]any{
			"username": username,
		}).WithError(err).Warn("list dedicated queue failed")
		return
	}
	for _, qy := range queues {
		if qy.Name == dedicatedQueueName(username) {
			if err := action(qy); err != nil {
				logger.WithComponent("AccountManager").WithFields(map[string]any{
					"username": username,
					"queue":    qy.Name,
				}).WithError(err).Warn("dedicated queue action failed")
			}
			return
		}
	}
}

// ensurePlanProfile memastikan profil paket ada di router — auto-buat dari
// kolom service_plans (via port.SubscriberAccount) bila belum ada.
func (m *routerAccountManager) ensurePlanProfile(ctx context.Context, driver port.DeviceDriver, serviceType string, acct port.SubscriberAccount) error {
	if acct.Profile == "" || acct.RateLimit == "" {
		return nil // tak cukup data untuk auto-buat; asumsikan profil manual
	}
	if !isHotspot(serviceType) {
		existing, err := m.ppp.ListProfiles(ctx, driver, acct.Profile)
		if err != nil {
			return fmt.Errorf("list profiles: %w", err)
		}
		for _, pr := range existing {
			if pr.Name == acct.Profile {
				return nil
			}
		}
		if _, err := m.ppp.AddProfile(ctx, driver, pppProfileParams(acct)); err != nil {
			return fmt.Errorf("add profile %s: %w", acct.Profile, err)
		}
		return nil
	}
	existing, err := m.hot.GetUserProfiles(ctx, driver)
	if err != nil {
		return fmt.Errorf("list hotspot profiles: %w", err)
	}
	for _, pr := range existing {
		if pr.Name == acct.Profile {
			return nil
		}
	}
	if _, err := m.hot.CreateUserProfile(ctx, driver, hotspotProfileParams(acct)); err != nil {
		return fmt.Errorf("add hotspot profile %s: %w", acct.Profile, err)
	}
	return nil
}

func (m *routerAccountManager) ensurePPPProfileFromSpec(ctx context.Context, driver port.DeviceDriver, spec domainSub.PPPoEProfileSpec) error {
	params := pppProfileParamsFromSpec(spec)
	existing, err := m.ppp.ListProfiles(ctx, driver, params.Name)
	if err != nil {
		return fmt.Errorf("list ppp profiles: %w", err)
	}
	for _, pr := range existing {
		if pr.Name == params.Name {
			return nil
		}
	}
	if _, err := m.ppp.AddProfile(ctx, driver, params); err != nil {
		return fmt.Errorf("add ppp profile %s: %w", params.Name, err)
	}
	return nil
}

func (m *routerAccountManager) ensureHotspotProfileFromSpec(ctx context.Context, driver port.DeviceDriver, spec domainSub.HotspotProfileSpec) error {
	params := hotspotProfileParamsFromSpec(spec)
	existing, err := m.hot.GetUserProfiles(ctx, driver)
	if err != nil {
		return fmt.Errorf("list hotspot profiles: %w", err)
	}
	for _, pr := range existing {
		if pr.Name == params.Name {
			return nil
		}
	}
	if _, err := m.hot.CreateUserProfile(ctx, driver, params); err != nil {
		return fmt.Errorf("add hotspot profile %s: %w", params.Name, err)
	}
	return nil
}

// ─── Update / Isolate / Restore / Suspend / Terminate ───────────────────

// EnsureProfile memastikan profil ada di router dengan rate tertentu
// (auto-create bila belum ada) sebelum akun dipindah ke profil tersebut.
func (m *routerAccountManager) EnsureProfile(ctx context.Context, deviceID, serviceType, profileName, rateLimit string) error {
	driver, err := m.resolve(ctx, deviceID)
	if err != nil {
		return err
	}
	return m.ensurePlanProfile(ctx, driver, serviceType,
		port.SubscriberAccount{Profile: profileName, RateLimit: rateLimit})
}

func (m *routerAccountManager) UpdateAccount(ctx context.Context, deviceID, serviceType, username, newProfile string) error {
	driver, err := m.resolve(ctx, deviceID)
	if err != nil {
		return err
	}
	if isHotspot(serviceType) {
		u, rosID, err := m.findHotspotUser(ctx, driver, username)
		if err != nil {
			return err
		}
		u.Profile = newProfile
		if _, err := m.hot.UpdateUser(ctx, driver, rosID, u); err != nil {
			return fmt.Errorf("update hotspot user %s: %w", username, err)
		}
		m.kickHotspotIfActive(ctx, driver, username)
		return nil
	}
	sec, err := m.findSecret(ctx, driver, username)
	if err != nil {
		return err
	}
	sec.Profile = newProfile
	if _, err := m.ppp.UpdateSecret(ctx, driver, sec.RosID, sec.Params()); err != nil {
		return fmt.Errorf("update ppp secret %s: %w", username, err)
	}
	m.kickPPP(ctx, driver, username)
	return nil
}

func (m *routerAccountManager) Isolate(ctx context.Context, deviceID, serviceType, username string, opt port.IsolationOptions) error {
	driver, err := m.resolve(ctx, deviceID)
	if err != nil {
		return err
	}
	// Profil isolir harus ada sebelum akun dipindah — auto-buat bila belum
	// (konvensi Mikhmon: rate 0/0).
	if opt.IsolirProfile != "" {
		if err := m.ensurePlanProfile(ctx, driver, serviceType, isolirAccount(opt.IsolirProfile, "0/0")); err != nil {
			return fmt.Errorf("ensure isolir profile: %w", err)
		}
	}
	isolirProfile := opt.IsolirProfile
	if err := m.UpdateAccount(ctx, deviceID, serviceType, username, isolirProfile); err != nil {
		return err
	}
	if opt.Redirect != nil && opt.AddressList != "" {
		if err := m.fw.EnsureIsolationRedirect(ctx, driver, *opt.Redirect); err != nil {
			return fmt.Errorf("ensure redirect rules: %w", err)
		}
	}
	// Tandai IP pelanggan agar rule dst-nat mengenai trafiknya.
	ip := m.activeAddress(ctx, driver, serviceType, username)
	if ip != "" && opt.AddressList != "" {
		if err := m.fw.AddToAddressList(ctx, driver, opt.AddressList, ip, "isolir:"+username); err != nil {
			return fmt.Errorf("address-list add %s: %w", ip, err)
		}
	}
	// DEDICATED: queue ikut dimatikan saat isolir (best-effort).
	if isDedicated(serviceType) {
		m.setDedicatedQueueEnabled(ctx, driver, username, false)
	}
	return nil
}

func (m *routerAccountManager) Restore(ctx context.Context, deviceID, serviceType, username, normalProfile, addressList string) error {
	driver, err := m.resolve(ctx, deviceID)
	if err != nil {
		return err
	}
	if addressList != "" {
		if err := m.fw.RemoveFromAddressListByComment(ctx, driver, addressList, "isolir:"+username); err != nil {
			return err
		}
	}
	// DEDICATED: aktifkan kembali queue (best-effort).
	if isDedicated(serviceType) {
		m.setDedicatedQueueEnabled(ctx, driver, username, true)
	}
	return m.UpdateAccount(ctx, deviceID, serviceType, username, normalProfile)
}

func (m *routerAccountManager) Suspend(ctx context.Context, deviceID, serviceType, username string) error {
	driver, err := m.resolve(ctx, deviceID)
	if err != nil {
		return err
	}
	if isHotspot(serviceType) {
		u, rosID, err := m.findHotspotUser(ctx, driver, username)
		if err != nil {
			return err
		}
		u.Disabled = true
		if _, err := m.hot.UpdateUser(ctx, driver, rosID, u); err != nil {
			return fmt.Errorf("suspend hotspot user %s: %w", username, err)
		}
		m.kickHotspotIfActive(ctx, driver, username)
		return nil
	}
	sec, err := m.findSecret(ctx, driver, username)
	if err != nil {
		return err
	}
	if _, err := m.ppp.SetSecretDisabled(ctx, driver, sec.RosID, true); err != nil {
		return fmt.Errorf("disable ppp secret %s: %w", username, err)
	}
	m.kickPPP(ctx, driver, username)
	return nil
}

func (m *routerAccountManager) Terminate(ctx context.Context, deviceID, serviceType, username string) error {
	driver, err := m.resolve(ctx, deviceID)
	if err != nil {
		return err
	}
	if isHotspot(serviceType) {
		_, rosID, err := m.findHotspotUser(ctx, driver, username)
		if err != nil {
			return err
		}
		if _, err := m.hot.RemoveUser(ctx, driver, rosID); err != nil {
			return fmt.Errorf("remove hotspot user %s: %w", username, err)
		}
		return nil
	}
	sec, err := m.findSecret(ctx, driver, username)
	if err != nil {
		return err
	}
	m.kickPPP(ctx, driver, username)
	if _, err := m.ppp.RemoveSecret(ctx, driver, sec.RosID); err != nil {
		return fmt.Errorf("remove ppp secret %s: %w", username, err)
	}
	// DEDICATED: hapus queue pelanggan (best-effort).
	if isDedicated(serviceType) {
		m.removeDedicatedQueue(ctx, driver, username)
	}
	return nil
}

// ─── helpers ────────────────────────────────────────────────────────────

type secretView struct {
	RosID   string
	Name    string
	Profile string
}

// params mengembalikan param update yang hanya mengubah profil.
func (s secretView) Params() port.PPPoESecretParams {
	return port.PPPoESecretParams{Name: s.Name, Profile: s.Profile, Service: "pppoe"}
}

func (m *routerAccountManager) findSecret(ctx context.Context, driver port.DeviceDriver, username string) (secretView, error) {
	secrets, err := m.ppp.ListSecrets(ctx, driver, username)
	if err != nil {
		return secretView{}, fmt.Errorf("list ppp secrets: %w", err)
	}
	for _, s := range secrets {
		if s.Name == username {
			return secretView{RosID: s.RosID, Name: s.Name, Profile: s.Profile}, nil
		}
	}
	return secretView{}, fmt.Errorf("ppp secret %q not found", username)
}

func (m *routerAccountManager) findHotspotUser(ctx context.Context, driver port.DeviceDriver, username string) (port.HotspotUserParams, string, error) {
	users, err := m.hot.ListUsers(ctx, driver, port.ListUsersFilter{Name: username})
	if err != nil {
		return port.HotspotUserParams{}, "", fmt.Errorf("list hotspot users: %w", err)
	}
	for _, u := range users {
		if u.Name == username {
			return port.HotspotUserParams{
				Name: u.Name, Password: u.Password, Profile: u.Profile, Server: u.Server,
				MACAddress: u.MACAddress, Address: u.Address,
				LimitUptime: u.LimitUptime, LimitBytesIn: u.LimitBytesIn, LimitBytesOut: u.LimitBytesOut,
				Comment: u.Comment, Disabled: u.Disabled,
			}, u.RosID, nil
		}
	}
	return port.HotspotUserParams{}, "", fmt.Errorf("hotspot user %q not found", username)
}

func (m *routerAccountManager) kickPPP(ctx context.Context, driver port.DeviceDriver, username string) {
	sessions, err := m.ppp.ListActive(ctx, driver, username)
	if err != nil {
		return
	}
	for _, s := range sessions {
		_, _ = m.ppp.KickActive(ctx, driver, s.RosID)
	}
}

func (m *routerAccountManager) kickHotspotIfActive(ctx context.Context, driver port.DeviceDriver, username string) {
	sessions, err := m.hot.ListActiveSessions(ctx, driver)
	if err != nil {
		return
	}
	for _, s := range sessions {
		if s.User == username {
			_, _ = m.hot.RemoveActiveSession(ctx, driver, s.RosID)
		}
	}
}

// activeAddress mengambil IP sesi aktif pelanggan (untuk address-list).
func (m *routerAccountManager) activeAddress(ctx context.Context, driver port.DeviceDriver, serviceType, username string) string {
	if isHotspot(serviceType) {
		sessions, err := m.hot.ListActiveSessions(ctx, driver)
		if err != nil {
			return ""
		}
		for _, s := range sessions {
			if s.User == username {
				return s.Address
			}
		}
		return ""
	}
	sessions, err := m.ppp.ListActive(ctx, driver, username)
	if err != nil {
		return ""
	}
	for _, s := range sessions {
		if s.Name == username && s.Address != "" {
			return s.Address
		}
	}
	return ""
}

// ─── Hook un-isolir otomatis pasca-lunas ────────────────────────────────

// buildOnPaidRestore mengembalikan hook PaymentProcessor.OnPaid: bila
// invoice yang lunas terkait subscription berstatus ISOLATED, akun di router
// dipulihkan ke profil paket (hapus address-list isolir) dan status kembali
// ACTIVE. Kegagalan router hanya dilog — pembayaran tetap sah.
func buildOnPaidRestore(
	mgr port.RouterAccountManager,
	subs port.SubscriptionRepository,
	plans port.ServicePlanRepository,
	settings port.SettingReader,
) func(ctx context.Context, inv billing.Invoice, pay billing.Payment) {
	return func(ctx context.Context, inv billing.Invoice, pay billing.Payment) {
		if inv.SubscriptionID == nil || *inv.SubscriptionID == "" {
			return
		}
		sub, err := subs.FindByID(ctx, *inv.SubscriptionID)
		if err != nil || sub.Status != "ISOLATED" {
			return
		}
		cfg := port.LoadISPSettings(ctx, settings)

		normalProfile := sub.RouterProfile
		if normalProfile == "" && sub.PlanID != "" {
			if pl, perr := plans.FindByID(ctx, sub.PlanID); perr == nil {
				normalProfile = pl.Name
			}
		}
		if normalProfile == "" {
			logger.WithComponent("OnPaidRestore").WithFields(map[string]any{
				"subscription_id": sub.ID,
			}).Warn("profil paket tidak diketahui; restore dilewati")
			return
		}

		restoreErr := mgr.Restore(ctx, deref(sub.DeviceID), sub.ServiceType,
			sub.RemoteUsername, normalProfile, cfg.IsolirAddressList)
		if restoreErr != nil {
			logger.WithComponent("OnPaidRestore").WithFields(map[string]any{
				"subscription_id": sub.ID,
			}).WithError(restoreErr).Warn("restore router gagal; akan dicoba worker")
			return
		}
		if err := subs.UpdateStatus(ctx, sub.ID, "ACTIVE"); err != nil {
			logger.WithComponent("OnPaidRestore").WithError(err).Warn("update status ACTIVE gagal")
			return
		}
		logger.WithComponent("OnPaidRestore").WithFields(map[string]any{
			"subscription_id": sub.ID,
		}).Info("pelanggan dipulihkan otomatis setelah lunas")
	}
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func timeNowUTC() time.Time {
	n := time.Now().UTC()
	return time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, time.UTC)
}
