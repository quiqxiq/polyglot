// DEVIASI: fake ONTRepository + ACSRepository digabung dalam satu file
// sesuai kontrak tugas Fase A (fakes_ont.go), sehingga melewati batas 500
// baris; keduanya tetap terpisah jelas dengan penanda seksi.
package mocktest

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/quixiq/polyglot/pkg/idgen"

	domainONT "github.com/quixiq/polyglot/internal/domain/ont"
	"github.com/quixiq/polyglot/internal/port"
)

// fakeDefaultTenantID menjaga perilaku fake sama dengan repository Postgres.
const fakeDefaultTenantID = "tenant-default"

// ─── ONTRepository ──────────────────────────────────────────────────────

// FakeONTRepository adalah fake in-memory inventori ONT + timeline + sampel.
type FakeONTRepository struct {
	mu      sync.Mutex
	byID    map[string]domainONT.ONT
	events  []domainONT.Event
	samples []domainONT.MetricSample
}

var _ port.ONTRepository = (*FakeONTRepository)(nil)

// NewFakeONTRepository constructs an empty fake ONT repository.
func NewFakeONTRepository() *FakeONTRepository {
	return &FakeONTRepository{byID: map[string]domainONT.ONT{}}
}

// Seed menambahkan ONT langsung ke store (test helper).
func (f *FakeONTRepository) Seed(o domainONT.ONT) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.byID[o.ID] = o
}

// SeedEvent menambahkan entri timeline langsung (test helper).
func (f *FakeONTRepository) SeedEvent(e domainONT.Event) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.events = append(f.events, e)
}

// SeedSample menambahkan sampel metrik langsung (test helper).
func (f *FakeONTRepository) SeedSample(s domainONT.MetricSample) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.samples = append(f.samples, s)
}

// Save membuat/memperbarui ONT (upsert berdasarkan id).
func (f *FakeONTRepository) Save(_ context.Context, o domainONT.ONT) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if o.ID == "" {
		o.ID = idgen.New("ont")
	}
	f.byID[o.ID] = o
	return nil
}

// FindByID mencari ONT berdasarkan id.
func (f *FakeONTRepository) FindByID(_ context.Context, id string) (domainONT.ONT, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	o, ok := f.byID[id]
	if !ok {
		return domainONT.ONT{}, ErrFakeNotFound
	}
	return o, nil
}

// FindByOLTAndMAC mencari ONT berdasarkan OLT + MAC ternormalisasi.
func (f *FakeONTRepository) FindByOLTAndMAC(_ context.Context, oltDeviceID, mac string) (domainONT.ONT, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if norm, ok := domainONT.NormalizeMAC(mac); ok {
		mac = norm
	}
	for _, o := range f.byID {
		if o.OLTDeviceID == oltDeviceID && o.MACAddress == mac {
			return o, nil
		}
	}
	return domainONT.ONT{}, ErrFakeNotFound
}

// FindByOLTPosition mencari ONT berdasarkan posisi PON/ONU.
func (f *FakeONTRepository) FindByOLTPosition(_ context.Context, oltDeviceID string, ponIndex, onuIndex int) (domainONT.ONT, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, o := range f.byID {
		if o.OLTDeviceID == oltDeviceID && o.PONIndex == ponIndex && o.ONUIndex == onuIndex {
			return o, nil
		}
	}
	return domainONT.ONT{}, ErrFakeNotFound
}

// ListByOLT mengembalikan ONT satu OLT terurut posisi PON/ONU.
func (f *FakeONTRepository) ListByOLT(_ context.Context, oltDeviceID string) ([]domainONT.ONT, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]domainONT.ONT, 0)
	for _, o := range f.byID {
		if o.OLTDeviceID == oltDeviceID {
			out = append(out, o)
		}
	}
	sortONTs(out)
	return out, nil
}

