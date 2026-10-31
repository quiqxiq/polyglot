package wifi

import (
	"strconv"

	"github.com/quixiq/polyglot/internal/domain/command"
)

// WiFiScanResult represents one wireless network seen in a frequency scan.
type WiFiScanResult struct {
	SSID      string `json:"ssid"`
	BSSID     string `json:"bssid"`
	Channel   string `json:"channel"`
	Frequency int    `json:"frequency"`
	Signal    int    `json:"signal"`
	SNR       int    `json:"snr"`
	RadioName string `json:"radio_name"`
}

// NewWiFiFrequencyScanCommand triggers a frequency scan on RouterOS v7 wifi interface.
func NewWiFiFrequencyScanCommand(iface, duration string) command.Command {
	args := map[string]string{
		"numbers": iface,
	}
	if duration != "" {
		args["duration"] = duration
	}
	return command.Command{
		Raw:  "/interface/wifi/frequency-scan",
		Args: args,
	}
}

// NewWirelessScanCommand triggers a scan on RouterOS legacy wireless interface.
func NewWirelessScanCommand(iface, duration string) command.Command {
	args := map[string]string{
		"numbers": iface,
	}
	if duration != "" {
		args["duration"] = duration
	}
	return command.Command{
		Raw:  "/interface/wireless/scan",
		Args: args,
	}
}

// ParseScanResults converts reply rows into WiFiScanResult slice.
func ParseScanResults(rows []map[string]string) []WiFiScanResult {
	out := make([]WiFiScanResult, 0, len(rows))
	for _, r := range rows {
		freq, _ := strconv.Atoi(r["frequency"])
		sig, _ := strconv.Atoi(r["signal"])
		snr, _ := strconv.Atoi(r["snr"])
		out = append(out, WiFiScanResult{
			SSID:      r["ssid"],
			BSSID:     r["bssid"],
			Channel:   r["channel"],
			Frequency: freq,
			Signal:    sig,
			SNR:       snr,
			RadioName: r["radio-name"],
		})
	}
	return out
}
