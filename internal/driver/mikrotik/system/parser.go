package system

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/quixiq/polyglot/internal/domain/command"
)

// ParseResource converts a command.Result from /system/resource/print into a SystemResource struct.
func ParseResource(result command.Result) SystemResource {
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
		FirmwareType:  row["current-firmware"],
		Voltage:       voltage,
		Temperature:   row["temperature"],
		BadBlocks:     row["bad-blocks"],
		Uptime:        row["uptime"],
		Version:       row["version"],
		BoardName:     row["board-name"],
	}
}

// ParseHealth converts a command.Result from /system/health/print into a Health struct.
func ParseHealth(result command.Result) Health {
	if len(result.Rows) == 0 {
		return Health{}
	}
	row := result.Rows[0]
	return Health{
		Voltage:        row["voltage"],
		Temperature:    row["temperature"],
		CPUTemperature: row["cpu-temperature"],
		PSUVoltage:     row["psu-voltage"],
		PSUCurrent:     row["psu-current"],
		PSUTemperature: row["psu-temperature"],
		Fan1Speed:      row["fan1-speed"],
		Fan2Speed:      row["fan2-speed"],
	}
}

// ParseClock converts command.Result from /system/clock/print into a Clock struct.
func ParseClock(result command.Result) Clock {
	if len(result.Rows) == 0 {
		return Clock{}
	}
	row := result.Rows[0]
	return Clock{
		Time:         row["time"],
		Date:         row["date"],
		TimeZoneName: row["time-zone-name"],
		GMTOffset:    row["gmt-offset"],
		DSTActive:    strings.EqualFold(row["dst-active"], "yes") || strings.EqualFold(row["dst-active"], "true"),
	}
}

// ParseIdentity converts a command.Result from /system/identity/print into a Identity struct.
func ParseIdentity(result command.Result) Identity {
	if len(result.Rows) == 0 {
		return Identity{}
	}
	return Identity{Name: result.Rows[0]["name"]}
}

// ParseRouterboard converts a command.Result from /system/routerboard/print into a Routerboard struct.
func ParseRouterboard(result command.Result) Routerboard {
	if len(result.Rows) == 0 {
		return Routerboard{}
	}
	row := result.Rows[0]
	return Routerboard{
		BoardName:       row["board-name"],
		Model:           row["model"],
		SerialNumber:    row["serial-number"],
		FirmwareType:    row["firmware-type"],
		FactoryFirmware: row["factory-firmware"],
		CurrentFirmware: row["current-firmware"],
		UpgradeFirmware: row["upgrade-firmware"],
	}
}

// ParseLogs converts command.Result rows from /log/print into a slice of LogEntry.
func ParseLogs(result command.Result) []LogEntry {
	logs := make([]LogEntry, 0, len(result.Rows))
	for i, row := range result.Rows {
		msg := row["message"]
		id := row[".id"]
		timeStr := row["time"]
		topics := row["topics"]

		if msg == "" && timeStr == "" && topics == "" {
			continue
		}

		if id == "" {
			id = fmt.Sprintf("log-%d-%d", time.Now().UnixNano(), i)
		}

		logs = append(logs, LogEntry{
			RosID:   id,
			Time:    timeStr,
			Topics:  topics,
			Message: msg,
		})
	}
	return logs
}

// ParsePing converts command.Result rows from /ping into a slice of PingResult.
func ParsePing(result command.Result) []PingResult {
	pings := make([]PingResult, 0, len(result.Rows))
	for _, row := range result.Rows {
		seq, _ := strconv.Atoi(row["seq"])
		sent, _ := strconv.Atoi(row["sent"])
		recv, _ := strconv.Atoi(row["received"])
		loss, _ := strconv.Atoi(row["packet-loss"])

		pings = append(pings, PingResult{
			Seq:        seq,
			Host:       row["host"],
			Sent:       sent,
			Received:   recv,
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

// ParseScheduler converts command.Result rows from /system/scheduler/print.
func ParseScheduler(result command.Result) []SystemScheduler {
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
			StartDate: row["start-date"],
			StartTime: row["start-time"],
			Interval:  row["interval"],
			OnEvent:   row["on-event"],
			RunCount:  row["run-count"],
			NextRun:   row["next-run"],
			Comment:   row["comment"],
			Disabled:  strings.EqualFold(row["disabled"], "true"),
		})
	}
	return schedulers
}

// ParseScript converts command.Result rows from /system/script/print.
func ParseScript(result command.Result) []SystemScript {
	scripts := make([]SystemScript, 0, len(result.Rows))
	for _, row := range result.Rows {
		id := row[".id"]
		name := row["name"]
		if id == "" || name == "" {
			continue
		}
		scripts = append(scripts, SystemScript{
			RosID:      id,
			Name:       name,
			Source:     row["source"],
			Owner:      row["owner"],
			RunCount:   row["run-count"],
			LastRun:    row["last-run"],
			Comment:    row["comment"],
			DontReqPwr: strings.EqualFold(row["dont-require-permissions"], "true") || strings.EqualFold(row["dont-require-permissions"], "yes"),
		})
	}
	return scripts
}
