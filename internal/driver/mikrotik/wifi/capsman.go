package wifi

import (
	"strconv"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/driver/mikrotik/internal/rosutil"
)

// CAPsMANManager represents the RouterOS CAPsMAN controller state (/interface/wifi/capsman).
type CAPsMANManager struct {
	Enabled       bool   `json:"enabled"`
	Interfaces    string `json:"interfaces"`
	UpgradePolicy string `json:"upgrade_policy"`
}

// RemoteCAP represents a managed access point connected to CAPsMAN.
type RemoteCAP struct {
	ID         string `json:"id"`
	Address    string `json:"address"`
	Identity   string `json:"identity"`
	BoardName  string `json:"board_name"`
	Serial     string `json:"serial"`
	Version    string `json:"version"`
	BaseMac    string `json:"base_mac"`
	CommonName string `json:"common_name"`
	State      string `json:"state"`
}

// CAPsChannel represents a CAPsMAN channel frequency profile.
type CAPsChannel struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Band      string `json:"band"`
	Frequency string `json:"frequency"`
	Width     string `json:"width"`
	Disabled  bool   `json:"disabled"`
	Comment   string `json:"comment"`
}

// CAPsDatapath represents a CAPsMAN datapath routing profile.
type CAPsDatapath struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Bridge          string `json:"bridge"`
	VLANID          int    `json:"vlan_id"`
	LocalForwarding bool   `json:"local_forwarding"`
	Disabled        bool   `json:"disabled"`
	Comment         string `json:"comment"`
}

// CAPsSecurity represents a CAPsMAN wireless security profile.
type CAPsSecurity struct {
	ID                  string `json:"id"`
	Name                string `json:"name"`
	AuthenticationTypes string `json:"authentication_types"`
	Encryption          string `json:"encryption"`
	WPS                 string `json:"wps"`
	Disabled            bool   `json:"disabled"`
	Comment             string `json:"comment"`
}

// CAPsConfiguration represents a master CAPsMAN WiFi configuration profile.
type CAPsConfiguration struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Mode     string `json:"mode"`
	SSID     string `json:"ssid"`
	Channel  string `json:"channel"`
	Security string `json:"security"`
	Datapath string `json:"datapath"`
	Disabled bool   `json:"disabled"`
	Comment  string `json:"comment"`
}

// CAPsProvisioning represents an AP provisioning rule in CAPsMAN.
type CAPsProvisioning struct {
	ID                  string `json:"id"`
	RadioMAC            string `json:"radio_mac"`
	MasterConfiguration string `json:"master_configuration"`
	SlaveConfigurations string `json:"slave_configurations"`
	Action              string `json:"action"`
	Disabled            bool   `json:"disabled"`
	Comment             string `json:"comment"`
}

// NewPrintCAPsMANManagerCommand queries CAPsMAN manager configuration.
func NewPrintCAPsMANManagerCommand() command.Command {
	args := make(map[string]string)
	rosutil.SetProplist(args, "enabled", "interfaces", "upgrade-policy")
	return command.Command{
		Raw:  "/interface/wifi/capsman/print",
		Args: args,
	}
}

// NewPrintRemoteCAPsCommand queries connected remote CAP access points.
func NewPrintRemoteCAPsCommand() command.Command {
	args := make(map[string]string)
	rosutil.SetProplist(args, ".id", "address", "identity", "board-name", "serial", "version", "base-mac", "common-name", "state")
	return command.Command{
		Raw:  "/interface/wifi/capsman/remote-cap/print",
		Args: args,
	}
}

// NewUpgradeRemoteCAPCommand triggers firmware upgrade of a remote CAP.
func NewUpgradeRemoteCAPCommand(id string) command.Command {
	return command.Command{
		Raw: "/interface/wifi/capsman/remote-cap/upgrade",
		Args: map[string]string{
			".id": id,
		},
	}
}

// NewPrintCAPsChannelsCommand queries CAPsMAN channel profiles.
func NewPrintCAPsChannelsCommand() command.Command {
	args := make(map[string]string)
	rosutil.SetProplist(args, ".id", "name", "band", "frequency", "width", "disabled", "comment")
	return command.Command{
		Raw:  "/interface/wifi/channel/print",
		Args: args,
	}
}

// NewPrintCAPsDatapathsCommand queries CAPsMAN datapath profiles.
func NewPrintCAPsDatapathsCommand() command.Command {
	args := make(map[string]string)
	rosutil.SetProplist(args, ".id", "name", "bridge", "vlan-id", "local-forwarding", "disabled", "comment")
	return command.Command{
		Raw:  "/interface/wifi/datapath/print",
		Args: args,
	}
}

// NewPrintCAPsSecurityCommand queries CAPsMAN security profiles.
func NewPrintCAPsSecurityCommand() command.Command {
	args := make(map[string]string)
	rosutil.SetProplist(args, ".id", "name", "authentication-types", "encryption", "wps", "disabled", "comment")
	return command.Command{
		Raw:  "/interface/wifi/security/print",
		Args: args,
	}
}

// NewPrintCAPsConfigurationsCommand queries CAPsMAN configurations.
func NewPrintCAPsConfigurationsCommand() command.Command {
	args := make(map[string]string)
	rosutil.SetProplist(args, ".id", "name", "mode", "ssid", "channel", "security", "datapath", "disabled", "comment")
	return command.Command{
		Raw:  "/interface/wifi/configuration/print",
		Args: args,
	}
}

// NewPrintCAPsProvisioningCommand queries CAPsMAN provisioning rules.
func NewPrintCAPsProvisioningCommand() command.Command {
	args := make(map[string]string)
	rosutil.SetProplist(args, ".id", "radio-mac", "master-configuration", "slave-configurations", "action", "disabled", "comment")
	return command.Command{
		Raw:  "/interface/wifi/provisioning/print",
		Args: args,
	}
}

// ParseCAPsMANManager converts reply rows into CAPsMANManager.
func ParseCAPsMANManager(rows []map[string]string) *CAPsMANManager {
	if len(rows) == 0 {
		return nil
	}
	r := rows[0]
	return &CAPsMANManager{
		Enabled:       r["enabled"] == "yes" || r["enabled"] == "true",
		Interfaces:    r["interfaces"],
		UpgradePolicy: r["upgrade-policy"],
	}
}

// ParseRemoteCAPs converts reply rows into RemoteCAP slice.
func ParseRemoteCAPs(rows []map[string]string) []RemoteCAP {
	out := make([]RemoteCAP, 0, len(rows))
	for _, r := range rows {
		out = append(out, RemoteCAP{
			ID:         r[".id"],
			Address:    r["address"],
			Identity:   r["identity"],
			BoardName:  r["board-name"],
			Serial:     r["serial"],
			Version:    r["version"],
			BaseMac:    r["base-mac"],
			CommonName: r["common-name"],
			State:      r["state"],
		})
	}
	return out
}

// ParseCAPsDatapaths converts reply rows into CAPsDatapath slice.
func ParseCAPsDatapaths(rows []map[string]string) []CAPsDatapath {
	out := make([]CAPsDatapath, 0, len(rows))
	for _, r := range rows {
		vid, _ := strconv.Atoi(r["vlan-id"])
		out = append(out, CAPsDatapath{
			ID:              r[".id"],
			Name:            r["name"],
			Bridge:          r["bridge"],
			VLANID:          vid,
			LocalForwarding: r["local-forwarding"] == "yes" || r["local-forwarding"] == "true",
			Disabled:        r["disabled"] == "true",
			Comment:         r["comment"],
		})
	}
	return out
}
