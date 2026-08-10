package mikrotik

import (
	"strconv"
	"strings"

	"github.com/quixiq/polyglot/internal/domain/command"
)

// SystemResource holds the parsed data from /system/resource/print.
// All numeric fields are kept as raw strings from RouterOS — unit conversion
// (bytes → MB, Hz → GHz, etc.) is the caller's responsibility, since
// different display contexts need different formats.
//
// Field notes (from RouterOS /system/resource reference):
//   - CPULoad      : current CPU utilization as percentage (0–100).
//   - CPUCount     : number of CPU cores.
//   - CPUFrequency : CPU clock in MHz.
//   - FreeMemory   : free RAM in bytes.
//   - TotalMemory  : total RAM in bytes.
//   - FreeHDDSpace : free storage in bytes.
//   - TotalHDDSpace: total storage in bytes.
//   - Architecture : CPU architecture name (e.g. "mipsbe", "arm", "x86").
//   - Model        : hardware model string.
//   - SerialNumber : device serial number.
//   - FirmwareType : firmware variant.
//   - Voltage      : board voltage in millivolts (may be empty on some models).
//   - Temperature  : board temperature in Celsius (may be empty).
//   - BadBlocks    : number of bad storage blocks (flash health).
//   - Uptime       : system uptime as RouterOS time string (e.g. "3d2h10m5s").
//   - Version      : RouterOS version string (e.g. "7.10.2 (stable)").
//   - BoardName    : board model name.
type SystemResource struct {
	CPULoad       int
	CPUCount      int
	CPUFrequency  string // MHz
	FreeMemory    string // bytes
	TotalMemory   string // bytes
	FreeHDDSpace  string // bytes
	TotalHDDSpace string // bytes
	Architecture  string
	Model         string
	SerialNumber  string
	FirmwareType  string
	Voltage       string // millivolts, may be empty
	Temperature   string // Celsius, may be empty
	BadBlocks     string
	Uptime        string
	Version       string
	BoardName     string
}

// SystemIdentity holds the router's configured identity (name).
type SystemIdentity struct {
	Name string
}

// SystemClock holds the current clock state from /system/clock/print.
type SystemClock struct {
	Time         string
	Date         string
	TimeZoneName string
	GMTOffset    string
}

// LogEntry represents one row returned by /log/print.
//
// RouterOS logs are append-only — there is no add/set/remove for log entries.
type LogEntry struct {
	RosID   string
	Time    string
	Topics  string // comma-separated categories (e.g. "ppp,info")
	Message string
}

// PingResult holds one "!re" row from a /ping command.
//
// Field notes (from RouterOS /ping reference):
//   - Seq         : sequence number (0-based).
//   - Host        : target hostname or IP.
//   - Sent        : packets sent so far.
//   - Received    : packets received.
//   - PacketLoss  : percentage (0–100).
//   - MinRTT      : minimum round-trip time (RouterOS duration string, e.g. "1ms").
//   - AvgRTT      : average round-trip time.
//   - MaxRTT      : maximum round-trip time.
//   - TTL         : time-to-live of received packet.
//   - Time        : RTT for this specific packet.
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

// NewPrintSystemResourceCommand builds the command.Command for
// /system/resource/print (one-shot, returns current resource snapshot).
func NewPrintSystemResourceCommand() command.Command {
	return command.Command{
		Raw:  "/system/resource/print",
		Args: map[string]string{},
	}
}

// NewStreamSystemResourceCommand builds the command.Command for
// /system/resource/print with interval (streaming, pushes system resource
// updates periodically, e.g. CPU load, memory usage). Use with Driver.Stream.
// Pass empty string for default interval "1s".
func NewStreamSystemResourceCommand(interval string) command.Command {
	if interval == "" {
		interval = "1s"
	}
	return command.Command{
		Raw:  "/system/resource/print",
		Args: map[string]string{"interval": interval},
	}
}


// NewPrintSystemIdentityCommand builds the command.Command for
// /system/identity/print (get router hostname).
func NewPrintSystemIdentityCommand() command.Command {
	return command.Command{
		Raw:  "/system/identity/print",
		Args: map[string]string{},
	}
}

// NewSetSystemIdentityCommand builds the command.Command for
// /system/identity/set (rename the router).
func NewSetSystemIdentityCommand(name string) command.Command {
	return command.Command{
		Raw:  "/system/identity/set",
		Args: map[string]string{"name": name},
	}
}

