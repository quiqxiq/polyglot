package genericssh

import (
	"fmt"

	"github.com/quixiq/polyglot/internal/domain/command"
)

// Classify reports the AI Router Ops classification of cmd for generic
// SSH-CLI vendors. Uncurated vendors have no read allowlist, so every
// command is MUTATION (fail-closed, always requires approval) — consistent
// with Polyglot-Architecture.md §5.4.
func Classify(cmd command.Command) command.Classification {
	return command.Classification{Mode: command.ModeMutation}
}

// Translate always errors: generic SSH vendors have no curated operation
// catalog until a specific platformdef is added for them.
func Translate(op command.Operation) (command.Command, error) {
	return command.Command{}, fmt.Errorf("genericssh: no curated operations until a platformdef is added")
}
