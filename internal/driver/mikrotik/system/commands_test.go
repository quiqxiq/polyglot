package system

import (
	"testing"
)

func TestSystemCommands(t *testing.T) {
	cmdRes := NewPrintResourceCommand()
	if cmdRes.Raw != "/system/resource/print" {
		t.Fatalf("expected /system/resource/print, got %s", cmdRes.Raw)
	}

	cmdHealth := NewPrintHealthCommand()
	if cmdHealth.Raw != "/system/health/print" {
		t.Fatalf("expected /system/health/print, got %s", cmdHealth.Raw)
	}

	cmdIdent := NewSetIdentityCommand("Router-Utama")
	if cmdIdent.Raw != "/system/identity/set" || cmdIdent.Args["name"] != "Router-Utama" {
		t.Fatalf("unexpected set identity command: %+v", cmdIdent)
	}

	cmdPing := NewPingCommand("8.8.8.8", "5")
	if cmdPing.Raw != "/ping" || cmdPing.Args["address"] != "8.8.8.8" || cmdPing.Args["count"] != "5" {
		t.Fatalf("unexpected ping command: %+v", cmdPing)
	}
}

