package bot

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/quixiq/polyglot/internal/config"
	"github.com/quixiq/polyglot/internal/port"
)

type RateLimitStatus int

const (
	StatusAllowed RateLimitStatus = iota
	StatusWarned
	StatusBlocked
	StatusMuted
	StatusDailyQuotaExceeded
)

type RateLimitResult struct {
	Status  RateLimitStatus
	Message string
}

type RateLimitStatusInfo struct {
	PhoneNumber     string
	IsMuted         bool
	MuteType        string // "temp_1h", "ban_24h"
	DailyChatCount  int
	DailyQuotaLimit int
	IsWhitelisted   bool
}

type RateLimiter struct {
	cache       port.CacheStore
	burstLimit  int
	burstWindow int
	mute1hSecs  int
	ban24hSecs  int
	dailyLimit  int
	whitelist   map[string]bool
}

func cleanPhoneNumber(phone string) string {
	phone = strings.TrimSpace(phone)
	phone = strings.TrimPrefix(phone, "+")
	phone = strings.Split(phone, "@")[0]
	return phone
}

func NewRateLimiter(cache port.CacheStore, cfg config.Config) *RateLimiter {
	burstLimit := cfg.BotBurstLimit
	if burstLimit <= 0 {
		burstLimit = 3
	}
	burstWindow := cfg.BotBurstWindowSecs
	if burstWindow <= 0 {
		burstWindow = 5
	}
	mute1h := cfg.BotMute1HourSecs
	if mute1h <= 0 {
		mute1h = 3600
	}
	ban24h := cfg.BotBan24HourSecs
	if ban24h <= 0 {
		ban24h = 86400
	}
	dailyLimit := cfg.BotDailyChatLimit
	if dailyLimit <= 0 {
		dailyLimit = 10
	}

	wl := make(map[string]bool)
	for _, p := range cfg.BotWhitelistPhones {
		cleaned := cleanPhoneNumber(p)
		if cleaned != "" {
			wl[cleaned] = true
		}
	}

	return &RateLimiter{
		cache:       cache,
		burstLimit:  burstLimit,
		burstWindow: burstWindow,
		mute1hSecs:  mute1h,
		ban24hSecs:  ban24h,
		dailyLimit:  dailyLimit,
		whitelist:   wl,
	}
}

// Check evaluates incoming message for spam burst, active mute penalties, and daily quota.
func (r *RateLimiter) Check(ctx context.Context, customerNumber string, messageContent string) (RateLimitResult, error) {
	customerNumber = cleanPhoneNumber(customerNumber)
	if customerNumber == "" {
		return RateLimitResult{Status: StatusAllowed}, nil
	}

	// 1. Whitelist Check
	if r.whitelist[customerNumber] {
		return RateLimitResult{Status: StatusAllowed}, nil
	}

	if r.cache == nil {
		return RateLimitResult{Status: StatusAllowed}, nil
	}

	// 2. Active Mute / Ban Check (Tier 1 & 2)
	muteVal, err := r.cache.Get(ctx, "mute:"+customerNumber)
	if err == nil && muteVal != "" {
		// Increment spam strike counter during active mute
		strikesKey := "spam_strikes:" + customerNumber
		strikesStr, _ := r.cache.Get(ctx, strikesKey)
		strikes, _ := strconv.Atoi(strikesStr)
		strikes++
		_ = r.cache.Set(ctx, strikesKey, strconv.Itoa(strikes), r.ban24hSecs)

		// If user continues to spam during 1h mute (>= 3 strikes), escalate to 24h ban
		if strikes >= 3 && muteVal != "ban_24h" {
			_ = r.cache.Set(ctx, "mute:"+customerNumber, "ban_24h", r.ban24hSecs)
		}

		return RateLimitResult{
			Status:  StatusMuted,
			Message: "Nomor Anda sedang dinonaktifkan sementara karena spam.",
		}, nil
	}

	// 3. Burst Anti-Spam Check (Tier 1)
	burstKey := "burst:" + customerNumber
	burstStr, _ := r.cache.Get(ctx, burstKey)
	burstCount, _ := strconv.Atoi(burstStr)
	burstCount++
	_ = r.cache.Set(ctx, burstKey, strconv.Itoa(burstCount), r.burstWindow)

	if burstCount > r.burstLimit {
		// Apply 1 Hour Mute Penalty
		_ = r.cache.Set(ctx, "mute:"+customerNumber, "temp_1h", r.mute1hSecs)
		return RateLimitResult{
			Status:  StatusWarned,
			Message: "⚠️ Anda mengirim pesan terlalu cepat. Layanan asisten otomatis dinonaktifkan sementara untuk nomor Anda selama 1 jam.",
		}, nil
	}

	// 4. Daily Quota Check (Tier 3)
	today := time.Now().Format("2006-01-02")
	dailyKey := fmt.Sprintf("daily_chat:%s:%s", customerNumber, today)
	dailyStr, _ := r.cache.Get(ctx, dailyKey)
	dailyCount, _ := strconv.Atoi(dailyStr)

	if dailyCount >= r.dailyLimit {
		return RateLimitResult{
			Status:  StatusDailyQuotaExceeded,
			Message: fmt.Sprintf("ℹ️ Anda telah mencapai batas harian %d percakapan AI untuk hari ini. Untuk bantuan lebih lanjut, silakan hubungi Customer Service kami.", r.dailyLimit),
		}, nil
	}

	return RateLimitResult{Status: StatusAllowed}, nil
}

