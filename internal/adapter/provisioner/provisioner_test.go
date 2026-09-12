package provisioner

import (
	"context"
	"errors"
	"strconv"
	"testing"

	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	"github.com/quixiq/polyglot/internal/domain/command"
	domainDevice "github.com/quixiq/polyglot/internal/domain/device"
	domainPlan "github.com/quixiq/polyglot/internal/domain/plan"
	domainSubscription "github.com/quixiq/polyglot/internal/domain/subscription"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/internal/port/mocktest"
)

type mockDriver struct {
	port.DeviceDriver
}

func (m *mockDriver) Execute(_ context.Context, _ command.Command) (command.Result, error) {
	return command.Result{Output: "ok"}, nil
}

type mockPPPGateway struct {
	port.PPPGateway
	secrets  []port.PPPoESecret
	profiles []port.PPPProfile
}

func (m *mockPPPGateway) ListSecrets(_ context.Context, _ port.DeviceDriver, _ string) ([]port.PPPoESecret, error) {
	return m.secrets, nil
}

func (m *mockPPPGateway) AddSecret(_ context.Context, _ port.DeviceDriver, p port.PPPoESecretParams) (command.Result, error) {
	m.secrets = append(m.secrets, port.PPPoESecret{
		RosID:   "*1",
		Name:    p.Name,
		Profile: p.Profile,
	})
	return command.Result{Output: "ok"}, nil
}

func (m *mockPPPGateway) UpdateSecret(_ context.Context, _ port.DeviceDriver, rosID string, p port.PPPoESecretParams) (command.Result, error) {
	for i := range m.secrets {
		if m.secrets[i].RosID == rosID || m.secrets[i].Name == p.Name {
			m.secrets[i].Profile = p.Profile
		}
	}
	return command.Result{Output: "ok"}, nil
}

func (m *mockPPPGateway) SetSecretDisabled(_ context.Context, _ port.DeviceDriver, rosID string, disabled bool) (command.Result, error) {
	for i := range m.secrets {
		if m.secrets[i].RosID == rosID {
			m.secrets[i].Disabled = disabled
		}
	}
	return command.Result{Output: "ok"}, nil
}

func (m *mockPPPGateway) RemoveSecret(_ context.Context, _ port.DeviceDriver, rosID string) (command.Result, error) {
	filtered := m.secrets[:0]
	for _, s := range m.secrets {
		if s.RosID != rosID {
			filtered = append(filtered, s)
		}
	}
	m.secrets = filtered
	return command.Result{Output: "ok"}, nil
}

func (m *mockPPPGateway) ListProfiles(_ context.Context, _ port.DeviceDriver, _ string) ([]port.PPPProfile, error) {
	return m.profiles, nil
}

func (m *mockPPPGateway) AddProfile(_ context.Context, _ port.DeviceDriver, p port.PPPProfileParams) (command.Result, error) {
	m.profiles = append(m.profiles, port.PPPProfile{
		RosID: "*p1",
		Name:  p.Name,
	})
	return command.Result{Output: "ok"}, nil
}

func (m *mockPPPGateway) ListActive(_ context.Context, _ port.DeviceDriver, _ string) ([]port.PPPActiveSession, error) {
	return nil, nil
}

func (m *mockPPPGateway) RemoveProfile(_ context.Context, _ port.DeviceDriver, rosID string) (command.Result, error) {
	filtered := m.profiles[:0]
	for _, pr := range m.profiles {
		if pr.RosID != rosID {
			filtered = append(filtered, pr)
		}
	}
	m.profiles = filtered
	return command.Result{Output: "ok"}, nil
}

func (m *mockPPPGateway) KickActive(_ context.Context, _ port.DeviceDriver, _ string) (command.Result, error) {
	return command.Result{Output: "ok"}, nil
}

type mockFirewallGateway struct {
	port.FirewallGateway
	addressList          []string
	filterCalls          int
	disableRedirectCalls int
	flushCalls           int
	removedByComment     []string
}

func (m *mockFirewallGateway) AddToAddressList(_ context.Context, _ port.DeviceDriver, list, address, _ string) error {
	m.addressList = append(m.addressList, list+":"+address)
	return nil
}

func (m *mockFirewallGateway) RemoveFromAddressListByComment(_ context.Context, _ port.DeviceDriver, list, comment string) error {
	m.removedByComment = append(m.removedByComment, list+":"+comment)
	return nil
}

