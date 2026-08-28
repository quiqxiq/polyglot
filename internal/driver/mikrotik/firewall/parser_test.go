package firewall

import (
	"testing"

	"github.com/quixiq/polyglot/internal/domain/command"
)

func TestParseFilters(t *testing.T) {
	res := command.Result{
		Rows: []map[string]string{
			{
				".id":         "*1",
				"chain":       "forward",
				"action":      "drop",
				"src-address": "192.168.1.100",
				"comment":     "SUSPENDED",
				"disabled":    "false",
			},
		},
	}

	filters := ParseFilters(res)
	if len(filters) != 1 || filters[0].SrcAddress != "192.168.1.100" || filters[0].Disabled {
		t.Fatalf("unexpected filters parsed: %+v", filters)
	}
}

func TestParseAddressList(t *testing.T) {
	res := command.Result{
		Rows: []map[string]string{
			{
				".id":      "*10",
				"list":     "ISOLIR_USERS",
				"address":  "10.0.0.50",
				"comment":  "budi",
				"disabled": "false",
			},
		},
	}

	entries := ParseAddressList(res)
	if len(entries) != 1 || entries[0].List != "ISOLIR_USERS" || entries[0].Address != "10.0.0.50" {
		t.Fatalf("unexpected address list parsed: %+v", entries)
	}
}

func TestParseNATRules(t *testing.T) {
	res := command.Result{
		Rows: []map[string]string{
			{
				".id":          "*5",
				"chain":        "dstnat",
				"action":       "dst-nat",
				"to-addresses": "10.0.0.1",
				"to-ports":     "80",
				"comment":      "ISOLATION_REDIRECT",
				"disabled":     "false",
			},
		},
	}

	rules := ParseNATRules(res)
	if len(rules) != 1 || rules[0].ToAddresses != "10.0.0.1" || rules[0].Disabled {
		t.Fatalf("unexpected nat rules parsed: %+v", rules)
	}
}

