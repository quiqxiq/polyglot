package hotspot

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/driver/mikrotik"
	"github.com/quixiq/polyglot/internal/port"
)

// ExpireMode defines how expired vouchers are handled by Mikhmon's expire monitor script.
// Canonical definition lives in internal/port (see port.ExpireMode).
type ExpireMode = port.ExpireMode

const (
	ExpireModeNotify       ExpireMode = "ntf"  // Mode "N": set limit-uptime=1s and kick
	ExpireModeNotifyRecord ExpireMode = "ntfc" // Mode "N" + record sale transaction
	ExpireModeRemove       ExpireMode = "rem"  // Mode "X": remove user completely
	ExpireModeRemoveRecord ExpireMode = "remc" // Mode "X" + record sale transaction
	ExpireModeNone         ExpireMode = "0"    // No expiration
)

// MikhmonProfileParams is the vendor-neutral Mikhmon profile parameter set.
// Canonical definition lives in internal/port (see port.MikhmonProfileParams docs).
type MikhmonProfileParams = port.MikhmonProfileParams

// BuildOnLoginScript generates the exact RouterOS RouterScript string used by
// Mikhmon v4 in the profile's on-login event. Directly ported from Mikhmon v4.
func BuildOnLoginScript(p MikhmonProfileParams) string {
	expModeCode := string(p.ExpireMode)
	if expModeCode == "" {
		expModeCode = "ntf"
	}

	modeChar := "N"
	switch p.ExpireMode {
	case ExpireModeRemove, ExpireModeRemoveRecord:
		modeChar = "X"
	}

	validity := strings.ToLower(p.Validity)
	if validity == "" {
		validity = "1d"
	}

	price := p.Price
	if price == "" {
		price = "0"
	}
	sprice := p.SellingPrice
	if sprice == "" {
		sprice = "0"
	}

	lockUserStr := "Disable"
	lockScript := ""
	if p.LockUser {
		lockUserStr = "Enable"
		lockScript = `; [:local mac $"mac-address"; /ip hotspot user set mac-address=$mac [find where name=$user]]`
	}

	lockServerStr := "Disable"
	serverLockScript := ""
	if p.LockServer {
		lockServerStr = "Enable"
		serverLockScript = `; [:local mac $"mac-address"; :local srv [/ip hotspot host get [find where mac-address="$mac"] server]; /ip hotspot user set server=$srv [find where name=$user]]`
	}

	// Transaction recording script segment
	recordScript := ""
	if p.EnableRecording || p.ExpireMode == ExpireModeNotifyRecord || p.ExpireMode == ExpireModeRemoveRecord {
		recordScript = fmt.Sprintf(
			`; :local mac $"mac-address"; :local time [/system clock get time ]; /system script add name="$date-|-$time-|-$user-|-%s-|-$address-|-$mac-|-%s-|-%s-|-$comment" owner="$month$year" source=$date comment=mikhmon`,
			price, validity, p.Name,
		)
	}

	// If no expiration mode selected
	if p.ExpireMode == ExpireModeNone {
		if price != "0" && price != "" {
			return fmt.Sprintf(`:put (",,%s,,%s,noexp,%s,%s,")%s%s`, price, sprice, lockUserStr, lockServerStr, lockScript, serverLockScript)
		}
		return ""
	}

	// Base expiration monitoring script segment
	base := fmt.Sprintf(
		`:put (",%s,%s,%s,%s,,%s,%s,"); `+
			`:local mode "%s"; `+
			`{:local date [ /system clock get date ];`+
			`:local year [ :pick $date 7 11 ];`+
			`:local month [ :pick $date 0 3 ];`+
			`:local comment [ /ip hotspot user get [/ip hotspot user find where name="$user"] comment];`+
			` :local ucode [:pic $comment 0 2];`+
			` :if ($ucode = "vc" or $ucode = "up" or $comment = "") do={`+
			` /sys sch add name="$user" disable=no start-date=$date interval="%s";`+
			` :delay 2s;`+
			` :local exp [ /sys sch get [ /sys sch find where name="$user" ] next-run];`+
			` :local getxp [len $exp];`+
			` :if ($getxp = 15) do={`+
			` :local d [:pic $exp 0 6]; :local t [:pic $exp 7 16]; :local s ("/"); :local exp ("$d$s$year $t");`+
			` /ip hotspot user set comment="$exp %s $comment" [find where name="$user"];};`+
			` :if ($getxp = 8) do={ /ip hotspot user set comment="$date $exp %s $comment" [find where name="$user"];};`+
			` :if ($getxp > 15) do={ /ip hotspot user set comment="$exp %s $comment" [find where name="$user"];};`+
			` /sys sch remove [find where name="$user"]`,
		expModeCode, price, validity, sprice, lockUserStr, lockServerStr,
		modeChar, validity, modeChar, modeChar, modeChar,
	)

	return base + recordScript + lockScript + serverLockScript + "}}"
}

