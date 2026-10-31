package network

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/domain/restorepoint"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/idgen"
)

// RestorePointConfig holds policy knobs used by the restore-point usecase.
type RestorePointConfig struct {
	TTL           time.Duration
	MaxBytes      int64
	RetentionDays int
	KeepLast      int
	Dedupe        bool
}

// RestorePointDeps wires the restore point usecase.
type RestorePointDeps struct {
	Repo      port.RestorePointRepository
	Storage   port.RestorePointStorage
	Devices   port.DeviceRepository
	Ops       func(ctx context.Context, deviceID string) (port.RouterOps, error)
	Approvals port.ApprovalRepository
	Config    RestorePointConfig
	Now       func() time.Time
}

// RestorePointUseCase creates, lists, and applies encrypted router backups.
type RestorePointUseCase struct {
	deps RestorePointDeps
}

// NewRestorePointUseCase builds the usecase. Ops may be nil in tests that only
// exercise metadata operations.
func NewRestorePointUseCase(deps RestorePointDeps) *RestorePointUseCase {
	if deps.Now == nil {
		deps.Now = func() time.Time { return time.Now().UTC() }
	}
	if deps.Config.TTL <= 0 {
		deps.Config.TTL = 14 * 24 * time.Hour
	}
	if deps.Config.MaxBytes <= 0 {
		deps.Config.MaxBytes = 50 * 1024 * 1024
	}
	return &RestorePointUseCase{deps: deps}
}

// CreateForMutation builds and persists the mandatory restore point for one
// proposed mutation. Any failure returns an error so no approval can be
// created (fail-closed).
func (uc *RestorePointUseCase) CreateForMutation(ctx context.Context, tenantID, deviceID, createdBy string, scopes []string, inverseScript []byte) (restorepoint.RestorePoint, error) {
	return uc.create(ctx, tenantID, deviceID, createdBy, restorepoint.TriggerMutation, scopes, inverseScript)
}

// CreateManual builds a restore point without a pending mutation.
func (uc *RestorePointUseCase) CreateManual(ctx context.Context, tenantID, deviceID, createdBy string) (restorepoint.RestorePoint, error) {
	return uc.create(ctx, tenantID, deviceID, createdBy, restorepoint.TriggerManual, nil, nil)
}

// CreateScheduled builds a scheduled backup (same artifacts, SYSTEM actor).
func (uc *RestorePointUseCase) CreateScheduled(ctx context.Context, tenantID, deviceID string) (restorepoint.RestorePoint, error) {
	return uc.create(ctx, tenantID, deviceID, "system:scheduler", restorepoint.TriggerScheduled, nil, nil)
}

// Cleanup removes expired bundles beyond the retention window while keeping
// the newest keepLast bundles per device. It never deletes a bundle that is
// still referenced by a PENDING/APPROVED/EXECUTING approval (caller supplies
// protectedIDs).
func (uc *RestorePointUseCase) Cleanup(ctx context.Context, now time.Time, keepLast int, protectedIDs map[string]bool) (deleted int, err error) {
	candidates, err := uc.deps.Repo.ListExpired(ctx, now, keepLast)
	if err != nil {
		return 0, err
	}
	for _, rp := range candidates {
		if protectedIDs[rp.ID] {
			continue
		}
		if _, derr := uc.deps.Repo.Delete(ctx, rp.ID); derr != nil {
			return deleted, derr
		}
		if uc.deps.Storage != nil {
			_ = uc.deps.Storage.DeleteBundle(ctx, rp.TenantID, rp.DeviceID, rp.ID)
		}
		deleted++
	}
	return deleted, nil
}

