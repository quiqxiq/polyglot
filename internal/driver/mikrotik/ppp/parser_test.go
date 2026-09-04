package ppp

import (
	"testing"

	"github.com/quixiq/polyglot/internal/domain/command"
)

func TestParseSecrets(t *testing.T) {
	res := command.Result{
		Rows: []map[string]string{
			{
				".id":      "*1",
				"name":     "user1",
				"password": "secretpassword",
				"profile":  "10mbps",
				"service":  "pppoe",
				"disabled": "false",
			},
			{
				".id":             "*2",
				"name":            "user2",
				"profile":         "20mbps",
				"service":         "pppoe",
				"last-caller-id":  "AA:BB:CC:DD:EE:FF",
				"last-logged-out": "mar/04/2026 18:23:45",
			},
		},
	}

	secrets := ParseSecrets(res)
	if len(secrets) != 2 {
		t.Fatalf("expected 2 secrets, got %d", len(secrets))
	}
	if secrets[0].Name != "user1" || secrets[0].Profile != "10mbps" || secrets[0].Disabled {
		t.Fatalf("unexpected secret parsed: %+v", secrets[0])
	}
	if secrets[1].CallerID != "AA:BB:CC:DD:EE:FF" || secrets[1].LastLoggedOut != "mar/04/2026 18:23:45" {
		t.Fatalf("expected last-caller-id and last-logged-out to be parsed: %+v", secrets[1])
	}
}

func TestParseActiveSessions(t *testing.T) {
	res := command.Result{
		Rows: []map[string]string{
			{
				".id":        "*A",
				"name":       "user1",
				"service":    "pppoe",
				"caller-id":  "00:11:22:33:44:55",
				"address":    "10.10.10.2",
				"session-id": "0x81000001",
				"radius":     "false",
				"uptime":     "1d02:15:30",
			},
		},
	}

	active := ParseActiveSessions(res)
	if len(active) != 1 {
		t.Fatalf("expected 1 active session, got %d", len(active))
	}
	if active[0].Name != "user1" || active[0].Address != "10.10.10.2" || active[0].Uptime != "1d02:15:30" {
		t.Fatalf("unexpected active session parsed: %+v", active[0])
	}
}

func TestEnrichActiveSessionsWithProfiles(t *testing.T) {
	active := []PPPActiveSession{
		{RosID: "*1", Name: "user1"},
	}
	secrets := []PPPoESecret{
		{RosID: "*A", Name: "user1", Profile: "premium"},
	}
	enriched := EnrichActiveSessionsWithProfiles(active, secrets)
	if enriched[0].Profile != "premium" {
		t.Fatalf("expected profile premium, got %s", enriched[0].Profile)
	}
}
