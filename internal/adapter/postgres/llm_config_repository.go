package postgres

import (
	"errors"

	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres/models"
	"github.com/quixiq/polyglot/internal/domain/llm"
)

type tokenStats struct {
	TotalIn  int64
	TotalOut int64
	Count    int64
}

func (s *Store) Create(config *llm.LLMConfig) error {
	return s.CreateLLMConfig(config)
}

func (s *Store) FindByID(id uint) (*llm.LLMConfig, error) {
	return s.FindLLMConfigByID(id)
}

func (s *Store) FindActive() (*llm.LLMConfig, error) {
	return s.FindActiveLLMConfig()
}

func (s *Store) FindAll() ([]llm.LLMConfig, error) {
	return s.FindAllLLMConfigs()
}

func (s *Store) Update(config *llm.LLMConfig) error {
	return s.UpdateLLMConfig(config)
}

func (s *Store) SetActive(id uint) error {
	return s.SetActiveLLMConfig(id)
}

func (s *Store) Delete(id uint) error {
	return s.DeleteLLMConfig(id)
}

func (s *Store) CreateLLMConfig(config *llm.LLMConfig) error {
	m := models.LLMConfigModelFromDomain(config)
	if err := s.db.Create(m).Error; err != nil {
		return err
	}
	config.ID = m.ID
	return nil
}

func (s *Store) PopulateLLMConfigAnalytics(config *llm.LLMConfig) {
	if config == nil {
		return
	}

	costInRate := config.CostPer1MInput
	costOutRate := config.CostPer1MOutput
	if costInRate <= 0 && costOutRate <= 0 {
		costInRate, costOutRate = llm.GetDefaultModelPricing(config.Provider, config.Model)
		config.CostPer1MInput = costInRate
		config.CostPer1MOutput = costOutRate
		_ = s.db.Model(&models.LLMConfigModel{}).Where("id = ?", config.ID).
			Updates(map[string]interface{}{
				"cost_per_1m_input":  costInRate,
				"cost_per_1m_output": costOutRate,
			})
	}

	var stats tokenStats
	err := s.db.Model(&models.MessageModel{}).
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

func (s *Store) FindLLMConfigByID(id uint) (*llm.LLMConfig, error) {
	var m models.LLMConfigModel
	if err := s.db.First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	cfg := m.ToDomain()
	s.PopulateLLMConfigAnalytics(cfg)
	return cfg, nil
}

func (s *Store) FindActiveLLMConfig() (*llm.LLMConfig, error) {
	var m models.LLMConfigModel
	if err := s.db.Where("is_active = ?", true).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	cfg := m.ToDomain()
	s.PopulateLLMConfigAnalytics(cfg)
	return cfg, nil
}

func (s *Store) FindAllLLMConfigs() ([]llm.LLMConfig, error) {
	var mList []models.LLMConfigModel
	if err := s.db.Order("created_at DESC").Find(&mList).Error; err != nil {
		return nil, err
	}
	var res []llm.LLMConfig
	for _, m := range mList {
		cfg := m.ToDomain()
		if cfg != nil {
			s.PopulateLLMConfigAnalytics(cfg)
			res = append(res, *cfg)
		}
	}
	return res, nil
}

func (s *Store) UpdateLLMConfig(config *llm.LLMConfig) error {
	m := models.LLMConfigModelFromDomain(config)
	return s.db.Save(m).Error
}

func (s *Store) SetActiveLLMConfig(id uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.LLMConfigModel{}).Where("is_active = ?", true).Update("is_active", false).Error; err != nil {
			return err
		}
		res := tx.Model(&models.LLMConfigModel{}).Where("id = ?", id).Update("is_active", true)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return ErrNotFound
		}
		return nil
	})
}

func (s *Store) DeleteLLMConfig(id uint) error {
	return s.db.Delete(&models.LLMConfigModel{}, id).Error
}
