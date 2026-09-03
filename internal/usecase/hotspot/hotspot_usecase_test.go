package hotspot

import (
	"context"
	"testing"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockDriver struct{}

func (m *mockDriver) Execute(ctx context.Context, cmd command.Command) (command.Result, error) {
	return command.Result{}, nil
}

func (m *mockDriver) Classify(cmd command.Command) command.Class {
	return command.ClassReadOnly
}

func (m *mockDriver) Translate(op command.Operation) (command.Command, error) {
	return command.Command{}, nil
}

func (m *mockDriver) Close() error {
	return nil
}

type mockHotspotGateway struct {
	port.HotspotGateway
	createUserProfileFn  func(ctx context.Context, driver port.DeviceDriver, p port.MikhmonProfileParams) (command.Result, error)
	generateVouchersFn   func(ctx context.Context, driver port.DeviceDriver, p port.VoucherGenerateParams, count int) (port.VoucherBatch, error)
	getSystemResourceFn  func(ctx context.Context, driver port.DeviceDriver) (port.SystemResource, error)
	getSystemIdentityFn  func(ctx context.Context, driver port.DeviceDriver) (string, error)
	listUsersFn          func(ctx context.Context, driver port.DeviceDriver, f port.ListUsersFilter) ([]port.HotspotUser, error)
	listActiveSessionsFn func(ctx context.Context, driver port.DeviceDriver) ([]port.HotspotActiveSession, error)
	listReportsFn        func(ctx context.Context, driver port.DeviceDriver, f port.ReportFilter) ([]port.MikhmonTransaction, error)
}

func (m *mockHotspotGateway) ListReports(ctx context.Context, driver port.DeviceDriver, f port.ReportFilter) ([]port.MikhmonTransaction, error) {
	if m.listReportsFn != nil {
		return m.listReportsFn(ctx, driver, f)
	}
	return nil, nil
}

func (m *mockHotspotGateway) CreateUserProfile(ctx context.Context, driver port.DeviceDriver, p port.MikhmonProfileParams) (command.Result, error) {
	if m.createUserProfileFn != nil {
		return m.createUserProfileFn(ctx, driver, p)
	}
	return command.Result{}, nil
}

func (m *mockHotspotGateway) GenerateVouchers(ctx context.Context, driver port.DeviceDriver, p port.VoucherGenerateParams, count int) (port.VoucherBatch, error) {
	if m.generateVouchersFn != nil {
		return m.generateVouchersFn(ctx, driver, p, count)
	}
	return port.VoucherBatch{}, nil
}

func (m *mockHotspotGateway) GetSystemResource(ctx context.Context, driver port.DeviceDriver) (port.SystemResource, error) {
	if m.getSystemResourceFn != nil {
		return m.getSystemResourceFn(ctx, driver)
	}
	return port.SystemResource{}, nil
}

func (m *mockHotspotGateway) GetSystemIdentity(ctx context.Context, driver port.DeviceDriver) (string, error) {
	if m.getSystemIdentityFn != nil {
		return m.getSystemIdentityFn(ctx, driver)
	}
	return "", nil
}

func (m *mockHotspotGateway) ListUsers(ctx context.Context, driver port.DeviceDriver, f port.ListUsersFilter) ([]port.HotspotUser, error) {
	if m.listUsersFn != nil {
		return m.listUsersFn(ctx, driver, f)
	}
	return nil, nil
}

func (m *mockHotspotGateway) ListActiveSessions(ctx context.Context, driver port.DeviceDriver) ([]port.HotspotActiveSession, error) {
	if m.listActiveSessionsFn != nil {
		return m.listActiveSessionsFn(ctx, driver)
	}
	return nil, nil
}

func TestHotspotUseCase(t *testing.T) {
	ctx := context.Background()
	driver := &mockDriver{}

	t.Run("CreateProfile", func(t *testing.T) {
		called := false
		gw := &mockHotspotGateway{
			createUserProfileFn: func(ctx context.Context, drv port.DeviceDriver, p port.MikhmonProfileParams) (command.Result, error) {
				assert.Equal(t, "1Day_10K", p.Name)
				assert.Equal(t, "10000", p.Price)
				called = true
				return command.Result{}, nil
			},
		}
		uc := New("", gw)
		_, err := uc.CreateProfile(ctx, driver, port.MikhmonProfileParams{
			Name:  "1Day_10K",
			Price: "10000",
		})
		require.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("GenerateVouchers", func(t *testing.T) {
		gw := &mockHotspotGateway{
			generateVouchersFn: func(ctx context.Context, drv port.DeviceDriver, p port.VoucherGenerateParams, count int) (port.VoucherBatch, error) {
				assert.Equal(t, "1Day_10K", p.Profile)
				assert.Equal(t, 3, count)
				return port.VoucherBatch{
					Vouchers: make([]port.GeneratedVoucher, count),
				}, nil
			},
		}
		uc := New("", gw)
		batch, err := uc.GenerateVouchers(ctx, driver, port.VoucherGenerateParams{
			Profile: "1Day_10K",
		}, 3)
		require.NoError(t, err)
		assert.Len(t, batch.Vouchers, 3)
	})

	t.Run("GetDashboardSummary", func(t *testing.T) {
		gw := &mockHotspotGateway{
			getSystemResourceFn: func(ctx context.Context, drv port.DeviceDriver) (port.SystemResource, error) {
				return port.SystemResource{
					CPULoad:   12,
					Uptime:    "1d",
					Version:   "7.10",
					BoardName: "hEX",
				}, nil
			},
			getSystemIdentityFn: func(ctx context.Context, drv port.DeviceDriver) (string, error) {
				return "Router-Hotspot", nil
			},
			listUsersFn: func(ctx context.Context, drv port.DeviceDriver, f port.ListUsersFilter) ([]port.HotspotUser, error) {
				return []port.HotspotUser{
					{RosID: "*1", Name: "user1"},
					{RosID: "*2", Name: "user2"},
				}, nil
			},
			listActiveSessionsFn: func(ctx context.Context, drv port.DeviceDriver) ([]port.HotspotActiveSession, error) {
				return []port.HotspotActiveSession{
					{RosID: "*1", User: "user1"},
				}, nil
			},
		}
		uc := New("", gw)

		summary, err := uc.GetDashboardSummary(ctx, driver)
		require.NoError(t, err)
		assert.Equal(t, 12, summary.CPULoad)
		assert.Equal(t, "Router-Hotspot", summary.Identity)
		assert.Equal(t, 2, summary.TotalUsers)
		assert.Equal(t, 1, summary.ActiveUsers)
	})
}
