package hotspot

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/quixiq/polyglot/internal/domain/command"
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

// BuildPermanentHotspotOnLoginScript generates a clean on-login script for permanent
// hotspot members/subscriptions without voucher expiration or transaction recording scripts.
func BuildPermanentHotspotOnLoginScript(lockUser, lockServer bool) string {
	var sb strings.Builder
	if lockUser {
		sb.WriteString(`[:local mac $"mac-address"; /ip hotspot user set mac-address=$mac [find where name=$user]]; `)
	}
	if lockServer {
		sb.WriteString(`[:local mac $"mac-address"; :local srv [/ip hotspot host get [find where mac-address="$mac"] server]; /ip hotspot user set server=$srv [find where name=$user]]; `)
	}
	return strings.TrimSpace(sb.String())
}

// BuildVoucherOnLoginScript is an alias for BuildOnLoginScript used for voucher profiles.
func BuildVoucherOnLoginScript(p MikhmonProfileParams) string {
	return BuildOnLoginScript(p)
}

// BuildOnLoginScript generates the exact RouterOS RouterScript string used by
// Mikhmon v4 in the profile's on-login event. Directly ported from Mikhmon v4.
func BuildOnLoginScript(p MikhmonProfileParams) string {
	if p.OnLogin != "" {
		return p.OnLogin
	}

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

	return NewAddUserProfileCommand(HotspotProfileParams{
		Name:           p.Name,
		AddressPool:    p.AddressPool,
		SharedUsers:    sharedUsers,
		RateLimit:      p.RateLimit,
		ParentQueue:    p.ParentQueue,
		AddressList:    p.AddressList,
		SessionTimeout: p.SessionTimeout,
		IdleTimeout:    p.IdleTimeout,
		Comment:        p.Comment,
		OnLogin:        onLoginScript,
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

	// Extract the CSV inside :put ("...")
	start := strings.Index(onLogin, ":put (\",")
	if start == -1 {
		return meta, fmt.Errorf("on-login script does not contain Mikhmon metadata payload")
	}
	start += len(":put (\"") // point to the first comma
	end := strings.Index(onLogin[start:], "\")")
	if end == -1 {
		return meta, fmt.Errorf("malformed Mikhmon metadata payload in on-login script")
	}
	payload := onLogin[start : start+end]

	parts := strings.Split(payload, ",")
	if len(parts) < 8 {
		return meta, fmt.Errorf("unexpected field count in Mikhmon payload: %d (expected at least 8)", len(parts))
	}

	meta.ExpireMode = parts[1]
	if meta.ExpireMode == "" && len(parts) > 5 && parts[5] == "noexp" {
		meta.ExpireMode = "0"
	}
	if p, err := strconv.ParseFloat(parts[2], 64); err == nil {
		meta.Price = p
	}
	meta.Validity = parts[3]
	if sp, err := strconv.ParseFloat(parts[4], 64); err == nil {
		meta.SellingPrice = sp
	}
	meta.LockUser = parts[6]
	meta.LockServer = parts[7]

	return meta, nil
}

// NormalizeProfileName replaces all whitespace sequences with single hyphens, matching Mikhmon's
// legacy preg_replace('/\s+/','-') applied to profile names before they are
// embedded in report script names (report names are split by "-|-").
func NormalizeProfileName(name string) string {
	return strings.Join(strings.Fields(name), "-")
}

// NewSetMikhmonProfileCommand builds the command.Command for /ip/hotspot/user/profile/set
// updating an existing profile's configuration and on-login script.
func NewSetMikhmonProfileCommand(rosID string, p MikhmonProfileParams) command.Command {
	onLoginScript := BuildOnLoginScript(p)

	return NewSetUserProfileCommand(rosID, HotspotProfileParams{
		Name:           p.Name,
		AddressPool:    p.AddressPool,
		SharedUsers:    p.SharedUsers,
		RateLimit:      p.RateLimit,
		ParentQueue:    p.ParentQueue,
		AddressList:    p.AddressList,
		SessionTimeout: p.SessionTimeout,
		IdleTimeout:    p.IdleTimeout,
		Comment:        p.Comment,
		OnLogin:        onLoginScript,
	})
}
