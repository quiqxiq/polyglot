// DEVIASI: empat fake untuk port AI Router Ops (approval, restore point,
// command log, active device) sengaja digabung di satu file sesuai kontrak
// tugas M1 (fakes_ai_ops.go), sehingga melewati batas 500 baris; tiap fake
// tetap dipisah jelas dengan penanda seksi.
package mocktest

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/quixiq/polyglot/internal/domain/approval"
	"github.com/quixiq/polyglot/internal/domain/audit"
	"github.com/quixiq/polyglot/internal/domain/restorepoint"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/idgen"
)

// fakeApprovalTTL adalah pengaman yang sama dengan repository Postgres bila
// usecase tidak mengisi ExpiresAt.
const fakeApprovalTTL = 15 * time.Minute

// fakeRestorePointRetention adalah pengaman retensi restore point pada fake.
const fakeRestorePointRetention = 30 * 24 * time.Hour

// ─── ApprovalRepository ─────────────────────────────────────────────────

// FakeApprovalRepo adalah fake in-memory port.ApprovalRepository dengan
// semantik CAS revision yang sama seperti adapter Postgres.
type FakeApprovalRepo struct {
	mu   sync.Mutex
	byID map[string]approval.Request
}

var _ port.ApprovalRepository = (*FakeApprovalRepo)(nil)

// NewFakeApprovalRepo constructs an empty fake approval repository.
func NewFakeApprovalRepo() *FakeApprovalRepo {
	return &FakeApprovalRepo{byID: map[string]approval.Request{}}
}

// Seed menambahkan approval langsung ke store (test helper) sambil
// menormalkan field wajib.
func (f *FakeApprovalRepo) Seed(r approval.Request) approval.Request {
	f.mu.Lock()
	defer f.mu.Unlock()
	r = normalizeFakeApproval(r)
	f.byID[r.ID] = r
	return r
}

// Create menyimpan pengajuan approval baru.
func (f *FakeApprovalRepo) Create(_ context.Context, req approval.Request) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	r := normalizeFakeApproval(req)
	f.byID[r.ID] = r
	return nil
}

// FindByID mencari approval berdasarkan id.
func (f *FakeApprovalRepo) FindByID(_ context.Context, id string) (approval.Request, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	r, ok := f.byID[id]
	if !ok {
		return approval.Request{}, approval.ErrNotFound
	}
	return r, nil
}

// FindByIdempotencyKey mencari approval berdasarkan kunci idempotensi tenant.
func (f *FakeApprovalRepo) FindByIdempotencyKey(_ context.Context, tenantID, key string) (approval.Request, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if key == "" {
		return approval.Request{}, approval.ErrNotFound
	}
	tenant := tenantID
	if tenant == "" {
		tenant = fakeDefaultTenantID
	}
	for _, r := range f.byID {
		if r.TenantID == tenant && r.IdempotencyKey == key {
			return r, nil
		}
	}
	return approval.Request{}, approval.ErrNotFound
}

// List mengembalikan approval sesuai filter, terbaru lebih dulu.
func (f *FakeApprovalRepo) List(_ context.Context, filter port.ApprovalFilter) ([]approval.Request, int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	tenant := filter.TenantID
	if tenant == "" {
		tenant = fakeDefaultTenantID
	}
	out := make([]approval.Request, 0)
	for _, r := range f.byID {
		if r.TenantID != tenant {
			continue
		}
		if filter.DeviceID != "" && r.DeviceID != filter.DeviceID {
			continue
		}
		if filter.Status != "" && string(r.Status) != filter.Status {
			continue
		}
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool {
		if !out[i].CreatedAt.Equal(out[j].CreatedAt) {
			return out[i].CreatedAt.After(out[j].CreatedAt)
		}
		return out[i].ID > out[j].ID
	})
	total := len(out)
	if filter.Offset > 0 {
		if filter.Offset >= len(out) {
			return []approval.Request{}, total, nil
		}
		out = out[filter.Offset:]
	}
	if filter.Limit > 0 && filter.Limit < len(out) {
		out = out[:filter.Limit]
	}
	return out, total, nil
}

