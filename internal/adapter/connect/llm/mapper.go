package llm

import (
	"fmt"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	domainllm "github.com/quixiq/polyglot/internal/domain/llm"
)

// ToProtoLLMConfig maps domain Config to proto LLMConfig message.
func ToProtoLLMConfig(cfg *domainllm.Config) *devicepb.LLMConfig {
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

// ToProtoLLMConfigList converts a slice of domain Configs into proto messages.
func ToProtoLLMConfigList(configs []domainllm.Config) []*devicepb.LLMConfig {
	items := make([]*devicepb.LLMConfig, len(configs))
	for i := range configs {
		items[i] = ToProtoLLMConfig(&configs[i])
	}
	return items
}
