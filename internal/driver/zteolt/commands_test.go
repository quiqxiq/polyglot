package zteolt

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/quixiq/polyglot/internal/domain/command"
)

func TestClassify(t *testing.T) {
	tests := []struct {
		name       string
		raw        string
		wantMode   command.Mode
		wantDenied bool
	}{
		{"show is read", "show gpon onu detail-info gpon-olt_1/1/1", command.ModeRead, false},
		{"display is read", "display version", command.ModeRead, false},
		{"ping is read", "ping 10.0.0.1", command.ModeRead, false},
		{"traceroute is read", "traceroute 8.8.8.8", command.ModeRead, false},
		{"read prefix is case-insensitive and trimmed", "  SHOW version  ", command.ModeRead, false},
		{"reboot is mutation but not denied", "reboot", command.ModeMutation, false},
		{"deregister onu is mutation but not denied", "deregister-onu 1/1/1", command.ModeMutation, false},
		{"reset is hard denied", "reset", command.ModeMutation, true},
		{"restore is hard denied", "restore factory-default", command.ModeMutation, true},
		{"config terminal is mutation", "config terminal", command.ModeMutation, false},
		{"bare show is mutation (fail-closed)", "show", command.ModeMutation, false},
		{"unknown command is mutation", "hello world", command.ModeMutation, false},
		{"empty command is mutation", "", command.ModeMutation, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Classify(command.Command{Raw: tt.raw})
			assert.Equal(t, tt.wantMode, got.Mode)
			assert.Equal(t, tt.wantDenied, got.IsDenied())
		})
	}
}

func TestClassifyPolicy(t *testing.T) {
	assert.Equal(t, command.DecisionAutoApprove,
		command.Decide(Classify(command.Command{Raw: "show version"})))
	assert.Equal(t, command.DecisionRequireApproval,
		command.Decide(Classify(command.Command{Raw: "reboot"})))
	assert.Equal(t, command.DecisionDeny,
		command.Decide(Classify(command.Command{Raw: "reset"})))
}

func TestDriverClassifyDelegates(t *testing.T) {
	d := &Driver{}
	assert.Equal(t, command.ModeRead, d.Classify(command.Command{Raw: "show version"}).Mode)
	assert.True(t, d.Classify(command.Command{Raw: "restore factory-default"}).IsDenied())
}
