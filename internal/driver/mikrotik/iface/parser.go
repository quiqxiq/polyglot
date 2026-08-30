package iface

import (
	"strings"

	"github.com/quixiq/polyglot/internal/domain/command"
)

func parseRosBool(val string) bool {
	return strings.EqualFold(val, "true") || strings.EqualFold(val, "yes") || val == "1"
}

func getRowField(row map[string]string, keys ...string) string {
	for _, k := range keys {
		if val, ok := row[k]; ok && val != "" {
			return val
		}
	}
	return ""
}

// ParseInterfaces converts command.Result rows from /interface/print into typed Interface values.
func ParseInterfaces(result command.Result) []Interface {
	ifaces := make([]Interface, 0, len(result.Rows))
	for _, row := range result.Rows {
		id := row[".id"]
		name := row["name"]
		if id == "" || name == "" {
			continue
		}
		ifaces = append(ifaces, Interface{
			RosID:      id,
			Name:       name,
			Type:       row["type"],
			MTU:        row["mtu"],
			ActualMTU:  row["actual-mtu"],
			L2MTU:      row["l2mtu"],
			MACAddress: row["mac-address"],
			Running:    parseRosBool(row["running"]),
			Disabled:   parseRosBool(row["disabled"]),
			RxByte:     row["rx-byte"],
			TxByte:     row["tx-byte"],
			RxPacket:   row["rx-packet"],
			TxPacket:   row["tx-packet"],
			Comment:    row["comment"],
		})
	}
	return ifaces
}

// ParseInterfaceTrafficStats converts the first row from a /interface/monitor-traffic result into InterfaceTrafficStats.
func ParseInterfaceTrafficStats(result command.Result) InterfaceTrafficStats {
	if len(result.Rows) == 0 {
		return InterfaceTrafficStats{}
	}
	row := result.Rows[0]
	return InterfaceTrafficStats{
		RxBitsPerSecond:    getRowField(row, "rx-bits-per-second", "rx-bps"),
		TxBitsPerSecond:    getRowField(row, "tx-bits-per-second", "tx-bps"),
		RxPacketsPerSecond: getRowField(row, "rx-packets-per-second", "rx-pps"),
		TxPacketsPerSecond: getRowField(row, "tx-packets-per-second", "tx-pps"),
		RxDropsPerSecond:   getRowField(row, "rx-drops-per-second"),
		TxDropsPerSecond:   getRowField(row, "tx-drops-per-second"),
		RxErrorsPerSecond:  getRowField(row, "rx-errors-per-second"),
		TxErrorsPerSecond:  getRowField(row, "tx-errors-per-second"),
	}
}