func (m *mockFirewallGateway) EnsureIsolationRedirect(_ context.Context, _ port.DeviceDriver, _ port.IsolationRedirectConfig) error {
	return nil
}

func (m *mockFirewallGateway) EnsureIsolationFilter(_ context.Context, _ port.DeviceDriver, _, _ string) error {
	m.filterCalls++
	return nil
}

func (m *mockFirewallGateway) DisableIsolationRedirect(_ context.Context, _ port.DeviceDriver) error {
	m.disableRedirectCalls++
	return nil
}

func (m *mockFirewallGateway) RemoveIsolationFilter(_ context.Context, _ port.DeviceDriver, _ string) error {
	return nil
}

func (m *mockFirewallGateway) FlushAddressList(_ context.Context, _ port.DeviceDriver, _ string) error {
	m.flushCalls++
	return nil
}

func TestProvisioner_PPPoELifecycle(t *testing.T) {
	ctx := context.Background()
	drv := &mockDriver{}
	pppGW := &mockPPPGateway{}
	fwGW := &mockFirewallGateway{}

	prov := NewWithResolver(
		func(_ context.Context, _ string) (port.DeviceDriver, error) {
			return drv, nil
		},
		pppGW, nil, fwGW, nil,
	)

	// 1. Provision
	acct := port.SubscriberAccount{
		Username:  "user1",
		Password:  "pass1",
		Profile:   "PLAN-10M",
		RateLimit: "10M/10M",
	}
	if err := prov.Provision(ctx, "dev1", "PPPOE", acct); err != nil {
		t.Fatalf("Provision failed: %v", err)
	}
	if len(pppGW.secrets) != 1 || pppGW.secrets[0].Name != "user1" {
		t.Fatalf("Secret not created properly: %+v", pppGW.secrets)
	}

	// 2. Isolate
	opt := port.IsolationOptions{
		IsolirProfile: "ISOLIR",
		AddressList:   "ISOLIR_LIST",
	}
	if err := prov.Isolate(ctx, "dev1", "PPPOE", "user1", opt); err != nil {
		t.Fatalf("Isolate failed: %v", err)
	}
	if pppGW.secrets[0].Profile != "ISOLIR" {
		t.Fatalf("Secret profile not updated to ISOLIR: %+v", pppGW.secrets[0])
	}

	// 3. Restore
	if err := prov.Restore(ctx, "dev1", "PPPOE", "user1", "PLAN-10M", "ISOLIR_LIST"); err != nil {
		t.Fatalf("Restore failed: %v", err)
	}
	if pppGW.secrets[0].Profile != "PLAN-10M" {
		t.Fatalf("Secret profile not restored: %+v", pppGW.secrets[0])
	}

	// 4. Suspend
	if err := prov.Suspend(ctx, "dev1", "PPPOE", "user1"); err != nil {
		t.Fatalf("Suspend failed: %v", err)
	}
	if !pppGW.secrets[0].Disabled {
		t.Fatalf("Secret not disabled on suspend")
	}

	// 5. Terminate
	if err := prov.Terminate(ctx, "dev1", "PPPOE", "user1"); err != nil {
		t.Fatalf("Terminate failed: %v", err)
	}
	if len(pppGW.secrets) != 0 {
		t.Fatalf("Secret not removed on terminate: %+v", pppGW.secrets)
	}
}

func TestProvisioner_SyncPlanProfile(t *testing.T) {
	ctx := context.Background()
	drv := &mockDriver{}
	pppGW := &mockPPPGateway{}

	prov := NewWithResolver(
		func(_ context.Context, _ string) (port.DeviceDriver, error) {
			return drv, nil
		},
		pppGW, nil, nil, nil,
	)

	plan := domainPlan.ServicePlan{
		ID:                    "p1",
		Name:                  "PLAN-20M",
		ServiceType:           "PPPOE",
		BandwidthDownloadKbps: 20000,
		BandwidthUploadKbps:   20000,
		PPPoE: &domainPlan.PPPoEPlanConfig{
			RemoteAddressPool: "pool-pppoe",
		},
	}

	if err := prov.SyncPlanProfile(ctx, "dev1", plan); err != nil {
		t.Fatalf("SyncPlanProfile failed: %v", err)
	}
	if len(pppGW.profiles) != 1 || pppGW.profiles[0].Name != "PLAN-20M" {
		t.Fatalf("PPP profile not created: %+v", pppGW.profiles)
	}
}

// ─── Mock hotspot gateway ───────────────────────────────────────────────

