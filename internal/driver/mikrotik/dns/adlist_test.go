package dns

import (
	"testing"
)

func TestDNSAdlistAndForwarders(t *testing.T) {
	fwdCmd := NewPrintDNSForwardersCommand()
	if fwdCmd.Raw != "/ip/dns/forwarders/print" || fwdCmd.Args[".proplist"] == "" {
		t.Fatalf("invalid forwarder print cmd: %v", fwdCmd)
	}

	adlistCmd := NewPrintDNSAdlistsCommand()
	if adlistCmd.Raw != "/ip/dns/adlist/print" || adlistCmd.Args[".proplist"] == "" {
		t.Fatalf("invalid adlist print cmd: %v", adlistCmd)
	}

	reloadCmd := NewReloadDNSAdlistCommand()
	if reloadCmd.Raw != "/ip/dns/adlist/reload" {
		t.Errorf("invalid reload cmd: %v", reloadCmd)
	}

	flushCmd := NewFlushDNSCacheCommand()
	if flushCmd.Raw != "/ip/dns/cache/flush" {
		t.Errorf("invalid flush cmd: %v", flushCmd)
	}

	fwds := ParseDNSForwarders([]map[string]string{
		{
			".id":             "*1",
			"name":            "cloudflare",
			"dns-servers":     "1.1.1.1,1.0.0.1",
			"verify-doh-cert": "yes",
			"disabled":        "false",
		},
	})
	if len(fwds) != 1 || len(fwds[0].DNSServers) != 2 || !fwds[0].VerifyDoHCert {
		t.Errorf("unexpected fwds parse: %+v", fwds)
	}

	adlists := ParseDNSAdlists([]map[string]string{
		{
			".id":         "*1",
			"url":         "https://example.com/hosts.txt",
			"ssl-verify":  "yes",
			"name-count":  "50000",
			"match-count": "120",
		},
	})
	if len(adlists) != 1 || adlists[0].NameCount != 50000 || adlists[0].MatchCount != 120 {
		t.Errorf("unexpected adlists parse: %+v", adlists)
	}
}
