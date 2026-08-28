package iface

import (
	"testing"
)

func TestNewPrintInterfacesCommand(t *testing.T) {
	cmd := NewPrintInterfacesCommand("ether", "ether1")
	if cmd.Raw != "/interface/print" || cmd.Args["?type"] != "ether" || cmd.Args["?name"] != "ether1" {
		t.Fatalf("unexpected print interfaces command: %+v", cmd)
	}
}

func TestNewMonitorTrafficOnceCommand(t *testing.T) {
	cmd := NewMonitorTrafficOnceCommand("ether1")
	if cmd.Raw != "/interface/monitor-traffic" || cmd.Args["interface"] != "ether1" {
		t.Fatalf("unexpected monitor traffic command: %+v", cmd)
	}
	if _, ok := cmd.Args["once"]; !ok {
		t.Fatalf("expected once arg in monitor traffic command")
	}
}