type mockHotspotGateway struct {
	port.HotspotGateway
	users              []port.HotspotUser
	profiles           []port.HotspotUserProfile
	seq                int
	walledGardenRemove int
}

func (m *mockHotspotGateway) ListUsers(_ context.Context, _ port.DeviceDriver, _ port.ListUsersFilter) ([]port.HotspotUser, error) {
	return m.users, nil
}

func (m *mockHotspotGateway) AddUser(_ context.Context, _ port.DeviceDriver, p port.HotspotUserParams) (command.Result, error) {
	m.seq++
	m.users = append(m.users, port.HotspotUser{
		RosID: "*h" + strconv.Itoa(m.seq), Name: p.Name, Profile: p.Profile, Disabled: p.Disabled,
	})
	return command.Result{Output: "ok"}, nil
}

func (m *mockHotspotGateway) UpdateUser(_ context.Context, _ port.DeviceDriver, rosID string, p port.HotspotUserParams) (command.Result, error) {
	for i := range m.users {
		if m.users[i].RosID == rosID {
			m.users[i].Profile = p.Profile
			m.users[i].Disabled = p.Disabled
		}
	}
	return command.Result{Output: "ok"}, nil
}

func (m *mockHotspotGateway) ListActiveSessions(_ context.Context, _ port.DeviceDriver) ([]port.HotspotActiveSession, error) {
	return nil, nil
}

func (m *mockHotspotGateway) ListCookies(_ context.Context, _ port.DeviceDriver) ([]port.HotspotCookie, error) {
	return nil, nil
}

func (m *mockHotspotGateway) GetUserProfiles(_ context.Context, _ port.DeviceDriver) ([]port.HotspotUserProfile, error) {
	return m.profiles, nil
}

func (m *mockHotspotGateway) CreateUserProfile(_ context.Context, _ port.DeviceDriver, p port.MikhmonProfileParams) (command.Result, error) {
	m.profiles = append(m.profiles, port.HotspotUserProfile{Name: p.Name})
	return command.Result{Output: "ok"}, nil
}

func (m *mockHotspotGateway) DeleteUserProfile(_ context.Context, _ port.DeviceDriver, rosID string) (command.Result, error) {
	filtered := m.profiles[:0]
	for _, pr := range m.profiles {
		if pr.RosID != rosID {
			filtered = append(filtered, pr)
		}
	}
	m.profiles = filtered
	return command.Result{Output: "ok"}, nil
}

func (m *mockHotspotGateway) RemoveWalledGarden(_ context.Context, _ port.DeviceDriver) error {
	m.walledGardenRemove++
	return nil
}

// ─── F1-7: resume PPP harus re-enable secret ────────────────────────────

func TestProvisioner_PPPoEResumeReEnablesSecret(t *testing.T) {
	ctx := context.Background()
	pppGW := &mockPPPGateway{}
	prov := NewWithResolver(
		func(_ context.Context, _ string) (port.DeviceDriver, error) { return &mockDriver{}, nil },
		pppGW, nil, nil, nil,
	)

	acct := port.SubscriberAccount{Username: "user1", Password: "pass1", Profile: "PLAN-10M"}
	if err := prov.Provision(ctx, "dev1", "PPPOE", acct); err != nil {
		t.Fatalf("Provision: %v", err)
	}
	if err := prov.Suspend(ctx, "dev1", "PPPOE", "user1"); err != nil {
		t.Fatalf("Suspend: %v", err)
	}
	if !pppGW.secrets[0].Disabled {
		t.Fatalf("secret harus disabled setelah suspend")
	}
	if err := prov.Restore(ctx, "dev1", "PPPOE", "user1", "PLAN-10M", ""); err != nil {
		t.Fatalf("Restore: %v", err)
	}
	if pppGW.secrets[0].Disabled {
		t.Fatalf("secret masih disabled setelah restore — pelanggan tidak akan bisa dial")
	}
}

// ─── F1-6: suspend/restore hotspot harus membersihkan disabled ──────────

