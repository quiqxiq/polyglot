package storage_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quixiq/polyglot/internal/adapter/storage"
	"github.com/quixiq/polyglot/internal/domain/restorepoint"
	cryptoutil "github.com/quixiq/polyglot/pkg/crypto"
)

const (
	testStorageKey = "01234567890123456789012345678901"
	testTenantID   = "tenant-default"
	testDeviceID   = "dev-0001"
	testRPID       = "rp-0001"
)

// newRestorePointStorage membuat storage pada direktori temporer test.
func newRestorePointStorage(t *testing.T) (*storage.RestorePointStorage, string) {
	t.Helper()
	root := filepath.Join(t.TempDir(), "restorepoints")
	st, err := storage.NewRestorePointStorage(root, testStorageKey)
	require.NoError(t, err)
	return st, root
}

// saveTestArtifact menyimpan satu artifact kecil untuk dipakai lintas test.
func saveTestArtifact(t *testing.T, st *storage.RestorePointStorage, kind restorepoint.ArtifactKind, name string, payload []byte) restorepoint.Artifact {
	t.Helper()
	artifact, err := st.Save(context.Background(), testTenantID, testDeviceID, testRPID,
		kind, []string{"/ip/firewall"}, name, bytes.NewReader(payload), 1<<20)
	require.NoError(t, err)
	return artifact
}

// assertFileMode memverifikasi permission file/direktori. Permission hanya
// deterministik di Linux (umask Windows/macOS berbeda).
func assertFileMode(t *testing.T, path string, want os.FileMode) {
	t.Helper()
	if runtime.GOOS != "linux" {
		t.Skip("permission test hanya di Linux")
	}
	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, want, info.Mode().Perm(), "mode %s", path)
}

func TestNewRestorePointStorage_Validation(t *testing.T) {
	t.Run("root kosong ditolak", func(t *testing.T) {
		_, err := storage.NewRestorePointStorage("   ", testStorageKey)
		require.Error(t, err)
	})

	t.Run("key bukan 32 byte ditolak", func(t *testing.T) {
		_, err := storage.NewRestorePointStorage(t.TempDir(), "too-short")
		assert.ErrorIs(t, err, cryptoutil.ErrInvalidKeyLength)
	})

	t.Run("key base64 32 byte diterima", func(t *testing.T) {
		root := filepath.Join(t.TempDir(), "restorepoints")
		encodedKey := base64.StdEncoding.EncodeToString([]byte(testStorageKey))
		stB64, err := storage.NewRestorePointStorage(root, encodedKey)
		require.NoError(t, err)

		artifact := saveTestArtifact(t, stB64, restorepoint.ArtifactScopedExport, "scoped.rsc", []byte("hostname polyglot"))
		rc, err := stB64.Open(context.Background(), artifact)
		require.NoError(t, err)
		defer func() { _ = rc.Close() }()
		got, err := io.ReadAll(rc)
		require.NoError(t, err)
		assert.Equal(t, "hostname polyglot", string(got))

		// Storage dengan key raw yang sama harus bisa membuka artifact.
		stRaw, err := storage.NewRestorePointStorage(root, testStorageKey)
		require.NoError(t, err)
		rawRC, err := stRaw.Open(context.Background(), artifact)
		require.NoError(t, err)
		defer func() { _ = rawRC.Close() }()
		raw, err := io.ReadAll(rawRC)
		require.NoError(t, err)
		assert.Equal(t, "hostname polyglot", string(raw))
	})

	t.Run("root dibuat 0700", func(t *testing.T) {
		root := filepath.Join(t.TempDir(), "restorepoints")
		_, err := storage.NewRestorePointStorage(root, testStorageKey)
		require.NoError(t, err)
		assertFileMode(t, root, 0o700)
	})
}

