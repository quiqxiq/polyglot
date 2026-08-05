package bot

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/quixiq/polyglot/internal/adapter/redis"
	"github.com/quixiq/polyglot/internal/config"
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
	redisStore *redis.Store
	cfg        config.Config
}

func NewRateLimiter(redisStore *redis.Store, cfg config.Config) *RateLimiter {
	return &RateLimiter{
		redisStore: redisStore,
		cfg:        cfg,
	}
}

func (r *RateLimiter) Check(ctx context.Context, customerNumber string, messageContent string) (RateLimitResult, error) {
	muted, err := r.redisStore.IsMuted(ctx, customerNumber)
	if err != nil {
		return RateLimitResult{Status: StatusAllowed}, err
	}
	if muted {
		return RateLimitResult{
			Status:  StatusMuted,
			Message: "Nomor Anda sedang dibatasi sementara karena aktivitas pengiriman pesan berlebih.",
		}, nil
	}

	msgHash := fmt.Sprintf("%x", sha256.Sum256([]byte(messageContent)))
	identicalCount, err := r.redisStore.RecordMessageForSpamDetection(ctx, customerNumber, msgHash)
	if err == nil && identicalCount >= 5 {
		_ = r.redisStore.MuteNumber(ctx, customerNumber, 15*time.Minute)
		return RateLimitResult{
			Status:  StatusMuted,
			Message: "Terdeteksi aktivitas spam. Nomor Anda dibatasi selama 15 menit.",
		}, nil
	}

	minCount, err := r.redisStore.IncrementRateLimit(ctx, customerNumber, "min", 1*time.Minute)
	if err != nil {
		return RateLimitResult{Status: StatusAllowed}, nil
	}

	hourCount, err := r.redisStore.IncrementRateLimit(ctx, customerNumber, "hour", 1*time.Hour)
	if err != nil {
		return RateLimitResult{Status: StatusAllowed}, nil
	}

	maxMin := int64(r.cfg.RateLimitPerMinute)
	maxHour := int64(r.cfg.RateLimitPerHour)

	if minCount > maxMin || hourCount > maxHour {
		if minCount == maxMin+1 || hourCount == maxHour+1 {
			return RateLimitResult{
				Status:  StatusWarned,
				Message: "Mohon tunggu sebentar. Anda telah mencapai batas pengiriman pesan. Silakan coba lagi beberapa saat lagi.",
			}, nil
		}
		return RateLimitResult{
			Status: StatusBlocked,
		}, nil
	}

	return RateLimitResult{Status: StatusAllowed}, nil
}
