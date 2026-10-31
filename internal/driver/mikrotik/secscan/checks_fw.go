package secscan

import (
	"sort"
	"strings"
)

var checksFW = []Check{
	{
		ID: "fw.input-default-drop", Category: "Firewall", Severity: Critical,
		Title: "The input chain does not drop what it has not accepted",
		Why:   "With no final drop, every service on the router is reachable from every interface, the internet included.",
		Fix:   "End the input chain with a drop for everything not from the LAN (the defconf's drop in-interface-list=!LAN).",
		Link:  "firewall", Menus: []string{"/ip/firewall/filter"},
		Eval: func(in Inputs) (Status, []string) {
			rows, ok := in.Get("/ip/firewall/filter")
			if !ok {
				return Unknown, nil
			}
			if endsInDrop(rows, "input") {
				return Pass, nil
			}
			return Fail, []string{"no drop or reject that catches the rest of the input chain"}
		},
	},
	{
		ID: "fw.input-established", Category: "Firewall", Severity: Medium,
		Title: "The input chain does not accept established connections",
		Why:   "Without it, replies to the router's own traffic are judged rule by rule, and a final drop breaks them.",
		Fix:   "Put accept connection-state=established,related,untracked at the top of the input chain.",
		Link:  "firewall", Menus: []string{"/ip/firewall/filter"},
		Eval: fwHas("input", acceptsEstablished, "no accept for established,related"),
	},
	{
		ID: "fw.input-invalid", Category: "Firewall", Severity: Low,
		Title: "The input chain does not drop invalid packets",
		Why:   "Invalid packets are malformed or out of state, and are a common probe.",
		Fix:   "Add drop connection-state=invalid near the top of the input chain.",
		Link:  "firewall", Menus: []string{"/ip/firewall/filter"},
		Eval: fwHas("input", dropsInvalid, "no drop for connection-state=invalid"),
	},
	{
		ID: "fw.forward-wan", Category: "Firewall", Severity: High,
		Title: "Forwarded traffic from the WAN is not dropped",
		Why:   "Without it, anything on the internet that can route to the LAN reaches it, not only port-forwarded services.",
		Fix:   "Add to the forward chain: drop connection-nat-state=!dstnat in-interface-list=WAN (the defconf rule).",
		Link:  "firewall", Menus: []string{"/ip/firewall/filter"},
		Eval: func(in Inputs) (Status, []string) {
			rows, ok := in.Get("/ip/firewall/filter")
			if !ok {
				return Unknown, nil
			}
			if endsInDrop(rows, "forward") {
				return Pass, nil
			}
			for _, r := range rows {
				if r["chain"] == "forward" && enabledRow(r) && dropping(r) && r["connection-nat-state"] == "!dstnat" &&
					(positive(r["in-interface-list"]) || positive(r["in-interface"])) {
					return Pass, nil
				}
			}
			return Fail, []string{"no drop for new connections arriving from the WAN"}
		},
	},
	{
		ID: "fw.forward-invalid", Category: "Firewall", Severity: Low,
		Title: "The forward chain does not drop invalid packets",
		Why:   "Invalid packets are malformed or out of state, and the LAN should not see them.",
		Fix:   "Add drop connection-state=invalid to the forward chain.",
		Link:  "firewall", Menus: []string{"/ip/firewall/filter"},
		Eval: fwHas("forward", dropsInvalid, "no drop for connection-state=invalid"),
	},
	{
		ID: "fw.dns-open-resolver", Category: "Firewall", Severity: High,
		Title: "The DNS server answers remote requests with nothing blocking the WAN",
		Why:   "An open resolver is used to amplify denial-of-service attacks against others, from your address.",
		Fix:   "Drop input to port 53 from the WAN, or end the input chain with a drop; or turn off Allow Remote Requests.",
		Link:  "dns", Menus: []string{"/ip/dns", "/ip/firewall/filter"},
		Eval: func(in Inputs) (Status, []string) {
			dns, ok := in.One("/ip/dns")
			if !ok {
				return Unknown, nil
			}
			if !yes(dns["allow-remote-requests"]) {
				return Pass, nil
			}
			rows, ok := in.Get("/ip/firewall/filter")
			if !ok {
				return Unknown, nil
			}
			if endsInDrop(rows, "input") {
				return Pass, nil
			}
			for _, r := range rows {
				if r["chain"] == "input" && enabledRow(r) && dropping(r) && portIn(r["dst-port"], 53) {
					return Pass, nil
				}
			}
			return Fail, []string{"allow-remote-requests=yes, and no input rule drops port 53"}
		},
	},
	{
		ID: "fw.wan-in-lan", Category: "Firewall", Severity: High,
		Title: "An interface is in both the WAN and LAN lists",
		Why:   "Rules written as \"not LAN\" or \"from WAN\" then disagree about it, and the defconf's input drop never applies to it.",
		Fix:   "Remove the WAN interface from the LAN interface list.",
		Link:  "interface-lists", Menus: []string{"/interface/list/member"},
		Eval: func(in Inputs) (Status, []string) {
			rows, ok := in.Get("/interface/list/member")
			if !ok {
				return Unknown, nil
			}
			wan, lan := map[string]bool{}, map[string]bool{}
			for _, r := range rows {
				if !enabledRow(r) {
					continue
				}
				switch strings.ToUpper(r["list"]) {
				case "WAN":
					wan[r["interface"]] = true
				case "LAN":
					lan[r["interface"]] = true
				}
			}
			var bad []string
			for i := range wan {
				if lan[i] {
					bad = append(bad, i)
				}
			}
			sort.Strings(bad)
			return verdict(bad)
		},
	},
	{
		ID: "fw.ipv6-input", Category: "Firewall", Severity: High,
		Title: "IPv6 is enabled and its input chain does not drop the rest",
		Why:   "IPv6 addresses are globally routable: an open IPv6 input chain exposes the router however tight the IPv4 one is.",
		Fix:   "End the IPv6 input chain with a drop for everything not from the LAN, or disable IPv6.",
		Link:  "firewall", Menus: []string{"/ipv6/settings", "/ipv6/firewall/filter"},
		Eval: func(in Inputs) (Status, []string) {
			s, ok := in.One("/ipv6/settings")
			if !ok {
				return Unknown, nil
			}
			if yes(s["disable-ipv6"]) {
				return Pass, nil
			}
			rows, ok := in.Get("/ipv6/firewall/filter")
			if !ok {
				return Unknown, nil
			}
			if endsInDrop(rows, "input") {
				return Pass, nil
			}
			return Fail, []string{"no drop or reject that catches the rest of the IPv6 input chain"}
		},
	},
	ipSetting("fw.rp-filter", "Reverse-path filtering is off", "rp-filter", "no",
		"Without it the router accepts packets with spoofed source addresses on the wrong interface.",
		"/ip settings set rp-filter=loose (or strict on a single-homed router)."),
	ipSetting("fw.syncookies", "TCP SYN cookies are off", "tcp-syncookies", "false",
		"SYN cookies keep the router answering during a SYN flood.",
		"/ip settings set tcp-syncookies=yes"),
}
