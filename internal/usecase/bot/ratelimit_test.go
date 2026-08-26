package bot_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quixiq/polyglot/internal/domain/customer"
	settingDomain "github.com/quixiq/polyglot/internal/domain/setting"
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

type mockSettingRepoForRateLimit struct {
	settings *settingDomain.BotSettings
}

func (m *mockSettingRepoForRateLimit) Get(ctx context.Context, key string) (*settingDomain.Setting, error) {
	return nil, nil
}
func (m *mockSettingRepoForRateLimit) GetAll(ctx context.Context) ([]settingDomain.Setting, error) {
	return nil, nil
}
func (m *mockSettingRepoForRateLimit) GetByCategory(ctx context.Context, category string) ([]settingDomain.Setting, error) {
	return nil, nil
}
func (m *mockSettingRepoForRateLimit) GetValue(ctx context.Context, key string, fallback string) string {
	return fallback
}
func (m *mockSettingRepoForRateLimit) Set(ctx context.Context, key, value, category, description string) error {
	return nil
}
func (m *mockSettingRepoForRateLimit) BatchSet(ctx context.Context, settings []settingDomain.Setting) error {
	return nil
}
func (m *mockSettingRepoForRateLimit) GetBotSettings(ctx context.Context) (*settingDomain.BotSettings, error) {
	if m.settings == nil {
		return settingDomain.DefaultBotSettings(), nil
	}
	return m.settings, nil
}
func (m *mockSettingRepoForRateLimit) SaveBotSettings(ctx context.Context, s *settingDomain.BotSettings) error {
	m.settings = s
	return nil
}

type mockUserRepoForRateLimit struct {
	users []*customer.User
}

func (m *mockUserRepoForRateLimit) FindAll(ctx context.Context) ([]*customer.User, error) {
	return m.users, nil
}
func (m *mockUserRepoForRateLimit) FindByID(ctx context.Context, id uint) (*customer.User, error) {
	return nil, nil
}
func (m *mockUserRepoForRateLimit) FindByUsername(ctx context.Context, username string) (*customer.User, error) {
	return nil, nil
}
func (m *mockUserRepoForRateLimit) FindByEmail(ctx context.Context, email string) (*customer.User, error) {
	return nil, nil
}
func (m *mockUserRepoForRateLimit) FindByRoles(ctx context.Context, roles []string, activeOnly bool) ([]*customer.User, error) {
	return nil, nil
}
func (m *mockUserRepoForRateLimit) Create(ctx context.Context, user *customer.User) error {
	return nil
}
func (m *mockUserRepoForRateLimit) Update(ctx context.Context, user *customer.User) error {
	return nil
}
func (m *mockUserRepoForRateLimit) Delete(ctx context.Context, id uint) error {
	return nil
}
func (m *mockUserRepoForRateLimit) UpdatePassword(ctx context.Context, id uint, passwordHash string) error {
	return nil
}
func (m *mockUserRepoForRateLimit) UpdateStatus(ctx context.Context, id uint, isActive bool) error {
	return nil
}
func (m *mockUserRepoForRateLimit) Count(ctx context.Context) (int64, error) {
	return int64(len(m.users)), nil
}
func (m *mockUserRepoForRateLimit) List(ctx context.Context, offset, limit int, search string) ([]*customer.User, int64, error) {
	return m.users, int64(len(m.users)), nil
}
func (m *mockUserRepoForRateLimit) AssignDevices(ctx context.Context, userID uint, deviceIDs []string, assignedBy *uint) error {
	return nil
}
func (m *mockUserRepoForRateLimit) GetAssignedDeviceIDs(ctx context.Context, userID uint) ([]string, error) {
	return nil, nil
}
func (m *mockUserRepoForRateLimit) IsDeviceAccessibleByUser(ctx context.Context, userID uint, deviceID string) (bool, error) {
	return true, nil
}

func TestRateLimiter_NormalChatAllowed(t *testing.T) {
	cache := newMockCacheStore()
	limiter := bot.NewRateLimiter(cache, nil, nil)
	ctx := context.Background()

	// Pesan 1-3 dalam toleransi (default burst limit: 3)
	for i := 0; i < 3; i++ {
		res, err := limiter.Check(ctx, "628123456789", "Halo admin")
		require.NoError(t, err)
		assert.Equal(t, bot.StatusAllowed, res.Status)
	}
}

func TestRateLimiter_BurstSpam_Mutes1Hour(t *testing.T) {
	cache := newMockCacheStore()
	limiter := bot.NewRateLimiter(cache, nil, nil)
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
	limiter := bot.NewRateLimiter(cache, nil, nil)
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
	settingRepo := &mockSettingRepoForRateLimit{
		settings: &settingDomain.BotSettings{
			BurstLimit:     100, // Longgar untuk fokus test daily limit
			DailyChatLimit: 10,
		},
	}
	limiter := bot.NewRateLimiter(cache, settingRepo, nil)
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
	settingRepo := &mockSettingRepoForRateLimit{
		settings: &settingDomain.BotSettings{
			BurstLimit:            1,
			DailyChatLimit:        1,
			WhitelistAllStaff:     true,
			CustomWhitelistPhones: "628999999999, +628111111111",
		},
	}
	userRepo := &mockUserRepoForRateLimit{
		users: []*customer.User{
			{
				ID:          1,
				Username:    "staff_ali",
				PhoneNumber: "628777777777",
				IsActive:    true,
			},
		},
	}
	limiter := bot.NewRateLimiter(cache, settingRepo, userRepo)
	ctx := context.Background()

	// Custom Whitelisted numbers bebas limit
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

	// Staff Whitelist bebas limit
	for i := 0; i < 20; i++ {
		res, err := limiter.Check(ctx, "628777777777", "Spam staf")
		require.NoError(t, err)
		assert.Equal(t, bot.StatusAllowed, res.Status)
	}
}

func TestRateLimiter_ResetRateLimitAndStatus(t *testing.T) {
	cache := newMockCacheStore()
	limiter := bot.NewRateLimiter(cache, nil, nil)
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

func TestRateLimiter_DynamicSettings_CustomMuteAndBanDuration(t *testing.T) {
	cache := newMockCacheStore()
	mockRepo := &mockSettingRepoForRateLimit{
		settings: &settingDomain.BotSettings{
			BurstLimit:            3,
			BurstWindowSecs:       5,
			Mute1HourSecs:         7200,   // 2 jam
			Ban24HourSecs:         172800, // 2 hari
			DailyChatLimit:        20,     // 20 chat
			WhitelistAllStaff:     false,
			CustomWhitelistPhones: "",
		},
	}
	limiter := bot.NewRateLimiter(cache, mockRepo, nil)
	ctx := context.Background()
	phone := "628999999999"

	// 1. Trigger burst spam
	for i := 0; i < 3; i++ {
		_, _ = limiter.Check(ctx, phone, "Normal")
	}
	res, err := limiter.Check(ctx, phone, "Spam 4")
	require.NoError(t, err)
	assert.Equal(t, bot.StatusWarned, res.Status)
	// Harus menyebut 2 jam, bukan 1 jam!
	assert.Contains(t, res.Message, "2 jam")

	// 2. Status saat di-mute
	resMuted, err := limiter.Check(ctx, phone, "Spam while muted")
	require.NoError(t, err)
	assert.Equal(t, bot.StatusMuted, resMuted.Status)
	assert.Contains(t, resMuted.Message, "2 jam")

	// 3. Kuota status harus 20
	status, err := limiter.GetRateLimitStatus(ctx, phone)
	require.NoError(t, err)
	assert.Equal(t, 20, status.DailyQuotaLimit)
}
