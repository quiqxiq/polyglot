package system

import (
	"strings"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/driver/mikrotik/internal/rosutil"
)

// NTPClientConfig holds current RouterOS NTP client settings and status.
type NTPClientConfig struct {
	Enabled bool     `json:"enabled"`
	Mode    string   `json:"mode"`
	Servers []string `json:"servers"`
	Status  string   `json:"status"`
}

// NTPServer represents one configured NTP server address.
type NTPServer struct {
	ID              string `json:"id"`
	Address         string `json:"address"`
	ResolvedAddress string `json:"resolved_address"`
	Disabled        bool   `json:"disabled"`
	Comment         string `json:"comment"`
}

// NewPrintNTPClientCommand queries NTP client settings with strict proplist.
func NewPrintNTPClientCommand() command.Command {
	args := make(map[string]string)
	rosutil.SetProplist(args, "enabled", "mode", "servers", "status")
	return command.Command{
		Raw:  "/system/ntp/client/print",
		Args: args,
	}
}

// NewSetNTPClientCommand updates NTP client settings.
func NewSetNTPClientCommand(enabled bool, mode string, servers []string) command.Command {
	val := "no"
	if enabled {
		val = "yes"
	}
	args := map[string]string{
		"enabled": val,
	}
	if mode != "" {
		args["mode"] = mode
	}
	if len(servers) > 0 {
		args["servers"] = strings.Join(servers, ",")
	}
	return command.Command{
		Raw:  "/system/ntp/client/set",
		Args: args,
	}
}

// NewPrintNTPServersCommand queries configured NTP server targets.
func NewPrintNTPServersCommand() command.Command {
	args := make(map[string]string)
	rosutil.SetProplist(args, ".id", "address", "resolved-address", "disabled", "comment")
	return command.Command{
		Raw:  "/system/ntp/client/servers/print",
		Args: args,
	}
}

// NewAddNTPServerCommand adds an NTP server target.
func NewAddNTPServerCommand(address, comment string) command.Command {
	args := map[string]string{
		"address": address,
	}
	if comment != "" {
		args["comment"] = comment
	}
	return command.Command{
		Raw:  "/system/ntp/client/servers/add",
		Args: args,
	}
}

// NewSetNTPServerCommand updates an NTP server target.
func NewSetNTPServerCommand(id, address, comment string, disabled bool) command.Command {
	args := map[string]string{
		".id": id,
	}
	if address != "" {
		args["address"] = address
	}
	args["comment"] = comment
	if disabled {
		args["disabled"] = "yes"
	} else {
		args["disabled"] = "no"
	}
	return command.Command{
		Raw:  "/system/ntp/client/servers/set",
		Args: args,
	}
}

// NewRemoveNTPServerCommand removes an NTP server target.
func NewRemoveNTPServerCommand(id string) command.Command {
	return command.Command{
		Raw: "/system/ntp/client/servers/remove",
		Args: map[string]string{
			".id": id,
		},
	}
}

// ParseNTPClient converts RouterOS reply rows into NTPClientConfig.
func ParseNTPClient(rows []map[string]string) *NTPClientConfig {
	if len(rows) == 0 {
		return nil
	}
	s := rows[0]
	var servers []string
	if raw := s["servers"]; raw != "" {
		for _, srv := range strings.Split(raw, ",") {
			if srv = strings.TrimSpace(srv); srv != "" {
				servers = append(servers, srv)
			}
		}
	}
	return &NTPClientConfig{
		Enabled: s["enabled"] == "yes" || s["enabled"] == "true",
		Mode:    s["mode"],
		Servers: servers,
		Status:  s["status"],
	}
}

// ParseNTPServers converts reply rows into NTPServer slice.
func ParseNTPServers(rows []map[string]string) []NTPServer {
	out := make([]NTPServer, 0, len(rows))
	for _, s := range rows {
		out = append(out, NTPServer{
			ID:              s[".id"],
			Address:         s["address"],
			ResolvedAddress: s["resolved-address"],
			Disabled:        s["disabled"] == "true",
			Comment:         s["comment"],
		})
	}
	return out
}

// NTPServerSettings holds NTP server daemon configuration (/system/ntp/server).
type NTPServerSettings struct {
	Enabled   bool `json:"enabled"`
	Manycast  bool `json:"manycast"`
	Multicast bool `json:"multicast"`
	Broadcast bool `json:"broadcast"`
}

// NewPrintNTPServerSettingsCommand queries NTP server daemon settings.
func NewPrintNTPServerSettingsCommand() command.Command {
	args := make(map[string]string)
	rosutil.SetProplist(args, "enabled", "manycast", "multicast", "broadcast")
	return command.Command{
		Raw:  "/system/ntp/server/print",
		Args: args,
	}
}

// NewSetNTPServerSettingsCommand updates NTP server daemon settings.
func NewSetNTPServerSettingsCommand(s NTPServerSettings) command.Command {
	val := "no"
	if s.Enabled {
		val = "yes"
	}
	many := "no"
	if s.Manycast {
		many = "yes"
	}
	multi := "no"
	if s.Multicast {
		multi = "yes"
	}
	broad := "no"
	if s.Broadcast {
		broad = "yes"
	}
	return command.Command{
		Raw: "/system/ntp/server/set",
		Args: map[string]string{
			"enabled":   val,
			"manycast":  many,
			"multicast": multi,
			"broadcast": broad,
		},
	}
}

// ParseNTPServerSettings converts reply rows into NTPServerSettings.
func ParseNTPServerSettings(rows []map[string]string) *NTPServerSettings {
	if len(rows) == 0 {
		return nil
	}
	r := rows[0]
	return &NTPServerSettings{
		Enabled:   r["enabled"] == "yes" || r["enabled"] == "true",
		Manycast:  r["manycast"] == "yes" || r["manycast"] == "true",
		Multicast: r["multicast"] == "yes" || r["multicast"] == "true",
		Broadcast: r["broadcast"] == "yes" || r["broadcast"] == "true",
	}
}
