package mocktest

import (
	"context"
	"fmt"
	"strings"
	"sync"

	domainDevice "github.com/quixiq/polyglot/internal/domain/device"
	domainPlan "github.com/quixiq/polyglot/internal/domain/plan"
	domainSub "github.com/quixiq/polyglot/internal/domain/subscription"
	"github.com/quixiq/polyglot/internal/port"
)

// ─── RouterAccountManager ───────────────────────────────────────────────

// FakeRouterAccountManager merekam semua call lifecycle akun router.
type FakeRouterAccountManager struct {
	mu sync.Mutex

	Calls []string // "Provision", "Isolate", ... dengan detail
	Fail  map[string]error
}

func (f *FakeRouterAccountManager) record(op string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Calls = append(f.Calls, op)
	for prefix, err := range f.Fail {
		if strings.HasPrefix(op, prefix) {
			return err
		}
	}
	return nil
}

func (f *FakeRouterAccountManager) Provision(_ context.Context, _, _ string, a port.SubscriberAccount) error {
	return f.record("Provision:" + a.Username + "@" + a.Profile)
}

// ProvisionPPPoE records a mock PPPoE provisioning call.
func (f *FakeRouterAccountManager) ProvisionPPPoE(_ context.Context, _ string, spec domainSub.PPPoEProvisionSpec) error {
	return f.record("ProvisionPPPoE:" + spec.Secret.Username + "@" + spec.Secret.Profile)
}

// ProvisionHotspot records a mock Hotspot provisioning call.
func (f *FakeRouterAccountManager) ProvisionHotspot(_ context.Context, _ string, spec domainSub.HotspotProvisionSpec) error {
	return f.record("ProvisionHotspot:" + spec.User.Username + "@" + spec.User.Profile)
}

// ProvisionDedicated records a mock Dedicated provisioning call.
func (f *FakeRouterAccountManager) ProvisionDedicated(_ context.Context, _ string, spec domainSub.DedicatedProvisionSpec) error {
	return f.record("ProvisionDedicated:" + spec.PPPoE.Secret.Username + "@" + spec.PPPoE.Secret.Profile)
}

func (f *FakeRouterAccountManager) UpdateAccount(_ context.Context, _, _, username, newProfile string) error {
	return f.record("UpdateAccount:" + username + "->" + newProfile)
}

func (f *FakeRouterAccountManager) EnsureProfile(_ context.Context, _, _, profileName, rateLimit string) error {
	return f.record("EnsureProfile:" + profileName + "@" + rateLimit)
}

func (f *FakeRouterAccountManager) Isolate(_ context.Context, _, _, username string, opt port.IsolationOptions) error {
	return f.record(fmt.Sprintf("Isolate:%s->%s(list=%s)", username, opt.IsolirProfile, opt.AddressList))
}

func (f *FakeRouterAccountManager) Restore(_ context.Context, _, _, username, normalProfile, addressList string) error {
	return f.record("Restore:" + username + "->" + normalProfile)
}

func (f *FakeRouterAccountManager) Suspend(_ context.Context, _, _, username string) error {
	return f.record("Suspend:" + username)
}

// Terminate records account termination.
func (f *FakeRouterAccountManager) Terminate(_ context.Context, _, _, username string) error {
	return f.record("Terminate:" + username)
}

// EnsureIsolationInfrastructure records isolation infrastructure provisioning.
func (f *FakeRouterAccountManager) EnsureIsolationInfrastructure(_ context.Context, deviceID string, _ domainDevice.IsolationConfig) error {
	return f.record("EnsureIsolationInfrastructure:" + deviceID)
}

// GetIsolationInfrastructureStatus returns stub isolation infrastructure status.
func (f *FakeRouterAccountManager) GetIsolationInfrastructureStatus(_ context.Context, _ string) (domainDevice.IsolationStatus, error) {
	return domainDevice.IsolationStatus{
		PPPoEProfileExists:   true,
		HotspotProfileExists: true,
		Config:               domainDevice.DefaultIsolationConfig(),
	}, nil
}

// ApplyIntegrationScript records applying integration script.
func (f *FakeRouterAccountManager) ApplyIntegrationScript(_ context.Context, deviceID, profileName, serviceType, scriptType, _ string) error {
	return f.record(fmt.Sprintf("ApplyIntegrationScript:%s->%s:%s:%s", deviceID, profileName, serviceType, scriptType))
}

// SyncPlanProfile records syncing plan profile to router.
func (f *FakeRouterAccountManager) SyncPlanProfile(_ context.Context, deviceID string, plan domainPlan.ServicePlan) error {
	return f.record(fmt.Sprintf("SyncPlanProfile:%s:%s", deviceID, plan.Name))
}

// DeletePlanProfile records deleting plan profile from router.
func (f *FakeRouterAccountManager) DeletePlanProfile(_ context.Context, deviceID string, serviceType, profileName string) error {
	return f.record(fmt.Sprintf("DeletePlanProfile:%s:%s:%s", deviceID, serviceType, profileName))
}

// Count returns how many calls start with the given prefix.
func (f *FakeRouterAccountManager) Count(prefix string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	n := 0
	for _, c := range f.Calls {
		if strings.HasPrefix(c, prefix) {
			n++
		}
	}
	return n
}

func NewFakeRouterAccountManager() *FakeRouterAccountManager {
	return &FakeRouterAccountManager{}
}
