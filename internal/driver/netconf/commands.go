package netconf

import (
	"fmt"
	"strings"

	"github.com/quixiq/polyglot/internal/domain/command"
)

// Command classification for AI Router Ops (PLAN-AI-ROUTER-OPS v2).
//
// NETCONF operations are classified by operation name: get/get-config/
// get-schema are READ, everything else is MUTATION (fail-closed for unknown
// operations). edit-config/commit therefore require approval, while
// delete-config is hard-denied: it wipes a whole configuration datastore and
// is never safe for an AI proposal. Human restore flows use dedicated
// usecases with typed confirmation instead.

// readOperations is the allowlist of NETCONF operations that only read state.
var readOperations = map[string]bool{
	"get":        true,
	"get-config": true,
	"get-schema": true,
}

// deniedOperations maps hard-denied NETCONF operations to their reason.
var deniedOperations = map[string]string{
	"delete-config": "hard deny for AI: deletes a configuration datastore",
}

// operationMap translates abstract Operations to NETCONF operations.
// OpReboot is intentionally absent: NETCONF has no standard reboot RPC
// across vendors. TODO: add a vendor-specific extension if/when needed.
var operationMap = map[command.Operation]command.Command{
	command.OpGetStatus: {Raw: "get-config"},
}

// Classify reports the AI Router Ops classification of cmd according to
// NETCONF conventions.
func Classify(cmd command.Command) command.Classification {
	op := strings.ToLower(strings.TrimSpace(cmd.Raw))
	if op == "" {
		return command.Classification{Mode: command.ModeMutation}
	}
	// Operation name is the first whitespace-delimited token; NETCONF
	// arguments (e.g. "get-config source=running") do not change the risk.
	name, _, _ := strings.Cut(op, " ")
	if reason, denied := deniedOperations[name]; denied {
		return command.Classification{Mode: command.ModeMutation, DenyReason: reason}
	}
	if readOperations[name] {
		return command.Classification{Mode: command.ModeRead}
	}
	return command.Classification{Mode: command.ModeMutation}
}

// Translate maps an abstract Operation to a NETCONF Command.
func Translate(op command.Operation) (command.Command, error) {
	cmd, ok := operationMap[op]
	if !ok {
		return command.Command{}, fmt.Errorf("netconf: unsupported operation %q", op)
	}
	return cmd, nil
}
