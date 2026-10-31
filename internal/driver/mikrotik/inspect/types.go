package inspect

// Node is one child node from /console/inspect request=child.
type Node struct {
	Type     string `json:"type"`
	Name     string `json:"name"`
	NodeType string `json:"node-type"`
}

// Syntax is one syntax/argument node from request=syntax.
type Syntax struct {
	Type       string `json:"type"`
	Symbol     string `json:"symbol"`
	SymbolType string `json:"symbol-type"`
	Nested     string `json:"nested"`
	Nonorm     string `json:"nonorm"`
	Text       string `json:"text"`
}

// Completion is one completion suggestion from request=completion.
type Completion struct {
	Type       string `json:"type"`
	Completion string `json:"completion"`
	Style      string `json:"style"`
	Offset     int    `json:"offset"`
	Preference int    `json:"preference"`
	Show       bool   `json:"show"`
	Text       string `json:"text"`
}

// Highlight is one highlighted token from request=highlight.
type Highlight struct {
	Type  string `json:"type"`
	Text  string `json:"text"`
	Start int    `json:"start"`
	End   int    `json:"end"`
}

// ValidationResult is the pre-approval syntax check result.
//
// Supported=false means the device is RouterOS v6 (or the path/version is
// deliberately skipped), so the caller must show a degraded-mode warning
// instead of claiming the command is valid.
type ValidationResult struct {
	Supported bool        `json:"supported"`
	Valid     bool        `json:"valid"`
	Errors    []Highlight `json:"errors,omitempty"`
	Warnings  []string    `json:"warnings,omitempty"`
}