// NewPrintSystemClockCommand builds the command.Command for
// /system/clock/print (get router date/time and timezone).
func NewPrintSystemClockCommand() command.Command {
	return command.Command{
		Raw:  "/system/clock/print",
		Args: map[string]string{},
	}
}

// NewPrintLogCommand builds the command.Command for /log/print.
// Pass a non-empty topicsFilter to match entries containing a keyword in
// their topics field (RouterOS partial match via ?topics~<keyword>).
// Example: topicsFilter="ppp" returns all PPPoE login/logout events.
func NewPrintLogCommand(topicsFilter string) command.Command {
	args := map[string]string{}
	if topicsFilter != "" {
		// RouterOS partial-match operator: ?field~value
		args["?topics~"+topicsFilter] = ""
	}
	return command.Command{
		Raw:  "/log/print",
		Args: args,
	}
}

// NewPingCommand builds the command.Command for /ping.
// count is the number of ICMP packets to send; defaults to "4" if empty.
//
// /ping is a streaming command — execute via Driver.Stream, not Driver.Execute.
// The driver's isStreamingCommand function recognises "/ping" and Driver.Execute
// will return ErrStreamingCommand if you call Execute on it directly.
func NewPingCommand(host, count string) command.Command {
	if count == "" {
		count = "4"
	}
	return command.Command{
		Raw: "/ping",
		Args: map[string]string{
			"address": host,
			"count":   count,
		},
	}
}

// NewPingStreamCommand builds the command.Command for continuous /ping streaming.
// Without the "count" argument, RouterOS sends streaming ping responses continuously
// until cancelled via StreamHandle.Cancel.
func NewPingStreamCommand(host string) command.Command {
	return command.Command{
		Raw: "/ping",
		Args: map[string]string{
			"address": host,
		},
	}
}

// ParseSystemResource converts the first row from /system/resource/print
// into a typed SystemResource. Returns zero-value if result is empty.
func ParseSystemResource(result command.Result) SystemResource {
	if len(result.Rows) == 0 {
		return SystemResource{}
	}
	row := result.Rows[0]

	cpuLoad, _ := strconv.Atoi(row["cpu-load"])
	cpuCount, _ := strconv.Atoi(row["cpu-count"])

	// RouterOS may report voltage/temperature under different field names
	// depending on the hardware model — try both variants.
	voltage := row["voltage"]
	if voltage == "" {
		voltage = row["board-voltage"]
	}
	temperature := row["temperature"]
	if temperature == "" {
		temperature = row["board-temperature"]
	}

	return SystemResource{
		CPULoad:       cpuLoad,
		CPUCount:      cpuCount,
		CPUFrequency:  row["cpu-frequency"],
		FreeMemory:    row["free-memory"],
		TotalMemory:   row["total-memory"],
		FreeHDDSpace:  row["free-hdd-space"],
		TotalHDDSpace: row["total-hdd-space"],
		Architecture:  row["architecture-name"],
		Model:         row["model"],
		SerialNumber:  row["serial-number"],
		FirmwareType:  row["firmware-type"],
		Voltage:       voltage,
		Temperature:   temperature,
		BadBlocks:     row["bad-blocks"],
		Uptime:        row["uptime"],
		Version:       row["version"],
		BoardName:     row["board-name"],
	}
}

// ParseSystemIdentity converts the first row from /system/identity/print.
func ParseSystemIdentity(result command.Result) SystemIdentity {
	if len(result.Rows) == 0 {
		return SystemIdentity{}
	}
	return SystemIdentity{Name: result.Rows[0]["name"]}
}

// ParseSystemClock converts the first row from /system/clock/print.
func ParseSystemClock(result command.Result) SystemClock {
	if len(result.Rows) == 0 {
		return SystemClock{}
	}
	row := result.Rows[0]
	return SystemClock{
		Time:         row["time"],
		Date:         row["date"],
		TimeZoneName: row["time-zone-name"],
		GMTOffset:    row["gmt-offset"],
	}
}

// ParseLogEntries converts command.Result rows from /log/print into typed
// LogEntry values. Rows missing ".id" are skipped.
func ParseLogEntries(result command.Result) []LogEntry {
	entries := make([]LogEntry, 0, len(result.Rows))
	for _, row := range result.Rows {
		id := row[".id"]
		if id == "" {
			continue
		}
		entries = append(entries, LogEntry{
			RosID:   id,
			Time:    row["time"],
			Topics:  row["topics"],
			Message: row["message"],
		})
	}
	return entries
}

