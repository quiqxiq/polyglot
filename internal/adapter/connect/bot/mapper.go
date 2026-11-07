package bot

import (
	"fmt"
	"time"

	polyglotv1 "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/internal/domain/bot"
	"github.com/quixiq/polyglot/internal/domain/llm"
	"github.com/quixiq/polyglot/internal/domain/skill"
)

// toProtoWASession converts a domain bot.WASession into its Protobuf representation.
func toProtoWASession(s *bot.WASession) *polyglotv1.WASession {
	if s == nil {
		return &polyglotv1.WASession{}
	}
	return &polyglotv1.WASession{
		Id:          fmt.Sprintf("%d", s.ID),
		Name:        s.DeviceName,
		PhoneNumber: s.PhoneNumber,
		Status:      string(s.Status),
		IsBotActive: s.IsBotEnabled,
		CreatedAt:   s.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

// toProtoWASessionList converts a slice of domain bot.WASession into Protobuf representations.
func toProtoWASessionList(sessions []bot.WASession) []*polyglotv1.WASession {
	items := make([]*polyglotv1.WASession, len(sessions))
	for i := range sessions {
		items[i] = toProtoWASession(&sessions[i])
	}
	return items
}

// toProtoLLMConfig converts a domain llm.Config into its Protobuf representation.
func toProtoLLMConfig(cfg *llm.Config) *polyglotv1.LLMConfig {
	if cfg == nil {
		return &polyglotv1.LLMConfig{}
	}
	return &polyglotv1.LLMConfig{
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
func toProtoLLMConfigList(configs []llm.Config) []*polyglotv1.LLMConfig {
	items := make([]*polyglotv1.LLMConfig, len(configs))
	for i := range configs {
		items[i] = toProtoLLMConfig(&configs[i])
	}
	return items
}

// toProtoConversation converts a domain bot.Conversation into its Protobuf representation.
func toProtoConversation(c *bot.Conversation) *polyglotv1.Conversation {
	if c == nil {
		return &polyglotv1.Conversation{}
	}
	return &polyglotv1.Conversation{
		Id:          fmt.Sprintf("%d", c.ID),
		SessionId:   fmt.Sprintf("%d", c.SessionID),
		ClientPhone: c.CustomerWANumber,
		Status:      string(c.Status),
		UpdatedAt:   c.UpdatedAt.Format(time.RFC3339),
	}
}

// toProtoConversationList converts a slice of domain bot.Conversation into Protobuf representations.
func toProtoConversationList(convs []bot.Conversation) []*polyglotv1.Conversation {
	items := make([]*polyglotv1.Conversation, len(convs))
	for i := range convs {
		items[i] = toProtoConversation(&convs[i])
	}
	return items
}

// --- Skills Mappers ---

func toProtoSkill(s *skill.Skill) *polyglotv1.Skill {
	if s == nil {
		return &polyglotv1.Skill{}
	}
	return &polyglotv1.Skill{
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

func toProtoSkillList(skills []skill.Skill) []*polyglotv1.Skill {
	items := make([]*polyglotv1.Skill, len(skills))
	for i := range skills {
		items[i] = toProtoSkill(&skills[i])
	}
	return items
}

func toProtoSkillResource(r *skill.SkillResource) *polyglotv1.SkillResource {
	if r == nil {
		return &polyglotv1.SkillResource{}
	}
	return &polyglotv1.SkillResource{
		Path:     r.Path,
		Name:     r.Name,
		Type:     r.Type,
		Size:     r.Size,
		MimeType: r.MimeType,
		Readable: r.Readable,
		Modified: r.Modified.Format(time.RFC3339),
	}
}

func toProtoSkillResourceList(resources []skill.SkillResource) []*polyglotv1.SkillResource {
	items := make([]*polyglotv1.SkillResource, len(resources))
	for i := range resources {
		items[i] = toProtoSkillResource(&resources[i])
	}
	return items
}

func toProtoResourceContent(c *skill.ResourceContent) *polyglotv1.ResourceContent {
	if c == nil {
		return &polyglotv1.ResourceContent{}
	}
	return &polyglotv1.ResourceContent{
		Content:  c.Content,
		Encoding: c.Encoding,
		MimeType: c.MimeType,
		Size:     c.Size,
	}
}

func toProtoGitRepoInfo(g *skill.GitRepoInfo) *polyglotv1.GitRepoInfo {
	if g == nil {
		return &polyglotv1.GitRepoInfo{}
	}
	return &polyglotv1.GitRepoInfo{
		Id:           g.ID,
		Url:          g.URL,
		Name:         g.Name,
		Enabled:      g.Enabled,
		LastSyncedAt: g.LastSyncedAt.Format(time.RFC3339),
	}
}

func toProtoGitRepoInfoList(repos []skill.GitRepoInfo) []*polyglotv1.GitRepoInfo {
	items := make([]*polyglotv1.GitRepoInfo, len(repos))
	for i := range repos {
		items[i] = toProtoGitRepoInfo(&repos[i])
	}
	return items
}
