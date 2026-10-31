// Package routerjob memodelkan antrean job asynchronous per perangkat router
// (tabel router_jobs, planning v2 §3.5): provisioning, isolir/restore,
// hapus akun, dan sinkronisasi profil dengan retry exponential backoff.
package routerjob

import (
	"encoding/json"
	"time"
)

// Status job dalam antrean.
type Status string

const (
	StatusQueued    Status = "QUEUED"
	StatusRunning   Status = "RUNNING"
	StatusSucceeded Status = "SUCCEEDED"
	StatusFailed    Status = "FAILED"
	StatusRetryWait Status = "RETRY_WAIT"
	StatusDead      Status = "DEAD"
)

// Type jenis pekerjaan router.
type Type string

const (
	TypeProvisionPPPoE     Type = "PROVISION_PPPOE"
	TypeProvisionHotspot   Type = "PROVISION_HOTSPOT"
	TypeProvisionDedicated Type = "PROVISION_DEDICATED"
	TypeIsolate            Type = "ISOLATE"
	TypeRestore            Type = "RESTORE"
	TypeDeleteAccount      Type = "DELETE_ACCOUNT"
	TypeSyncProfile        Type = "SYNC_PROFILE"

	// Fase B/C PLAN-ACS-OLT-HIOSO.md — OLT/ONT.
	TypeSyncOLT       Type = "SYNC_OLT"
	TypeONTRename     Type = "ONT_RENAME"
	TypeONTReboot     Type = "ONT_REBOOT"
	TypeONTSetAdmin   Type = "ONT_SET_ADMIN"
	TypeONTDelete     Type = "ONT_DELETE"
	TypeOLTSaveConfig Type = "OLT_SAVE_CONFIG"

	// Fase D — GenieACS TR-069.
	TypeACSReboot           Type = "ACS_REBOOT"
	TypeACSSetWiFi          Type = "ACS_SET_WIFI"
	TypeACSPushPPPoE        Type = "ACS_PUSH_PPPOE"
	TypeACSRefreshTelemetry Type = "ACS_REFRESH_TELEMETRY"
	TypeACSSyncMatch        Type = "ACS_SYNC_MATCH"
	TypeONTBind             Type = "ONT_BIND"
	TypeACSBootstrapPreset  Type = "ACS_BOOTSTRAP_PRESET"

	// Billing Ops v3 (PLAN-BILLING-OPS-V3).
	TypeMassPayment Type = "MASS_PAYMENT"
)

// Job adalah satu entri antrean job router.
type Job struct {
	ID             string          `json:"id"`
	TenantID       string          `json:"tenant_id"`
	DeviceID       string          `json:"device_id"`
	JobType        Type            `json:"job_type"`
	Payload        json.RawMessage `json:"payload"`
	IdempotencyKey string          `json:"idempotency_key"`
	Status         Status          `json:"status"`
	Attempts       int             `json:"attempts"`
	MaxAttempts    int             `json:"max_attempts"`
	NextRunAt      time.Time       `json:"next_run_at"`
	LastError      string          `json:"last_error,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

// CanRetry melaporkan apakah job masih boleh dicoba ulang.
func (j Job) CanRetry() bool {
	max := j.MaxAttempts
	if max <= 0 {
		max = 5
	}
	return j.Attempts < max
}

// Backoff menghitung jeda percobaan berikutnya: 30s, 1m, 2m, 4m, 8m, ...
// dengan batas maksimum 30 menit (attempts sudah termasuk percobaan terakhir).
func Backoff(attempts int) time.Duration {
	if attempts < 1 {
		attempts = 1
	}
	d := 30 * time.Second
	for i := 1; i < attempts; i++ {
		d *= 2
		if d >= 30*time.Minute {
			return 30 * time.Minute
		}
	}
	return d
}

// ValidType melaporkan apakah tipe job dikenal.
func ValidType(t Type) bool {
	switch t {
	case TypeProvisionPPPoE, TypeProvisionHotspot, TypeProvisionDedicated,
		TypeIsolate, TypeRestore, TypeDeleteAccount, TypeSyncProfile,
		TypeSyncOLT, TypeONTRename, TypeONTReboot, TypeONTSetAdmin,
		TypeONTDelete, TypeOLTSaveConfig,
		TypeACSReboot, TypeACSSetWiFi, TypeACSPushPPPoE, TypeACSRefreshTelemetry,
		TypeACSSyncMatch, TypeONTBind, TypeACSBootstrapPreset,
		TypeMassPayment:
		return true
	default:
		return false
	}
}
