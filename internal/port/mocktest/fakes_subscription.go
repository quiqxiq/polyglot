package mocktest

import (
	"context"
	"sync"

	domainSubscription "github.com/quixiq/polyglot/internal/domain/subscription"
)

// FakeSubscriptionRepo is an in-memory subscription repository for tests.
type FakeSubscriptionRepo struct {
	mu   sync.Mutex
	byID map[string]domainSubscription.Subscription
}

// NewFakeSubscriptionRepo creates an empty fake subscription repository.
func NewFakeSubscriptionRepo() *FakeSubscriptionRepo {
	return &FakeSubscriptionRepo{byID: map[string]domainSubscription.Subscription{}}
}

// Seed adds a subscription to in-memory store directly.
func (f *FakeSubscriptionRepo) Seed(s domainSubscription.Subscription) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.byID[s.ID] = s
}

// Save stores a subscription in memory.
func (f *FakeSubscriptionRepo) Save(_ context.Context, s domainSubscription.Subscription) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.byID[s.ID] = s
	return nil
}

// FindByID finds a subscription by ID.
func (f *FakeSubscriptionRepo) FindByID(_ context.Context, id string) (domainSubscription.Subscription, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	s, ok := f.byID[id]
	if !ok {
		return domainSubscription.Subscription{}, ErrFakeNotFound
	}
	return s, nil
}

// FindByCustomerID lists subscriptions for a customer ID.
func (f *FakeSubscriptionRepo) FindByCustomerID(_ context.Context, cid string) ([]domainSubscription.Subscription, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []domainSubscription.Subscription
	for _, s := range f.byID {
		if s.CustomerID == cid {
			out = append(out, s)
		}
	}
	return out, nil
}

// FindAll returns all stored subscriptions.
func (f *FakeSubscriptionRepo) FindAll(_ context.Context) ([]domainSubscription.Subscription, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]domainSubscription.Subscription, 0, len(f.byID))
	for _, s := range f.byID {
		out = append(out, s)
	}
	return out, nil
}

// UpdateStatus updates the status of a subscription.
func (f *FakeSubscriptionRepo) UpdateStatus(_ context.Context, id, status string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	s, ok := f.byID[id]
	if !ok {
		return ErrFakeNotFound
	}
	s.Status = status
	f.byID[id] = s
	return nil
}

// FindByDeviceAndUsername finds a subscription by device ID and remote username.
func (f *FakeSubscriptionRepo) FindByDeviceAndUsername(_ context.Context, deviceID, username string) (domainSubscription.Subscription, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, s := range f.byID {
		if s.DeviceID != nil && *s.DeviceID == deviceID && s.RemoteUsername == username {
			return s, nil
		}
	}
	return domainSubscription.Subscription{}, ErrFakeNotFound
}

// ListActive lists all non-deleted active subscriptions.
func (f *FakeSubscriptionRepo) ListActive(_ context.Context) ([]domainSubscription.Subscription, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []domainSubscription.Subscription
	for _, s := range f.byID {
		if s.Status == domainSubscription.StatusActive && s.DeletedAt == nil {
			out = append(out, s)
		}
	}
	return out, nil
}

// ListLifecycle implements the lifecycle view (ACTIVE + ISOLATED).
func (f *FakeSubscriptionRepo) ListLifecycle(_ context.Context) ([]domainSubscription.Subscription, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []domainSubscription.Subscription
	for _, s := range f.byID {
		if s.DeletedAt != nil {
			continue
		}
		if s.Status == domainSubscription.StatusActive || s.Status == domainSubscription.StatusIsolated {
			out = append(out, s)
		}
	}
	return out, nil
}

// HasActiveForPlan implements the delete-guard lookup.
func (f *FakeSubscriptionRepo) HasActiveForPlan(_ context.Context, planID string) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, s := range f.byID {
		if s.PlanID == planID && s.DeletedAt == nil &&
			(s.Status == domainSubscription.StatusActive || s.Status == domainSubscription.StatusIsolated) {
			return true, nil
		}
	}
	return false, nil
}

// Delete implements hard-delete for the manage-subscription flow.
func (f *FakeSubscriptionRepo) Delete(_ context.Context, id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.byID[id]; !ok {
		return ErrFakeNotFound
	}
	delete(f.byID, id)
	return nil
}
