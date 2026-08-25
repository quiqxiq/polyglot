package port

import (
	"context"
)

// IsolirConfig describes the router-side isolation infrastructure needed
// before a PPPoE subscriber can be suspended with a payment-portal
// redirect. Values come from system_settings (keys "isolir.*").
type IsolirConfig struct {
	ProfileName    string // e.g. "isolir"
	PoolName       string // e.g. "pool-isolir"
	PoolRange      string // e.g. "172.16.99.10-172.16.99.254"
	PortalIP       string // empty → plain redirect keeps destination, rewrites port only
	PortalHTTPPort string // e.g. "8080"
	RedirectPorts  string // e.g. "80,443"
}

// IsolirSetupResult reports which RouterOS objects existed or were created.
type IsolirSetupResult struct {
	PoolExisted    bool
	ProfileExisted bool
	NATRuleIDs     []string // .id of ISOLIR_REDIRECT rules (existing + created)
	CreatedNATIDs  []string // subset of NATRuleIDs created by this call
}

// IsolirNATRuleStatus reports one expected redirect rule's presence on the
// router, keyed by the destination port it should match.
type IsolirNATRuleStatus struct {
	Port        string
	Exists      bool
	RosID       string
	Action      string
	ToAddresses string
	ToPorts     string
	Comment     string
}

// IsolirInspection is the read-only health snapshot of the isolation
// infrastructure on one router, produced by InspectIsolirInfrastructure.
type IsolirInspection struct {
	PoolExists           bool
	PoolName             string
	PoolRanges           string
	ProfileExists        bool
	ProfileName          string
	ProfileRateLimit     string
	ProfileRemoteAddress string
	NATRules             []IsolirNATRuleStatus
}

// IsolationProvisioner guarantees the suspended-subscriber plumbing exists
// on a router. Implementations must be idempotent — safe to run before
// every suspension.
type IsolationProvisioner interface {
	EnsureIsolirProfile(ctx context.Context, driver DeviceDriver, cfg IsolirConfig) (existed bool, err error)
	EnsureIsolirInfrastructure(ctx context.Context, driver DeviceDriver, cfg IsolirConfig) (IsolirSetupResult, error)
	RemoveIsolirInfrastructure(ctx context.Context, driver DeviceDriver, cfg IsolirConfig) error
	// InspectIsolirInfrastructure reports which pieces of the configured
	// isolation setup currently exist on the router. Read-only.
	InspectIsolirInfrastructure(ctx context.Context, driver DeviceDriver, cfg IsolirConfig) (IsolirInspection, error)
}