func TestProvisioner_HotspotRestoreClearsDisabled(t *testing.T) {
	ctx := context.Background()
	hotGW := &mockHotspotGateway{}
	prov := NewWithResolver(
		func(_ context.Context, _ string) (port.DeviceDriver, error) { return &mockDriver{}, nil },
		nil, hotGW, nil, nil,
	)

	acct := port.SubscriberAccount{Username: "hs1", Password: "pw", Profile: "PLAN-HS"}
	if err := prov.Provision(ctx, "dev1", "HOTSPOT", acct); err != nil {
		t.Fatalf("Provision: %v", err)
	}
	if err := prov.Suspend(ctx, "dev1", "HOTSPOT", "hs1"); err != nil {
		t.Fatalf("Suspend: %v", err)
	}
	if !hotGW.users[0].Disabled {
		t.Fatalf("user hotspot harus disabled setelah suspend")
	}
	if err := prov.Restore(ctx, "dev1", "HOTSPOT", "hs1", "PLAN-HS", ""); err != nil {
		t.Fatalf("Restore: %v", err)
	}
	if hotGW.users[0].Disabled {
		t.Fatalf("user hotspot masih disabled setelah restore")
	}
}

// ─── F1-8: OnPaid hanya untuk invoice lunas ─────────────────────────────

func seedIsolatedSub(t *testing.T, subs *mocktest.FakeSubscriptionRepo) {
	t.Helper()
	dev := "dev-1"
	subs.Seed(domainSubscription.Subscription{
		ID: "sub-1", PlanID: "plan-1", DeviceID: &dev,
		ServiceType: "PPPOE", RemoteUsername: "bs1234",
		RouterProfile: "PLAN-10M", Status: domainSubscription.StatusIsolated,
	})
}

func TestBuildOnPaidRestore_RestoresIsolatedWhenInvoicePaid(t *testing.T) {
	ctx := context.Background()
	mgr := mocktest.NewFakeRouterAccountManager()
	subs := mocktest.NewFakeSubscriptionRepo()
	planRepo := mocktest.NewFakeServicePlanRepo()
	settings := mocktest.NewFakeSettingReader(nil)
	seedIsolatedSub(t, subs)

	subID := "sub-1"
	hook := BuildOnPaidRestore(mgr, subs, planRepo, settings)
	hook(ctx, domainBilling.Invoice{
		ID: "inv-1", SubscriptionID: &subID, Status: domainBilling.StatusPaid,
	}, domainBilling.Payment{ID: "pay-1"})

	if got := mgr.Count("Restore:"); got != 1 {
		t.Fatalf("Restore dipanggil %d kali, mau 1", got)
	}
	sub, err := subs.FindByID(ctx, "sub-1")
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if sub.Status != domainSubscription.StatusActive {
		t.Fatalf("status subscription = %s, mau ACTIVE", sub.Status)
	}
}

func TestBuildOnPaidRestore_FailureMarksProvisionFailed(t *testing.T) {
	ctx := context.Background()
	mgr := &mocktest.FakeRouterAccountManager{Fail: map[string]error{
		"Restore:": errors.New("router unreachable"),
	}}
	subs := mocktest.NewFakeSubscriptionRepo()
	planRepo := mocktest.NewFakeServicePlanRepo()
	settings := mocktest.NewFakeSettingReader(nil)
	seedIsolatedSub(t, subs)

	subID := "sub-1"
	hook := BuildOnPaidRestore(mgr, subs, planRepo, settings)
	hook(ctx, domainBilling.Invoice{
		ID: "inv-1", SubscriptionID: &subID, Status: domainBilling.StatusPaid,
	}, domainBilling.Payment{ID: "pay-1"})

	sub, err := subs.FindByID(ctx, "sub-1")
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if sub.ProvisionStatus != domainSubscription.ProvisionFailed {
		t.Fatalf("provision_status = %s, mau FAILED agar worker retry", sub.ProvisionStatus)
	}
	if sub.Status != domainSubscription.StatusIsolated {
		t.Fatalf("status = %s, harus tetap ISOLATED sampai restore sukses", sub.Status)
	}
}

func TestBuildOnPaidRestore_SkipsPartialInvoice(t *testing.T) {
	ctx := context.Background()
	mgr := mocktest.NewFakeRouterAccountManager()
	subs := mocktest.NewFakeSubscriptionRepo()
	planRepo := mocktest.NewFakeServicePlanRepo()
	settings := mocktest.NewFakeSettingReader(nil)
	seedIsolatedSub(t, subs)

	subID := "sub-1"
	hook := BuildOnPaidRestore(mgr, subs, planRepo, settings)
	hook(ctx, domainBilling.Invoice{
		ID: "inv-1", SubscriptionID: &subID, Status: domainBilling.StatusPartial,
	}, domainBilling.Payment{ID: "pay-1"})

	if got := mgr.Count("Restore:"); got != 0 {
		t.Fatalf("Restore dipanggil %d kali untuk pembayaran parsial", got)
	}
	sub, err := subs.FindByID(ctx, "sub-1")
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if sub.Status != domainSubscription.StatusIsolated {
		t.Fatalf("status subscription = %s, harus tetap ISOLATED", sub.Status)
	}
}

