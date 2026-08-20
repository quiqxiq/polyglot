package bot

import (
	"regexp"
	"strings"
)

var (
	thinkTagRegex    = regexp.MustCompile(`(?is)<(?:think|thought)>.*?</(?:think|thought)>`)
	unclosedThinkReg = regexp.MustCompile(`(?is)<(?:think|thought)>.*$`)
)

type Guardrail struct{}

func NewGuardrail() *Guardrail {
	return &Guardrail{}
}

// SanitizeResponse membersihkan output model dari tag <think>, reasoning monologue, dan draf internal.
func (g *Guardrail) SanitizeResponse(response string) string {
	res := strings.TrimSpace(response)
	// 1. Strip complete <think>...</think> and <thought>...</thought> blocks
	res = thinkTagRegex.ReplaceAllString(res, "")
	// 2. Strip unclosed <think>... blocks if model reached token limit during reasoning
	if strings.Contains(strings.ToLower(res), "<think>") || strings.Contains(strings.ToLower(res), "<thought>") {
		res = unclosedThinkReg.ReplaceAllString(res, "")
	}
	// 3. Remove stray tags
	res = strings.ReplaceAll(res, "</think>", "")
	res = strings.ReplaceAll(res, "</thought>", "")
	res = strings.ReplaceAll(res, "<think>", "")
	res = strings.ReplaceAll(res, "<thought>", "")
	res = strings.TrimSpace(res)

	// 4. Strip internal monologue reasoning if model outputs draft steps directly
	for _, marker := range []string{
		"**Final Output Generation:**",
		"**Constructing the final Indonesian response:**",
		"**Final Response:**",
		"**Final Output:**",
	} {
		if idx := strings.Index(res, marker); idx != -1 {
			res = strings.TrimSpace(res[idx+len(marker):])
		}
	}

	return strings.TrimSpace(res)
}
