package secscan

import (
	"strconv"
	"strings"
)

func yes(v string) bool { return v == "true" || v == "yes" }

func enabledRow(r Row) bool { return !yes(r["disabled"]) }

func verdict(bad []string) (Status, []string) {
	if len(bad) > 0 {
		return Fail, bad
	}
	return Pass, nil
}

func restricted(r Row) bool { return strings.Trim(r["available-from"]+r["address"], ", ") != "" }

func positive(v string) bool { return v != "" && !strings.HasPrefix(v, "!") }

var fwNotMatchers = map[string]bool{
	".id": true, "chain": true, "action": true, "disabled": true, "dynamic": true, "invalid": true,
	"bytes": true, "packets": true, "comment": true, "log": true, "log-prefix": true,
	"reject-with": true, "jump-target": true,
}

func catchAll(r Row) bool {
	for k, v := range r {
		if fwNotMatchers[k] || v == "" {
			continue
		}
		if (k == "in-interface" || k == "in-interface-list") && strings.HasPrefix(v, "!") {
			continue
		}
		return false
	}
	return true
}

func dropping(r Row) bool { return r["action"] == "drop" || r["action"] == "reject" }

func endsInDrop(rows []Row, chain string) bool {
	for _, r := range rows {
		if r["chain"] == chain && enabledRow(r) && dropping(r) && catchAll(r) {
			return true
		}
	}
	return false
}

func hasState(r Row, state string) bool {
	for _, s := range strings.Split(r["connection-state"], ",") {
		if s == state {
			return true
		}
	}
	return false
}

func acceptsEstablished(rows []Row, chain string) bool {
	for _, r := range rows {
		if r["chain"] == chain && enabledRow(r) && hasState(r, "established") &&
			(r["action"] == "accept" || r["action"] == "fasttrack-connection") {
			return true
		}
	}
	return false
}

func dropsInvalid(rows []Row, chain string) bool {
	for _, r := range rows {
		if r["chain"] == chain && enabledRow(r) && dropping(r) && hasState(r, "invalid") {
			return true
		}
	}
	return false
}

func portIn(spec string, p int) bool {
	for _, part := range strings.Split(spec, ",") {
		lo, hi, isRange := strings.Cut(part, "-")
		a, err := strconv.Atoi(strings.TrimSpace(lo))
		if err != nil {
			continue
		}
		b := a
		if isRange {
			if b, err = strconv.Atoi(strings.TrimSpace(hi)); err != nil {
				continue
			}
		}
		if p >= a && p <= b {
			return true
		}
	}
	return false
}

func fullGroups(groups []Row) map[string]bool {
	out := map[string]bool{}
	for _, g := range groups {
		has := map[string]bool{}
		for _, p := range strings.Split(g["policy"], ",") {
			has[p] = true
		}
		if has["write"] && has["policy"] {
			out[g["name"]] = true
		}
	}
	return out
}

var plaintextServices = map[string]bool{"telnet": true, "ftp": true, "www": true, "api": true}

func macCheck(id, title, menu, why string) Check {
	return Check{
		ID: id, Category: "Management access", Severity: Medium, Title: title, Why: why,
		Fix:   "Set its Allowed Interface List to the LAN list (or none).",
		Menus: []string{menu},
		Eval: func(in Inputs) (Status, []string) {
			r, ok := in.One(menu)
			if !ok {
				return Unknown, nil
			}
			if l := r["allowed-interface-list"]; l == "all" || l == "static" || l == "dynamic" {
				return Fail, []string{"allowed-interface-list=" + l}
			}
			return Pass, nil
		},
	}
}

func fwHas(chain string, has func([]Row, string) bool, missing string) func(Inputs) (Status, []string) {
	return func(in Inputs) (Status, []string) {
		rows, ok := in.Get("/ip/firewall/filter")
		if !ok {
			return Unknown, nil
		}
		if has(rows, chain) {
			return Pass, nil
		}
		return Fail, []string{missing}
	}
}

