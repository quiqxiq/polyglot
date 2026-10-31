package huaweiolt

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
		{"display is read", "display version", command.ModeRead, false},
		{"show is read", "show running-config", command.ModeRead, false},
		{"ping is read", "ping 10.0.0.1", command.ModeRead, false},
		{"traceroute is read", "traceroute 8.8.8.8", command.ModeRead, false},
		{"read prefix is case-insensitive and trimmed", "  DISPLAY version  ", command.ModeRead, false},
		{"reboot is mutation but not denied", "reboot", command.ModeMutation, false},
		{"reset saved-configuration is hard denied", "reset saved-configuration", command.ModeMutation, true},
		{"restore factory-configuration is hard denied", "restore factory-configuration", command.ModeMutation, true},
		{"system-view is mutation", "system-view", command.ModeMutation, false},
		{"bare display is mutation (fail-closed)", "display", command.ModeMutation, false},
		{"unknown command is mutation", "configure terminal", command.ModeMutation, false},
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
		command.Decide(Classify(command.Command{Raw: "display version"})))
	assert.Equal(t, command.DecisionRequireApproval,
		command.Decide(Classify(command.Command{Raw: "reboot"})))
	assert.Equal(t, command.DecisionDeny,
		command.Decide(Classify(command.Command{Raw: "reset saved-configuration"})))
}

func TestDriverClassifyDelegates(t *testing.T) {
	d := &Driver{}
	assert.Equal(t, command.ModeRead, d.Classify(command.Command{Raw: "show version"}).Mode)
	assert.True(t, d.Classify(command.Command{Raw: "reset saved-configuration"}).IsDenied())
}
