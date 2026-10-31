package netconf

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
		{"get is read", "get", command.ModeRead, false},
		{"get-config is read", "get-config", command.ModeRead, false},
		{"get-schema is read", "get-schema", command.ModeRead, false},
		{"read operation is case-insensitive and trimmed", "  GET-CONFIG  ", command.ModeRead, false},
		{"read operation with arguments is read", "get-config source=running", command.ModeRead, false},
		{"edit-config is mutation but not denied", "edit-config", command.ModeMutation, false},
		{"commit is mutation but not denied", "commit", command.ModeMutation, false},
		{"copy-config is mutation", "copy-config", command.ModeMutation, false},
		{"delete-config is hard denied", "delete-config", command.ModeMutation, true},
		{"unknown rpc is mutation", "custom-rpc", command.ModeMutation, false},
		{"empty operation is mutation", "", command.ModeMutation, false},
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
		command.Decide(Classify(command.Command{Raw: "get-config"})))
	assert.Equal(t, command.DecisionRequireApproval,
		command.Decide(Classify(command.Command{Raw: "edit-config"})))
	assert.Equal(t, command.DecisionDeny,
		command.Decide(Classify(command.Command{Raw: "delete-config"})))
}

func TestDriverClassifyDelegates(t *testing.T) {
	d := &Driver{}
	assert.Equal(t, command.ModeRead, d.Classify(command.Command{Raw: "get"}).Mode)
	assert.True(t, d.Classify(command.Command{Raw: "delete-config"}).IsDenied())
}
