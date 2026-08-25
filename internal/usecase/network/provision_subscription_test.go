package network

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainPlan "github.com/quixiq/polyglot/internal/domain/plan"
	domainSub "github.com/quixiq/polyglot/internal/domain/subscription"
	"github.com/quixiq/polyglot/internal/port"
)

// ─── fakes ───────────────────────────────────────────────────────────────

type fakeSubRepo struct {
	subs map[string]domainSub.Subscription
}

func newFakeSubRepo() *fakeSubRepo { return &fakeSubRepo{subs: map[string]domainSub.Subscription{}} }

func (f *fakeSubRepo) Save(ctx context.Context, s domainSub.Subscription) error {
	f.subs[s.ID] = s
	return nil
}

func (f *fakeSubRepo) FindByID(ctx context.Context, id string) (domainSub.Subscription, error) {
	s, ok := f.subs[id]
	if !ok {
		return domainSub.Subscription{}, domainSub.ErrNotFound
	}
	return s, nil
}

func (f *fakeSubRepo) FindByCustomerID(ctx context.Context, customerID string) ([]domainSub.Subscription, error) {
	return nil, nil
}

func (f *fakeSubRepo) FindAll(ctx context.Context) ([]domainSub.Subscription, error) { return nil, nil }

func (f *fakeSubRepo) UpdateStatus(ctx context.Context, id, status string) error {
	s, ok := f.subs[id]
	if !ok {
		return domainSub.ErrNotFound
	}
	s.Status = status
	f.subs[id] = s
	return nil
}

func (f *fakeSubRepo) UpdateMapping(ctx context.Context, id, remoteUsername, remoteID, deviceID string) error {
	s, ok := f.subs[id]
	if !ok {
		return domainSub.ErrNotFound
	}
	s.RemoteUsername = remoteUsername
	s.RemoteID = remoteID
	s.DeviceID = deviceID
	f.subs[id] = s
	return nil
}

func (f *fakeSubRepo) FindByDeviceAndUsername(ctx context.Context, deviceID, username string) (domainSub.Subscription, error) {
	for _, s := range f.subs {
		if s.DeviceID == deviceID && s.RemoteUsername == username {
			return s, nil
		}
	}
	return domainSub.Subscription{}, domainSub.ErrNotFound
}

func (f *fakeSubRepo) SetIsolation(ctx context.Context, id, status string, isolatedAt *time.Time, reason string) error {
	return f.UpdateStatus(ctx, id, status)
}

type fakePlanRepo struct {
	port.PlanRepository // panic on unstubbed calls
	plans               map[string]domainPlan.Plan
}

func (f *fakePlanRepo) FindByID(ctx context.Context, id string) (domainPlan.Plan, error) {
	p, ok := f.plans[id]
	if !ok {
		return domainPlan.Plan{}, domainPlan.ErrNotFound
	}
	return p, nil
}

type fakeSecretVault struct{ secrets map[string]string }

func newFakeSecretVault() *fakeSecretVault { return &fakeSecretVault{secrets: map[string]string{}} }

func (f *fakeSecretVault) Put(ctx context.Context, key, secret string) error {
	f.secrets[key] = secret
	return nil
}

func (f *fakeSecretVault) Get(ctx context.Context, key string) (string, error) {
	return f.secrets[key], nil
}

func (f *fakeSecretVault) Delete(ctx context.Context, key string) error {
	delete(f.secrets, key)
	return nil
}

type fakeSettingRepo struct {
	port.SettingRepository // panic on unstubbed calls
	values                 map[string]string
}

func (f *fakeSettingRepo) GetValue(ctx context.Context, key, fallback string) string {
	if v, ok := f.values[key]; ok && v != "" {
		return v
	}
	return fallback
}

type fakePPPGateway struct {
	port.PPPGateway // panic on unstubbed calls
	secrets         map[string]port.PPPoESecret
	profiles        []port.PPPProfile
	nextID          int
	updatedProfiles []string
	kicked          []string
}

func newFakePPP() *fakePPPGateway {
	return &fakePPPGateway{
		secrets:  make(map[string]port.PPPoESecret),
		profiles: nil,
		nextID:   1,
	}
}

func (f *fakePPPGateway) ListSecrets(ctx context.Context, d port.DeviceDriver, nameFilter string) ([]port.PPPoESecret, error) {
	out := make([]port.PPPoESecret, 0)
	for _, s := range f.secrets {
		if nameFilter == "" || s.Name == nameFilter {
			out = append(out, s)
		}
	}
	return out, nil
}

