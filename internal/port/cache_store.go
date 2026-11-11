package port

import (
	"context"
	"errors"
)

// ErrCacheMiss reports that a key is absent or expired.
var ErrCacheMiss = errors.New("cache key not found")

// CacheStore abstracts key-value caching operations.
type CacheStore interface {
	// Get returns the value stored at key. A missing or expired key returns
	// ErrCacheMiss.
	Get(ctx context.Context, key string) (string, error)
	// Set stores value at key for expirationSeconds. A non-positive value
	// stores the value without an expiry.
	Set(ctx context.Context, key string, value string, expirationSeconds int) error
	// Delete removes key. Deleting a missing key is not an error.
	Delete(ctx context.Context, key string) error
	// Take atomically returns the value stored at key and deletes it, so
	// concurrent callers cannot both consume the same value. A missing or
	// expired key returns ErrCacheMiss.
	Take(ctx context.Context, key string) (string, error)
}
