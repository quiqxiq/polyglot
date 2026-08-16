package hotspot

import (
	"strings"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/driver/mikrotik"
	"github.com/quixiq/polyglot/internal/port"
)

// MikhmonTransaction is the vendor-neutral Mikhmon transaction row.
// Canonical definition lives in internal/port (see port.MikhmonTransaction docs).
type MikhmonTransaction = port.MikhmonTransaction

// NewPrintMikhmonReportsCommand builds a command.Command to fetch transaction logs
// from RouterOS /system/script entries matching comment="mikhmon".
// Reuses mikrotik.NewPrintSystemScriptsCommand.
func NewPrintMikhmonReportsCommand() command.Command {
	return mikrotik.NewPrintSystemScriptsCommand("", "mikhmon")
}

// ParseMikhmonTransactions parses command.Result rows from /system/script/print
// (filtered by comment="mikhmon") into typed MikhmonTransaction structs.
//
// Mikhmon encodes transaction metadata into the script's `name` attribute using
// "-|-" delimiter: "date-|-time-|-user-|-price-|-address-|-mac-|-validity-|-profile-|-comment".
func ParseMikhmonTransactions(result command.Result) []MikhmonTransaction {
	scripts := mikrotik.ParseSystemScripts(result)
	records := make([]MikhmonTransaction, 0, len(scripts))

	for _, s := range scripts {
		if !strings.Contains(s.Name, "-|-") {
			continue
		}

		parts := strings.Split(s.Name, "-|-")
		rec := MikhmonTransaction{
			RosID:   s.RosID,
			RawName: s.Name,
		}

		if len(parts) >= 1 {
			rec.Date = parts[0]
		}
		if len(parts) >= 2 {
			rec.Time = parts[1]
		}
		if len(parts) >= 3 {
			rec.Username = parts[2]
		}
		if len(parts) >= 4 {
			rec.Price = parts[3]
		}
		if len(parts) >= 5 {
			rec.Address = parts[4]
		}
		if len(parts) >= 6 {
			rec.MAC = parts[5]
		}
		if len(parts) >= 7 {
			rec.Validity = parts[6]
		}
		if len(parts) >= 8 {
			rec.Profile = parts[7]
		}
		if len(parts) >= 9 {
			rec.Comment = parts[8]
		}

		records = append(records, rec)
	}

	return records
}
