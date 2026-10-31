// Package storage menyediakan penyimpanan berkas lokal untuk skill dan bundle
// restore point terenkripsi.
package storage

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/quixiq/polyglot/internal/domain/restorepoint"
	cryptoutil "github.com/quixiq/polyglot/pkg/crypto"
)

const (
	restorePointDirMode      os.FileMode = 0o700
	restorePointFileMode     os.FileMode = 0o600
	restorePointManifest     string      = "manifest.json"
	restorePointArtifactDir  string      = "artifacts"
	restorePointEncSuffix    string      = ".enc"
	restorePointComponentMax             = 160
)

// RestorePointStorage menyimpan bundle restore point terenkripsi AES-256-GCM
// di disk lokal dengan struktur:
//
//	<root>/<tenant>/<device>/<restore_point>/manifest.json
//	<root>/<tenant>/<device>/<restore_point>/artifacts/<nama>.enc
//
// Path hanya dibentuk dari identifier yang divalidasi ketat (bukan join
// mentah input pengguna) dan semua baca/tulis/hapus wajib tetap berada di
// dalam root (PLAN-AI-ROUTER-OPS v2 M1, D9).
type RestorePointStorage struct {
	rootDir string
	key     string
}

// NewRestorePointStorage membuat penyimpanan bundle di rootDir dengan kunci
// enkripsi 32 byte. encryptionKey boleh berupa 32 byte raw atau representasi
// base64 dari 32 byte (format konfigurasi RESTORE_POINT_ENCRYPTION_KEY).
// Root dibuat dengan permission 0700 dan diperiksa dapat ditulis.
func NewRestorePointStorage(rootDir, encryptionKey string) (*RestorePointStorage, error) {
	if strings.TrimSpace(rootDir) == "" {
		return nil, fmt.Errorf("%w: root directory kosong", restorepoint.ErrStorage)
	}
	absRoot, err := filepath.Abs(rootDir)
	if err != nil {
		return nil, fmt.Errorf("%w: resolve root directory: %w", restorepoint.ErrStorage, err)
	}
	if err := os.MkdirAll(absRoot, restorePointDirMode); err != nil {
		return nil, fmt.Errorf("%w: buat root directory: %w", restorepoint.ErrStorage, err)
	}
	if err := os.Chmod(absRoot, restorePointDirMode); err != nil {
		return nil, fmt.Errorf("%w: set permission root directory: %w", restorepoint.ErrStorage, err)
	}
	probe, err := os.CreateTemp(absRoot, ".write-probe-*")
	if err != nil {
		return nil, fmt.Errorf("%w: root directory tidak writable: %w", restorepoint.ErrStorage, err)
	}
	probeName := probe.Name()
	_ = probe.Close()
	_ = os.Remove(probeName)

	key, err := decodeRestorePointKey(encryptionKey)
	if err != nil {
		return nil, err
	}
	return &RestorePointStorage{rootDir: absRoot, key: key}, nil
}

// decodeRestorePointKey menerima key raw 32 byte atau base64 dari 32 byte.
func decodeRestorePointKey(encryptionKey string) (string, error) {
	key := strings.TrimSpace(encryptionKey)
	if key == "" {
		return "", cryptoutil.ErrInvalidKeyLength
	}
	if decoded, err := base64.StdEncoding.DecodeString(key); err == nil && len(decoded) == 32 {
		return string(decoded), nil
	}
	if len(key) != 32 {
		return "", cryptoutil.ErrInvalidKeyLength
	}
	return key, nil
}

