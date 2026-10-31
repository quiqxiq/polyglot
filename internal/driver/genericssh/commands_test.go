package genericssh

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/quixiq/polyglot/internal/domain/command"
)

func TestClassify(t *testing.T) {
	// Uncurated SSH vendors have no read allowlist: every command is
	// MUTATION (fail-closed), including unknown and read-looking commands.
	tests := []struct {
		name string
		cmd  command.Command
	}{
		{"empty command", command.Command{}},
		{"read-looking command", command.Command{Raw: "show version"}},
		{"destructive command", command.Command{Raw: "reload"}},
		{"unknown command", command.Command{Raw: "no_such_command"}},
	}
	want := command.Classification{Mode: command.ModeMutation}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, want, Classify(tt.cmd))
		})
	}
}
