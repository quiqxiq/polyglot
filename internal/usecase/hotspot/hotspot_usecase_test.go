package hotspot

import (
	"context"
	"testing"

	"github.com/quixiq/polyglot/internal/domain/command"
	mikhmon "github.com/quixiq/polyglot/internal/driver/mikrotik/hotspot"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockDriver struct {
	executeFn func(ctx context.Context, cmd command.Command) (command.Result, error)
}

func (m *mockDriver) Execute(ctx context.Context, cmd command.Command) (command.Result, error) {
	if m.executeFn != nil {
		return m.executeFn(ctx, cmd)
	}
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

func TestHotspotUseCase(t *testing.T) {
	// Real driver gateway bound to a policy-bypassing executor: the test
	// asserts command construction/parsing through the full seam without
	// needing the policy gate (mock driver classifies everything read-only).
	exec := func(ctx context.Context, driver port.DeviceDriver, cmd command.Command) (command.Result, error) {
		return driver.Execute(ctx, cmd)
	}
	uc := New("", mikhmon.NewGateway(exec))
	ctx := context.Background()

	t.Run("CreateProfile", func(t *testing.T) {
		driver := &mockDriver{
			executeFn: func(ctx context.Context, cmd command.Command) (command.Result, error) {
				assert.Equal(t, "/ip/hotspot/user/profile/add", cmd.Raw)
				assert.Equal(t, "1Day_10K", cmd.Args["name"])
				return command.Result{}, nil
			},
		}
		_, err := uc.CreateProfile(ctx, driver, mikhmon.MikhmonProfileParams{
			Name:  "1Day_10K",
			Price: "10000",
		})
		require.NoError(t, err)
	})

	t.Run("GenerateVouchers", func(t *testing.T) {
		executedCount := 0
		driver := &mockDriver{
			executeFn: func(ctx context.Context, cmd command.Command) (command.Result, error) {
				assert.Equal(t, "/ip/hotspot/user/add", cmd.Raw)
				executedCount++
				return command.Result{}, nil
			},
		}
		batch, err := uc.GenerateVouchers(ctx, driver, mikhmon.VoucherGenerateParams{
			Profile: "1Day_10K",
		}, 3)
		require.NoError(t, err)
		assert.Len(t, batch.Vouchers, 3)
		assert.Equal(t, 3, executedCount)
	})

	t.Run("GetDashboardSummary", func(t *testing.T) {
		driver := &mockDriver{
			executeFn: func(ctx context.Context, cmd command.Command) (command.Result, error) {
				switch cmd.Raw {
				case "/system/resource/print":
					return command.Result{Rows: []map[string]string{
						{"cpu-load": "12", "uptime": "1d", "version": "7.10", "board-name": "hEX"},
					}}, nil
				case "/system/identity/print":
					return command.Result{Rows: []map[string]string{
						{"name": "Router-Hotspot"},
					}}, nil
				case "/ip/hotspot/user/print":
					return command.Result{Rows: []map[string]string{
						{".id": "*1", "name": "user1"},
						{".id": "*2", "name": "user2"},
					}}, nil
				case "/ip/hotspot/active/print":
					return command.Result{Rows: []map[string]string{
						{".id": "*1", "user": "user1"},
					}}, nil
				case "/system/script/print":
					return command.Result{Rows: []map[string]string{}}, nil
				default:
					return command.Result{}, nil
				}
			},
		}

		summary, err := uc.GetDashboardSummary(ctx, driver)
		require.NoError(t, err)
		assert.Equal(t, 12, summary.CPULoad)
		assert.Equal(t, "Router-Hotspot", summary.Identity)
		assert.Equal(t, 2, summary.TotalUsers)
		assert.Equal(t, 1, summary.ActiveUsers)
	})
}
