package provisioner

import (
	"context"
	"fmt"
	"strings"

	billing "github.com/quixiq/polyglot/internal/domain/billing"
	domainDevice "github.com/quixiq/polyglot/internal/domain/device"
	domainSubscription "github.com/quixiq/polyglot/internal/domain/subscription"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/logger"
)

// Provision provisions a subscriber account on the router according to service type.
func (p *Provisioner) Provision(ctx context.Context, deviceID, serviceType string, acct port.SubscriberAccount) error {
	driver, err := p.resolve(ctx, deviceID)
	if err != nil {
		return fmt.Errorf("resolve driver %s: %w", deviceID, err)
	}
	resolved, err := p.ensurePlanProfile(ctx, driver, serviceType, acct)
	if err != nil {
		return err
	}
	if resolved != "" {
		acct.Profile = resolved
	}
	if isHotspot(serviceType) {
		userParams := port.HotspotUserParams{
			Name: acct.Username, Password: acct.Password, Profile: acct.Profile,
			Comment: acct.Comment,
		}
		if _, rosID, err := p.findHotspotUser(ctx, driver, acct.Username); err == nil {
			if _, err := p.hot.UpdateUser(ctx, driver, rosID, userParams); err != nil {
				return fmt.Errorf("update hotspot user %s: %w", acct.Username, err)
			}
			p.kickHotspotIfActive(ctx, driver, acct.Username)
			return nil
		}
		if _, err := p.hot.AddUser(ctx, driver, userParams); err != nil {
			return fmt.Errorf("add hotspot user %s: %w", acct.Username, err)
		}
		return nil
	}
	secParams := port.PPPoESecretParams{
		Name: acct.Username, Password: acct.Password, Profile: acct.Profile,
		Service: "pppoe", Comment: acct.Comment,
	}
	if sec, err := p.findSecret(ctx, driver, acct.Username); err == nil {
		if _, err := p.ppp.UpdateSecret(ctx, driver, sec.RosID, secParams); err != nil {
			return fmt.Errorf("update ppp secret %s: %w", acct.Username, err)
		}
		p.kickPPP(ctx, driver, acct.Username)
	} else {
		if _, err := p.ppp.AddSecret(ctx, driver, secParams); err != nil {
			return fmt.Errorf("add ppp secret %s: %w", acct.Username, err)
		}
	}
	if isDedicated(serviceType) {
		if err := p.ensureDedicatedQueue(ctx, driver, acct); err != nil {
			return fmt.Errorf("dedicated queue: %w", err)
		}
	}
	return nil
}

// UpdateAccount switches the subscriber's profile on the router and kicks the
// active session. Akun yang sebelumnya di-disable (suspend) di-enable kembali:
// pindah profil berarti akun diaktifkan pada profil tujuan (isolir/restore).
func (p *Provisioner) UpdateAccount(ctx context.Context, deviceID, serviceType, username, newProfile string) error {
	driver, err := p.resolve(ctx, deviceID)
	if err != nil {
		return err
	}
	if isHotspot(serviceType) {
		u, rosID, err := p.findHotspotUser(ctx, driver, username)
		if err != nil {
			return err
		}
		u.Profile = newProfile
		u.Disabled = false
		if _, err := p.hot.UpdateUser(ctx, driver, rosID, u); err != nil {
			return fmt.Errorf("update hotspot user %s: %w", username, err)
		}
		p.kickHotspotIfActive(ctx, driver, username)
		return nil
	}
	sec, err := p.findSecret(ctx, driver, username)
	if err != nil {
		return err
	}
	sec.Profile = newProfile
	if _, err := p.ppp.UpdateSecret(ctx, driver, sec.RosID, sec.Params()); err != nil {
		return fmt.Errorf("update ppp secret %s: %w", username, err)
	}
	// Pastikan akun aktif kembali: pindah profil (isolir/restore/change-plan)
	// tidak boleh menyisakan disabled=yes dari suspend (F1-7).
	if _, err := p.ppp.SetSecretDisabled(ctx, driver, sec.RosID, false); err != nil {
		return fmt.Errorf("enable ppp secret %s: %w", username, err)
	}
	p.kickPPP(ctx, driver, username)
	return nil
}

// isolirRateLimit adalah throttle profil isolir. Bukan 0/0 (unlimited):
// pelanggan terisolir tetap dibatasi walau redirect/filter tidak terpasang.
const isolirRateLimit = "64k/64k"

