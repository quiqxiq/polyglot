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
