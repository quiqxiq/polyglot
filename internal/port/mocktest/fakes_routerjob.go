package mocktest

import (
	"context"
	"sort"
	"sync"
	"time"

	domainRouterJob "github.com/quixiq/polyglot/internal/domain/routerjob"
	"github.com/quixiq/polyglot/internal/port"
)

// FakeRouterJobRepo adalah fake in-memory antrean job router.
type FakeRouterJobRepo struct {
	mu   sync.Mutex
	byID map[string]domainRouterJob.Job
}

// NewFakeRouterJobRepo constructs an empty fake router job queue.
func NewFakeRouterJobRepo() *FakeRouterJobRepo {
	return &FakeRouterJobRepo{byID: map[string]domainRouterJob.Job{}}
}

// Seed adds a job directly (test helper).
func (f *FakeRouterJobRepo) Seed(j domainRouterJob.Job) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.byID[j.ID] = j
}

// Enqueue menambah job secara idempoten berdasarkan idempotency_key.
func (f *FakeRouterJobRepo) Enqueue(_ context.Context, job domainRouterJob.Job) (domainRouterJob.Job, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, existing := range f.byID {
		if existing.IdempotencyKey == job.IdempotencyKey {
			return existing, nil
		}
	}
	f.byID[job.ID] = job
	return job, nil
}

// ClaimDue mengambil job yang siap dijalankan.
func (f *FakeRouterJobRepo) ClaimDue(_ context.Context, limit int, now time.Time) ([]domainRouterJob.Job, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var due []domainRouterJob.Job
	for _, j := range f.byID {
		if (j.Status == domainRouterJob.StatusQueued || j.Status == domainRouterJob.StatusRetryWait) &&
			!j.NextRunAt.After(now) {
			due = append(due, j)
		}
	}
	sort.Slice(due, func(i, k int) bool { return due[i].NextRunAt.Before(due[k].NextRunAt) })
	if limit > 0 && len(due) > limit {
		due = due[:limit]
	}
	for i := range due {
		j := due[i]
		j.Status = domainRouterJob.StatusRunning
		j.Attempts++
		j.UpdatedAt = now
		f.byID[j.ID] = j
		due[i] = j
	}
	return due, nil
}

// MarkSucceeded menandai job sukses.
func (f *FakeRouterJobRepo) MarkSucceeded(_ context.Context, id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	j, ok := f.byID[id]
	if !ok {
		return ErrFakeNotFound
	}
	j.Status = domainRouterJob.StatusSucceeded
	j.LastError = ""
	f.byID[id] = j
	return nil
}

// MarkRetry menandai job menunggu retry.
func (f *FakeRouterJobRepo) MarkRetry(_ context.Context, id string, nextRunAt time.Time, lastErr string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	j, ok := f.byID[id]
	if !ok {
		return ErrFakeNotFound
	}
	j.Status = domainRouterJob.StatusRetryWait
	j.NextRunAt = nextRunAt
	j.LastError = lastErr
	f.byID[id] = j
	return nil
}

// MarkDead menandai job gagal permanen.
func (f *FakeRouterJobRepo) MarkDead(_ context.Context, id string, lastErr string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	j, ok := f.byID[id]
	if !ok {
		return ErrFakeNotFound
	}
	j.Status = domainRouterJob.StatusDead
	j.LastError = lastErr
	f.byID[id] = j
	return nil
}

// Requeue mengembalikan job DEAD/FAILED ke QUEUED.
func (f *FakeRouterJobRepo) Requeue(_ context.Context, id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	j, ok := f.byID[id]
	if !ok {
		return ErrFakeNotFound
	}
	j.Status = domainRouterJob.StatusQueued
	j.Attempts = 0
	j.LastError = ""
	j.NextRunAt = time.Now()
	f.byID[id] = j
	return nil
}

// FindByID mencari job berdasarkan id.
func (f *FakeRouterJobRepo) FindByID(_ context.Context, id string) (domainRouterJob.Job, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	j, ok := f.byID[id]
	if !ok {
		return domainRouterJob.Job{}, ErrFakeNotFound
	}
	return j, nil
}

// List mengembalikan job terbaru, opsional difilter status.
func (f *FakeRouterJobRepo) List(_ context.Context, filter port.RouterJobFilter) ([]domainRouterJob.Job, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]domainRouterJob.Job, 0, len(f.byID))
	for _, j := range f.byID {
		if filter.Status != "" && string(j.Status) != filter.Status {
			continue
		}
		out = append(out, j)
	}
	sort.Slice(out, func(i, k int) bool { return out[i].CreatedAt.After(out[k].CreatedAt) })
	if filter.Limit > 0 && len(out) > filter.Limit {
		out = out[:filter.Limit]
	}
	return out, nil
}