func (f *fakePPPGateway) AddSecret(ctx context.Context, d port.DeviceDriver, p port.PPPoESecretParams) (commandResultAlias, error) {
	id := "*" + itoa(f.nextID)
	f.nextID++
	f.secrets[p.Name] = port.PPPoESecret{RosID: id, Name: p.Name, Profile: p.Profile, RemoteAddress: p.RemoteAddress}
	rows := []map[string]string{{".id": id}}
	return commandResultAlias{Rows: rows}, nil
}

func (f *fakePPPGateway) UpdateSecret(ctx context.Context, d port.DeviceDriver, rosID string, p port.PPPoESecretParams) (commandResultAlias, error) {
	for name, s := range f.secrets {
		if s.RosID == rosID {
			if p.Profile != "" {
				s.Profile = p.Profile
				f.updatedProfiles = append(f.updatedProfiles, p.Profile)
			}
			f.secrets[name] = s
			return commandResultAlias{}, nil
		}
	}
	return commandResultAlias{}, assert.AnError
}

func (f *fakePPPGateway) ListProfiles(ctx context.Context, d port.DeviceDriver, nameFilter string) ([]port.PPPProfile, error) {
	out := make([]port.PPPProfile, 0)
	for _, pr := range f.profiles {
		if nameFilter == "" || pr.Name == nameFilter {
			out = append(out, pr)
		}
	}
	return out, nil
}

func (f *fakePPPGateway) AddProfile(ctx context.Context, d port.DeviceDriver, p port.PPPProfileParams) (commandResultAlias, error) {
	f.profiles = append(f.profiles, port.PPPProfile{Name: p.Name, RateLimit: p.RateLimit, RemoteAddress: p.RemoteAddress})
	return commandResultAlias{}, nil
}

func (f *fakePPPGateway) ListActive(ctx context.Context, d port.DeviceDriver, nameFilter string) ([]port.PPPActiveSession, error) {
	return nil, nil
}

func (f *fakePPPGateway) KickActive(ctx context.Context, d port.DeviceDriver, rosID string) (commandResultAlias, error) {
	f.kicked = append(f.kicked, rosID)
	return commandResultAlias{}, nil
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	digits := ""
	for n > 0 {
		digits = string(rune('0'+n%10)) + digits
		n /= 10
	}
	return digits
}

type fakeHotspotGateway struct {
	port.HotspotGateway // panic on unstubbed calls
	users               map[string]port.HotspotUser
	disabled            []bool
	nextID              int
}

func newFakeHotspot() *fakeHotspotGateway {
	return &fakeHotspotGateway{users: make(map[string]port.HotspotUser), nextID: 1}
}

func (f *fakeHotspotGateway) ListUsers(ctx context.Context, d port.DeviceDriver, fl port.ListUsersFilter) ([]port.HotspotUser, error) {
	out := make([]port.HotspotUser, 0)
	for _, u := range f.users {
		if fl.Name == "" || u.Name == fl.Name {
			out = append(out, u)
		}
	}
	return out, nil
}

func (f *fakeHotspotGateway) AddUser(ctx context.Context, d port.DeviceDriver, p port.HotspotUserParams) (commandResultAlias, error) {
	id := "*H" + itoa(f.nextID)
	f.nextID++
	f.users[p.Name] = port.HotspotUser{RosID: id, Name: p.Name, Profile: p.Profile}
	return commandResultAlias{Rows: []map[string]string{{".id": id}}}, nil
}

func (f *fakeHotspotGateway) SetUserDisabled(ctx context.Context, d port.DeviceDriver, rosID string, disabled bool) (commandResultAlias, error) {
	for name, u := range f.users {
		if u.RosID == rosID {
			u.Disabled = disabled
			f.users[name] = u
			f.disabled = append(f.disabled, disabled)
			return commandResultAlias{}, nil
		}
	}
	return commandResultAlias{}, assert.AnError
}

func (f *fakeHotspotGateway) UpdateUser(ctx context.Context, d port.DeviceDriver, rosID string, p port.HotspotUserParams) (commandResultAlias, error) {
	for name, u := range f.users {
		if u.RosID == rosID {
			if p.Profile != "" {
				u.Profile = p.Profile
			}
			f.users[name] = u
			return commandResultAlias{}, nil
		}
	}
	return commandResultAlias{}, assert.AnError
}

func (f *fakeHotspotGateway) GetUserProfiles(ctx context.Context, d port.DeviceDriver) ([]port.HotspotUserProfile, error) {
	return nil, nil
}

func (f *fakeHotspotGateway) CreateUserProfile(ctx context.Context, d port.DeviceDriver, p port.MikhmonProfileParams) (commandResultAlias, error) {
	return commandResultAlias{}, nil
}

