package dhcp

import (
	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/driver/mikrotik/internal/rosutil"
)

// DHCPClient represents a WAN DHCP client running on an interface.
type DHCPClient struct {
	ID           string `json:"id"`
	Interface    string `json:"interface"`
	Status       string `json:"status"`
	Address      string `json:"address"`
	Gateway      string `json:"gateway"`
	DHCPServer   string `json:"dhcp_server"`
	PrimaryDNS   string `json:"primary_dns"`
	SecondaryDNS string `json:"secondary_dns"`
	ExpiresAfter string `json:"expires_after"`
	Disabled     bool   `json:"disabled"`
	Comment      string `json:"comment"`
}

// NewPrintDHCPClientsCommand queries DHCP clients with strict proplist.
func NewPrintDHCPClientsCommand() command.Command {
	args := make(map[string]string)
	rosutil.SetProplist(args, ".id", "interface", "status", "address", "gateway", "dhcp-server", "primary-dns", "secondary-dns", "expires-after", "disabled", "comment")
	return command.Command{
		Raw:  "/ip/dhcp-client/print",
		Args: args,
	}
}

// NewAddDHCPClientCommand creates a DHCP client on the given interface.
func NewAddDHCPClientCommand(iface, comment string, addDefaultRoute bool) command.Command {
	route := "no"
	if addDefaultRoute {
		route = "yes"
	}
	args := map[string]string{
		"interface":         iface,
		"add-default-route": route,
	}
	if comment != "" {
		args["comment"] = comment
	}
	return command.Command{
		Raw:  "/ip/dhcp-client/add",
		Args: args,
	}
}

// NewSetDHCPClientCommand updates an existing DHCP client.
func NewSetDHCPClientCommand(id, iface, comment string, disabled bool) command.Command {
	args := map[string]string{
		".id":       id,
		"interface": iface,
		"comment":   comment,
	}
	if disabled {
		args["disabled"] = "yes"
	} else {
		args["disabled"] = "no"
	}
	return command.Command{
		Raw:  "/ip/dhcp-client/set",
		Args: args,
	}
}

// NewRemoveDHCPClientCommand removes a DHCP client by ID.
func NewRemoveDHCPClientCommand(id string) command.Command {
	return command.Command{
		Raw: "/ip/dhcp-client/remove",
		Args: map[string]string{
			".id": id,
		},
	}
}

// NewRenewDHCPClientCommand triggers a DHCP lease renewal.
func NewRenewDHCPClientCommand(id string) command.Command {
	return command.Command{
		Raw: "/ip/dhcp-client/renew",
		Args: map[string]string{
			"numbers": id,
		},
	}
}

// NewReleaseDHCPClientCommand triggers a DHCP lease release.
func NewReleaseDHCPClientCommand(id string) command.Command {
	return command.Command{
		Raw: "/ip/dhcp-client/release",
		Args: map[string]string{
			"numbers": id,
		},
	}
}

// ParseDHCPClients converts reply rows into DHCPClient slice.
func ParseDHCPClients(rows []map[string]string) []DHCPClient {
	out := make([]DHCPClient, 0, len(rows))
	for _, r := range rows {
		out = append(out, DHCPClient{
			ID:           r[".id"],
			Interface:    r["interface"],
			Status:       r["status"],
			Address:      r["address"],
			Gateway:      r["gateway"],
			DHCPServer:   r["dhcp-server"],
			PrimaryDNS:   r["primary-dns"],
			SecondaryDNS: r["secondary-dns"],
			ExpiresAfter: r["expires-after"],
			Disabled:     r["disabled"] == "true",
			Comment:      r["comment"],
		})
	}
	return out
}
