// Package mocktest menyediakan fake in-memory untuk port ISP — dipakai
// unit test usecase (pola memRepo/memCredVault ala server_test.go).
package mocktest

import (
	"context"
	"errors"
	"sync"
	"time"

	domainPlan "github.com/quixiq/polyglot/internal/domain/plan"
	domainRegistration "github.com/quixiq/polyglot/internal/domain/registration"
	"github.com/quixiq/polyglot/internal/port"
)

// ErrFakeNotFound adalah error not-found generik fakes.
var ErrFakeNotFound = errors.New("mocktest: not found")

// ─── RegistrationRepository ─────────────────────────────────────────────

type FakeRegistrationRepo struct {
	mu   sync.Mutex
	byID map[string]domainRegistration.Registration
}

func NewFakeRegistrationRepo() *FakeRegistrationRepo {
	return &FakeRegistrationRepo{byID: map[string]domainRegistration.Registration{}}
}

func (f *FakeRegistrationRepo) Save(_ context.Context, r domainRegistration.Registration) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.byID[r.ID] = r
	return nil
}

func (f *FakeRegistrationRepo) FindByID(_ context.Context, id string) (domainRegistration.Registration, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	r, ok := f.byID[id]
	if !ok {
		return domainRegistration.Registration{}, ErrFakeNotFound
	}
	return r, nil
}

func (f *FakeRegistrationRepo) FindByNo(_ context.Context, no string) (domainRegistration.Registration, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, r := range f.byID {
		if r.RegistrationNo == no {
			return r, nil
		}
	}
	return domainRegistration.Registration{}, ErrFakeNotFound
}

func (f *FakeRegistrationRepo) List(_ context.Context, _ port.RegistrationFilter) ([]domainRegistration.Registration, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]domainRegistration.Registration, 0, len(f.byID))
	for _, r := range f.byID {
		out = append(out, r)
	}
	return out, nil
}

func (f *FakeRegistrationRepo) UpdateStatus(_ context.Context, id, status string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	r, ok := f.byID[id]
	if !ok {
		return ErrFakeNotFound
	}
	r.Status = status
	f.byID[id] = r
	return nil
}

func (f *FakeRegistrationRepo) CountPendingSince(_ context.Context, _ time.Time) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var n int64
	for _, r := range f.byID {
		if r.Status == domainRegistration.StatusPending {
			n++
		}
	}
	return n, nil
}

// ─── ServicePlanRepository ──────────────────────────────────────────────

type FakeServicePlanRepo struct {
	mu   sync.Mutex
	byID map[string]domainPlan.ServicePlan
}

func NewFakeServicePlanRepo() *FakeServicePlanRepo {
	return &FakeServicePlanRepo{byID: map[string]domainPlan.ServicePlan{}}
}

func (f *FakeServicePlanRepo) Seed(p domainPlan.ServicePlan) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.byID[p.ID] = p
}

func (f *FakeServicePlanRepo) Save(_ context.Context, p domainPlan.ServicePlan) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.byID[p.ID] = p
	return nil
}

func (f *FakeServicePlanRepo) FindByID(_ context.Context, id string) (domainPlan.ServicePlan, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	p, ok := f.byID[id]
	if !ok {
		return domainPlan.ServicePlan{}, ErrFakeNotFound
	}
	return p, nil
}

func (f *FakeServicePlanRepo) FindByName(_ context.Context, tenant, name string) (domainPlan.ServicePlan, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, p := range f.byID {
		if p.TenantID == tenant && p.Name == name {
			return p, nil
		}
	}
	return domainPlan.ServicePlan{}, ErrFakeNotFound
}

func (f *FakeServicePlanRepo) List(_ context.Context, activeOnly bool) ([]domainPlan.ServicePlan, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []domainPlan.ServicePlan
	for _, p := range f.byID {
		if activeOnly && !p.IsActive {
			continue
		}
		out = append(out, p)
	}
	return out, nil
}

func (f *FakeServicePlanRepo) Delete(_ context.Context, id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.byID, id)
	return nil
}
