package guard

// CheckServiceEdit evaluates an /ip/service mutation and refuses any change that would
// disable or disconnect the management API channel (api / api-ssl).

import (
	"strings"
)

// ServiceRow is one side of an /ip/service write.
type ServiceRow struct {
	Name     string
	Port     string
	Address  string // comma-separated prefixes; empty admits every address
	VRF      string
	Disabled bool
}

// CheckServiceEdit judges one /ip/service write. `ours` is "api" or
// "api-ssl"; `self` is every address the router sees MikroDash arriving from,
// and `resolved` whether that could be read at all.
func CheckServiceEdit(ours string, self []string, resolved bool, before, after ServiceRow) Verdict {
	if !strings.EqualFold(strings.TrimSpace(before.Name), ours) {
		return Verdict{Level: "none"}
	}
	refuse := func(code string, value any) Verdict {
		return Verdict{Level: "refuse", Code: code, Detail: map[string]any{"service": ours, "value": value}}
	}
	if after.Disabled && !before.Disabled {
		return refuse("service-disable", ours)
	}
	if strings.TrimSpace(after.Port) != strings.TrimSpace(before.Port) {
		return refuse("service-port", strings.TrimSpace(after.Port))
	}
	if strings.TrimSpace(after.VRF) != strings.TrimSpace(before.VRF) {
		return refuse("service-vrf", strings.TrimSpace(after.VRF))
	}
	addr := strings.TrimSpace(after.Address)
	if addr != strings.TrimSpace(before.Address) && addr != "" {
		if !resolved || len(self) == 0 {
			return refuse("service-address-unknown", addr)
		}
		for _, a := range self {
			if !serviceAdmits(addr, a) {
				return refuse("service-address", addr)
			}
		}
	}
	return Verdict{Level: "none"}
}

// serviceAdmits reports whether a comma-separated /ip/service address list
// admits one address. An entry that does not parse admits nothing, which is
// the safe direction for a refusal.
func serviceAdmits(list, address string) bool {
	for _, entry := range strings.Split(list, ",") {
		if covers, decided := AddressCovers(strings.TrimSpace(entry), []string{address}); decided && covers {
			return true
		}
	}
	return false
}

// ── THE CERTIFICATE api-ssl PRESENTS ────────────────────────────────────────
//
// MikroDash speaking api-ssl needs a certificate on the other end: remove the
// one the service names, and the service has nothing to present, the TLS
// handshake fails, and the fix is WinBox. MikroDash accepts a self-signed
// certificate, so which one is presented does not matter — only that it still
// exists. So removing it is refused; nothing else about a certificate is.
//
// Not ours when this session speaks plain api, and not ours when api-ssl names
// no certificate at all (`none`, or empty): there is nothing of ours to remove.

// CheckCertificateRemove judges removing the certificate `name`.
func CheckCertificateRemove(tls bool, apiSSLCert, name string) Verdict {
	used := strings.TrimSpace(apiSSLCert)
	if !tls || used == "" || strings.EqualFold(used, "none") {
		return Verdict{Level: "none"}
	}
	if strings.TrimSpace(name) != used {
		return Verdict{Level: "none"}
	}
	return Verdict{Level: "refuse", Code: "certificate-in-use",
		Detail: map[string]any{"service": "api-ssl", "value": used}}
}
