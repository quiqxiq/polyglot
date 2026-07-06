package zteolt

import (
	"fmt"

	"github.com/quixiq/polyglot/internal/domain/command"
)

// destructiveCommands lists Telnet provisioning commands considered
// destructive for ZTE OLT. TODO: fill in per actual firmware CLI reference.
var destructiveCommands = map[string]bool{
	"reboot":         true,
	"reset":          true,
	"deregister-onu": true,
}

// operationMap translates abstract Operations to ZTE OLT Telnet commands.
// OpGetStatus is intentionally absent — see telnet.go's Translate comment.
var operationMap = map[command.Operation]command.Command{
	command.OpReboot: {Raw: "reboot"},
}

// Classify reports the risk class of cmd according to ZTE OLT conventions.
func Classify(cmd command.Command) command.Class {
	if destructiveCommands[cmd.Raw] {
		return command.ClassDestructive
	}
	return command.ClassReadOnly
}

// Translate maps an abstract Operation to a ZTE OLT Telnet Command.
func Translate(op command.Operation) (command.Command, error) {
	cmd, ok := operationMap[op]
	if !ok {
		return command.Command{}, fmt.Errorf("zteolt: unsupported operation %q", op)
	}
	return cmd, nil
}
