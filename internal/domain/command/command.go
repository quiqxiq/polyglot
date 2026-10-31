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
//
// Meta carries vendor reply metadata that is not tabular data — for RouterOS
// this is the "!done" sentence map (e.g. ret/.id returned by an add
// command). It is used by the AI Router Ops approval pipeline to derive the
// inverse of an "add" without relying on a synthetic comment marker.
type Result struct {
	Output string
	Rows   []map[string]string
	Meta   map[string]string
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

// Mode is the only risk classification the AI Router Ops policy
// distinguishes: a command either reads/introspects state or it can mutate
// state. Unknown commands MUST be treated as ModeMutation (fail-closed) —
// see Polyglot-Architecture.md §6.1 and PLAN-AI-ROUTER-OPS v2.
type Mode int

const (
	// ModeRead covers commands that only read state or introspect syntax.
	// Safe to auto-execute for AI and humans.
	ModeRead Mode = iota
	// ModeMutation covers any command that can create, change, delete,
	// enable/disable, reset, reboot, upload, or otherwise alter device state.
	ModeMutation
)

// Classification is the vendor driver's result for one Command: the mode plus
// an optional hard-deny reason. A non-empty DenyReason means the command must
// never be executed by the AI/proposal pipeline (e.g. reset-configuration,
// backup/load, AI-created scripts). Human restore flows use dedicated
// usecases with typed confirmation instead of this pipeline.
type Classification struct {
	Mode       Mode
	DenyReason string
}

// IsDenied reports whether the command is hard-denied.
func (c Classification) IsDenied() bool { return c.DenyReason != "" }