// UpdateIfRevision menerapkan compare-and-set pada Revision.
func (f *FakeApprovalRepo) UpdateIfRevision(_ context.Context, req approval.Request, expectedRevision int) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	stored, ok := f.byID[req.ID]
	if !ok {
		return false, approval.ErrNotFound
	}
	if stored.Revision != expectedRevision {
		return false, nil
	}
	req.UpdatedAt = time.Now().UTC()
	f.byID[req.ID] = req
	return true, nil
}

// ClaimForExecution mengklaim APPROVED→EXECUTING secara atomik.
func (f *FakeApprovalRepo) ClaimForExecution(_ context.Context, id string, now time.Time) (approval.Request, bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	r, ok := f.byID[id]
	if !ok {
		return approval.Request{}, false, approval.ErrNotFound
	}
	if !r.CanExecute(now) {
		return approval.Request{}, false, nil
	}
	if r.ExecutionNonce == "" {
		r.ExecutionNonce = idgen.New("exec")
	}
	r.Status = approval.StatusExecuting
	r.Revision++
	r.UpdatedAt = now
	f.byID[id] = r
	return r, true, nil
}

// ExpirePending menandai PENDING kedaluwarsa menjadi EXPIRED.
func (f *FakeApprovalRepo) ExpirePending(_ context.Context, now time.Time) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	expired := 0
	for id, r := range f.byID {
		if r.Status != approval.StatusPending || !r.IsExpired(now) {
			continue
		}
		r.Status = approval.StatusExpired
		r.Revision++
		r.UpdatedAt = now
		f.byID[id] = r
		expired++
	}
	return expired, nil
}

// normalizeFakeApproval menyamakan default dengan adapter Postgres.
func normalizeFakeApproval(r approval.Request) approval.Request {
	if r.ID == "" {
		r.ID = idgen.New("appr")
	}
	if r.TenantID == "" {
		r.TenantID = fakeDefaultTenantID
	}
	if r.Status == "" {
		r.Status = approval.StatusPending
	}
	if r.Revision <= 0 {
		r.Revision = 1
	}
	if r.CreatedAt.IsZero() {
		r.CreatedAt = time.Now().UTC()
	}
	if r.ExpiresAt.IsZero() {
		r.ExpiresAt = r.CreatedAt.Add(fakeApprovalTTL)
	}
	if r.UpdatedAt.IsZero() {
		r.UpdatedAt = r.CreatedAt
	}
	return r
}

// ─── RestorePointRepository ─────────────────────────────────────────────

// FakeRestorePointRepo adalah fake in-memory port.RestorePointRepository.
type FakeRestorePointRepo struct {
	mu   sync.Mutex
	byID map[string]restorepoint.RestorePoint
}

var _ port.RestorePointRepository = (*FakeRestorePointRepo)(nil)

// NewFakeRestorePointRepo constructs an empty fake restore point repository.
func NewFakeRestorePointRepo() *FakeRestorePointRepo {
	return &FakeRestorePointRepo{byID: map[string]restorepoint.RestorePoint{}}
}

// Seed menambahkan restore point langsung ke store (test helper).
func (f *FakeRestorePointRepo) Seed(rp restorepoint.RestorePoint) restorepoint.RestorePoint {
	f.mu.Lock()
	defer f.mu.Unlock()
	rp = normalizeFakeRestorePoint(rp)
	f.byID[rp.ID] = rp
	return rp
}

// Create menyimpan bundle restore point beserta artefaknya.
func (f *FakeRestorePointRepo) Create(_ context.Context, rp restorepoint.RestorePoint) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	rp = normalizeFakeRestorePoint(rp)
	f.byID[rp.ID] = rp
	return nil
}

