package secscan

import (
	"testing"
	"time"
)

func makeInputs(menus map[string][]Row, now time.Time) Inputs {
	out := Inputs{
		Rows:   map[string][]Row{},
		Absent: map[string]bool{},
		Now:    now,
	}
	for m, rows := range menus {
		if rows == nil {
			out.Absent[m] = true
			continue
		}
		out.Rows[m] = rows
	}
	return out
}

func TestSecurityScanRun(t *testing.T) {
	now := time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC)

	// Hardened router input
	hardened := map[string][]Row{
		"/ip/service": {
			{"name": "telnet", "port": "23", "disabled": "true"},
			{"name": "ftp", "port": "21", "disabled": "true"},
			{"name": "www", "port": "80", "disabled": "true"},
			{"name": "api", "port": "8728", "disabled": "true"},
			{"name": "ssh", "port": "22", "disabled": "false", "available-from": "10.0.0.0/8"},
			{"name": "www-ssl", "port": "443", "disabled": "false", "certificate": "cert1", "available-from": "10.0.0.0/8"},
		},
		"/tool/mac-server":            {{"allowed-interface-list": "LAN"}},
		"/tool/mac-server/mac-winbox": {{"allowed-interface-list": "LAN"}},
		"/tool/mac-server/ping":       {{"enabled": "false"}},
		"/ip/neighbor/discovery-settings": {
			{"discover-interface-list": "LAN"},
		},
		"/tool/romon":            {{"enabled": "false"}},
		"/tool/bandwidth-server": {{"enabled": "false"}},
		"/ip/ssh":                {{"strong-crypto": "yes", "forwarding-enabled": "no"}},
		"/ip/firewall/filter": {
			{"chain": "input", "action": "accept", "connection-state": "established,related,untracked"},
			{"chain": "input", "action": "drop", "connection-state": "invalid"},
			{"chain": "input", "action": "drop", "in-interface-list": "!LAN"},
			{"chain": "forward", "action": "accept", "connection-state": "established,related,untracked"},
			{"chain": "forward", "action": "drop", "connection-state": "invalid"},
			{"chain": "forward", "action": "drop", "connection-nat-state": "!dstnat", "in-interface-list": "WAN"},
		},
		"/interface/list/member": {
			{"interface": "ether1", "list": "WAN"},
			{"interface": "ether2", "list": "LAN"},
		},
		"/ipv6/settings":                {{"disable-ipv6": "yes"}},
		"/ip/settings":                  {{"rp-filter": "loose", "tcp-syncookies": "true"}},
		"/ip/dns":                       {{"allow-remote-requests": "no"}},
		"/ip/upnp":                      {{"enabled": "false"}},
		"/ip/socks":                     {{"enabled": "false"}},
		"/ip/proxy":                     {{"enabled": "false"}},
		"/ip/smb":                       {{"enabled": "no", "status": "disabled"}},
		"/interface/pptp-server/server": {{"enabled": "false"}},
		"/interface/l2tp-server/server": {{"enabled": "false", "use-ipsec": "yes"}},
		"/snmp":                         {{"enabled": "false"}},
		"/ip/cloud":                     {{"ddns-enabled": "no"}},
		"/user": {
			{"name": "admin", "disabled": "true", "address": "10.0.0.0/8"},
			{"name": "superops", "disabled": "false", "group": "full", "address": "10.0.0.0/8"},
		},
		"/user/group": {
			{"name": "full", "policy": "read,write,policy"},
		},
		"/user/settings": {{"minimum-password-length": "12"}},
		"/user/active":   {{"name": "superops", "via": "ssh"}},
		"/system/package/update": {
			{"installed-version": "7.24.4", "latest-version": "7.24.4"},
		},
		"/system/routerboard": {
			{"current-firmware": "7.24.4", "upgrade-firmware": "7.24.4"},
		},
		"/system/ntp/client":     {{"enabled": "yes"}},
		"/system/logging/action": {{"name": "syslog", "target": "remote", "remote": "10.0.0.50"}},
		"/system/logging":        {{"action": "syslog", "disabled": "false"}},
		"/system/script":         {{"name": "backup", "dont-require-permissions": "false"}},
		"/system/device-mode":    {{"mode": "home"}},
		"/interface/wifi/security": {
			{"name": "sec1", "authentication-types": "wpa2-psk,wpa3-psk", "encryption": "ccmp", "wps": "disable"},
		},
		"/certificate": {
			{"name": "cert1", "invalid-after": "2027-10-28 12:00:00"},
		},
	}

	rep := Run(makeInputs(hardened, now))
	if rep.Score != 100 {
		t.Errorf("expected 100 score on hardened router, got %d. Failed checks:", rep.Score)
		for _, f := range rep.Findings {
			if f.Status == Fail {
				t.Errorf("  - %s: %s (detail: %v)", f.ID, f.Title, f.Detail)
			}
		}
	}
	if rep.Passed == 0 {
		t.Error("expected passed checks")
	}

	// Test critical cap: vulnerable router where input chain does not drop catch-all
	vulnerable := map[string][]Row{
		"/ip/firewall/filter": {
			{"chain": "input", "action": "accept", "protocol": "tcp", "dst-port": "80"},
		},
		"/ip/service": {
			{"name": "telnet", "port": "23", "disabled": "false"},
		},
	}
	vulnRep := Run(makeInputs(vulnerable, now))
	if vulnRep.Score > CriticalCap {
		t.Errorf("expected score <= CriticalCap (%d), got %d", CriticalCap, vulnRep.Score)
	}
	if vulnRep.Failed["critical"] == 0 {
		t.Error("expected critical failure for input-default-drop")
	}
}

func TestCheckCatalogueCompleteness(t *testing.T) {
	if len(Checks) < 35 {
		t.Errorf("expected at least 35 checks in catalogue, got %d", len(Checks))
	}
	seen := map[string]bool{}
	for _, c := range Checks {
		if c.ID == "" {
			t.Error("found check with empty ID")
		}
		if seen[c.ID] {
			t.Errorf("duplicate check ID: %s", c.ID)
		}
		seen[c.ID] = true
		if c.Category == "" {
			t.Errorf("check %s has empty category", c.ID)
		}
		if c.Eval == nil {
			t.Errorf("check %s has nil Eval function", c.ID)
		}
	}
}
