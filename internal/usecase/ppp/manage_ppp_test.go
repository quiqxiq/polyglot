package ppp_test

import (
	"context"
	"testing"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/internal/usecase/ppp"
)

type mockPPPGateway struct {
	secrets  []port.PPPoESecret
	profiles []port.PPPProfile
	active   []port.PPPActiveSession
}

func (m *mockPPPGateway) ListSecrets(ctx context.Context, driver port.DeviceDriver, nameFilter string) ([]port.PPPoESecret, error) {
	return m.secrets, nil
}

func (m *mockPPPGateway) GetSecret(ctx context.Context, driver port.DeviceDriver, rosID string) (port.PPPoESecret, error) {
	for _, s := range m.secrets {
		if s.RosID == rosID {
			return s, nil
		}
	}
	return port.PPPoESecret{}, nil
}

func (m *mockPPPGateway) AddSecret(ctx context.Context, driver port.DeviceDriver, p port.PPPoESecretParams) (command.Result, error) {
	m.secrets = append(m.secrets, port.PPPoESecret{
		RosID:   "*1",
		Name:    p.Name,
		Profile: p.Profile,
		Service: p.Service,
	})
	return command.Result{Output: "added"}, nil
}

func (m *mockPPPGateway) UpdateSecret(ctx context.Context, driver port.DeviceDriver, rosID string, p port.PPPoESecretParams) (command.Result, error) {
	return command.Result{Output: "updated"}, nil
}

func (m *mockPPPGateway) RemoveSecret(ctx context.Context, driver port.DeviceDriver, rosID string) (command.Result, error) {
	return command.Result{Output: "removed"}, nil
}

func (m *mockPPPGateway) SetSecretDisabled(ctx context.Context, driver port.DeviceDriver, rosID string, disabled bool) (command.Result, error) {
	return command.Result{Output: "set"}, nil
}

func (m *mockPPPGateway) ListProfiles(ctx context.Context, driver port.DeviceDriver, nameFilter string) ([]port.PPPProfile, error) {
	return m.profiles, nil
}

func (m *mockPPPGateway) GetProfile(ctx context.Context, driver port.DeviceDriver, rosID string) (port.PPPProfile, error) {
	for _, p := range m.profiles {
		if p.RosID == rosID {
			return p, nil
		}
	}
	return port.PPPProfile{}, nil
}

func (m *mockPPPGateway) AddProfile(ctx context.Context, driver port.DeviceDriver, p port.PPPProfileParams) (command.Result, error) {
	m.profiles = append(m.profiles, port.PPPProfile{
		RosID:     "*p1",
		Name:      p.Name,
		RateLimit: p.RateLimit,
	})
	return command.Result{Output: "added"}, nil
}

func (m *mockPPPGateway) UpdateProfile(ctx context.Context, driver port.DeviceDriver, rosID string, p port.PPPProfileParams) (command.Result, error) {
	return command.Result{Output: "updated"}, nil
}

func (m *mockPPPGateway) RemoveProfile(ctx context.Context, driver port.DeviceDriver, rosID string) (command.Result, error) {
	return command.Result{Output: "removed"}, nil
}

func (m *mockPPPGateway) ListActive(ctx context.Context, driver port.DeviceDriver, nameFilter string) ([]port.PPPActiveSession, error) {
	return m.active, nil
}

func (m *mockPPPGateway) KickActive(ctx context.Context, driver port.DeviceDriver, rosID string) (command.Result, error) {
	return command.Result{Output: "kicked"}, nil
}

func (m *mockPPPGateway) ListInactive(ctx context.Context, driver port.DeviceDriver) ([]port.PPPoESecret, error) {
	return m.secrets, nil
}

func TestPPPUseCase_Secrets(t *testing.T) {
	mockGW := &mockPPPGateway{
		secrets: []port.PPPoESecret{
			{RosID: "*1", Name: "client1", Profile: "10M", Service: "pppoe"},
		},
	}
	uc := ppp.New(mockGW)
	ctx := context.Background()

	secrets, err := uc.ListSecrets(ctx, nil, "")
	if err != nil || len(secrets) != 1 {
		t.Fatalf("expected 1 secret, got %v (err: %v)", len(secrets), err)
	}

	res, err := uc.AddSecret(ctx, nil, port.PPPoESecretParams{
		Name:    "client2",
		Profile: "20M",
		Service: "pppoe",
	})
	if err != nil || res.Output != "added" {
		t.Fatalf("failed to add secret: %v", err)
	}

	if _, err := uc.AddSecret(ctx, nil, port.PPPoESecretParams{}); err == nil {
		t.Fatalf("expected error for empty secret name, got nil")
	}

	kicked, err := uc.KickActiveBatch(ctx, nil, []string{"*a1", "*a2"})
	if err != nil || kicked != 2 {
		t.Fatalf("expected 2 kicked sessions, got %d (err: %v)", kicked, err)
	}
}
