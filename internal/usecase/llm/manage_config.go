package llm

import (
	"context"
	"fmt"

	domainllm "github.com/quixiq/polyglot/internal/domain/llm"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/crypto"
	"github.com/quixiq/polyglot/pkg/fault"
	"github.com/quixiq/polyglot/pkg/llmcost"
)

// ProviderTester tests connectivity to an LLM provider.
type ProviderTester func(ctx context.Context, cfg *domainllm.Config, apiKey string) error

// CreateConfigInput represents parameters for creating an LLM configuration.
type CreateConfigInput struct {
	Provider       string
	ModelName      string
	APIKey         string
	BaseURL        string
	Temperature    float64
	MaxTokens      int
	SystemPrompt   string
	SkillsMode     string
	EnableSkills   bool
	SkillsPrompt   string
	SelectedSkills []string
}

// UpdateConfigInput represents parameters for updating an LLM configuration.
type UpdateConfigInput struct {
	ID             uint
	Provider       string
	ModelName      string
	APIKey         string
	BaseURL        string
	Temperature    float64
	MaxTokens      int
	SystemPrompt   string
	SkillsMode     string
	EnableSkills   bool
	SkillsPrompt   string
	SelectedSkills []string
}

// ManageConfigUseCase coordinates LLM model configurations and credential management.
type ManageConfigUseCase struct {
	repo          port.LLMConfigRepository
	encryptionKey string
	tester        ProviderTester
}

// NewManageConfigUseCase constructs a ManageConfigUseCase.
func NewManageConfigUseCase(repo port.LLMConfigRepository, encryptionKey string, tester ProviderTester) *ManageConfigUseCase {
	return &ManageConfigUseCase{
		repo:          repo,
		encryptionKey: encryptionKey,
		tester:        tester,
	}
}

// List returns all saved LLM configurations.
func (uc *ManageConfigUseCase) List(ctx context.Context) ([]domainllm.Config, error) {
	if uc.repo == nil {
		return nil, domainllm.ErrNotFound
	}
	configs, err := uc.repo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("find all configs: %w", err)
	}
	return configs, nil
}

// GetByID returns an LLM configuration by its primary key ID.
func (uc *ManageConfigUseCase) GetByID(ctx context.Context, id uint) (*domainllm.Config, error) {
	if uc.repo == nil {
		return nil, domainllm.ErrNotFound
	}
	cfg, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("find config by id: %w", err)
	}
	return cfg, nil
}

// Create stores a new LLM configuration, auto-calculating cost rates and encrypting credentials.
func (uc *ManageConfigUseCase) Create(ctx context.Context, input CreateConfigInput) (*domainllm.Config, error) {
	if uc.repo == nil {
		return nil, domainllm.ErrInvalidInput
	}
	if input.Provider == "" || input.ModelName == "" {
		return nil, fmt.Errorf("%w: provider and model name are required", domainllm.ErrInvalidInput)
	}

	var encryptedKey string
	if input.APIKey != "" {
		enc, err := crypto.Encrypt(input.APIKey, uc.encryptionKey)
		if err != nil {
			return nil, fault.Wrap(fault.KindUnknown, fmt.Errorf("failed to encrypt api key: %w", err))
		}
		encryptedKey = enc
	}

	inRate, outRate := llmcost.GetDefaultPricing(input.Provider, input.ModelName)

	isActive := false
	if activeCfg, _ := uc.repo.FindActive(ctx); activeCfg == nil {
		isActive = true
	}

	cfg := &domainllm.Config{
		Provider:        input.Provider,
		Model:           input.ModelName,
		APIKeyEncrypted: encryptedKey,
		BaseURL:         input.BaseURL,
		Temperature:     input.Temperature,
		MaxOutputTokens: input.MaxTokens,
		SystemPrompt:    input.SystemPrompt,
		SkillsMode:      input.SkillsMode,
		EnableSkills:    input.EnableSkills,
		SkillsPrompt:    input.SkillsPrompt,
		SelectedSkills:  input.SelectedSkills,
		CostPer1MInput:  inRate,
		CostPer1MOutput: outRate,
		IsActive:        isActive,
	}

	if err := uc.repo.Create(ctx, cfg); err != nil {
		return nil, fmt.Errorf("create config: %w", err)
	}
	return cfg, nil
}

// Update modifies an existing LLM configuration, re-encrypting credentials if updated.
func (uc *ManageConfigUseCase) Update(ctx context.Context, input UpdateConfigInput) (*domainllm.Config, error) {
	if uc.repo == nil {
		return nil, domainllm.ErrNotFound
	}
	cfg, err := uc.repo.FindByID(ctx, input.ID)
	if err != nil {
		return nil, fmt.Errorf("find config by id: %w", err)
	}

	cfg.Provider = input.Provider
	cfg.Model = input.ModelName
	cfg.BaseURL = input.BaseURL
	cfg.Temperature = input.Temperature
	cfg.SystemPrompt = input.SystemPrompt
	cfg.SkillsMode = input.SkillsMode
	cfg.EnableSkills = input.EnableSkills
	cfg.SkillsPrompt = input.SkillsPrompt
	cfg.SelectedSkills = input.SelectedSkills

	inRate, outRate := llmcost.GetDefaultPricing(input.Provider, input.ModelName)
	cfg.CostPer1MInput = inRate
	cfg.CostPer1MOutput = outRate

	if input.MaxTokens > 0 {
		cfg.MaxOutputTokens = input.MaxTokens
	}
	if input.APIKey != "" {
		enc, err := crypto.Encrypt(input.APIKey, uc.encryptionKey)
		if err != nil {
			return nil, fault.Wrap(fault.KindUnknown, fmt.Errorf("failed to encrypt api key: %w", err))
		}
		cfg.APIKeyEncrypted = enc
	}

	if err := uc.repo.Update(ctx, cfg); err != nil {
		return nil, fmt.Errorf("update config: %w", err)
	}
	return cfg, nil
}

// SetActive marks the specified LLM configuration as the default active model.
func (uc *ManageConfigUseCase) SetActive(ctx context.Context, id uint) error {
	if uc.repo == nil {
		return domainllm.ErrNotFound
	}
	if err := uc.repo.SetActive(ctx, id); err != nil {
		return fmt.Errorf("set active config: %w", err)
	}
	return nil
}

// Delete removes an LLM configuration by ID.
func (uc *ManageConfigUseCase) Delete(ctx context.Context, id uint) error {
	if uc.repo == nil {
		return domainllm.ErrNotFound
	}
	if err := uc.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete config: %w", err)
	}
	return nil
}

// TestConnection verifies connectivity to the LLM model.
func (uc *ManageConfigUseCase) TestConnection(ctx context.Context, id uint) (string, error) {
	if uc.repo == nil {
		return "", domainllm.ErrNotFound
	}
	cfg, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return "", fmt.Errorf("find config by id: %w", err)
	}

	var apiKey string
	if cfg.APIKeyEncrypted != "" {
		dec, err := crypto.Decrypt(cfg.APIKeyEncrypted, uc.encryptionKey)
		if err != nil {
			return "", fmt.Errorf("failed to decrypt api key: %w", err)
		}
		apiKey = dec
	}

	if uc.tester != nil {
		if err := uc.tester(ctx, cfg, apiKey); err != nil {
			return "", err
		}
	}

	return fmt.Sprintf("Koneksi ke model AI (%s - %s) berhasil diuji!", cfg.Provider, cfg.Model), nil
}
