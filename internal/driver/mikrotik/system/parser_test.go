package system

import (
	"testing"

	"github.com/quixiq/polyglot/internal/domain/command"
)

func TestParseResource(t *testing.T) {
	res := command.Result{
		Rows: []map[string]string{
			{
				"cpu-load":     "15",
				"cpu-count":    "4",
				"free-memory":  "536870912",
				"total-memory": "1073741824",
				"uptime":       "5d12h30m",
				"version":      "7.12.1",
				"board-name":   "RB4011iGS+",
			},
		},
	}

	resource := ParseResource(res)
	if resource.CPULoad != 15 || resource.CPUCount != 4 || resource.BoardName != "RB4011iGS+" {
		t.Fatalf("unexpected resource parsed: %+v", resource)
	}
}

func TestParsePing(t *testing.T) {
	res := command.Result{
		Rows: []map[string]string{
			{
				"seq":         "1",
				"host":        "8.8.8.8",
				"sent":        "1",
				"received":    "1",
				"packet-loss": "0",
				"time":        "12ms",
				"ttl":         "116",
			},
		},
	}

	pings := ParsePing(res)
	if len(pings) != 1 || pings[0].Host != "8.8.8.8" || pings[0].Time != "12ms" {
		t.Fatalf("unexpected ping parsed: %+v", pings)
	}
}
