package genieacs

import (
	"fmt"

	"github.com/quixiq/polyglot/internal/domain/command"
)

// destructiveTaskTypes lists GenieACS task types considered destructive —
// e.g. factory reset or reboot of a customer's CPE/ONT.
var destructiveTaskTypes = map[string]bool{
	"reboot":       true,
	"factoryReset": true,
}

// operationMap translates abstract Operations to GenieACS task types.
var operationMap = map[command.Operation]command.Command{
	command.OpGetStatus: {Raw: "getParameterValues"},
	command.OpReboot:    {Raw: "reboot"},
}

// Classify reports the risk class of cmd according to GenieACS task types.
func Classify(cmd command.Command) command.Class {
	if destructiveTaskTypes[cmd.Raw] {
		return command.ClassDestructive
	}
	return command.ClassReadOnly
}

// Translate maps an abstract Operation to a GenieACS task type Command.
func Translate(op command.Operation) (command.Command, error) {
	cmd, ok := operationMap[op]
	if !ok {
		return command.Command{}, fmt.Errorf("genieacs: unsupported operation %q", op)
	}
	return cmd, nil
}
