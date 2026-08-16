package hotspot

import (
	"fmt"
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
