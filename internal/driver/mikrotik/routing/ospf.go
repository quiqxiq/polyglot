package routing

import (
	"strconv"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/driver/mikrotik/internal/rosutil"
)

// OSPFInstance represents an OSPF process/instance in RouterOS v7.
type OSPFInstance struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Version  int    `json:"version"`
	RouterID string `json:"router_id"`
	Disabled bool   `json:"disabled"`
	Comment  string `json:"comment"`
}

// OSPFArea represents an OSPF area under an instance.
type OSPFArea struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Instance    string `json:"instance"`
	AreaID      string `json:"area_id"`
	NoSummaries bool   `json:"no_summaries"`
	Disabled    bool   `json:"disabled"`
	Comment     string `json:"comment"`
}

// OSPFTemplate represents an OSPF interface template configuration.
type OSPFTemplate struct {
	ID         string `json:"id"`
	Area       string `json:"area"`
	Interfaces string `json:"interfaces"`
	Type       string `json:"type"`
	Cost       int    `json:"cost"`
	Priority   int    `json:"priority"`
	Passive    bool   `json:"passive"`
	Disabled   bool   `json:"disabled"`
	Comment    string `json:"comment"`
}

// NewPrintOSPFInstancesCommand queries OSPF instances with strict proplist.
func NewPrintOSPFInstancesCommand() command.Command {
	args := make(map[string]string)
	rosutil.SetProplist(args, ".id", "name", "version", "router-id", "disabled", "comment")
	return command.Command{
		Raw:  "/routing/ospf/instance/print",
		Args: args,
	}
}

// NewAddOSPFInstanceCommand creates an OSPF instance.
func NewAddOSPFInstanceCommand(inst OSPFInstance) command.Command {
	args := map[string]string{
		"name": inst.Name,
	}
	if inst.Version > 0 {
		args["version"] = strconv.Itoa(inst.Version)
	}
	if inst.RouterID != "" {
		args["router-id"] = inst.RouterID
	}
	if inst.Comment != "" {
		args["comment"] = inst.Comment
	}
	if inst.Disabled {
		args["disabled"] = "yes"
	}
	return command.Command{
		Raw:  "/routing/ospf/instance/add",
		Args: args,
	}
}

// NewRemoveOSPFInstanceCommand deletes an OSPF instance by ID.
func NewRemoveOSPFInstanceCommand(id string) command.Command {
	return command.Command{
		Raw: "/routing/ospf/instance/remove",
		Args: map[string]string{
			".id": id,
		},
	}
}

// NewPrintOSPFAreasCommand queries OSPF areas with strict proplist.
func NewPrintOSPFAreasCommand() command.Command {
	args := make(map[string]string)
	rosutil.SetProplist(args, ".id", "name", "instance", "area-id", "no-summaries", "disabled", "comment")
	return command.Command{
		Raw:  "/routing/ospf/area/print",
		Args: args,
	}
}

// NewAddOSPFAreaCommand creates an OSPF area.
func NewAddOSPFAreaCommand(area OSPFArea) command.Command {
	args := map[string]string{
		"name":     area.Name,
		"instance": area.Instance,
		"area-id":  area.AreaID,
	}
	if area.NoSummaries {
		args["no-summaries"] = ""
	}
	if area.Comment != "" {
		args["comment"] = area.Comment
	}
	if area.Disabled {
		args["disabled"] = "yes"
	}
	return command.Command{
		Raw:  "/routing/ospf/area/add",
		Args: args,
	}
}

// NewRemoveOSPFAreaCommand removes an OSPF area by ID.
func NewRemoveOSPFAreaCommand(id string) command.Command {
	return command.Command{
		Raw: "/routing/ospf/area/remove",
		Args: map[string]string{
			".id": id,
		},
	}
}

// NewPrintOSPFTemplatesCommand queries OSPF interface templates with strict proplist.
func NewPrintOSPFTemplatesCommand() command.Command {
	args := make(map[string]string)
	rosutil.SetProplist(args, ".id", "area", "interfaces", "type", "cost", "priority", "passive", "disabled", "comment")
	return command.Command{
		Raw:  "/routing/ospf/interface-template/print",
		Args: args,
	}
}

// NewAddOSPFTemplateCommand creates an OSPF interface template.
func NewAddOSPFTemplateCommand(tmpl OSPFTemplate) command.Command {
	args := map[string]string{
		"area":       tmpl.Area,
		"interfaces": tmpl.Interfaces,
	}
	if tmpl.Type != "" {
		args["type"] = tmpl.Type
	}
	if tmpl.Cost > 0 {
		args["cost"] = strconv.Itoa(tmpl.Cost)
	}
	if tmpl.Priority > 0 {
		args["priority"] = strconv.Itoa(tmpl.Priority)
	}
	if tmpl.Passive {
		args["passive"] = ""
	}
	if tmpl.Comment != "" {
		args["comment"] = tmpl.Comment
	}
	if tmpl.Disabled {
		args["disabled"] = "yes"
	}
	return command.Command{
		Raw:  "/routing/ospf/interface-template/add",
		Args: args,
	}
}

// NewMoveOSPFTemplateCommand moves an ordered template before destination.
func NewMoveOSPFTemplateCommand(id, destination string) command.Command {
	args := map[string]string{
		"numbers": id,
	}
	if destination != "" {
		args["destination"] = destination
	}
	return command.Command{
		Raw:  "/routing/ospf/interface-template/move",
		Args: args,
	}
}

// NewRemoveOSPFTemplateCommand removes an OSPF template by ID.
func NewRemoveOSPFTemplateCommand(id string) command.Command {
	return command.Command{
		Raw: "/routing/ospf/interface-template/remove",
		Args: map[string]string{
			".id": id,
		},
	}
}

// ParseOSPFInstances converts reply rows into OSPFInstance slice.
func ParseOSPFInstances(rows []map[string]string) []OSPFInstance {
	out := make([]OSPFInstance, 0, len(rows))
	for _, r := range rows {
		v, _ := strconv.Atoi(r["version"])
		out = append(out, OSPFInstance{
			ID:       r[".id"],
			Name:     r["name"],
			Version:  v,
			RouterID: r["router-id"],
			Disabled: r["disabled"] == "true",
			Comment:  r["comment"],
		})
	}
	return out
}

// ParseOSPFAreas converts reply rows into OSPFArea slice.
func ParseOSPFAreas(rows []map[string]string) []OSPFArea {
	out := make([]OSPFArea, 0, len(rows))
	for _, r := range rows {
		out = append(out, OSPFArea{
			ID:          r[".id"],
			Name:        r["name"],
			Instance:    r["instance"],
			AreaID:      r["area-id"],
			NoSummaries: r["no-summaries"] == "true",
			Disabled:    r["disabled"] == "true",
			Comment:     r["comment"],
		})
	}
	return out
}

// ParseOSPFTemplates converts reply rows into OSPFTemplate slice.
func ParseOSPFTemplates(rows []map[string]string) []OSPFTemplate {
	out := make([]OSPFTemplate, 0, len(rows))
	for _, r := range rows {
		cost, _ := strconv.Atoi(r["cost"])
		pri, _ := strconv.Atoi(r["priority"])
		out = append(out, OSPFTemplate{
			ID:         r[".id"],
			Area:       r["area"],
			Interfaces: r["interfaces"],
			Type:       r["type"],
			Cost:       cost,
			Priority:   pri,
			Passive:    r["passive"] == "true",
			Disabled:   r["disabled"] == "true",
			Comment:    r["comment"],
		})
	}
	return out
}
