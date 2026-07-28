package postgres

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/testutil"
)

func TestDeviceRepository_CRUD(t *testing.T) {
	ctx := context.Background()
	db, err := testutil.NewMemoryDB()
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&deviceModel{}))

	repo := NewDeviceRepository(db)

	d := device.Device{
		ID:         "dev-001",
		Name:       "Mikrotik Border",
		Vendor:     "mikrotik",
		DriverType: "mikrotik",
		Host:       "192.168.1.1",
		Port:       8728,
		TimeoutMS:  30000,
		Extra:      map[string]string{"note": "border router"},
		Enabled:    true,
	}

	created, err := repo.Create(ctx, d)
	require.NoError(t, err)
	assert.Equal(t, d.Name, created.Name)
	assert.Equal(t, d.Host, created.Host)

	found, err := repo.FindByID(ctx, d.ID)
	require.NoError(t, err)
	assert.Equal(t, d.Name, found.Name)
	assert.Equal(t, "border router", found.Extra["note"])

	all, err := repo.FindAll(ctx)
	require.NoError(t, err)
	assert.Len(t, all, 1)

	updated := d
	updated.Name = "Mikrotik Border Updated"
	_, err = repo.Update(ctx, updated)
	require.NoError(t, err)

	found, err = repo.FindByID(ctx, d.ID)
	require.NoError(t, err)
	assert.Equal(t, "Mikrotik Border Updated", found.Name)

	require.NoError(t, repo.Delete(ctx, d.ID))
	_, err = repo.FindByID(ctx, d.ID)
	require.Error(t, err)
	assert.ErrorIs(t, err, device.ErrNotFound)
}

func TestDeviceRepository_FindByID_NotFound(t *testing.T) {
	ctx := context.Background()
	db, err := testutil.NewMemoryDB()
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&deviceModel{}))

	repo := NewDeviceRepository(db)

	_, err = repo.FindByID(ctx, "non-existent")
	require.Error(t, err)
	assert.ErrorIs(t, err, device.ErrNotFound)
}

func TestDeviceRepository_toDomain_NilExtra(t *testing.T) {
	m := deviceModel{
		ID:     "dev-002",
		Name:   "Test",
		Vendor: "mikrotik",
		Host:   "1.2.3.4",
	}
	m.Extra = nil

	d := m.toDomain()
	assert.NotNil(t, d.Extra)
	assert.Equal(t, map[string]string{}, d.Extra)
}