func (uc *RestorePointUseCase) create(ctx context.Context, tenantID, deviceID, createdBy string, trigger restorepoint.Trigger, scopes []string, inverseScript []byte) (restorepoint.RestorePoint, error) {
	if uc.deps.Repo == nil || uc.deps.Storage == nil {
		return restorepoint.RestorePoint{}, fmt.Errorf("%w: restore point dependencies not configured", restorepoint.ErrCreationFailed)
	}
	if tenantID == "" {
		tenantID = "tenant-default"
	}
	now := uc.deps.Now()
	rp := restorepoint.RestorePoint{
		ID:        idgen.New("rp"),
		TenantID:  tenantID,
		DeviceID:  deviceID,
		Trigger:   trigger,
		Scope:     normalizeScopes(scopes),
		Status:    restorepoint.StatusCreating,
		CreatedBy: createdBy,
		CreatedAt: now,
		UpdatedAt: now,
		ExpiresAt: now.Add(uc.deps.Config.TTL),
	}
	_ = rp

	ops, err := uc.opsFor(ctx, deviceID)
	if err != nil {
		return restorepoint.RestorePoint{}, fmt.Errorf("%w: %v", restorepoint.ErrCreationFailed, err)
	}
	version, freeBytes, err := ops.ResourceInfo(ctx)
	if err != nil {
		return restorepoint.RestorePoint{}, fmt.Errorf("%w: preflight: %v", restorepoint.ErrCreationFailed, err)
	}
	rp.RouterOSVersion = version
	if freeBytes > 0 && freeBytes < uc.deps.Config.MaxBytes {
		return restorepoint.RestorePoint{}, fmt.Errorf("%w: router free space %d bytes below cap %d", restorepoint.ErrCreationFailed, freeBytes, uc.deps.Config.MaxBytes)
	}

	raw, err := ops.CreateArtifacts(ctx, rp.ID, rp.Scope, inverseScript, uc.deps.Config.MaxBytes)
	if err != nil {
		return restorepoint.RestorePoint{}, fmt.Errorf("%w: %v", restorepoint.ErrCreationFailed, err)
	}

	var total int64
	saveArtifact := func(art port.RawArtifact) error {
		if art.Data != nil {
			defer func() { _ = art.Data.Close() }()
		}
		stored, serr := uc.deps.Storage.Save(ctx, tenantID, deviceID, rp.ID, art.Kind, art.Scope, art.Name, art.Data, uc.deps.Config.MaxBytes)
		if serr != nil {
			return fmt.Errorf("%w: store %s: %v", restorepoint.ErrCreationFailed, art.Kind, serr)
		}
		rp.Artifacts = append(rp.Artifacts, stored)
		total += stored.SizeBytes
		return nil
	}

	// 1) Scoped exports first: their hashes become the configuration hash used
	//    for binary-backup dedupe.
	var binaryRaw *port.RawArtifact
	var inverseRaw *port.RawArtifact
	for i := range raw {
		switch raw[i].Kind {
		case restorepoint.ArtifactBinaryBackup:
			binaryRaw = &raw[i]
		case restorepoint.ArtifactInverseScript:
			inverseRaw = &raw[i]
		default:
			if err := saveArtifact(raw[i]); err != nil {
				return restorepoint.RestorePoint{}, err
			}
		}
	}
	rp.ConfigHash = configHash(version, rp.Artifacts)

	// 2) Binary backup: reuse an existing identical artifact when dedupe is
	//    enabled and the mutation is not on a critical path. Critical menus
	//    always get a fresh backup by policy.
	if binaryRaw != nil {
		reused := false
		if uc.deps.Config.Dedupe && rp.ConfigHash != "" && !isCriticalScope(rp.Scope) {
			if prev, ferr := uc.deps.Repo.FindLatestByConfigHash(ctx, deviceID, rp.ConfigHash); ferr == nil {
				for _, old := range prev.ArtifactsOfKind(restorepoint.ArtifactBinaryBackup) {
					linked, lerr := uc.deps.Storage.LinkArtifact(ctx, tenantID, deviceID, rp.ID, old)
					if lerr == nil {
						_ = binaryRaw.Data.Close()
						rp.Artifacts = append(rp.Artifacts, linked)
						total += linked.SizeBytes
						reused = true
						break
					}
				}
			}
		}
		if !reused {
			if err := saveArtifact(*binaryRaw); err != nil {
				return restorepoint.RestorePoint{}, err
			}
		}
	}
	if inverseRaw != nil {
		if err := saveArtifact(*inverseRaw); err != nil {
			return restorepoint.RestorePoint{}, err
		}
	}
	rp.TotalSizeBytes = total

	manifest, _ := json.Marshal(map[string]any{
		"id": rp.ID, "device_id": deviceID, "trigger": trigger,
		"routeros_version": version, "scope": rp.Scope,
		"artifacts": rp.Artifacts, "total_size_bytes": total,
	})
	rp.Manifest = manifest
	if err := uc.deps.Storage.SaveManifest(ctx, tenantID, deviceID, rp.ID, json.RawMessage(manifest)); err != nil {
		return restorepoint.RestorePoint{}, fmt.Errorf("%w: manifest: %v", restorepoint.ErrCreationFailed, err)
	}
	rp.Status = restorepoint.StatusAvailable
	if err := uc.deps.Repo.Create(ctx, rp); err != nil {
		return restorepoint.RestorePoint{}, fmt.Errorf("%w: persist: %v", restorepoint.ErrCreationFailed, err)
	}
	return rp, nil
}

