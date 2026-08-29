package gateway

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestChargeMalformedJSONReturnsErrorEnvelope(t *testing.T) {
	handler := NewHandler(nil)
	mux := http.NewServeMux()
	handler.RegisterProtected(mux)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/cashier/charge", nil)

	mux.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("charge malformed JSON status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	if recorder.Header().Get("Content-Type") != "application/json" {
		t.Errorf("charge malformed JSON content type = %q, want application/json", recorder.Header().Get("Content-Type"))
	}
	var payload struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode charge error envelope: %v", err)
	}
	if payload.Error.Code != "INVALID_ARGUMENT" {
		t.Errorf("charge malformed JSON code = %q, want %q", payload.Error.Code, "INVALID_ARGUMENT")
	}
	if payload.Error.Message == "" {
		t.Error("charge malformed JSON message is empty")
	}
}
