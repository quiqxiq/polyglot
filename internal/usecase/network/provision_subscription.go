package network

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/quixiq/polyglot/internal/domain/command"
	domainPlan "github.com/quixiq/polyglot/internal/domain/plan"
	domainSub "github.com/quixiq/polyglot/internal/domain/subscription"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/logger"
)

// ErrDeviceRequired is returned when isolation or provisioning is attempted
// before the subscription has a device mapping.
var ErrDeviceRequired = fmt.Errorf("subscription has no device mapping yet")

// commandResultAlias keeps the helper signature readable.
type commandResultAlias = command.Result

// SubscriptionProvisioner orchestrates the device-side lifecycle of a
// subscription account: provision, isolate (with payment-portal redirect),
// un-isolate, and deprovision. The DB stays mapping-only — every router
// detail is read or ensured on the device itself through the gateways.
type SubscriptionProvisioner struct {
	subs      port.SubscriptionRepository
	plans     port.PlanRepository
	secrets   port.SecretVault
	settings  port.SettingRepository
	ppp       port.PPPGateway
	hotspot   port.HotspotGateway
	isolir    port.IsolationProvisioner
	driverFor func(ctx context.Context, deviceID string) (port.DeviceDriver, error)
}

func NewSubscriptionProvisioner(
	subs port.SubscriptionRepository,
	plans port.PlanRepository,
	secrets port.SecretVault,
	settings port.SettingRepository,
	ppp port.PPPGateway,
	hotspot port.HotspotGateway,
	isolir port.IsolationProvisioner,
	driverFor func(ctx context.Context, deviceID string) (port.DeviceDriver, error),
) *SubscriptionProvisioner {
	return &SubscriptionProvisioner{
		subs: subs, plans: plans, secrets: secrets, settings: settings,
		ppp: ppp, hotspot: hotspot, isolir: isolir, driverFor: driverFor,
	}
}

// Provision ensures the plan's profile exists on the target router, creates
// the account (PPPoE secret or permanent hotspot user), and persists the
// resulting mapping. Idempotent: an existing account with the same username
// is adopted instead of duplicated.
func (p *SubscriptionProvisioner) Provision(ctx context.Context, sub domainSub.Subscription, password string) (domainSub.Subscription, error) {
	if sub.DeviceID == "" || sub.RemoteUsername == "" {
		return sub, ErrDeviceRequired
	}
	planRow, err := p.plans.FindByID(ctx, sub.PlanID)
	if err != nil {
		return sub, fmt.Errorf("load plan: %w", err)
	}
	driver, err := p.driverFor(ctx, sub.DeviceID)
	if err != nil {
		return sub, fmt.Errorf("connect to device %s: %w", sub.DeviceID, err)
	}

	switch sub.ServiceType {
	case domainPlan.ServiceTypePPPoE:
		rosID, err := p.provisionPPPoE(ctx, driver, sub, planRow, password)
		if err != nil {
			return sub, err
		}
		sub.RemoteID = rosID
	case domainPlan.ServiceTypeHotspot:
		rosID, err := p.provisionHotspot(ctx, driver, sub, planRow, password)
		if err != nil {
			return sub, err
		}
		sub.RemoteID = rosID
	default:
		return sub, domainSub.ErrInvalidServiceType
	}

	if err := p.secrets.Put(ctx, secretKey(sub.ID), password); err != nil {
		return sub, fmt.Errorf("store password in vault: %w", err)
	}

	sub.Status = domainSub.StatusActive
	if err := p.subs.Save(ctx, sub); err != nil {
		return sub, fmt.Errorf("persist subscription status: %w", err)
	}
	logger.WithComponent("SubscriptionProvisioner").WithFields(map[string]any{
		"subscription_id": sub.ID, "device_id": sub.DeviceID, "username": sub.RemoteUsername,
	}).Info("subscription provisioned")
	return sub, nil
}

