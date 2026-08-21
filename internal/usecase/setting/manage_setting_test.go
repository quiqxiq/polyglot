package setting_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quixiq/polyglot/internal/domain/setting"
	settingUC "github.com/quixiq/polyglot/internal/usecase/setting"
)

type mockSettingRepo struct {
	settings    map[string]setting.Setting
	botSettings *setting.BotSettings
}

func newMockSettingRepo() *mockSettingRepo {
	return &mockSettingRepo{
		settings:    make(map[string]setting.Setting),
		botSettings: setting.DefaultBotSettings(),
	}
}

func (m *mockSettingRepo) Get(ctx context.Context, key string) (*setting.Setting, error) {
	if s, ok := m.settings[key]; ok {
		return &s, nil
	}
	return nil, nil
}

func (m *mockSettingRepo) GetValue(ctx context.Context, key string, fallback string) string {
	if s, ok := m.settings[key]; ok {
		return s.Value
	}
	return fallback
}

func (m *mockSettingRepo) Set(ctx context.Context, key, value, category, description string) error {
	m.settings[key] = setting.Setting{
		Key:         key,
		Value:       value,
		Category:    category,
		Description: description,
	}
	return nil
}

func (m *mockSettingRepo) GetByCategory(ctx context.Context, category string) ([]setting.Setting, error) {
	var res []setting.Setting
	for _, s := range m.settings {
		if s.Category == category {
			res = append(res, s)
		}
	}
	return res, nil
}

func (m *mockSettingRepo) GetAll(ctx context.Context) ([]setting.Setting, error) {
	var res []setting.Setting
	for _, s := range m.settings {
		res = append(res, s)
	}
	return res, nil
}

func (m *mockSettingRepo) BatchSet(ctx context.Context, settings []setting.Setting) error {
	for _, s := range settings {
		m.settings[s.Key] = s
	}
	return nil
}

func (m *mockSettingRepo) GetBotSettings(ctx context.Context) (*setting.BotSettings, error) {
	return m.botSettings, nil
}

func (m *mockSettingRepo) SaveBotSettings(ctx context.Context, s *setting.BotSettings) error {
	m.botSettings = s
	return nil
}

func TestManageSettingUseCase(t *testing.T) {
	ctx := context.Background()
	repo := newMockSettingRepo()
	uc := settingUC.NewManageSettingUseCase(repo)

	// Test Update & Get
	_, err := uc.UpdateSetting(ctx, "app.name", "Polyglot NetOps")
	require.NoError(t, err)

	val, err := uc.GetSettingsByCategory(ctx, "general")
	require.NoError(t, err)
	assert.Len(t, val, 1)
	assert.Equal(t, "Polyglot NetOps", val[0].Value)

	// Test Bot Settings
	botCfg, err := uc.GetBotSettings(ctx)
	require.NoError(t, err)
	assert.Equal(t, 3, botCfg.BurstLimit)

	botCfg.BurstLimit = 5
	botCfg.DailyChatLimit = 25
	botCfg.CustomWhitelistPhones = "628111222333"
	err = uc.UpdateBotSettings(ctx, botCfg)
	require.NoError(t, err)

	updated, err := uc.GetBotSettings(ctx)
	require.NoError(t, err)
	assert.Equal(t, 5, updated.BurstLimit)
	assert.Equal(t, 25, updated.DailyChatLimit)
	assert.Equal(t, "628111222333", updated.CustomWhitelistPhones)
}
