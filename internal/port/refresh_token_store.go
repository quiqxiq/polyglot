package port

import "context"

// RefreshTokenStore persists opaque refresh tokens (rotated on every refresh,
// revoked on logout). Implemented by the Redis adapter; defined here so the
// auth service depends on a contract, not a concrete store.
type RefreshTokenStore interface {
	// Set stores value at key with the given TTL in seconds.
	Set(ctx context.Context, key, value string, ttlSeconds int) error
	// Get returns the value at key, or an empty string when missing.
	Get(ctx context.Context, key string) (string, error)
	// Delete removes key.
	Delete(ctx context.Context, key string) error
}
