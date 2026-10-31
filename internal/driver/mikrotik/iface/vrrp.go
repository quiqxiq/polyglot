package iface

import (
	"strconv"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/driver/mikrotik/internal/rosutil"
)

// VRRPInterface represents a Virtual Router Redundancy Protocol interface.
type VRRPInterface struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Interface string `json:"interface"`
	VRID      int    `json:"vrid"`
	Priority  int    `json:"priority"`
	Interval  int    `json:"interval"`
	Version   int    `json:"version"`
	Running   bool   `json:"running"`
	Backup    bool   `json:"backup"`
	Disabled  bool   `json:"disabled"`
	Comment   string `json:"comment"`
}

// DetectInternetState represents WAN internet reachability auto-detection state.
type DetectInternetState struct {
	Interface string `json:"interface"`
	State     string `json:"state"`
	Internet  string `json:"internet"`
}

// NewPrintVRRPCommand queries VRRP interfaces with strict proplist.
func NewPrintVRRPCommand() command.Command {
	args := make(map[string]string)
	rosutil.SetProplist(args, ".id", "name", "interface", "vrid", "priority", "interval", "version", "running", "backup", "disabled", "comment")
	return command.Command{
		Raw:  "/interface/vrrp/print",
		Args: args,
	}
}

// NewAddVRRPCommand creates a VRRP interface.
func NewAddVRRPCommand(v VRRPInterface) command.Command {
	args := map[string]string{
		"name":      v.Name,
		"interface": v.Interface,
		"vrid":      strconv.Itoa(v.VRID),
	}
	if v.Priority > 0 {
		args["priority"] = strconv.Itoa(v.Priority)
	}
	if v.Interval > 0 {
		args["interval"] = strconv.Itoa(v.Interval)
	}
	if v.Version > 0 {
		args["version"] = strconv.Itoa(v.Version)
	}
	if v.Comment != "" {
		args["comment"] = v.Comment
	}
	if v.Disabled {
		args["disabled"] = "yes"
	}
	return command.Command{
		Raw:  "/interface/vrrp/add",
		Args: args,
	}
}

// NewSetVRRPCommand updates a VRRP interface.
func NewSetVRRPCommand(v VRRPInterface) command.Command {
	args := map[string]string{
		".id": v.ID,
	}
	if v.Name != "" {
		args["name"] = v.Name
	}
	if v.VRID > 0 {
		args["vrid"] = strconv.Itoa(v.VRID)
	}
	if v.Priority > 0 {
		args["priority"] = strconv.Itoa(v.Priority)
	}
	args["comment"] = v.Comment
	if v.Disabled {
		args["disabled"] = "yes"
	} else {
		args["disabled"] = "no"
	}
	return command.Command{
		Raw:  "/interface/vrrp/set",
		Args: args,
	}
}

// NewRemoveVRRPCommand deletes a VRRP interface by ID.
func NewRemoveVRRPCommand(id string) command.Command {
	return command.Command{
		Raw: "/interface/vrrp/remove",
		Args: map[string]string{
			".id": id,
		},
	}
}

// NewPrintDetectInternetCommand queries WAN internet detection status.
func NewPrintDetectInternetCommand() command.Command {
	args := make(map[string]string)
	rosutil.SetProplist(args, "interface", "state", "internet")
	return command.Command{
		Raw:  "/interface/detect-internet/state/print",
		Args: args,
	}
}

// ParseVRRPInterfaces converts reply rows into VRRPInterface slice.
func ParseVRRPInterfaces(rows []map[string]string) []VRRPInterface {
	out := make([]VRRPInterface, 0, len(rows))
	for _, r := range rows {
		vrid, _ := strconv.Atoi(r["vrid"])
		pri, _ := strconv.Atoi(r["priority"])
		interval, _ := strconv.Atoi(r["interval"])
		ver, _ := strconv.Atoi(r["version"])
		out = append(out, VRRPInterface{
			ID:        r[".id"],
			Name:      r["name"],
			Interface: r["interface"],
			VRID:      vrid,
			Priority:  pri,
			Interval:  interval,
			Version:   ver,
			Running:   r["running"] == "true",
			Backup:    r["backup"] == "true",
			Disabled:  r["disabled"] == "true",
			Comment:   r["comment"],
		})
	}
	return out
}

// ParseDetectInternetState converts reply rows into DetectInternetState slice.
func ParseDetectInternetState(rows []map[string]string) []DetectInternetState {
	out := make([]DetectInternetState, 0, len(rows))
	for _, r := range rows {
		out = append(out, DetectInternetState{
			Interface: r["interface"],
			State:     r["state"],
			Internet:  r["internet"],
		})
	}
	return out
}
