package cisco

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/quixiq/polyglot/internal/domain/command"
)

func TestClassify(t *testing.T) {
	tests := []struct {
		name string
		cmd  command.Command
		want command.Classification
	}{
		{"empty command is denied", command.Command{Raw: ""},
			command.Classification{Mode: command.ModeMutation, DenyReason: "empty command"}},
		{"reload is hard-denied", command.Command{Raw: "reload"},
			command.Classification{Mode: command.ModeMutation, DenyReason: "hard deny for AI: reload"}},
		{"reload with args is hard-denied", command.Command{Raw: "reload in 5"},
			command.Classification{Mode: command.ModeMutation, DenyReason: "hard deny for AI: reload"}},
		{"write erase is hard-denied", command.Command{Raw: "write erase"},
			command.Classification{Mode: command.ModeMutation, DenyReason: "hard deny for AI: write erase"}},
		{"erase startup-config is hard-denied", command.Command{Raw: "erase startup-config"},
			command.Classification{Mode: command.ModeMutation, DenyReason: "hard deny for AI: erase startup-config"}},
		{"show is read", command.Command{Raw: "show ip interface brief"},
			command.Classification{Mode: command.ModeRead}},
		{"show is read regardless of case", command.Command{Raw: "Show Version"},
			command.Classification{Mode: command.ModeRead}},
		{"display is read", command.Command{Raw: "display current-configuration"},
			command.Classification{Mode: command.ModeRead}},
		{"more is read", command.Command{Raw: "more flash:config.txt"},
			command.Classification{Mode: command.ModeRead}},
		{"bare dir is read", command.Command{Raw: "dir"},
			command.Classification{Mode: command.ModeRead}},
		{"dir with argument is read", command.Command{Raw: "dir flash:"},
			command.Classification{Mode: command.ModeRead}},
		{"ping is read", command.Command{Raw: "ping 1.1.1.1"},
			command.Classification{Mode: command.ModeRead}},
		{"traceroute is read", command.Command{Raw: "traceroute 8.8.8.8"},
			command.Classification{Mode: command.ModeRead}},
		{"terminal length is read", command.Command{Raw: "terminal length 0"},
			command.Classification{Mode: command.ModeRead}},
		{"configure terminal is mutation", command.Command{Raw: "configure terminal"},
			command.Classification{Mode: command.ModeMutation}},
		{"unknown command defaults mutation", command.Command{Raw: "no_such_command"},
			command.Classification{Mode: command.ModeMutation}},
		{"directory does not match dir word", command.Command{Raw: "directory"},
			command.Classification{Mode: command.ModeMutation}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, Classify(tt.cmd))
		})
	}
}
