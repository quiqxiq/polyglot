// Package monitor defines abstract, vendor-neutral monitoring intents. A
// monitoring Operation describes WHAT live observation to open on a device
// (e.g. "stream the hotspot host table", "stream one interface's traffic")
// without naming any vendor's concrete command syntax — that translation lives
// in each vendor's driver (see port.StreamingDeviceDriver and
// internal/driver/<vendor>/commands.go).
//
// Every Operation here is inherently STREAMING and read-only: it is fulfilled
// through port.StreamingDeviceDriver.Stream, not DeviceDriver.Execute, and it
// only observes device state, so it never passes through the command policy
// gate (consistent with how /ping is treated). It is a sibling of
// domain/provision, kept separate because the semantics differ: provisioning
// changes/registers configuration, monitoring opens a live observation stream.
//
// This package is domain-pure per AGENTS.md §1.2: no I/O, no external imports,
// only data.
package monitor

// Operation is an abstract, vendor-neutral monitoring intent. A vendor's
// port.StreamingDeviceDriver.TranslateStream maps it to that vendor's native
// streaming command. The interface is sealed by the unexported marker method so
// only this package can define valid operations — a driver's exhaustive type
// switch can rely on the set being closed. Every Operation maps to exactly one
// native command (streaming is a single long-running command, never a
// sequence), which is why TranslateStream returns one command rather than a
// slice.
type Operation interface {
	isMonitorOperation()
}

// HotspotHosts streams the device's hotspot host table as it changes (RouterOS
// "/ip/hotspot/host/print follow"). Fields optionally restricts which columns
// the device returns (mapped by the driver to its native field-projection
// mechanism, e.g. RouterOS .proplist); an empty Fields means "return every
// field, raw".
type HotspotHosts struct {
	Fields []string
}

func (HotspotHosts) isMonitorOperation() {}

// InterfaceTraffic streams live rx/tx traffic counters for a single interface
// (RouterOS "/interface/monitor-traffic"). Interface is the interface name to
// monitor and is required; the driver rejects the operation if it is empty.
type InterfaceTraffic struct {
	Interface string
}

func (InterfaceTraffic) isMonitorOperation() {}

// DHCPLeases streams DHCP lease changes as they happen (RouterOS
// "/ip/dhcp-server/lease/print follow-only"). Unlike follow, follow-only emits
// only subsequent changes, not the current table first. Fields behaves exactly
// like HotspotHosts.Fields: empty means "return every field, raw".
type DHCPLeases struct {
	Fields []string
}

func (DHCPLeases) isMonitorOperation() {}

// QueueStats streams live statistics for one simple queue (RouterOS
// "/queue/simple/print stats follow" filtered by queue name). QueueName selects
// the queue and is required; the driver rejects the operation if it is empty.
// Fields behaves exactly like HotspotHosts.Fields: empty means "return every
// field, raw".
type QueueStats struct {
	QueueName string
	Fields    []string
}

func (QueueStats) isMonitorOperation() {}
