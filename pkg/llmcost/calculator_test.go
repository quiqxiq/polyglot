package llmcost_test

import (
	"testing"

	"github.com/quixiq/polyglot/pkg/llmcost"
	"github.com/stretchr/testify/assert"
)

func TestGetDefaultPricing(t *testing.T) {
	tests := []struct {
		provider string
		model    string
		wantIn   float64
		wantOut  float64
	}{
		{"gemini", "gemini-2.0-flash", 0.10, 0.40},
		{"openai", "gpt-4o-mini", 0.15, 0.60},
		{"groq", "llama-3.3-70b-versatile", 0.59, 0.79},
		{"ollama", "llama3.2", 0.0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.provider+"_"+tt.model, func(t *testing.T) {
			in, out := llmcost.GetDefaultPricing(tt.provider, tt.model)
			assert.Equal(t, tt.wantIn, in)
			assert.Equal(t, tt.wantOut, out)
		})
	}
}

func TestCalculateCost(t *testing.T) {
	t.Run("calculates correctly for 1000 in and 500 out on gemini-2.0-flash", func(t *testing.T) {
		// in: 1000 / 1M * 0.10 = $0.0001
		// out: 500 / 1M * 0.40 = $0.0002
		// total USD: $0.0003
		// total IDR: 0.0003 * 16000 = Rp 4.80
		costUSD, costIDR := llmcost.CalculateCost("gemini", "gemini-2.0-flash", 1000, 500, 0, 0)
		assert.InDelta(t, 0.0003, costUSD, 0.00001)
		assert.InDelta(t, 4.80, costIDR, 0.01)
	})

	t.Run("uses custom rates if provided", func(t *testing.T) {
		costUSD, _ := llmcost.CalculateCost("custom", "custom-model", 1_000_000, 1_000_000, 1.0, 2.0)
		assert.InDelta(t, 3.0, costUSD, 0.0001)
	})
}
