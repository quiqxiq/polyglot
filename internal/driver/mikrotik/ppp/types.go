package ppp

import (
	"github.com/quixiq/polyglot/internal/port"
)

// PPPoESecret is the vendor-neutral PPPoE secret row.
type PPPoESecret = port.PPPoESecret

// PPPoESecretParams holds the parameters needed to create or update a PPPoE secret.
type PPPoESecretParams = port.PPPoESecretParams

// PPPProfile represents one row returned by /ppp/profile/print.
type PPPProfile = port.PPPProfile

// PPPProfileParams holds the parameters for creating or updating a RouterOS PPP profile.
type PPPProfileParams = port.PPPProfileParams

// PPPActiveSession is the vendor-neutral PPP active session row.
type PPPActiveSession = port.PPPActiveSession

// PPPActiveStat is the vendor-neutral PPP active session telemetry stats row.
type PPPActiveStat = port.PPPActiveStat

// IsolirProfileParams returns a pre-filled PPPProfileParams for the standard
// "isolir" (suspension) profile used in ISP billing.
func IsolirProfileParams() PPPProfileParams {
	return PPPProfileParams{
		Name:          "isolir",
		LocalAddress:  "0.0.0.0",
		RemoteAddress: "0.0.0.0",
		RateLimit:     "0/0",
		Comment:       "SUSPENDED_PROFILE",
		SharedUsers:   "1",
	}
}