// Save membaca data (maksimal maxBytes bila > 0), mengenkripsi, lalu menulis
// artifact secara atomik. SHA256 dan SizeBytes dihitung dari plaintext agar
// Verify dapat memeriksa integritas setelah dekripsi.
func (s *RestorePointStorage) Save(
	ctx context.Context,
	tenantID, deviceID, restorePointID string,
	kind restorepoint.ArtifactKind,
	scope []string,
	name string,
	data io.Reader,
	maxBytes int64,
) (restorepoint.Artifact, error) {
	if err := ctx.Err(); err != nil {
		return restorepoint.Artifact{}, err
	}
	if err := validateArtifactKind(kind); err != nil {
		return restorepoint.Artifact{}, err
	}
	if data == nil {
		return restorepoint.Artifact{}, fmt.Errorf("%w: data artifact nil", restorepoint.ErrStorage)
	}
	if name == "" {
		name = defaultArtifactName(kind)
	}
	if err := validatePathComponent("artifact name", name); err != nil {
		return restorepoint.Artifact{}, err
	}
	artifactsDir, err := s.artifactDir(tenantID, deviceID, restorePointID)
	if err != nil {
		return restorepoint.Artifact{}, err
	}
	if err := os.MkdirAll(artifactsDir, restorePointDirMode); err != nil {
		return restorepoint.Artifact{}, fmt.Errorf("%w: buat direktori artifact: %w", restorepoint.ErrStorage, err)
	}

	var src io.Reader = data
	if maxBytes > 0 {
		src = io.LimitReader(data, maxBytes+1)
	}
	plaintext, err := io.ReadAll(src)
	if err != nil {
		return restorepoint.Artifact{}, fmt.Errorf("%w: baca data artifact: %w", restorepoint.ErrStorage, err)
	}
	if maxBytes > 0 && int64(len(plaintext)) > maxBytes {
		return restorepoint.Artifact{}, fmt.Errorf("%w: artifact %s melebihi batas %d byte", restorepoint.ErrSizeLimit, name, maxBytes)
	}
	ciphertext, err := cryptoutil.EncryptBytes(plaintext, s.key)
	if err != nil {
		return restorepoint.Artifact{}, fmt.Errorf("%w: enkripsi artifact: %w", restorepoint.ErrStorage, err)
	}

	relPath := filepath.ToSlash(filepath.Join(tenantID, deviceID, restorePointID, restorePointArtifactDir, name+restorePointEncSuffix))
	fullPath, err := s.resolveArtifactPath(relPath)
	if err != nil {
		return restorepoint.Artifact{}, err
	}
	if err := writeFileAtomic(fullPath, ciphertext, restorePointFileMode); err != nil {
		return restorepoint.Artifact{}, fmt.Errorf("%w: tulis artifact: %w", restorepoint.ErrStorage, err)
	}

	sum := sha256.Sum256(plaintext)
	return restorepoint.Artifact{
		Kind:        kind,
		Scope:       append([]string(nil), scope...),
		StoragePath: relPath,
		SHA256:      hex.EncodeToString(sum[:]),
		SizeBytes:   int64(len(plaintext)),
	}, nil
}

// SaveManifest menulis manifest.json bundle (metadata + hash artifact, tanpa
// secret) secara atomik. Pemanggil bertanggung jawab tidak memasukkan
// credential/secret ke dalam manifest.
func (s *RestorePointStorage) SaveManifest(ctx context.Context, tenantID, deviceID, restorePointID string, manifest any) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if manifest == nil {
		return fmt.Errorf("%w: manifest nil", restorepoint.ErrStorage)
	}
	bundleDir, err := s.bundleDir(tenantID, deviceID, restorePointID)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(bundleDir, restorePointDirMode); err != nil {
		return fmt.Errorf("%w: buat direktori bundle: %w", restorepoint.ErrStorage, err)
	}
	payload, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("%w: marshal manifest: %w", restorepoint.ErrStorage, err)
	}
	payload = append(payload, '\n')
	if err := writeFileAtomic(filepath.Join(bundleDir, restorePointManifest), payload, restorePointFileMode); err != nil {
		return fmt.Errorf("%w: tulis manifest: %w", restorepoint.ErrStorage, err)
	}
	return nil
}