// IncrementDailyQuota increments the successful AI chat count for the customer for today.
func (r *RateLimiter) IncrementDailyQuota(ctx context.Context, customerNumber string) error {
	if r.cache == nil {
		return nil
	}
	customerNumber = cleanPhoneNumber(customerNumber)
	today := time.Now().Format("2006-01-02")
	key := fmt.Sprintf("daily_chat:%s:%s", customerNumber, today)
	dailyStr, _ := r.cache.Get(ctx, key)
	dailyCount, _ := strconv.Atoi(dailyStr)
	dailyCount++
	return r.cache.Set(ctx, key, strconv.Itoa(dailyCount), 86400)
}

// ResetRateLimit removes all active mute, burst, strikes, and daily quota counters for a number.
func (r *RateLimiter) ResetRateLimit(ctx context.Context, customerNumber string) error {
	if r.cache == nil {
		return nil
	}
	customerNumber = cleanPhoneNumber(customerNumber)
	today := time.Now().Format("2006-01-02")
	_ = r.cache.Delete(ctx, "mute:"+customerNumber)
	_ = r.cache.Delete(ctx, "burst:"+customerNumber)
	_ = r.cache.Delete(ctx, "spam_strikes:"+customerNumber)
	_ = r.cache.Delete(ctx, fmt.Sprintf("daily_chat:%s:%s", customerNumber, today))
	return nil
}

// GetRateLimitStatus queries the current rate limit and quota state of a phone number.
func (r *RateLimiter) GetRateLimitStatus(ctx context.Context, customerNumber string) (*RateLimitStatusInfo, error) {
	customerNumber = cleanPhoneNumber(customerNumber)
	isWhitelisted := r.whitelist[customerNumber]
	if isWhitelisted || r.cache == nil {
		return &RateLimitStatusInfo{
			PhoneNumber:     customerNumber,
			IsMuted:         false,
			DailyChatCount:  0,
			DailyQuotaLimit: r.dailyLimit,
			IsWhitelisted:   isWhitelisted,
		}, nil
	}

	muteVal, _ := r.cache.Get(ctx, "mute:"+customerNumber)
	isMuted := muteVal != ""

	today := time.Now().Format("2006-01-02")
	dailyStr, _ := r.cache.Get(ctx, fmt.Sprintf("daily_chat:%s:%s", customerNumber, today))
	dailyCount, _ := strconv.Atoi(dailyStr)

	return &RateLimitStatusInfo{
		PhoneNumber:     customerNumber,
		IsMuted:         isMuted,
		MuteType:        muteVal,
		DailyChatCount:  dailyCount,
		DailyQuotaLimit: r.dailyLimit,
		IsWhitelisted:   false,
	}, nil
}

