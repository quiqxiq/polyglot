package postgres_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres"
	domainApproval "github.com/quixiq/polyglot/internal/domain/approval"
	domainAudit "github.com/quixiq/polyglot/internal/domain/audit"
	domainRestorePoint "github.com/quixiq/polyglot/internal/domain/restorepoint"
	"github.com/quixiq/polyglot/internal/port"
)

// newAIOpsTestDB membuka GORM ke PostgreSQL ephemeral yang seluruh migrasi SQL
// sudah diterapkan (pola startMigratedPostgres).
func newAIOpsTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	if testing.Short() {
		t.Skip("short mode")
	}
	_, _, dsn := startMigratedPostgres(t)
	db, err := gorm.Open(gormpostgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	return db
}

// seedAIOpsDevice membuat satu perangkat router untuk memenuhi FK device_id.
func seedAIOpsDevice(t *testing.T, db *gorm.DB, name, host string) string {
	t.Helper()
	var id string
	require.NoError(t, db.Raw(
		`INSERT INTO devices (name, vendor, driver_type, host) VALUES (?, 'mikrotik', 'mikrotik', ?) RETURNING id`,
		name, host).Scan(&id).Error)
	require.NotEmpty(t, id)
	return id
}

// TestAIOpsRepositories memverifikasi persistensi M1 PLAN-AI-ROUTER-OPS v2:
// approval HITL (CAS/claim/expire), restore point + artefak, command log, dan
// active device binding.
func TestAIOpsRepositories(t *testing.T) {
	db := newAIOpsTestDB(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Microsecond)
	deviceID := seedAIOpsDevice(t, db, "AI Router 1", "10.20.0.1")
	deviceID2 := seedAIOpsDevice(t, db, "AI Router 2", "10.20.0.2")

	var approvalID, restorePointID string

	t.Run("approval_lifecycle", func(t *testing.T) {
		repo := postgres.NewApprovalRepository(db)
		req := domainApproval.Request{
			ID:             "appr-test-1",
			TenantID:       "tenant-default",
			DeviceID:       deviceID,
			SessionID:      "sess-1",
			ActorType:      domainApproval.ActorAgent,
			ActorID:        "agent-1",
			RequestedBy:    "agent-1",
			Command:        "ip firewall filter add",
			Args:           map[string]string{"chain": "input", "action": "accept"},
			Revision:       1,
			Diff:           json.RawMessage(`{"before":[],"after":[{"chain":"input"}]}`),
			Status:         domainApproval.StatusPending,
			IdempotencyKey: "idem-1",
			CreatedAt:      now,
			ExpiresAt:      now.Add(10 * time.Minute),
		}
		require.NoError(t, repo.Create(ctx, req))
		approvalID = req.ID

		got, err := repo.FindByID(ctx, req.ID)
		require.NoError(t, err)
		assert.Equal(t, deviceID, got.DeviceID)
		assert.Equal(t, domainApproval.ActorAgent, got.ActorType)
		assert.Equal(t, req.Args, got.Args)
		assert.JSONEq(t, string(req.Diff), string(got.Diff))
		assert.Equal(t, domainApproval.StatusPending, got.Status)
		assert.Equal(t, 1, got.Revision)
		assert.Equal(t, "idem-1", got.IdempotencyKey)
		assert.WithinDuration(t, req.ExpiresAt, got.ExpiresAt, time.Microsecond)

		byKey, err := repo.FindByIdempotencyKey(ctx, "tenant-default", "idem-1")
		require.NoError(t, err)
		assert.Equal(t, req.ID, byKey.ID)
		_, err = repo.FindByIdempotencyKey(ctx, "tenant-default", "idem-missing")
		assert.ErrorIs(t, err, domainApproval.ErrNotFound)

		list, total, err := repo.List(ctx, port.ApprovalFilter{
			TenantID: "tenant-default", DeviceID: deviceID, Status: "PENDING",
		})
		require.NoError(t, err)
		assert.Equal(t, 1, total)
		require.Len(t, list, 1)
		assert.Equal(t, req.ID, list[0].ID)

		// CAS: expected revision salah → false tanpa error; revision benar → tersimpan.
		decidedAt := now.Add(time.Minute)
		decision := got
		decision.Status = domainApproval.StatusApproved
		decision.Revision = 2
		decision.DecidedAt = &decidedAt
		decision.Reason = "disetujui supervisor"
		ok, err := repo.UpdateIfRevision(ctx, decision, 99)
		require.NoError(t, err)
		assert.False(t, ok, "revision mismatch harus menghasilkan false")
		ok, err = repo.UpdateIfRevision(ctx, decision, 1)
		require.NoError(t, err)
		assert.True(t, ok)

		approved, err := repo.FindByID(ctx, req.ID)
		require.NoError(t, err)
		assert.Equal(t, domainApproval.StatusApproved, approved.Status)
		assert.Equal(t, 2, approved.Revision)
		assert.Equal(t, "disetujui supervisor", approved.Reason)
		require.NotNil(t, approved.DecidedAt)

		// Revision lama sudah tidak berlaku.
		ok, err = repo.UpdateIfRevision(ctx, decision, 1)
		require.NoError(t, err)
		assert.False(t, ok)

		// Claim: APPROVED → EXECUTING sekali saja.
		claimed, ok, err := repo.ClaimForExecution(ctx, req.ID, now)
		require.NoError(t, err)
		require.True(t, ok)
		assert.Equal(t, domainApproval.StatusExecuting, claimed.Status)
		assert.NotEmpty(t, claimed.ExecutionNonce)
		assert.Equal(t, 3, claimed.Revision)
		_, ok, err = repo.ClaimForExecution(ctx, req.ID, now)
		require.NoError(t, err)
		assert.False(t, ok, "claim kedua harus gagal")

		// ExpirePending hanya menyentuh PENDING yang lewat expires_at.
		expiredReq := req
		expiredReq.ID = "appr-expired-1"
		expiredReq.IdempotencyKey = ""
		expiredReq.CreatedAt = now.Add(-time.Hour)
		expiredReq.ExpiresAt = now.Add(-time.Minute)
		require.NoError(t, repo.Create(ctx, expiredReq))
		expiredCount, err := repo.ExpirePending(ctx, now)
		require.NoError(t, err)
		assert.Equal(t, 1, expiredCount)
		expired, err := repo.FindByID(ctx, expiredReq.ID)
		require.NoError(t, err)
		assert.Equal(t, domainApproval.StatusExpired, expired.Status)
		assert.Equal(t, 2, expired.Revision)

		_, err = repo.FindByID(ctx, "appr-missing")
		assert.ErrorIs(t, err, domainApproval.ErrNotFound)
	})

	t.Run("restore_point_bundle", func(t *testing.T) {
		repo := postgres.NewRestorePointRepository(db)
		rp := domainRestorePoint.RestorePoint{
			ID:              "rp-test-1",
			TenantID:        "tenant-default",
			DeviceID:        deviceID,
			Trigger:         domainRestorePoint.TriggerMutation,
			Scope:           []string{"/ip/firewall/filter", "/ip/address"},
			RouterOSVersion: "7.16.1",
			ConfigHash:      "hash-abc",
			Artifacts: []domainRestorePoint.Artifact{
				{
					Kind:        domainRestorePoint.ArtifactScopedExport,
					Scope:       []string{"ip/firewall"},
					StoragePath: "restore/rp-test-1/export.rsc",
					SHA256:      "aa11",
					SizeBytes:   1234,
				},
				{
					Kind:        domainRestorePoint.ArtifactBinaryBackup,
					StoragePath: "restore/rp-test-1/backup.backup",
					SHA256:      "bb22",
					SizeBytes:   5678,
				},
			},
			RollbackScript: json.RawMessage(`{"lines":["/ip firewall filter remove"]}`),
			Manifest:       json.RawMessage(`{"version":1,"files":2}`),
			TotalSizeBytes: 6912,
			Status:         domainRestorePoint.StatusAvailable,
			CreatedBy:      "user-1",
			CreatedAt:      now.Add(-2 * time.Hour),
			ExpiresAt:      now.Add(24 * time.Hour),
		}
		require.NoError(t, repo.Create(ctx, rp))
		restorePointID = rp.ID

		got, err := repo.FindByID(ctx, rp.ID)
		require.NoError(t, err)
		assert.Equal(t, deviceID, got.DeviceID)
		assert.Equal(t, domainRestorePoint.TriggerMutation, got.Trigger)
		assert.ElementsMatch(t, rp.Scope, got.Scope)
		assert.Equal(t, "hash-abc", got.ConfigHash)
		assert.JSONEq(t, string(rp.RollbackScript), string(got.RollbackScript))
		assert.JSONEq(t, string(rp.Manifest), string(got.Manifest))
		assert.Equal(t, int64(6912), got.TotalSizeBytes)
		require.Len(t, got.Artifacts, 2)
		assert.True(t, got.HasKind(domainRestorePoint.ArtifactScopedExport))
		assert.True(t, got.HasKind(domainRestorePoint.ArtifactBinaryBackup))
		export := got.ArtifactsOfKind(domainRestorePoint.ArtifactScopedExport)
		require.Len(t, export, 1)
		assert.Equal(t, "restore/rp-test-1/export.rsc", export[0].StoragePath)
		assert.Equal(t, "aa11", export[0].SHA256)
		assert.ElementsMatch(t, []string{"ip/firewall"}, export[0].Scope)

		latest, err := repo.FindLatestByConfigHash(ctx, deviceID, "hash-abc")
		require.NoError(t, err)
		assert.Equal(t, rp.ID, latest.ID)
		assert.Len(t, latest.Artifacts, 2)
		_, err = repo.FindLatestByConfigHash(ctx, deviceID, "hash-missing")
		assert.ErrorIs(t, err, domainRestorePoint.ErrNotFound)

		list, total, err := repo.List(ctx, port.RestorePointFilter{
			TenantID: "tenant-default", DeviceID: deviceID,
		})
		require.NoError(t, err)
		assert.Equal(t, 1, total)
		require.Len(t, list, 1)
		assert.Empty(t, list[0].Artifacts, "List hanya memuat header")

		usedAt := now.Add(-time.Minute)
		require.NoError(t, repo.UpdateStatus(ctx, rp.ID, domainRestorePoint.StatusUsed, &usedAt))
		used, err := repo.FindByID(ctx, rp.ID)
		require.NoError(t, err)
		assert.Equal(t, domainRestorePoint.StatusUsed, used.Status)
		require.NotNil(t, used.UsedAt)
		assert.WithinDuration(t, usedAt, *used.UsedAt, time.Microsecond)

		// Retensi: tiga bundle kedaluwarsa per device; keepLast=1 menyisakan
		// yang terbaru, dua terlamanya dikembalikan dari yang paling lama.
		for i := 0; i < 3; i++ {
			expired := domainRestorePoint.RestorePoint{
				ID:         fmt.Sprintf("rp-expired-%d", i),
				TenantID:   "tenant-default",
				DeviceID:   deviceID,
				Trigger:    domainRestorePoint.TriggerScheduled,
				Status:     domainRestorePoint.StatusAvailable,
				CreatedAt:  now.Add(time.Duration(i-3) * time.Hour),
				ExpiresAt:  now.Add(-time.Hour),
				ConfigHash: fmt.Sprintf("hash-old-%d", i),
				Artifacts: []domainRestorePoint.Artifact{{
					Kind:        domainRestorePoint.ArtifactInverseScript,
					StoragePath: fmt.Sprintf("restore/rp-expired-%d/inverse.rsc", i),
					SHA256:      "cc33",
					SizeBytes:   10,
				}},
			}
			require.NoError(t, repo.Create(ctx, expired))
		}
		expiredPoints, err := repo.ListExpired(ctx, now, 1)
		require.NoError(t, err)
		require.Len(t, expiredPoints, 2)
		assert.Equal(t, "rp-expired-0", expiredPoints[0].ID, "paling lama lebih dulu")
		assert.Equal(t, "rp-expired-1", expiredPoints[1].ID)
		assert.Len(t, expiredPoints[0].Artifacts, 1)

		allExpired, err := repo.ListExpired(ctx, now, 0)
		require.NoError(t, err)
		assert.Len(t, allExpired, 3)

		deleted, err := repo.Delete(ctx, rp.ID)
		require.NoError(t, err)
		assert.Equal(t, domainRestorePoint.StatusDeleted, deleted.Status)
		assert.Len(t, deleted.Artifacts, 2, "artefak dikembalikan untuk cleanup file")
		afterDelete, err := repo.FindByID(ctx, rp.ID)
		require.NoError(t, err)
		assert.Equal(t, domainRestorePoint.StatusDeleted, afterDelete.Status)

		_, err = repo.FindByID(ctx, "rp-missing")
		assert.ErrorIs(t, err, domainRestorePoint.ErrNotFound)
		err = repo.UpdateStatus(ctx, "rp-missing", domainRestorePoint.StatusExpired, nil)
		assert.ErrorIs(t, err, domainRestorePoint.ErrNotFound)
		_, err = repo.Delete(ctx, "rp-missing")
		assert.ErrorIs(t, err, domainRestorePoint.ErrNotFound)
	})

	t.Run("command_log", func(t *testing.T) {
		repo := postgres.NewCommandLogRepository(db)
		written := now.Add(-time.Minute)
		require.NoError(t, repo.Write(ctx, domainAudit.CommandLog{
			TenantID:       "tenant-default",
			DeviceID:       deviceID,
			ActorType:      "AGENT",
			ActorID:        "agent-1",
			SessionID:      "sess-1",
			Mode:           "MUTATION",
			Command:        "ip firewall filter add",
			CommandArgs:    json.RawMessage(`{"chain":"input"}`),
			ApprovalID:     approvalID,
			RestorePointID: restorePointID,
			Status:         "PENDING_APPROVAL",
			OutputSummary:  "menunggu approval",
			DurationMS:     12,
			CreatedAt:      written,
		}))
		require.NoError(t, repo.Write(ctx, domainAudit.CommandLog{
			TenantID:      "tenant-default",
			DeviceID:      deviceID,
			ActorType:     "USER",
			ActorID:       "user-1",
			Mode:          "READ",
			Command:       "/system resource print",
			Status:        "SUCCESS",
			OutputSummary: "ok",
			DurationMS:    7,
			CreatedAt:     now,
		}))
		require.ErrorIs(t, repo.Write(ctx, domainAudit.CommandLog{Mode: "READ"}), domainAudit.ErrInvalidInput)

		entries, total, err := repo.List(ctx, port.CommandLogFilter{
			TenantID: "tenant-default", DeviceID: deviceID,
		})
		require.NoError(t, err)
		assert.Equal(t, 2, total)
		require.Len(t, entries, 2)
		assert.Equal(t, "SUCCESS", entries[0].Status, "terbaru lebih dulu")
		assert.Equal(t, "/system resource print", entries[0].Command)
		assert.Positive(t, entries[0].ID)
		assert.Equal(t, "PENDING_APPROVAL", entries[1].Status)
		assert.JSONEq(t, `{"chain":"input"}`, string(entries[1].CommandArgs))
		assert.Equal(t, approvalID, entries[1].ApprovalID)
		assert.Equal(t, restorePointID, entries[1].RestorePointID)
		assert.Equal(t, int64(12), entries[1].DurationMS)

		failed, failedTotal, err := repo.List(ctx, port.CommandLogFilter{
			TenantID: "tenant-default", Status: "SUCCESS",
		})
		require.NoError(t, err)
		assert.Equal(t, 1, failedTotal)
		require.Len(t, failed, 1)
	})

	t.Run("active_device", func(t *testing.T) {
		store := postgres.NewActiveDeviceRepository(db)
		binding := port.ActiveDevice{
			TenantID:  "tenant-default",
			UserID:    "user-1",
			SessionID: "sess-1",
			DeviceID:  deviceID,
		}
		require.NoError(t, store.Set(ctx, binding, time.Hour))
		got, err := store.Get(ctx, "tenant-default", "user-1", "sess-1")
		require.NoError(t, err)
		assert.Equal(t, deviceID, got.DeviceID)

		// Upsert: binding (tenant, user, session) yang sama diperbarui, bukan
		// menambah baris.
		binding.DeviceID = deviceID2
		require.NoError(t, store.Set(ctx, binding, time.Hour))
		got, err = store.Get(ctx, "tenant-default", "user-1", "sess-1")
		require.NoError(t, err)
		assert.Equal(t, deviceID2, got.DeviceID)
		var rows int64
		require.NoError(t, db.Table("active_devices").Count(&rows).Error)
		assert.Equal(t, int64(1), rows)

		// TTL: binding kedaluwarsa tidak lagi dikembalikan.
		require.NoError(t, store.Set(ctx, port.ActiveDevice{
			TenantID: "tenant-default", UserID: "user-2", SessionID: "sess-2", DeviceID: deviceID,
		}, time.Millisecond))
		time.Sleep(10 * time.Millisecond)
		_, err = store.Get(ctx, "tenant-default", "user-2", "sess-2")
		assert.ErrorIs(t, err, postgres.ErrNotFound)

		require.NoError(t, store.Clear(ctx, "tenant-default", "user-1", "sess-1"))
		_, err = store.Get(ctx, "tenant-default", "user-1", "sess-1")
		assert.ErrorIs(t, err, postgres.ErrNotFound)
		require.NoError(t, store.Clear(ctx, "tenant-default", "user-1", "sess-1"), "clear idempoten")
	})
}
