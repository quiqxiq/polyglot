package network

import (
	"sort"
	"strings"

	"github.com/quixiq/polyglot/internal/domain/command"
)

// DiffLine is one semantic change in a proposed mutation.
type DiffLine struct {
	Op     string `json:"op"` // context | add | remove | change
	Path   string `json:"path"`
	Before string `json:"before,omitempty"`
	After  string `json:"after,omitempty"`
	Text   string `json:"text"`
}

// Diff is the human-reviewable change set for one command.
type Diff struct {
	Menu    string            `json:"menu"`
	Verb    string            `json:"verb"`
	Summary string            `json:"summary"`
	Lines   []DiffLine        `json:"lines"`
	Before  map[string]string `json:"before,omitempty"`
	After   map[string]string `json:"after,omitempty"`
}

// CommandDiff pairs a proposed command with its diff (multi-command config).
type CommandDiff struct {
	Command string `json:"command"`
	Diff    Diff   `json:"diff"`
}

// SplitCommandPath splits "/ip/firewall/filter/add" into menu and verb.
func SplitCommandPath(raw string) (menu, verb string) {
	trimmed := strings.Trim(strings.TrimSpace(raw), "/")
	if trimmed == "" {
		return "", ""
	}
	if i := strings.LastIndex(trimmed, "/"); i >= 0 {
		return trimmed[:i], trimmed[i+1:]
	}
	return "", trimmed
}

// BuildCommandDiff builds a reviewable diff for cmd given a pre-mutation
// snapshot of the affected menu (before may be nil for add).
func BuildCommandDiff(cmd command.Command, before []map[string]string) Diff {
	menu, verb := SplitCommandPath(cmd.Raw)
	d := Diff{Menu: menu, Verb: verb, Summary: cmd.Raw}
	if len(before) > 0 {
		d.Before = before[0]
	}

	switch verb {
	case "add":
		d.After = cloneArgs(cmd.Args)
		d.Summary = "Tambah objek di /" + menu
		for _, key := range sortedKeys(cmd.Args) {
			d.Lines = append(d.Lines, DiffLine{Op: "add", Path: key, After: cmd.Args[key], Text: "+ " + key + "=" + cmd.Args[key]})
		}
	case "set":
		if len(before) > 0 {
			d.Before = before[0]
		}
		d.After = mergedAfter(before, cmd.Args)
		d.Summary = "Ubah objek di /" + menu
		for _, key := range sortedKeys(cmd.Args) {
			if key == ".id" || key == "numbers" || strings.HasPrefix(key, "?") {
				continue
			}
			oldVal := ""
			if len(before) > 0 {
				oldVal = before[0][key]
			}
			if oldVal == cmd.Args[key] {
				continue
			}
			d.Lines = append(d.Lines, DiffLine{Op: "change", Path: key, Before: oldVal, After: cmd.Args[key], Text: "~ " + key + ": " + oldVal + " → " + cmd.Args[key]})
		}
	case "remove":
		d.Summary = "Hapus objek di /" + menu
		if len(before) > 0 {
			for _, key := range sortedKeys(before[0]) {
				if key == ".id" || key == "dynamic" {
					continue
				}
				d.Lines = append(d.Lines, DiffLine{Op: "remove", Path: key, Before: before[0][key], Text: "- " + key + "=" + before[0][key]})
			}
		}
	case "enable", "disable":
		d.Summary = "Ubah status objek di /" + menu
		d.Lines = append(d.Lines, DiffLine{Op: "change", Path: "disabled", Before: verb, After: verb, Text: "~ " + verb})
	case "move":
		d.Summary = "Pindahkan objek di /" + menu
		d.Lines = append(d.Lines, DiffLine{Op: "change", Path: "position", Before: cmd.Args["numbers"], After: cmd.Args["destination"], Text: "~ move " + cmd.Args["numbers"] + " → " + cmd.Args["destination"]})
	default:
		d.Summary = "Aksi /" + menu + "/" + verb
		for _, key := range sortedKeys(cmd.Args) {
			d.Lines = append(d.Lines, DiffLine{Op: "context", Path: key, After: cmd.Args[key], Text: "  " + key + "=" + cmd.Args[key]})
		}
	}
	return d
}

// BuildScriptDiff builds one combined diff for a multi-command script.
func BuildScriptDiff(commands []command.Command, snapshots [][]map[string]string) []CommandDiff {
	out := make([]CommandDiff, len(commands))
	for i, cmd := range commands {
		var before []map[string]string
		if i < len(snapshots) {
			before = snapshots[i]
		}
		out[i] = CommandDiff{Command: cmd.Raw, Diff: BuildCommandDiff(cmd, before)}
	}
	return out
}

func cloneArgs(args map[string]string) map[string]string {
	out := make(map[string]string, len(args))
	for k, v := range args {
		if k == ".id" || k == "numbers" || strings.HasPrefix(k, "?") {
			continue
		}
		out[k] = v
	}
	return out
}

func mergedAfter(before []map[string]string, args map[string]string) map[string]string {
	out := map[string]string{}
	if len(before) > 0 {
		for k, v := range before[0] {
			out[k] = v
		}
	}
	for k, v := range args {
		out[k] = v
	}
	return out
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