func (p *SubscriptionProvisioner) provisionPPPoE(ctx context.Context, driver port.DeviceDriver, sub domainSub.Subscription, planRow domainPlan.Plan, password string) (string, error) {
	profileName := PlanProfileName(planRow)
	if err := p.ensurePPPProfile(ctx, driver, planRow, profileName); err != nil {
		return "", err
	}
	// Adopt an existing secret from a previous partial provisioning run.
	existing, lookupErr := p.ppp.ListSecrets(ctx, driver, sub.RemoteUsername)
	if lookupErr == nil && len(existing) > 0 && existing[0].RosID != "" {
		return existing[0].RosID, p.subs.UpdateMapping(ctx, sub.ID, sub.RemoteUsername, existing[0].RosID, sub.DeviceID)
	}
	res, err := p.ppp.AddSecret(ctx, driver, port.PPPoESecretParams{
		Name:          sub.RemoteUsername,
		Password:      password,
		Profile:       profileName,
		Service:       "pppoe",
		RemoteAddress: planRow.IPPoolName,
		Comment:       mappingComment(sub),
	})
	if err != nil {
		return "", fmt.Errorf("create pppoe secret %q: %w", sub.RemoteUsername, err)
	}
	rosID := firstResultID(res)
	if rosID == "" {
		// Fallback: resolve via print (some RouterOS versions return no body).
		secrets, ferr := p.ppp.ListSecrets(ctx, driver, sub.RemoteUsername)
		if ferr != nil || len(secrets) == 0 {
			return "", fmt.Errorf("create pppoe secret %q: cannot resolve .id", sub.RemoteUsername)
		}
		rosID = secrets[0].RosID
	}
	return rosID, p.subs.UpdateMapping(ctx, sub.ID, sub.RemoteUsername, rosID, sub.DeviceID)
}

func (p *SubscriptionProvisioner) provisionHotspot(ctx context.Context, driver port.DeviceDriver, sub domainSub.Subscription, planRow domainPlan.Plan, password string) (string, error) {
	profileName := PlanProfileName(planRow)
	if err := p.ensureHotspotProfile(ctx, driver, planRow, profileName); err != nil {
		return "", err
	}
	users, lookupErr := p.hotspot.ListUsers(ctx, driver, port.ListUsersFilter{Name: sub.RemoteUsername})
	if lookupErr == nil && len(users) > 0 && users[0].RosID != "" {
		return users[0].RosID, p.subs.UpdateMapping(ctx, sub.ID, sub.RemoteUsername, users[0].RosID, sub.DeviceID)
	}
	res, err := p.hotspot.AddUser(ctx, driver, port.HotspotUserParams{
		Name:     sub.RemoteUsername,
		Password: password,
		Profile:  profileName,
		Comment:  mappingComment(sub),
	})
	if err != nil {
		return "", fmt.Errorf("create hotspot user %q: %w", sub.RemoteUsername, err)
	}
	rosID := firstResultID(res)
	if rosID == "" {
		users, ferr := p.hotspot.ListUsers(ctx, driver, port.ListUsersFilter{Name: sub.RemoteUsername})
		if ferr != nil || len(users) == 0 {
			return "", fmt.Errorf("create hotspot user %q: cannot resolve .id", sub.RemoteUsername)
		}
		rosID = users[0].RosID
	}
	return rosID, p.subs.UpdateMapping(ctx, sub.ID, sub.RemoteUsername, rosID, sub.DeviceID)
}

// Deprovision removes the device-side account and clears the mapping.
// With removeAccount=false only the DB mapping is cleared.
func (p *SubscriptionProvisioner) Deprovision(ctx context.Context, subscriptionID string, removeAccount bool) error {
	sub, err := p.subs.FindByID(ctx, subscriptionID)
	if err != nil {
		return err
	}
	if removeAccount && sub.DeviceID != "" && sub.RemoteID != "" {
		driver, err := p.driverFor(ctx, sub.DeviceID)
		if err != nil {
			return fmt.Errorf("connect to device %s: %w", sub.DeviceID, err)
		}
		switch sub.ServiceType {
		case domainPlan.ServiceTypePPPoE:
			_, err = p.ppp.RemoveSecret(ctx, driver, sub.RemoteID)
		case domainPlan.ServiceTypeHotspot:
			_, err = p.hotspot.RemoveUser(ctx, driver, sub.RemoteID)
		}
		if err != nil {
			return fmt.Errorf("remove device account %q: %w", sub.RemoteUsername, err)
		}
	}
	_ = p.secrets.Delete(ctx, secretKey(sub.ID))
	if err := p.subs.UpdateStatus(ctx, sub.ID, domainSub.StatusTerminated); err != nil {
		return fmt.Errorf("mark terminated: %w", err)
	}
	logger.WithComponent("SubscriptionProvisioner").WithField("subscription_id", sub.ID).
		Info("subscription deprovisioned")
	return nil
}

