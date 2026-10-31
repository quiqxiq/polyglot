package routing

import (
	"strconv"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/driver/mikrotik/internal/rosutil"
)

// IPv6Route represents an IPv6 routing table entry (/ipv6/route).
type IPv6Route struct {
	ID         string `json:"id"`
	DstAddress string `json:"dst_address"`
	Gateway    string `json:"gateway"`
	Distance   int    `json:"distance"`
	Active     bool   `json:"active"`
	Disabled   bool   `json:"disabled"`
	Comment    string `json:"comment"`
}

// IPv6Address represents an IPv6 interface address assignment (/ipv6/address).
type IPv6Address struct {
	ID        string `json:"id"`
	Address   string `json:"address"`
	Interface string `json:"interface"`
	Disabled  bool   `json:"disabled"`
	Comment   string `json:"comment"`
}

// IPv6Settings holds global IPv6 protocol stack settings (/ipv6/settings).
type IPv6Settings struct {
	DisableIPv6 bool `json:"disable_ipv6"`
}

// NewPrintIPv6RoutesCommand queries IPv6 routes with strict proplist.
func NewPrintIPv6RoutesCommand() command.Command {
	args := make(map[string]string)
	rosutil.SetProplist(args, ".id", "dst-address", "gateway", "distance", "active", "disabled", "comment")
	return command.Command{
		Raw:  "/ipv6/route/print",
		Args: args,
	}
}

// NewStreamIPv6RoutesCommand builds a Mode A follow command for IPv6 route updates.
func NewStreamIPv6RoutesCommand() command.Command {
	args := map[string]string{
		"follow": "",
	}
	rosutil.SetProplist(args, ".id", "dst-address", "gateway", "distance", "active", "disabled", "comment")
	return command.Command{
		Raw:  "/ipv6/route/print",
		Args: args,
	}
}

// NewAddIPv6RouteCommand creates a static IPv6 route.
func NewAddIPv6RouteCommand(r IPv6Route) command.Command {
	args := map[string]string{
		"dst-address": r.DstAddress,
		"gateway":     r.Gateway,
	}
	if r.Distance > 0 {
		args["distance"] = strconv.Itoa(r.Distance)
	}
	if r.Comment != "" {
		args["comment"] = r.Comment
	}
	if r.Disabled {
		args["disabled"] = "yes"
	}
	return command.Command{
		Raw:  "/ipv6/route/add",
		Args: args,
	}
}

// NewRemoveIPv6RouteCommand removes an IPv6 route by ID.
func NewRemoveIPv6RouteCommand(id string) command.Command {
	return command.Command{
		Raw: "/ipv6/route/remove",
		Args: map[string]string{
			".id": id,
		},
	}
}

// NewPrintIPv6AddressesCommand queries IPv6 addresses with strict proplist.
func NewPrintIPv6AddressesCommand() command.Command {
	args := make(map[string]string)
	rosutil.SetProplist(args, ".id", "address", "interface", "disabled", "comment")
	return command.Command{
		Raw:  "/ipv6/address/print",
		Args: args,
	}
}

// NewAddIPv6AddressCommand adds an IPv6 address to an interface.
func NewAddIPv6AddressCommand(addr IPv6Address) command.Command {
	args := map[string]string{
		"address":   addr.Address,
		"interface": addr.Interface,
	}
	if addr.Comment != "" {
		args["comment"] = addr.Comment
	}
	if addr.Disabled {
		args["disabled"] = "yes"
	}
	return command.Command{
		Raw:  "/ipv6/address/add",
		Args: args,
	}
}

// NewRemoveIPv6AddressCommand deletes an IPv6 address by ID.
func NewRemoveIPv6AddressCommand(id string) command.Command {
	return command.Command{
		Raw: "/ipv6/address/remove",
		Args: map[string]string{
			".id": id,
		},
	}
}

// NewPrintIPv6SettingsCommand queries global IPv6 settings.
func NewPrintIPv6SettingsCommand() command.Command {
	args := make(map[string]string)
	rosutil.SetProplist(args, "disable-ipv6")
	return command.Command{
		Raw:  "/ipv6/settings/print",
		Args: args,
	}
}

// NewSetIPv6SettingsCommand toggles IPv6 stack disable status.
func NewSetIPv6SettingsCommand(disableIPv6 bool) command.Command {
	val := "no"
	if disableIPv6 {
		val = "yes"
	}
	return command.Command{
		Raw: "/ipv6/settings/set",
		Args: map[string]string{
			"disable-ipv6": val,
		},
	}
}

// ParseIPv6Routes converts reply rows into IPv6Route slice.
func ParseIPv6Routes(rows []map[string]string) []IPv6Route {
	out := make([]IPv6Route, 0, len(rows))
	for _, r := range rows {
		dist, _ := strconv.Atoi(r["distance"])
		out = append(out, IPv6Route{
			ID:         r[".id"],
			DstAddress: r["dst-address"],
			Gateway:    r["gateway"],
			Distance:   dist,
			Active:     r["active"] == "true",
			Disabled:   r["disabled"] == "true",
			Comment:    r["comment"],
		})
	}
	return out
}

// ParseIPv6Addresses converts reply rows into IPv6Address slice.
func ParseIPv6Addresses(rows []map[string]string) []IPv6Address {
	out := make([]IPv6Address, 0, len(rows))
	for _, r := range rows {
		out = append(out, IPv6Address{
			ID:        r[".id"],
			Address:   r["address"],
			Interface: r["interface"],
			Disabled:  r["disabled"] == "true",
			Comment:   r["comment"],
		})
	}
	return out
}

// ParseIPv6Settings converts reply rows into IPv6Settings.
func ParseIPv6Settings(rows []map[string]string) *IPv6Settings {
	if len(rows) == 0 {
		return nil
	}
	return &IPv6Settings{
		DisableIPv6: rows[0]["disable-ipv6"] == "yes" || rows[0]["disable-ipv6"] == "true",
	}
}