// FindByID mencari bundle restore point berdasarkan id.
func (f *FakeRestorePointRepo) FindByID(_ context.Context, id string) (restorepoint.RestorePoint, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	rp, ok := f.byID[id]
	if !ok {
		return restorepoint.RestorePoint{}, restorepoint.ErrNotFound
	}
	return rp, nil
}

// List mengembalikan header restore point sesuai filter, terbaru lebih dulu.
func (f *FakeRestorePointRepo) List(_ context.Context, filter port.RestorePointFilter) ([]restorepoint.RestorePoint, int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	tenant := filter.TenantID
	if tenant == "" {
		tenant = fakeDefaultTenantID
	}
	out := make([]restorepoint.RestorePoint, 0)
	for _, rp := range f.byID {
		if rp.TenantID != tenant {
			continue
		}
		if filter.DeviceID != "" && rp.DeviceID != filter.DeviceID {
			continue
		}
		if filter.Status != "" && string(rp.Status) != filter.Status {
			continue
		}
		stripped := rp
		stripped.Artifacts = nil
		out = append(out, stripped)
	}
	sort.Slice(out, func(i, j int) bool {
		if !out[i].CreatedAt.Equal(out[j].CreatedAt) {
			return out[i].CreatedAt.After(out[j].CreatedAt)
		}
		return out[i].ID > out[j].ID
	})
	total := len(out)
	if filter.Offset > 0 {
		if filter.Offset >= len(out) {
			return []restorepoint.RestorePoint{}, total, nil
		}
		out = out[filter.Offset:]
	}
	if filter.Limit > 0 && filter.Limit < len(out) {
		out = out[:filter.Limit]
	}
	return out, total, nil
}

// UpdateStatus mengubah status bundle dan (opsional) used_at.
func (f *FakeRestorePointRepo) UpdateStatus(_ context.Context, id string, status restorepoint.Status, usedAt *time.Time) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	rp, ok := f.byID[id]
	if !ok {
		return restorepoint.ErrNotFound
	}
	rp.Status = status
	if usedAt != nil {
		rp.UsedAt = usedAt
	}
	rp.UpdatedAt = time.Now().UTC()
	f.byID[id] = rp
	return nil
}

// FindLatestByConfigHash mencari bundle AVAILABLE terbaru dengan config hash
// yang sama untuk device.
func (f *FakeRestorePointRepo) FindLatestByConfigHash(_ context.Context, deviceID, configHash string) (restorepoint.RestorePoint, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var best restorepoint.RestorePoint
	found := false
	for _, rp := range f.byID {
		if rp.DeviceID != deviceID || rp.ConfigHash != configHash || rp.Status != restorepoint.StatusAvailable {
			continue
		}
		if !found || rp.CreatedAt.After(best.CreatedAt) || (rp.CreatedAt.Equal(best.CreatedAt) && rp.ID > best.ID) {
			best = rp
			found = true
		}
	}
	if !found {
		return restorepoint.RestorePoint{}, restorepoint.ErrNotFound
	}
	return best, nil
}

// ListExpired meniru semantik retensi: kedaluwarsa sebelum `before`, tetapi
// selalu mempertahankan `keepLast` bundle terbaru per device.
func (f *FakeRestorePointRepo) ListExpired(_ context.Context, before time.Time, keepLast int) ([]restorepoint.RestorePoint, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if keepLast < 0 {
		keepLast = 0
	}
	byDevice := map[string][]restorepoint.RestorePoint{}
	for _, rp := range f.byID {
		if rp.Status == restorepoint.StatusDeleted || !rp.ExpiresAt.Before(before) {
			continue
		}
		byDevice[rp.DeviceID] = append(byDevice[rp.DeviceID], rp)
	}
	out := make([]restorepoint.RestorePoint, 0)
	for _, group := range byDevice {
		sort.Slice(group, func(i, j int) bool {
			if !group[i].CreatedAt.Equal(group[j].CreatedAt) {
				return group[i].CreatedAt.After(group[j].CreatedAt)
			}
			return group[i].ID > group[j].ID
		})
		if keepLast >= len(group) {
			continue
		}
		out = append(out, group[keepLast:]...)
	}
	sort.Slice(out, func(i, j int) bool {
		if !out[i].CreatedAt.Equal(out[j].CreatedAt) {
			return out[i].CreatedAt.Before(out[j].CreatedAt)
		}
		return out[i].ID < out[j].ID
	})
	return out, nil
}

