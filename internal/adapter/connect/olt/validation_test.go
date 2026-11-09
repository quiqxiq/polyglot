package olt_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	connectOlt "github.com/quixiq/polyglot/internal/adapter/connect/olt"
)

func TestOltServiceValidationRejectsEmptyGetOltID(t *testing.T) {
	_, handler := connectOlt.NewOltServiceHandler(nil, nil)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(
		http.MethodPost,
		"/polyglot.v1.OltService/GetOlt",
		bytes.NewBufferString(`{"id":""}`),
	)
	req.Header.Set("Content-Type", "application/json")

	handler.ServeHTTP(recorder, req)

	if got, want := recorder.Code, http.StatusBadRequest; got != want {
		t.Errorf("GetOlt empty id status = %d, want %d", got, want)
	}
}

func TestOltServiceValidationRejectsCreateOltWithoutHost(t *testing.T) {
	_, handler := connectOlt.NewOltServiceHandler(nil, nil)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(
		http.MethodPost,
		"/polyglot.v1.OltService/CreateOlt",
		bytes.NewBufferString(`{"name":"test-olt","host":"","community_ro":"public"}`),
	)
	req.Header.Set("Content-Type", "application/json")

	handler.ServeHTTP(recorder, req)

	if got, want := recorder.Code, http.StatusBadRequest; got != want {
		t.Errorf("CreateOlt empty host status = %d, want %d", got, want)
	}
}
