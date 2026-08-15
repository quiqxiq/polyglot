package mikrotik

import (
	"strconv"
	"strings"

	"github.com/quixiq/polyglot/internal/domain/command"
)

// ParseSystemResource converts the first row from /system/resource/print into a typed SystemResource.
func ParseSystemResource(result command.Result) SystemResource {
	if len(result.Rows) == 0 {
		return SystemResource{}
	}
	row := result.Rows[0]

	cpuLoad, _ := strconv.Atoi(row["cpu-load"])
	cpuCount, _ := strconv.Atoi(row["cpu-count"])

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

// ParseLogEntries converts command.Result rows from /log/print into typed LogEntry values.
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

// ParsePingResults converts command.Result rows from a /ping streaming result into typed PingResult values.
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
