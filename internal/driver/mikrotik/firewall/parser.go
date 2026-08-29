package firewall

import (
	"strings"

	"github.com/quixiq/polyglot/internal/domain/command"
)

// ParseFilters converts command.Result rows from /ip/firewall/filter/print.
func ParseFilters(result command.Result) []FirewallFilter {
	filters := make([]FirewallFilter, 0, len(result.Rows))
	for _, row := range result.Rows {
		id := row[".id"]
		if id == "" {
			continue
		}
		filters = append(filters, FirewallFilter{
			RosID:          id,
			Chain:          row["chain"],
			Action:         row["action"],
			SrcAddress:     row["src-address"],
			SrcAddressList: row["src-address-list"],
			DstAddress:     row["dst-address"],
			Protocol:       row["protocol"],
			Bytes:          row["bytes"],
			Packets:        row["packets"],
			Comment:        row["comment"],
			Disabled:       strings.EqualFold(row["disabled"], "true"),
		})
	}
	return filters
}

// ParseAddressList converts command.Result rows from /ip/firewall/address-list/print.
func ParseAddressList(result command.Result) []AddressListEntry {
	entries := make([]AddressListEntry, 0, len(result.Rows))
	for _, row := range result.Rows {
		id := row[".id"]
		if id == "" {
			continue
		}
		entries = append(entries, AddressListEntry{
			RosID:    id,
			List:     row["list"],
			Address:  row["address"],
			Timeout:  row["timeout"],
			Comment:  row["comment"],
			Disabled: strings.EqualFold(row["disabled"], "true"),
			Dynamic:  strings.EqualFold(row["dynamic"], "true"),
		})
	}
	return entries
}

// ParseNATRules converts command.Result rows from /ip/firewall/nat/print.
func ParseNATRules(result command.Result) []NATRule {
	rules := make([]NATRule, 0, len(result.Rows))
	for _, row := range result.Rows {
		id := row[".id"]
		if id == "" {
			continue
		}
		rules = append(rules, NATRule{
			RosID:          id,
			Chain:          row["chain"],
			Action:         row["action"],
			ToAddresses:    row["to-addresses"],
			ToPorts:        row["to-ports"],
			SrcAddress:     row["src-address"],
			SrcAddressList: row["src-address-list"],
			DstAddress:     row["dst-address"],
			Protocol:       row["protocol"],
			DstPort:        row["dst-port"],
			Bytes:          row["bytes"],
			Packets:        row["packets"],
			Comment:        row["comment"],
			Disabled:       strings.EqualFold(row["disabled"], "true"),
		})
	}
	return rules
}

// ParsePools converts command.Result rows from /ip/pool/print.
func ParsePools(result command.Result) []IPPool {
	pools := make([]IPPool, 0, len(result.Rows))
	for _, row := range result.Rows {
		id := row[".id"]
		name := row["name"]
		if id == "" || name == "" {
			continue
		}
		pools = append(pools, IPPool{
			RosID:    id,
			Name:     name,
			Ranges:   row["ranges"],
			NextPool: row["next-pool"],
			Comment:  row["comment"],
		})
	}
	return pools
}

// ParseIPAddresses converts command.Result rows from /ip/address/print.
func ParseIPAddresses(result command.Result) []IPAddress {
	addresses := make([]IPAddress, 0, len(result.Rows))
	for _, row := range result.Rows {
		id := row[".id"]
		addr := row["address"]
		if id == "" || addr == "" {
			continue
		}
		addresses = append(addresses, IPAddress{
			RosID:           id,
			Address:         addr,
			Network:         row["network"],
			Interface:       row["interface"],
			ActualInterface: row["actual-interface"],
			Disabled:        strings.EqualFold(row["disabled"], "true"),
			Dynamic:         strings.EqualFold(row["dynamic"], "true"),
			Comment:         row["comment"],
		})
	}
	return addresses
}

// ParseIPRoutes converts command.Result rows from /ip/route/print.
func ParseIPRoutes(result command.Result) []IPRoute {
	routes := make([]IPRoute, 0, len(result.Rows))
	for _, row := range result.Rows {
		id := row[".id"]
		dst := row["dst-address"]
		if id == "" || dst == "" {
			continue
		}
		routes = append(routes, IPRoute{
			RosID:      id,
			DstAddress: dst,
			Gateway:    row["gateway"],
			Distance:   row["distance"],
			Active:     strings.EqualFold(row["active"], "true"),
			Dynamic:    strings.EqualFold(row["dynamic"], "true"),
			Static:     strings.EqualFold(row["static"], "true"),
			Comment:    row["comment"],
		})
	}
	return routes
}
