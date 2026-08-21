package setting

import (
	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/internal/domain/setting"
)

// DomainSettingToPb converts domain Setting entity to proto SettingItem message.
func DomainSettingToPb(s *setting.Setting) *devicepb.SettingItem {
	if s == nil {
		return nil
	}
	return &devicepb.SettingItem{
		Key:           s.Key,
		Value:         s.Value,
		Category:      s.Category,
		Description:   s.Description,
		UpdatedAtUnix: s.UpdatedAt.Unix(),
	}
}

// DomainSettingsToPb converts slice of domain Setting entities to proto slice.
func DomainSettingsToPb(items []setting.Setting) []*devicepb.SettingItem {
	res := make([]*devicepb.SettingItem, 0, len(items))
	for i := range items {
		res = append(res, DomainSettingToPb(&items[i]))
	}
	return res
}

// PbSettingToDomain converts proto SettingItem to domain Setting entity.
func PbSettingToDomain(pb *devicepb.SettingItem) *setting.Setting {
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

// DomainBotSettingsToPb converts domain BotSettings entity to proto BotSettings message.
func DomainBotSettingsToPb(s *setting.BotSettings) *devicepb.BotSettings {
	if s == nil {
		return nil
	}
	return &devicepb.BotSettings{
		BurstLimit:            int32(s.BurstLimit),
		BurstWindowSecs:       int32(s.BurstWindowSecs),
		Mute_1HSecs:           int32(s.Mute1HourSecs),
		Ban_24HSecs:           int32(s.Ban24HourSecs),
		DailyChatLimit:        int32(s.DailyChatLimit),
		SessionTimeoutMinutes: int32(s.SessionTimeoutMinutes),
		SlidingWindowSize:     int32(s.SlidingWindowSize),
		LlmMaxOutputTokens:    int32(s.LLMMaxOutputTokens),
		WhitelistAllStaff:     s.WhitelistAllStaff,
		CustomWhitelistPhones: s.CustomWhitelistPhones,
	}
}

// PbBotSettingsToDomain converts proto BotSettings message to domain BotSettings entity.
func PbBotSettingsToDomain(pb *devicepb.BotSettings) *setting.BotSettings {
	if pb == nil {
		return setting.DefaultBotSettings()
	}
	return &setting.BotSettings{
		BurstLimit:            int(pb.BurstLimit),
		BurstWindowSecs:       int(pb.BurstWindowSecs),
		Mute1HourSecs:         int(pb.Mute_1HSecs),
		Ban24HourSecs:         int(pb.Ban_24HSecs),
		DailyChatLimit:        int(pb.DailyChatLimit),
		SessionTimeoutMinutes: int(pb.SessionTimeoutMinutes),
		SlidingWindowSize:     int(pb.SlidingWindowSize),
		LLMMaxOutputTokens:    int(pb.LlmMaxOutputTokens),
		WhitelistAllStaff:     pb.WhitelistAllStaff,
		CustomWhitelistPhones: pb.CustomWhitelistPhones,
	}
}
