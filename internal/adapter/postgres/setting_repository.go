package postgres

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/quixiq/polyglot/internal/adapter/postgres/model"
	"github.com/quixiq/polyglot/internal/domain/setting"
	"github.com/quixiq/polyglot/internal/port"
)

type SettingRepository struct {
	db        *gorm.DB
	cacheLock sync.RWMutex
	cache     map[string]string
	cacheTime time.Time
}

var _ port.SettingRepository = (*SettingRepository)(nil)

// NewSettingRepository creates a new instance of SettingRepository with memory read-through cache.
func NewSettingRepository(db *gorm.DB) *SettingRepository {
	return &SettingRepository{
		db:    db,
		cache: make(map[string]string),
	}
}

func (r *SettingRepository) Get(ctx context.Context, key string) (*setting.Setting, error) {
	var m model.SystemSettingModel
	if err := r.db.WithContext(ctx).Where("key = ?", key).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return m.ToDomain(), nil
}

func (r *SettingRepository) GetValue(ctx context.Context, key string, fallback string) string {
	r.cacheLock.RLock()
	if time.Since(r.cacheTime) < 30*time.Second {
		if val, ok := r.cache[key]; ok {
			r.cacheLock.RUnlock()
			return val
		}
	}
	r.cacheLock.RUnlock()

	// Cache miss or expired -> refresh cache
	var models []model.SystemSettingModel
	if err := r.db.WithContext(ctx).Find(&models).Error; err == nil {
		r.cacheLock.Lock()
		r.cache = make(map[string]string, len(models))
		for _, m := range models {
			r.cache[m.Key] = m.Value
		}
		r.cacheTime = time.Now()
		val, ok := r.cache[key]
		r.cacheLock.Unlock()
		if ok {
			return val
		}
	}

	return fallback
}

func (r *SettingRepository) Set(ctx context.Context, key, value, category, description string) error {
	m := model.SystemSettingModel{
		Key:         key,
		Value:       value,
		Category:    category,
		Description: description,
		UpdatedAt:   time.Now(),
	}

	err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value", "category", "description", "updated_at"}),
	}).Create(&m).Error

	if err == nil {
		r.invalidateCache()
	}
	return err
}

func (r *SettingRepository) GetByCategory(ctx context.Context, category string) ([]setting.Setting, error) {
	var models []model.SystemSettingModel
	if err := r.db.WithContext(ctx).Where("category = ?", category).Order("key ASC").Find(&models).Error; err != nil {
		return nil, err
	}

	result := make([]setting.Setting, 0, len(models))
	for _, m := range models {
		if d := m.ToDomain(); d != nil {
			result = append(result, *d)
		}
	}
	return result, nil
}

func (r *SettingRepository) GetAll(ctx context.Context) ([]setting.Setting, error) {
	var models []model.SystemSettingModel
	if err := r.db.WithContext(ctx).Order("category ASC, key ASC").Find(&models).Error; err != nil {
		return nil, err
	}

	result := make([]setting.Setting, 0, len(models))
	for _, m := range models {
		if d := m.ToDomain(); d != nil {
			result = append(result, *d)
		}
	}
	return result, nil
}

func (r *SettingRepository) BatchSet(ctx context.Context, settings []setting.Setting) error {
	if len(settings) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, s := range settings {
			m := model.SystemSettingModel{
				Key:         s.Key,
				Value:       s.Value,
				Category:    s.Category,
				Description: s.Description,
				UpdatedAt:   time.Now(),
			}
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "key"}},
				DoUpdates: clause.AssignmentColumns([]string{"value", "category", "description", "updated_at"}),
			}).Create(&m).Error; err != nil {
				return err
			}
		}
		r.invalidateCache()
		return nil
	})
}

