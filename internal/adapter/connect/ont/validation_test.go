package ont_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	connectOnt "github.com/quixiq/polyglot/internal/adapter/connect/ont"
)

func TestOntServiceValidationRejectsEmptyRebootOntID(t *testing.T) {
	_, handler := connectOnt.NewOntServiceHandler(nil)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(
		http.MethodPost,
		"/polyglot.v1.OntService/RebootOnt",
		bytes.NewBufferString(`{"ont_id":""}`),
	)
	req.Header.Set("Content-Type", "application/json")

	handler.ServeHTTP(recorder, req)

	if got, want := recorder.Code, http.StatusBadRequest; got != want {
		t.Errorf("RebootOnt empty ont_id status = %d, want %d", got, want)
	}
}

func TestOntServiceValidationRejectsEmptySetOntWifiSSID(t *testing.T) {
	_, handler := connectOnt.NewOntServiceHandler(nil)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(
		http.MethodPost,
		"/polyglot.v1.OntService/SetOntWifi",
		bytes.NewBufferString(`{"ont_id":"ont-123","ssid":"","password":"secretpassword"}`),
	)
	req.Header.Set("Content-Type", "application/json")

	handler.ServeHTTP(recorder, req)

	if got, want := recorder.Code, http.StatusBadRequest; got != want {
		t.Errorf("SetOntWifi empty ssid status = %d, want %d", got, want)
	}
}
