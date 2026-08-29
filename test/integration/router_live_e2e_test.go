//go:build integration

package integration

// E2E: tarik data PPP secret LANGSUNG dari MikroTik sungguhan (.env) lalu
// upsert ke sqlite staging. Router hanya DIBACA — tidak ada perubahan state.

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres"
	"github.com/quixiq/polyglot/internal/adapter/postgres/model"
	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/driver/mikrotik"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/internal/usecase/importer"
	"github.com/quixiq/polyglot/internal/usecase/network"
)

type integrationVault struct{}

func (integrationVault) EncryptString(_ context.Context, plaintext string) (string, error) {
	return "enc:" + plaintext, nil
}

func (integrationVault) DecryptString(_ context.Context, ciphertext string) (string, error) {
	if strings.HasPrefix(ciphertext, "enc:") {
		return strings.TrimPrefix(ciphertext, "enc:"), nil
	}
	return "", errors.New("not encrypted")
}

func (integrationVault) Get(context.Context, string) (device.Credentials, error) {
	return device.Credentials{}, device.ErrNotFound
}

func (integrationVault) Save(context.Context, string, device.Credentials) error { return nil }

var networkExecutePreApproved = func(ctx context.Context, driver port.DeviceDriver, cmd command.Command) (command.Result, error) {
	return network.ExecuteCommandPreApproved(ctx, driver, cmd)
}

func TestImportFromRouter_E2E(t *testing.T) {
	if testing.Short() {
		t.Skip("short mode")
	}
	_ = godotenv.Load(filepath.Join("..", "..", ".env"))
	host := os.Getenv("MIKROTIK_TEST_HOST")
	if host == "" {
		t.Skip("MIKROTIK_TEST_HOST tidak diset — lewati E2E import router")
	}
	p := 8728
	if v := os.Getenv("MIKROTIK_TEST_PORT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			p = n
		}
	}
	ctx := context.Background()
	drv, err := mikrotik.NewDriver(ctx, device.Target{
		Host: host, Port: p,
		Username: os.Getenv("MIKROTIK_TEST_USER"),
		Password: os.Getenv("MIKROTIK_TEST_PASS"),
		Timeout:  15 * time.Second,
	})
	if err != nil && strings.Contains(strings.ToLower(err.Error()), "connect") {
		t.Skipf("router unreachable (%v) — lewati E2E", err)
	}
	require.NoError(t, err, "connect router")

	gw := mikrotik.NewGateway(networkExecutePreApproved)
	src := importer.NewRouterSource(gw)

	rows, err := src.PullPPPoERows(ctx, drv, "E2E-ROUTER")
	require.NoError(t, err, "pull secrets dari router")
	t.Logf("router mengembalikan %d baris secret", len(rows))

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&model.ServicePlanModel{}, &model.CustomerModel{},
		&model.SubscriptionModel{},
	))
	plansRepo := postgres.NewServicePlanRepository(db)
	customsRepo := postgres.NewCustomerRepository(db)
	subsRepo := postgres.NewSubscriptionRepository(db, integrationVault{})
	upsert := importer.NewUpsertUseCase(plansRepo, customsRepo, subsRepo, nil, "e2e-device-id")

	res, err := upsert.Import(ctx, rows)
	require.NoError(t, err)
	assert.Equal(t, len(rows), res.RowsTotal)
	t.Logf("impor selesai: customers +%d subs +%d plans +%d",
		res.CustomersCreated, res.SubsCreated, res.PlansCreated)

	if len(rows) > 0 {
		all, ferr := subsRepo.FindAll(ctx)
		require.NoError(t, ferr)
		require.NotEmpty(t, all)
		for _, s := range all {
			assert.Equal(t, "OK", s.ProvisionStatus,
				"data dari router langsung = provisioned OK")
			break // cukup satu sampel
		}
	}
}
