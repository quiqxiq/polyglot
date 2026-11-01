package mikrotik

import (
	"strings"

	"github.com/quixiq/polyglot/internal/domain/command"
)

// RouterOS command classification.
//
// Commands are categorized into READ or MUTATION (fail-closed for unknown verbs).
// A hard-deny list refuses commands that are never safe to execute directly:
// reset-configuration, full backup restore, arbitrary file removal, and
// arbitrary /tool/fetch (SSRF). Human recovery flows use dedicated usecases
// with typed confirmation instead.

// readVerbs is the allowlist of RouterOS verbs that only read state or
// introspect syntax. Everything else is MUTATION.
var readVerbs = map[string]bool{
	"print":           true,
	"get":             true,
	"export":          true,
	"monitor":         true,
	"monitor-traffic": true,
	"ping":            true,
	"traceroute":      true,
	"check":           true,
	"inspect":         true,
	"sniff":           true,
	"torch":           true,
	"find":            true,
	"discover":        true,
	"scan":            true,
	"whois":           true,
}

// hardDenyPrefixes are API paths that must never be executed directly. Checked against the
// normalized raw path (queries/args excluded).
var hardDenyPrefixes = []string{
	"/system/reset-configuration",
	"/system/backup/load",
	"/system/script/add",
	"/system/script/set",
	"/system/scheduler/add",
	"/system/scheduler/set",
	"/file/remove",
	"/tool/fetch",
}

// Classify reports the risk classification of cmd.
func Classify(cmd command.Command) command.Classification {
	raw := strings.TrimSpace(cmd.Raw)
	if raw == "" {
		return command.Classification{Mode: command.ModeMutation, DenyReason: "empty command path"}
	}
	// Normalize for deny-list checks: strip trailing whitespace and any
	// "/print" style suffix is NOT stripped — deny rules target exact paths.
	lower := strings.ToLower(raw)
	for _, prefix := range hardDenyPrefixes {
		if lower == prefix || strings.HasPrefix(lower, prefix+" ") || strings.HasPrefix(lower, prefix+"/") {
			return command.Classification{Mode: command.ModeMutation, DenyReason: "hard deny: " + prefix}
		}
	}

	verb := lastPathSegment(raw)
	if verb == "" {
		return command.Classification{Mode: command.ModeMutation}
	}

	// A read verb with write-ish attributes is still a mutation (defense in
	// depth against e.g. /file/print with a crafted arg).
	if readVerbs[verb] && !hasWriteArgs(cmd.Args) {
		return command.Classification{Mode: command.ModeRead}
	}
	return command.Classification{Mode: command.ModeMutation}
}

// lastPathSegment returns the final non-empty segment of a RouterOS API path.
// It accepts leading "/" and tolerates a trailing "/".
func lastPathSegment(raw string) string {
	trimmed := strings.Trim(raw, "/")
	if trimmed == "" {
		return ""
	}
	if i := strings.LastIndex(trimmed, "/"); i >= 0 {
		return trimmed[i+1:]
	}
	return trimmed
}

// hasWriteArgs reports whether the argument map contains a key that can never
// appear on a genuine read command. /export file=... is deliberately NOT
// treated as a write: exporting to a file only reads configuration.
func hasWriteArgs(args map[string]string) bool {
	_, ok := args["contents"]
	return ok
}
