package mocktest

import (
	"context"
	"sort"
	"sync"

	domainRegistration "github.com/quixiq/polyglot/internal/domain/registration"
)

// FakeRegistrationEventRepo adalah fake in-memory timeline PSB.
type FakeRegistrationEventRepo struct {
	mu     sync.Mutex
	events []domainRegistration.RegistrationEvent
}

// NewFakeRegistrationEventRepo membuat timeline kosong.
func NewFakeRegistrationEventRepo() *FakeRegistrationEventRepo {
	return &FakeRegistrationEventRepo{}
}

// Append menambahkan event ke timeline.
func (f *FakeRegistrationEventRepo) Append(_ context.Context, e domainRegistration.RegistrationEvent) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.events = append(f.events, e)
	return nil
}

// ListByRegistration mengembalikan event satu pendaftaran, terurut waktu.
func (f *FakeRegistrationEventRepo) ListByRegistration(_ context.Context, registrationID string) ([]domainRegistration.RegistrationEvent, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []domainRegistration.RegistrationEvent
	for _, e := range f.events {
		if e.RegistrationID == registrationID {
			out = append(out, e)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out, nil
}

// Count mengembalikan jumlah event bertipe tertentu (helper test).
func (f *FakeRegistrationEventRepo) Count(eventType string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	n := 0
	for _, e := range f.events {
		if e.EventType == eventType {
			n++
		}
	}
	return n
}

// Events mengembalikan salinan seluruh event (helper test).
func (f *FakeRegistrationEventRepo) Events() []domainRegistration.RegistrationEvent {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]domainRegistration.RegistrationEvent, len(f.events))
	copy(out, f.events)
	return out
}
