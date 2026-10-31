package iface

import (
	"strconv"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/driver/mikrotik/internal/rosutil"
)

// VLANInterface represents an 802.1Q tagged VLAN interface.
type VLANInterface struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	VLANID    int    `json:"vlan_id"`
	Interface string `json:"interface"`
	MTU       int    `json:"mtu"`
	Disabled  bool   `json:"disabled"`
	Comment   string `json:"comment"`
}

// NewPrintVLANInterfacesCommand queries VLAN sub-interfaces with strict proplist.
func NewPrintVLANInterfacesCommand() command.Command {
	args := make(map[string]string)
	rosutil.SetProplist(args, ".id", "name", "vlan-id", "interface", "mtu", "disabled", "comment")
	return command.Command{
		Raw:  "/interface/vlan/print",
		Args: args,
	}
}

// NewStreamVLANInterfacesCommand builds a Mode A follow command for VLAN events.
func NewStreamVLANInterfacesCommand() command.Command {
	args := map[string]string{
		"follow": "",
	}
	rosutil.SetProplist(args, ".id", "name", "vlan-id", "interface", "mtu", "disabled", "comment")
	return command.Command{
		Raw:  "/interface/vlan/print",
		Args: args,
	}
}

// NewAddVLANInterfaceCommand creates a VLAN sub-interface.
func NewAddVLANInterfaceCommand(vlan VLANInterface) command.Command {
	args := map[string]string{
		"name":      vlan.Name,
		"vlan-id":   strconv.Itoa(vlan.VLANID),
		"interface": vlan.Interface,
	}
	if vlan.MTU > 0 {
		args["mtu"] = strconv.Itoa(vlan.MTU)
	}
	if vlan.Comment != "" {
		args["comment"] = vlan.Comment
	}
	if vlan.Disabled {
		args["disabled"] = "yes"
	}
	return command.Command{
		Raw:  "/interface/vlan/add",
		Args: args,
	}
}

// NewSetVLANInterfaceCommand updates a VLAN sub-interface.
func NewSetVLANInterfaceCommand(vlan VLANInterface) command.Command {
	args := map[string]string{
		".id": vlan.ID,
	}
	if vlan.Name != "" {
		args["name"] = vlan.Name
	}
	if vlan.VLANID > 0 {
		args["vlan-id"] = strconv.Itoa(vlan.VLANID)
	}
	if vlan.Interface != "" {
		args["interface"] = vlan.Interface
	}
	if vlan.MTU > 0 {
		args["mtu"] = strconv.Itoa(vlan.MTU)
	}
	args["comment"] = vlan.Comment
	if vlan.Disabled {
		args["disabled"] = "yes"
	} else {
		args["disabled"] = "no"
	}
	return command.Command{
		Raw:  "/interface/vlan/set",
		Args: args,
	}
}

// NewRemoveVLANInterfaceCommand deletes a VLAN sub-interface by ID.
func NewRemoveVLANInterfaceCommand(id string) command.Command {
	return command.Command{
		Raw: "/interface/vlan/remove",
		Args: map[string]string{
			".id": id,
		},
	}
}

// ParseVLANInterfaces converts reply rows into VLANInterface slice.
func ParseVLANInterfaces(rows []map[string]string) []VLANInterface {
	out := make([]VLANInterface, 0, len(rows))
	for _, r := range rows {
		vlanID, _ := strconv.Atoi(r["vlan-id"])
		mtu, _ := strconv.Atoi(r["mtu"])
		out = append(out, VLANInterface{
			ID:        r[".id"],
			Name:      r["name"],
			VLANID:    vlanID,
			Interface: r["interface"],
			MTU:       mtu,
			Disabled:  r["disabled"] == "true",
			Comment:   r["comment"],
		})
	}
	return out
}
