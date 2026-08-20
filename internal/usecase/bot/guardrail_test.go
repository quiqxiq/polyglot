package bot_test

import (
	"testing"

	"github.com/quixiq/polyglot/internal/usecase/bot"
	"github.com/stretchr/testify/assert"
)

func TestGuardrail_SanitizeResponse(t *testing.T) {
	g := bot.NewGuardrail([]string{"paket", "harga"})

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Normal clean response",
			input:    "Halo Kak, ada yang bisa dibantu?",
			expected: "Halo Kak, ada yang bisa dibantu?",
		},
		{
			name:     "Complete <think> tag stripped",
			input:    "<think>User wants help with package info. Let's provide it.</think>Halo Kak! Ini paket internet kami.",
			expected: "Halo Kak! Ini paket internet kami.",
		},
		{
			name:     "Complete <thought> tag stripped",
			input:    "<thought>\nStep 1: Check greeting\nStep 2: Respond\n</thought>\nSelamat pagi Kak!",
			expected: "Selamat pagi Kak!",
		},
		{
			name:     "Unclosed <think> tag stripped",
			input:    "<think>Internal reasoning step 1, 2, 3...",
			expected: "",
		},
		{
			name: "Reasoning monologue with final output marker",
			input: `1. **Acknowledge & Empathize:** Friendly greeting.
2. **Drafting:** Halo kak.
**Final Output Generation:**
Halo Kak! Selamat datang di layanan kami.`,
			expected: "Halo Kak! Selamat datang di layanan kami.",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := g.SanitizeResponse(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}
