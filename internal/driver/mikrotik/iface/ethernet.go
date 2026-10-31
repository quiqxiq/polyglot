package iface

import (
	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/driver/mikrotik/internal/rosutil"
)

// EthernetInterface represents physical Ethernet hardware port configuration and status.
type EthernetInterface struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	MacAddress      string `json:"mac_address"`
	Speed           string `json:"speed"`
	FullDuplex      bool   `json:"full_duplex"`
	AutoNegotiation bool   `json:"auto_negotiation"`
	Running         bool   `json:"running"`
	PoEOut          string `json:"poe_out"`
	Disabled        bool   `json:"disabled"`
	Comment         string `json:"comment"`
}

// CableTestResult holds Ethernet physical cable diagnostic findings.
type CableTestResult struct {
	Interface  string `json:"interface"`
	Status     string `json:"status"`
	CablePairs string `json:"cable_pairs"`
}

// NewPrintEthernetCommand queries physical ethernet interfaces with strict proplist.
func NewPrintEthernetCommand() command.Command {
	args := make(map[string]string)
	rosutil.SetProplist(args, ".id", "name", "mac-address", "speed", "full-duplex", "auto-negotiation", "running", "poe-out", "disabled", "comment")
	return command.Command{
		Raw:  "/interface/ethernet/print",
		Args: args,
	}
}

// NewSetEthernetCommand updates ethernet port settings.
func NewSetEthernetCommand(id, comment string, disabled bool) command.Command {
	args := map[string]string{
		".id":     id,
		"comment": comment,
	}
	if disabled {
		args["disabled"] = "yes"
	} else {
		args["disabled"] = "no"
	}
	return command.Command{
		Raw:  "/interface/ethernet/set",
		Args: args,
	}
}

// NewCableTestCommand runs a single physical cable test on an ethernet port.
func NewCableTestCommand(ifaceName string) command.Command {
	return command.Command{
		Raw: "/interface/ethernet/cable-test",
		Args: map[string]string{
			"numbers": ifaceName,
			"once":    "",
		},
	}
}

// ParseEthernetInterfaces converts reply rows into EthernetInterface slice.
func ParseEthernetInterfaces(rows []map[string]string) []EthernetInterface {
	out := make([]EthernetInterface, 0, len(rows))
	for _, r := range rows {
		out = append(out, EthernetInterface{
			ID:              r[".id"],
			Name:            r["name"],
			MacAddress:      r["mac-address"],
			Speed:           r["speed"],
			FullDuplex:      r["full-duplex"] == "true",
			AutoNegotiation: r["auto-negotiation"] != "false",
			Running:         r["running"] == "true",
			PoEOut:          r["poe-out"],
			Disabled:        r["disabled"] == "true",
			Comment:         r["comment"],
		})
	}
	return out
}

// ParseCableTestResult converts reply rows into CableTestResult.
func ParseCableTestResult(rows []map[string]string) *CableTestResult {
	if len(rows) == 0 {
		return nil
	}
	r := rows[0]
	return &CableTestResult{
		Interface:  r["name"],
		Status:     r["status"],
		CablePairs: r["cable-pairs"],
	}
}
