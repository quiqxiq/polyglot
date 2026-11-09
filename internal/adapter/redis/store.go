package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Store provides Redis-backed cache and session operations.
type Store struct {
	client *redis.Client
}

func NewStore(url string) (*Store, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis URL: %w", err)
	}

	client := redis.NewClient(opts)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &Store{client: client}, nil
}

func (s *Store) Get(ctx context.Context, key string) (string, error) {
	return s.client.Get(ctx, key).Result()
}

func (s *Store) Set(ctx context.Context, key string, value string, expirationSeconds int) error {
	return s.client.Set(ctx, key, value, time.Duration(expirationSeconds)*time.Second).Err()
}

func (s *Store) Delete(ctx context.Context, key string) error {
	return s.client.Del(ctx, key).Err()
}

func rateLimitKey(scope string, window string) string {
	return fmt.Sprintf("rl:%s:%s", scope, window)
}

// IncrementRateLimit increments the hit counter for the given scope and window.
func (s *Store) IncrementRateLimit(ctx context.Context, scope string, window string, ttl time.Duration) (int64, error) {
	key := rateLimitKey(scope, window)
	pipe := s.client.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, ttl)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}
	return incr.Val(), nil
}

// GetRateLimitCount returns the current hit counter for the given scope and window.
func (s *Store) GetRateLimitCount(ctx context.Context, scope string, window string) (int64, error) {
	count, err := s.client.Get(ctx, rateLimitKey(scope, window)).Int64()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, nil
		}
		return 0, err
	}
	return count, nil
}

// ResetRateLimit clears a rate-limit counter (e.g. after a successful login).
func (s *Store) ResetRateLimit(ctx context.Context, scope string, window string) error {
	return s.client.Del(ctx, rateLimitKey(scope, window)).Err()
}