// Open mendekripsi artifact dan mengembalikan reader atas plaintext.
func (s *RestorePointStorage) Open(ctx context.Context, artifact restorepoint.Artifact) (io.ReadCloser, error) {
	plaintext, err := s.decryptArtifact(ctx, artifact)
	if err != nil {
		return nil, err
	}
	return io.NopCloser(bytes.NewReader(plaintext)), nil
}

// Verify mendekripsi artifact lalu membandingkan SHA256 (dan ukuran bila
// diketahui) dengan metadata Artifact. Mismatch menandakan file korup atau
// termodifikasi di luar sistem.
func (s *RestorePointStorage) Verify(ctx context.Context, artifact restorepoint.Artifact) error {
	plaintext, err := s.decryptArtifact(ctx, artifact)
	if err != nil {
		return err
	}
	sum := sha256.Sum256(plaintext)
	got := hex.EncodeToString(sum[:])
	want := strings.ToLower(strings.TrimSpace(artifact.SHA256))
	if want == "" || got != want {
		return fmt.Errorf("%w: sha256 artifact %s tidak cocok", restorepoint.ErrStorage, artifact.StoragePath)
	}
	if artifact.SizeBytes > 0 && int64(len(plaintext)) != artifact.SizeBytes {
		return fmt.Errorf("%w: ukuran artifact %s tidak cocok", restorepoint.ErrStorage, artifact.StoragePath)
	}
	return nil
}

// DeleteBundle menghapus direktori bundle beserta seluruh artifact-nya.
// Idempoten: bundle yang tidak ada bukan error.
func (s *RestorePointStorage) DeleteBundle(ctx context.Context, tenantID, deviceID, restorePointID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	bundleDir, err := s.bundleDir(tenantID, deviceID, restorePointID)
	if err != nil {
		return err
	}
	if !isStrictlyWithin(s.rootDir, bundleDir) {
		return fmt.Errorf("%w: path bundle %q keluar dari root", restorepoint.ErrStorage, bundleDir)
	}
	if err := os.RemoveAll(bundleDir); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("%w: hapus bundle: %w", restorepoint.ErrStorage, err)
	}
	return nil
}

// Stat mengembalikan metadata filesystem artifact terenkripsi (ukuran file
// tersimpan, bukan ukuran plaintext).
func (s *RestorePointStorage) Stat(ctx context.Context, artifact restorepoint.Artifact) (os.FileInfo, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	fullPath, err := s.resolveArtifactPath(artifact.StoragePath)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(fullPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("%w: artifact %s", restorepoint.ErrNotFound, artifact.StoragePath)
		}
		return nil, fmt.Errorf("%w: stat artifact: %w", restorepoint.ErrStorage, err)
	}
	return info, nil
}

// TotalSize menghitung total byte seluruh file di dalam satu bundle
// (artifact terenkripsi + manifest). Bundle yang belum ada mengembalikan 0.
func (s *RestorePointStorage) TotalSize(ctx context.Context, tenantID, deviceID, restorePointID string) (int64, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	bundleDir, err := s.bundleDir(tenantID, deviceID, restorePointID)
	if err != nil {
		return 0, err
	}
	var total int64
	walkErr := filepath.WalkDir(bundleDir, func(_ string, entry fs.DirEntry, err error) error {
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return nil
			}
			return err
		}
		if !entry.Type().IsRegular() {
			return nil
		}
		info, infoErr := entry.Info()
		if infoErr != nil {
			if errors.Is(infoErr, os.ErrNotExist) {
				return nil
			}
			return infoErr
		}
		total += info.Size()
		return nil
	})
	if walkErr != nil {
		if errors.Is(walkErr, os.ErrNotExist) {
			return 0, nil
		}
		return 0, fmt.Errorf("%w: hitung total size bundle: %w", restorepoint.ErrStorage, walkErr)
	}
	return total, nil
}