// Delete menandai bundle DELETED dan mengembalikan bundle beserta artefaknya.
func (f *FakeRestorePointRepo) Delete(_ context.Context, id string) (restorepoint.RestorePoint, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	rp, ok := f.byID[id]
	if !ok {
		return restorepoint.RestorePoint{}, restorepoint.ErrNotFound
	}
	rp.Status = restorepoint.StatusDeleted
	rp.UpdatedAt = time.Now().UTC()
	f.byID[id] = rp
	return rp, nil
}

// normalizeFakeRestorePoint menyamakan default dengan adapter Postgres.
func normalizeFakeRestorePoint(rp restorepoint.RestorePoint) restorepoint.RestorePoint {
	if rp.ID == "" {
		rp.ID = idgen.New("rp")
	}
	if rp.TenantID == "" {
		rp.TenantID = fakeDefaultTenantID
	}
	if rp.Status == "" {
		rp.Status = restorepoint.StatusCreating
	}
	if rp.CreatedAt.IsZero() {
		rp.CreatedAt = time.Now().UTC()
	}
	if rp.ExpiresAt.IsZero() {
		rp.ExpiresAt = rp.CreatedAt.Add(fakeRestorePointRetention)
	}
	if rp.UpdatedAt.IsZero() {
		rp.UpdatedAt = rp.CreatedAt
	}
	return rp
}

// ─── CommandLogRepository ───────────────────────────────────────────────

// FakeCommandLogRepo adalah fake in-memory port.CommandLogRepository.
type FakeCommandLogRepo struct {
	mu      sync.Mutex
	entries []audit.CommandLog
}

var _ port.CommandLogRepository = (*FakeCommandLogRepo)(nil)

// NewFakeCommandLogRepo constructs an empty fake command log repository.
func NewFakeCommandLogRepo() *FakeCommandLogRepo {
	return &FakeCommandLogRepo{}
}

// Seed menambahkan entri command log langsung (test helper).
func (f *FakeCommandLogRepo) Seed(entry audit.CommandLog) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.entries = append(f.entries, normalizeFakeCommandLog(entry, int64(len(f.entries)+1)))
}

// Write menambahkan satu entri command log append-only (ID auto).
func (f *FakeCommandLogRepo) Write(_ context.Context, entry audit.CommandLog) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if entry.Command == "" || entry.Mode == "" || entry.Status == "" {
		return audit.ErrInvalidInput
	}
	f.entries = append(f.entries, normalizeFakeCommandLog(entry, int64(len(f.entries)+1)))
	return nil
}

// List mengembalikan command log sesuai filter, terbaru lebih dulu.
func (f *FakeCommandLogRepo) List(_ context.Context, filter port.CommandLogFilter) ([]audit.CommandLog, int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	tenant := filter.TenantID
	if tenant == "" {
		tenant = fakeDefaultTenantID
	}
	out := make([]audit.CommandLog, 0)
	for _, entry := range f.entries {
		if entry.TenantID != tenant {
			continue
		}
		if filter.DeviceID != "" && entry.DeviceID != filter.DeviceID {
			continue
		}
		if filter.Status != "" && entry.Status != filter.Status {
			continue
		}
		out = append(out, entry)
	}
	sort.Slice(out, func(i, j int) bool {
		if !out[i].CreatedAt.Equal(out[j].CreatedAt) {
			return out[i].CreatedAt.After(out[j].CreatedAt)
		}
		return out[i].ID > out[j].ID
	})
	total := len(out)
	if filter.Offset > 0 {
		if filter.Offset >= len(out) {
			return []audit.CommandLog{}, total, nil
		}
		out = out[filter.Offset:]
	}
	if filter.Limit > 0 && filter.Limit < len(out) {
		out = out[:filter.Limit]
	}
	return out, total, nil
}

