package vpn

import (
	"strconv"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/driver/mikrotik/internal/rosutil"
)

// OVPNServer represents the OpenVPN server daemon configuration on RouterOS.
type OVPNServer struct {
	Enabled     bool   `json:"enabled"`
	Port        int    `json:"port"`
	Mode        string `json:"mode"`
	Protocol    string `json:"protocol"`
	Certificate string `json:"certificate"`
	Auth        string `json:"auth"`
	Cipher      string `json:"cipher"`
}

// OVPNClient represents an outbound OpenVPN client interface (/interface/ovpn-client).
type OVPNClient struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	ConnectTo string `json:"connect_to"`
	Port      int    `json:"port"`
	User      string `json:"user"`
	Profile   string `json:"profile"`
	Cipher    string `json:"cipher"`
	Auth      string `json:"auth"`
	Disabled  bool   `json:"disabled"`
	Comment   string `json:"comment"`
}

// NewPrintOVPNServerCommand queries OpenVPN server configuration.
func NewPrintOVPNServerCommand() command.Command {
	args := make(map[string]string)
	rosutil.SetProplist(args, "enabled", "port", "mode", "protocol", "certificate", "auth", "cipher")
	return command.Command{
		Raw:  "/interface/ovpn-server/server/print",
		Args: args,
	}
}

// NewSetOVPNServerCommand updates OpenVPN server settings.
func NewSetOVPNServerCommand(s OVPNServer) command.Command {
	val := "no"
	if s.Enabled {
		val = "yes"
	}
	args := map[string]string{
		"enabled": val,
	}
	if s.Port > 0 {
		args["port"] = strconv.Itoa(s.Port)
	}
	if s.Mode != "" {
		args["mode"] = s.Mode
	}
	if s.Protocol != "" {
		args["protocol"] = s.Protocol
	}
	if s.Certificate != "" {
		args["certificate"] = s.Certificate
	}
	if s.Auth != "" {
		args["auth"] = s.Auth
	}
	if s.Cipher != "" {
		args["cipher"] = s.Cipher
	}
	return command.Command{
		Raw:  "/interface/ovpn-server/server/set",
		Args: args,
	}
}

// NewPrintOVPNClientCommand queries OpenVPN client interfaces with strict proplist.
func NewPrintOVPNClientCommand() command.Command {
	args := make(map[string]string)
	rosutil.SetProplist(args, ".id", "name", "connect-to", "port", "user", "profile", "cipher", "auth", "disabled", "comment")
	return command.Command{
		Raw:  "/interface/ovpn-client/print",
		Args: args,
	}
}

// NewAddOVPNClientCommand creates an OpenVPN client interface. Password is never read back.
func NewAddOVPNClientCommand(c OVPNClient, password string) command.Command {
	args := map[string]string{
		"name":       c.Name,
		"connect-to": c.ConnectTo,
		"user":       c.User,
	}
	if c.Port > 0 {
		args["port"] = strconv.Itoa(c.Port)
	}
	if password != "" {
		args["password"] = password
	}
	if c.Profile != "" {
		args["profile"] = c.Profile
	}
	if c.Cipher != "" {
		args["cipher"] = c.Cipher
	}
	if c.Auth != "" {
		args["auth"] = c.Auth
	}
	if c.Comment != "" {
		args["comment"] = c.Comment
	}
	if c.Disabled {
		args["disabled"] = "yes"
	}
	return command.Command{
		Raw:  "/interface/ovpn-client/add",
		Args: args,
	}
}

// NewRemoveOVPNClientCommand deletes an OpenVPN client interface by ID.
func NewRemoveOVPNClientCommand(id string) command.Command {
	return command.Command{
		Raw: "/interface/ovpn-client/remove",
		Args: map[string]string{
			".id": id,
		},
	}
}

// ParseOVPNServer converts reply rows into OVPNServer.
func ParseOVPNServer(rows []map[string]string) *OVPNServer {
	if len(rows) == 0 {
		return nil
	}
	r := rows[0]
	port, _ := strconv.Atoi(r["port"])
	return &OVPNServer{
		Enabled:     r["enabled"] == "yes" || r["enabled"] == "true",
		Port:        port,
		Mode:        r["mode"],
		Protocol:    r["protocol"],
		Certificate: r["certificate"],
		Auth:        r["auth"],
		Cipher:      r["cipher"],
	}
}

// ParseOVPNClients converts reply rows into OVPNClient slice.
func ParseOVPNClients(rows []map[string]string) []OVPNClient {
	out := make([]OVPNClient, 0, len(rows))
	for _, r := range rows {
		port, _ := strconv.Atoi(r["port"])
		out = append(out, OVPNClient{
			ID:        r[".id"],
			Name:      r["name"],
			ConnectTo: r["connect-to"],
			Port:      port,
			User:      r["user"],
			Profile:   r["profile"],
			Cipher:    r["cipher"],
			Auth:      r["auth"],
			Disabled:  r["disabled"] == "true",
			Comment:   r["comment"],
		})
	}
	return out
}
