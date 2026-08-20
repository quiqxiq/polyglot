package llmcost

import (
	"strings"
)

const (
	// DefaultUSDToIDR defines the baseline currency conversion rate.
	DefaultUSDToIDR = 16000.0
)

// ModelPricing represents the cost in USD per 1 Million tokens.
type ModelPricing struct {
	InputPer1M  float64
	OutputPer1M float64
}

type modelRate struct {
	fragment string
	pricing  ModelPricing
}

// orderedPricing ensures specific model prefixes (e.g. gpt-4o-mini) are checked before generic ones (gpt-4o).
var orderedPricing = []modelRate{
	{"gpt-4o-mini", ModelPricing{InputPer1M: 0.15, OutputPer1M: 0.60}},
	{"gpt-4o", ModelPricing{InputPer1M: 2.50, OutputPer1M: 10.00}},
	{"o3-mini", ModelPricing{InputPer1M: 1.10, OutputPer1M: 4.40}},
	{"gemini-2.0-flash-lite", ModelPricing{InputPer1M: 0.075, OutputPer1M: 0.30}},
	{"gemini-2.0-flash", ModelPricing{InputPer1M: 0.10, OutputPer1M: 0.40}},
	{"gemini-1.5-flash", ModelPricing{InputPer1M: 0.075, OutputPer1M: 0.30}},
	{"gemini-1.5-pro", ModelPricing{InputPer1M: 1.25, OutputPer1M: 5.00}},
	{"qwen3.6-27b", ModelPricing{InputPer1M: 0.60, OutputPer1M: 3.00}},
	{"gpt-oss-120b", ModelPricing{InputPer1M: 0.15, OutputPer1M: 0.60}},
	{"gpt-oss-20b", ModelPricing{InputPer1M: 0.075, OutputPer1M: 0.30}},
	{"llama-3.3-70b-versatile", ModelPricing{InputPer1M: 0.59, OutputPer1M: 0.79}},
	{"llama-3.1-8b-instant", ModelPricing{InputPer1M: 0.05, OutputPer1M: 0.08}},
	{"deepseek-reasoner", ModelPricing{InputPer1M: 0.55, OutputPer1M: 2.19}},
	{"deepseek-chat", ModelPricing{InputPer1M: 0.14, OutputPer1M: 0.28}},
	{"claude-3-5-haiku", ModelPricing{InputPer1M: 0.80, OutputPer1M: 4.00}},
	{"claude-3-5-sonnet", ModelPricing{InputPer1M: 3.00, OutputPer1M: 15.00}},
}

// GetDefaultPricing returns the default cost per 1M tokens in USD for a given provider/model.
func GetDefaultPricing(provider, model string) (inputCost float64, outputCost float64) {
	prov := strings.ToLower(strings.TrimSpace(provider))
	if prov == "ollama" {
		return 0.0, 0.0
	}

	modelLower := strings.ToLower(strings.TrimSpace(model))
	for _, item := range orderedPricing {
		if strings.Contains(modelLower, item.fragment) {
			return item.pricing.InputPer1M, item.pricing.OutputPer1M
		}
	}

	// Default fallbacks based on provider
	switch prov {
	case "gemini":
		return 0.10, 0.40
	case "openai":
		return 0.15, 0.60
	case "groq":
		return 0.59, 0.79
	case "claude":
		return 0.80, 4.00
	case "deepseek":
		return 0.14, 0.28
	default:
		return 0.10, 0.40
	}
}

// CalculateCost computes total cost in USD and IDR for given input/output tokens.
func CalculateCost(provider, model string, inputTokens, outputTokens int, customInputRate, customOutputRate float64) (costUSD float64, costIDR float64) {
	inRate := customInputRate
	outRate := customOutputRate

	if inRate <= 0 && outRate <= 0 {
		inRate, outRate = GetDefaultPricing(provider, model)
	}

	inCost := (float64(inputTokens) / 1_000_000.0) * inRate
	outCost := (float64(outputTokens) / 1_000_000.0) * outRate

	costUSD = inCost + outCost
	costIDR = costUSD * DefaultUSDToIDR
	return costUSD, costIDR
}
