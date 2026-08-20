package genkit_test

import (
	"context"
	"testing"

	"github.com/quixiq/polyglot/internal/adapter/llm/genkit"
	"github.com/quixiq/polyglot/internal/domain/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewProvider_Validation(t *testing.T) {
	ctx := context.Background()

	t.Run("nil config returns error", func(t *testing.T) {
		p, err := genkit.NewProvider(ctx, nil, "any-key")
		assert.Error(t, err)
		assert.Nil(t, p)
	})

	t.Run("missing api key for gemini returns error", func(t *testing.T) {
		p, err := genkit.NewProvider(ctx, &llm.Config{Provider: "gemini", Model: "gemini-2.0-flash"}, "")
		assert.Error(t, err)
		assert.Nil(t, p)
	})

	t.Run("missing api key for openai returns error", func(t *testing.T) {
		p, err := genkit.NewProvider(ctx, &llm.Config{Provider: "openai", Model: "gpt-4o"}, "")
		assert.Error(t, err)
		assert.Nil(t, p)
	})

	t.Run("missing api key for groq returns error", func(t *testing.T) {
		p, err := genkit.NewProvider(ctx, &llm.Config{Provider: "groq", Model: "llama-3.3-70b-versatile"}, "")
		assert.Error(t, err)
		assert.Nil(t, p)
	})

	t.Run("missing api key for claude returns error", func(t *testing.T) {
		p, err := genkit.NewProvider(ctx, &llm.Config{Provider: "claude", Model: "claude-3-5-sonnet"}, "")
		assert.Error(t, err)
		assert.Nil(t, p)
	})

	t.Run("unsupported provider returns error", func(t *testing.T) {
		p, err := genkit.NewProvider(ctx, &llm.Config{Provider: "unsupported-provider"}, "any-key")
		assert.Error(t, err)
		assert.Nil(t, p)
	})

	t.Run("gemini provider initialized successfully", func(t *testing.T) {
		p, err := genkit.NewProvider(ctx, &llm.Config{
			Provider:        "gemini",
			Model:           "gemini-2.0-flash",
			MaxOutputTokens: 2048,
		}, "fake-gemini-key")
		require.NoError(t, err)
		assert.NotNil(t, p)
	})

	t.Run("openai provider initialized successfully", func(t *testing.T) {
		p, err := genkit.NewProvider(ctx, &llm.Config{
			Provider:        "openai",
			Model:           "gpt-4o-mini",
			MaxOutputTokens: 1024,
		}, "fake-openai-key")
		require.NoError(t, err)
		assert.NotNil(t, p)
	})

	t.Run("groq provider initialized successfully with default base url", func(t *testing.T) {
		p, err := genkit.NewProvider(ctx, &llm.Config{
			Provider:        "groq",
			Model:           "llama-3.3-70b-versatile",
			MaxOutputTokens: 2048,
		}, "fake-groq-key")
		require.NoError(t, err)
		assert.NotNil(t, p)
	})

	t.Run("claude provider initialized successfully", func(t *testing.T) {
		p, err := genkit.NewProvider(ctx, &llm.Config{
			Provider:        "claude",
			Model:           "claude-3-5-sonnet",
			MaxOutputTokens: 4096,
		}, "fake-claude-key")
		require.NoError(t, err)
		assert.NotNil(t, p)
	})

	t.Run("ollama provider initialized successfully", func(t *testing.T) {
		p, err := genkit.NewProvider(ctx, &llm.Config{
			Provider:        "ollama",
			Model:           "llama3.2",
			MaxOutputTokens: 1024,
		}, "")
		require.NoError(t, err)
		assert.NotNil(t, p)
	})

	t.Run("custom provider initialized successfully", func(t *testing.T) {
		p, err := genkit.NewProvider(ctx, &llm.Config{
			Provider:        "custom",
			Model:           "meta-llama/llama-3.3-70b-instruct",
			Params:          `{"base_url":"https://openrouter.ai/api/v1"}`,
			MaxOutputTokens: 1024,
		}, "fake-openrouter-key")
		require.NoError(t, err)
		assert.NotNil(t, p)
	})
}
