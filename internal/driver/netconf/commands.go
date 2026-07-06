package netconf

import (
	"fmt"

	"github.com/quixiq/polyglot/internal/domain/command"
)

// destructiveOperations lists NETCONF operations considered destructive —
// edit-config/delete-config/commit always require HITL approval.
// get-config/get are read-only.
var destructiveOperations = map[string]bool{
	"edit-config":   true,
	"delete-config": true,
	"commit":        true,
}

// operationMap translates abstract Operations to NETCONF operations.
// OpReboot is intentionally absent: NETCONF has no standard reboot RPC
// across vendors. TODO: add a vendor-specific extension if/when needed.
var operationMap = map[command.Operation]command.Command{
	command.OpGetStatus: {Raw: "get-config"},
}

// Classify reports the risk class of cmd according to NETCONF conventions.
func Classify(cmd command.Command) command.Class {
	if destructiveOperations[cmd.Raw] {
		return command.ClassDestructive
	}
	return command.ClassReadOnly
}

// Translate maps an abstract Operation to a NETCONF Command.
func Translate(op command.Operation) (command.Command, error) {
	cmd, ok := operationMap[op]
	if !ok {
		return command.Command{}, fmt.Errorf("netconf: unsupported operation %q", op)
	}
	return cmd, nil
}