func TestSaveOpenVerify_RoundTrip(t *testing.T) {
	st, root := newRestorePointStorage(t)
	ctx := context.Background()
	payload := []byte("hostname polyglot\r\n/ip firewall filter print detail\r\n")

	artifact := saveTestArtifact(t, st, restorepoint.ArtifactScopedExport, "scoped-ip-firewall.rsc", payload)

	expectedPath := filepath.ToSlash(filepath.Join(testTenantID, testDeviceID, testRPID,
		"artifacts", "scoped-ip-firewall.rsc.enc"))
	assert.Equal(t, restorepoint.ArtifactScopedExport, artifact.Kind)
	assert.Equal(t, []string{"/ip/firewall"}, artifact.Scope)
	assert.Equal(t, expectedPath, artifact.StoragePath)
	assert.Equal(t, int64(len(payload)), artifact.SizeBytes)
	sum := sha256.Sum256(payload)
	assert.Equal(t, hex.EncodeToString(sum[:]), artifact.SHA256)

	fullPath := filepath.Join(root, filepath.FromSlash(artifact.StoragePath))
	encrypted, err := os.ReadFile(fullPath)
	require.NoError(t, err)
	assert.Len(t, encrypted, len(payload)+12+16, "nonce + GCM tag")
	assert.False(t, bytes.Contains(encrypted, payload), "file at-rest tidak boleh berisi plaintext")
	assertFileMode(t, fullPath, 0o600)
	assertFileMode(t, filepath.Dir(fullPath), 0o700)

	rc, err := st.Open(ctx, artifact)
	require.NoError(t, err)
	got, err := io.ReadAll(rc)
	require.NoError(t, err)
	require.NoError(t, rc.Close())
	assert.Equal(t, payload, got)

	require.NoError(t, st.Verify(ctx, artifact))
	info, err := st.Stat(ctx, artifact)
	require.NoError(t, err)
	assert.Equal(t, int64(len(encrypted)), info.Size())

	t.Run("hash tidak cocok", func(t *testing.T) {
		corrupt := artifact
		corrupt.SHA256 = strings.Repeat("0", 64)
		assert.Error(t, st.Verify(ctx, corrupt))
	})

	t.Run("ukuran tidak cocok", func(t *testing.T) {
		corrupt := artifact
		corrupt.SizeBytes++
		assert.Error(t, st.Verify(ctx, corrupt))
	})

	t.Run("file di disk korup", func(t *testing.T) {
		tampered := append([]byte(nil), encrypted...)
		tampered[len(tampered)-1] ^= 0xff
		require.NoError(t, os.WriteFile(fullPath, tampered, 0o600))
		assert.Error(t, st.Verify(ctx, artifact))
		_, openErr := st.Open(ctx, artifact)
		assert.Error(t, openErr, "GCM harus menolak file yang diubah")
	})

	t.Run("context dibatalkan", func(t *testing.T) {
		cancelled, cancel := context.WithCancel(ctx)
		cancel()
		_, openErr := st.Open(cancelled, artifact)
		assert.ErrorIs(t, openErr, context.Canceled)
	})
}

func TestSave_DefaultArtifactName(t *testing.T) {
	st, _ := newRestorePointStorage(t)
	artifact := saveTestArtifact(t, st, restorepoint.ArtifactBinaryBackup, "", []byte("binary-backup"))
	assert.True(t, strings.HasSuffix(artifact.StoragePath, "/artifacts/binary.backup.enc"), artifact.StoragePath)
}

func TestSave_SizeLimit(t *testing.T) {
	st, _ := newRestorePointStorage(t)
	ctx := context.Background()

	_, err := st.Save(ctx, testTenantID, testDeviceID, testRPID, restorepoint.ArtifactBinaryBackup,
		nil, "oversize.backup", bytes.NewReader(make([]byte, 9)), 8)
	assert.ErrorIs(t, err, restorepoint.ErrSizeLimit)

	artifact, err := st.Save(ctx, testTenantID, testDeviceID, testRPID, restorepoint.ArtifactBinaryBackup,
		nil, "exact.backup", bytes.NewReader(make([]byte, 8)), 8)
	require.NoError(t, err)
	assert.Equal(t, int64(8), artifact.SizeBytes)

	// maxBytes <= 0 berarti tanpa batas.
	unbounded, err := st.Save(ctx, testTenantID, testDeviceID, testRPID, restorepoint.ArtifactBinaryBackup,
		nil, "unbounded.backup", bytes.NewReader(make([]byte, 32)), 0)
	require.NoError(t, err)
	assert.Equal(t, int64(32), unbounded.SizeBytes)
}

