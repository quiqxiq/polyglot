package secscan

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func parseCertExpiry(raw string) (time.Time, error) {
	formats := []string{
		"2006-01-02 15:04:05",
		time.RFC3339,
		"jan/02/2006 15:04:05",
		"Jan/02/2006 15:04:05",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, strings.TrimSpace(raw)); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unknown date format: %s", raw)
}

var checksAcctSys = []Check{
	// ── Accounts ───────────────────────────────────────────────────────────
	{
		ID: "acct.admin-user", Category: "Accounts", Severity: Medium,
		Title: "The default admin account is enabled",
		Why:   "Every attack on a MikroTik router tries \"admin\" first.",
		Fix:   "Create a personal full account, then disable or remove admin.",
		Link:  "users", Menus: []string{"/user"},
		Eval: func(in Inputs) (Status, []string) {
			rows, ok := in.Get("/user")
			if !ok {
				return Unknown, nil
			}
			for _, r := range rows {
				if r["name"] == "admin" && enabledRow(r) {
					return Fail, []string{"admin"}
				}
			}
			return Pass, nil
		},
	},
	{
		ID: "acct.no-address", Category: "Accounts", Severity: Low,
		Title: "Accounts can log in from any address",
		Why:   "A stolen password works from anywhere; an allowed-address list limits where it works.",
		Fix:   "Set Allowed Address on each account to the networks it is used from.",
		Link:  "users", Menus: []string{"/user"},
		Eval: func(in Inputs) (Status, []string) {
			rows, ok := in.Get("/user")
			if !ok {
				return Unknown, nil
			}
			var bad []string
			for _, r := range rows {
				if enabledRow(r) && strings.Trim(r["address"], ", ") == "" {
					bad = append(bad, r["name"])
				}
			}
			return verdict(bad)
		},
	},
	{
		ID: "acct.password-policy", Category: "Accounts", Severity: Low,
		Title: "The minimum password length is under 8",
		Why:   "Without a policy RouterOS accepts any password, a one-letter one included.",
		Fix:   "/user settings set minimum-password-length=12 minimum-categories=2",
		Link:  "users", Menus: []string{"/user/settings"},
		Eval: func(in Inputs) (Status, []string) {
			r, ok := in.One("/user/settings")
			if !ok {
				return Unknown, nil
			}
			n, err := strconv.Atoi(r["minimum-password-length"])
			if err != nil {
				return Unknown, nil
			}
			if n < 8 {
				return Fail, []string{"minimum-password-length=" + strconv.Itoa(n)}
			}
			return Pass, nil
		},
	},
	{
		ID: "acct.full-users", Category: "Accounts", Severity: Info,
		Title: "More than two accounts have full rights",
		Why:   "Every account that can change policy is one more password that can hand the router over.",
		Fix:   "Give day-to-day accounts a group without the policy right.",
		Link:  "users", Menus: []string{"/user", "/user/group"},
		Eval: func(in Inputs) (Status, []string) {
			users, ok := in.Get("/user")
			groups, ok2 := in.Get("/user/group")
			if !ok || !ok2 {
				return Unknown, nil
			}
			full := fullGroups(groups)
			var who []string
			for _, u := range users {
				if enabledRow(u) && full[u["group"]] {
					who = append(who, u["name"])
				}
			}
			if len(who) > 2 {
				return Fail, who
			}
			return Pass, nil
		},
	},
	{
		ID: "acct.plaintext-sessions", Category: "Accounts", Severity: Medium,
		Title: "Someone is logged in over a plaintext protocol",
		Why:   "A telnet or FTP session's password crossed the network readable.",
		Fix:   "Change that account's password, and disable telnet and ftp.",
		Link:  "users", Menus: []string{"/user/active"},
		Eval: func(in Inputs) (Status, []string) {
			rows, ok := in.Get("/user/active")
			if !ok {
				return Unknown, nil
			}
			var bad []string
			for _, r := range rows {
				if r["via"] == "telnet" || r["via"] == "ftp" {
					bad = append(bad, r["name"]+" via "+r["via"])
				}
			}
			return verdict(bad)
		},
	},

	// ── System ─────────────────────────────────────────────────────────────
	{
		ID: "sys.update", Category: "System", Severity: Medium,
		Title: "A newer RouterOS is available",
		Why:   "RouterOS releases carry security fixes, some for remotely exploitable bugs.",
		Fix:   "Install the update from the Packages page.",
		Link:  "packages", Menus: []string{"/system/package/update"},
		Eval: func(in Inputs) (Status, []string) {
			r, ok := in.One("/system/package/update")
			if !ok || (r["latest-version"] == "" && r["status"] == "") {
				return Unknown, nil
			}
			latest := r["latest-version"]
			installed := r["installed-version"]
			status := r["status"]
			isNewer := false
			if latest != "" && installed != "" && latest != installed {
				isNewer = true
			} else if strings.Contains(strings.ToLower(status), "new version") {
				isNewer = true
			}
			if isNewer {
				return Fail, []string{installed + " installed, " + latest + " available"}
			}
			return Pass, nil
		},
	},
	{
		ID: "sys.firmware", Category: "System", Severity: Low,
		Title: "The RouterBOARD firmware is behind RouterOS",
		Why:   "The boot firmware is updated separately, and carries its own fixes.",
		Fix:   "Upgrade the firmware from the Packages page, then reboot.",
		Link:  "packages", Menus: []string{"/system/routerboard"},
		Eval: func(in Inputs) (Status, []string) {
			r, ok := in.One("/system/routerboard")
			if !ok || r["current-firmware"] == "" || r["upgrade-firmware"] == "" {
				return Unknown, nil
			}
			if r["current-firmware"] != r["upgrade-firmware"] {
				return Fail, []string{r["current-firmware"] + " installed, " + r["upgrade-firmware"] + " available"}
			}
			return Pass, nil
		},
	},
	{
		ID: "sys.ntp", Category: "System", Severity: Low,
		Title: "The NTP client is disabled",
		Why:   "A wrong clock breaks certificate checks and makes the logs useless in an investigation.",
		Fix:   "Enable the NTP client with at least two servers.",
		Link:  "ntp-client", Menus: []string{"/system/ntp/client"},
		Eval: func(in Inputs) (Status, []string) {
			r, ok := in.One("/system/ntp/client")
			if !ok {
				return Unknown, nil
			}
			if !yes(r["enabled"]) {
				return Fail, []string{"enabled=no"}
			}
			return Pass, nil
		},
	},
	{
		ID: "sys.remote-log", Category: "System", Severity: Low,
		Title: "Logs are not sent off the router",
		Why:   "The router's own log is lost on reboot and can be cleared by whoever gets in.",
		Fix:   "Point a logging action at a syslog server, and send the info, warning, error and critical topics to it.",
		Link:  "logging", Menus: []string{"/system/logging", "/system/logging/action"},
		Eval: func(in Inputs) (Status, []string) {
			rules, ok := in.Get("/system/logging")
			actions, ok2 := in.Get("/system/logging/action")
			if !ok || !ok2 {
				return Unknown, nil
			}
			remote := map[string]bool{}
			for _, a := range actions {
				if a["target"] == "remote" && a["remote"] != "" && a["remote"] != "0.0.0.0" && a["remote"] != "::" {
					remote[a["name"]] = true
				}
			}
			for _, r := range rules {
				if enabledRow(r) && remote[r["action"]] {
					return Pass, nil
				}
			}
			return Fail, []string{"no logging rule uses a remote action"}
		},
	},
	{
		ID: "sys.script-permissions", Category: "System", Severity: Medium,
		Title: "A script runs without permission checks",
		Why:   "dont-require-permissions lets a script do what its owner's policy would not, when anything with lower rights starts it.",
		Fix:   "Clear Don't Require Permissions on each listed script.",
		Link:  "scripts", Menus: []string{"/system/script"},
		Eval: func(in Inputs) (Status, []string) {
			rows, ok := in.Get("/system/script")
			if !ok {
				return Unknown, nil
			}
			var bad []string
			for _, r := range rows {
				if yes(r["dont-require-permissions"]) {
					bad = append(bad, r["name"])
				}
			}
			return verdict(bad)
		},
	},
	{
		ID: "sys.device-mode", Category: "System", Severity: Info,
		Title: "Device mode allows every feature",
		Why:   "Device mode (RouterOS 7.17+) can switch off features an attacker would use, such as the scheduler, fetch and the sniffer.",
		Fix:   "Consider /system device-mode update mode=home, confirmed with the reset button.",
		Menus: []string{"/system/device-mode"},
		Eval: func(in Inputs) (Status, []string) {
			r, ok := in.One("/system/device-mode")
			if !ok || r["mode"] == "" {
				return Unknown, nil
			}
			if r["mode"] == "advanced" || r["mode"] == "enterprise" {
				return Fail, []string{"mode=" + r["mode"]}
			}
			return Pass, nil
		},
	},

	// ── Wireless ───────────────────────────────────────────────────────────
	wifiCheck("wifi.open", "A wireless network has no encryption", High,
		"Anyone in range can join an open network and read its traffic.",
		"Set WPA2 or WPA3 on every security profile.",
		func(auth, _ string) bool { return strings.TrimSpace(auth) == "" }),
	wifiCheck("wifi.weak", "A wireless network allows WPA or TKIP", Medium,
		"WPA (version 1) and TKIP are broken, and let a client downgrade the whole network.",
		"Use WPA2-PSK or WPA3 with CCMP only.",
		func(auth, cipher string) bool {
			for _, a := range strings.Split(auth, ",") {
				if a == "wpa-psk" || a == "wpa-eap" {
					return true
				}
			}
			return strings.Contains(cipher, "tkip")
		}),
	{
		ID: "wifi.wps", Category: "Wireless", Severity: Medium,
		Title: "WPS is enabled",
		Why:   "WPS lets a device join by pressing a button, or by brute-forcing its PIN, without the passphrase.",
		Fix:   "Set WPS to disable on every wifi security profile.",
		Link:  "wifi-networks", Menus: []string{"/interface/wifi/security"},
		Eval: func(in Inputs) (Status, []string) {
			rows, ok := in.Get("/interface/wifi/security")
			if !ok {
				return Unknown, nil
			}
			var bad []string
			for _, r := range rows {
				if w := r["wps"]; w != "" && w != "disable" {
					bad = append(bad, r["name"]+" (wps="+w+")")
				}
			}
			return verdict(bad)
		},
	},

	// ── Certificates ───────────────────────────────────────────────────────
	{
		ID: "cert.expiring", Category: "Certificates", Severity: Medium,
		Title: "A certificate has expired or expires within 30 days",
		Why:   "An expired certificate breaks the SSL services and VPNs that use it, and invites turning verification off.",
		Fix:   "Renew or replace the listed certificates.",
		Link:  "certificates", Menus: []string{"/certificate"},
		Eval: func(in Inputs) (Status, []string) {
			rows, ok := in.Get("/certificate")
			if !ok {
				return Unknown, nil
			}
			var bad []string
			for _, r := range rows {
				t, err := parseCertExpiry(r["invalid-after"])
				if err != nil {
					continue
				}
				now := in.Now
				if now.IsZero() {
					now = time.Now()
				}
				switch days := t.Sub(now).Hours() / 24; {
				case days < 0:
					bad = append(bad, r["name"]+" (expired "+t.Format("2006-01-02")+")")
				case days < 30:
					bad = append(bad, fmt.Sprintf("%s (expires in %d days)", r["name"], int(days)))
				}
			}
			return verdict(bad)
		},
	},
}
