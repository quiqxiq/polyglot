package importer_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/internal/usecase/importer"
)

type mockPPPGateway struct {
	port.PPPGateway
	listSecretsFn func(ctx context.Context, driver port.DeviceDriver, service string) ([]port.PPPoESecret, error)
}

func (m *mockPPPGateway) ListSecrets(ctx context.Context, driver port.DeviceDriver, service string) ([]port.PPPoESecret, error) {
	if m.listSecretsFn != nil {
		return m.listSecretsFn(ctx, driver, service)
	}
	return nil, nil
}

type mockHotspotGateway struct {
	port.HotspotGateway
	getUserProfilesFn func(ctx context.Context, driver port.DeviceDriver) ([]port.HotspotUserProfile, error)
	listIPBindingsFn  func(ctx context.Context, driver port.DeviceDriver) ([]port.HotspotIPBinding, error)
	listUsersFn       func(ctx context.Context, driver port.DeviceDriver, f port.ListUsersFilter) ([]port.HotspotUser, error)
}

func (m *mockHotspotGateway) GetUserProfiles(ctx context.Context, driver port.DeviceDriver) ([]port.HotspotUserProfile, error) {
	if m.getUserProfilesFn != nil {
		return m.getUserProfilesFn(ctx, driver)
	}
	return nil, nil
}

func (m *mockHotspotGateway) ListIPBindings(ctx context.Context, driver port.DeviceDriver) ([]port.HotspotIPBinding, error) {
	if m.listIPBindingsFn != nil {
		return m.listIPBindingsFn(ctx, driver)
	}
	return nil, nil
}

func (m *mockHotspotGateway) ListUsers(ctx context.Context, driver port.DeviceDriver, f port.ListUsersFilter) ([]port.HotspotUser, error) {
	if m.listUsersFn != nil {
		return m.listUsersFn(ctx, driver, f)
	}
	return nil, nil
}

func TestIsVoucherProfile(t *testing.T) {
	// Profil voucher Mikhmon: ada script on-login
	voucherProfile1 := port.HotspotUserProfile{
		Name:    "Paket-2Jam",
		OnLogin: `/tool fetch mode=http address=192.168.88.1 src-path=("/mikhmon/status.php?user=".$user) dst-path=test.txt`,
	}
	assert.True(t, importer.IsVoucherProfile(voucherProfile1))

	voucherProfile2 := port.HotspotUserProfile{
		Name:    "Voucher-1Hari",
		Comment: "validity=1d,price=5000",
	}
	assert.True(t, importer.IsVoucherProfile(voucherProfile2))

	// Profil permanen: on-login kosong
	permProfile := port.HotspotUserProfile{
		Name:    "Member-Bulanan",
		OnLogin: "",
		Comment: "Pelanggan Tetap Kos",
	}
	assert.False(t, importer.IsVoucherProfile(permProfile))
}

