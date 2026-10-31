package system

import (
	"strconv"
	"strings"

	"github.com/quixiq/polyglot/internal/domain/command"
)

// This file groups the Read/Export/Backup primitives used by the AI Router
// Ops restore-point pipeline (PLAN-AI-ROUTER-OPS v2). It only builds
// vendor-native commands and parses replies; orchestration lives in
// usecase/network.

// NewScopedExportCommand exports one menu subtree to a file on the router.
// scope is a RouterOS menu path without leading slash, e.g. "ip/firewall".
func NewScopedExportCommand(scope, filename string) command.Command {
	raw := "/export"
	scope = strings.Trim(scope, "/")
	if scope != "" {
		raw = "/" + scope + "/export"
	}
	return command.Command{
		Raw: raw,
		Args: map[string]string{
			"file":           filename,
			"compact":        "",
			"terse":          "",
			"hide-sensitive": "",
		},
	}
}

// NewFullExportCommand exports the whole configuration (secrets hidden).
func NewFullExportCommand(filename string) command.Command {
	return NewScopedExportCommand("", filename)
}

// NewResourcePrintCommand reads version/free space/uptime for preflight.
func NewResourcePrintCommand() command.Command {
	return command.Command{
		Raw:  "/system/resource/print",
		Args: map[string]string{".proplist": "version,board-name,uptime,free-hdd-space,free-memory"},
	}
}

// NewClockPrintCommand reads the router clock (watchdog uses router time).
func NewClockPrintCommand() command.Command {
	return command.Command{
		Raw:  "/system/clock/print",
		Args: map[string]string{".proplist": "date,time,time-zone-name"},
	}
}

// ParseReadChunk extracts file bytes from a /file/read reply. RouterOS
// returns the chunk in the "contents" attribute; older builds may use "data".
// It returns ok=false when no content attribute is present.
func ParseReadChunk(result command.Result) (data []byte, ok bool) {
	for _, row := range result.Rows {
		for _, key := range []string{"contents", "data", "text"} {
			if v, exists := row[key]; exists {
				return []byte(v), true
			}
		}
	}
	if result.Meta != nil {
		for _, key := range []string{"contents", "data", "text"} {
			if v, exists := result.Meta[key]; exists {
				return []byte(v), true
			}
		}
	}
	return nil, false
}

// ParseRetID extracts the created object id (.id) from an add command reply
// (RouterOS returns it in the !done sentence as "ret" or ".id").
func ParseRetID(result command.Result) string {
	if result.Meta != nil {
		for _, key := range []string{"ret", ".id", "id"} {
			if v := strings.TrimSpace(result.Meta[key]); v != "" {
				return v
			}
		}
	}
	return ""
}

// ParseFileSize returns the "size" field of the first /file/print row.
func ParseFileSize(result command.Result) int64 {
	if len(result.Rows) == 0 {
		return 0
	}
	raw := strings.ReplaceAll(strings.TrimSpace(result.Rows[0]["size"]), " ", "")
	n, _ := strconv.ParseInt(raw, 10, 64)
	return n
}
