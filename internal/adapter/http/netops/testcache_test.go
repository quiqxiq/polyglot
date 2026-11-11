package netops

import (
	"context"
	"sync"
	"time"

	"github.com/quixiq/polyglot/internal/port"
)

type testCacheItem struct {
	value     string
	expiresAt time.Time
}

type testCache struct {
	mu    sync.Mutex
	items map[string]testCacheItem
	now   time.Time
}

func newTestCache() *testCache {
	return &testCache{items: make(map[string]testCacheItem), now: time.Now()}
}

func (c *testCache) Get(_ context.Context, key string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	item, ok := c.items[key]
	if !ok {
		return "", port.ErrCacheMiss
	}
	if !item.expiresAt.IsZero() && !c.now.Before(item.expiresAt) {
		delete(c.items, key)
		return "", port.ErrCacheMiss
	}
	return item.value, nil
}

func (c *testCache) Set(_ context.Context, key, value string, expirationSeconds int) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	item := testCacheItem{value: value}
	if expirationSeconds > 0 {
		item.expiresAt = c.now.Add(time.Duration(expirationSeconds) * time.Second)
	}
	c.items[key] = item
	return nil
}

func (c *testCache) Delete(_ context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
	return nil
}

func (c *testCache) Take(_ context.Context, key string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	item, ok := c.items[key]
	if !ok {
		return "", port.ErrCacheMiss
	}
	delete(c.items, key)
	if !item.expiresAt.IsZero() && !c.now.Before(item.expiresAt) {
		return "", port.ErrCacheMiss
	}
	return item.value, nil
}

func (c *testCache) advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
}
