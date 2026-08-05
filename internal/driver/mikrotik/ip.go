package mikrotik

import (
	"strconv"
	"strings"

	"github.com/quixiq/polyglot/internal/domain/command"
)

// IPAddress represents one row returned by /ip/address/print.
//
// Field notes (from RouterOS /ip/address reference):
//   - RosID     : internal ID — required for remove.
//   - Address   : IP address with prefix length in CIDR notation (e.g. "192.168.1.1/24").
//   - Network   : network address derived from Address (e.g. "192.168.1.0").
//   - Interface : name of the RouterOS interface this address is bound to.
//   - Disabled  : true when the address is configured but not active.
//   - Dynamic   : true when the address was assigned by DHCP client (not static).
//   - Comment   : free-text label.
type IPAddress struct {
	RosID     string
	Address   string // CIDR format: "ip/prefix"
	Network   string
	Interface string
	Disabled  bool
	Dynamic   bool
	Comment   string
}

// IPAddressPrintParams holds filter criteria for /ip/address/print.
type IPAddressPrintParams struct {
	Interface string // filter by interface name
	Address   string // filter by exact IP (without prefix)
}

// NewPrintIPAddressesCommand builds the command.Command for /ip/address/print
// with optional filters.
func NewPrintIPAddressesCommand(p IPAddressPrintParams) command.Command {
	args := map[string]string{}
	setIfNonEmpty(args, "?interface", p.Interface)
	setIfNonEmpty(args, "?address", p.Address)
	return command.Command{
		Raw:  "/ip/address/print",
		Args: args,
	}
}

// NewAddIPAddressCommand builds the command.Command for /ip/address/add.
// address must be in CIDR format (e.g. "10.0.0.1/24").
func NewAddIPAddressCommand(iface, address string) command.Command {
	return command.Command{
		Raw: "/ip/address/add",
		Args: map[string]string{
			"interface": iface,
			"address":   address,
		},
	}
}

// NewRemoveIPAddressCommand builds the command.Command for /ip/address/remove.
// Classified as ClassDestructive — removing an IP address can disrupt routing.
func NewRemoveIPAddressCommand(rosID string) command.Command {
	return command.Command{
		Raw:  "/ip/address/remove",
		Args: map[string]string{".id": rosID},
	}
}

// ParseIPAddresses converts command.Result rows from /ip/address/print
// into typed IPAddress values. Rows missing ".id" or "address" are skipped.
func ParseIPAddresses(result command.Result) []IPAddress {
	addrs := make([]IPAddress, 0, len(result.Rows))
	for _, row := range result.Rows {
		id := row[".id"]
		addr := row["address"]
		if id == "" || addr == "" {
			continue
		}
		addrs = append(addrs, IPAddress{
			RosID:     id,
			Address:   addr,
			Network:   row["network"],
			Interface: row["interface"],
			Disabled:  strings.EqualFold(row["disabled"], "true"),
			Dynamic:   strings.EqualFold(row["dynamic"], "true"),
			Comment:   row["comment"],
		})
	}
	return addrs
}

// IPRoute represents one row returned by /ip/route/print.
//
// Field notes (from RouterOS /ip/route reference):
//   - RosID      : internal ID — required for remove.
//   - DstAddress : destination network in CIDR format (e.g. "0.0.0.0/0" for default).
//   - Gateway    : next-hop IP or interface name.
//   - Distance   : administrative distance (1–255, lower = preferred).
//   - Active     : true when RouterOS is currently using this route.
//   - Dynamic    : true when the route was learned (OSPF, BGP, DHCP client, etc.).
//   - Static     : true for manually configured routes.
//   - Comment    : free-text label.
type IPRoute struct {
	RosID      string
	DstAddress string
	Gateway    string
	Distance   int
	Active     bool
	Dynamic    bool
	Static     bool
	Comment    string
}

// NewPrintIPRoutesCommand builds the command.Command for /ip/route/print.
// Pass a non-empty dstFilter to filter by destination network (exact match).
func NewPrintIPRoutesCommand(dstFilter string) command.Command {
	args := map[string]string{}
	if dstFilter != "" {
		args["?dst-address"] = dstFilter
	}
	return command.Command{
		Raw:  "/ip/route/print",
		Args: args,
	}
}

// NewAddIPRouteCommand builds the command.Command for /ip/route/add.
// distance defaults to "1" when empty — matching RouterOS default.
func NewAddIPRouteCommand(dstAddress, gateway, distance string) command.Command {
	dist := distance
	if dist == "" {
		dist = "1"
	}
	return command.Command{
		Raw: "/ip/route/add",
		Args: map[string]string{
			"dst-address": dstAddress,
			"gateway":     gateway,
			"distance":    dist,
		},
	}
}

// NewRemoveIPRouteCommand builds the command.Command for /ip/route/remove.
// Classified as ClassDestructive — removing a route can disrupt network connectivity.
func NewRemoveIPRouteCommand(rosID string) command.Command {
	return command.Command{
		Raw:  "/ip/route/remove",
		Args: map[string]string{".id": rosID},
	}
}

// ParseIPRoutes converts command.Result rows from /ip/route/print into
// typed IPRoute values. Rows missing ".id" or "dst-address" are skipped.
func ParseIPRoutes(result command.Result) []IPRoute {
	routes := make([]IPRoute, 0, len(result.Rows))
	for _, row := range result.Rows {
		id := row[".id"]
		dst := row["dst-address"]
		if id == "" || dst == "" {
			continue
		}
		dist, _ := strconv.Atoi(row["distance"])
		routes = append(routes, IPRoute{
			RosID:      id,
			DstAddress: dst,
			Gateway:    row["gateway"],
			Distance:   dist,
			Active:     strings.EqualFold(row["active"], "true"),
			Dynamic:    strings.EqualFold(row["dynamic"], "true"),
			Static:     strings.EqualFold(row["static"], "true"),
			Comment:    row["comment"],
		})
	}
	return routes
}
