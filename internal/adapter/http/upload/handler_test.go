package upload

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestUploadedFileIsServed(t *testing.T) {
	dir := t.TempDir()
	proofDir := filepath.Join(dir, proofSubdir)
	if err := os.MkdirAll(proofDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	name := "proof-test.jpg"
	if err := os.WriteFile(filepath.Join(proofDir, name), []byte("image"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	mux := http.NewServeMux()
	NewHandler(dir).RegisterPublic(mux)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/uploads/"+proofSubdir+"/"+name, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET uploaded file status = %d, want 200", rec.Code)
	}
	if rec.Body.String() != "image" {
		t.Fatalf("GET uploaded file body = %q, want %q", rec.Body.String(), "image")
	}
}

func TestUploadsDirectoryListingIsDenied(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, proofSubdir), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	mux := http.NewServeMux()
	NewHandler(dir).RegisterPublic(mux)

	for _, target := range []string{
		"/uploads/",
		"/uploads/" + proofSubdir,
		"/uploads/" + proofSubdir + "/",
	} {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, target, nil))
		if rec.Code != http.StatusNotFound {
			t.Errorf("GET %s status = %d, want 404", target, rec.Code)
		}
	}
}

func TestUploadsMissingFileIsNotFound(t *testing.T) {
	mux := http.NewServeMux()
	NewHandler(t.TempDir()).RegisterPublic(mux)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/uploads/payment-proofs/missing.jpg", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("GET missing upload status = %d, want 404", rec.Code)
	}
}
