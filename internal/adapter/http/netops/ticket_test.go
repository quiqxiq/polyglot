package netops

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTicketStoreRedeemsExactlyOnce(t *testing.T) {
	cache := newTestCache()
	store, err := NewTicketStore(cache, time.Minute)
	require.NoError(t, err)

	ticket, err := store.Issue(context.Background(), 42)
	require.NoError(t, err)
	assert.NotEmpty(t, ticket)

	userID, err := store.Redeem(context.Background(), ticket)
	require.NoError(t, err)
	assert.Equal(t, uint(42), userID)

	_, err = store.Redeem(context.Background(), ticket)
	require.ErrorIs(t, err, errInvalidTicket)
}

func TestTicketStoreRejectsExpiredAndUnknownTickets(t *testing.T) {
	cache := newTestCache()
	store, err := NewTicketStore(cache, time.Minute)
	require.NoError(t, err)

	ticket, err := store.Issue(context.Background(), 7)
	require.NoError(t, err)
	cache.advance(2 * time.Minute)

	_, err = store.Redeem(context.Background(), ticket)
	require.ErrorIs(t, err, errInvalidTicket)

	_, err = store.Redeem(context.Background(), "unknown")
	require.ErrorIs(t, err, errInvalidTicket)
}

func TestTicketStoreRejectsInvalidConfiguration(t *testing.T) {
	_, err := NewTicketStore(nil, time.Minute)
	require.Error(t, err)
	_, err = NewTicketStore(newTestCache(), 0)
	require.Error(t, err)

	store, err := NewTicketStore(newTestCache(), time.Minute)
	require.NoError(t, err)
	_, err = store.Issue(context.Background(), 0)
	require.Error(t, err)
}
