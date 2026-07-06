package genericssh

import (
	"fmt"

	"github.com/quixiq/polyglot/internal/domain/command"
)

// Classify reports the risk class of cmd for generic SSH-CLI vendors.
// TODO: populate once a custom platformdef is added for a specific vendor —
// until then, default to requiring approval for anything, since risk is
// genuinely unknown for an uncurated vendor (fail-safe, not fail-open).
func Classify(cmd command.Command) command.Class {
	return command.ClassDestructive
}

// Translate always errors: generic SSH vendors have no curated operation
// catalog until a specific platformdef is added for them.
func Translate(op command.Operation) (command.Command, error) {
	return command.Command{}, fmt.Errorf("genericssh: no curated operations until a platformdef is added")
}
