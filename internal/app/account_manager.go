package app

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	billing "github.com/quixiq/polyglot/internal/domain/billing"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/internal/registry"
	"github.com/quixiq/polyglot/pkg/logger"
)

// routerAccountManager implements port.RouterAccountManager dengan menyusun
// driver-resolver + PPPGateway + HotspotGateway + FirewallGateway.
// Semua semantik mengikuti ISP nyata: isolir = pindah profil + kick +
// address-list redirect; suspend = disable; terminate = hapus.
type routerAccountManager struct {
	// resolve memetakan deviceID → DeviceDriver (di produksi: registry.Get;
	// di E2E/test: closure ke driver tetap).
	resolve func(ctx context.Context, deviceID string) (port.DeviceDriver, error)
	ppp     port.PPPGateway
	hot     port.HotspotGateway
	fw      port.FirewallGateway
}

var _ port.RouterAccountManager = (*routerAccountManager)(nil)

func newRouterAccountManager(reg *registry.Registry, ppp port.PPPGateway, hot port.HotspotGateway, fw port.FirewallGateway) *routerAccountManager {
	return &routerAccountManager{
		resolve: func(ctx context.Context, deviceID string) (port.DeviceDriver, error) {
			return reg.Get(ctx, deviceID)
		},
		ppp: ppp, hot: hot, fw: fw,
	}
}

const serviceHotspot = "HOTSPOT"

func isHotspot(serviceType string) bool { return strings.EqualFold(serviceType, serviceHotspot) }

// ─── Provision ──────────────────────────────────────────────────────────

func (m *routerAccountManager) Provision(ctx context.Context, deviceID, serviceType string, acct port.SubscriberAccount) error {
	driver, err := m.resolve(ctx, deviceID)
	if err != nil {
		return fmt.Errorf("resolve driver %s: %w", deviceID, err)
	}
	if err := m.ensurePlanProfile(ctx, driver, serviceType, acct.Profile, acct.RateLimit); err != nil {
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
	return nil
}

// ensurePlanProfile memastikan profil paket ada di router — auto-buat dari
// kolom service_plans bila belum ada.
func (m *routerAccountManager) ensurePlanProfile(ctx context.Context, driver port.DeviceDriver, serviceType, profileName, rateLimit string) error {
	if profileName == "" || rateLimit == "" {
		return nil // tak cukup data untuk auto-buat; asumsikan profil manual
	}
	if !isHotspot(serviceType) {
		existing, err := m.ppp.ListProfiles(ctx, driver, profileName)
		if err != nil {
			return fmt.Errorf("list profiles: %w", err)
		}
		for _, pr := range existing {
			if pr.Name == profileName {
				return nil
			}
		}
		if _, err := m.ppp.AddProfile(ctx, driver, port.PPPProfileParams{
			Name:      profileName,
			RateLimit: rateLimit,
			Comment:   "AUTO plan profile",
		}); err != nil {
			return fmt.Errorf("add profile %s: %w", profileName, err)
		}
		return nil
	}
	existing, err := m.hot.GetUserProfiles(ctx, driver)
	if err != nil {
		return fmt.Errorf("list hotspot profiles: %w", err)
	}
	for _, pr := range existing {
		if pr.Name == profileName {
			return nil
		}
	}
	if _, err := m.hot.CreateUserProfile(ctx, driver, port.MikhmonProfileParams{
		Name: profileName, RateLimit: rateLimit, Comment: "AUTO plan profile",
	}); err != nil {
		return fmt.Errorf("add hotspot profile %s: %w", profileName, err)
	}
	return nil
}

// ─── Update / Isolate / Restore / Suspend / Terminate ───────────────────

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
		if err := m.ensurePlanProfile(ctx, driver, serviceType, opt.IsolirProfile, "0/0"); err != nil {
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

// kbpsToMikrotikRate converts kbps → RouterOS rate string ("5M", "512k").
func kbpsToMikrotikRate(kbps int) string {
	if kbps <= 0 {
		return ""
	}
	if kbps >= 1000 {
		return strconv.Itoa((kbps+500)/1000) + "M"
	}
	return strconv.Itoa(kbps) + "k"
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
