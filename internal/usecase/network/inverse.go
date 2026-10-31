package network

import (
	"strings"

	"github.com/quixiq/polyglot/internal/domain/command"
)

// BuiltInverse is a best-effort inverse command plus warnings for cases where
// a perfect inverse is impossible (reboot, move).
type BuiltInverse struct {
	Commands []command.Command
	Warnings []string
}

// InverseVerb maps a mutation verb to its opposite where one exists.
func InverseVerb(verb string) string {
	switch verb {
	case "add":
		return "remove"
	case "remove":
		return "add"
	case "enable":
		return "disable"
	case "disable":
		return "enable"
	default:
		return ""
	}
}

// BuildInverse derives rollback commands for one mutation. createdID is the
// router-returned .id for add commands (from command.Result.Meta["ret"]).
func BuildInverse(cmd command.Command, before []map[string]string, createdID string) BuiltInverse {
	menu, verb := SplitCommandPath(cmd.Raw)
	out := BuiltInverse{}
	switch verb {
	case "add":
		if createdID != "" {
			out.Commands = append(out.Commands, command.Command{Raw: joinMenuVerb(menu, "remove"), Args: map[string]string{".id": createdID}})
			return out
		}
		// Fallback: unique selector from the args themselves (comment is a
		// stable RouterOS field when the caller supplied one).
		selector := uniqueSelector(cmd.Args)
		if selector == nil {
			out.Warnings = append(out.Warnings, "inverse add: no .id returned and no unique selector; rollback manual")
			return out
		}
		out.Commands = append(out.Commands, command.Command{Raw: joinMenuVerb(menu, "remove"), Args: selector})
	case "set":
		if len(before) == 0 {
			out.Warnings = append(out.Warnings, "inverse set: no snapshot; rollback manual")
			return out
		}
		args := map[string]string{}
		if id := selectorID(cmd.Args); id != "" {
			args[".id"] = id
		} else {
			out.Warnings = append(out.Warnings, "inverse set: no .id selector; rollback manual")
			return out
		}
		for k := range cmd.Args {
			if k == ".id" || k == "numbers" || strings.HasPrefix(k, "?") {
				continue
			}
			if old, ok := before[0][k]; ok {
				args[k] = old
			}
		}
		out.Commands = append(out.Commands, command.Command{Raw: joinMenuVerb(menu, "set"), Args: args})
	case "remove":
		if len(before) == 0 {
			out.Warnings = append(out.Warnings, "inverse remove: no snapshot; rollback manual")
			return out
		}
		args := map[string]string{}
		for k, v := range before[0] {
			if k == ".id" || k == "dynamic" || k == "invalid" {
				continue
			}
			args[k] = v
		}
		out.Commands = append(out.Commands, command.Command{Raw: joinMenuVerb(menu, "add"), Args: args})
	case "enable", "disable":
		args := map[string]string{}
		if id := selectorID(cmd.Args); id != "" {
			args[".id"] = id
		} else {
			out.Warnings = append(out.Warnings, "inverse status: no .id selector; rollback manual")
			return out
		}
		out.Commands = append(out.Commands, command.Command{Raw: joinMenuVerb(menu, InverseVerb(verb)), Args: args})
	default:
		out.Warnings = append(out.Warnings, "no deterministic inverse for verb "+verb)
	}
	return out
}

// RenderInverseScript renders inverse commands as a RouterOS import script.
func RenderInverseScript(cmds []command.Command) []byte {
	var b strings.Builder
	b.WriteString("# Polyglot AI Router Ops inverse script (best effort)\n")
	for _, cmd := range cmds {
		b.WriteString(cmd.Raw)
		for _, key := range sortedKeys(cmd.Args) {
			b.WriteString(" ")
			b.WriteString(key)
			b.WriteString("=")
			b.WriteString(quoteRouterValue(cmd.Args[key]))
		}
		b.WriteString("\n")
	}
	return []byte(b.String())
}

func joinMenuVerb(menu, verb string) string {
	menu = strings.Trim(menu, "/")
	if menu == "" {
		return "/" + verb
	}
	return "/" + menu + "/" + verb
}

func selectorID(args map[string]string) string {
	for _, key := range []string{".id", "numbers"} {
		if v := strings.TrimSpace(args[key]); v != "" {
			return v
		}
	}
	return ""
}

func uniqueSelector(args map[string]string) map[string]string {
	for _, key := range []string{"comment", "name", "address"} {
		if v := strings.TrimSpace(args[key]); v != "" {
			out := map[string]string{"?" + key: v}
			return out
		}
	}
	return nil
}

func quoteRouterValue(v string) string {
	if v == "" || strings.ContainsAny(v, " \"'=;") {
		return `"` + strings.ReplaceAll(v, `"`, `\"`) + `"`
	}
	return v
}
