package system

import (
	"github.com/quixiq/polyglot/internal/port"
)

// SystemResource is the vendor-neutral router system resource snapshot.
type SystemResource = port.SystemResource

// Health holds the parsed sensor data from /system/health/print.
type Health = port.SystemHealth

// Clock holds the parsed clock data from /system/clock/print.
type Clock = port.SystemClock

// Identity holds the router's configured identity name.
type Identity = port.SystemIdentity

// Routerboard holds the parsed data from /system/routerboard/print.
type Routerboard = port.SystemRouterboard

// LogEntry represents one row returned by /log/print.
type LogEntry = port.LogEntry

// PingResult holds one "!re" row from a /ping command.
type PingResult struct {
	Seq        int
	Host       string
	Sent       int
	Received   int
	PacketLoss int
	MinRTT     string
	AvgRTT     string
	MaxRTT     string
	TTL        string
	Time       string
}

// SystemScheduler represents one row from /system/scheduler/print.
type SystemScheduler struct {
	RosID     string
	Name      string
	StartDate string
	StartTime string
	Interval  string
	OnEvent   string
	RunCount  string
	NextRun   string
	Comment   string
	Disabled  bool
}

// SystemSchedulerParams holds parameters for adding or modifying a scheduler entry.
type SystemSchedulerParams struct {
	Name      string
	StartDate string
	StartTime string
	Interval  string
	OnEvent   string
	Comment   string
	Disabled  bool
}

// SystemScript represents one row from /system/script/print.
type SystemScript struct {
	RosID      string
	Name       string
	Source     string
	Owner      string
	RunCount   string
	LastRun    string
	Comment    string
	DontReqPwr bool
}

// SystemScriptParams holds parameters for adding or modifying a script entry.
type SystemScriptParams struct {
	Name       string
	Source     string
	Comment    string
	DontReqPwr bool
}
