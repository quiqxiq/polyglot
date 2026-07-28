package vault

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/testutil"
)

func TestAESVault_RoundTrip(t *testing.T) {
	ctx := context.Background()
	db, err := testutil.NewMemoryDB()
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&credentialRecord{}))

	key := make([]byte, 32)
	vault, err := NewAESVault(db, key)
	require.NoError(t, err)

	creds := device.Credentials{
		Username: "admin",
		Password: "secret",
		Extra:    map[string]string{"api_key": "abc123"},
	}

	require.NoError(t, vault.Store(ctx, "dev-001", creds))

	got, err := vault.Get(ctx, "dev-001")
	require.NoError(t, err)
	assert.Equal(t, creds.Username, got.Username)
	assert.Equal(t, creds.Password, got.Password)
	assert.Equal(t, creds.Extra, got.Extra)
}

func TestAESVault_Get_NotFound(t *testing.T) {
	ctx := context.Background()
	db, err := testutil.NewMemoryDB()
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&credentialRecord{}))

	key := make([]byte, 32)
	vault, err := NewAESVault(db, key)
	require.NoError(t, err)

	_, err = vault.Get(ctx, "missing")
	require.Error(t, err)
	assert.ErrorIs(t, err, device.ErrNotFound)
}

func TestAESVault_New_InvalidKey(t *testing.T) {
	key := make([]byte, 16)
	_, err := NewAESVault(nil, key)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "32 bytes")
}

func TestAESVault_WrongKey_FailsDecryption(t *testing.T) {
	ctx := context.Background()
	db, err := testutil.NewMemoryDB()
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&credentialRecord{}))

	key := make([]byte, 32)
	vault, err := NewAESVault(db, key)
	require.NoError(t, err)

	require.NoError(t, vault.Store(ctx, "dev-001", device.Credentials{
		Username: "admin",
		Password: "secret",
	}))

	wrongKey := make([]byte, 32)
	wrongKey[0] = 1
	wrongVault, err := NewAESVault(db, wrongKey)
	require.NoError(t, err)

	_, err = wrongVault.Get(ctx, "dev-001")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "decrypt failed")
}
