// Package memorycache provides a process-local CacheStore used when Redis is
// unavailable. Values are lost on restart, so it is a fallback rather than a
// replacement for the shared cache.
package memorycache

import (
	"context"
	"sync"
	"time"

	"github.com/quixiq/polyglot/internal/port"
)

type entry struct {
	value     string
	expiresAt time.Time
}

// Store is a concurrency-safe in-memory cache.
type Store struct {
	mu    sync.Mutex
	items map[string]entry
	now   func() time.Time
}

var _ port.CacheStore = (*Store)(nil)

// NewStore returns an empty in-memory cache.
func NewStore() *Store {
	return &Store{
		items: make(map[string]entry),
		now:   time.Now,
	}
}

// Get returns the value stored at key or port.ErrCacheMiss.
func (s *Store) Get(_ context.Context, key string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, ok := s.items[key]
	if !ok {
		return "", port.ErrCacheMiss
	}
	if s.expired(item) {
		delete(s.items, key)
		return "", port.ErrCacheMiss
	}
	return item.value, nil
}

// Set stores value at key for expirationSeconds; a non-positive value never expires.
func (s *Store) Set(_ context.Context, key string, value string, expirationSeconds int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	item := entry{value: value}
	if expirationSeconds > 0 {
		item.expiresAt = s.now().Add(time.Duration(expirationSeconds) * time.Second)
	}
	s.items[key] = item
	return nil
}

// Delete removes key; deleting a missing key is not an error.
func (s *Store) Delete(_ context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.items, key)
	return nil
}

// Take atomically returns and removes key, or port.ErrCacheMiss.
func (s *Store) Take(_ context.Context, key string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, ok := s.items[key]
	if !ok {
		return "", port.ErrCacheMiss
	}
	delete(s.items, key)
	if s.expired(item) {
		return "", port.ErrCacheMiss
	}
	return item.value, nil
}

func (s *Store) expired(item entry) bool {
	return !item.expiresAt.IsZero() && !s.now().Before(item.expiresAt)
}
