package port

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/olt"
)

// OLTHealth ringkasan kesehatan/identitas OLT hasil probe.
type OLTHealth struct {
	Reachable   bool
	Variant     olt.Variant
	SysDescr    string
	SysObjectID string
	Firmware    string
	Uptime      string
	LatencyMS   int64
}

// OLTONU satu baris ONU hasil pembacaan OLT.
type OLTONU struct {
	PONIndex    int
	ONUIndex    int
	LLID        int
	MACAddress  string
	Name        string
	Model       string
	Firmware    string
	LinkUp      bool
	AdminEnable bool
	RxPowerDBM  *float64
	TxPowerDBM  *float64
	DistanceM   *int
}

// OLTManager adalah operasi terstruktur ke satu OLT EPON (read SNMP + write CLI).
// Implementasi wajib serialisasi akses per perangkat (sesi CLI terbatas).
type OLTManager interface {
	// Detect mengembalikan capability varian + kesehatan OLT.
	Detect(ctx context.Context) (olt.Capabilities, OLTHealth, error)
	// ListONUs membaca inventori ONU; ponIndex 0 = semua PON.
	ListONUs(ctx context.Context, ponIndex int) ([]OLTONU, error)
	RenameONU(ctx context.Context, ponIndex, onuIndex int, name string) error
	RebootONU(ctx context.Context, ponIndex, onuIndex int) error
	SetONUAdmin(ctx context.Context, ponIndex, onuIndex int, enabled bool) error
	DeleteONU(ctx context.Context, ponIndex, onuIndex int) error
	SaveConfig(ctx context.Context) error
	RunningConfig(ctx context.Context) (string, error)
	SystemInfo(ctx context.Context) (map[string]string, error)
}

// OLTManagerProvider membangun OLTManager yang sudah terikat ke satu OLT
// (device inventory). Implementasi biasanya memakai registry.Get + type
// assertion ke driver vendor, sehingga usecase tidak pernah menyentuh driver.
type OLTManagerProvider interface {
	ForDevice(ctx context.Context, oltDeviceID string) (OLTManager, error)
}