func normalizeFakeCommandLog(entry audit.CommandLog, nextID int64) audit.CommandLog {
	if entry.TenantID == "" {
		entry.TenantID = fakeDefaultTenantID
	}
	if entry.ID <= 0 {
		entry.ID = nextID
	}
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now().UTC()
	}
	return entry
}

// ─── ActiveDeviceStore ──────────────────────────────────────────────────

// FakeActiveDeviceStore adalah fake in-memory port.ActiveDeviceStore.
type FakeActiveDeviceStore struct {
	mu       sync.Mutex
	bindings map[string]port.ActiveDevice
}

var _ port.ActiveDeviceStore = (*FakeActiveDeviceStore)(nil)

// NewFakeActiveDeviceStore constructs an empty fake active device store.
func NewFakeActiveDeviceStore() *FakeActiveDeviceStore {
	return &FakeActiveDeviceStore{bindings: map[string]port.ActiveDevice{}}
}

// Seed menambahkan binding langsung (test helper).
func (f *FakeActiveDeviceStore) Seed(a port.ActiveDevice, expiresAt time.Time) {
	f.mu.Lock()
	defer f.mu.Unlock()
	a.UpdatedAt = expiresAt
	f.bindings[fakeActiveDeviceKey(a.TenantID, a.UserID, a.SessionID)] = a
}

// Set meng-upsert binding dengan masa berlaku now+ttl.
func (f *FakeActiveDeviceStore) Set(_ context.Context, a port.ActiveDevice, ttl time.Duration) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	now := time.Now().UTC()
	expiresAt := now.Add(100 * 365 * 24 * time.Hour)
	if ttl > 0 {
		expiresAt = now.Add(ttl)
	}
	a.TenantID = normalizeFakeTenant(a.TenantID)
	a.UpdatedAt = expiresAt
	f.bindings[fakeActiveDeviceKey(a.TenantID, a.UserID, a.SessionID)] = a
	return nil
}

// Get mengambil binding yang belum kedaluwarsa.
func (f *FakeActiveDeviceStore) Get(_ context.Context, tenantID, userID, sessionID string) (port.ActiveDevice, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	a, ok := f.bindings[fakeActiveDeviceKey(normalizeFakeTenant(tenantID), userID, sessionID)]
	if !ok || !a.UpdatedAt.After(time.Now().UTC()) {
		return port.ActiveDevice{}, ErrFakeNotFound
	}
	return a, nil
}

// Clear menghapus binding (idempoten).
func (f *FakeActiveDeviceStore) Clear(_ context.Context, tenantID, userID, sessionID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.bindings, fakeActiveDeviceKey(normalizeFakeTenant(tenantID), userID, sessionID))
	return nil
}

func normalizeFakeTenant(tenantID string) string {
	if tenantID == "" {
		return fakeDefaultTenantID
	}
	return tenantID
}

func fakeActiveDeviceKey(tenantID, userID, sessionID string) string {
	return tenantID + "/" + userID + "/" + sessionID
}

func (f *FakeCommandLogRepo) DeleteBefore(_ context.Context, cutoff time.Time) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	kept := f.entries[:0]
	deleted := 0
	for _, l := range f.entries {
		if l.CreatedAt.Before(cutoff) {
			deleted++
			continue
		}
		kept = append(kept, l)
	}
	f.entries = kept
	return deleted, nil
}
