package bot

import (
	"fmt"
	"time"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/internal/domain/bot"
	"github.com/quixiq/polyglot/internal/domain/llm"
	"github.com/quixiq/polyglot/internal/domain/skill"
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
		Id:             fmt.Sprintf("%d", cfg.ID),
		Provider:       cfg.Provider,
		ModelName:      cfg.Model,
		BaseUrl:        cfg.BaseURL,
		IsActive:       cfg.IsActive,
		Temperature:    cfg.Temperature,
		MaxTokens:      int32(cfg.MaxOutputTokens),
		SystemPrompt:   cfg.SystemPrompt,
		SkillsMode:     cfg.SkillsMode,
		EnableSkills:   cfg.EnableSkills,
		SkillsPrompt:   cfg.SkillsPrompt,
		SelectedSkills: cfg.SelectedSkills,
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

// --- Skills Mappers ---

func toProtoSkill(s *skill.Skill) *devicepb.Skill {
	if s == nil {
		return &devicepb.Skill{}
	}
	return &devicepb.Skill{
		Id:            s.ID,
		Name:          s.Name,
		Description:   s.Description,
		Content:       s.Content,
		License:       s.License,
		Compatibility: s.Compatibility,
		AllowedTools:  s.AllowedTools,
		Metadata:      s.Metadata,
		ReadOnly:      s.ReadOnly,
		SourceType:    s.SourceType,
		SourceUrl:     s.SourceURL,
		CreatedAt:     s.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     s.UpdatedAt.Format(time.RFC3339),
	}
}

func toProtoSkillList(skills []skill.Skill) []*devicepb.Skill {
	items := make([]*devicepb.Skill, len(skills))
	for i := range skills {
		items[i] = toProtoSkill(&skills[i])
	}
	return items
}

func toProtoSkillResource(r *skill.SkillResource) *devicepb.SkillResource {
	if r == nil {
		return &devicepb.SkillResource{}
	}
	return &devicepb.SkillResource{
		Path:     r.Path,
		Name:     r.Name,
		Type:     r.Type,
		Size:     r.Size,
		MimeType: r.MimeType,
		Readable: r.Readable,
		Modified: r.Modified.Format(time.RFC3339),
	}
}

func toProtoSkillResourceList(resources []skill.SkillResource) []*devicepb.SkillResource {
	items := make([]*devicepb.SkillResource, len(resources))
	for i := range resources {
		items[i] = toProtoSkillResource(&resources[i])
	}
	return items
}

func toProtoResourceContent(c *skill.ResourceContent) *devicepb.ResourceContent {
	if c == nil {
		return &devicepb.ResourceContent{}
	}
	return &devicepb.ResourceContent{
		Content:  c.Content,
		Encoding: c.Encoding,
		MimeType: c.MimeType,
		Size:     c.Size,
	}
}

func toProtoGitRepoInfo(g *skill.GitRepoInfo) *devicepb.GitRepoInfo {
	if g == nil {
		return &devicepb.GitRepoInfo{}
	}
	return &devicepb.GitRepoInfo{
		Id:           g.ID,
		Url:          g.URL,
		Name:         g.Name,
		Enabled:      g.Enabled,
		LastSyncedAt: g.LastSyncedAt.Format(time.RFC3339),
	}
}

func toProtoGitRepoInfoList(repos []skill.GitRepoInfo) []*devicepb.GitRepoInfo {
	items := make([]*devicepb.GitRepoInfo, len(repos))
	for i := range repos {
		items[i] = toProtoGitRepoInfo(&repos[i])
	}
	return items
}
