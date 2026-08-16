package llm

import "strings"

// ModelPricing defines official pricing per 1 Million (1M) tokens in USD.
type ModelPricing struct {
	CostPer1MInput  float64
	CostPer1MOutput float64
}

// GetDefaultModelPricing returns official pricing rates per 1M tokens in USD.
func GetDefaultModelPricing(provider, model string) (costInput, costOutput float64) {
	provLower := strings.ToLower(strings.TrimSpace(provider))
	modLower := strings.ToLower(strings.TrimSpace(model))

	switch provLower {
	case "gemini":
		if strings.Contains(modLower, "flash-lite") {
			return 0.0375, 0.150
		}
		if strings.Contains(modLower, "pro") {
			return 1.250, 5.000
		}
		return 0.075, 0.300

	case "openai":
		if strings.Contains(modLower, "gpt-4o-mini") || strings.Contains(modLower, "mini") {
			return 0.150, 0.600
		}
		if strings.Contains(modLower, "gpt-4o") || strings.Contains(modLower, "gpt-4") {
			return 2.500, 10.000
		}
		if strings.Contains(modLower, "gpt-3.5") {
			return 0.500, 1.500
		}
		return 0.150, 0.600

	case "claude":
		if strings.Contains(modLower, "haiku") {
			return 0.800, 4.000
		}
		if strings.Contains(modLower, "sonnet") || strings.Contains(modLower, "opus") {
			return 3.000, 15.000
		}
		return 1.000, 5.000

	case "groq":
		if strings.Contains(modLower, "70b") {
			return 0.590, 0.790
		}
		if strings.Contains(modLower, "8b") {
			return 0.050, 0.080
		}
		if strings.Contains(modLower, "mixtral") {
			return 0.240, 0.240
		}
		return 0.100, 0.200
	}

	return 0.100, 0.300
}
