package iface

import (
	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/driver/mikrotik/internal/rosutil"
)

// InterfaceList represents an interface list definition.
type InterfaceList struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Comment string `json:"comment"`
	Builtin bool   `json:"builtin"`
}

// InterfaceListMember represents an interface membership in a list.
type InterfaceListMember struct {
	ID        string `json:"id"`
	Interface string `json:"interface"`
	List      string `json:"list"`
	Disabled  bool   `json:"disabled"`
	Comment   string `json:"comment"`
}

// NewPrintInterfaceListsCommand queries interface lists with strict proplist.
func NewPrintInterfaceListsCommand() command.Command {
	args := make(map[string]string)
	rosutil.SetProplist(args, ".id", "name", "comment", "builtin")
	return command.Command{
		Raw:  "/interface/list/print",
		Args: args,
	}
}

// NewAddInterfaceListCommand creates an interface list.
func NewAddInterfaceListCommand(name, comment string) command.Command {
	args := map[string]string{
		"name": name,
	}
	if comment != "" {
		args["comment"] = comment
	}
	return command.Command{
		Raw:  "/interface/list/add",
		Args: args,
	}
}

// NewRemoveInterfaceListCommand deletes an interface list by ID.
func NewRemoveInterfaceListCommand(id string) command.Command {
	return command.Command{
		Raw: "/interface/list/remove",
		Args: map[string]string{
			".id": id,
		},
	}
}

// NewPrintInterfaceListMembersCommand queries members of interface lists with strict proplist.
func NewPrintInterfaceListMembersCommand() command.Command {
	args := make(map[string]string)
	rosutil.SetProplist(args, ".id", "interface", "list", "disabled", "comment")
	return command.Command{
		Raw:  "/interface/list/member/print",
		Args: args,
	}
}

// NewAddInterfaceListMemberCommand assigns an interface to an interface list.
func NewAddInterfaceListMemberCommand(iface, list, comment string) command.Command {
	args := map[string]string{
		"interface": iface,
		"list":      list,
	}
	if comment != "" {
		args["comment"] = comment
	}
	return command.Command{
		Raw:  "/interface/list/member/add",
		Args: args,
	}
}

// NewRemoveInterfaceListMemberCommand removes an interface from a list by ID.
func NewRemoveInterfaceListMemberCommand(id string) command.Command {
	return command.Command{
		Raw: "/interface/list/member/remove",
		Args: map[string]string{
			".id": id,
		},
	}
}

// ParseInterfaceLists converts reply rows into InterfaceList slice.
func ParseInterfaceLists(rows []map[string]string) []InterfaceList {
	out := make([]InterfaceList, 0, len(rows))
	for _, r := range rows {
		out = append(out, InterfaceList{
			ID:      r[".id"],
			Name:    r["name"],
			Comment: r["comment"],
			Builtin: r["builtin"] == "true" || r["builtin"] == "yes",
		})
	}
	return out
}

// ParseInterfaceListMembers converts reply rows into InterfaceListMember slice.
func ParseInterfaceListMembers(rows []map[string]string) []InterfaceListMember {
	out := make([]InterfaceListMember, 0, len(rows))
	for _, r := range rows {
		out = append(out, InterfaceListMember{
			ID:        r[".id"],
			Interface: r["interface"],
			List:      r["list"],
			Disabled:  r["disabled"] == "true",
			Comment:   r["comment"],
		})
	}
	return out
}