// Isolate suspends a subscriber. PPPoE gets the full redirect treatment:
// isolation infrastructure is guaranteed once per router, the secret's
// profile switches to the isolir profile, and the active session is kicked
// so the subscriber reconnects into the isolation pool where NAT dst-nat
// rules steer web traffic to the payment portal. Permanent hotspot users
// are disabled outright (portal redirect is a PPPoE-only feature by design).
func (p *SubscriptionProvisioner) Isolate(ctx context.Context, subscriptionID, reason string) (domainSub.Subscription, error) {
	sub, err := p.mustIsolatable(ctx, subscriptionID)
	if err != nil {
		return domainSub.Subscription{}, err
	}
	driver, err := p.driverFor(ctx, sub.DeviceID)
	if err != nil {
		return domainSub.Subscription{}, fmt.Errorf("connect to device %s: %w", sub.DeviceID, err)
	}

	switch sub.ServiceType {
	case domainPlan.ServiceTypePPPoE:
		cfg := p.isolirConfig(ctx)
		if _, err := p.isolir.EnsureIsolirInfrastructure(ctx, driver, cfg); err != nil {
			return sub, err
		}
		if _, err := p.ppp.UpdateSecret(ctx, driver, sub.RemoteID, port.PPPoESecretParams{Profile: cfg.ProfileName}); err != nil {
			return sub, fmt.Errorf("switch secret %q to isolir profile: %w", sub.RemoteUsername, err)
		}
		p.kickPPPoEByName(ctx, driver, sub.RemoteUsername)
	case domainPlan.ServiceTypeHotspot:
		if _, err := p.hotspot.SetUserDisabled(ctx, driver, sub.RemoteID, true); err != nil {
			return sub, fmt.Errorf("disable hotspot user %q: %w", sub.RemoteUsername, err)
		}
		p.disconnectHotspotByName(ctx, driver, sub.RemoteUsername)
	}

	now := time.Now()
	if err := p.subs.SetIsolation(ctx, sub.ID, domainSub.StatusIsolated, &now, reason); err != nil {
		return sub, fmt.Errorf("record isolation: %w", err)
	}
	logger.WithComponent("SubscriptionProvisioner").WithFields(map[string]any{
		"subscription_id": sub.ID, "reason": reason,
	}).Info("subscription isolated")
	sub.Status = domainSub.StatusIsolated
	sub.IsolatedAt = &now
	sub.IsolationReason = reason
	return sub, nil
}

