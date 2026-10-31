package secscan

import (
	"strings"
)

var checksServices = []Check{
	onCheck("svc.upnp", "Exposed services", "UPnP is enabled", High, "/ip/upnp",
		"UPnP lets any device on the LAN open ports on the router to the internet, malware included.",
		"/ip upnp set enabled=no", ""),
	onCheck("svc.socks", "Exposed services", "The SOCKS proxy is enabled", High, "/ip/socks",
		"An open SOCKS proxy relays anyone's traffic through your address.",
		"/ip socks set enabled=no", ""),
	onCheck("svc.proxy", "Exposed services", "The web proxy is enabled", Medium, "/ip/proxy",
		"A web proxy reachable from outside becomes an open relay for anyone's traffic.",
		"/ip proxy set enabled=no, or restrict it with access rules.", ""),
	{
		ID: "svc.smb", Category: "Exposed services", Severity: Medium,
		Title: "The SMB file server is running",
		Why:   "SMB has a long history of remote vulnerabilities and should not run on a router.",
		Fix:   "/ip smb set enabled=no",
		Menus: []string{"/ip/smb"},
		Eval: func(in Inputs) (Status, []string) {
			r, ok := in.One("/ip/smb")
			if !ok {
				return Unknown, nil
			}
			if r["enabled"] == "yes" || (r["status"] != "" && r["status"] != "disabled") {
				return Fail, []string{"enabled=" + r["enabled"] + ", status=" + r["status"]}
			}
			return Pass, nil
		},
	},
	onCheck("svc.pptp", "Exposed services", "A PPTP server is enabled", High, "/interface/pptp-server/server",
		"PPTP's MS-CHAPv2 authentication is broken: captured logins are cracked offline.",
		"Disable the PPTP server and move clients to WireGuard, IKEv2 or L2TP/IPsec.", "vpn"),
	{
		ID: "svc.l2tp-no-ipsec", Category: "Exposed services", Severity: Medium,
		Title: "The L2TP server runs without IPsec",
		Why:   "L2TP alone does not encrypt: tunnel traffic and logins cross the network in the clear.",
		Fix:   "/interface l2tp-server server set use-ipsec=required",
		Link:  "vpn", Menus: []string{"/interface/l2tp-server/server"},
		Eval: func(in Inputs) (Status, []string) {
			r, ok := in.One("/interface/l2tp-server/server")
			if !ok {
				return Unknown, nil
			}
			if yes(r["enabled"]) && (r["use-ipsec"] == "no" || r["use-ipsec"] == "") {
				return Fail, []string{"use-ipsec=" + r["use-ipsec"]}
			}
			return Pass, nil
		},
	},
	snmpCheck("svc.snmp-public", "SNMP answers the default community or no security", Medium,
		"The default community \"public\", or security=none, lets anyone who can reach SNMP read the router's configuration.",
		"Rename or disable the public community, and set security to authorized or private.",
		func(r Row) bool { return r["name"] == "public" || r["security"] == "none" }),
	snmpCheck("svc.snmp-write", "An SNMP community has write access", High,
		"SNMP write access can change the router's configuration.",
		"Turn off write-access on every community.",
		func(r Row) bool { return yes(r["write-access"]) }),
	snmpCheck("svc.snmp-anywhere", "An SNMP community is open to any address", Medium,
		"A community with addresses 0.0.0.0/0 or ::/0 answers the whole internet if the firewall lets SNMP through.",
		"Restrict each community's addresses to the monitoring hosts.",
		func(r Row) bool {
			a := r["addresses"]
			return a == "" || strings.Contains(a, "0.0.0.0/0") || strings.Contains(a, "::/0")
		}),
	{
		ID: "svc.cloud-ddns", Category: "Exposed services", Severity: Info,
		Title: "IP Cloud DDNS publishes the router's address",
		Why:   "The router's public address is published under a predictable MikroTik hostname.",
		Fix:   "/ip cloud set ddns-enabled=no if nothing relies on it.",
		Menus: []string{"/ip/cloud"},
		Eval: func(in Inputs) (Status, []string) {
			r, ok := in.One("/ip/cloud")
			if !ok {
				return Unknown, nil
			}
			if yes(r["ddns-enabled"]) {
				return Fail, []string{"ddns-enabled=yes"}
			}
			return Pass, nil
		},
	},
}
