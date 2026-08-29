package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres/model"
	"github.com/quixiq/polyglot/internal/domain/llm"
	"github.com/quixiq/polyglot/internal/port"
)

// LLMConfigRepository implements port.LLMConfigRepository for GORM/Postgres.
type LLMConfigRepository struct {
	db *gorm.DB
}

var _ port.LLMConfigRepository = (*LLMConfigRepository)(nil)

// NewLLMConfigRepository creates a new LLMConfigRepository.
func NewLLMConfigRepository(db *gorm.DB) *LLMConfigRepository {
	return &LLMConfigRepository{db: db}
}

type tokenStats struct {
	TotalIn  int64
	TotalOut int64
	Count    int64
}

func (r *LLMConfigRepository) Create(ctx context.Context, config *llm.Config) error {
	m := model.LLMConfigModelFromDomain(config)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	config.ID = m.ID
	return nil
}

func (r *LLMConfigRepository) PopulateLLMConfigAnalytics(ctx context.Context, config *llm.Config) {
	if config == nil {
		return
	}

	costInRate := config.CostPer1MInput
	costOutRate := config.CostPer1MOutput
	if costInRate <= 0 && costOutRate <= 0 {
		costInRate, costOutRate = llm.GetDefaultModelPricing(config.Provider, config.Model)
		config.CostPer1MInput = costInRate
		config.CostPer1MOutput = costOutRate
		_ = r.db.WithContext(ctx).Model(&model.LLMConfigModel{}).Where("id = ?", config.ID).
			Updates(map[string]any{
				"cost_per_1m_input":  costInRate,
				"cost_per_1m_output": costOutRate,
			})
	}

	var stats tokenStats
	err := r.db.WithContext(ctx).Model(&model.MessageModel{}).
		Select("COALESCE(SUM(token_in), 0) as total_in, COALESCE(SUM(token_out), 0) as total_out, COUNT(*) as count").
		Where("llm_config_id = ? OR (llm_config_id IS NULL AND sender_type = 'bot')", config.ID).
		Scan(&stats).Error

	if err == nil {
		config.TotalInputTokens = stats.TotalIn
		config.TotalOutputTokens = stats.TotalOut
		config.TotalMessages = stats.Count

		costUSD := (float64(stats.TotalIn)/1_000_000.0)*costInRate +
			(float64(stats.TotalOut)/1_000_000.0)*costOutRate

		config.TotalCostUSD = costUSD
		config.TotalCostIDR = costUSD * 16000.0
	}
}

func (r *LLMConfigRepository) FindByID(ctx context.Context, id uint) (*llm.Config, error) {
	var m model.LLMConfigModel
	if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	cfg := m.ToDomain()
	r.PopulateLLMConfigAnalytics(ctx, cfg)
	return cfg, nil
}

func (r *LLMConfigRepository) FindActive(ctx context.Context) (*llm.Config, error) {
	var m model.LLMConfigModel
	// 1. Coba cari konfigurasi yang ditandai aktif
	if err := r.db.WithContext(ctx).Where("is_active = ?", true).First(&m).Error; err == nil {
		cfg := m.ToDomain()
		r.PopulateLLMConfigAnalytics(ctx, cfg)
		return cfg, nil
	}

	// 2. Fallback: jika belum ada yang aktif, ambil konfigurasi terbaru dan aktifkan otomatis
	if err := r.db.WithContext(ctx).Order("updated_at DESC, id DESC").First(&m).Error; err == nil {
		m.IsActive = true
		_ = r.db.WithContext(ctx).Model(&m).Update("is_active", true)
		cfg := m.ToDomain()
		r.PopulateLLMConfigAnalytics(ctx, cfg)
		return cfg, nil
	}

	return nil, ErrNotFound
}

func (r *LLMConfigRepository) FindAll(ctx context.Context) ([]llm.Config, error) {
	var mList []model.LLMConfigModel
	if err := r.db.WithContext(ctx).Order("created_at DESC").Find(&mList).Error; err != nil {
		return nil, err
	}
	var res []llm.Config
	for _, m := range mList {
		cfg := m.ToDomain()
		if cfg != nil {
			r.PopulateLLMConfigAnalytics(ctx, cfg)
			res = append(res, *cfg)
		}
	}
	return res, nil
}

func (r *LLMConfigRepository) Update(ctx context.Context, config *llm.Config) error {
	m := model.LLMConfigModelFromDomain(config)
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *LLMConfigRepository) SetActive(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.LLMConfigModel{}).Where("is_active = ?", true).Update("is_active", false).Error; err != nil {
			return err
		}
		res := tx.Model(&model.LLMConfigModel{}).Where("id = ?", id).Update("is_active", true)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return ErrNotFound
		}
		return nil
	})
}

func (r *LLMConfigRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.LLMConfigModel{}, id).Error
}

// ─── Backward compatibility delegations on Store ──────────────────────────────

func (s *Store) Create(ctx context.Context, config *llm.Config) error {
	return NewLLMConfigRepository(s.db).Create(ctx, config)
}

func (s *Store) PopulateLLMConfigAnalytics(ctx context.Context, config *llm.Config) {
	NewLLMConfigRepository(s.db).PopulateLLMConfigAnalytics(ctx, config)
}

func (s *Store) FindByID(ctx context.Context, id uint) (*llm.Config, error) {
	return NewLLMConfigRepository(s.db).FindByID(ctx, id)
}

func (s *Store) FindActive(ctx context.Context) (*llm.Config, error) {
	return NewLLMConfigRepository(s.db).FindActive(ctx)
}

func (s *Store) FindAll(ctx context.Context) ([]llm.Config, error) {
	return NewLLMConfigRepository(s.db).FindAll(ctx)
}

func (s *Store) Update(ctx context.Context, config *llm.Config) error {
	return NewLLMConfigRepository(s.db).Update(ctx, config)
}

func (s *Store) SetActive(ctx context.Context, id uint) error {
	return NewLLMConfigRepository(s.db).SetActive(ctx, id)
}

func (s *Store) Delete(ctx context.Context, id uint) error {
	return NewLLMConfigRepository(s.db).Delete(ctx, id)
}