// decryptArtifact membaca file terenkripsi berdasarkan StoragePath relatif
// artifact dan mengembalikan plaintext-nya.
func (s *RestorePointStorage) decryptArtifact(ctx context.Context, artifact restorepoint.Artifact) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	fullPath, err := s.resolveArtifactPath(artifact.StoragePath)
	if err != nil {
		return nil, err
	}
	ciphertext, err := os.ReadFile(fullPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("%w: artifact %s", restorepoint.ErrNotFound, artifact.StoragePath)
		}
		return nil, fmt.Errorf("%w: baca artifact: %w", restorepoint.ErrStorage, err)
	}
	plaintext, err := cryptoutil.DecryptBytes(ciphertext, s.key)
	if err != nil {
		return nil, fmt.Errorf("%w: dekripsi artifact: %w", restorepoint.ErrStorage, err)
	}
	return plaintext, nil
}

// bundleDir membentuk path bundle dari identifier tervalidasi.
func (s *RestorePointStorage) bundleDir(tenantID, deviceID, restorePointID string) (string, error) {
	components := []struct {
		role  string
		value string
	}{
		{"tenant id", tenantID},
		{"device id", deviceID},
		{"restore point id", restorePointID},
	}
	for _, c := range components {
		if err := validatePathComponent(c.role, c.value); err != nil {
			return "", err
		}
	}
	full := filepath.Join(s.rootDir, tenantID, deviceID, restorePointID)
	if !isStrictlyWithin(s.rootDir, full) {
		return "", fmt.Errorf("%w: path bundle tidak valid", restorepoint.ErrStorage)
	}
	return full, nil
}

// artifactDir membentuk path direktori artifacts sebuah bundle.
func (s *RestorePointStorage) artifactDir(tenantID, deviceID, restorePointID string) (string, error) {
	bundleDir, err := s.bundleDir(tenantID, deviceID, restorePointID)
	if err != nil {
		return "", err
	}
	return filepath.Join(bundleDir, restorePointArtifactDir), nil
}

// resolveArtifactPath mengubah StoragePath relatif menjadi path absolut yang
// dijamin berada di dalam root.
func (s *RestorePointStorage) resolveArtifactPath(relPath string) (string, error) {
	if strings.TrimSpace(relPath) == "" {
		return "", fmt.Errorf("%w: storage path artifact kosong", restorepoint.ErrStorage)
	}
	clean := filepath.Clean(filepath.FromSlash(relPath))
	if filepath.IsAbs(clean) || clean == "." || clean == ".." ||
		strings.HasPrefix(clean, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("%w: storage path %q keluar dari root", restorepoint.ErrStorage, relPath)
	}
	full := filepath.Join(s.rootDir, clean)
	if !isStrictlyWithin(s.rootDir, full) {
		return "", fmt.Errorf("%w: storage path %q keluar dari root", restorepoint.ErrStorage, relPath)
	}
	return full, nil
}

// validatePathComponent memastikan satu komponen path hanya berisi karakter
// aman dan bukan "." / "..". Karakter "/" dan "\" otomatis ditolak karena
// tidak termasuk allowlist.
func validatePathComponent(role, value string) error {
	if value == "" || len(value) > restorePointComponentMax || value == "." || value == ".." {
		return fmt.Errorf("%w: %s %q tidak valid", restorepoint.ErrStorage, role, value)
	}
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z',
			r >= 'A' && r <= 'Z',
			r >= '0' && r <= '9',
			r == '.', r == '_', r == '-':
			continue
		default:
			return fmt.Errorf("%w: %s %q mengandung karakter terlarang", restorepoint.ErrStorage, role, value)
		}
	}
	return nil
}

// validateArtifactKind membatasi kind ke nilai domain yang dikenal (selaras
// dengan CHECK constraint restore_point_artifacts).
func validateArtifactKind(kind restorepoint.ArtifactKind) error {
	switch kind {
	case restorepoint.ArtifactScopedExport, restorepoint.ArtifactBinaryBackup, restorepoint.ArtifactInverseScript:
		return nil
	default:
		return fmt.Errorf("%w: kind artifact %q tidak dikenal", restorepoint.ErrStorage, string(kind))
	}
}

