package routing

import (
	"testing"

	"github.com/quixiq/polyglot/internal/domain/command"
)

func TestRoutingCommands(t *testing.T) {
	cmdBGP := NewPrintBGPSessionsCommand("isp-upstream")
	if cmdBGP.Raw != "/routing/bgp/session/print" || cmdBGP.Args["?name"] != "isp-upstream" || cmdBGP.Args[".proplist"] == "" {
		t.Fatalf("unexpected bgp session command: %+v", cmdBGP)
	}

	cmdOSPF := NewPrintOSPFNeighborsCommand()
	if cmdOSPF.Raw != "/routing/ospf/neighbor/print" || cmdOSPF.Args[".proplist"] == "" {
		t.Fatalf("unexpected ospf neighbor command: %+v", cmdOSPF)
	}

	cmdTables := NewPrintRoutingTablesCommand()
	if cmdTables.Raw != "/routing/table/print" || cmdTables.Args[".proplist"] == "" {
		t.Fatalf("unexpected routing table command: %+v", cmdTables)
	}
}

func TestParseBGPSessions(t *testing.T) {
	res := command.Result{
		Rows: []map[string]string{
			{
				".id":            "*1",
				"name":           "isp-upstream",
				"remote.address": "203.0.113.1",
				"remote.as":      "64512",
				"local.address":  "203.0.113.2",
				"local.as":       "64513",
				"state":          "established",
				"uptime":         "2w3d",
				"prefix-count":   "150",
				"disabled":       "false",
			},
		},
	}
	sessions := ParseBGPSessions(res)
	if len(sessions) != 1 || sessions[0].Name != "isp-upstream" || sessions[0].State != "established" || sessions[0].PrefixesReceived != "150" {
		t.Fatalf("unexpected bgp session parsed: %+v", sessions)
	}
}
