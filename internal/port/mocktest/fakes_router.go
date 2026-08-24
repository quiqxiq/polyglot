package mocktest

import (
	"context"
	"fmt"
	"strings"
	"sync"

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

func (f *FakeRouterAccountManager) Terminate(_ context.Context, _, _, username string) error {
	return f.record("Terminate:" + username)
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