func TestPullHotspotRows_PermanentVsVoucher(t *testing.T) {
	ctx := context.Background()
	hgw := &mockHotspotGateway{
		getUserProfilesFn: func(ctx context.Context, driver port.DeviceDriver) ([]port.HotspotUserProfile, error) {
			return []port.HotspotUserProfile{
				{Name: "Paket-2Jam", OnLogin: `/tool fetch mode=http ... mikhmon`},
				{Name: "Member-Bulanan", OnLogin: ""},
			}, nil
		},
		listIPBindingsFn: func(ctx context.Context, driver port.DeviceDriver) ([]port.HotspotIPBinding, error) {
			return []port.HotspotIPBinding{
				{
					MACAddress: "AA:BB:CC:DD:EE:FF",
					Address:    "192.168.88.50",
					ToAddress:  "192.168.88.50",
					Type:       "bypassed",
					Comment:    "Warnet RT (Bypass)",
				},
			}, nil
		},
		listUsersFn: func(ctx context.Context, driver port.DeviceDriver, f port.ListUsersFilter) ([]port.HotspotUser, error) {
			return []port.HotspotUser{
				// User voucher sementara
				{
					Name:        "a89k",
					Profile:     "Paket-2Jam",
					LimitUptime: "02:00:00",
					Comment:     "vc-2jam-01",
				},
				// User member permanen
				{
					Name:        "budi-member",
					Password:    "rahasia123",
					Profile:     "Member-Bulanan",
					LimitUptime: "",
					Comment:     "Rp 100.000 08123456789",
				},
			}, nil
		},
	}

	src := importer.NewRouterSource(nil, hgw)

	// Uji 1: includeVouchers = false
	rows, permCount, ipbCount, vSkipped, err := src.PullHotspotRows(ctx, nil, "ROUTER-UTAMA", true, false)
	require.NoError(t, err)
	assert.Equal(t, 1, permCount, "harus ada 1 member permanen")
	assert.Equal(t, 1, ipbCount, "harus ada 1 IP binding bypass")
	assert.Equal(t, 1, vSkipped, "harus ada 1 voucher yang dilewati")
	require.Len(t, rows, 2, "total baris permanen adalah 2 (1 IP binding + 1 member)")

	// Verifikasi entri IP binding
	assert.Equal(t, "Warnet RT (Bypass)", rows[0].Name)
	assert.Equal(t, "HOTSPOT", rows[0].ServiceType)
	assert.Equal(t, "IP_BINDING", rows[0].HotspotType)
	assert.Equal(t, "AA:BB:CC:DD:EE:FF", rows[0].MACAddress)

	// Verifikasi entri user permanen
	assert.Equal(t, "budi-member", rows[1].Name)
	assert.Equal(t, "Member-Bulanan", rows[1].PlanName)
	assert.Equal(t, "08123456789", rows[1].Phone)
	assert.InDelta(t, 100000, rows[1].Price, 0.01)
	assert.Equal(t, "PERMANENT_USER", rows[1].HotspotType)
}

func TestPullRouterRows_Unified(t *testing.T) {
	ctx := context.Background()
	pgw := &mockPPPGateway{
		listSecretsFn: func(ctx context.Context, driver port.DeviceDriver, service string) ([]port.PPPoESecret, error) {
			return []port.PPPoESecret{
				{Name: "user-pppoe", Profile: "10Mbps", Comment: "Rp 150.000 0857123456"},
			}, nil
		},
	}
	hgw := &mockHotspotGateway{
		getUserProfilesFn: func(ctx context.Context, driver port.DeviceDriver) ([]port.HotspotUserProfile, error) {
			return []port.HotspotUserProfile{
				{Name: "Hotspot-Member", OnLogin: ""},
			}, nil
		},
		listIPBindingsFn: func(ctx context.Context, driver port.DeviceDriver) ([]port.HotspotIPBinding, error) {
			return nil, nil
		},
		listUsersFn: func(ctx context.Context, driver port.DeviceDriver, f port.ListUsersFilter) ([]port.HotspotUser, error) {
			return []port.HotspotUser{
				{Name: "user-hotspot", Profile: "Hotspot-Member", Comment: "Rp 50.000"},
			}, nil
		},
	}

	src := importer.NewRouterSource(pgw, hgw)

	// Tarik ALL
	res, err := src.PullRouterRows(ctx, nil, "ROUTER-1", importer.PullOptions{
		ServiceType:       "ALL",
		IncludeIPBindings: false,
		IncludeVouchers:   false,
	})
	require.NoError(t, err)
	assert.Equal(t, 1, res.PPPoEDetected)
	assert.Equal(t, 1, res.HotspotPermanentDetected)
	require.Len(t, res.Rows, 2)
	assert.Equal(t, "PPPOE", res.Rows[0].ServiceType)
	assert.Equal(t, "HOTSPOT", res.Rows[1].ServiceType)
}
