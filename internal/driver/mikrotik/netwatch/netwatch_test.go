package netwatch

import (
	"testing"
)

func TestNetwatchCommandsAndParsers(t *testing.T) {
	printCmd := NewPrintNetwatchCommand()
	if printCmd.Raw != "/tool/netwatch/print" || printCmd.Args[".proplist"] == "" {
		t.Fatalf("invalid netwatch print: %v", printCmd)
	}

	streamCmd := NewStreamNetwatchCommand()
	if streamCmd.Raw != "/tool/netwatch/print" || streamCmd.Args["follow"] != "" {
		t.Fatalf("invalid stream netwatch: %v", streamCmd)
	}

	addCmd := NewAddNetwatchCommand(NetwatchHost{
		Host:     "8.8.8.8",
		Interval: "10s",
		Timeout:  "1s",
	})
	if addCmd.Raw != "/tool/netwatch/add" || addCmd.Args["host"] != "8.8.8.8" {
		t.Errorf("unexpected add netwatch: %v", addCmd)
	}

	hosts := ParseNetwatchHosts([]map[string]string{
		{".id": "*1", "host": "8.8.8.8", "status": "up", "since": "oct/28/2026 10:00:00"},
	})
	if len(hosts) != 1 || hosts[0].Host != "8.8.8.8" || hosts[0].Status != "up" {
		t.Errorf("unexpected hosts parse: %+v", hosts)
	}
}
