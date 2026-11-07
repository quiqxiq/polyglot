package main

// InspectChild represents a child node returned by /console/inspect request=child.
type InspectChild struct {
	Type     string `json:"type"`                // "child"
	Name     string `json:"name"`                // e.g. "secret", "profile", "add", "set"
	NodeType string `json:"node-type,omitempty"` // "dir", "path", "cmd", "arg"
}

// InspectSyntax represents syntax or argument info returned by /console/inspect request=syntax.
type InspectSyntax struct {
	Type        string   `json:"type"`                  // "syntax"
	Name        string   `json:"name,omitempty"`        // argument name
	SyntaxType  string   `json:"syntax-type,omitempty"` // e.g. "string", "ip-prefix", "bool"
	Text        string   `json:"text,omitempty"`        // help / description text
	Default     string   `json:"default,omitempty"`     // default value
	Required    bool     `json:"required,omitempty"`    // whether argument is required
	Values      []string `json:"values,omitempty"`      // enum values
	Help        string   `json:"help,omitempty"`
	Description string   `json:"description,omitempty"`
}

// SubmenuSchema holds all introspected commands and properties for a RouterOS submenu (e.g. /ppp/secret).
type SubmenuSchema struct {
	Name        string             // e.g. "secret"
	FullPath    string             // e.g. "ppp,secret"
	Description string             // Help text for submenu
	Commands    map[string]Command // e.g. "add", "set", "print", "remove"
}

// Command holds the argument definitions for a single command.
type Command struct {
	Name       string
	FullPath   string
	HasNumbers bool // true if command targets item numbers/selectors (e.g. in /set)
	Arguments  []Argument
}

// Argument represents a single parameter/field on a RouterOS command or entity.
type Argument struct {
	Name        string   // e.g. "local-address"
	Type        string   // e.g. "string", "ip", "bool", "num"
	Description string   // Help text
	Default     string   // Default value
	Required    bool     // Is required for add
	Values      []string // Allowed values (enum)
}

// GeneratedField represents a Go struct field ready for rendering.
type GeneratedField struct {
	Name     string // Go PascalCase field name, e.g. "LocalAddress"
	Type     string // Go type, e.g. "string", "bool", "*bool", "int64"
	RosName  string // Original RouterOS kebab-case name, e.g. "local-address"
	Tag      string // Go struct tag, e.g. `json:"local-address,omitempty" ros:"local-address"`
	Doc      string // Comment line(s)
	Required bool
}

// GeneratedStruct represents a Go struct ready for rendering.
type GeneratedStruct struct {
	Name    string // e.g. "PPPSecret", "PPPSecretAddRequest"
	Doc     string // Struct doc comment
	Command string // "print", "add", "set"
	Fields  []GeneratedField
}
