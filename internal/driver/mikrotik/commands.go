package mikrotik

import (
	"fmt"

	"github.com/quixiq/polyglot/internal/domain/command"
)

// destructivePaths lists RouterOS API paths considered destructive — these
// always require HITL approval regardless of caller.
// TODO: fill in the full list per NetOps-Architecture.md §5.3.
var destructivePaths = map[string]bool{
	"/system/reboot":              true,
	"/system/reset-configuration": true,
}

// operationMap translates abstract Operations to RouterOS API paths.
// TODO: add more operations as new usecases need them.
var operationMap = map[command.Operation]command.Command{
	command.OpGetStatus: {Raw: "/system/resource/print"},
	command.OpReboot:    {Raw: "/system/reboot"},
}

// Classify reports the risk class of cmd according to RouterOS API conventions.
func Classify(cmd command.Command) command.Class {
	if destructivePaths[cmd.Raw] {
		return command.ClassDestructive
	}
	return command.ClassReadOnly
}

// Translate maps an abstract Operation to a RouterOS-native Command.
func Translate(op command.Operation) (command.Command, error) {
	cmd, ok := operationMap[op]
	if !ok {
		return command.Command{}, fmt.Errorf("mikrotik: unsupported operation %q", op)
	}
	return cmd, nil
}
