package genericcli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quixiq/polyglot/internal/domain/command"
)

func TestCatalog_Classify(t *testing.T) {
	catalog := Catalog{
		DestructivePrefixes: []string{"reload", "write erase", "delete"},
		ReadPrefixes:        []string{"show ", "ping ", "dir"},
	}

	tests := []struct {
		name string
		cmd  command.Command
		want command.Classification
	}{
		{"reload is hard-denied", command.Command{Raw: "reload"},
			command.Classification{Mode: command.ModeMutation, DenyReason: "hard deny for AI: reload"}},
		{"reload with args is hard-denied", command.Command{Raw: "reload in 5"},
			command.Classification{Mode: command.ModeMutation, DenyReason: "hard deny for AI: reload"}},
		{"write erase is hard-denied", command.Command{Raw: "write erase"},
			command.Classification{Mode: command.ModeMutation, DenyReason: "hard deny for AI: write erase"}},
		{"erase startup-config denied even without catalog entry", command.Command{Raw: "erase startup-config"},
			command.Classification{Mode: command.ModeMutation, DenyReason: "hard deny for AI: erase startup-config"}},
		{"vendor-phrased reset is hard-denied", command.Command{Raw: "system reset-configuration"},
			command.Classification{Mode: command.ModeMutation, DenyReason: "hard deny for AI: reset-configuration"}},
		{"delete is mutation", command.Command{Raw: "delete vlan 10"},
			command.Classification{Mode: command.ModeMutation}},
		{"show command is read", command.Command{Raw: "show version"},
			command.Classification{Mode: command.ModeRead}},
		{"show startup-config is read, not denied", command.Command{Raw: "show startup-config"},
			command.Classification{Mode: command.ModeRead}},
		{"ping is read", command.Command{Raw: "ping 1.1.1.1"},
			command.Classification{Mode: command.ModeRead}},
		{"bare dir is read", command.Command{Raw: "dir"},
			command.Classification{Mode: command.ModeRead}},
		{"dir with argument is read", command.Command{Raw: "dir flash:"},
			command.Classification{Mode: command.ModeRead}},
		{"directory does not match dir word", command.Command{Raw: "directory"},
			command.Classification{Mode: command.ModeMutation}},
		{"unrelated command defaults mutation", command.Command{Raw: "clear counters"},
			command.Classification{Mode: command.ModeMutation}},
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

func TestCatalog_ZeroValueIsUsable(t *testing.T) {
	// A Catalog{} (no prefixes, no operations, no failure patterns) must
	// not panic — and must classify everything as MUTATION (fail-closed),
	// never as READ.
	var catalog Catalog

	assert.Equal(t, command.Classification{Mode: command.ModeMutation}, catalog.Classify(command.Command{Raw: "anything"}))

	_, err := catalog.Translate(command.OpGetStatus)
	require.Error(t, err)
}