func (f *fakeHotspotGateway) ListActiveSessions(ctx context.Context, d port.DeviceDriver) ([]port.HotspotActiveSession, error) {
	return nil, nil
}

type fakeIsolir struct {
	ensureCalls  int
	inspectCalls int
	inspection   port.IsolirInspection
	removedCfg   []port.IsolirConfig
	removedErr   error
}

func (f *fakeIsolir) EnsureIsolirProfile(ctx context.Context, d port.DeviceDriver, cfg port.IsolirConfig) (bool, error) {
	return true, nil
}

func (f *fakeIsolir) EnsureIsolirInfrastructure(ctx context.Context, d port.DeviceDriver, cfg port.IsolirConfig) (port.IsolirSetupResult, error) {
	f.ensureCalls++
	return port.IsolirSetupResult{NATRuleIDs: []string{"*N1", "*N2"}}, nil
}

func (f *fakeIsolir) RemoveIsolirInfrastructure(ctx context.Context, d port.DeviceDriver, cfg port.IsolirConfig) error {
	if f.removedErr != nil {
		return f.removedErr
	}
	f.removedCfg = append(f.removedCfg, cfg)
	return nil
}

func (f *fakeIsolir) InspectIsolirInfrastructure(ctx context.Context, d port.DeviceDriver, cfg port.IsolirConfig) (port.IsolirInspection, error) {
	f.inspectCalls++
	return f.inspection, nil
}

// ─── fixtures ────────────────────────────────────────────────────────────

const testDeviceID = "11111111-1111-1111-1111-111111111111"

func fixtureProvisioner(t *testing.T) (*SubscriptionProvisioner, *fakeSubRepo, *fakePPPGateway, *fakeHotspotGateway, *fakeIsolir) {
	t.Helper()
	subs := newFakeSubRepo()
	plans := &fakePlanRepo{plans: map[string]domainPlan.Plan{
		"plan-1": {
			ID: "plan-1", Name: "100 RB", ServiceType: domainPlan.ServiceTypePPPoE,
			RateDownKbps: 10240, RateUpKbps: 10240, Price: 100000, IPPoolName: "PPPOE-POOL",
		},
		"plan-hs": {
			ID: "plan-hs", Name: "HS 20M", ServiceType: domainPlan.ServiceTypeHotspot,
			RateDownKbps: 20480, RateUpKbps: 20480, Price: 200000, SharedUsers: 2,
		},
	}}
	ppp := newFakePPP()
	hs := newFakeHotspot()
	isolir := &fakeIsolir{}
	driverFor := func(ctx context.Context, deviceID string) (port.DeviceDriver, error) {
		return nil, nil // driver tidak dipakai oleh fake gateways
	}
	p := NewSubscriptionProvisioner(subs, plans, newFakeSecretVault(), &fakeSettingRepo{values: map[string]string{}}, ppp, hs, isolir, driverFor)
	return p, subs, ppp, hs, isolir
}

func activePPPoESub() domainSub.Subscription {
	return domainSub.Subscription{
		ID: "sub-1", CustomerID: "cust-1", PlanID: "plan-1",
		DeviceID: testDeviceID, ServiceType: domainPlan.ServiceTypePPPoE,
		RemoteUsername: "BUDI-01", RemoteID: "*5",
		Status: domainSub.StatusActive,
	}
}

// ─── tests ───────────────────────────────────────────────────────────────

func TestSubscriptionProvisioner_Provision_PPPoE(t *testing.T) {
	ctx := context.Background()
	p, subs, ppp, hs, _ := fixtureProvisioner(t)
	sub := activePPPoESub()
	sub.Status = domainSub.StatusPendingProvision
	sub.RemoteID = ""
	require.NoError(t, subs.Save(ctx, sub))
	assert.Empty(t, hs.users)

	updated, err := p.Provision(ctx, sub, "rahasia123")
	require.NoError(t, err)

	assert.Equal(t, domainSub.StatusActive, updated.Status)
	assert.NotEmpty(t, updated.RemoteID)
	secret, ok := ppp.secrets["BUDI-01"]
	require.True(t, ok, "secret harus dibuat di router")
	assert.Equal(t, "100-RB", secret.Profile, "profile mengikuti nama plan")
	assert.Equal(t, "PPPOE-POOL", secret.RemoteAddress)
	stored, err := p.secrets.Get(ctx, "subscription:sub-1:password")
	require.NoError(t, err)
	assert.Equal(t, "rahasia123", stored)
}

