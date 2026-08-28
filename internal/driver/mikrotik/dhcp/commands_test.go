package dhcp

import (
	"testing"
)

func TestNewPrintLeasesCommand(t *testing.T) {
	cmd := NewPrintLeasesCommand("")
	if cmd.Raw != "/ip/dhcp-server/lease/print" {
		t.Fatalf("expected raw /ip/dhcp-server/lease/print, got %s", cmd.Raw)
	}
	if len(cmd.Args) != 0 {
		t.Fatalf("expected 0 args, got %v", cmd.Args)
	}

	cmdMac := NewPrintLeasesCommand("00:11:22:33:44:55")
	if cmdMac.Args["?mac-address"] != "00:11:22:33:44:55" {
		t.Fatalf("expected mac-address filter, got %v", cmdMac.Args)
	}
}

func TestNewStreamLeasesCommand(t *testing.T) {
	cmd := NewStreamLeasesCommand("")
	if _, ok := cmd.Args["follow"]; !ok {
		t.Fatalf("expected follow arg in streaming command, got %v", cmd.Args)
	}
}

func TestNewSetLeaseBlockCommand(t *testing.T) {
	cmd := NewSetLeaseBlockCommand("*1", DHCPLeaseBlockParams{
		Blocked: true,
		Comment: "SUSPENDED",
	})
	if cmd.Raw != "/ip/dhcp-server/lease/set" {
		t.Fatalf("expected /ip/dhcp-server/lease/set, got %s", cmd.Raw)
	}
	if cmd.Args[".id"] != "*1" || cmd.Args["blocked"] != "yes" || cmd.Args["comment"] != "SUSPENDED" {
		t.Fatalf("unexpected args: %v", cmd.Args)
	}
}

func TestNewPrintServersCommand(t *testing.T) {
	cmd := NewPrintServersCommand()
	if cmd.Raw != "/ip/dhcp-server/print" {
		t.Fatalf("expected /ip/dhcp-server/print, got %s", cmd.Raw)
	}
}

