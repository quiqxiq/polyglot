package wifi

import (
	"github.com/quixiq/polyglot/internal/domain/command"
)

// NewPrintWirelessInterfacesCommand builds the command.Command for legacy /interface/wireless/print.
func NewPrintWirelessInterfacesCommand(nameFilter string) command.Command {
	args := map[string]string{
		".proplist": ".id,name,master-interface,ssid,band,channel-width,frequency,security-profile,running,disabled,comment",
	}
	if nameFilter != "" {
		args["?name"] = nameFilter
	}
	return command.Command{
		Raw:  "/interface/wireless/print",
		Args: args,
	}
}

// NewPrintWiFiInterfacesCommand builds the command.Command for RouterOS v7 /interface/wifi/print.
func NewPrintWiFiInterfacesCommand(nameFilter string) command.Command {
	args := map[string]string{
		".proplist": ".id,name,master-interface,configuration.ssid,channel.band,channel.width,channel.frequency,security,running,disabled,comment",
	}
	if nameFilter != "" {
		args["?name"] = nameFilter
	}
	return command.Command{
		Raw:  "/interface/wifi/print",
		Args: args,
	}
}

// NewPrintWirelessRegistrationsCommand builds the command.Command for legacy /interface/wireless/registration-table/print.
func NewPrintWirelessRegistrationsCommand(ifaceFilter, macFilter string) command.Command {
	args := map[string]string{
		".proplist": ".id,interface,mac-address,ssid,signal-strength,tx-rate,rx-rate,uptime,bytes,packets,comment",
	}
	if ifaceFilter != "" {
		args["?interface"] = ifaceFilter
	}
	if macFilter != "" {
		args["?mac-address"] = macFilter
	}
	return command.Command{
		Raw:  "/interface/wireless/registration-table/print",
		Args: args,
	}
}

// NewStreamWirelessRegistrationsCommand builds the command.Command for streaming legacy registration table follow.
func NewStreamWirelessRegistrationsCommand(ifaceFilter string) command.Command {
	args := map[string]string{
		"follow":    "",
		".proplist": ".id,interface,mac-address,ssid,signal-strength,tx-rate,rx-rate,uptime,bytes,packets,comment",
	}
	if ifaceFilter != "" {
		args["?interface"] = ifaceFilter
	}
	return command.Command{
		Raw:  "/interface/wireless/registration-table/print",
		Args: args,
	}
}

// NewPrintWiFiRegistrationsCommand builds the command.Command for RouterOS v7 /interface/wifi/registration-table/print.
func NewPrintWiFiRegistrationsCommand(ifaceFilter, macFilter string) command.Command {
	args := map[string]string{
		".proplist": ".id,interface,mac-address,ssid,signal,tx-rate,rx-rate,uptime,tx-bytes,rx-bytes,tx-packets,rx-packets,comment",
	}
	if ifaceFilter != "" {
		args["?interface"] = ifaceFilter
	}
	if macFilter != "" {
		args["?mac-address"] = macFilter
	}
	return command.Command{
		Raw:  "/interface/wifi/registration-table/print",
		Args: args,
	}
}

// NewStreamWiFiRegistrationsCommand builds the command.Command for streaming RouterOS v7 wifi registration table follow.
func NewStreamWiFiRegistrationsCommand(ifaceFilter string) command.Command {
	args := map[string]string{
		"follow":    "",
		".proplist": ".id,interface,mac-address,ssid,signal,tx-rate,rx-rate,uptime,tx-bytes,rx-bytes,tx-packets,rx-packets,comment",
	}
	if ifaceFilter != "" {
		args["?interface"] = ifaceFilter
	}
	return command.Command{
		Raw:  "/interface/wifi/registration-table/print",
		Args: args,
	}
}