// List mengembalikan ONT sesuai filter (semua field kosong = tanpa filter).
func (f *FakeONTRepository) List(_ context.Context, filter port.ONTFilter) ([]domainONT.ONT, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	search := strings.ToLower(strings.TrimSpace(filter.Search))
	out := make([]domainONT.ONT, 0)
	for _, o := range f.byID {
		if filter.OLTDeviceID != "" && o.OLTDeviceID != filter.OLTDeviceID {
			continue
		}
		if filter.PONIndex > 0 && o.PONIndex != filter.PONIndex {
			continue
		}
		if filter.State != "" && string(o.State) != filter.State {
			continue
		}
		if filter.Linked != nil && o.ISOLinked() != *filter.Linked {
			continue
		}
		if filter.ACSMatched != nil && (o.ACSDeviceID != nil) != *filter.ACSMatched {
			continue
		}
		if search != "" && !strings.Contains(strings.ToLower(o.MACAddress), search) &&
			!strings.Contains(strings.ToLower(o.SerialNumber), search) &&
			!strings.Contains(strings.ToLower(o.Model), search) &&
			!strings.Contains(strings.ToLower(o.PPPoEUsername), search) {
			continue
		}
		out = append(out, o)
	}
	sortONTs(out)
	if filter.Offset > 0 {
		if filter.Offset >= len(out) {
			return []domainONT.ONT{}, nil
		}
		out = out[filter.Offset:]
	}
	if filter.Limit > 0 && filter.Limit < len(out) {
		out = out[:filter.Limit]
	}
	return out, nil
}

// UpdateLink menautkan ONT ke pelanggan/langganan dan username PPPoE.
func (f *FakeONTRepository) UpdateLink(_ context.Context, id string, customerID, subscriptionID *string, pppoeUsername string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	o, ok := f.byID[id]
	if !ok {
		return ErrFakeNotFound
	}
	o.CustomerID = customerID
	o.SubscriptionID = subscriptionID
	o.PPPoEUsername = pppoeUsername
	f.byID[id] = o
	return nil
}

// UpdateACS memperbarui kolom integrasi ACS.
func (f *FakeONTRepository) UpdateACS(_ context.Context, id string, acsDeviceID *string, matchSource string, lastInform *time.Time, telemetry []byte, wlanSlots int) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	o, ok := f.byID[id]
	if !ok {
		return ErrFakeNotFound
	}
	o.ACSDeviceID = acsDeviceID
	o.ACSMatchSource = matchSource
	o.ACSLastInform = lastInform
	o.ACSTelemetry = append([]byte(nil), telemetry...)
	o.WLANSlots = wlanSlots
	f.byID[id] = o
	return nil
}

// RecordEvent menambahkan satu entri timeline ONT.
func (f *FakeONTRepository) RecordEvent(_ context.Context, e domainONT.Event) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if e.ID == "" {
		e.ID = idgen.New("ontev")
	}
	if e.TenantID == "" {
		e.TenantID = fakeDefaultTenantID
	}
	if e.CreatedAt.IsZero() {
		e.CreatedAt = time.Now()
	}
	f.events = append(f.events, e)
	return nil
}

// ListEvents mengembalikan timeline ONT terbaru lebih dulu; limit <= 0 tanpa batas.
func (f *FakeONTRepository) ListEvents(_ context.Context, ontID string, limit int) ([]domainONT.Event, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]domainONT.Event, 0)
	for _, e := range f.events {
		if e.ONTID == ontID {
			out = append(out, e)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	if limit > 0 && limit < len(out) {
		out = out[:limit]
	}
	return out, nil
}

// AddSample menambahkan satu sampel metrik optik ONT.
func (f *FakeONTRepository) AddSample(_ context.Context, s domainONT.MetricSample) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if s.ID == "" {
		s.ID = idgen.New("ontsmp")
	}
	if s.SampledAt.IsZero() {
		s.SampledAt = time.Now()
	}
	f.samples = append(f.samples, s)
	return nil
}

// ListSamples mengembalikan sampel terbaru lebih dulu; limit <= 0 tanpa batas.
func (f *FakeONTRepository) ListSamples(_ context.Context, ontID string, from time.Time, limit int) ([]domainONT.MetricSample, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]domainONT.MetricSample, 0)
	for _, s := range f.samples {
		if s.ONTID != ontID {
			continue
		}
		if !from.IsZero() && s.SampledAt.Before(from) {
			continue
		}
		out = append(out, s)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].SampledAt.After(out[j].SampledAt) })
	if limit > 0 && limit < len(out) {
		out = out[:limit]
	}
	return out, nil
}