// ParsePingResults converts command.Result rows from a /ping streaming result
// into typed PingResult values. Each row corresponds to one ICMP reply.
func ParsePingResults(result command.Result) []PingResult {
	pings := make([]PingResult, 0, len(result.Rows))
	for _, row := range result.Rows {
		seq, _ := strconv.Atoi(row["seq"])
		sent, _ := strconv.Atoi(row["sent"])
		received, _ := strconv.Atoi(row["received"])
		loss, _ := strconv.Atoi(strings.TrimSuffix(row["packet-loss"], "%"))
		pings = append(pings, PingResult{
			Seq:        seq,
			Host:       row["host"],
			Sent:       sent,
			Received:   received,
			PacketLoss: loss,
			MinRTT:     row["min-rtt"],
			AvgRTT:     row["avg-rtt"],
			MaxRTT:     row["max-rtt"],
			TTL:        row["ttl"],
			Time:       row["time"],
		})
	}
	return pings
}

// SystemSchedulerParams holds parameters for /system/scheduler/add or /set.
type SystemSchedulerParams struct {
	Name      string
	StartTime string
	StartDate string
	Interval  string
	OnEvent   string
	Comment   string
	Disabled  bool
}

// SystemScheduler represents one row from /system/scheduler/print.
type SystemScheduler struct {
	RosID     string
	Name      string
	StartTime string
	StartDate string
	Interval  string
	OnEvent   string
	NextRun   string
	Comment   string
	Disabled  bool
}

// SystemScript represents one row from /system/script/print.
type SystemScript struct {
	RosID   string
	Name    string
	Owner   string
	Source  string
	Comment string
}

// NewPrintSystemSchedulersCommand builds the command.Command for /system/scheduler/print.
func NewPrintSystemSchedulersCommand(nameFilter string) command.Command {
	args := map[string]string{}
	if nameFilter != "" {
		args["?name"] = nameFilter
	}
	return command.Command{
		Raw:  "/system/scheduler/print",
		Args: args,
	}
}

// NewAddSystemSchedulerCommand builds the command.Command for /system/scheduler/add.
func NewAddSystemSchedulerCommand(p SystemSchedulerParams) command.Command {
	args := map[string]string{
		"name":     p.Name,
		"on-event": p.OnEvent,
	}
	setIfNonEmpty(args, "start-time", p.StartTime)
	setIfNonEmpty(args, "start-date", p.StartDate)
	setIfNonEmpty(args, "interval", p.Interval)
	setIfNonEmpty(args, "comment", p.Comment)
	if p.Disabled {
		args["disabled"] = "yes"
	} else {
		args["disabled"] = "no"
	}
	return command.Command{Raw: "/system/scheduler/add", Args: args}
}

// NewSetSystemSchedulerCommand builds the command.Command for /system/scheduler/set.
func NewSetSystemSchedulerCommand(rosID string, p SystemSchedulerParams) command.Command {
	args := map[string]string{".id": rosID}
	setIfNonEmpty(args, "name", p.Name)
	setIfNonEmpty(args, "start-time", p.StartTime)
	setIfNonEmpty(args, "start-date", p.StartDate)
	setIfNonEmpty(args, "interval", p.Interval)
	setIfNonEmpty(args, "on-event", p.OnEvent)
	setIfNonEmpty(args, "comment", p.Comment)
	return command.Command{Raw: "/system/scheduler/set", Args: args}
}

// ParseSystemSchedulers converts command.Result rows from /system/scheduler/print.
func ParseSystemSchedulers(result command.Result) []SystemScheduler {
	schedulers := make([]SystemScheduler, 0, len(result.Rows))
	for _, row := range result.Rows {
		id := row[".id"]
		name := row["name"]
		if id == "" || name == "" {
			continue
		}
		schedulers = append(schedulers, SystemScheduler{
			RosID:     id,
			Name:      name,
			StartTime: row["start-time"],
			StartDate: row["start-date"],
			Interval:  row["interval"],
			OnEvent:   row["on-event"],
			NextRun:   row["next-run"],
			Comment:   row["comment"],
			Disabled:  strings.EqualFold(row["disabled"], "true"),
		})
	}
	return schedulers
}

// NewPrintSystemScriptsCommand builds the command.Command for /system/script/print.
// Pass ownerFilter or commentFilter to filter entries (e.g. commentFilter="mikhmon").
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

