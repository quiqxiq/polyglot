package vpn

import (
	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/driver/mikrotik/internal/rosutil"
)

// PPPoEServer represents a PPPoE service instance (/interface/pppoe-server/server).
type PPPoEServer struct {
	ID                string `json:"id"`
	ServiceName       string `json:"service_name"`
	Interface         string `json:"interface"`
	DefaultProfile    string `json:"default_profile"`
	OneSessionPerHost bool   `json:"one_session_per_host"`
	Disabled          bool   `json:"disabled"`
	Comment           string `json:"comment"`
}

// PPPoEClient represents a client PPPoE interface (/interface/pppoe-client).
type PPPoEClient struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Interface       string `json:"interface"`
	User            string `json:"user"`
	Profile         string `json:"profile"`
	AddDefaultRoute bool   `json:"add_default_route"`
	Disabled        bool   `json:"disabled"`
	Comment         string `json:"comment"`
}

// NewPrintPPPoEServerCommand queries PPPoE server instances with strict proplist.
func NewPrintPPPoEServerCommand() command.Command {
	args := make(map[string]string)
	rosutil.SetProplist(args, ".id", "service-name", "interface", "default-profile", "one-session-per-host", "disabled", "comment")
	return command.Command{
		Raw:  "/interface/pppoe-server/server/print",
		Args: args,
	}
}

// NewAddPPPoEServerCommand creates a PPPoE server instance.
func NewAddPPPoEServerCommand(s PPPoEServer) command.Command {
	args := map[string]string{
		"service-name": s.ServiceName,
		"interface":    s.Interface,
	}
	if s.DefaultProfile != "" {
		args["default-profile"] = s.DefaultProfile
	}
	if s.OneSessionPerHost {
		args["one-session-per-host"] = "yes"
	} else {
		args["one-session-per-host"] = "no"
	}
	if s.Comment != "" {
		args["comment"] = s.Comment
	}
	if s.Disabled {
		args["disabled"] = "yes"
	}
	return command.Command{
		Raw:  "/interface/pppoe-server/server/add",
		Args: args,
	}
}

// NewRemovePPPoEServerCommand deletes a PPPoE server instance by ID.
func NewRemovePPPoEServerCommand(id string) command.Command {
	return command.Command{
		Raw: "/interface/pppoe-server/server/remove",
		Args: map[string]string{
			".id": id,
		},
	}
}

// NewPrintPPPoEClientCommand queries PPPoE client interfaces with strict proplist.
func NewPrintPPPoEClientCommand() command.Command {
	args := make(map[string]string)
	rosutil.SetProplist(args, ".id", "name", "interface", "user", "profile", "add-default-route", "disabled", "comment")
	return command.Command{
		Raw:  "/interface/pppoe-client/print",
		Args: args,
	}
}

// NewAddPPPoEClientCommand creates a PPPoE client interface. Password is never read back.
func NewAddPPPoEClientCommand(c PPPoEClient, password string) command.Command {
	route := "no"
	if c.AddDefaultRoute {
		route = "yes"
	}
	args := map[string]string{
		"name":              c.Name,
		"interface":         c.Interface,
		"user":              c.User,
		"add-default-route": route,
	}
	if password != "" {
		args["password"] = password
	}
	if c.Profile != "" {
		args["profile"] = c.Profile
	}
	if c.Comment != "" {
		args["comment"] = c.Comment
	}
	if c.Disabled {
		args["disabled"] = "yes"
	}
	return command.Command{
		Raw:  "/interface/pppoe-client/add",
		Args: args,
	}
}

// NewRemovePPPoEClientCommand deletes a PPPoE client interface by ID.
func NewRemovePPPoEClientCommand(id string) command.Command {
	return command.Command{
		Raw: "/interface/pppoe-client/remove",
		Args: map[string]string{
			".id": id,
		},
	}
}

// ParsePPPoEServers converts reply rows into PPPoEServer slice.
func ParsePPPoEServers(rows []map[string]string) []PPPoEServer {
	out := make([]PPPoEServer, 0, len(rows))
	for _, r := range rows {
		out = append(out, PPPoEServer{
			ID:                r[".id"],
			ServiceName:       r["service-name"],
			Interface:         r["interface"],
			DefaultProfile:    r["default-profile"],
			OneSessionPerHost: r["one-session-per-host"] == "yes" || r["one-session-per-host"] == "true",
			Disabled:          r["disabled"] == "true",
			Comment:           r["comment"],
		})
	}
	return out
}

// ParsePPPoEClients converts reply rows into PPPoEClient slice.
func ParsePPPoEClients(rows []map[string]string) []PPPoEClient {
	out := make([]PPPoEClient, 0, len(rows))
	for _, r := range rows {
		out = append(out, PPPoEClient{
			ID:              r[".id"],
			Name:            r["name"],
			Interface:       r["interface"],
			User:            r["user"],
			Profile:         r["profile"],
			AddDefaultRoute: r["add-default-route"] == "yes" || r["add-default-route"] == "true",
			Disabled:        r["disabled"] == "true",
			Comment:         r["comment"],
		})
	}
	return out
}