// defaultArtifactName memberi nama file default bila pemanggil tidak
// menyuplai nama eksplisit.
func defaultArtifactName(kind restorepoint.ArtifactKind) string {
	switch kind {
	case restorepoint.ArtifactBinaryBackup:
		return "binary.backup"
	case restorepoint.ArtifactInverseScript:
		return "inverse.rsc"
	default:
		return "scoped.rsc"
	}
}

// isStrictlyWithin melaporkan apakah path berada di bawah root (bukan root
// itu sendiri). Dipakai sebagai pertahanan terakhir sebelum I/O file.
func isStrictlyWithin(root, path string) bool {
	if path == root {
		return false
	}
	return strings.HasPrefix(path, root+string(os.PathSeparator))
}

// writeFileAtomic menulis data ke path lewat file temporer di direktori yang
// sama lalu rename, sehingga pembaca tidak pernah melihat file setengah jadi.
func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }()

	if err := tmp.Chmod(perm); err != nil {
		_ = tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}

// LinkArtifact membuat artifact baru di bundle tujuan dengan mereferensikan
// file terenkripsi yang sudah ada (hardlink bila memungkinkan, fallback copy).
// Dipakai dedupe binary backup: dua bundle berbagi blok disk yang sama dan
// menghapus salah satu bundle tidak menghapus file bundle lain.
func (s *RestorePointStorage) LinkArtifact(
	ctx context.Context,
	tenantID, deviceID, newRestorePointID string,
	src restorepoint.Artifact,
) (restorepoint.Artifact, error) {
	if err := ctx.Err(); err != nil {
		return restorepoint.Artifact{}, err
	}
	if err := validateArtifactKind(src.Kind); err != nil {
		return restorepoint.Artifact{}, err
	}
	srcPath, err := s.resolveArtifactPath(src.StoragePath)
	if err != nil {
		return restorepoint.Artifact{}, err
	}
	dstDir, err := s.artifactDir(tenantID, deviceID, newRestorePointID)
	if err != nil {
		return restorepoint.Artifact{}, err
	}
	if err := os.MkdirAll(dstDir, restorePointDirMode); err != nil {
		return restorepoint.Artifact{}, fmt.Errorf("%w: buat direktori artifact: %w", restorepoint.ErrStorage, err)
	}
	base := filepath.Base(srcPath)
	dstPath := filepath.Join(dstDir, base)
	if !isStrictlyWithin(s.rootDir, dstPath) {
		return restorepoint.Artifact{}, fmt.Errorf("%w: path artifact tujuan tidak valid", restorepoint.ErrStorage)
	}
	if err := os.Link(srcPath, dstPath); err != nil {
		// Fallback lintas filesystem: salin byte terenkripsi apa adanya.
		in, oerr := os.Open(srcPath)
		if oerr != nil {
			return restorepoint.Artifact{}, fmt.Errorf("%w: buka artifact sumber: %w", restorepoint.ErrStorage, oerr)
		}
		defer func() { _ = in.Close() }()
		out, cerr := os.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, restorePointFileMode)
		if cerr != nil {
			return restorepoint.Artifact{}, fmt.Errorf("%w: buat artifact tujuan: %w", restorepoint.ErrStorage, cerr)
		}
		if _, errCopy := io.Copy(out, in); errCopy != nil {
			_ = out.Close()
			return restorepoint.Artifact{}, fmt.Errorf("%w: salin artifact: %w", restorepoint.ErrStorage, errCopy)
		}
		if errClose := out.Close(); errClose != nil {
			return restorepoint.Artifact{}, fmt.Errorf("%w: tutup artifact: %w", restorepoint.ErrStorage, errClose)
		}
	}
	rel, err := filepath.Rel(s.rootDir, dstPath)
	if err != nil {
		return restorepoint.Artifact{}, fmt.Errorf("%w: hitung path relatif: %w", restorepoint.ErrStorage, err)
	}
	out := src
	out.StoragePath = filepath.ToSlash(rel)
	return out, nil
}
