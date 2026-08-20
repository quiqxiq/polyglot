package mikrotik

import (
	"strings"

	"github.com/quixiq/polyglot/internal/domain/command"
)

// FirewallFilterParams holds the parameters for adding a RouterOS firewall
// filter rule (/ip/firewall/filter/add). The application uses filter rules
// for hard-block suspension of static-IP subscribers.
//
// Field notes (from RouterOS /ip/firewall/filter reference):
//   - Chain          : "forward" (transit traffic) or "input" (to router).
//   - Action         : "drop", "accept", "reject", "passthrough", etc.
//   - SrcAddress     : source IP/CIDR to match. Used for per-IP block rules.
//   - SrcAddressList : address-list name to match. Used for list-based blocks.
//   - DstAddress     : destination IP/CIDR (often omitted for subscriber blocks).
//   - Protocol       : "tcp", "udp", "icmp", etc. Leave empty to match all.
//   - PlaceBefore    : rule position — "0" inserts at the top of the chain,
//                      which ensures the block rule is evaluated before any
//                      accept rules. Leave empty to append at the end.
//   - Comment        : free-text label. Convention for suspension:
//                      "SUSPENDED block_<ip> - <reason> - <timestamp>".
//   - Disabled       : when true the rule is stored but not active.
type FirewallFilterParams struct {
	Chain          string
	Action         string
	SrcAddress     string
	SrcAddressList string
	DstAddress     string
	Protocol       string
	PlaceBefore    string
	Comment        string
	Disabled       bool
}

// FirewallFilter represents one row returned by /ip/firewall/filter/print.
//
// Counter fields (Bytes, Packets) accumulate since last reset/reboot and
// are useful for monitoring which rules are actually matching traffic.
type FirewallFilter struct {
	RosID          string
	Chain          string
	Action         string
	SrcAddress     string
	SrcAddressList string
	DstAddress     string
	Protocol       string
	Bytes          string
	Packets        string
	Comment        string
	Disabled       bool
}

// BlockIPFilterParams returns a pre-filled FirewallFilterParams for blocking
// a specific subscriber IP address. This is the per-IP variant of static-IP
// suspension (chain=forward, action=drop).
func BlockIPFilterParams(customerIP, reason string) FirewallFilterParams {
	return FirewallFilterParams{
		Chain:      "forward",
		Action:     "drop",
		SrcAddress: customerIP,
		Comment:    "SUSPENDED block_" + customerIP + " - " + reason,
	}
}

// BlockAddressListFilterParams returns a pre-filled FirewallFilterParams for
// a list-based block rule (matching the "blocked_customers" address-list).
// This is created once per chain (forward and input) and serves as the base
// rule for all address-list suspension. Use PlaceBefore="0" to insert at top.
func BlockAddressListFilterParams(chain, listName string) FirewallFilterParams {
	comment := "Block suspended customers"
	if chain == "input" {
		comment = "Block suspended customers from accessing router (static IP)"
	} else {
		comment = "Block suspended customers (static IP)"
	}
	return FirewallFilterParams{
		Chain:          chain,
		Action:         "drop",
		SrcAddressList: listName,
		Comment:        comment,
		PlaceBefore:    "0",
	}
}

// FirewallFilterPrintParams holds filter criteria for /ip/firewall/filter/print.
// All fields are optional — omitting a field returns all rules in that dimension.
type FirewallFilterPrintParams struct {
	Chain          string // filter by chain name
	SrcAddress     string // filter by exact source IP
	SrcAddressList string // filter by address-list name
	Action         string // filter by action (e.g. "drop")
}

// NewPrintFirewallFiltersCommand builds the command.Command for
// /ip/firewall/filter/print with optional query filters.
func NewPrintFirewallFiltersCommand(p FirewallFilterPrintParams) command.Command {
	args := map[string]string{}
	setIfNonEmpty(args, "?chain", p.Chain)
	setIfNonEmpty(args, "?src-address", p.SrcAddress)
	setIfNonEmpty(args, "?src-address-list", p.SrcAddressList)
	setIfNonEmpty(args, "?action", p.Action)
	return command.Command{
		Raw:  "/ip/firewall/filter/print",
		Args: args,
	}
}