// Isolate isolates a subscriber by changing their profile, adding their IP to the isolation address-list,
// ensuring redirect rules, and kicking active sessions.
func (p *Provisioner) Isolate(ctx context.Context, deviceID, serviceType, username string, opt port.IsolationOptions) error {
	driver, err := p.resolve(ctx, deviceID)
	if err != nil {
		return err
	}
	if opt.IsolirProfile != "" {
		resolved, err := p.ensurePlanProfile(ctx, driver, serviceType,
			isolirAccount(opt.IsolirProfile, isolirRateLimit, opt.AddressList))
		if err != nil {
			return fmt.Errorf("ensure isolir profile: %w", err)
		}
		if resolved != "" {
			opt.IsolirProfile = resolved
		}
	}
	isolirProfile := opt.IsolirProfile
	ip := p.activeAddress(ctx, driver, serviceType, username)
	if err := p.UpdateAccount(ctx, deviceID, serviceType, username, isolirProfile); err != nil {
		return err
	}
	if opt.Redirect != nil && opt.AddressList != "" {
		if err := p.fw.EnsureIsolationRedirect(ctx, driver, *opt.Redirect); err != nil {
			return fmt.Errorf("ensure redirect rules: %w", err)
		}
		// Filter drop juga dipasang agar isolir otomatis membatasi trafik
		// non-HTTP (F6-7).
		if err := p.fw.EnsureIsolationFilter(ctx, driver, opt.AddressList, opt.Redirect.PaymentHost); err != nil {
			return fmt.Errorf("ensure isolation filter: %w", err)
		}
	}
	if ip != "" && opt.AddressList != "" {
		if err := p.fw.AddToAddressList(ctx, driver, opt.AddressList, ip, "isolir:"+username); err != nil {
			return fmt.Errorf("address-list add %s: %w", ip, err)
		}
	}
	if isDedicated(serviceType) {
		p.setDedicatedQueueEnabled(ctx, driver, username, false)
	}
	return nil
}

// Restore restores an isolated subscriber back to their normal plan profile and removes them from address-list.
func (p *Provisioner) Restore(ctx context.Context, deviceID, serviceType, username, normalProfile, addressList string) error {
	driver, err := p.resolve(ctx, deviceID)
	if err != nil {
		return err
	}
	if addressList != "" {
		if err := p.fw.RemoveFromAddressListByComment(ctx, driver, addressList, "isolir:"+username); err != nil {
			return err
		}
	}
	if isDedicated(serviceType) {
		p.setDedicatedQueueEnabled(ctx, driver, username, true)
	}
	return p.UpdateAccount(ctx, deviceID, serviceType, username, normalProfile)
}

// Suspend temporarily disables the subscriber account on the router.
func (p *Provisioner) Suspend(ctx context.Context, deviceID, serviceType, username string) error {
	driver, err := p.resolve(ctx, deviceID)
	if err != nil {
		return err
	}
	if isHotspot(serviceType) {
		return p.suspendHotspot(ctx, driver, username)
	}
	return p.suspendPPPoE(ctx, driver, username)
}

// Terminate permanently removes the subscriber account and dedicated queues from the router.
func (p *Provisioner) Terminate(ctx context.Context, deviceID, serviceType, username string) error {
	driver, err := p.resolve(ctx, deviceID)
	if err != nil {
		return err
	}
	if isHotspot(serviceType) {
		return p.terminateHotspot(ctx, driver, username)
	}
	if err := p.terminatePPPoE(ctx, driver, username); err != nil {
		return err
	}
	if isDedicated(serviceType) {
		p.removeDedicatedQueue(ctx, driver, username)
	}
	return nil
}

