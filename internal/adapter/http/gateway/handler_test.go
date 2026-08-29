package gateway

import (
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
	if body := recorder.Body.String(); body == "" || body[0] != '{' {
		t.Errorf("charge malformed JSON body = %q, want JSON object", body)
	}
}
