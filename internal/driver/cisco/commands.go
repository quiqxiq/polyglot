package cisco

import (
	"fmt"
	"strings"

	"github.com/quixiq/polyglot/internal/domain/command"
)

// destructivePrefixes lists CLI command prefixes considered destructive —
// these always require HITL approval regardless of caller.
// TODO: fill in the full list per Polyglot-Architecture.md §5.3.
var destructivePrefixes = []string{"reload", "write erase", "erase startup-config"}

// operationMap translates abstract Operations to Cisco CLI commands.
var operationMap = map[command.Operation]command.Command{
	command.OpGetStatus: {Raw: "show version"},
	command.OpReboot:    {Raw: "reload"},
}

// Classify reports the risk class of cmd according to Cisco CLI conventions.
func Classify(cmd command.Command) command.Class {
	for _, prefix := range destructivePrefixes {
		if strings.HasPrefix(cmd.Raw, prefix) {
			return command.ClassDestructive
		}
	}
	return command.ClassReadOnly
}

// Translate maps an abstract Operation to a Cisco CLI Command.
func Translate(op command.Operation) (command.Command, error) {
	cmd, ok := operationMap[op]
	if !ok {
		return command.Command{}, fmt.Errorf("cisco: unsupported operation %q", op)
	}
	return cmd, nil
}
