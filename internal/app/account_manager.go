// DEVIASI: RouterAccountManager menyatukan seluruh operasi provisi lifecycle PPP, Hotspot, Firewall, Plan Profile Sync & Script router.
package app

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	billing "github.com/quixiq/polyglot/internal/domain/billing"
	domainCommand "github.com/quixiq/polyglot/internal/domain/command"
	domainDevice "github.com/quixiq/polyglot/internal/domain/device"
	domainPlan "github.com/quixiq/polyglot/internal/domain/plan"
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
				// Profil sudah ada — pastikan address-list sesuai bila diisi.
				if acct.AddressList != "" && pr.AddressList != acct.AddressList {
					p := pppProfileParams(acct)
					if _, err := m.ppp.UpdateProfile(ctx, driver, pr.RosID, p); err != nil {
						return fmt.Errorf("update profile %s address-list: %w", acct.Profile, err)
					}
				}
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
			// Profil sudah ada — pastikan address-list sesuai bila diisi.
			if acct.AddressList != "" && pr.AddressList != acct.AddressList {
				p := hotspotProfileParams(acct)
				if _, err := m.hot.UpdateUserProfile(ctx, driver, pr.RosID, p); err != nil {
					return fmt.Errorf("update hotspot profile %s address-list: %w", acct.Profile, err)
				}
			}
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
	// Ambil IP sesi aktif sebelum akun dipindah dan sesi di-kick
	ip := m.activeAddress(ctx, driver, serviceType, username)
	if err := m.UpdateAccount(ctx, deviceID, serviceType, username, isolirProfile); err != nil {
		return err
	}
	if opt.Redirect != nil && opt.AddressList != "" {
		if err := m.fw.EnsureIsolationRedirect(ctx, driver, *opt.Redirect); err != nil {
			return fmt.Errorf("ensure redirect rules: %w", err)
		}
	}
	// Tandai IP pelanggan agar rule dst-nat mengenai trafiknya.
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
	if err == nil {
		for _, s := range sessions {
			if s.User == username {
				_, _ = m.hot.RemoveActiveSession(ctx, driver, s.RosID)
			}
		}
	}
	cookies, err := m.hot.ListCookies(ctx, driver)
	if err == nil {
		for _, c := range cookies {
			if c.User == username {
				_, _ = m.hot.DeleteCookie(ctx, driver, c.RosID)
			}
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

func (m *routerAccountManager) EnsureIsolationInfrastructure(ctx context.Context, deviceID string, cfg domainDevice.IsolationConfig) error {
	driver, err := m.resolve(ctx, deviceID)
	if err != nil {
		return err
	}
	rate := cfg.RateLimit
	if rate == "" {
		rate = "0/0"
	}
	pppoeProf := cfg.PPPoEProfileName
	if pppoeProf == "" {
		pppoeProf = "ISOLIR"
	}
	hotspotProf := cfg.HotspotProfileName
	if hotspotProf == "" {
		hotspotProf = "ISOLIR"
	}
	addrList := cfg.AddressListName
	if addrList == "" {
		addrList = "ISOLIR_USERS"
	}

	localAddr := cfg.LocalAddress
	if localAddr == "" {
		localAddr = "10.100.0.1"
	}
	remotePool := cfg.RemoteAddressPool
	if remotePool == "" {
		remotePool = "pool-isolir"
	}

	// 1. Ensure dedicated IP pool for PPPoE isolation
	poolRes, err := driver.Execute(ctx, domainCommand.Command{
		Raw:  "/ip/pool/print",
		Args: map[string]string{"?name": remotePool},
	})
	if err == nil && len(poolRes.Rows) == 0 {
		_, _ = driver.Execute(ctx, domainCommand.Command{
			Raw: "/ip/pool/add",
			Args: map[string]string{
				"name":    remotePool,
				"ranges":  "10.100.0.10-10.100.0.250",
				"comment": "polyglot:isolation",
			},
		})
	}

	// 2. Ensure DNS static entries for isolation domains pointing to redirect IP
	if cfg.RedirectIP != "" {
		for _, domain := range cfg.WalledGardenDomains {
			if strings.HasSuffix(domain, ".test") || strings.HasSuffix(domain, ".isp.net") {
				cleanDom := strings.TrimPrefix(domain, "*")
				dnsRes, _ := driver.Execute(ctx, domainCommand.Command{
					Raw:  "/ip/dns/static/print",
					Args: map[string]string{"?name": cleanDom},
				})
				if len(dnsRes.Rows) == 0 {
					_, _ = driver.Execute(ctx, domainCommand.Command{
						Raw: "/ip/dns/static/add",
						Args: map[string]string{
							"name":    cleanDom,
							"address": cfg.RedirectIP,
							"comment": "polyglot:isolation",
						},
					})
				}
			}
		}
	}

	// 4. Ensure PPPoE isolation profile
	if err := m.ensurePlanProfile(ctx, driver, "PPPOE", port.SubscriberAccount{
		Profile:           pppoeProf,
		RateLimit:         rate,
		AddressList:       addrList,
		LocalAddress:      localAddr,
		RemoteAddressPool: remotePool,
		DNSServer:         localAddr + ",8.8.8.8",
	}); err != nil {
		return fmt.Errorf("ensure pppoe isolir profile: %w", err)
	}

	// 5. Ensure Hotspot isolation profile
	if err := m.ensurePlanProfile(ctx, driver, "HOTSPOT", port.SubscriberAccount{
		Profile:     hotspotProf,
		RateLimit:   rate,
		AddressList: addrList,
	}); err != nil {
		return fmt.Errorf("ensure hotspot isolir profile: %w", err)
	}

	// 6. Ensure NAT Redirect and Firewall Filter rules if configured
	if cfg.NATRedirectEnabled && cfg.RedirectIP != "" {
		redirCfg := port.IsolationRedirectConfig{
			SrcAddressList: addrList,
			PaymentHost:    cfg.RedirectIP,
			PaymentPort:    strconv.Itoa(cfg.RedirectPort),
		}
		if err := m.fw.EnsureIsolationRedirect(ctx, driver, redirCfg); err != nil {
			return fmt.Errorf("ensure isolation redirect: %w", err)
		}
		if err := m.fw.EnsureIsolationFilter(ctx, driver, addrList, cfg.RedirectIP); err != nil {
			return fmt.Errorf("ensure isolation filter: %w", err)
		}
	}

	// 7. Ensure Hotspot Walled Garden for domains and portal
	if len(cfg.WalledGardenDomains) > 0 || cfg.RedirectIP != "" {
		portStr := ""
		if cfg.RedirectPort > 0 {
			portStr = strconv.Itoa(cfg.RedirectPort)
		}
		if err := m.hot.EnsureWalledGarden(ctx, driver, cfg.WalledGardenDomains, cfg.RedirectIP, portStr); err != nil {
			return fmt.Errorf("ensure walled garden: %w", err)
		}
	}
	return nil
}

func (m *routerAccountManager) GetIsolationInfrastructureStatus(ctx context.Context, deviceID string) (domainDevice.IsolationStatus, error) {
	driver, err := m.resolve(ctx, deviceID)
	if err != nil {
		return domainDevice.IsolationStatus{}, err
	}
	status := domainDevice.IsolationStatus{
		Config: domainDevice.DefaultIsolationConfig(),
	}

	// Check PPP profile
	pppProfiles, err := m.ppp.ListProfiles(ctx, driver, "")
	if err == nil {
		for _, p := range pppProfiles {
			if strings.EqualFold(p.Name, status.Config.PPPoEProfileName) {
				status.PPPoEProfileExists = true
				break
			}
		}
	}

	// Check Hotspot profile
	hotProfiles, err := m.hot.GetUserProfiles(ctx, driver)
	if err == nil {
		for _, p := range hotProfiles {
			if strings.EqualFold(p.Name, status.Config.HotspotProfileName) {
				status.HotspotProfileExists = true
				break
			}
		}
	}

	// Check NAT redirect rule
	hasRedirect, err := m.fw.HasIsolationRedirect(ctx, driver, status.Config.AddressListName)
	if err == nil {
		status.NATRedirectExists = hasRedirect
	}

	// Count isolated users in the address list
	count, err := m.fw.CountAddressListEntries(ctx, driver, status.Config.AddressListName)
	if err == nil {
		status.IsolatedUsersCount = count
	}

	return status, nil
}

func (m *routerAccountManager) ApplyIntegrationScript(ctx context.Context, deviceID, profileName, serviceType, scriptType, script string) error {
	driver, err := m.resolve(ctx, deviceID)
	if err != nil {
		return err
	}
	if isHotspot(serviceType) {
		profs, err := m.hot.GetUserProfiles(ctx, driver)
		if err != nil {
			return fmt.Errorf("get hotspot user profiles: %w", err)
		}
		var targetProf *port.HotspotUserProfile
		for i := range profs {
			if strings.EqualFold(profs[i].Name, profileName) {
				targetProf = &profs[i]
				break
			}
		}
		if targetProf == nil {
			return fmt.Errorf("hotspot profile %q not found", profileName)
		}
		_, err = m.hot.UpdateUserProfile(ctx, driver, targetProf.RosID, port.MikhmonProfileParams{
			Name:        targetProf.Name,
			AddressPool: targetProf.AddressPool,
			SharedUsers: targetProf.SharedUsers,
			RateLimit:   targetProf.RateLimit,
			ParentQueue: targetProf.ParentQueue,
			Comment:     targetProf.Comment,
			ExpireMode:  port.ExpireModeNone,
		})
		if err != nil {
			return fmt.Errorf("update hotspot user profile: %w", err)
		}
		return nil
	}

	// PPPoE profile
	profs, err := m.ppp.ListProfiles(ctx, driver, "")
	if err != nil {
		return fmt.Errorf("list ppp profiles: %w", err)
	}
	var targetProf *port.PPPProfile
	for i := range profs {
		if strings.EqualFold(profs[i].Name, profileName) {
			targetProf = &profs[i]
			break
		}
	}
	if targetProf == nil {
		return fmt.Errorf("ppp profile %q not found", profileName)
	}
	onUp := targetProf.OnUp
	onDown := targetProf.OnDown
	switch scriptType {
	case "on-up":
		onUp = script
	case "on-down":
		onDown = script
	}
	_, err = m.ppp.UpdateProfile(ctx, driver, targetProf.RosID, port.PPPProfileParams{
		Name:          targetProf.Name,
		LocalAddress:  targetProf.LocalAddress,
		RemoteAddress: targetProf.RemoteAddress,
		RateLimit:     targetProf.RateLimit,
		ParentQueue:   targetProf.ParentQueue,
		Comment:       targetProf.Comment,
		OnUp:          onUp,
		OnDown:        onDown,
	})
	if err != nil {
		return fmt.Errorf("update ppp profile: %w", err)
	}
	return nil
}

// SyncPlanProfile creates or updates the corresponding profile on the target router.
func (m *routerAccountManager) SyncPlanProfile(ctx context.Context, deviceID string, pl domainPlan.ServicePlan) error {
	if deviceID == "" || pl.Name == "" {
		return nil
	}
	driver, err := m.resolve(ctx, deviceID)
	if err != nil {
		return err
	}

	rate := pl.RateLimitWithBurst()
	if !isHotspot(pl.ServiceType) {
		// PPPoE Profile
		remotePool := pl.RemoteAddressPool
		if pl.PPPoE != nil && pl.PPPoE.RemoteAddressPool != "" {
			remotePool = pl.PPPoE.RemoteAddressPool
		}
		addrList := pl.AddressList
		if pl.PPPoE != nil && pl.PPPoE.AddressList != "" {
			addrList = pl.PPPoE.AddressList
		}
		sessionTimeout := pl.SessionTimeout
		if pl.PPPoE != nil && pl.PPPoE.SessionTimeout != "" {
			sessionTimeout = pl.PPPoE.SessionTimeout
		}
		idleTimeout := pl.IdleTimeout
		if pl.PPPoE != nil && pl.PPPoE.IdleTimeout != "" {
			idleTimeout = pl.PPPoE.IdleTimeout
		}

		profs, err := m.ppp.ListProfiles(ctx, driver, "")
		if err != nil {
			return fmt.Errorf("list ppp profiles: %w", err)
		}
		var existing *port.PPPProfile
		for i := range profs {
			if strings.EqualFold(profs[i].Name, pl.Name) {
				existing = &profs[i]
				break
			}
		}

		params := port.PPPProfileParams{
			Name:           pl.Name,
			RateLimit:      rate,
			RemoteAddress:  remotePool,
			AddressList:    addrList,
			ParentQueue:    pl.ParentQueue,
			SessionTimeout: sessionTimeout,
			IdleTimeout:    idleTimeout,
			Comment:        "polyglot plan:" + pl.ID,
		}
		if existing != nil {
			params.LocalAddress = existing.LocalAddress
			params.OnUp = existing.OnUp
			params.OnDown = existing.OnDown
			_, err = m.ppp.UpdateProfile(ctx, driver, existing.RosID, params)
			if err != nil {
				return fmt.Errorf("update ppp profile %s: %w", pl.Name, err)
			}
		} else {
			_, err = m.ppp.AddProfile(ctx, driver, params)
			if err != nil {
				return fmt.Errorf("create ppp profile %s: %w", pl.Name, err)
			}
		}
		return nil
	}

	// Hotspot User Profile
	ipPool := pl.IPPoolName
	if pl.Hotspot != nil && pl.Hotspot.IPPoolName != "" {
		ipPool = pl.Hotspot.IPPoolName
	}
	addrList := pl.AddressList
	if pl.Hotspot != nil && pl.Hotspot.AddressList != "" {
		addrList = pl.Hotspot.AddressList
	}
	sharedUsers := 1
	if pl.Hotspot != nil && pl.Hotspot.SharedUsers > 0 {
		sharedUsers = pl.Hotspot.SharedUsers
	} else if pl.SharedUsers > 0 {
		sharedUsers = pl.SharedUsers
	}
	sessionTimeout := pl.SessionTimeout
	if pl.Hotspot != nil && pl.Hotspot.SessionTimeout != "" {
		sessionTimeout = pl.Hotspot.SessionTimeout
	}
	idleTimeout := pl.IdleTimeout
	if pl.Hotspot != nil && pl.Hotspot.IdleTimeout != "" {
		idleTimeout = pl.Hotspot.IdleTimeout
	}

	profs, err := m.hot.GetUserProfiles(ctx, driver)
	if err != nil {
		return fmt.Errorf("list hotspot profiles: %w", err)
	}
	var existing *port.HotspotUserProfile
	for i := range profs {
		if strings.EqualFold(profs[i].Name, pl.Name) {
			existing = &profs[i]
			break
		}
	}

	params := port.MikhmonProfileParams{
		Name:           pl.Name,
		RateLimit:      rate,
		AddressPool:    ipPool,
		AddressList:    addrList,
		SharedUsers:    intToStrOr(sharedUsers, "1"),
		ParentQueue:    pl.ParentQueue,
		SessionTimeout: sessionTimeout,
		IdleTimeout:    idleTimeout,
		Comment:        "polyglot plan:" + pl.ID,
	}
	if existing != nil {
		params.OnLogin = existing.OnLogin
		_, err = m.hot.UpdateUserProfile(ctx, driver, existing.RosID, params)
		if err != nil {
			return fmt.Errorf("update hotspot profile %s: %w", pl.Name, err)
		}
	} else {
		_, err = m.hot.CreateUserProfile(ctx, driver, params)
		if err != nil {
			return fmt.Errorf("create hotspot profile %s: %w", pl.Name, err)
		}
	}
	return nil
}

// DeletePlanProfile removes the corresponding profile from the router if it exists.
func (m *routerAccountManager) DeletePlanProfile(ctx context.Context, deviceID string, serviceType, profileName string) error {
	if deviceID == "" || profileName == "" {
		return nil
	}
	driver, err := m.resolve(ctx, deviceID)
	if err != nil {
		return err
	}

	if !isHotspot(serviceType) {
		profs, err := m.ppp.ListProfiles(ctx, driver, "")
		if err != nil {
			return fmt.Errorf("list ppp profiles: %w", err)
		}
		for _, pr := range profs {
			if strings.EqualFold(pr.Name, profileName) {
				_, err := m.ppp.RemoveProfile(ctx, driver, pr.RosID)
				if err != nil {
					return fmt.Errorf("remove ppp profile: %w", err)
				}
				return nil
			}
		}
		return nil
	}

	profs, err := m.hot.GetUserProfiles(ctx, driver)
	if err != nil {
		return fmt.Errorf("get hotspot user profiles: %w", err)
	}
	for _, pr := range profs {
		if strings.EqualFold(pr.Name, profileName) {
			_, err := m.hot.DeleteUserProfile(ctx, driver, pr.RosID)
			if err != nil {
				return fmt.Errorf("delete hotspot user profile: %w", err)
			}
			return nil
		}
	}
	return nil
}

func timeNowUTC() time.Time {
	n := time.Now().UTC()
	return time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, time.UTC)
}
