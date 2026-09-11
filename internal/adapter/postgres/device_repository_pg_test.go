// F5-9: verifikasi mapping kolom produksi `extra` JSONB & `tags` TEXT[]
// terhadap PostgreSQL nyata (bukan hanya SQLite test).
package postgres_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres"
	"github.com/quixiq/polyglot/internal/domain/device"
)

func TestDeviceRepository_PostgresJSONBRoundTrip(t *testing.T) {
	if testing.Short() {
		t.Skip("short mode")
	}
	_, _, dsn := startMigratedPostgres(t)
	gormDB, err := gorm.Open(gormpostgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	repo := postgres.NewDeviceRepository(gormDB)
	vault := postgres.NewCredentialVault(gormDB, testEncryptionKey)
	ctx := context.Background()

	dev := device.Device{
		ID: "11111111-1111-1111-1111-111111111111", TenantID: "tenant-default",
		Name: "PG Device", Vendor: "mikrotik", DriverType: "mikrotik", Host: "10.0.0.1",
		Port: 8728, SSHPort: 22, TimeoutMS: 5000, PollIntervalMS: 30000,
		Extra:   map[string]string{"use_tls": "false", "webhook_token": "rtr_pg"},
		Tags:    []string{"core", "pg"},
		Enabled: true,
	}
	require.NoError(t, repo.Save(ctx, dev))
	require.NoError(t, vault.Save(ctx, dev.ID, device.Credentials{
		Username: "admin", Password: "secret",
	}))

	got, err := repo.FindByID(ctx, dev.ID)
	require.NoError(t, err)
	assert.Equal(t, dev.Extra, got.Extra)
	assert.ElementsMatch(t, dev.Tags, got.Tags)

	creds, err := vault.Get(ctx, dev.ID)
	require.NoError(t, err)
	assert.Equal(t, "admin", creds.Username)
	assert.Equal(t, "secret", creds.Password)
}
