package llm

import (
	"context"
	"errors"
	"testing"

	domainllm "github.com/quixiq/polyglot/internal/domain/llm"
)

type mockLLMRepo struct {
	configs []domainllm.Config
	active  *domainllm.Config
	byID    map[uint]*domainllm.Config
}

func newMockLLMRepo() *mockLLMRepo {
	return &mockLLMRepo{
		byID: make(map[uint]*domainllm.Config),
	}
}

func (m *mockLLMRepo) Create(ctx context.Context, cfg *domainllm.Config) error {
	cfg.ID = uint(len(m.configs) + 1)
	m.configs = append(m.configs, *cfg)
	m.byID[cfg.ID] = cfg
	if cfg.IsActive {
		m.active = cfg
	}
	return nil
}

func (m *mockLLMRepo) FindByID(ctx context.Context, id uint) (*domainllm.Config, error) {
	if cfg, ok := m.byID[id]; ok {
		return cfg, nil
	}
	return nil, domainllm.ErrNotFound
}

func (m *mockLLMRepo) FindActive(ctx context.Context) (*domainllm.Config, error) {
	if m.active != nil {
		return m.active, nil
	}
	return nil, domainllm.ErrNotFound
}

func (m *mockLLMRepo) FindAll(ctx context.Context) ([]domainllm.Config, error) {
	return m.configs, nil
}

func (m *mockLLMRepo) Update(ctx context.Context, cfg *domainllm.Config) error {
	m.byID[cfg.ID] = cfg
	return nil
}

func (m *mockLLMRepo) SetActive(ctx context.Context, id uint) error {
	if cfg, ok := m.byID[id]; ok {
		cfg.IsActive = true
		m.active = cfg
		return nil
	}
	return domainllm.ErrNotFound
}

func (m *mockLLMRepo) Delete(ctx context.Context, id uint) error {
	delete(m.byID, id)
	return nil
}

func TestManageConfigUseCase(t *testing.T) {
	ctx := context.Background()
	repo := newMockLLMRepo()
	key := "12345678901234567890123456789012"
	testerCalled := false
	tester := func(ctx context.Context, cfg *domainllm.Config, apiKey string) error {
		testerCalled = true
		if apiKey != "secret123" {
			return errors.New("invalid api key")
		}
		return nil
	}

	uc := NewManageConfigUseCase(repo, key, tester)

	// Test Create
	created, err := uc.Create(ctx, CreateConfigInput{
		Provider:    "openai",
		ModelName:   "gpt-4o",
		APIKey:      "secret123",
		Temperature: 0.7,
		MaxTokens:   2048,
	})
	if err != nil {
		t.Fatalf("unexpected error creating config: %v", err)
	}
	if !created.IsActive {
		t.Errorf("expected first config to be auto-activated")
	}
	if created.CostPer1MInput <= 0 {
		t.Errorf("expected default pricing to be computed")
	}

	// Test List
	list, err := uc.List(ctx)
	if err != nil || len(list) != 1 {
		t.Fatalf("expected 1 config, got %d (err: %v)", len(list), err)
	}

	// Test GetByID
	got, err := uc.GetByID(ctx, created.ID)
	if err != nil || got.Model != "gpt-4o" {
		t.Fatalf("expected gpt-4o, got %v", got)
	}

	// Test Update
	updated, err := uc.Update(ctx, UpdateConfigInput{
		ID:          created.ID,
		Provider:    "openai",
		ModelName:   "gpt-4o-mini",
		Temperature: 0.5,
	})
	if err != nil || updated.Model != "gpt-4o-mini" {
		t.Fatalf("expected updated model gpt-4o-mini, got %v", updated)
	}

	// Test TestConnection
	msg, err := uc.TestConnection(ctx, created.ID)
	if err != nil || !testerCalled {
		t.Fatalf("expected successful test connection, got msg: %s, err: %v", msg, err)
	}

	// Test SetActive
	if err := uc.SetActive(ctx, created.ID); err != nil {
		t.Fatalf("failed to set active: %v", err)
	}

	// Test Delete
	if err := uc.Delete(ctx, created.ID); err != nil {
		t.Fatalf("failed to delete: %v", err)
	}
}