// ─── F6-7: isolir otomatis memasang filter drop ─────────────────────────

func TestProvisioner_IsolateInstallsFilter(t *testing.T) {
	ctx := context.Background()
	pppGW := &mockPPPGateway{}
	fwGW := &mockFirewallGateway{}
	prov := NewWithResolver(
		func(_ context.Context, _ string) (port.DeviceDriver, error) { return &mockDriver{}, nil },
		pppGW, nil, fwGW, nil,
	)
	acct := port.SubscriberAccount{Username: "user1", Password: "pass1", Profile: "PLAN-10M"}
	if err := prov.Provision(ctx, "dev1", "PPPOE", acct); err != nil {
		t.Fatalf("Provision: %v", err)
	}
	opt := port.IsolationOptions{
		IsolirProfile: "isolir",
		AddressList:   "ISOLIR_USERS",
		Redirect: &port.IsolationRedirectConfig{
			SrcAddressList: "ISOLIR_USERS", PaymentHost: "bayar.example.com", PaymentPort: "8080",
		},
	}
	if err := prov.Isolate(ctx, "dev1", "PPPOE", "user1", opt); err != nil {
		t.Fatalf("Isolate: %v", err)
	}
	if fwGW.filterCalls != 1 {
		t.Fatalf("EnsureIsolationFilter dipanggil %d kali, mau 1", fwGW.filterCalls)
	}
}

// ─── F6-6: hapus infrastruktur isolir ───────────────────────────────────

func TestProvisioner_DeleteIsolationInfrastructure(t *testing.T) {
	ctx := context.Background()
	pppGW := &mockPPPGateway{profiles: []port.PPPProfile{{RosID: "*p1", Name: "ISOLIR"}}}
	hotGW := &mockHotspotGateway{profiles: []port.HotspotUserProfile{{RosID: "*h1", Name: "ISOLIR"}}}
	fwGW := &mockFirewallGateway{}
	prov := NewWithResolver(
		func(_ context.Context, _ string) (port.DeviceDriver, error) { return &mockDriver{}, nil },
		pppGW, hotGW, fwGW, nil,
	)
	cfg := domainDevice.DefaultIsolationConfig()
	if err := prov.DeleteIsolationInfrastructure(ctx, "dev1", cfg, true); err != nil {
		t.Fatalf("DeleteIsolationInfrastructure: %v", err)
	}
	if len(pppGW.profiles) != 0 {
		t.Fatalf("profil PPP isolir belum terhapus: %+v", pppGW.profiles)
	}
	if len(hotGW.profiles) != 0 {
		t.Fatalf("profil hotspot isolir belum terhapus: %+v", hotGW.profiles)
	}
	if fwGW.disableRedirectCalls != 1 || fwGW.flushCalls != 1 {
		t.Fatalf("cleanup firewall tidak lengkap: disable=%d flush=%d", fwGW.disableRedirectCalls, fwGW.flushCalls)
	}
	if hotGW.walledGardenRemove != 1 {
		t.Fatalf("walled-garden tidak dibersihkan")
	}
}

// ─── F6-3: cleanup address-list saat terminate ──────────────────────────

func TestProvisioner_CleanupIsolationAddressList(t *testing.T) {
	ctx := context.Background()
	fwGW := &mockFirewallGateway{}
	prov := NewWithResolver(
		func(_ context.Context, _ string) (port.DeviceDriver, error) { return &mockDriver{}, nil },
		nil, nil, fwGW, nil,
	)
	if err := prov.CleanupIsolationAddressList(ctx, "dev1", "ISOLIR_USERS", "bs1234"); err != nil {
		t.Fatalf("CleanupIsolationAddressList: %v", err)
	}
	if len(fwGW.removedByComment) != 1 || fwGW.removedByComment[0] != "ISOLIR_USERS:isolir:bs1234" {
		t.Fatalf("address-list tidak dibersihkan: %+v", fwGW.removedByComment)
	}

	// No-op untuk argumen kosong.
	if err := prov.CleanupIsolationAddressList(ctx, "", "L", "u"); err != nil {
		t.Fatalf("empty device harus no-op: %v", err)
	}
}
