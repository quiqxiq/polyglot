package memorycache

import (
	"context"
	"testing"
	"time"

	"github.com/quixiq/polyglot/internal/port"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStoreGetSetDelete(t *testing.T) {
	store := NewStore()
	ctx := context.Background()

	_, err := store.Get(ctx, "missing")
	require.ErrorIs(t, err, port.ErrCacheMiss)

	require.NoError(t, store.Set(ctx, "key", "value", 60))
	got, err := store.Get(ctx, "key")
	require.NoError(t, err)
	assert.Equal(t, "value", got)

	require.NoError(t, store.Delete(ctx, "key"))
	_, err = store.Get(ctx, "key")
	require.ErrorIs(t, err, port.ErrCacheMiss)
}

func TestStoreExpiresValues(t *testing.T) {
	store := NewStore()
	now := time.Now()
	store.now = func() time.Time { return now }
	ctx := context.Background()

	require.NoError(t, store.Set(ctx, "key", "value", 1))
	now = now.Add(2 * time.Second)

	_, err := store.Get(ctx, "key")
	require.ErrorIs(t, err, port.ErrCacheMiss)
}

func TestStoreTakeConsumesValueOnce(t *testing.T) {
	store := NewStore()
	ctx := context.Background()

	require.NoError(t, store.Set(ctx, "key", "value", 60))
	got, err := store.Take(ctx, "key")
	require.NoError(t, err)
	assert.Equal(t, "value", got)

	_, err = store.Take(ctx, "key")
	require.ErrorIs(t, err, port.ErrCacheMiss)
}

func TestStoreTakeRejectsExpiredValue(t *testing.T) {
	store := NewStore()
	now := time.Now()
	store.now = func() time.Time { return now }
	ctx := context.Background()

	require.NoError(t, store.Set(ctx, "key", "value", 1))
	now = now.Add(2 * time.Second)

	_, err := store.Take(ctx, "key")
	require.ErrorIs(t, err, port.ErrCacheMiss)
}
