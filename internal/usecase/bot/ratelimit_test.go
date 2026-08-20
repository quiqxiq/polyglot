package bot_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quixiq/polyglot/internal/config"
	"github.com/quixiq/polyglot/internal/usecase/bot"
)

type mockCacheStore struct {
	data map[string]string
}

func newMockCacheStore() *mockCacheStore {
	return &mockCacheStore{data: make(map[string]string)}
}

func (m *mockCacheStore) Get(_ context.Context, key string) (string, error) {
	v, ok := m.data[key]
	if !ok {
		return "", errors.New("key not found")
	}
	return v, nil
}

func (m *mockCacheStore) Set(_ context.Context, key string, value string, _ int) error {
	m.data[key] = value
	return nil
}

func (m *mockCacheStore) Delete(_ context.Context, key string) error {
	delete(m.data, key)
	return nil
}

func TestRateLimiter_NormalChatAllowed(t *testing.T) {
	cache := newMockCacheStore()
	cfg := config.Config{
		BotBurstLimit:      3,
		BotBurstWindowSecs: 5,
		BotMute1HourSecs:   3600,
		BotBan24HourSecs:   86400,
		BotDailyChatLimit:  10,
	}
	limiter := bot.NewRateLimiter(cache, cfg)
	ctx := context.Background()

	// Pesan 1-3 dalam toleransi
	for i := 0; i < 3; i++ {
		res, err := limiter.Check(ctx, "628123456789", "Halo admin")
		require.NoError(t, err)
		assert.Equal(t, bot.StatusAllowed, res.Status)
	}
}

func TestRateLimiter_BurstSpam_Mutes1Hour(t *testing.T) {
	cache := newMockCacheStore()
	cfg := config.Config{
		BotBurstLimit:      3,
		BotBurstWindowSecs: 5,
		BotMute1HourSecs:   3600,
		BotBan24HourSecs:   86400,
		BotDailyChatLimit:  10,
	}
	limiter := bot.NewRateLimiter(cache, cfg)
	ctx := context.Background()

	// Kirim 3 pesan normal
	for i := 0; i < 3; i++ {
		res, err := limiter.Check(ctx, "628123456789", "Halo")
		require.NoError(t, err)
		assert.Equal(t, bot.StatusAllowed, res.Status)
	}

	// Pesan ke-4 melanggar burst limit (Maks 3 / 5s) -> Warning + Mute 1h
	res, err := limiter.Check(ctx, "628123456789", "Spam 4")
	require.NoError(t, err)
	assert.Equal(t, bot.StatusWarned, res.Status)
	assert.Contains(t, res.Message, "1 jam")

	// Pesan ke-5 saat status mute aktif -> StatusMuted
	resMuted, err := limiter.Check(ctx, "628123456789", "Spam 5")
	require.NoError(t, err)
	assert.Equal(t, bot.StatusMuted, resMuted.Status)
}

func TestRateLimiter_EscalatesTo24HourBan(t *testing.T) {
	cache := newMockCacheStore()
	cfg := config.Config{
		BotBurstLimit:      3,
		BotBurstWindowSecs: 5,
		BotMute1HourSecs:   3600,
		BotBan24HourSecs:   86400,
		BotDailyChatLimit:  10,
	}
	limiter := bot.NewRateLimiter(cache, cfg)
	ctx := context.Background()

	// Trigger burst mute
	for i := 0; i < 4; i++ {
		_, _ = limiter.Check(ctx, "628123456789", "Spam")
	}

	// Cek status mute awal: temp_1h
	muteVal, _ := cache.Get(ctx, "mute:628123456789")
	assert.Equal(t, "temp_1h", muteVal)

	// Lakukan spam berulang saat sedang di-mute (3 strikes)
	for i := 0; i < 3; i++ {
		res, err := limiter.Check(ctx, "628123456789", "Spam while muted")
		require.NoError(t, err)
		assert.Equal(t, bot.StatusMuted, res.Status)
	}

	// Verifikasi penalti dinaikkan menjadi ban_24h
	muteValEscalated, _ := cache.Get(ctx, "mute:628123456789")
	assert.Equal(t, "ban_24h", muteValEscalated)
}

func TestRateLimiter_DailyQuotaExceeded(t *testing.T) {
	cache := newMockCacheStore()
	cfg := config.Config{
		BotBurstLimit:      100, // Longgar untuk fokus test daily limit
		BotDailyChatLimit:  10,
	}
	limiter := bot.NewRateLimiter(cache, cfg)
	ctx := context.Background()

	// Simulasikan 10 chat sukses
	for i := 0; i < 10; i++ {
		res, err := limiter.Check(ctx, "628123456789", "Tanya paket")
		require.NoError(t, err)
		assert.Equal(t, bot.StatusAllowed, res.Status)
		_ = limiter.IncrementDailyQuota(ctx, "628123456789")
	}

	// Chat ke-11 -> StatusDailyQuotaExceeded
	resExceeded, err := limiter.Check(ctx, "628123456789", "Tanya lagi")
	require.NoError(t, err)
	assert.Equal(t, bot.StatusDailyQuotaExceeded, resExceeded.Status)
	assert.Contains(t, resExceeded.Message, "10 percakapan AI")
}

func TestRateLimiter_WhitelistBypass(t *testing.T) {
	cache := newMockCacheStore()
	cfg := config.Config{
		BotBurstLimit:      1,
		BotDailyChatLimit:  1,
		BotWhitelistPhones: []string{"628999999999", "+628111111111"},
	}
	limiter := bot.NewRateLimiter(cache, cfg)
	ctx := context.Background()

	// Whitelisted number bebas spam
	for i := 0; i < 20; i++ {
		res, err := limiter.Check(ctx, "628999999999@s.whatsapp.net", "Spam admin")
		require.NoError(t, err)
		assert.Equal(t, bot.StatusAllowed, res.Status)
	}

	for i := 0; i < 20; i++ {
		res, err := limiter.Check(ctx, "628111111111", "Spam teknisi")
		require.NoError(t, err)
		assert.Equal(t, bot.StatusAllowed, res.Status)
	}
}

func TestRateLimiter_ResetRateLimitAndStatus(t *testing.T) {
	cache := newMockCacheStore()
	cfg := config.Config{
		BotBurstLimit:     3,
		BotDailyChatLimit: 10,
	}
	limiter := bot.NewRateLimiter(cache, cfg)
	ctx := context.Background()

	phone := "628123456789"

	// Trigger burst mute + increment quota
	for i := 0; i < 4; i++ {
		_, _ = limiter.Check(ctx, phone, "Spam")
	}
	_ = limiter.IncrementDailyQuota(ctx, phone)

	// Status sebelum reset
	status, err := limiter.GetRateLimitStatus(ctx, phone)
	require.NoError(t, err)
	assert.True(t, status.IsMuted)
	assert.Equal(t, "temp_1h", status.MuteType)
	assert.Equal(t, 1, status.DailyChatCount)

	// Admin melakukan reset limit
	err = limiter.ResetRateLimit(ctx, phone)
	require.NoError(t, err)

	// Status setelah reset
	statusAfter, err := limiter.GetRateLimitStatus(ctx, phone)
	require.NoError(t, err)
	assert.False(t, statusAfter.IsMuted)
	assert.Equal(t, 0, statusAfter.DailyChatCount)

	// Cek pesan selanjutnya kembali allowed
	res, err := limiter.Check(ctx, phone, "Halo kembali")
	require.NoError(t, err)
	assert.Equal(t, bot.StatusAllowed, res.Status)
}