// Unisolate restores normal service: the secret/user goes back to the
// plan-derived profile and the session is refreshed.
func (p *SubscriptionProvisioner) Unisolate(ctx context.Context, subscriptionID string) (domainSub.Subscription, error) {
	sub, err := p.subs.FindByID(ctx, subscriptionID)
	if err != nil {
		return domainSub.Subscription{}, err
	}
	if sub.DeviceID == "" || sub.RemoteID == "" {
		return sub, ErrDeviceRequired
	}
	planRow, err := p.plans.FindByID(ctx, sub.PlanID)
	if err != nil {
		return sub, fmt.Errorf("load plan: %w", err)
	}
	driver, err := p.driverFor(ctx, sub.DeviceID)
	if err != nil {
		return sub, fmt.Errorf("connect to device %s: %w", sub.DeviceID, err)
	}
	profileName := PlanProfileName(planRow)

	switch sub.ServiceType {
	case domainPlan.ServiceTypePPPoE:
		if err := p.ensurePPPProfile(ctx, driver, planRow, profileName); err != nil {
			return sub, err
		}
		if _, err := p.ppp.UpdateSecret(ctx, driver, sub.RemoteID, port.PPPoESecretParams{Profile: profileName}); err != nil {
			return sub, fmt.Errorf("restore profile on secret %q: %w", sub.RemoteUsername, err)
		}
		p.kickPPPoEByName(ctx, driver, sub.RemoteUsername)
	case domainPlan.ServiceTypeHotspot:
		if err := p.ensureHotspotProfile(ctx, driver, planRow, profileName); err != nil {
			return sub, err
		}
		if _, err := p.hotspot.UpdateUser(ctx, driver, sub.RemoteID, port.HotspotUserParams{Profile: profileName}); err != nil {
			return sub, fmt.Errorf("restore profile on user %q: %w", sub.RemoteUsername, err)
		}
		if _, err := p.hotspot.SetUserDisabled(ctx, driver, sub.RemoteID, false); err != nil {
			return sub, fmt.Errorf("re-enable hotspot user %q: %w", sub.RemoteUsername, err)
		}
	}

	if err := p.subs.SetIsolation(ctx, sub.ID, domainSub.StatusActive, nil, ""); err != nil {
		return sub, fmt.Errorf("clear isolation: %w", err)
	}
	logger.WithComponent("SubscriptionProvisioner").WithField("subscription_id", sub.ID).
		Info("subscription restored")
	sub.Status = domainSub.StatusActive
	sub.IsolatedAt = nil
	sub.IsolationReason = ""
	return sub, nil
}

// ─── helpers ─────────────────────────────────────────────────────────────

func (p *SubscriptionProvisioner) mustIsolatable(ctx context.Context, id string) (domainSub.Subscription, error) {
	sub, err := p.subs.FindByID(ctx, id)
	if err != nil {
		return sub, err
	}
	if sub.Status == domainSub.StatusTerminated {
		return sub, fmt.Errorf("subscription %s is terminated", id)
	}
	if sub.DeviceID == "" || sub.RemoteID == "" {
		return sub, ErrDeviceRequired
	}
	return sub, nil
}

func (p *SubscriptionProvisioner) ensurePPPProfile(ctx context.Context, driver port.DeviceDriver, planRow domainPlan.Plan, profileName string) error {
	profiles, err := p.ppp.ListProfiles(ctx, driver, profileName)
	if err != nil {
		return fmt.Errorf("list profiles: %w", err)
	}
	if len(profiles) > 0 {
		return nil
	}
	if _, err := p.ppp.AddProfile(ctx, driver, port.PPPProfileParams{
		Name:          profileName,
		RateLimit:     rateLimitForPlan(planRow),
		RemoteAddress: planRow.IPPoolName,
		ParentQueue:   planRow.ParentQueue,
		AddressList:   planRow.AddressList,
		Comment:       "PLAN " + planRow.Name + " (auto)",
	}); err != nil {
		return fmt.Errorf("create ppp profile %q: %w", profileName, err)
	}
	return nil
}

func (p *SubscriptionProvisioner) ensureHotspotProfile(ctx context.Context, driver port.DeviceDriver, planRow domainPlan.Plan, profileName string) error {
	profiles, err := p.hotspot.GetUserProfiles(ctx, driver)
	if err != nil {
		return fmt.Errorf("list hotspot profiles: %w", err)
	}
	for _, pr := range profiles {
		if strings.EqualFold(pr.Name, profileName) {
			return nil
		}
	}
	if _, err := p.hotspot.CreateUserProfile(ctx, driver, port.MikhmonProfileParams{
		Name:        profileName,
		RateLimit:   rateLimitForPlan(planRow),
		SharedUsers: strconv.Itoa(planRow.SharedUsers),
		AddressPool: planRow.IPPoolName,
		ParentQueue: planRow.ParentQueue,
		Comment:     "PLAN " + planRow.Name + " (auto)",
	}); err != nil {
		return fmt.Errorf("create hotspot profile %q: %w", profileName, err)
	}
	return nil
}

