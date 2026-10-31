package container

import (
	"strings"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/driver/mikrotik/internal/rosutil"
)

// ContainerEnv represents a container environment variable (/container/envs).
type ContainerEnv struct {
	ID      string `json:"id"`
	List    string `json:"list"`
	Key     string `json:"key"`
	Value   string `json:"value,omitempty"` // Omitted from readback for secret protection
	Comment string `json:"comment"`
}

// ContainerMount represents a host directory bind mount (/container/mounts).
type ContainerMount struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Src     string `json:"src"`
	Dst     string `json:"dst"`
	Comment string `json:"comment"`
}

// VethInterface represents a Virtual Ethernet interface (/interface/veth).
type VethInterface struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Address  string `json:"address"`
	Gateway  string `json:"gateway"`
	Running  bool   `json:"running"`
	Disabled bool   `json:"disabled"`
	Comment  string `json:"comment"`
}

// NewPrintContainerEnvsCommand queries container environment variables with strict proplist.
// Note: value is deliberately not in proplist to prevent credential leakage.
func NewPrintContainerEnvsCommand(listFilter string) command.Command {
	args := make(map[string]string)
	rosutil.SetProplist(args, ".id", "list", "key", "comment")
	if listFilter != "" {
		args["?list"] = listFilter
	}
	return command.Command{
		Raw:  "/container/envs/print",
		Args: args,
	}
}

// NewAddContainerEnvCommand adds an environment variable.
func NewAddContainerEnvCommand(e ContainerEnv) command.Command {
	args := map[string]string{
		"list":  e.List,
		"key":   e.Key,
		"value": e.Value,
	}
	rosutil.SetIfNonEmpty(args, "comment", e.Comment)
	return command.Command{
		Raw:  "/container/envs/add",
		Args: args,
	}
}

// NewSetContainerEnvCommand updates an environment variable.
func NewSetContainerEnvCommand(e ContainerEnv) command.Command {
	args := map[string]string{
		".id": e.ID,
	}
	rosutil.SetIfNonEmpty(args, "list", e.List)
	rosutil.SetIfNonEmpty(args, "key", e.Key)
	rosutil.SetIfNonEmpty(args, "value", e.Value)
	args["comment"] = e.Comment
	return command.Command{
		Raw:  "/container/envs/set",
		Args: args,
	}
}

// NewRemoveContainerEnvCommand deletes an environment variable.
func NewRemoveContainerEnvCommand(id string) command.Command {
	return command.Command{
		Raw: "/container/envs/remove",
		Args: map[string]string{
			".id": id,
		},
	}
}

// NewPrintContainerMountsCommand queries container bind mounts with strict proplist.
func NewPrintContainerMountsCommand() command.Command {
	args := make(map[string]string)
	rosutil.SetProplist(args, ".id", "name", "src", "dst", "comment")
	return command.Command{
		Raw:  "/container/mounts/print",
		Args: args,
	}
}

// NewAddContainerMountCommand creates a bind mount.
func NewAddContainerMountCommand(m ContainerMount) command.Command {
	args := map[string]string{
		"name": m.Name,
		"src":  m.Src,
		"dst":  m.Dst,
	}
	rosutil.SetIfNonEmpty(args, "comment", m.Comment)
	return command.Command{
		Raw:  "/container/mounts/add",
		Args: args,
	}
}

// NewSetContainerMountCommand updates a bind mount.
func NewSetContainerMountCommand(m ContainerMount) command.Command {
	args := map[string]string{
		".id": m.ID,
	}
	rosutil.SetIfNonEmpty(args, "name", m.Name)
	rosutil.SetIfNonEmpty(args, "src", m.Src)
	rosutil.SetIfNonEmpty(args, "dst", m.Dst)
	args["comment"] = m.Comment
	return command.Command{
		Raw:  "/container/mounts/set",
		Args: args,
	}
}

// NewRemoveContainerMountCommand removes a bind mount.
func NewRemoveContainerMountCommand(id string) command.Command {
	return command.Command{
		Raw: "/container/mounts/remove",
		Args: map[string]string{
			".id": id,
		},
	}
}

// NewPrintVethInterfacesCommand queries Virtual Ethernet interfaces with strict proplist.
func NewPrintVethInterfacesCommand() command.Command {
	args := make(map[string]string)
	rosutil.SetProplist(args, ".id", "name", "address", "gateway", "running", "disabled", "comment")
	return command.Command{
		Raw:  "/interface/veth/print",
		Args: args,
	}
}

// NewAddVethInterfaceCommand creates a Virtual Ethernet interface.
func NewAddVethInterfaceCommand(v VethInterface) command.Command {
	args := map[string]string{
		"name": v.Name,
	}
	rosutil.SetIfNonEmpty(args, "address", v.Address)
	rosutil.SetIfNonEmpty(args, "gateway", v.Gateway)
	rosutil.SetIfNonEmpty(args, "comment", v.Comment)
	if v.Disabled {
		args["disabled"] = "yes"
	}
	return command.Command{
		Raw:  "/interface/veth/add",
		Args: args,
	}
}

// NewSetVethInterfaceCommand updates a Virtual Ethernet interface.
func NewSetVethInterfaceCommand(rosID string, v VethInterface) command.Command {
	args := map[string]string{
		".id": rosID,
	}
	rosutil.SetIfNonEmpty(args, "name", v.Name)
	rosutil.SetIfNonEmpty(args, "address", v.Address)
	rosutil.SetIfNonEmpty(args, "gateway", v.Gateway)
	args["comment"] = v.Comment
	if v.Disabled {
		args["disabled"] = "yes"
	} else {
		args["disabled"] = "no"
	}
	return command.Command{
		Raw:  "/interface/veth/set",
		Args: args,
	}
}

// NewRemoveVethInterfaceCommand deletes a Virtual Ethernet interface.
func NewRemoveVethInterfaceCommand(rosID string) command.Command {
	return command.Command{
		Raw: "/interface/veth/remove",
		Args: map[string]string{
			".id": rosID,
		},
	}
}

// ParseContainerEnvs converts reply rows into ContainerEnv slice.
func ParseContainerEnvs(rows []map[string]string) []ContainerEnv {
	out := make([]ContainerEnv, 0, len(rows))
	for _, r := range rows {
		out = append(out, ContainerEnv{
			ID:      r[".id"],
			List:    r["list"],
			Key:     r["key"],
			Comment: r["comment"],
		})
	}
	return out
}

// ParseContainerMounts converts reply rows into ContainerMount slice.
func ParseContainerMounts(rows []map[string]string) []ContainerMount {
	out := make([]ContainerMount, 0, len(rows))
	for _, r := range rows {
		out = append(out, ContainerMount{
			ID:      r[".id"],
			Name:    r["name"],
			Src:     r["src"],
			Dst:     r["dst"],
			Comment: r["comment"],
		})
	}
	return out
}

// ParseVethInterfaces converts reply rows into VethInterface slice.
func ParseVethInterfaces(rows []map[string]string) []VethInterface {
	out := make([]VethInterface, 0, len(rows))
	for _, r := range rows {
		out = append(out, VethInterface{
			ID:       r[".id"],
			Name:     r["name"],
			Address:  r["address"],
			Gateway:  r["gateway"],
			Running:  strings.EqualFold(r["running"], "yes") || strings.EqualFold(r["running"], "true"),
			Disabled: strings.EqualFold(r["disabled"], "yes") || strings.EqualFold(r["disabled"], "true"),
			Comment:  r["comment"],
		})
	}
	return out
}