// NewAddMikhmonProfileCommand builds the command.Command for /ip/hotspot/user/profile/add
// configured with Mikhmon's on-login script and metadata.
func NewAddMikhmonProfileCommand(p MikhmonProfileParams) command.Command {
	onLoginScript := BuildOnLoginScript(p)

	sharedUsers := p.SharedUsers
	if sharedUsers == "" {
		sharedUsers = "1"
	}

	return mikrotik.NewAddHotspotUserProfileCommand(mikrotik.HotspotUserProfileParams{
		Name:        p.Name,
		AddressPool: p.AddressPool,
		SharedUsers: sharedUsers,
		RateLimit:   p.RateLimit,
		ParentQueue: p.ParentQueue,
		Comment:     p.Comment,
		OnLogin:     onLoginScript,
	})
}

// ProfileMeta holds the structured Mikhmon metadata embedded in a profile's
// on-login script CSV payload (the ":put (\",...\")" string built by
// BuildOnLoginScript). Legacy Mikhmon parses this CSV on every edit screen;
// this is the Go equivalent.
type ProfileMeta struct {
	ExpireMode   string  // "0" (noexp), "ntf", "ntfc", "rem", "remc"
	Price        float64 // index [2] — selling price
	Validity     string  // index [3] — e.g. "1h", "1d" (empty for noexp)
	SellingPrice float64 // index [4] — cost price
	LockUser     string  // index [6] — "Enable" | "Disable"
	LockServer   string  // index [7] — "Enable" | "Disable"
}

// ParseOnLoginScript extracts structured ProfileMeta from a profile's on-login
// script. Both layouts produced by BuildOnLoginScript are recognised:
//
//	standard: :put (",<mode>,<price>,<validity>,<sprice>,,<lockUser>,<lockServer>,")
//	noexp   : :put (",,<price>,,<sprice>,noexp,<lockUser>,<lockServer>,")
//
// An empty script (noexp with zero price) returns zero metadata without error;
// a script without a ":put (\",...\")" payload returns an error.
func ParseOnLoginScript(onLogin string) (ProfileMeta, error) {
	meta := ProfileMeta{}
	if strings.TrimSpace(onLogin) == "" {
		return meta, nil
	}

	start := strings.Index(onLogin, ":put (")
	if start < 0 {
		return meta, fmt.Errorf("parse on-login: no :put payload")
	}
	rest := onLogin[start+len(":put ("):]
	end := strings.Index(rest, ")")
	if end < 0 {
		return meta, fmt.Errorf("parse on-login: unterminated :put payload")
	}
	payload := strings.Trim(rest[:end], `"`)
	tokens := strings.Split(payload, ",")
	if len(tokens) < 8 {
		return meta, fmt.Errorf("parse on-login: unexpected payload %q", payload)
	}

	meta.ExpireMode = tokens[1]
	// noexp layout puts the "noexp" marker at index 5 with mode empty.
	if meta.ExpireMode == "" && tokens[5] == "noexp" {
		meta.ExpireMode = "0"
	}
	meta.Price, _ = strconv.ParseFloat(tokens[2], 64)
	meta.Validity = tokens[3]
	meta.SellingPrice, _ = strconv.ParseFloat(tokens[4], 64)
	meta.LockUser = tokens[6]
	meta.LockServer = tokens[7]
	return meta, nil
}

// NormalizeProfileName replaces every whitespace run with "-", mirroring the
// legacy preg_replace('/\s+/','-') applied to profile names before they are
// embedded in report script names (report names are split by "-|-").
func NormalizeProfileName(name string) string {
	return strings.Join(strings.Fields(name), "-")
}

// NewSetMikhmonProfileCommand builds the command.Command for /ip/hotspot/user/profile/set
// updating an existing profile's configuration and on-login script.
func NewSetMikhmonProfileCommand(rosID string, p MikhmonProfileParams) command.Command {
	onLoginScript := BuildOnLoginScript(p)

	return mikrotik.NewSetHotspotUserProfileCommand(rosID, mikrotik.HotspotUserProfileParams{
		Name:        p.Name,
		AddressPool: p.AddressPool,
		SharedUsers: p.SharedUsers,
		RateLimit:   p.RateLimit,
		ParentQueue: p.ParentQueue,
		Comment:     p.Comment,
		OnLogin:     onLoginScript,
	})
}
