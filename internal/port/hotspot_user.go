package port

import (
	domainHotspot "github.com/quixiq/polyglot/internal/domain/hotspot"
)

// ListUsersFilter holds optional filters for /ip/hotspot/user/print.
// Mirrors the legacy Mikhmon list/query filters: profile, comment (batch tag)
// and only_unused (uptime=0s — vouchers that have never logged in).
type ListUsersFilter struct {
	Name       string // ?name=   — exact username lookup
	Profile    string // ?profile= — profile filter
	Comment    string // ?comment= — batch/tag filter (e.g. "MyTag")
	OnlyUnused bool   // ?uptime=0s — never logged in
}

// HotspotUserParams holds the parameters for creating or updating a RouterOS
// hotspot user (/ip/hotspot/user/add or /ip/hotspot/user/set).
type HotspotUserParams struct {
	Name          string
	Password      string
	Profile       string
	Server        string
	MACAddress    string
	Address       string
	LimitUptime   string
	LimitBytesIn  string
	LimitBytesOut string
	Comment       string
	Disabled      bool
}

// HotspotUser alias to domain model.
type HotspotUser = domainHotspot.HotspotUser
