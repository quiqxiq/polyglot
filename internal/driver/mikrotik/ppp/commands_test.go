package ppp

import (
	"testing"
)

func TestNewAddSecretCommand(t *testing.T) {
	cmd := NewAddSecretCommand(PPPoESecretParams{
		Name:     "user1",
		Password: "pass",
		Profile:  "default",
		Service:  "pppoe",
		Disabled: false,
	})
	if cmd.Raw != "/ppp/secret/add" || cmd.Args["name"] != "user1" {
		t.Fatalf("unexpected add secret command: %+v", cmd)
	}
}

func TestNewPrintActiveCommand(t *testing.T) {
	cmd := NewPrintActiveCommand("user1")
	if cmd.Raw != "/ppp/active/print" || cmd.Args["?name"] != "user1" {
		t.Fatalf("unexpected print active command: %+v", cmd)
	}
}

func TestNewPrintProfilesCommand(t *testing.T) {
	cmd := NewPrintProfilesCommand("10mbps")
	if cmd.Raw != "/ppp/profile/print" || cmd.Args["?name"] != "10mbps" {
		t.Fatalf("unexpected print profile command: %+v", cmd)
	}
}

// F2-9: binding MAC pelanggan (caller-id) harus diteruskan ke /ppp/secret.
func TestNewAddSecretCommand_WritesCallerID(t *testing.T) {
	cmd := NewAddSecretCommand(PPPoESecretParams{Name: "u1", CallerID: "AA:BB:CC:DD:EE:FF"})
	if cmd.Args["caller-id"] != "AA:BB:CC:DD:EE:FF" {
		t.Fatalf("caller-id tidak ditulis: %+v", cmd.Args)
	}
}

func TestNewSetSecretCommand_WritesCallerID(t *testing.T) {
	cmd := NewSetSecretCommand("*1", PPPoESecretParams{CallerID: "AA:BB:CC:DD:EE:FF"})
	if cmd.Args["caller-id"] != "AA:BB:CC:DD:EE:FF" {
		t.Fatalf("caller-id tidak ditulis: %+v", cmd.Args)
	}
}
