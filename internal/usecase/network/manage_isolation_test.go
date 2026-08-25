package network

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quixiq/polyglot/internal/port"
)

func TestManageIsolationUseCase_Setup(t *testing.T) {
	ctx := context.Background()
	isolir := &fakeIsolir{}
	uc := NewManageIsolationUseCase(
		&fakeSettingRepo{values: map[string]string{
			"isolir.pool_range": "172.16.99.10-172.16.99.254",
		}},
		isolir,
		func(ctx context.Context, deviceID string) (port.DeviceDriver, error) { return nil, nil },
	)

	res, cfg, err := uc.Setup(ctx, testDeviceID, IsolirConfigOverride{PoolRange: "10.10.0.1-10.10.0.254", PortalIP: "192.0.2.10"})
	require.NoError(t, err)

	assert.Equal(t, 1, isolir.ensureCalls)
	assert.Len(t, res.NATRuleIDs, 2)
	assert.Equal(t, "10.10.0.1-10.10.0.254", cfg.PoolRange, "override menang atas settings")
	assert.Equal(t, "192.0.2.10", cfg.PortalIP)
	assert.Equal(t, "isolir", cfg.ProfileName, "default settings saat kosong")
}

func TestManageIsolationUseCase_SetupDeviceUnreachable(t *testing.T) {
	ctx := context.Background()
	uc := NewManageIsolationUseCase(
		&fakeSettingRepo{values: map[string]string{}},
		&fakeIsolir{},
		func(ctx context.Context, deviceID string) (port.DeviceDriver, error) {
			return nil, errors.New("device offline")
		},
	)

	_, _, err := uc.Setup(ctx, testDeviceID, IsolirConfigOverride{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "device offline")
}

func TestManageIsolationUseCase_StatusWarnings(t *testing.T) {
	ctx := context.Background()
	isolir := &fakeIsolir{inspection: port.IsolirInspection{
		PoolExists:       false,
		PoolName:         "pool-isolir",
		ProfileExists:    true,
		ProfileName:      "isolir",
		ProfileRateLimit: "512k/512k",
		NATRules: []port.IsolirNATRuleStatus{
			{Port: "80", Exists: true, RosID: "*A", Action: "dst-nat", ToAddresses: "192.0.2.10"},
			{Port: "443", Exists: false},
		},
	}}
	uc := NewManageIsolationUseCase(
		&fakeSettingRepo{values: map[string]string{}},
		isolir,
		func(ctx context.Context, deviceID string) (port.DeviceDriver, error) { return nil, nil },
	)

	ins, cfg, warnings, err := uc.Status(ctx, testDeviceID)
	require.NoError(t, err)
	assert.Equal(t, 1, isolir.inspectCalls)
	assert.False(t, ins.PoolExists)
	assert.Equal(t, "pool-isolir", cfg.PoolName)

	var poolWarn, ruleWarn, portalWarn bool
	for _, w := range warnings {
		switch {
		case containsStr(w, "pool"):
			poolWarn = true
		case containsStr(w, "443"):
			ruleWarn = true
		case containsStr(w, "portal_ip"):
			portalWarn = true
		}
	}
	assert.True(t, poolWarn, "pool hilang harus di-warning")
	assert.True(t, ruleWarn, "rule 443 hilang harus di-warning")
	assert.True(t, portalWarn, "portal_ip kosong harus di-warning")
}

func TestManageIsolationUseCase_Remove(t *testing.T) {
	ctx := context.Background()
	isolir := &fakeIsolir{}
	uc := NewManageIsolationUseCase(
		&fakeSettingRepo{values: map[string]string{"isolir.redirect_ports": "80,443"}},
		isolir,
		func(ctx context.Context, deviceID string) (port.DeviceDriver, error) { return nil, nil },
	)

	require.NoError(t, uc.Remove(ctx, testDeviceID))
	require.Len(t, isolir.removedCfg, 1)
	assert.Equal(t, "80,443", isolir.removedCfg[0].RedirectPorts)
}

func containsStr(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
