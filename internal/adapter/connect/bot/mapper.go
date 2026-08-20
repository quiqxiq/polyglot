package bot

import (
	"fmt"
	"time"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/internal/domain/bot"
	"github.com/quixiq/polyglot/internal/domain/llm"
)

// toProtoWASession converts a domain bot.WASession into its Protobuf representation.
func toProtoWASession(s *bot.WASession) *devicepb.WASession {
	if s == nil {
		return &devicepb.WASession{}
	}
	return &devicepb.WASession{
		Id:          fmt.Sprintf("%d", s.ID),
		Name:        s.DeviceName,
		PhoneNumber: s.PhoneNumber,
		Status:      string(s.Status),
		IsBotActive: s.IsBotEnabled,
		CreatedAt:   s.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

// toProtoWASessionList converts a slice of domain bot.WASession into Protobuf representations.
func toProtoWASessionList(sessions []bot.WASession) []*devicepb.WASession {
	items := make([]*devicepb.WASession, len(sessions))
	for i := range sessions {
		items[i] = toProtoWASession(&sessions[i])
	}
	return items
}

// toProtoLLMConfig converts a domain llm.Config into its Protobuf representation.
func toProtoLLMConfig(cfg *llm.Config) *devicepb.LLMConfig {
	if cfg == nil {
		return &devicepb.LLMConfig{}
	}
	return &devicepb.LLMConfig{
		Id:           fmt.Sprintf("%d", cfg.ID),
		Provider:     cfg.Provider,
		ModelName:    cfg.Model,
		IsActive:     cfg.IsActive,
		MaxTokens:    int32(cfg.MaxOutputTokens),
		SystemPrompt: "",
	}
}

// toProtoLLMConfigList converts a slice of domain llm.Config into Protobuf representations.
func toProtoLLMConfigList(configs []llm.Config) []*devicepb.LLMConfig {
	items := make([]*devicepb.LLMConfig, len(configs))
	for i := range configs {
		items[i] = toProtoLLMConfig(&configs[i])
	}
	return items
}

// toProtoConversation converts a domain bot.Conversation into its Protobuf representation.
func toProtoConversation(c *bot.Conversation) *devicepb.Conversation {
	if c == nil {
		return &devicepb.Conversation{}
	}
	return &devicepb.Conversation{
		Id:          fmt.Sprintf("%d", c.ID),
		SessionId:   fmt.Sprintf("%d", c.SessionID),
		ClientPhone: c.CustomerWANumber,
		Status:      string(c.Status),
		UpdatedAt:   c.UpdatedAt.Format(time.RFC3339),
	}
}

// toProtoConversationList converts a slice of domain bot.Conversation into Protobuf representations.
func toProtoConversationList(convs []bot.Conversation) []*devicepb.Conversation {
	items := make([]*devicepb.Conversation, len(convs))
	for i := range convs {
		items[i] = toProtoConversation(&convs[i])
	}
	return items
}