// List returns restore points for a device (or tenant when deviceID empty).
func (uc *RestorePointUseCase) List(ctx context.Context, f port.RestorePointFilter) ([]restorepoint.RestorePoint, int, error) {
	return uc.deps.Repo.List(ctx, f)
}

// Get returns one restore point bundle.
func (uc *RestorePointUseCase) Get(ctx context.Context, id string) (restorepoint.RestorePoint, error) {
	return uc.deps.Repo.FindByID(ctx, id)
}

// Delete soft-deletes a restore point and removes its encrypted files. It
// refuses bundles still referenced by a PENDING/APPROVED/EXECUTING approval.
func (uc *RestorePointUseCase) Delete(ctx context.Context, id string) error {
	rp, err := uc.deps.Repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if uc.deps.Approvals != nil {
		for _, status := range []string{"PENDING", "APPROVED", "EXECUTING"} {
			items, _, lerr := uc.deps.Approvals.List(ctx, port.ApprovalFilter{TenantID: rp.TenantID, Status: status, Limit: 500})
			if lerr != nil {
				return lerr
			}
			for _, a := range items {
				if a.RestorePointID == id {
					return fmt.Errorf("%w: restore point masih direferensikan approval %s (%s)", restorepoint.ErrNotAvailable, a.ID, a.Status)
				}
			}
		}
	}
	if _, derr := uc.deps.Repo.Delete(ctx, id); derr != nil {
		return derr
	}
	if uc.deps.Storage != nil {
		return uc.deps.Storage.DeleteBundle(ctx, rp.TenantID, rp.DeviceID, rp.ID)
	}
	return nil
}

// OpenArtifact returns a decrypted artifact stream for download.
func (uc *RestorePointUseCase) OpenArtifact(ctx context.Context, rpID string, kind restorepoint.ArtifactKind) (io.ReadCloser, restorepoint.Artifact, error) {
	rp, err := uc.deps.Repo.FindByID(ctx, rpID)
	if err != nil {
		return nil, restorepoint.Artifact{}, err
	}
	arts := rp.ArtifactsOfKind(kind)
	if len(arts) == 0 {
		return nil, restorepoint.Artifact{}, fmt.Errorf("%w: artifact %s missing", restorepoint.ErrNotAvailable, kind)
	}
	rc, err := uc.deps.Storage.Open(ctx, arts[0])
	if err != nil {
		return nil, restorepoint.Artifact{}, err
	}
	return rc, arts[0], nil
}

