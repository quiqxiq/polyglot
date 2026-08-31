package provisioner

import (
	"context"
	"fmt"

	billing "github.com/quixiq/polyglot/internal/domain/billing"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/logger"
)

// Provision provisions a subscriber account on the router according to service type.
func (p *Provisioner) Provision(ctx context.Context, deviceID, serviceType string, acct port.SubscriberAccount) error {
	driver, err := p.resolve(ctx, deviceID)
	if err != nil {
		return fmt.Errorf("resolve driver %s: %w", deviceID, err)
	}
	if err := p.ensurePlanProfile(ctx, driver, serviceType, acct); err != nil {
		return err
	}
	if isHotspot(serviceType) {
		if _, err := p.hot.AddUser(ctx, driver, port.HotspotUserParams{
			Name: acct.Username, Password: acct.Password, Profile: acct.Profile,
			Comment: acct.Comment,
		}); err != nil {
			return fmt.Errorf("add hotspot user %s: %w", acct.Username, err)
		}
		return nil
	}
	if _, err := p.ppp.AddSecret(ctx, driver, port.PPPoESecretParams{
		Name: acct.Username, Password: acct.Password, Profile: acct.Profile,
		Service: "pppoe", Comment: acct.Comment,
	}); err != nil {
		return fmt.Errorf("add ppp secret %s: %w", acct.Username, err)
	}
	if isDedicated(serviceType) {
		if err := p.ensureDedicatedQueue(ctx, driver, acct); err != nil {
			return fmt.Errorf("dedicated queue: %w", err)
		}
	}
	return nil
}

// UpdateAccount switches the subscriber's profile on the router and kicks the active session.
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
	p.kickPPP(ctx, driver, username)
	return nil
}

// Isolate isolates a subscriber by changing their profile, adding their IP to the isolation address-list,
// ensuring redirect rules, and kicking active sessions.
func (p *Provisioner) Isolate(ctx context.Context, deviceID, serviceType, username string, opt port.IsolationOptions) error {
	driver, err := p.resolve(ctx, deviceID)
	if err != nil {
		return err
	}
	if opt.IsolirProfile != "" {
		if err := p.ensurePlanProfile(ctx, driver, serviceType, isolirAccount(opt.IsolirProfile, "0/0")); err != nil {
			return fmt.Errorf("ensure isolir profile: %w", err)
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