func TestSave_RejectsUnsafeInput(t *testing.T) {
	st, root := newRestorePointStorage(t)
	ctx := context.Background()
	parent := filepath.Dir(root)

	cases := []struct {
		name         string
		tenantID     string
		deviceID     string
		restorePoint string
		artifactName string
	}{
		{"tenant traversal", "../escape", testDeviceID, testRPID, "a.rsc"},
		{"tenant kosong", "", testDeviceID, testRPID, "a.rsc"},
		{"device slash", testTenantID, "dev/other", testRPID, "a.rsc"},
		{"device dotdot", testTenantID, "..", testRPID, "a.rsc"},
		{"restore point dotdot", testTenantID, testDeviceID, "..", "a.rsc"},
		{"artifact traversal", testTenantID, testDeviceID, testRPID, "../escape"},
		{"artifact backslash", testTenantID, testDeviceID, testRPID, `a\b`},
		{"artifact absolute", testTenantID, testDeviceID, testRPID, "/etc/passwd"},
		{"artifact kosong", testTenantID, testDeviceID, testRPID, " "},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := st.Save(ctx, tc.tenantID, tc.deviceID, tc.restorePoint,
				restorepoint.ArtifactScopedExport, nil, tc.artifactName, bytes.NewReader([]byte("x")), 1024)
			assert.Error(t, err)
		})
	}

	t.Run("kind tidak dikenal", func(t *testing.T) {
		_, err := st.Save(ctx, testTenantID, testDeviceID, testRPID,
			restorepoint.ArtifactKind("BOGUS"), nil, "x.rsc", bytes.NewReader([]byte("x")), 1024)
		assert.Error(t, err)
	})

	assert.NoDirExists(t, filepath.Join(parent, "escape"), "tidak boleh ada direktori di luar root")
}

func TestOpen_RejectsTraversalPath(t *testing.T) {
	st, _ := newRestorePointStorage(t)
	ctx := context.Background()

	for _, storagePath := range []string{
		"",
		"..",
		"../escape.enc",
		"a/../../escape.enc",
		"/etc/passwd",
	} {
		_, err := st.Open(ctx, restorepoint.Artifact{StoragePath: storagePath})
		assert.Error(t, err, "path %q harus ditolak", storagePath)
	}
}

func TestOpen_WrongKey(t *testing.T) {
	stA, root := newRestorePointStorage(t)
	artifact := saveTestArtifact(t, stA, restorepoint.ArtifactScopedExport, "scoped.rsc", []byte("secret config"))

	stB, err := storage.NewRestorePointStorage(root, "abcdefghijklmnopqrstuvwxyz012345")
	require.NoError(t, err)
	_, err = stB.Open(context.Background(), artifact)
	require.Error(t, err, "key berbeda harus gagal dekripsi")
	assert.Error(t, stB.Verify(context.Background(), artifact))
}

