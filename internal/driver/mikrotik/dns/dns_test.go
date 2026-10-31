package dns

import (
	"testing"

	"github.com/quixiq/polyglot/internal/domain/command"
)

func TestDNSCommands(t *testing.T) {
	cmdSettings := NewPrintDNSSettingsCommand()
	if cmdSettings.Raw != "/ip/dns/print" || cmdSettings.Args[".proplist"] == "" {
		t.Fatalf("unexpected dns settings command: %+v", cmdSettings)
	}

	cmdStatic := NewPrintDNSStaticCommand("myhost.local")
	if cmdStatic.Raw != "/ip/dns/static/print" || cmdStatic.Args["?name"] != "myhost.local" || cmdStatic.Args[".proplist"] == "" {
		t.Fatalf("unexpected dns static print command: %+v", cmdStatic)
	}

	cmdStream := NewStreamDNSStaticCommand("")
	if cmdStream.Raw != "/ip/dns/static/print" || cmdStream.Args["follow"] != "" || cmdStream.Args[".proplist"] == "" {
		t.Fatalf("unexpected dns static stream command: %+v", cmdStream)
	}

	cmdAdd := NewAddDNSStaticCommand(DNSStaticParams{
		Name:    "router.local",
		Address: "192.168.88.1",
		TTL:     "1d",
		Comment: "Gateway",
	})
	if cmdAdd.Raw != "/ip/dns/static/add" || cmdAdd.Args["name"] != "router.local" || cmdAdd.Args["address"] != "192.168.88.1" {
		t.Fatalf("unexpected dns add static command: %+v", cmdAdd)
	}
}

func TestParseDNSSettings(t *testing.T) {
	res := command.Result{
		Rows: []map[string]string{
			{
				"servers":               "1.1.1.1,8.8.8.8",
				"dynamic-servers":       "192.168.1.1",
				"allow-remote-requests": "yes",
				"cache-size":            "2048KiB",
				"cache-used":            "128KiB",
			},
		},
	}
	s := ParseDNSSettings(res)
	if s.Servers != "1.1.1.1,8.8.8.8" || !s.AllowRemote || s.CacheSize != "2048KiB" {
		t.Fatalf("unexpected dns settings parsed: %+v", s)
	}
}

func TestParseDNSStatic(t *testing.T) {
	res := command.Result{
		Rows: []map[string]string{
			{
				".id":      "*1",
				"name":     "router.local",
				"address":  "192.168.88.1",
				"ttl":      "1d",
				"comment":  "Gateway",
				"disabled": "false",
				"dynamic":  "no",
			},
		},
	}
	records := ParseDNSStatic(res)
	if len(records) != 1 || records[0].Name != "router.local" || records[0].Address != "192.168.88.1" {
		t.Fatalf("unexpected dns static parsed: %+v", records)
	}
}
