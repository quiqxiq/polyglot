package iface

import (
	"testing"

	"github.com/quixiq/polyglot/internal/domain/command"
)

func TestParseInterfaces(t *testing.T) {
	res := command.Result{
		Rows: []map[string]string{
			{
				".id":         "*1",
				"name":        "ether1",
				"type":        "ether",
				"mac-address": "00:11:22:33:44:55",
				"running":     "true",
				"disabled":    "false",
			},
		},
	}

	ifaces := ParseInterfaces(res)
	if len(ifaces) != 1 {
		t.Fatalf("expected 1 interface, got %d", len(ifaces))
	}
	if ifaces[0].Name != "ether1" || !ifaces[0].Running || ifaces[0].Disabled {
		t.Fatalf("unexpected interface parsed: %+v", ifaces[0])
	}
}

func TestParseInterfaceTrafficStats(t *testing.T) {
	res := command.Result{
		Rows: []map[string]string{
			{
				"name":                "ether1",
				"rx-bits-per-second":  "10485760",
				"tx-bits-per-second":  "5242880",
				"rx-packets-per-second": "850",
				"tx-packets-per-second": "420",
			},
		},
	}

	stats := ParseInterfaceTrafficStats(res)
	if stats.RxBitsPerSecond != "10485760" || stats.TxBitsPerSecond != "5242880" {
		t.Fatalf("unexpected traffic stats parsed: %+v", stats)
	}
}

