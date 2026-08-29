package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/quixiq/polyglot/pkg/fault"
)

func TestWriteHTTPError_UsesStableEnvelopeAndStatus(t *testing.T) {
	recorder := httptest.NewRecorder()
	WriteHTTPError(recorder, fault.New(fault.KindNotFound, "customer: internal database identifier 42"))

	if got, want := recorder.Code, http.StatusNotFound; got != want {
		t.Errorf("WriteHTTPError status = %d, want %d", got, want)
	}
	if got, want := recorder.Header().Get("Content-Type"), "application/json"; got != want {
		t.Errorf("WriteHTTPError content type = %q, want %q", got, want)
	}

	var got HTTPErrorResponse
	if err := json.NewDecoder(recorder.Body).Decode(&got); err != nil {
		t.Fatalf("decode WriteHTTPError response: %v", err)
	}
	if got.Error.Code != "NOT_FOUND" {
		t.Errorf("WriteHTTPError code = %q, want %q", got.Error.Code, "NOT_FOUND")
	}
	if got.Error.Message != "resource not found" {
		t.Errorf("WriteHTTPError message = %q, want %q", got.Error.Message, "resource not found")
	}
}

func TestHTTPStatusForKind(t *testing.T) {
	tests := []struct {
		name string
		kind fault.Kind
		want int
	}{
		{name: "invalid input", kind: fault.KindInvalidInput, want: http.StatusBadRequest},
		{name: "unauthenticated", kind: fault.KindUnauthenticated, want: http.StatusUnauthorized},
		{name: "permission denied", kind: fault.KindPermissionDenied, want: http.StatusForbidden},
		{name: "conflict", kind: fault.KindConflict, want: http.StatusConflict},
		{name: "unavailable", kind: fault.KindUnavailable, want: http.StatusServiceUnavailable},
		{name: "unknown", kind: fault.KindUnknown, want: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HTTPStatusForKind(tt.kind); got != tt.want {
				t.Errorf("HTTPStatusForKind(%s) = %d, want %d", tt.kind, got, tt.want)
			}
		})
	}
}
