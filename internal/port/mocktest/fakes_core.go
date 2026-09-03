package mocktest

import (
	"context"
	"sync"
	"time"

	domainAudit "github.com/quixiq/polyglot/internal/domain/audit"
	domainNotification "github.com/quixiq/polyglot/internal/domain/notification"
)

// ─── NotificationRepository ─────────────────────────────────────────────

type FakeNotificationRepo struct {
	mu        sync.Mutex
	queued    []domainNotification.WANotification
	templates map[string]domainNotification.NotificationTemplate // key → tpl
}

func NewFakeNotificationRepo() *FakeNotificationRepo {
	return &FakeNotificationRepo{templates: map[string]domainNotification.NotificationTemplate{}}
}

func (f *FakeNotificationRepo) SeedTemplate(t domainNotification.NotificationTemplate) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.templates[t.TemplateKey] = t
}

func (f *FakeNotificationRepo) SaveTemplate(_ context.Context, t domainNotification.NotificationTemplate) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.templates[t.TemplateKey] = t
	return nil
}

func (f *FakeNotificationRepo) FindTemplateByKey(_ context.Context, _, key string) (domainNotification.NotificationTemplate, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	t, ok := f.templates[key]
	if !ok {
		return domainNotification.NotificationTemplate{}, ErrFakeNotFound
	}
	return t, nil
}

func (f *FakeNotificationRepo) ListTemplates(_ context.Context, _ bool) ([]domainNotification.NotificationTemplate, error) {
	return nil, nil
}

func (f *FakeNotificationRepo) Queue(_ context.Context, n domainNotification.WANotification) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.queued = append(f.queued, n)
	return nil
}

func (f *FakeNotificationRepo) Queued() []domainNotification.WANotification {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]domainNotification.WANotification, len(f.queued))
	copy(out, f.queued)
	return out
}

func (f *FakeNotificationRepo) FindByID(_ context.Context, id string) (domainNotification.WANotification, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, n := range f.queued {
		if n.ID == id {
			return n, nil
		}
	}
	return domainNotification.WANotification{}, ErrFakeNotFound
}

func (f *FakeNotificationRepo) Pending(_ context.Context, limit int) ([]domainNotification.WANotification, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []domainNotification.WANotification
	for _, n := range f.queued {
		if n.Status == domainNotification.StatusQueued {
			out = append(out, n)
			if limit > 0 && len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

func (f *FakeNotificationRepo) MarkSent(_ context.Context, id string, sentAt time.Time) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	for i := range f.queued {
		if f.queued[i].ID == id {
			f.queued[i].Status = domainNotification.StatusSent
			f.queued[i].SentAt = &sentAt
			return nil
		}
	}
	return ErrFakeNotFound
}

func (f *FakeNotificationRepo) MarkFailed(_ context.Context, id, errMsg string) error { return nil }

func (f *FakeNotificationRepo) ListByCustomer(_ context.Context, cid string, _ int) ([]domainNotification.WANotification, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []domainNotification.WANotification
	for _, n := range f.queued {
		if n.CustomerID != nil && *n.CustomerID == cid {
			out = append(out, n)
		}
	}
	return out, nil
}

// ─── AuditLogWriter ─────────────────────────────────────────────────────

type FakeAuditWriter struct {
	mu      sync.Mutex
	Entries []domainAudit.AuditLog
}

func (f *FakeAuditWriter) Write(_ context.Context, e domainAudit.AuditLog) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Entries = append(f.Entries, e)
	return nil
}

func (f *FakeAuditWriter) Count(action string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	n := 0
	for _, e := range f.Entries {
		if e.Action == action {
			n++
		}
	}
	return n
}
