// Package upload menyediakan endpoint publik untuk unggah bukti transfer
// pelanggan (planning v2 Fase 3). Berkas disimpan di disk lokal dengan nama
// acak; validasi ketat: ukuran maksimum, tipe konten gambar, dan ekstensi
// diturunkan dari hasil deteksi (bukan dari nama berkas klien).
package upload

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/quixiq/polyglot/pkg/logger"
	"github.com/quixiq/polyglot/pkg/response"
)

const (
	maxProofSize = 5 << 20 // 5 MB
	proofSubdir  = "payment-proofs"
)

var allowedImageTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

// Handler melayani unggah bukti pembayaran dan penyajian berkas hasil unggah.
type Handler struct {
	dir string
}

// NewHandler constructs an upload handler rooted at dir.
func NewHandler(dir string) *Handler {
	if strings.TrimSpace(dir) == "" {
		dir = "data/uploads"
	}
	return &Handler{dir: dir}
}

// RegisterPublic mounts endpoint unggah + penyajian berkas pada mux publik.
func (h *Handler) RegisterPublic(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/public/payment-proof", h.uploadPaymentProof)

	fileServer := http.FileServer(http.Dir(h.dir))
	uploads := http.StripPrefix("/uploads/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Bukti transfer memakai nama acak. Jangan pernah menampilkan isi
		// direktori: listing akan membatalkan sifat unguessable dari nama file.
		name := strings.TrimPrefix(r.URL.Path, "/")
		if name == "" || strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}
		f, err := http.Dir(h.dir).Open(name)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer func() { _ = f.Close() }()
		info, err := f.Stat()
		if err != nil || info.IsDir() {
			http.NotFound(w, r)
			return
		}
		fileServer.ServeHTTP(w, r)
	}))
	mux.Handle("GET /uploads/", uploads)
}

func (h *Handler) uploadPaymentProof(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxProofSize)
	if err := r.ParseMultipartForm(maxProofSize); err != nil {
		writeErr(w, http.StatusBadRequest, "multipart form diperlukan (field 'file')")
		return
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		writeErr(w, http.StatusBadRequest, "field file wajib")
		return
	}
	defer func() { _ = file.Close() }()

	head := make([]byte, 512)
	n, err := io.ReadFull(file, head)
	if err != nil && err != io.ErrUnexpectedEOF {
		writeErr(w, http.StatusBadRequest, "gagal membaca berkas")
		return
	}
	contentType := strings.Split(http.DetectContentType(head[:n]), ";")[0]
	ext, ok := allowedImageTypes[contentType]
	if !ok {
		writeErr(w, http.StatusUnsupportedMediaType, "hanya gambar JPG/PNG/WebP yang diterima")
		return
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		writeErr(w, http.StatusBadRequest, "gagal membaca berkas")
		return
	}

	name := randomName() + ext
	destDir := filepath.Join(h.dir, proofSubdir)
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		logger.WithComponent("UploadHTTP").WithError(err).Error("gagal membuat direktori unggah")
		writeErr(w, http.StatusInternalServerError, "gagal menyimpan berkas")
		return
	}
	dest := filepath.Join(destDir, name)
	out, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		logger.WithComponent("UploadHTTP").WithError(err).Error("gagal membuat berkas unggah")
		writeErr(w, http.StatusInternalServerError, "gagal menyimpan berkas")
		return
	}
	defer func() { _ = out.Close() }()
	if _, err := io.Copy(out, io.LimitReader(file, maxProofSize)); err != nil {
		_ = os.Remove(dest)
		writeErr(w, http.StatusInternalServerError, "gagal menyimpan berkas")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"url": "/uploads/" + proofSubdir + "/" + name,
	})
}

func randomName() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("proof-%d", time.Now().UnixNano())
	}
	return "proof-" + hex.EncodeToString(b)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	response.WriteHTTPStatusError(w, status, msg)
}
