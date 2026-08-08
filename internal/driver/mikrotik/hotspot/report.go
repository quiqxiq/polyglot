package hotspot

import (
	"strings"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/driver/mikrotik"
)

// MikhmonTransaction represents one transaction record parsed from RouterOS
// /system/script entries created by Mikhmon v4's on-login script.
//
// Field notes:
//   - RosID   : internal RouterOS script ID (e.g. "*1")
//   - Date    : transaction date string (e.g. "dec/25/2024")
//   - Time    : transaction time string (e.g. "10:30:15")
//   - Username: subscriber username
//   - Price   : price recorded
//   - Address : IP address assigned to subscriber during login
//   - MAC     : MAC address of subscriber device
//   - Validity: validity string (e.g. "1d")
//   - Profile : profile name
//   - Comment : user comment
type MikhmonTransaction struct {
	RosID    string
	Date     string
	Time     string
	Username string
	Price    string
	Address  string
	MAC      string
	Validity string
	Profile  string
	Comment  string
	RawName  string
}

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
