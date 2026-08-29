package adminapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestImportRouterMissingDeviceReturnsErrorEnvelope(t *testing.T) {
	handler := NewHandler(nil, nil, nil, nil, nil, nil)
	mux := http.NewServeMux()
	handler.RegisterProtected(mux)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/admin/import-router", nil)

	mux.ServeHTTP(recorder, request)

	assertHTTPErrorEnvelope(t, recorder, http.StatusBadRequest, "INVALID_ARGUMENT")
}

func assertHTTPErrorEnvelope(t *testing.T, recorder *httptest.ResponseRecorder, status int, code string) {
	t.Helper()
	if recorder.Code != status {
		t.Errorf("HTTP error status = %d, want %d", recorder.Code, status)
	}
	if recorder.Header().Get("Content-Type") != "application/json" {
		t.Errorf("HTTP error content type = %q, want application/json", recorder.Header().Get("Content-Type"))
	}
	var payload struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode HTTP error envelope: %v", err)
	}
	if payload.Error.Code != code {
		t.Errorf("HTTP error body = %s, want code %q", recorder.Body.Bytes(), code)
	}
}
