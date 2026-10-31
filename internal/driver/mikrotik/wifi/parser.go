package wifi

import (
	"strings"

	"github.com/quixiq/polyglot/internal/domain/command"
)

// ParseWirelessInterfaces converts command.Result from /interface/wireless/print into []WiFiInterface.
func ParseWirelessInterfaces(result command.Result) []WiFiInterface {
	ifaces := make([]WiFiInterface, 0, len(result.Rows))
	for _, row := range result.Rows {
		id := row[".id"]
		name := row["name"]
		if id == "" || name == "" {
			continue
		}
		ifaces = append(ifaces, WiFiInterface{
			RosID:       id,
			Name:        name,
			MasterIface: row["master-interface"],
			SSID:        row["ssid"],
			Band:        row["band"],
			Channel:     row["channel-width"],
			Frequency:   row["frequency"],
			Security:    row["security-profile"],
			Running:     strings.EqualFold(row["running"], "true") || strings.EqualFold(row["running"], "yes"),
			Disabled:    strings.EqualFold(row["disabled"], "true") || strings.EqualFold(row["disabled"], "yes"),
			Comment:     row["comment"],
		})
	}
	return ifaces
}

// ParseWiFiInterfaces converts command.Result from /interface/wifi/print into []WiFiInterface.
func ParseWiFiInterfaces(result command.Result) []WiFiInterface {
	ifaces := make([]WiFiInterface, 0, len(result.Rows))
	for _, row := range result.Rows {
		id := row[".id"]
		name := row["name"]
		if id == "" || name == "" {
			continue
		}
		ifaces = append(ifaces, WiFiInterface{
			RosID:       id,
			Name:        name,
			MasterIface: row["master-interface"],
			SSID:        row["configuration.ssid"],
			Band:        row["channel.band"],
			Channel:     row["channel.width"],
			Frequency:   row["channel.frequency"],
			Security:    row["security"],
			Running:     strings.EqualFold(row["running"], "true") || strings.EqualFold(row["running"], "yes"),
			Disabled:    strings.EqualFold(row["disabled"], "true") || strings.EqualFold(row["disabled"], "yes"),
			Comment:     row["comment"],
		})
	}
	return ifaces
}

// ParseWiFiRegistrations converts command.Result from registration-table/print into []WiFiRegistration.
func ParseWiFiRegistrations(result command.Result) []WiFiRegistration {
	regs := make([]WiFiRegistration, 0, len(result.Rows))
	for _, row := range result.Rows {
		id := row[".id"]
		mac := row["mac-address"]
		if id == "" || mac == "" {
			continue
		}
		signal := row["signal"]
		if signal == "" {
			signal = row["signal-strength"]
		}
		txBytes := row["tx-bytes"]
		if txBytes == "" {
			txBytes = row["bytes"]
		}
		rxBytes := row["rx-bytes"]
		txPackets := row["tx-packets"]
		if txPackets == "" {
			txPackets = row["packets"]
		}
		rxPackets := row["rx-packets"]

		regs = append(regs, WiFiRegistration{
			RosID:      id,
			Interface:  row["interface"],
			MACAddress: mac,
			SSID:       row["ssid"],
			Signal:     signal,
			TxRate:     row["tx-rate"],
			RxRate:     row["rx-rate"],
			Uptime:     row["uptime"],
			BytesTx:    txBytes,
			BytesRx:    rxBytes,
			PacketsTx:  txPackets,
			PacketsRx:  rxPackets,
			Comment:    row["comment"],
		})
	}
	return regs
}
