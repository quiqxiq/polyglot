package guard

import (
	"errors"
)

// ErrGuardRefused is returned when a safety guard refuses a command to prevent router lockout.
var ErrGuardRefused = errors.New("command refused by routeros safety guard")

// Level describes the severity of a guard verdict: "none", "warn", or "refuse".
type Level string

const (
	LevelNone   Level = "none"
	LevelWarn   Level = "warn"
	LevelRefuse Level = "refuse"
)

// Verdict is the evaluation result returned by a guard check.
type Verdict struct {
	Level       string         `json:"level"`
	Code        string         `json:"code,omitempty"`
	Detail      map[string]any `json:"detail,omitempty"`
	Fingerprint string         `json:"fingerprint,omitempty"`
}

// Warned reports whether this verdict generates a safety warning.
func (v Verdict) Warned() bool { return v.Level == "warn" }

// Refused reports whether this verdict outright refuses the operation.
func (v Verdict) Refused() bool { return v.Level == "refuse" }

// Reply represents a RouterOS sentence attribute map.
type Reply = map[string]string
