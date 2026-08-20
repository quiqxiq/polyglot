package postgres_test

import (
	"context"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres"
	"github.com/quixiq/polyglot/internal/adapter/postgres/model"
	"github.com/quixiq/polyglot/internal/domain/device"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.Exec(`CREATE TABLE devices (
		id TEXT PRIMARY KEY,
		tenant_id TEXT NOT NULL DEFAULT 'tenant-default',
		name TEXT NOT NULL,
		vendor TEXT NOT NULL DEFAULT 'mikrotik',
		driver_type TEXT NOT NULL DEFAULT 'mikrotik',
		host TEXT NOT NULL,
		port INTEGER NOT NULL DEFAULT 8728,
		ssh_port INTEGER NOT NULL DEFAULT 22,
		timeout_ms INTEGER NOT NULL DEFAULT 10000,
		poll_interval_ms INTEGER NOT NULL DEFAULT 30000,
		extra_json TEXT,
		tags_json TEXT,
		enabled NUMERIC NOT NULL DEFAULT 1,
		created_at DATETIME,
		updated_at DATETIME
	)`).Error
	require.NoError(t, err)

	err = db.AutoMigrate(&model.CredentialModel{})
	require.NoError(t, err)

	return db
}

func TestDeviceModel_Conversion(t *testing.T) {
	dev := device.Device{
		ID:             "router-1",
		TenantID:       "tenant-default",
		Name:           "Core MikroTik Router",
		Vendor:         "mikrotik",
		DriverType:     "mikrotik",
		Host:           "192.168.1.1",
		Port:           8728,
		TimeoutMS:      10000,
		PollIntervalMS: 30000,
		Extra:          map[string]string{"use_tls": "false"},
		Tags:           []string{"core", "mikrotik"},
		Enabled:        true,
	}

	model := model.DeviceModelFromDomain(dev)
	assert.Equal(t, "router-1", model.ID)
	assert.Equal(t, "Core MikroTik Router", model.Name)

	domainDev := model.ToDomain()
	assert.Equal(t, dev.ID, domainDev.ID)
	assert.Equal(t, dev.Name, domainDev.Name)
	assert.Equal(t, dev.Host, domainDev.Host)
	assert.Equal(t, dev.Tags, domainDev.Tags)
	assert.Equal(t, dev.Extra, domainDev.Extra)
}

func TestCredentialModel_Conversion(t *testing.T) {
	creds := device.Credentials{
		Username: "admin",
		Password: "secretpassword",
		Extra:    map[string]string{"api_key": "xyz123"},
	}

	model, err := model.CredentialModelFromDomain("router-1", creds, "")
	require.NoError(t, err)
	assert.Equal(t, "router-1", model.DeviceID)
	assert.NotEmpty(t, model.Ciphertext)
	assert.NotEmpty(t, model.Nonce)

	domainCreds, err := model.ToDomain("")
	require.NoError(t, err)
	assert.Equal(t, creds.Username, domainCreds.Username)
	assert.Equal(t, creds.Password, domainCreds.Password)
	assert.Equal(t, creds.Extra, domainCreds.Extra)
}

func TestPostgresDeviceRepository_CRUD(t *testing.T) {
	db := setupTestDB(t)
	repo := postgres.NewDeviceRepository(db)
	vault := postgres.NewCredentialVault(db)
	ctx := context.Background()

	dev := device.Device{
		ID:         "router-test-1",
		TenantID:   "tenant-default",
		Name:       "Test MikroTik",
		Vendor:     "mikrotik",
		DriverType: "mikrotik",
		Host:       "192.168.88.1",
		Port:       8728,
		TimeoutMS:  5000,
		Enabled:    true,
	}

	creds := device.Credentials{
		Username: "admin",
		Password: "password123",
	}

	// 1. Save Device & Credentials
	err := repo.Save(ctx, dev)
	require.NoError(t, err)

	err = vault.Save(ctx, dev.ID, creds)
	require.NoError(t, err)

	// 2. FindByID & Get Credentials
	foundDev, err := repo.FindByID(ctx, dev.ID)
	require.NoError(t, err)
	assert.Equal(t, dev.Name, foundDev.Name)

	foundCreds, err := vault.Get(ctx, dev.ID)
	require.NoError(t, err)
	assert.Equal(t, creds.Username, foundCreds.Username)

	// 3. FindAll
	all, err := repo.FindAll(ctx)
	require.NoError(t, err)
	assert.Len(t, all, 1)

	// 4. Delete
	err = repo.Delete(ctx, dev.ID)
	require.NoError(t, err)

	_, err = repo.FindByID(ctx, dev.ID)
	assert.ErrorIs(t, err, device.ErrNotFound)
}
