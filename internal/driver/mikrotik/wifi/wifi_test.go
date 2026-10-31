package wifi

import (
	"testing"

	"github.com/quixiq/polyglot/internal/domain/command"
)

func TestWiFiCommands(t *testing.T) {
	cmdWireless := NewPrintWirelessInterfacesCommand("wlan1")
	if cmdWireless.Raw != "/interface/wireless/print" || cmdWireless.Args["?name"] != "wlan1" || cmdWireless.Args[".proplist"] == "" {
		t.Fatalf("unexpected wireless command: %+v", cmdWireless)
	}

	cmdWiFi := NewPrintWiFiInterfacesCommand("wifi1")
	if cmdWiFi.Raw != "/interface/wifi/print" || cmdWiFi.Args["?name"] != "wifi1" || cmdWiFi.Args[".proplist"] == "" {
		t.Fatalf("unexpected wifi command: %+v", cmdWiFi)
	}

	cmdReg := NewPrintWiFiRegistrationsCommand("wifi1", "00:11:22:33:44:55")
	if cmdReg.Raw != "/interface/wifi/registration-table/print" || cmdReg.Args[".proplist"] == "" {
		t.Fatalf("unexpected wifi registrations command: %+v", cmdReg)
	}
}

func TestParseWiFiInterfaces(t *testing.T) {
	res := command.Result{
		Rows: []map[string]string{
			{
				".id":                "*1",
				"name":               "wifi1",
				"configuration.ssid": "Office-5G",
				"channel.band":       "5ghz-ax",
				"channel.width":      "80mhz",
				"channel.frequency":  "5180",
				"security":           "wpa3-psk",
				"running":            "true",
				"disabled":           "false",
			},
		},
	}
	ifaces := ParseWiFiInterfaces(res)
	if len(ifaces) != 1 || ifaces[0].Name != "wifi1" || ifaces[0].SSID != "Office-5G" || !ifaces[0].Running {
		t.Fatalf("unexpected wifi iface parsed: %+v", ifaces)
	}
}

func TestParseWiFiRegistrations(t *testing.T) {
	res := command.Result{
		Rows: []map[string]string{
			{
				".id":         "*12",
				"interface":   "wifi1",
				"mac-address": "A4:83:E7:11:22:33",
				"ssid":        "Office-5G",
				"signal":      "-55",
				"tx-rate":     "866Mbps",
				"rx-rate":     "866Mbps",
				"uptime":      "3h15m",
			},
		},
	}
	regs := ParseWiFiRegistrations(res)
	if len(regs) != 1 || regs[0].MACAddress != "A4:83:E7:11:22:33" || regs[0].Signal != "-55" {
		t.Fatalf("unexpected wifi registration parsed: %+v", regs)
	}
}
