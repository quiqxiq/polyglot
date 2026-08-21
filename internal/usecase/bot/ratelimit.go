package bot

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/quixiq/polyglot/internal/domain/setting"
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
	settingRepo port.SettingRepository
	userRepo    port.UserRepository
}

func cleanPhoneNumber(phone string) string {
	phone = strings.TrimSpace(phone)
	phone = strings.TrimPrefix(phone, "+")
	phone = strings.Split(phone, "@")[0]
	return phone
}

// NewRateLimiter creates a dynamic multi-tier rate limiter backed by system_settings and users repository.
func NewRateLimiter(cache port.CacheStore, settingRepo port.SettingRepository, userRepo port.UserRepository) *RateLimiter {
	return &RateLimiter{
		cache:       cache,
		settingRepo: settingRepo,
		userRepo:    userRepo,
	}
}

// WithSettings attaches dynamic setting and user repositories.
func (r *RateLimiter) WithSettings(settingRepo port.SettingRepository, userRepo port.UserRepository) *RateLimiter {
	r.settingRepo = settingRepo
	r.userRepo = userRepo
	return r
}

// HasCache reports whether the rate limiter has an active cache backend.
func (r *RateLimiter) HasCache() bool {
	return r != nil && r.cache != nil
}

func (r *RateLimiter) isWhitelisted(ctx context.Context, customerNumber string) bool {
	if customerNumber == "" {
		return false
	}
	if r.settingRepo != nil {
		s, err := r.settingRepo.GetBotSettings(ctx)
		if err == nil && s != nil {
			if s.WhitelistAllStaff && r.userRepo != nil {
				allUsers, err := r.userRepo.FindAll(ctx)
				if err == nil {
					for _, u := range allUsers {
						if cleanPhoneNumber(u.PhoneNumber) == customerNumber && u.PhoneNumber != "" {
							return true
						}
					}
				}
			}
			if s.CustomWhitelistPhones != "" {
				for _, p := range strings.Split(s.CustomWhitelistPhones, ",") {
					if cleanPhoneNumber(p) == customerNumber && strings.TrimSpace(p) != "" {
						return true
					}
				}
			}
		}
	}
	return false
}

func formatDurationID(secs int) string {
	if secs <= 0 {
		return "sementara"
	}
	if secs%86400 == 0 {
		days := secs / 86400
		if days == 1 {
			return "24 jam"
		}
		return fmt.Sprintf("%d hari", days)
	}
	if secs%3600 == 0 {
		hours := secs / 3600
		return fmt.Sprintf("%d jam", hours)
	}
	if secs%60 == 0 {
		minutes := secs / 60
		return fmt.Sprintf("%d menit", minutes)
	}
	if secs > 3600 {
		hours := secs / 3600
		rem := (secs % 3600) / 60
		if rem > 0 {
			return fmt.Sprintf("%d jam %d menit", hours, rem)
		}
		return fmt.Sprintf("%d jam", hours)
	}
	if secs > 60 {
		mins := secs / 60
		return fmt.Sprintf("%d menit", mins)
	}
	return fmt.Sprintf("%d detik", secs)
}

func (r *RateLimiter) getEffectiveLimits(ctx context.Context) (burstLimit, burstWindow, mute1h, ban24h, dailyLimit int) {
	defaults := setting.DefaultBotSettings()
	burstLimit = defaults.BurstLimit
	burstWindow = defaults.BurstWindowSecs
	mute1h = defaults.Mute1HourSecs
	ban24h = defaults.Ban24HourSecs
	dailyLimit = defaults.DailyChatLimit

	if r.settingRepo != nil {
		if s, err := r.settingRepo.GetBotSettings(ctx); err == nil && s != nil {
			if s.BurstLimit > 0 {
				burstLimit = s.BurstLimit
			}
			if s.BurstWindowSecs > 0 {
				burstWindow = s.BurstWindowSecs
			}
			if s.Mute1HourSecs > 0 {
				mute1h = s.Mute1HourSecs
			}
			if s.Ban24HourSecs > 0 {
				ban24h = s.Ban24HourSecs
			}
			if s.DailyChatLimit > 0 {
				dailyLimit = s.DailyChatLimit
			}
		}
	}
	return
}

