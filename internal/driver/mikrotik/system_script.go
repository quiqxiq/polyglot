package mikrotik

import (
	"github.com/quixiq/polyglot/internal/domain/command"
)

// SystemScript represents one row from /system/script/print.
type SystemScript struct {
	RosID   string
	Name    string
	Owner   string
	Source  string
	Comment string
}

// NewPrintSystemScriptsCommand builds the command.Command for /system/script/print.
func NewPrintSystemScriptsCommand(ownerFilter, commentFilter string) command.Command {
	args := map[string]string{}
	if ownerFilter != "" {
		args["?owner"] = ownerFilter
	}
	if commentFilter != "" {
		args["?comment"] = commentFilter
	}
	return command.Command{
		Raw:  "/system/script/print",
		Args: args,
	}
}

// NewRemoveSystemScriptCommand builds the command.Command for /system/script/remove.
func NewRemoveSystemScriptCommand(rosID string) command.Command {
	return command.Command{
		Raw:  "/system/script/remove",
		Args: map[string]string{".id": rosID},
	}
}

// ParseSystemScripts converts command.Result rows from /system/script/print.
func ParseSystemScripts(result command.Result) []SystemScript {
	scripts := make([]SystemScript, 0, len(result.Rows))
	for _, row := range result.Rows {
		id := row[".id"]
		name := row["name"]
		if id == "" || name == "" {
			continue
		}
		scripts = append(scripts, SystemScript{
			RosID:   id,
			Name:    name,
			Owner:   row["owner"],
			Source:  row["source"],
			Comment: row["comment"],
		})
	}
	return scripts
}
