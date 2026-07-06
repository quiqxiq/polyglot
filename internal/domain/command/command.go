package command

// Command is a vendor-agnostic instruction requested by a user or AI agent.
// Each DeviceDriver interprets Raw/Args according to its own protocol —
// e.g. Mikrotik treats Raw as a RouterOS API path ("/interface/print") and
// Args as sentence attributes; Cisco/scrapligo treats Raw as a literal CLI
// string. See NetOps-Architecture.md §5.3 for the rationale, and CLAUDE.md
// §1.2 for where vendor-specific command knowledge must live (never here).
type Command struct {
	Raw  string
	Args map[string]string
}

// Result is the outcome of executing a Command against a device.
type Result struct {
	Output string
}

// Operation is an abstract, vendor-agnostic operation name that usecase/
// may request (e.g. "get the device's status") without knowing which
// native command fulfills it for any given vendor. DeviceDriver.Translate
// maps an Operation to a concrete Command.
type Operation string

const (
	OpGetStatus Operation = "get_status"
	OpReboot    Operation = "reboot"
)

// Class represents the risk classification of a Command, as determined by
// the target device's specific DeviceDriver implementation (not by domain
// itself — domain never knows vendor command syntax).
type Class int

const (
	ClassReadOnly Class = iota
	ClassDestructive
)