func TestSaveManifest_AndTotalSize(t *testing.T) {
	st, root := newRestorePointStorage(t)
	ctx := context.Background()

	// Bundle belum ada: total size 0 tanpa error.
	total, err := st.TotalSize(ctx, testTenantID, testDeviceID, "rp-empty")
	require.NoError(t, err)
	assert.Zero(t, total)

	first := saveTestArtifact(t, st, restorepoint.ArtifactScopedExport, "scoped.rsc", []byte("abcde"))
	second := saveTestArtifact(t, st, restorepoint.ArtifactBinaryBackup, "binary.backup", []byte("abcdefg"))

	// TotalSize menghitung ukuran file terenkripsi di disk (plaintext + nonce
	// + GCM tag), bukan total plaintext.
	onDiskSize := func(a restorepoint.Artifact) int64 {
		info, statErr := os.Stat(filepath.Join(root, filepath.FromSlash(a.StoragePath)))
		require.NoError(t, statErr)
		return info.Size()
	}
	firstOnDisk, secondOnDisk := onDiskSize(first), onDiskSize(second)
	require.Greater(t, firstOnDisk, first.SizeBytes)
	require.Greater(t, secondOnDisk, second.SizeBytes)

	total, err = st.TotalSize(ctx, testTenantID, testDeviceID, testRPID)
	require.NoError(t, err)
	assert.Equal(t, firstOnDisk+secondOnDisk, total)

	manifest := map[string]any{
		"schema":           1,
		"tenant_id":        testTenantID,
		"device_id":        testDeviceID,
		"restore_point_id": testRPID,
		"artifacts": []map[string]any{
			{"kind": first.Kind, "storage_path": first.StoragePath, "sha256": first.SHA256, "size_bytes": first.SizeBytes},
			{"kind": second.Kind, "storage_path": second.StoragePath, "sha256": second.SHA256, "size_bytes": second.SizeBytes},
		},
	}
	require.NoError(t, st.SaveManifest(ctx, testTenantID, testDeviceID, testRPID, manifest))

	manifestPath := filepath.Join(root, testTenantID, testDeviceID, testRPID, "manifest.json")
	raw, err := os.ReadFile(manifestPath)
	require.NoError(t, err)
	assertFileMode(t, manifestPath, 0o600)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal(raw, &decoded))
	assert.Equal(t, float64(1), decoded["schema"])
	assert.Equal(t, testTenantID, decoded["tenant_id"])
	assert.NotContains(t, strings.ToLower(string(raw)), "password", "manifest tidak boleh berisi credential")

	totalAfterManifest, err := st.TotalSize(ctx, testTenantID, testDeviceID, testRPID)
	require.NoError(t, err)
	assert.Equal(t, firstOnDisk+secondOnDisk+int64(len(raw)), totalAfterManifest)

	// Identifier tidak aman ditolak.
	assert.Error(t, st.SaveManifest(ctx, testTenantID, testDeviceID, "../escape", manifest))
}

func TestDeleteBundle(t *testing.T) {
	st, root := newRestorePointStorage(t)
	ctx := context.Background()

	saveTestArtifact(t, st, restorepoint.ArtifactScopedExport, "scoped.rsc", []byte("data"))
	require.NoError(t, st.SaveManifest(ctx, testTenantID, testDeviceID, testRPID, map[string]any{"schema": 1}))
	bundleDir := filepath.Join(root, testTenantID, testDeviceID, testRPID)
	require.DirExists(t, bundleDir)

	// Bundle lain tidak boleh ikut terhapus.
	otherArtifact, err := st.Save(ctx, testTenantID, testDeviceID, "rp-0002",
		restorepoint.ArtifactScopedExport, nil, "scoped.rsc", bytes.NewReader([]byte("other")), 1024)
	require.NoError(t, err)

	require.NoError(t, st.DeleteBundle(ctx, testTenantID, testDeviceID, testRPID))
	require.NoDirExists(t, bundleDir)
	require.FileExists(t, filepath.Join(root, filepath.FromSlash(otherArtifact.StoragePath)))

	// Idempoten.
	require.NoError(t, st.DeleteBundle(ctx, testTenantID, testDeviceID, testRPID))

	// Identifier tidak aman ditolak.
	assert.Error(t, st.DeleteBundle(ctx, testTenantID, testDeviceID, ".."))
}

func TestRestorePointStorage_LinkArtifactDedup(t *testing.T) {
	st, _ := newRestorePointStorage(t)
	ctx := context.Background()
	src := saveTestArtifact(t, st, restorepoint.ArtifactBinaryBackup, "binary.backup", []byte("dedup-binary-payload"))

	linked, err := st.LinkArtifact(ctx, testTenantID, testDeviceID, "rp-new", src)
	require.NoError(t, err)
	assert.Equal(t, src.SHA256, linked.SHA256)
	assert.Equal(t, src.SizeBytes, linked.SizeBytes)
	assert.NotEqual(t, src.StoragePath, linked.StoragePath)

	rc, err := st.Open(ctx, linked)
	require.NoError(t, err)
	data, _ := io.ReadAll(rc)
	_ = rc.Close()
	assert.Equal(t, "dedup-binary-payload", string(data))

	// Deleting the source bundle must not break the linked artifact because
	// the file is hardlinked (or copied on cross-device fallback).
	require.NoError(t, st.DeleteBundle(ctx, testTenantID, testDeviceID, testRPID))
	rc2, err := st.Open(ctx, linked)
	require.NoError(t, err)
	data2, _ := io.ReadAll(rc2)
	_ = rc2.Close()
	assert.Equal(t, "dedup-binary-payload", string(data2))
}
