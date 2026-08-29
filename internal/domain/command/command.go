package command

// Command is a vendor-agnostic instruction requested by a user or AI agent.
// Each DeviceDriver interprets Raw/Args according to its own protocol —
// e.g. Mikrotik treats Raw as a RouterOS API path ("/interface/print") and
// Args as sentence attributes; Cisco/scrapligo treats Raw as a literal CLI
// string. See Polyglot-Architecture.md §5.3 for the rationale, and CLAUDE.md
// §1.2 for where vendor-specific command knowledge must live (never here).
type Command struct {
	Raw  string
	Args map[string]string
}

// Result is the outcome of executing a Command against a device.
//
// Rows holds structured key-value data when the vendor's reply is naturally
// tabular (e.g. each RouterOS "!re" sentence's word map for a /print
// command, or one row per streamed event for a streaming command such as
// /ping or /interface/monitor-traffic). A command with no tabular output
// (e.g. a bare action like /system/reboot) leaves Rows empty. A command
// that returns many rows (e.g. /interface/print with no filter) populates
// Rows with one map per row — Result deliberately does NOT flatten this
// into a single map, since that would silently drop rows beyond the first.
type Result struct {
	Output string
	Rows   []map[string]string
}

// Operation is an abstract, vendor-agnostic operation name that usecase/
// may request (e.g. "get the device's status") without knowing which
// native command fulfills it for any given vendor. DeviceDriver.Translate
// maps an Operation to a concrete Command.
type Operation string

const (
	// OpGetStatus requests the current device status.
	OpGetStatus Operation = "get_status"
	// OpReboot requests a device reboot.
	OpReboot Operation = "reboot"
)

// Class represents the risk classification of a Command, as determined by
// the target device's specific DeviceDriver implementation (not by domain
// itself — domain never knows vendor command syntax).
type Class int

const (
	// ClassReadOnly marks a command as non-destructive.
	ClassReadOnly Class = iota
	// ClassDestructive marks a command as potentially destructive.
	ClassDestructive
)
