package port

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/setting"
)

// SettingReader membaca konfigurasi dinamis dari system_settings.
// Interface kecil agar usecase/worker dapat dites tanpa DB.
type SettingReader interface {
	GetValue(ctx context.Context, key, fallback string) string
}

// ISPSettings adalah alias ke domain model per DEVELOPMENT-GUIDELINES.md §4.2.
type ISPSettings = setting.ISPSettings

// LoadISPSettings membaca seluruh key isp.* dengan fallback default.
func LoadISPSettings(ctx context.Context, r SettingReader) ISPSettings {
	return setting.LoadISPSettings(ctx, r)
}
