package restorepoint

import (
	"encoding/json"
	"time"
)

// Trigger describes why a restore point was created.
type Trigger string

const (
	TriggerMutation  Trigger = "MUTATION"
	TriggerScheduled Trigger = "SCHEDULED"
	TriggerManual    Trigger = "MANUAL"
)

// ArtifactKind is one file inside a restore point bundle.
type ArtifactKind string

const (
	ArtifactScopedExport  ArtifactKind = "SCOPED_EXPORT"
	ArtifactBinaryBackup  ArtifactKind = "BINARY_BACKUP"
	ArtifactInverseScript ArtifactKind = "INVERSE_SCRIPT"
)

// Status is the restore point lifecycle state.
type Status string

const (
	StatusCreating  Status = "CREATING"
	StatusAvailable Status = "AVAILABLE"
	StatusPartial   Status = "PARTIAL"
	StatusUsed      Status = "USED"
	StatusExpired   Status = "EXPIRED"
	StatusDeleted   Status = "DELETED"
)

// Artifact is one encrypted file in a bundle.
type Artifact struct {
	Kind        ArtifactKind
	Scope       []string
	StoragePath string
	SHA256      string
	SizeBytes   int64
}

// RestorePoint is a recoverable bundle for one device mutation/backup.
type RestorePoint struct {
	ID              string
	TenantID        string
	DeviceID        string
	Trigger         Trigger
	Scope           []string
	RouterOSVersion string
	ConfigHash      string
	Artifacts       []Artifact
	RollbackScript  json.RawMessage
	Manifest        json.RawMessage
	TotalSizeBytes  int64
	Status          Status
	CreatedBy       string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	ExpiresAt       time.Time
	UsedAt          *time.Time
}

// HasKind reports whether the bundle contains at least one artifact of kind.
func (rp RestorePoint) HasKind(kind ArtifactKind) bool {
	for _, a := range rp.Artifacts {
		if a.Kind == kind {
			return true
		}
	}
	return false
}

// ArtifactsOfKind returns all artifacts of kind.
func (rp RestorePoint) ArtifactsOfKind(kind ArtifactKind) []Artifact {
	out := make([]Artifact, 0, len(rp.Artifacts))
	for _, a := range rp.Artifacts {
		if a.Kind == kind {
			out = append(out, a)
		}
	}
	return out
}

// Usable reports whether this restore point may back a mutation approval.
func (rp RestorePoint) Usable(now time.Time) bool {
	if rp.Status != StatusAvailable && rp.Status != StatusUsed {
		return false
	}
	return rp.ExpiresAt.IsZero() || now.Before(rp.ExpiresAt)
}