func (p *SubscriptionProvisioner) kickPPPoEByName(ctx context.Context, driver port.DeviceDriver, username string) {
	sessions, err := p.ppp.ListActive(ctx, driver, username)
	if err != nil {
		logger.WithComponent("SubscriptionProvisioner").
			Warnf("kick %q: list active failed: %v", username, err)
		return
	}
	for _, s := range sessions {
		if _, err := p.ppp.KickActive(ctx, driver, s.RosID); err != nil {
			logger.WithComponent("SubscriptionProvisioner").
				Warnf("kick %q session %s failed: %v", username, s.RosID, err)
		}
	}
}

func (p *SubscriptionProvisioner) disconnectHotspotByName(ctx context.Context, driver port.DeviceDriver, username string) {
	sessions, err := p.hotspot.ListActiveSessions(ctx, driver)
	if err != nil {
		logger.WithComponent("SubscriptionProvisioner").
			Warnf("disconnect %q: list active failed: %v", username, err)
		return
	}
	for _, s := range sessions {
		if s.User != "" && !strings.EqualFold(s.User, username) {
			continue
		}
		if _, err := p.hotspot.RemoveActiveSession(ctx, driver, s.RosID); err != nil {
			logger.WithComponent("SubscriptionProvisioner").
				Warnf("disconnect %q session %s failed: %v", username, s.RosID, err)
		}
	}
}

// isolirConfig reads the "isolir.*" keys with safe defaults.
func (p *SubscriptionProvisioner) isolirConfig(ctx context.Context) port.IsolirConfig {
	return isolirConfigFromSettings(ctx, p.settings, IsolirConfigOverride{})
}

// PlanProfileName derives the router profile name from a plan. The name is
// a REFERENCE — the actual profile object lives on the device and is
// ensured at provisioning time.
func PlanProfileName(p domainPlan.Plan) string {
	var b strings.Builder
	for _, r := range p.Name {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_':
			b.WriteRune(r)
		case r == ' ':
			b.WriteByte('-')
		}
	}
	name := strings.Trim(b.String(), "-")
	if name == "" {
		return "plan-" + p.ID[:8]
	}
	return name
}

// rateLimitString converts kbps values into the RouterOS "rx/tx" form,
// preferring Mbps units when cleanly divisible ("10240" → "10M").
func rateLimitString(downKbps, upKbps int) string {
	return unitString(downKbps) + "/" + unitString(upKbps)
}

// rateLimitForPlan renders the full RouterOS rate-limit value for a plan,
// appending burst tokens when burst parameters are configured:
//
//	"10M/10M 20M/20M 15M/15M 16s"
//	(rx-rate/tx-rate rx-burst-rate/tx-burst-rate rx-burst-threshold/tx-burst-threshold rx-burst-time/tx-burst-time)
func rateLimitForPlan(p domainPlan.Plan) string {
	base := rateLimitString(p.RateDownKbps, p.RateUpKbps)
	if !p.HasBurst() {
		return base
	}
	return base + " " +
		rateLimitString(p.BurstDownKbps, p.BurstUpKbps) + " " +
		rateLimitString(p.BurstThresholdKbps, p.BurstThresholdKbps) + " " +
		strconv.Itoa(p.BurstTimeSeconds) + "s"
}

func unitString(kbps int) string {
	if kbps <= 0 {
		return "0"
	}
	if kbps%1000 == 0 {
		return strconv.Itoa(kbps/1000) + "M"
	}
	if kbps%1024 == 0 {
		return strconv.Itoa(kbps/1024) + "M"
	}
	return strconv.Itoa(kbps) + "k"
}

// secretKey is the vault key convention for a subscription's password.
func secretKey(subscriptionID string) string {
	return "subscription:" + subscriptionID + ":password"
}

func mappingComment(s domainSub.Subscription) string {
	return "POLYGLOT sub:" + s.ID + " cust:" + s.CustomerID
}

// firstResultID extracts the RouterOS .id from an add command response.
func firstResultID(res commandResultAlias) string {
	for _, row := range res.Rows {
		if id := row[".id"]; id != "" {
			return id
		}
	}
	return ""
}
