package secscan

var checksMgmt = []Check{
	{
		ID: "mgmt.plaintext-services", Category: "Management access", Severity: High,
		Title: "Plaintext management services are enabled",
		Why:   "Telnet, FTP, HTTP and the plain API send logins and everything after them unencrypted.",
		Fix:   "Disable telnet, ftp, www and api; manage with SSH, WinBox, www-ssl and api-ssl.",
		Link:  "ip-services", Menus: []string{"/ip/service"},
		Eval: func(in Inputs) (Status, []string) {
			rows, ok := in.Get("/ip/service")
			if !ok {
				return Unknown, nil
			}
			var bad []string
			for _, r := range rows {
				if !yes(r["dynamic"]) && enabledRow(r) && plaintextServices[r["name"]] {
					bad = append(bad, r["name"]+" (port "+r["port"]+")")
				}
			}
			return verdict(bad)
		},
	},
	{
		ID: "mgmt.unrestricted-services", Category: "Management access", Severity: Medium,
		Title: "Management services accept connections from any address",
		Why:   "A service with no allowed addresses answers the whole internet if the firewall ever lets it through.",
		Fix:   "Set Available From on every enabled service to the management networks.",
		Link:  "ip-services", Menus: []string{"/ip/service"},
		Eval: func(in Inputs) (Status, []string) {
			rows, ok := in.Get("/ip/service")
			if !ok {
				return Unknown, nil
			}
			var bad []string
			for _, r := range rows {
				if !yes(r["dynamic"]) && enabledRow(r) && !restricted(r) {
					bad = append(bad, r["name"])
				}
			}
			return verdict(bad)
		},
	},
	{
		ID: "mgmt.ssl-no-certificate", Category: "Management access", Severity: Medium,
		Title: "An SSL service is enabled without a certificate",
		Why:   "www-ssl and api-ssl with no certificate do not work, and invite switching back to the plaintext service.",
		Fix:   "Assign a certificate to www-ssl and api-ssl, or disable them.",
		Link:  "ip-services", Menus: []string{"/ip/service"},
		Eval: func(in Inputs) (Status, []string) {
			rows, ok := in.Get("/ip/service")
			if !ok {
				return Unknown, nil
			}
			var bad []string
			for _, r := range rows {
				n := r["name"]
				if (n == "www-ssl" || n == "api-ssl") && !yes(r["dynamic"]) && enabledRow(r) &&
					(r["certificate"] == "" || r["certificate"] == "none") {
					bad = append(bad, n)
				}
			}
			return verdict(bad)
		},
	},
	macCheck("mgmt.mac-telnet", "MAC Telnet is allowed on every interface", "/tool/mac-server",
		"MAC Telnet gives a login prompt to anything on the same layer-2 segment, bypassing the IP firewall."),
	macCheck("mgmt.mac-winbox", "MAC WinBox is allowed on every interface", "/tool/mac-server/mac-winbox",
		"MAC WinBox gives full management to anything on the same layer-2 segment, bypassing the IP firewall."),
	onCheck("mgmt.mac-ping", "Management access", "MAC Ping is enabled", Low, "/tool/mac-server/ping",
		"It lets layer-2 neighbours discover the router.",
		"/tool mac-server ping set enabled=no", ""),
	{
		ID: "mgmt.discovery", Category: "Management access", Severity: Low,
		Title: "Neighbour discovery runs on a list that includes the WAN",
		Why:   "MNDP, CDP and LLDP announce the model, version and identity to whoever is connected.",
		Fix:   "Set IP › Neighbors › Discovery Settings to the LAN interface list.",
		Menus: []string{"/ip/neighbor/discovery-settings"},
		Eval: func(in Inputs) (Status, []string) {
			r, ok := in.One("/ip/neighbor/discovery-settings")
			if !ok {
				return Unknown, nil
			}
			switch l := r["discover-interface-list"]; l {
			case "all", "static", "dynamic":
				return Fail, []string{"discover-interface-list=" + l}
			}
			return Pass, nil
		},
	},
	onCheck("mgmt.romon", "Management access", "RoMON is enabled", Medium, "/tool/romon",
		"RoMON lets any RoMON-capable neighbour manage this router over layer 2, outside the IP firewall.",
		"/tool romon set enabled=no, unless it is part of how the network is managed.", ""),
	{
		ID: "mgmt.bandwidth-server", Category: "Management access", Severity: Medium,
		Title: "The bandwidth-test server is enabled",
		Why:   "Anyone who reaches it can saturate the router's links and CPU; without authentication, with no login at all.",
		Fix:   "/tool bandwidth-server set enabled=no when no test is being run.",
		Menus: []string{"/tool/bandwidth-server"},
		Eval: func(in Inputs) (Status, []string) {
			r, ok := in.One("/tool/bandwidth-server")
			if !ok {
				return Unknown, nil
			}
			if !yes(r["enabled"]) {
				return Pass, nil
			}
			if !yes(r["authenticate"]) {
				return Fail, []string{"enabled, without authentication"}
			}
			return Fail, []string{"enabled"}
		},
	},
	{
		ID: "mgmt.ssh-crypto", Category: "Management access", Severity: Low,
		Title: "SSH allows weak cryptography",
		Why:   "Without strong-crypto, SSH still offers SHA-1 and older ciphers.",
		Fix:   "/ip ssh set strong-crypto=yes",
		Menus: []string{"/ip/ssh"},
		Eval: func(in Inputs) (Status, []string) {
			r, ok := in.One("/ip/ssh")
			if !ok {
				return Unknown, nil
			}
			if !yes(r["strong-crypto"]) {
				return Fail, []string{"strong-crypto=" + r["strong-crypto"]}
			}
			return Pass, nil
		},
	},
	{
		ID: "mgmt.ssh-forwarding", Category: "Management access", Severity: Low,
		Title: "SSH port forwarding is allowed",
		Why:   "A logged-in user can tunnel through the router to anything it can reach.",
		Fix:   "/ip ssh set forwarding-enabled=no",
		Menus: []string{"/ip/ssh"},
		Eval: func(in Inputs) (Status, []string) {
			r, ok := in.One("/ip/ssh")
			if !ok {
				return Unknown, nil
			}
			if f := r["forwarding-enabled"]; f != "" && f != "no" {
				return Fail, []string{"forwarding-enabled=" + f}
			}
			return Pass, nil
		},
	},
}