func TestSubscriptionProvisioner_IsolateUnisolate_PPPoE(t *testing.T) {
	ctx := context.Background()
	p, subs, ppp, _, isolir := fixtureProvisioner(t)
	sub := activePPPoESub()
	require.NoError(t, subs.Save(ctx, sub))
	// Secret sudah ada di router (hasil provisioning sebelumnya).
	ppp.secrets["BUDI-01"] = port.PPPoESecret{RosID: "*5", Name: "BUDI-01", Profile: "100-RB"}

	isolated, err := p.Isolate(ctx, sub.ID, "tagihan overdue")
	require.NoError(t, err)
	assert.Equal(t, domainSub.StatusIsolated, isolated.Status)
	assert.Equal(t, 1, isolir.ensureCalls, "infrastruktur isolir dipastikan sekali per isolate")
	secret := ppp.secrets["BUDI-01"]
	assert.Equal(t, "isolir", secret.Profile, "profile secret harus berubah ke isolir")

	restored, err := p.Unisolate(ctx, sub.ID)
	require.NoError(t, err)
	assert.Equal(t, domainSub.StatusActive, restored.Status)
	secret = ppp.secrets["BUDI-01"]
	assert.Equal(t, "100-RB", secret.Profile, "profile dikembalikan ke profile plan")
}

func TestSubscriptionProvisioner_Isolate_HotspotPermanent(t *testing.T) {
	ctx := context.Background()
	p, subs, _, hs, isolir := fixtureProvisioner(t)
	sub := activePPPoESub()
	sub.ServiceType = domainPlan.ServiceTypeHotspot
	sub.PlanID = "plan-hs"
	sub.RemoteUsername = "TINI-HS"
	sub.RemoteID = "*H1"
	require.NoError(t, subs.Save(ctx, sub))
	hs.users["TINI-HS"] = port.HotspotUser{RosID: "*H1", Name: "TINI-HS"}

	_, err := p.Isolate(ctx, sub.ID, "unpaid")
	require.NoError(t, err)
	user := hs.users["TINI-HS"]
	assert.True(t, user.Disabled, "hotspot permanent harus di-disable saat isolir")
	assert.Equal(t, 0, isolir.ensureCalls, "hotspot tidak memakai infrastruktur redirect portal")

	_, err = p.Unisolate(ctx, sub.ID)
	require.NoError(t, err)
	user = hs.users["TINI-HS"]
	assert.False(t, user.Disabled)
}

func TestRateLimitString(t *testing.T) {
	assert.Equal(t, "10M/10M", rateLimitString(10240, 10240))
	assert.Equal(t, "512k/512k", rateLimitString(512, 512))
	assert.Equal(t, "20M/10M", rateLimitString(20480, 10240))
}

func TestRateLimitForPlan_Burst(t *testing.T) {
	t.Run("tanpa burst", func(t *testing.T) {
		p := domainPlan.Plan{Name: "BURST-P", ServiceType: domainPlan.ServiceTypePPPoE, RateDownKbps: 10240, RateUpKbps: 10240}
		assert.Equal(t, "10M/10M", rateLimitForPlan(p))
	})
	t.Run("dengan burst penuh", func(t *testing.T) {
		p := domainPlan.Plan{
			Name: "BURST-FULL", ServiceType: domainPlan.ServiceTypePPPoE,
			RateDownKbps: 10240, RateUpKbps: 10240,
			BurstDownKbps: 20480, BurstUpKbps: 20480,
			BurstThresholdKbps: 15360, BurstTimeSeconds: 16,
		}
		assert.Equal(t, "10M/10M 20M/20M 15M/15M 16s", rateLimitForPlan(p))
	})
	t.Run("burst parsial tidak valid", func(t *testing.T) {
		p := domainPlan.Plan{
			Name: "BURST-PART", ServiceType: domainPlan.ServiceTypePPPoE,
			RateDownKbps: 10240, RateUpKbps: 10240,
			BurstDownKbps: 20480, // sisanya kosong
		}
		assert.ErrorIs(t, p.Validate(), domainPlan.ErrInvalidBurst)
	})
	t.Run("burst rate lebih kecil dari base ditolak", func(t *testing.T) {
		p := domainPlan.Plan{
			Name: "BURST-LOW", ServiceType: domainPlan.ServiceTypePPPoE,
			RateDownKbps: 10240, RateUpKbps: 10240,
			BurstDownKbps: 5120, BurstUpKbps: 20480,
			BurstThresholdKbps: 15360, BurstTimeSeconds: 16,
		}
		assert.ErrorIs(t, p.Validate(), domainPlan.ErrInvalidBurst)
	})
}

func TestPlanProfileName(t *testing.T) {
	p := domainPlan.Plan{Name: "Home 20 Mbps", ID: "abcdef12-0000"}
	assert.Equal(t, "Home-20-Mbps", PlanProfileName(p))
}
