package inspect

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/quixiq/polyglot/internal/domain/command"
)

// ParseChild maps /console/inspect request=child rows to typed nodes.
func ParseChild(result command.Result) []Node {
	out := make([]Node, 0, len(result.Rows))
	for _, row := range result.Rows {
		out = append(out, Node{
			Type:     row["type"],
			Name:     row["name"],
			NodeType: row["node-type"],
		})
	}
	return out
}

// ParseSyntax maps request=syntax rows to typed nodes.
func ParseSyntax(result command.Result) []Syntax {
	out := make([]Syntax, 0, len(result.Rows))
	for _, row := range result.Rows {
		out = append(out, Syntax{
			Type:       row["type"],
			Symbol:     row["symbol"],
			SymbolType: row["symbol-type"],
			Nested:     row["nested"],
			Nonorm:     row["nonorm"],
			Text:       row["text"],
		})
	}
	return out
}

// ParseCompletion maps request=completion rows to typed completions.
func ParseCompletion(result command.Result) []Completion {
	out := make([]Completion, 0, len(result.Rows))
	for _, row := range result.Rows {
		offset, _ := strconv.Atoi(strings.TrimSpace(row["offset"]))
		pref, _ := strconv.Atoi(strings.TrimSpace(row["preference"]))
		out = append(out, Completion{
			Type:       row["type"],
			Completion: row["completion"],
			Style:      row["style"],
			Offset:     offset,
			Preference: pref,
			Show:       row["show"] == "true",
			Text:       row["text"],
		})
	}
	return out
}

// ParseHighlight maps request=highlight rows to typed highlights.
func ParseHighlight(result command.Result) []Highlight {
	out := make([]Highlight, 0, len(result.Rows))
	for _, row := range result.Rows {
		start, _ := strconv.Atoi(strings.TrimSpace(row["start"]))
		end, _ := strconv.Atoi(strings.TrimSpace(row["end"]))
		out = append(out, Highlight{
			Type:  row["type"],
			Text:  row["text"],
			Start: start,
			End:   end,
		})
	}
	return out
}

// BuildValidation turns highlight output into a ValidationResult. A token
// with type "error" (RouterOS uses "error" and sometimes "syntax-error")
// makes the script invalid; offsets are returned for the UI.
func BuildValidation(highlights []Highlight) ValidationResult {
	res := ValidationResult{Supported: true, Valid: true}
	for _, h := range highlights {
		typeLower := strings.ToLower(h.Type)
		if typeLower == "error" || strings.Contains(typeLower, "error") {
			res.Valid = false
			res.Errors = append(res.Errors, h)
		}
	}
	return res
}

// ParseCapabilities extracts RouterOS version from /system/resource/print.
func ParseCapabilities(result command.Result) (version string, major, minor, patch int, err error) {
	if len(result.Rows) == 0 {
		return "", 0, 0, 0, fmt.Errorf("inspect: empty /system/resource/print response")
	}
	version = strings.TrimSpace(result.Rows[0]["version"])
	major, minor, patch = ParseVersion(version)
	return version, major, minor, patch, nil
}

// ParseVersion parses a RouterOS version string like "7.16.2 (stable)".
func ParseVersion(raw string) (major, minor, patch int) {
	fields := strings.Fields(strings.TrimSpace(raw))
	if len(fields) == 0 {
		return 0, 0, 0
	}
	parts := strings.Split(fields[0], ".")
	if len(parts) > 0 {
		major, _ = strconv.Atoi(parts[0])
	}
	if len(parts) > 1 {
		minor, _ = strconv.Atoi(parts[1])
	}
	if len(parts) > 2 {
		patch, _ = strconv.Atoi(parts[2])
	}
	return major, minor, patch
}
