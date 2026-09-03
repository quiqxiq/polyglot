package setting

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/setting"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/logger"
)

type ManageSettingUseCase struct {
	repo port.SettingRepository
}

func NewManageSettingUseCase(repo port.SettingRepository) *ManageSettingUseCase {
	return &ManageSettingUseCase{
		repo: repo,
	}
}

func (u *ManageSettingUseCase) GetAllSettings(ctx context.Context) ([]setting.Setting, error) {
	if u.repo == nil {
		return nil, setting.ErrRepositoryNotConfigured
	}
	return u.repo.GetAll(ctx)
}

func (u *ManageSettingUseCase) GetSettingsByCategory(ctx context.Context, category string) ([]setting.Setting, error) {
	if u.repo == nil {
		return nil, setting.ErrRepositoryNotConfigured
	}
	return u.repo.GetByCategory(ctx, category)
}

func (u *ManageSettingUseCase) UpdateSetting(ctx context.Context, key, value string) (*setting.Setting, error) {
	if u.repo == nil {
		return nil, setting.ErrRepositoryNotConfigured
	}

	existing, err := u.repo.Get(ctx, key)
	cat := "general"
	desc := ""
	if err == nil && existing != nil {
		cat = existing.Category
		desc = existing.Description
	}

	if err := u.repo.Set(ctx, key, value, cat, desc); err != nil {
		return nil, err
	}

	logger.WithComponent("SettingUseCase").WithField("key", key).Info("setting updated successfully")
	return u.repo.Get(ctx, key)
}

func (u *ManageSettingUseCase) BatchUpdateSettings(ctx context.Context, settings []setting.Setting) error {
	if u.repo == nil {
		return setting.ErrRepositoryNotConfigured
	}
	if len(settings) == 0 {
		return nil
	}

	if err := u.repo.BatchSet(ctx, settings); err != nil {
		return err
	}

	logger.WithComponent("SettingUseCase").WithField("setting_count", len(settings)).Info("settings batch updated")
	return nil
}

func (u *ManageSettingUseCase) GetBotSettings(ctx context.Context) (*setting.BotSettings, error) {
	if u.repo == nil {
		return setting.DefaultBotSettings(), nil
	}
	return u.repo.GetBotSettings(ctx)
}

func (u *ManageSettingUseCase) UpdateBotSettings(ctx context.Context, s *setting.BotSettings) error {
	if u.repo == nil {
		return setting.ErrRepositoryNotConfigured
	}
	if err := u.repo.SaveBotSettings(ctx, s); err != nil {
		return err
	}

	logger.WithComponent("SettingUseCase").Info("bot operational settings updated successfully")
	return nil
}