// NewAddFirewallFilterCommand builds the command.Command for
// /ip/firewall/filter/add.
func NewAddFirewallFilterCommand(p FirewallFilterParams) command.Command {
	args := map[string]string{
		"chain":  p.Chain,
		"action": p.Action,
	}
	setIfNonEmpty(args, "src-address", p.SrcAddress)
	setIfNonEmpty(args, "src-address-list", p.SrcAddressList)
	setIfNonEmpty(args, "dst-address", p.DstAddress)
	setIfNonEmpty(args, "protocol", p.Protocol)
	setIfNonEmpty(args, "place-before", p.PlaceBefore)
	setIfNonEmpty(args, "comment", p.Comment)
	if p.Disabled {
		args["disabled"] = "yes"
	}
	return command.Command{Raw: "/ip/firewall/filter/add", Args: args}
}

// NewRemoveFirewallFilterCommand builds the command.Command for
// /ip/firewall/filter/remove. Classified as ClassDestructive.
func NewRemoveFirewallFilterCommand(rosID string) command.Command {
	return command.Command{
		Raw:  "/ip/firewall/filter/remove",
		Args: map[string]string{".id": rosID},
	}
}

// ParseFirewallFilters converts command.Result rows from
// /ip/firewall/filter/print into typed FirewallFilter values.
// Rows missing ".id" are skipped.
func ParseFirewallFilters(result command.Result) []FirewallFilter {
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

// AddressListEntry represents one row returned by
// /ip/firewall/address-list/print.
//
// Field notes:
//   - List    : name of the address-list this entry belongs to.
//   - Dynamic : true when the entry was added by a RouterOS rule (not manually).
type AddressListEntry struct {
	RosID        string
	List         string
	Address      string
	Comment      string
	Disabled     bool
	Dynamic      bool
	CreationTime string
}

// AddressListPrintParams holds filter criteria for /ip/firewall/address-list/print.
type AddressListPrintParams struct {
	List    string // filter by list name
	Address string // filter by exact IP
}

// NewPrintAddressListCommand builds the command.Command for
// /ip/firewall/address-list/print with optional filters.
func NewPrintAddressListCommand(p AddressListPrintParams) command.Command {
	args := map[string]string{}
	setIfNonEmpty(args, "?list", p.List)
	setIfNonEmpty(args, "?address", p.Address)
	return command.Command{
		Raw:  "/ip/firewall/address-list/print",
		Args: args,
	}
}

// NewAddToAddressListCommand builds the command.Command for
// /ip/firewall/address-list/add, adding a customer IP to a named list
// (e.g. "blocked_customers"). This is the list-based suspension mechanism.
func NewAddToAddressListCommand(listName, address, comment string) command.Command {
	args := map[string]string{
		"list":    listName,
		"address": address,
	}
	setIfNonEmpty(args, "comment", comment)
	return command.Command{Raw: "/ip/firewall/address-list/add", Args: args}
}

// NewRemoveFromAddressListCommand builds the command.Command for
// /ip/firewall/address-list/remove. Classified as ClassDestructive.
func NewRemoveFromAddressListCommand(rosID string) command.Command {
	return command.Command{
		Raw:  "/ip/firewall/address-list/remove",
		Args: map[string]string{".id": rosID},
	}
}

// ParseAddressListEntries converts command.Result rows from
// /ip/firewall/address-list/print into typed AddressListEntry values.
func ParseAddressListEntries(result command.Result) []AddressListEntry {
	entries := make([]AddressListEntry, 0, len(result.Rows))
	for _, row := range result.Rows {
		id := row[".id"]
		if id == "" {
			continue
		}
		entries = append(entries, AddressListEntry{
			RosID:        id,
			List:         row["list"],
			Address:      row["address"],
			Comment:      row["comment"],
			Disabled:     strings.EqualFold(row["disabled"], "true"),
			Dynamic:      strings.EqualFold(row["dynamic"], "true"),
			CreationTime: row["creation-time"],
		})
	}
	return entries
}

// NewPrintFirewallNATRulesCommand builds the command.Command for /ip/firewall/nat/print.
func NewPrintFirewallNATRulesCommand() command.Command {
	return command.Command{
		Raw:  "/ip/firewall/nat/print",
		Args: map[string]string{},
	}
}
