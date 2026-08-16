package port

import (
	"context"
)

// CacheStore abstracts key-value caching operations.
type CacheStore interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, expirationSeconds int) error
	Delete(ctx context.Context, key string) error
}