// Restore applies a restore point in the requested mode. The caller MUST be a
// human user; typed confirmation (device name) is verified here.
func (uc *RestorePointUseCase) Restore(ctx context.Context, rpID string, mode port.RestoreMode, confirmation string) (restorepoint.RestorePoint, error) {
	rp, err := uc.deps.Repo.FindByID(ctx, rpID)
	if err != nil {
		return restorepoint.RestorePoint{}, err
	}
	if !rp.Usable(uc.deps.Now()) {
		return restorepoint.RestorePoint{}, restorepoint.ErrNotAvailable
	}
	dev, err := uc.deps.Devices.FindByID(ctx, rp.DeviceID)
	if err != nil {
		return restorepoint.RestorePoint{}, fmt.Errorf("%w: device: %v", restorepoint.ErrNotAvailable, err)
	}
	if !confirmDevice(confirmation, dev) {
		return restorepoint.RestorePoint{}, fmt.Errorf("%w: typed device-name confirmation required", restorepoint.ErrNotAvailable)
	}
	ops, err := uc.opsFor(ctx, rp.DeviceID)
	if err != nil {
		return restorepoint.RestorePoint{}, err
	}

	switch mode {
	case port.RestoreInverse, port.RestoreScopedExport:
		kind := restorepoint.ArtifactInverseScript
		name := "polyglot-restore-inverse.rsc"
		if mode == port.RestoreScopedExport {
			kind = restorepoint.ArtifactScopedExport
			name = "polyglot-restore-scoped.rsc"
		}
		arts := rp.ArtifactsOfKind(kind)
		if len(arts) == 0 {
			return restorepoint.RestorePoint{}, fmt.Errorf("%w: %s artifact missing", restorepoint.ErrNotAvailable, kind)
		}
		rc, err := uc.deps.Storage.Open(ctx, arts[0])
		if err != nil {
			return restorepoint.RestorePoint{}, err
		}
		data, rerr := io.ReadAll(io.LimitReader(rc, uc.deps.Config.MaxBytes))
		_ = rc.Close()
		if rerr != nil {
			return restorepoint.RestorePoint{}, rerr
		}
		if err := ops.UploadFile(ctx, name, data); err != nil {
			return restorepoint.RestorePoint{}, err
		}
		defer ops.RemoveFile(ctx, name)
		if err := ops.ImportScript(ctx, name); err != nil {
			return restorepoint.RestorePoint{}, err
		}
	case port.RestoreBinary:
		arts := rp.ArtifactsOfKind(restorepoint.ArtifactBinaryBackup)
		if len(arts) == 0 {
			return restorepoint.RestorePoint{}, fmt.Errorf("%w: binary artifact missing", restorepoint.ErrNotAvailable)
		}
		rc, err := uc.deps.Storage.Open(ctx, arts[0])
		if err != nil {
			return restorepoint.RestorePoint{}, err
		}
		data, rerr := io.ReadAll(io.LimitReader(rc, uc.deps.Config.MaxBytes))
		_ = rc.Close()
		if rerr != nil {
			return restorepoint.RestorePoint{}, rerr
		}
		const maxUpload = 60000
		if len(data) > maxUpload {
			return restorepoint.RestorePoint{}, fmt.Errorf("%w: binary restore via API limited to %d bytes in v1; use manual file transfer", restorepoint.ErrSizeLimit, maxUpload)
		}
		if err := ops.UploadFile(ctx, "polyglot-restore.backup", data); err != nil {
			return restorepoint.RestorePoint{}, err
		}
		if err := ops.LoadBinary(ctx, "polyglot-restore.backup"); err != nil {
			return restorepoint.RestorePoint{}, err
		}
	default:
		return restorepoint.RestorePoint{}, fmt.Errorf("%w: unknown restore mode %q", restorepoint.ErrNotAvailable, mode)
	}

	now := uc.deps.Now()
	if err := uc.deps.Repo.UpdateStatus(ctx, rp.ID, restorepoint.StatusUsed, &now); err != nil {
		return restorepoint.RestorePoint{}, err
	}
	rp.Status = restorepoint.StatusUsed
	rp.UsedAt = &now
	return rp, nil
}

func (uc *RestorePointUseCase) opsFor(ctx context.Context, deviceID string) (port.RouterOps, error) {
	if uc.deps.Ops == nil {
		return nil, fmt.Errorf("%w: router ops resolver not configured", restorepoint.ErrCreationFailed)
	}
	return uc.deps.Ops(ctx, deviceID)
}

func confirmDevice(confirmation string, dev device.Device) bool {
	return confirmation != "" && (confirmation == dev.Name || confirmation == dev.ID)
}

func normalizeScopes(scopes []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(scopes))
	for _, s := range scopes {
		s = trimMenu(s)
		if s == "" || seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	if len(out) == 0 {
		out = []string{""}
	}
	return out
}

// configHash derives a stable hash from the router version and scoped export
// artifact hashes. Identical configuration => identical hash.
func configHash(version string, artifacts []restorepoint.Artifact) string {
	var parts []string
	parts = append(parts, version)
	for _, a := range artifacts {
		if a.Kind == restorepoint.ArtifactScopedExport {
			parts = append(parts, a.SHA256)
		}
	}
	if len(parts) <= 1 {
		return ""
	}
	sum := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return hex.EncodeToString(sum[:])
}

// isCriticalScope reports whether a mutation touches a policy-critical menu;
// those always require a fresh binary backup (dedupe disabled).
func isCriticalScope(scopes []string) bool {
	critical := []string{"system", "ip/service", "interface", "ip/firewall", "ip/route", "ipv6"}
	for _, scope := range scopes {
		s := trimMenu(scope)
		for _, c := range critical {
			if s == c || strings.HasPrefix(s, c+"/") {
				return true
			}
		}
	}
	return false
}

func trimMenu(s string) string {
	for len(s) > 0 && s[0] == '/' {
		s = s[1:]
	}
	for len(s) > 0 && s[len(s)-1] == '/' {
		s = s[:len(s)-1]
	}
	return s
}