// PurgeSamplesBefore menghapus sampel lebih lama dari `before`.
func (f *FakeONTRepository) PurgeSamplesBefore(_ context.Context, before time.Time) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	kept := f.samples[:0]
	var removed int64
	for _, s := range f.samples {
		if s.SampledAt.Before(before) {
			removed++
			continue
		}
		kept = append(kept, s)
	}
	f.samples = kept
	return removed, nil
}

// CountByState menghitung jumlah ONT per state; oltDeviceID kosong = semua OLT.
func (f *FakeONTRepository) CountByState(_ context.Context, oltDeviceID string) (port.ONTStateCounts, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var counts port.ONTStateCounts
	for _, o := range f.byID {
		if oltDeviceID != "" && o.OLTDeviceID != oltDeviceID {
			continue
		}
		counts.Total++
		switch o.State {
		case domainONT.StateOnline:
			counts.Online++
		case domainONT.StateOffline:
			counts.Offline++
		case domainONT.StateUnknown:
			counts.Unknown++
		case domainONT.StateRemoved:
			counts.Removed++
		}
	}
	return counts, nil
}

// sortONTs mengurutkan inventori sesuai repositori: OLT lalu posisi PON/ONU.
func sortONTs(list []domainONT.ONT) {
	sort.Slice(list, func(i, j int) bool {
		if list[i].OLTDeviceID != list[j].OLTDeviceID {
			return list[i].OLTDeviceID < list[j].OLTDeviceID
		}
		if list[i].PONIndex != list[j].PONIndex {
			return list[i].PONIndex < list[j].PONIndex
		}
		return list[i].ONUIndex < list[j].ONUIndex
	})
}

// ─── ACSRepository ──────────────────────────────────────────────────────

// FakeACSRepository adalah fake in-memory cache device GenieACS + preset.
type FakeACSRepository struct {
	mu      sync.Mutex
	byACSID map[string]domainONT.ACSDevice
	presets map[string]domainONT.ACSPreset // key: tenant + "/" + name
}

var _ port.ACSRepository = (*FakeACSRepository)(nil)

// NewFakeACSRepository constructs an empty fake ACS repository.
func NewFakeACSRepository() *FakeACSRepository {
	return &FakeACSRepository{
		byACSID: map[string]domainONT.ACSDevice{},
		presets: map[string]domainONT.ACSPreset{},
	}
}

// SeedDevice menambahkan device cache langsung (test helper).
func (f *FakeACSRepository) SeedDevice(d domainONT.ACSDevice) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.byACSID[d.ACSDeviceID] = d
}

// SeedPreset menambahkan preset langsung (test helper).
func (f *FakeACSRepository) SeedPreset(p domainONT.ACSPreset) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if p.TenantID == "" {
		p.TenantID = fakeDefaultTenantID
	}
	f.presets[p.TenantID+"/"+p.Name] = p
}

// UpsertDevice menyimpan cache device secara idempoten berdasarkan acs_device_id.
func (f *FakeACSRepository) UpsertDevice(_ context.Context, d domainONT.ACSDevice) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if d.TenantID == "" {
		d.TenantID = fakeDefaultTenantID
	}
	existing, ok := f.byACSID[d.ACSDeviceID]
	switch {
	case ok:
		if d.ID == "" {
			d.ID = existing.ID
		}
		if d.FirstSeenAt.IsZero() {
			d.FirstSeenAt = existing.FirstSeenAt
		}
		if d.LastSeenAt.IsZero() {
			d.LastSeenAt = existing.LastSeenAt
		}
		if d.LastInformAt == nil {
			d.LastInformAt = existing.LastInformAt
		}
		if d.MatchedONTID == nil {
			d.MatchedONTID = existing.MatchedONTID
			if d.MatchSource == "" {
				d.MatchSource = existing.MatchSource
			}
		}
	default:
		if d.ID == "" {
			d.ID = idgen.New("acsdev")
		}
		if d.FirstSeenAt.IsZero() {
			d.FirstSeenAt = time.Now()
		}
		if d.LastSeenAt.IsZero() {
			d.LastSeenAt = d.FirstSeenAt
		}
	}
	f.byACSID[d.ACSDeviceID] = d
	return nil
}