// Check evaluates incoming message for spam burst, active mute penalties, and daily quota.
func (r *RateLimiter) Check(ctx context.Context, customerNumber string, messageContent string) (RateLimitResult, error) {
	customerNumber = cleanPhoneNumber(customerNumber)
	if customerNumber == "" {
		return RateLimitResult{Status: StatusAllowed}, nil
	}

	// 1. Whitelist Check
	if r.isWhitelisted(ctx, customerNumber) {
		return RateLimitResult{Status: StatusAllowed}, nil
	}

	if r.cache == nil {
		return RateLimitResult{Status: StatusAllowed}, nil
	}

	burstLimit, burstWindow, mute1hSecs, ban24hSecs, dailyLimit := r.getEffectiveLimits(ctx)

	// 2. Active Mute / Ban Check (Tier 1 & 2)
	muteVal, err := r.cache.Get(ctx, "mute:"+customerNumber)
	if err == nil && muteVal != "" {
		// Catat spam strike selama masa mute aktif
		strikesKey := "spam_strikes:" + customerNumber
		strikesStr, _ := r.cache.Get(ctx, strikesKey)
		strikes, _ := strconv.Atoi(strikesStr)
		strikes++
		_ = r.cache.Set(ctx, strikesKey, strconv.Itoa(strikes), ban24hSecs)

		// Jika terus melakukan spam (>= 3 strike) selama masa mute, eskalasi ke ban level 2
		if strikes >= 3 && muteVal != "ban_24h" {
			_ = r.cache.Set(ctx, "mute:"+customerNumber, "ban_24h", ban24hSecs)
			return RateLimitResult{
				Status:  StatusMuted,
				Message: fmt.Sprintf("⚠️ Nomor Anda telah dinonaktifkan selama %s karena aktivitas spam berulang. Silakan hubungi admin Customer Service kami.", formatDurationID(ban24hSecs)),
			}, nil
		}

		return RateLimitResult{
			Status:  StatusMuted,
			Message: fmt.Sprintf("⚠️ Nomor Anda sedang dalam masa pembatasan sementara (%s). Mohon menunggu atau hubungi admin jika butuh bantuan segera.", formatDurationID(mute1hSecs)),
		}, nil
	}

	// 3. Burst Anti-Spam Check (Tier 1)
	burstKey := "burst:" + customerNumber
	burstStr, _ := r.cache.Get(ctx, burstKey)
	burstCount, _ := strconv.Atoi(burstStr)
	burstCount++
	_ = r.cache.Set(ctx, burstKey, strconv.Itoa(burstCount), burstWindow)

	if burstCount > burstLimit {
		// Terapkan penalti mute
		_ = r.cache.Set(ctx, "mute:"+customerNumber, "temp_1h", mute1hSecs)
		return RateLimitResult{
			Status:  StatusWarned,
			Message: fmt.Sprintf("⚠️ Anda mengirim pesan terlalu cepat. Layanan asisten otomatis dinonaktifkan sementara untuk nomor Anda selama %s.", formatDurationID(mute1hSecs)),
		}, nil
	}

	// 4. Daily Quota Check (Tier 3)
	today := time.Now().Format("2006-01-02")
	dailyKey := fmt.Sprintf("daily_chat:%s:%s", customerNumber, today)
	dailyStr, _ := r.cache.Get(ctx, dailyKey)
	dailyCount, _ := strconv.Atoi(dailyStr)

	if dailyCount >= dailyLimit {
		return RateLimitResult{
			Status:  StatusDailyQuotaExceeded,
			Message: fmt.Sprintf("ℹ️ Anda telah mencapai batas harian %d percakapan AI untuk hari ini. Untuk bantuan lebih lanjut, silakan hubungi Customer Service kami.", dailyLimit),
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
	isWhitelisted := r.isWhitelisted(ctx, customerNumber)
	_, _, _, _, dailyLimit := r.getEffectiveLimits(ctx)

	if isWhitelisted || r.cache == nil {
		return &RateLimitStatusInfo{
			PhoneNumber:     customerNumber,
			IsMuted:         false,
			DailyChatCount:  0,
			DailyQuotaLimit: dailyLimit,
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
		DailyQuotaLimit: dailyLimit,
		IsWhitelisted:   false,
	}, nil
}

