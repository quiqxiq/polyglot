package huaweiolt

import (
	"fmt"
	"strings"

	"github.com/quixiq/polyglot/internal/domain/command"
)

// Command classification for AI Router Ops (PLAN-AI-ROUTER-OPS v2).
//
// Only READ commands may be auto-executed by the AI pipeline; everything
// else is MUTATION and needs approval. Unknown commands are MUTATION
// (fail-closed). A small hard-deny list refuses full reset/restore commands
// even with approval — human recovery flows use dedicated usecases with
// typed confirmation instead.

// readPrefixes are Huawei VRP CLI prefixes that only read/introspect state.
var readPrefixes = []string{"display ", "show ", "ping ", "traceroute "}

// hardDenyPrefixes are full reset/restore commands that AI must never
// propose. Reboot is deliberately NOT hard-denied — it is an ordinary
// mutation that needs approval.
var hardDenyPrefixes = []string{"reset", "restore", "factory-reset", "factory reset"}

// operationMap translates abstract Operations to Huawei VRP CLI commands.
var operationMap = map[command.Operation]command.Command{
	command.OpGetStatus: {Raw: "display version"},
	command.OpReboot:    {Raw: "reboot"},
}

// Classify reports the AI Router Ops classification of cmd according to
// Huawei VRP conventions.
func Classify(cmd command.Command) command.Classification {
	raw := strings.ToLower(strings.TrimSpace(cmd.Raw))
	if raw == "" {
		return command.Classification{Mode: command.ModeMutation}
	}
	for _, prefix := range hardDenyPrefixes {
		if raw == prefix || strings.HasPrefix(raw, prefix+" ") {
			return command.Classification{
				Mode:       command.ModeMutation,
				DenyReason: "hard deny for AI: " + prefix,
			}
		}
	}
	for _, prefix := range readPrefixes {
		if strings.HasPrefix(raw, prefix) {
			return command.Classification{Mode: command.ModeRead}
		}
	}
	return command.Classification{Mode: command.ModeMutation}
}

// Translate maps an abstract Operation to a Huawei VRP Command.
func Translate(op command.Operation) (command.Command, error) {
	cmd, ok := operationMap[op]
	if !ok {
		return command.Command{}, fmt.Errorf("huaweiolt: unsupported operation %q", op)
	}
	return cmd, nil
}
