package huaweiolt

import (
	"fmt"

	"github.com/quixiq/polyglot/internal/domain/command"
)

// destructiveCommands lists Huawei VRP CLI commands considered destructive.
// TODO: fill in the full list per Polyglot-Architecture.md §5.3.
var destructiveCommands = map[string]bool{
	"reboot":                    true,
	"reset saved-configuration": true,
}

// operationMap translates abstract Operations to Huawei VRP CLI commands.
var operationMap = map[command.Operation]command.Command{
	command.OpGetStatus: {Raw: "display version"},
	command.OpReboot:    {Raw: "reboot"},
}

// Classify reports the risk class of cmd according to Huawei VRP conventions.
func Classify(cmd command.Command) command.Class {
	if destructiveCommands[cmd.Raw] {
		return command.ClassDestructive
	}
	return command.ClassReadOnly
}

// Translate maps an abstract Operation to a Huawei VRP Command.
func Translate(op command.Operation) (command.Command, error) {
	cmd, ok := operationMap[op]
	if !ok {
		return command.Command{}, fmt.Errorf("huaweiolt: unsupported operation %q", op)
	}
	return cmd, nil
}
