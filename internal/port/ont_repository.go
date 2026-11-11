package port

import (
	"context"
	"time"

	"github.com/quixiq/polyglot/internal/domain/ont"
)

// ONTFilter filter daftar ONT (semua field opsional/kosong = tanpa filter).
type ONTFilter struct {
	OLTDeviceID string
	PONIndex    int
	State       string
	Search      string
	CustomerID  string
	Linked      *bool
	ACSMatched  *bool
	Limit       int
	Offset      int
}

// ONTStateCounts ringkasan jumlah ONT per state untuk satu OLT.
type ONTStateCounts struct {
	Total   int
	Online  int
	Offline int
	Unknown int
	Removed int
}

// ONTRepository adalah persistensi inventori ONT/ONU + timeline + sampel metrik.
type ONTRepository interface {
	// Save membuat/memperbarui satu ONT (upsert berdasarkan id).
	Save(ctx context.Context, o ont.ONT) error
	FindByID(ctx context.Context, id string) (ont.ONT, error)
	FindByOLTAndMAC(ctx context.Context, oltDeviceID, mac string) (ont.ONT, error)
	FindByOLTPosition(ctx context.Context, oltDeviceID string, ponIndex, onuIndex int) (ont.ONT, error)
	ListByOLT(ctx context.Context, oltDeviceID string) ([]ont.ONT, error)
	List(ctx context.Context, f ONTFilter) ([]ont.ONT, error)
	UpdateLink(ctx context.Context, id string, customerID, subscriptionID *string, pppoeUsername string) error
	UpdateACS(ctx context.Context, id string, acsDeviceID *string, matchSource string, lastInform *time.Time, telemetry []byte, wlanSlots int) error
	RecordEvent(ctx context.Context, e ont.Event) error
	ListEvents(ctx context.Context, ontID string, limit int) ([]ont.Event, error)
	AddSample(ctx context.Context, s ont.MetricSample) error
	ListSamples(ctx context.Context, ontID string, from time.Time, limit int) ([]ont.MetricSample, error)
	PurgeSamplesBefore(ctx context.Context, before time.Time) (int64, error)
	CountByState(ctx context.Context, oltDeviceID string) (ONTStateCounts, error)
}

// ACSDeviceFilter filter daftar cache device GenieACS.
type ACSDeviceFilter struct {
	Matched *bool
	Search  string
	Limit   int
	Offset  int
}

// ACSRepository adalah persistensi cache device GenieACS + preset yang dikelola.
type ACSRepository interface {
	UpsertDevice(ctx context.Context, d ont.ACSDevice) error
	ListDevices(ctx context.Context, f ACSDeviceFilter) ([]ont.ACSDevice, error)
	FindDeviceByACSID(ctx context.Context, acsDeviceID string) (ont.ACSDevice, error)
	UpdateDeviceMatch(ctx context.Context, acsDeviceID string, matchedONTID *string, source string) error
	SavePreset(ctx context.Context, p ont.ACSPreset) error
	GetPreset(ctx context.Context, name string) (ont.ACSPreset, error)
	ListPresets(ctx context.Context) ([]ont.ACSPreset, error)
}