func (p *Provisioner) activeAddress(ctx context.Context, driver port.DeviceDriver, serviceType, username string) string {
	if isHotspot(serviceType) {
		sessions, err := p.hot.ListActiveSessions(ctx, driver)
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
	sessions, err := p.ppp.ListActive(ctx, driver, username)
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

// BuildOnPaidRestore returns an OnPaid hook for PaymentProcessor.
// When an invoice for an ISOLATED subscription is paid, the subscriber is restored on the router
// and its status set back to ACTIVE. Failures on router are logged, leaving payment valid.
func BuildOnPaidRestore(
	mgr port.RouterAccountManager,
	subs port.SubscriptionRepository,
	plans port.ServicePlanRepository,
	settings port.SettingReader,
) func(ctx context.Context, inv billing.Invoice, pay billing.Payment) {
	return func(ctx context.Context, inv billing.Invoice, pay billing.Payment) {
		if inv.SubscriptionID == nil || *inv.SubscriptionID == "" {
			return
		}
		// Restore hanya setelah tagihan benar-benar lunas — pembayaran parsial
		// tidak boleh memulihkan layanan (F1-8).
		if inv.Status != billing.StatusPaid {
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
			}).Warn("plan profile unknown; restore skipped")
			return
		}

		restoreErr := mgr.Restore(ctx, deref(sub.DeviceID), sub.ServiceType,
			sub.RemoteUsername, normalProfile, cfg.IsolirAddressList)
		if restoreErr != nil {
			logger.WithComponent("OnPaidRestore").WithFields(map[string]any{
				"subscription_id": sub.ID,
			}).WithError(restoreErr).Warn("router restore failed; worker will retry")
			// Tandai gagal agar lifecycle worker mencoba restore ulang di
			// siklus berikutnya (F2-13).
			sub.ProvisionStatus = domainSubscription.ProvisionFailed
			if serr := subs.Save(ctx, sub); serr != nil {
				logger.WithComponent("OnPaidRestore").WithError(serr).Warn("mark provision failed: update subscription failed")
			}
			return
		}
		if err := subs.UpdateStatus(ctx, sub.ID, "ACTIVE"); err != nil {
			logger.WithComponent("OnPaidRestore").WithError(err).Warn("update status ACTIVE failed")
			return
		}
		logger.WithComponent("OnPaidRestore").WithFields(map[string]any{
			"subscription_id": sub.ID,
		}).Info("subscriber restored automatically after payment")
	}
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// CleanupIsolationAddressList menghapus penanda address-list isolir milik
// username (F6-3). Best-effort: list kosong/username kosong = no-op.
func (p *Provisioner) CleanupIsolationAddressList(ctx context.Context, deviceID, addressList, username string) error {
	if deviceID == "" || addressList == "" || username == "" {
		return nil
	}
	driver, err := p.resolve(ctx, deviceID)
	if err != nil {
		return err
	}
	if err := p.fw.RemoveFromAddressListByComment(ctx, driver, addressList, "isolir:"+username); err != nil {
		return fmt.Errorf("cleanup isolir address-list: %w", err)
	}
	return nil
}

// DeleteIsolationInfrastructure menghapus profil isolir PPP/Hotspot dan
// (opsional) rule firewall + walled-garden terkait (F6-6).
func (p *Provisioner) DeleteIsolationInfrastructure(ctx context.Context, deviceID string, cfg domainDevice.IsolationConfig, removeFirewallRules bool) error {
	if deviceID == "" {
		return fmt.Errorf("device id is required")
	}
	driver, err := p.resolve(ctx, deviceID)
	if err != nil {
		return err
	}
	if cfg.PPPoEProfileName != "" {
		if err := p.deletePPPProfileByName(ctx, driver, cfg.PPPoEProfileName); err != nil {
			return err
		}
	}
	if cfg.HotspotProfileName != "" && p.hot != nil {
		if err := p.deleteHotspotProfileByName(ctx, driver, cfg.HotspotProfileName); err != nil {
			return err
		}
	}
	if !removeFirewallRules {
		return nil
	}
	if err := p.fw.DisableIsolationRedirect(ctx, driver); err != nil {
		return fmt.Errorf("disable isolation redirect: %w", err)
	}
	if cfg.AddressListName != "" {
		if err := p.fw.RemoveIsolationFilter(ctx, driver, cfg.AddressListName); err != nil {
			return fmt.Errorf("remove isolation filter: %w", err)
		}
		if err := p.fw.FlushAddressList(ctx, driver, cfg.AddressListName); err != nil {
			return fmt.Errorf("flush isolation address-list: %w", err)
		}
	}
	if p.hot != nil {
		if err := p.hot.RemoveWalledGarden(ctx, driver); err != nil {
			return fmt.Errorf("remove walled garden: %w", err)
		}
	}
	return nil
}

func (p *Provisioner) deletePPPProfileByName(ctx context.Context, driver port.DeviceDriver, name string) error {
	profs, err := p.ppp.ListProfiles(ctx, driver, name)
	if err != nil {
		return fmt.Errorf("list ppp profiles: %w", err)
	}
	for _, pr := range profs {
		if strings.EqualFold(pr.Name, name) {
			if _, err := p.ppp.RemoveProfile(ctx, driver, pr.RosID); err != nil {
				return fmt.Errorf("remove ppp profile %s: %w", name, err)
			}
			return nil
		}
	}
	return nil
}

func (p *Provisioner) deleteHotspotProfileByName(ctx context.Context, driver port.DeviceDriver, name string) error {
	profs, err := p.hot.GetUserProfiles(ctx, driver)
	if err != nil {
		return fmt.Errorf("list hotspot profiles: %w", err)
	}
	for _, pr := range profs {
		if strings.EqualFold(pr.Name, name) {
			if _, err := p.hot.DeleteUserProfile(ctx, driver, pr.RosID); err != nil {
				return fmt.Errorf("delete hotspot profile %s: %w", name, err)
			}
			return nil
		}
	}
	return nil
}
