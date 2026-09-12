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
