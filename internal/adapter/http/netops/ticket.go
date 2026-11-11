package netops

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/quixiq/polyglot/internal/port"
)

const (
	ticketPrefix = "netops:ticket:"
	ticketBytes  = 32
)

var errInvalidTicket = errors.New("invalid netops ticket")

// TicketStore issues and atomically redeems one-time NetOps tickets. Only the
// SHA-256 digest is stored, so a cache dump does not expose usable tickets.
type TicketStore struct {
	cache port.CacheStore
	ttl   time.Duration
}

// NewTicketStore returns a ticket store backed by cache.
func NewTicketStore(cache port.CacheStore, ttl time.Duration) (*TicketStore, error) {
	if cache == nil {
		return nil, errors.New("netops ticket cache is required")
	}
	if ttl <= 0 {
		return nil, errors.New("netops ticket lifetime must be positive")
	}
	return &TicketStore{cache: cache, ttl: ttl}, nil
}

// Issue mints a random ticket bound to userID.
func (s *TicketStore) Issue(ctx context.Context, userID uint) (string, error) {
	if userID == 0 {
		return "", errors.New("netops ticket requires a user")
	}
	raw := make([]byte, ticketBytes)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("generate netops ticket: %w", err)
	}
	ticket := base64.RawURLEncoding.EncodeToString(raw)
	if err := s.cache.Set(ctx, ticketKey(ticket), strconv.FormatUint(uint64(userID), 10), int(s.ttl.Seconds())); err != nil {
		return "", fmt.Errorf("store netops ticket: %w", err)
	}
	return ticket, nil
}

// Redeem consumes ticket and returns its user ID. A missing, expired, or
// already-used ticket returns errInvalidTicket.
func (s *TicketStore) Redeem(ctx context.Context, ticket string) (uint, error) {
	if ticket == "" || len(ticket) > 128 {
		return 0, errInvalidTicket
	}
	raw, err := s.cache.Take(ctx, ticketKey(ticket))
	if err != nil {
		if errors.Is(err, port.ErrCacheMiss) {
			return 0, errInvalidTicket
		}
		return 0, fmt.Errorf("redeem netops ticket: %w", err)
	}
	userID, err := strconv.ParseUint(strings.TrimSpace(raw), 10, 64)
	if err != nil || userID == 0 {
		return 0, errInvalidTicket
	}
	return uint(userID), nil
}

// TTL returns the lifetime of newly issued tickets.
func (s *TicketStore) TTL() time.Duration {
	return s.ttl
}

func ticketKey(ticket string) string {
	sum := sha256.Sum256([]byte(ticket))
	return ticketPrefix + hex.EncodeToString(sum[:])
}
