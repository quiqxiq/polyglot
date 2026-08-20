package port

// ExpireMonitorStatus describes the install/enabled state of the Mikhmon
// expire-monitor scheduler on a device. It is produced by
// HotspotGateway.GetExpireMonitorStatus and consumed by the usecase/handler
// layers; the raw RouterOS .id is kept so disable/remove can act on the
// exact scheduler that was found.
type ExpireMonitorStatus struct {
	IsInstalled   bool
	IsEnabled     bool
	SchedulerID   string // RouterOS .id — used for disable/remove
	SchedulerName string // name as found on the router (legacy or gateway form)
}
