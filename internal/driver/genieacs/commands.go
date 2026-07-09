package genieacs

import (
	"encoding/json"
	"fmt"

	"github.com/quixiq/polyglot/internal/domain/command"
)

// destructiveTaskTypes lists GenieACS task types that mutate the CPE/ONT and
// therefore always require HITL approval. Read-only task types
// (getParameterValues, refreshObject) and the direct-GET pseudo-command
// "getDevice" are absent from this set. Mirrors the official task catalog in
// analisis-api-genieacs.md §3.
var destructiveTaskTypes = map[string]bool{
	"reboot":             true,
	"factoryReset":       true,
	"setParameterValues": true,
	"addObject":          true,
	"deleteObject":       true,
	"download":           true,
}

// directGetCommands lists cmd.Raw values that bypass the task queue entirely
// and map to a direct NBI GET instead of POST /devices/{id}/tasks. Currently
// only "getDevice" (GET /devices/{id}) — added because OpGetStatus should be
// a fast cached read, not an async CPE query (see analisis-api-genieacs.md
// §8 note 2 on the 200/202 split and §3 on getParameterValues storing values
// in DB rather than returning them inline).
var directGetCommands = map[string]bool{
	"getDevice": true,
}

// operationMap translates abstract Operations to GenieACS-native commands.
// OpGetStatus maps to "getDevice" (direct GET, not a task) deliberately —
// see directGetCommands' doc comment.
var operationMap = map[command.Operation]command.Command{
	command.OpGetStatus: {Raw: "getDevice"},
	command.OpReboot:    {Raw: "reboot"},
}

// Classify reports the risk class of cmd according to GenieACS task
// conventions. Direct-GET commands and read-only task types are
// ClassReadOnly; everything else in destructiveTaskTypes is
// ClassDestructive. Unknown Raw values default to read-only — the driver's
// Execute refuses unknown names separately via ErrUnsupportedTaskType, so a
// misclassified unknown can never actually reach the device.
func Classify(cmd command.Command) command.Class {
	if destructiveTaskTypes[cmd.Raw] {
		return command.ClassDestructive
	}
	return command.ClassReadOnly
}

// Translate maps an abstract Operation to a GenieACS-native Command.
func Translate(op command.Operation) (command.Command, error) {
	cmd, ok := operationMap[op]
	if !ok {
		return command.Command{}, fmt.Errorf("genieacs: unsupported operation %q", op)
	}
	return cmd, nil
}

// isTaskCommand reports whether cmd.Raw is a GenieACS task type that must be
// POSTed to /devices/{id}/tasks (as opposed to a direct-GET pseudo-command
// like "getDevice"). Returns false for unknown names — Execute handles
// rejection via ErrUnsupportedTaskType.
func isTaskCommand(cmd command.Command) bool {
	return !directGetCommands[cmd.Raw] && taskTypeKnown(cmd.Raw)
}

// taskTypeKnown reports whether name is one of the eight official GenieACS
// task types from analisis-api-genieacs.md §3.
func taskTypeKnown(name string) bool {
	switch name {
	case "getParameterValues", "refreshObject", "setParameterValues",
		"addObject", "deleteObject", "reboot", "factoryReset", "download":
		return true
	}
	return false
}

// taskBody is the JSON shape POSTed to /devices/{id}/tasks. Fields are
// optional per task type: only "name" is always present; the rest are added
// conditionally by buildTaskBody based on cmd.Args.
//
// ParameterValues uses [][]any — the value slot of each [name, value, type]
// tuple is genuinely dynamic at the JSON boundary (string, number, or
// boolean depending on the TR-069 xsd type), so a typed struct cannot
// represent it without custom marshalling. This is the JSON-deserialization
// exception permitted by CLAUDE.md §6.
type taskBody struct {
	Name            string   `json:"name"`
	ParameterNames  []string `json:"parameterNames,omitempty"`
	ParameterValues [][]any  `json:"parameterValues,omitempty"`
	ObjectName      string   `json:"objectName,omitempty"`
	File            string   `json:"file,omitempty"`
}

// buildTaskBody constructs the JSON body for a POST /devices/{id}/tasks
// request from cmd. cmd.Raw becomes the task "name"; cmd.Args carries the
// task-specific fields, with array/tuple values encoded as JSON strings
// (per the agreed Args encoding — CLAUDE.md limits Args to map[string]string
// and GenieACS needs arrays, so JSON-in-a-string is the least-lossy option):
//
//   - Args["parameterNames"]   : JSON array of strings
//   - Args["parameterValues"]  : JSON array of [name, value, xsd_type] tuples
//   - Args["objectName"]       : plain string
//   - Args["file"]             : plain string
//
// Fields absent from Args are omitted from the body (GenieACS treats missing
// optional fields correctly per task type — e.g. reboot needs none).
func buildTaskBody(cmd command.Command) ([]byte, error) {
	body := taskBody{Name: cmd.Raw}

	if raw, ok := cmd.Args["parameterNames"]; ok {
		if err := json.Unmarshal([]byte(raw), &body.ParameterNames); err != nil {
			return nil, fmt.Errorf("genieacs: parse parameterNames: %w", err)
		}
	}

	if raw, ok := cmd.Args["parameterValues"]; ok {
		if err := json.Unmarshal([]byte(raw), &body.ParameterValues); err != nil {
			return nil, fmt.Errorf("genieacs: parse parameterValues: %w", err)
		}
	}

	body.ObjectName = cmd.Args["objectName"]
	body.File = cmd.Args["file"]

	b, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("genieacs: marshal task body: %w", err)
	}
	return b, nil
}
