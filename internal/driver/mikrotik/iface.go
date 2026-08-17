package mikrotik

import (
	"strings"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/port"
)

// Interface is the vendor-neutral network interface row.
// Canonical definition lives in internal/port (see port.Interface docs).
type Interface = port.Interface

// InterfaceTrafficStats is the vendor-neutral interface traffic rate snapshot.
// Canonical definition lives in internal/port (see port.InterfaceTrafficStats docs).
type InterfaceTrafficStats = port.InterfaceTrafficStats

// NewPrintInterfacesCommand builds the command.Command for /interface/print.
// Pass a non-empty nameFilter to look up one interface by name.
func NewPrintInterfacesCommand(nameFilter string) command.Command {
	args := map[string]string{}
	if nameFilter != "" {
		args["?name"] = nameFilter
	}
	return command.Command{
		Raw:  "/interface/ethernet/print",
		Args: args,
	}
}

// NewStreamInterfacesCommand builds the command.Command for streaming
// /interface/ethernet/print interval=<n>, which makes RouterOS re-send the
// FULL list of interfaces periodically. interval defaults to "1s".
func NewStreamInterfacesCommand(nameFilter, interval string) command.Command {
	if interval == "" {
		interval = "1s"
	}
	args := map[string]string{"interval": interval}
	if nameFilter != "" {
		args["?name"] = nameFilter
	}
	return command.Command{
		Raw:  "/interface/ethernet/print",
		Args: args,
	}
}

// NewMonitorTrafficOnceCommand builds the command.Command for
// /interface/monitor-traffic with the "once" flag, which returns a single
// snapshot of current traffic rates and then finishes (RouterOS sends one
// "!re" then "!done"). Use with Driver.Execute — isStreamingCommand returns
// false for this command because "once" is present.
func NewMonitorTrafficOnceCommand(ifaceName string) command.Command {
	return command.Command{
		Raw: "/interface/monitor-traffic",
		Args: map[string]string{
			"interface": ifaceName,
			"once":      "",
		},
	}
}

// NewMonitorTrafficStreamCommand builds the command.Command for
// /interface/monitor-traffic without "once", which causes RouterOS to send
// updated traffic rate rows every second indefinitely until cancelled.
// Use with Driver.Stream — isStreamingCommand returns true for this command.
// Parse each Result from StreamHandle.Chan() with ParseInterfaceTrafficStats.
func NewMonitorTrafficStreamCommand(ifaceName string) command.Command {
	return command.Command{
		Raw: "/interface/monitor-traffic",
		Args: map[string]string{
			"interface": ifaceName,
			// deliberately no "once" — streaming mode
		},
	}
}

// NewEnableInterfaceCommand builds the command.Command for /interface/enable.
// rosID must come from a prior /interface/print result.
func NewEnableInterfaceCommand(rosID string) command.Command {
	return command.Command{
		Raw:  "/interface/enable",
		Args: map[string]string{".id": rosID},
	}
}

// NewDisableInterfaceCommand builds the command.Command for /interface/disable.
// Classified as ClassDestructive — disabling an interface drops all traffic.
func NewDisableInterfaceCommand(rosID string) command.Command {
	return command.Command{
		Raw:  "/interface/disable",
		Args: map[string]string{".id": rosID},
	}
}

func parseRosBool(val string) bool {
	return strings.EqualFold(val, "true") || strings.EqualFold(val, "yes") || val == "1"
}

// ParseInterfaces converts command.Result rows from /interface/print into
// typed Interface values. Rows missing ".id" or "name" are skipped.
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

func getRowField(row map[string]string, keys ...string) string {
	for _, k := range keys {
		if val, ok := row[k]; ok && val != "" {
			return val
		}
	}
	return ""
}

// ParseInterfaceTrafficStats converts the first row from a
// /interface/monitor-traffic result into InterfaceTrafficStats.
// Returns zero-value if result is empty (device returned no rows).
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
