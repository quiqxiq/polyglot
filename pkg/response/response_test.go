package response_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"

	"github.com/quixiq/polyglot/pkg/response"
)

func TestEnvelopeOK(t *testing.T) {
	rec := httptest.NewRecorder()
	response.OK(rec, http.StatusOK, "success operation", map[string]string{"foo": "bar"})

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var env response.Envelope[map[string]string]
	if err := json.NewDecoder(rec.Body).Decode(&env); err != nil {
		t.Fatalf("failed to decode response envelope: %v", err)
	}

	if !env.Success || env.Message != "success operation" || env.Data["foo"] != "bar" {
		t.Errorf("unexpected envelope content: %+v", env)
	}
}

func TestToConnectError(t *testing.T) {
	testCases := []struct {
		name     string
		input    error
		wantCode connect.Code
	}{
		{"NotFound", response.ErrNotFound, connect.CodeNotFound},
		{"Unauthorized", response.ErrUnauthorized, connect.CodeUnauthenticated},
		{"Forbidden", response.ErrForbidden, connect.CodePermissionDenied},
		{"InvalidInput", response.ErrInvalidInput, connect.CodeInvalidArgument},
		{"DeviceOffline", response.ErrDeviceOffline, connect.CodeUnavailable},
		{"Generic", errors.New("unknown internal error"), connect.CodeInternal},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			connErr := response.ToConnectError(tc.input)
			if connErr == nil {
				t.Fatal("expected non-nil connect error")
			}
			if connErr.Code() != tc.wantCode {
				t.Errorf("error code mismatch: got %v, want %v", connErr.Code(), tc.wantCode)
			}
		})
	}
}
