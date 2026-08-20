package llm

import "time"

// SkillsMode defines how skills are exposed to the agent runtime.
const (
	SkillsModePrompt = "prompt"
	SkillsModeTools  = "tools"
	SkillsModeBoth   = "both"
)

// Config represents a configured LLM provider with its settings.
type Config struct {
	ID              uint      `json:"id"`
	Provider        string    `json:"provider"`
	Model           string    `json:"model"`
	APIKeyEncrypted string    `json:"-"`
	Params          string    `json:"params"`
	MaxOutputTokens int       `json:"max_output_tokens"`
	BaseURL         string    `json:"base_url,omitempty"`
	Temperature     float64   `json:"temperature,omitempty"`
	SystemPrompt    string    `json:"system_prompt,omitempty"`
	IsActive        bool      `json:"is_active"`
	CostPer1MInput  float64   `json:"cost_per_1m_input"`
	CostPer1MOutput float64   `json:"cost_per_1m_output"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`

	// Skills Configuration (LocalAI compatible)
	EnableSkills   bool     `json:"enable_skills"`
	SkillsMode     string   `json:"skills_mode,omitempty"`     // "prompt", "tools", "both"
	SelectedSkills []string `json:"selected_skills,omitempty"` // Filter list skill aktif
	SkillsPrompt   string   `json:"skills_prompt,omitempty"`   // Custom template rendering prompt

	// Calculated Analytics DTO
	TotalInputTokens  int64   `json:"total_input_tokens"`
	TotalOutputTokens int64   `json:"total_output_tokens"`
	TotalMessages     int64   `json:"total_messages"`
	TotalCostUSD      float64 `json:"total_cost_usd"`
	TotalCostIDR      float64 `json:"total_cost_idr"`
}

// ChatMessage represents a single message in the LLM conversation context.
type ChatMessage struct {
	Role    string `json:"role"` // user, assistant, system, tool
	Content string `json:"content"`
}

// ChatResponse represents the LLM's response including token usage.
type ChatResponse struct {
	Content  string `json:"content"`
	TokenIn  int    `json:"token_in"`
	TokenOut int    `json:"token_out"`
}
