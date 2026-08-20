package llm

import "context"

// ToolFunc is the signature of the execution handler for a Tool.
// It receives the context and raw JSON arguments string, returning a text or JSON output string.
type ToolFunc func(ctx context.Context, argsJSON string) (string, error)

// Tool represents a callable function tool that can be exposed to the LLM agent runtime.
type Tool struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	InputSchema any      `json:"-"`
	Handler     ToolFunc `json:"-"`
}