func ipSetting(id, title, key, bad, why, fix string) Check {
	return Check{
		ID: id, Category: "Firewall", Severity: Low, Title: title, Why: why, Fix: fix,
		Menus: []string{"/ip/settings"},
		Eval: func(in Inputs) (Status, []string) {
			r, ok := in.One("/ip/settings")
			if !ok {
				return Unknown, nil
			}
			if r[key] == bad {
				return Fail, []string{key + "=" + bad}
			}
			return Pass, nil
		},
	}
}

func onCheck(id, category, title string, sev Severity, menu, why, fix, link string) Check {
	return Check{
		ID: id, Category: category, Severity: sev, Title: title, Why: why, Fix: fix, Link: link,
		Menus: []string{menu},
		Eval: func(in Inputs) (Status, []string) {
			r, ok := in.One(menu)
			if !ok {
				return Unknown, nil
			}
			if yes(r["enabled"]) {
				return Fail, []string{"enabled"}
			}
			return Pass, nil
		},
	}
}

func snmpCheck(id, title string, sev Severity, why, fix string, bad func(Row) bool) Check {
	return Check{
		ID: id, Category: "Exposed services", Severity: sev, Title: title, Why: why, Fix: fix, Link: "snmp",
		Menus: []string{"/snmp", "/snmp/community"},
		Eval: func(in Inputs) (Status, []string) {
			s, ok := in.One("/snmp")
			if !ok {
				return Unknown, nil
			}
			if !yes(s["enabled"]) {
				return Pass, nil
			}
			rows, ok := in.Get("/snmp/community")
			if !ok {
				return Unknown, nil
			}
			var hit []string
			for _, r := range rows {
				if enabledRow(r) && bad(r) {
					hit = append(hit, r["name"])
				}
			}
			return verdict(hit)
		},
	}
}

func wifiCheck(id, title string, sev Severity, why, fix string, bad func(auth, cipher string) bool) Check {
	return Check{
		ID: id, Category: "Wireless", Severity: sev, Title: title, Why: why, Fix: fix, Link: "wifi-networks",
		Menus: []string{"/interface/wifi/security", "/interface/wireless/security-profiles"},
		Eval: func(in Inputs) (Status, []string) {
			wifi, okW := in.Get("/interface/wifi/security")
			legacy, okL := in.Get("/interface/wireless/security-profiles")
			if !okW && !okL {
				return Unknown, nil
			}
			var hit []string
			for _, r := range wifi {
				if bad(r["authentication-types"], r["encryption"]) {
					hit = append(hit, r["name"])
				}
			}
			for _, r := range legacy {
				auth := r["authentication-types"]
				if r["mode"] == "none" {
					auth = ""
				}
				if bad(auth, r["unicast-ciphers"]) {
					hit = append(hit, r["name"])
				}
			}
			return verdict(hit)
		},
	}
}

func buildFacts(in Inputs) Facts {
	f := Facts{Services: []ServiceFact{}}
	if rows, ok := in.Get("/ip/service"); ok {
		for _, r := range rows {
			if yes(r["dynamic"]) {
				continue
			}
			f.Services = append(f.Services, ServiceFact{
				Name:       r["name"],
				Port:       r["port"],
				Enabled:    enabledRow(r),
				Restricted: restricted(r),
				Plaintext:  plaintextServices[r["name"]],
			})
		}
	}
	if r, ok := in.One("/system/package/update"); ok {
		f.Installed, f.Latest = r["installed-version"], r["latest-version"]
	}
	if r, ok := in.One("/system/routerboard"); ok {
		f.FirmwareCurrent, f.FirmwareUpgrade = r["current-firmware"], r["upgrade-firmware"]
	}
	if users, ok := in.Get("/user"); ok {
		full := map[string]bool{}
		if groups, ok := in.Get("/user/group"); ok {
			full = fullGroups(groups)
		}
		for _, u := range users {
			if !enabledRow(u) {
				continue
			}
			f.Users++
			if full[u["group"]] {
				f.FullUsers++
			}
			if u["name"] == "admin" {
				f.AdminEnabled = true
			}
		}
	}
	if r, ok := in.One("/user/settings"); ok {
		f.MinPasswordLen = r["minimum-password-length"]
	}
	return f
}
