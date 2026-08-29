package device_test

import (
	"bytes"
	"encoding/binary"
	"net/http"
	"net/http/httptest"
	"testing"

	connectDevice "github.com/quixiq/polyglot/internal/adapter/connect/device"
)

func TestDeviceServiceValidationRejectsInvalidUnaryRequest(t *testing.T) {
	_, handler := connectDevice.NewDeviceServiceHandler(nil, nil, nil, nil)
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(
		http.MethodPost,
		"/polyglot.v1.DeviceService/GetDevice",
		bytes.NewBufferString(`{"id":""}`),
	)
	req.Header.Set("Content-Type", "application/json")

	handler.ServeHTTP(recorder, req)

	if got, want := recorder.Code, http.StatusBadRequest; got != want {
		t.Errorf("GetDevice invalid request status = %d, want %d", got, want)
	}
}

func TestDeviceServiceValidationRejectsInvalidStreamingRequest(t *testing.T) {
	_, handler := connectDevice.NewDeviceServiceHandler(nil, nil, nil, nil)
	recorder := httptest.NewRecorder()
	payload := []byte(`{"id":""}`)
	body := make([]byte, 5+len(payload))
	body[0] = 0
	binary.BigEndian.PutUint32(body[1:5], uint32(len(payload)))
	copy(body[5:], payload)
	req := httptest.NewRequest(
		http.MethodPost,
		"/polyglot.v1.DeviceService/StreamPing",
		bytes.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/connect+json")
	req.Header.Set("Connect-Protocol-Version", "1")

	handler.ServeHTTP(recorder, req)

	if got, want := recorder.Code, http.StatusOK; got != want {
		t.Errorf("StreamPing invalid request status = %d, want %d", got, want)
	}
	if !bytes.Contains(recorder.Body.Bytes(), []byte(`"code":"invalid_argument"`)) {
		t.Errorf("StreamPing invalid response = %s, want invalid_argument stream error", recorder.Body.Bytes())
	}
}
