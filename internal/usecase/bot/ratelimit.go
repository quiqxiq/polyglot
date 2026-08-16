package bot

import (
	"context"

	"github.com/quixiq/polyglot/internal/port"
)

type RateLimitStatus int

const (
	StatusAllowed RateLimitStatus = iota
	StatusWarned
	StatusBlocked
	StatusMuted
)

type RateLimitResult struct {
	Status  RateLimitStatus
	Message string
}

type RateLimiter struct {
	cache port.CacheStore
}

func NewRateLimiter(cache port.CacheStore) *RateLimiter {
	return &RateLimiter{cache: cache}
}

func (r *RateLimiter) Check(ctx context.Context, customerNumber string, messageContent string) (RateLimitResult, error) {
	if r.cache == nil {
		return RateLimitResult{Status: StatusAllowed}, nil
	}
	mutedVal, err := r.cache.Get(ctx, "mute:"+customerNumber)
	if err == nil && mutedVal == "true" {
		return RateLimitResult{
			Status:  StatusMuted,
			Message: "Nomor Anda sedang di-mute sementara karena terlalu banyak pesan.",
		}, nil
	}
	return RateLimitResult{Status: StatusAllowed}, nil
}
