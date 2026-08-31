package hotspot

import (
	"github.com/quixiq/polyglot/internal/port"
)

// HotspotActiveSession is the vendor-neutral hotspot active session row.
type HotspotActiveSession = port.HotspotActiveSession

// HotspotActiveStat is the vendor-neutral hotspot active session telemetry stats row.
type HotspotActiveStat = port.HotspotActiveStat

// HotspotUser is the vendor-neutral hotspot user row.
type HotspotUser = port.HotspotUser

// HotspotUserProfile is the vendor-neutral hotspot user profile row.
type HotspotUserProfile = port.HotspotUserProfile

// HotspotServer represents one row returned by /ip/hotspot/print.
type HotspotServer struct {
	RosID       string
	Name        string
	Interface   string
	Profile     string
	AddressPool string
	Disabled    bool
}

// HotspotUserParams holds parameters for creating/modifying a hotspot user.
type HotspotUserParams struct {
	Server      string
	Name        string
	Password    string
	Profile     string
	Comment     string
	LimitUptime string
	LimitBytes  string
	Disabled    bool
}

// HotspotProfileParams holds parameters for creating/modifying a hotspot user profile.
type HotspotProfileParams struct {
	Name           string
	RateLimit      string
	SessionTimeout string
	IdleTimeout    string
	SharedUsers    string
	ParentQueue    string
	AddressPool    string
	AddressList    string
	Comment        string
	OnLogin        string
}
