package reports

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDailyInvalidDateReturnsErrorEnvelope(t *testing.T) {
	handler := NewHandler(nil)
	mux := http.NewServeMux()
	handler.Register(mux)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/reports/daily?date=invalid", nil)

	mux.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("daily invalid date status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	if recorder.Header().Get("Content-Type") != "application/json" {
		t.Errorf("daily invalid date content type = %q, want application/json", recorder.Header().Get("Content-Type"))
	}
	var payload struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode daily error envelope: %v", err)
	}
	if payload.Error.Code != "INVALID_ARGUMENT" {
		t.Errorf("daily invalid date code = %q, want %q", payload.Error.Code, "INVALID_ARGUMENT")
	}
}
