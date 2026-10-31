package cisco

import (
	"fmt"
	"strings"

	"github.com/quixiq/polyglot/internal/domain/command"
)

// readPrefixes is the allowlist of Cisco CLI command prefixes that only
// read state or introspect syntax. Prefixes end with a space (literal
// prefix); "dir" has none and matches the word "dir" (with or without a
// following argument). Everything not listed here is MUTATION.
var readPrefixes = []string{
	"show ",
	"display ",
	"more ",
	"dir",
	"ping ",
	"traceroute ",
	"terminal length ",
}

// denyPrefixes are CLI commands the AI pipeline must never execute, even
// with approval: reload and configuration wipes. Checked before the read
// allowlist.
var denyPrefixes = []string{"reload", "write erase", "erase startup-config"}

// operationMap translates abstract Operations to Cisco CLI commands.
var operationMap = map[command.Operation]command.Command{
	command.OpGetStatus: {Raw: "show version"},
	command.OpReboot:    {Raw: "reload"},
}

// Classify reports the AI Router Ops classification of cmd according to
// Cisco CLI conventions. Hard-deny commands are checked first, then the
// read allowlist; anything else is MUTATION (fail-closed).
func Classify(cmd command.Command) command.Classification {
	raw := strings.ToLower(strings.TrimSpace(cmd.Raw))
	if raw == "" {
		return command.Classification{Mode: command.ModeMutation, DenyReason: "empty command"}
	}
	for _, prefix := range denyPrefixes {
		if raw == prefix || strings.HasPrefix(raw, prefix+" ") {
			return command.Classification{Mode: command.ModeMutation, DenyReason: "hard deny for AI: " + prefix}
		}
	}
	for _, prefix := range readPrefixes {
		if hasCmdPrefix(raw, prefix) {
			return command.Classification{Mode: command.ModeRead}
		}
	}
	return command.Classification{Mode: command.ModeMutation}
}

// hasCmdPrefix reports whether raw (already lowercased) begins with prefix.
// A trailing space means a literal prefix ("show "); otherwise the prefix is
// treated as a whole command word ("dir" matches "dir" and "dir flash:",
// but not "directory").
func hasCmdPrefix(raw, prefix string) bool {
	if strings.HasSuffix(prefix, " ") {
		return strings.HasPrefix(raw, prefix)
	}
	return raw == prefix || strings.HasPrefix(raw, prefix+" ")
}

// Translate maps an abstract Operation to a Cisco CLI Command.
func Translate(op command.Operation) (command.Command, error) {
	cmd, ok := operationMap[op]
	if !ok {
		return command.Command{}, fmt.Errorf("cisco: unsupported operation %q", op)
	}
	return cmd, nil
}
