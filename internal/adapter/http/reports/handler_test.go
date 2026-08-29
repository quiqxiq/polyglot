package reports

import (
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
	if body := recorder.Body.String(); body == "" || body[0] != '{' {
		t.Errorf("daily invalid date body = %q, want JSON object", body)
	}
}