func (r *SettingRepository) GetBotSettings(ctx context.Context) (*setting.BotSettings, error) {
	s := setting.DefaultBotSettings()

	s.BurstLimit = r.getIntValue(ctx, "bot.burst_limit", s.BurstLimit)
	s.BurstWindowSecs = r.getIntValue(ctx, "bot.burst_window_secs", s.BurstWindowSecs)
	s.Mute1HourSecs = r.getIntValue(ctx, "bot.mute_1h_secs", s.Mute1HourSecs)
	s.Ban24HourSecs = r.getIntValue(ctx, "bot.ban_24h_secs", s.Ban24HourSecs)
	s.DailyChatLimit = r.getIntValue(ctx, "bot.daily_chat_limit", s.DailyChatLimit)
	s.SessionTimeoutMinutes = r.getIntValue(ctx, "bot.session_timeout_minutes", s.SessionTimeoutMinutes)
	s.SlidingWindowSize = r.getIntValue(ctx, "bot.sliding_window_size", s.SlidingWindowSize)
	s.LLMMaxOutputTokens = r.getIntValue(ctx, "bot.llm_max_output_tokens", s.LLMMaxOutputTokens)
	s.WhitelistAllStaff = r.getBoolValue(ctx, "bot.whitelist_all_staff", s.WhitelistAllStaff)
	s.CustomWhitelistPhones = r.GetValue(ctx, "bot.custom_whitelist_phones", s.CustomWhitelistPhones)

	return s, nil
}

func (r *SettingRepository) SaveBotSettings(ctx context.Context, s *setting.BotSettings) error {
	if s == nil {
		return errors.New("bot settings cannot be nil")
	}

	items := []setting.Setting{
		{Key: "bot.burst_limit", Value: strconv.Itoa(s.BurstLimit), Category: "bot_rate_limit", Description: "Jumlah pesan burst spam"},
		{Key: "bot.burst_window_secs", Value: strconv.Itoa(s.BurstWindowSecs), Category: "bot_rate_limit", Description: "Rentang waktu deteksi burst spam (detik)"},
		{Key: "bot.mute_1h_secs", Value: strconv.Itoa(s.Mute1HourSecs), Category: "bot_rate_limit", Description: "Durasi mute spam level 1"},
		{Key: "bot.ban_24h_secs", Value: strconv.Itoa(s.Ban24HourSecs), Category: "bot_rate_limit", Description: "Durasi ban spam level 2"},
		{Key: "bot.daily_chat_limit", Value: strconv.Itoa(s.DailyChatLimit), Category: "bot_rate_limit", Description: "Batas chat harian per nomor"},
		{Key: "bot.session_timeout_minutes", Value: strconv.Itoa(s.SessionTimeoutMinutes), Category: "bot_session", Description: "Timeout sesi obrolan (menit)"},
		{Key: "bot.sliding_window_size", Value: strconv.Itoa(s.SlidingWindowSize), Category: "bot_session", Description: "Jumlah history pesan ke LLM"},
		{Key: "bot.llm_max_output_tokens", Value: strconv.Itoa(s.LLMMaxOutputTokens), Category: "bot_session", Description: "Maksimal token output LLM"},
		{Key: "bot.whitelist_all_staff", Value: strconv.FormatBool(s.WhitelistAllStaff), Category: "bot_whitelist", Description: "Bebaskan limit untuk semua staf tabel users"},
		{Key: "bot.custom_whitelist_phones", Value: strings.TrimSpace(s.CustomWhitelistPhones), Category: "bot_whitelist", Description: "Nomor WhatsApp bebas limit"},
	}

	return r.BatchSet(ctx, items)
}

func (r *SettingRepository) getIntValue(ctx context.Context, key string, fallback int) int {
	valStr := r.GetValue(ctx, key, "")
	if valStr == "" {
		return fallback
	}
	if n, err := strconv.Atoi(valStr); err == nil {
		return n
	}
	return fallback
}

func (r *SettingRepository) getBoolValue(ctx context.Context, key string, fallback bool) bool {
	valStr := r.GetValue(ctx, key, "")
	if valStr == "" {
		return fallback
	}
	if b, err := strconv.ParseBool(valStr); err == nil {
		return b
	}
	return fallback
}

func (r *SettingRepository) invalidateCache() {
	r.cacheLock.Lock()
	r.cacheTime = time.Time{}
	r.cacheLock.Unlock()
}
