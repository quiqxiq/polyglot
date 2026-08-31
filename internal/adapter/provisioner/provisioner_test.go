package provisioner

import (
	"context"
	"testing"

	"github.com/quixiq/polyglot/internal/domain/command"
	domainPlan "github.com/quixiq/polyglot/internal/domain/plan"
	"github.com/quixiq/polyglot/internal/port"
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

func (m *mockPPPGateway) KickActive(_ context.Context, _ port.DeviceDriver, _ string) (command.Result, error) {
	return command.Result{Output: "ok"}, nil
}

type mockFirewallGateway struct {
	port.FirewallGateway
	addressList []string
}

func (m *mockFirewallGateway) AddToAddressList(_ context.Context, _ port.DeviceDriver, list, address, _ string) error {
	m.addressList = append(m.addressList, list+":"+address)
	return nil
}

func (m *mockFirewallGateway) RemoveFromAddressListByComment(_ context.Context, _ port.DeviceDriver, _, _ string) error {
	return nil
}

func (m *mockFirewallGateway) EnsureIsolationRedirect(_ context.Context, _ port.DeviceDriver, _ port.IsolationRedirectConfig) error {
	return nil
}

func (m *mockFirewallGateway) EnsureIsolationFilter(_ context.Context, _ port.DeviceDriver, _, _ string) error {
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
