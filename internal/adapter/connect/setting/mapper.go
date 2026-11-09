package setting

import (
	polyglotv1 "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/internal/domain/setting"
)

// DomainSettingToPb converts domain Setting entity to proto SettingItem message.
func DomainSettingToPb(s *setting.Setting) *polyglotv1.SettingItem {
	if s == nil {
		return nil
	}
	return &polyglotv1.SettingItem{
		Key:           s.Key,
		Value:         s.Value,
		Category:      s.Category,
		Description:   s.Description,
		UpdatedAtUnix: s.UpdatedAt.Unix(),
	}
}

// DomainSettingsToPb converts slice of domain Setting entities to proto slice.
func DomainSettingsToPb(items []setting.Setting) []*polyglotv1.SettingItem {
	res := make([]*polyglotv1.SettingItem, 0, len(items))
	for i := range items {
		res = append(res, DomainSettingToPb(&items[i]))
	}
	return res
}

// PbSettingToDomain converts proto SettingItem to domain Setting entity.
func PbSettingToDomain(pb *polyglotv1.SettingItem) *setting.Setting {
	if pb == nil {
		return nil
	}
	return &setting.Setting{
		Key:         pb.Key,
		Value:       pb.Value,
		Category:    pb.Category,
		Description: pb.Description,
	}
}
