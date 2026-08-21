package port

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/setting"
)

// SettingRepository defines persistence and retrieval operations for system configuration.
type SettingRepository interface {
	Get(ctx context.Context, key string) (*setting.Setting, error)
	GetValue(ctx context.Context, key string, fallback string) string
	Set(ctx context.Context, key, value, category, description string) error
	GetByCategory(ctx context.Context, category string) ([]setting.Setting, error)
	GetAll(ctx context.Context) ([]setting.Setting, error)
	BatchSet(ctx context.Context, settings []setting.Setting) error

	// Typed helper for Bot operational parameters
	GetBotSettings(ctx context.Context) (*setting.BotSettings, error)
	SaveBotSettings(ctx context.Context, s *setting.BotSettings) error
}
