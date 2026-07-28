package genericcli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quixiq/polyglot/internal/domain/command"
)

func TestCatalog_Classify(t *testing.T) {
	catalog := Catalog{
		Curated:             true,
		DestructivePrefixes: []string{"reload", "write erase", "delete"},
	}

	tests := []struct {
		name string
		cmd  command.Command
		want command.Class
	}{
		{"reload is destructive", command.Command{Raw: "reload"}, command.ClassDestructive},
		{"reload with args is destructive", command.Command{Raw: "reload in 5"}, command.ClassDestructive},
		{"write erase is destructive", command.Command{Raw: "write erase"}, command.ClassDestructive},
		{"show command is read-only", command.Command{Raw: "show version"}, command.ClassReadOnly},
		{"unrelated command defaults read-only", command.Command{Raw: "ping 1.1.1.1"}, command.ClassReadOnly},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, catalog.Classify(tt.cmd))
		})
	}
}

func TestCatalog_Translate(t *testing.T) {
	catalog := Catalog{
		Operations: map[command.Operation]command.Command{
			command.OpGetStatus: {Raw: "show version"},
			command.OpReboot:    {Raw: "reload"},
		},
	}

	t.Run("known operation", func(t *testing.T) {
		got, err := catalog.Translate(command.OpGetStatus)
		require.NoError(t, err)
		assert.Equal(t, command.Command{Raw: "show version"}, got)
	})

	t.Run("unknown operation errors", func(t *testing.T) {
		_, err := catalog.Translate(command.Operation("does_not_exist"))
		require.Error(t, err)
	})
}

func TestCatalog_ZeroValueIsFailSafe(t *testing.T) {
	// A Catalog{} (Curated false — no prefixes, no operations) must not
	// panic and must fail SAFE: with the vendor's risk unknown, every
	// command is destructive (needs approval), never silently read-only.
	var catalog Catalog

	assert.Equal(t, command.ClassDestructive, catalog.Classify(command.Command{Raw: "anything"}))

	_, err := catalog.Translate(command.OpGetStatus)
	require.Error(t, err)
}

func TestCatalog_CuratedUnlistedIsReadOnly(t *testing.T) {
	// Once curated, a command not in the destructive list is read-only.
	catalog := Catalog{Curated: true}

	assert.Equal(t, command.ClassReadOnly, catalog.Classify(command.Command{Raw: "show version"}))
}