// ListDevices mengembalikan device cache sesuai filter.
func (f *FakeACSRepository) ListDevices(_ context.Context, filter port.ACSDeviceFilter) ([]domainONT.ACSDevice, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	search := strings.ToLower(strings.TrimSpace(filter.Search))
	out := make([]domainONT.ACSDevice, 0)
	for _, d := range f.byACSID {
		if filter.Matched != nil && (d.MatchedONTID != nil) != *filter.Matched {
			continue
		}
		if search != "" &&
			!strings.Contains(strings.ToLower(d.ACSDeviceID), search) &&
			!strings.Contains(strings.ToLower(d.OUI), search) &&
			!strings.Contains(strings.ToLower(d.ProductClass), search) &&
			!strings.Contains(strings.ToLower(d.SerialNumber), search) &&
			!strings.Contains(strings.ToLower(d.SoftwareVersion), search) {
			continue
		}
		out = append(out, d)
	}
	sort.Slice(out, func(i, j int) bool {
		if !out[i].LastSeenAt.Equal(out[j].LastSeenAt) {
			return out[i].LastSeenAt.After(out[j].LastSeenAt)
		}
		return out[i].ACSDeviceID < out[j].ACSDeviceID
	})
	if filter.Offset > 0 {
		if filter.Offset >= len(out) {
			return []domainONT.ACSDevice{}, nil
		}
		out = out[filter.Offset:]
	}
	if filter.Limit > 0 && filter.Limit < len(out) {
		out = out[:filter.Limit]
	}
	return out, nil
}

// FindDeviceByACSID mencari device berdasarkan id GenieACS.
func (f *FakeACSRepository) FindDeviceByACSID(_ context.Context, acsDeviceID string) (domainONT.ACSDevice, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	d, ok := f.byACSID[acsDeviceID]
	if !ok {
		return domainONT.ACSDevice{}, ErrFakeNotFound
	}
	return d, nil
}

// UpdateDeviceMatch menetapkan/melepas kaitan device ke ONT dan melepas
// device lain yang masih memegang ONT yang sama.
func (f *FakeACSRepository) UpdateDeviceMatch(_ context.Context, acsDeviceID string, matchedONTID *string, source string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.byACSID[acsDeviceID]; !ok {
		return ErrFakeNotFound
	}
	if matchedONTID != nil {
		for id, d := range f.byACSID {
			if id == acsDeviceID || d.MatchedONTID == nil {
				continue
			}
			if *d.MatchedONTID == *matchedONTID {
				d.MatchedONTID = nil
				d.MatchSource = ""
				f.byACSID[id] = d
			}
		}
	}
	d := f.byACSID[acsDeviceID]
	d.MatchedONTID = matchedONTID
	d.MatchSource = ""
	if matchedONTID != nil {
		d.MatchSource = source
	}
	f.byACSID[acsDeviceID] = d
	return nil
}

// SavePreset menyimpan preset secara idempoten berdasarkan (tenant, name).
func (f *FakeACSRepository) SavePreset(_ context.Context, p domainONT.ACSPreset) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if p.TenantID == "" {
		p.TenantID = fakeDefaultTenantID
	}
	key := p.TenantID + "/" + p.Name
	existing, ok := f.presets[key]
	switch {
	case ok:
		if p.ID == "" {
			p.ID = existing.ID
		}
		if p.CreatedAt.IsZero() {
			p.CreatedAt = existing.CreatedAt
		}
	default:
		if p.ID == "" {
			p.ID = idgen.New("acsp")
		}
	}
	f.presets[key] = p
	return nil
}

// GetPreset mencari preset berdasarkan name pada tenant default.
func (f *FakeACSRepository) GetPreset(_ context.Context, name string) (domainONT.ACSPreset, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	p, ok := f.presets[fakeDefaultTenantID+"/"+name]
	if !ok {
		return domainONT.ACSPreset{}, ErrFakeNotFound
	}
	return p, nil
}

// ListPresets mengembalikan seluruh preset terurut nama.
func (f *FakeACSRepository) ListPresets(_ context.Context) ([]domainONT.ACSPreset, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]domainONT.ACSPreset, 0, len(f.presets))
	for _, p := range f.presets {
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}
