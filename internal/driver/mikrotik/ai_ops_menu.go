package mikrotik

import (
	"strings"
)

// NormalizeMenuPath maps common aliases or incomplete menu paths to valid RouterOS API menu paths.
func NormalizeMenuPath(menu string) string {
	m := strings.Trim(strings.ToLower(strings.TrimSpace(menu)), "/")
	switch m {
	case "resource", "system/resources", "resources", "system-resource":
		return "system/resource"
	case "system":
		// When asked to inspect /system, the intention is system resources and performance.
		return "system/resource"
	case "health", "system-health":
		return "system/health"
	case "identity", "system-identity":
		return "system/identity"
	case "routerboard", "board", "system-routerboard":
		return "system/routerboard"
	case "clock", "time", "system-clock":
		return "system/clock"
	case "package", "packages", "system-package":
		return "system/package"
	case "license", "system-license":
		return "system/license"
	case "interfaces":
		return "interface"
	case "ethernet", "ethernets", "interface/ethernets":
		return "interface/ethernet"
	case "bridge", "bridges", "interface/bridges":
		return "interface/bridge"
	case "bridge/port", "bridge-port", "bridge-ports", "interface/bridge/ports":
		return "interface/bridge/port"
	case "vlan", "vlans", "interface/vlans":
		return "interface/vlan"
	case "wireless", "wifi", "interface/wifi", "interface/wirelesses":
		return "interface/wireless"
	case "wireguard", "interface/wireguards":
		return "interface/wireguard"
	case "bonding", "interface/bondings":
		return "interface/bonding"
	case "address", "addresses", "ip/addresses", "ip-address":
		return "ip/address"
	case "route", "routes", "ip/routes", "routing", "ip-route":
		return "ip/route"
	case "arp", "arps", "ip/arps":
		return "ip/arp"
	case "dns", "ip/dns-cache":
		return "ip/dns"
	case "pool", "pools", "ip/pools":
		return "ip/pool"
	case "neighbor", "neighbors", "ip/neighbors":
		return "ip/neighbor"
	case "dhcp", "dhcp-lease", "dhcp-leases", "dhcp/lease", "dhcp/leases", "lease", "leases", "ip/dhcp-server/leases":
		return "ip/dhcp-server/lease"
	case "dhcp-server", "dhcp-servers", "ip/dhcp-servers":
		return "ip/dhcp-server"
	case "dhcp-client", "dhcp-clients", "ip/dhcp-clients":
		return "ip/dhcp-client"
	case "firewall", "firewall/filter", "filter", "ip/firewall/filters":
		return "ip/firewall/filter"
	case "firewall/nat", "nat", "ip/firewall/nats":
		return "ip/firewall/nat"
	case "firewall/mangle", "mangle", "ip/firewall/mangles":
		return "ip/firewall/mangle"
	case "firewall/raw", "raw":
		return "ip/firewall/raw"
	case "firewall/address-list", "address-list", "address-lists", "ip/firewall/address-lists":
		return "ip/firewall/address-list"
	case "firewall/connection", "connections", "connection", "ip/firewall/connections":
		return "ip/firewall/connection"
	case "service", "services", "ip/services":
		return "ip/service"
	case "queue", "queues", "simple-queue", "queue/simple-queues":
		return "queue/simple"
	case "hotspot/user", "hotspot-user", "hotspot-users", "hotspot/users":
		return "ip/hotspot/user"
	case "hotspot/active", "hotspot-active":
		return "ip/hotspot/active"
	case "ppp/active", "ppp-active":
		return "ppp/active"
	case "ppp/secret", "ppp-secret", "ppp-secrets", "ppp/secrets":
		return "ppp/secret"
	case "log", "logs", "system/log", "system/logs":
		return "log"
	case "certificate", "certificates":
		return "certificate"
	case "user", "users":
		return "user"
	default:
		return m
	}
}

func commandRaw(menu, verb string) string {
	menu = strings.Trim(menu, "/")
	if menu == "" {
		return "/" + verb
	}
	return "/" + menu + "/" + verb
}

func sanitizeSlug(in string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(in) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-', r == '_':
			b.WriteRune(r)
		case r == '/', r == '.':
			b.WriteByte('-')
		}
	}
	return strings.Trim(b.String(), "-")
}
