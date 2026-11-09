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
		{name: "not found", kind: fault.KindNotFound, want: http.StatusNotFound},
		{name: "invalid input", kind: fault.KindInvalidInput, want: http.StatusBadRequest},
		{name: "already exists", kind: fault.KindAlreadyExists, want: http.StatusConflict},
		{name: "unauthenticated", kind: fault.KindUnauthenticated, want: http.StatusUnauthorized},
		{name: "permission denied", kind: fault.KindPermissionDenied, want: http.StatusForbidden},
		{name: "failed precondition", kind: fault.KindFailedPrecondition, want: http.StatusPreconditionFailed},
		{name: "conflict", kind: fault.KindConflict, want: http.StatusConflict},
		{name: "unavailable", kind: fault.KindUnavailable, want: http.StatusServiceUnavailable},
		{name: "resource exhausted", kind: fault.KindResourceExhausted, want: http.StatusTooManyRequests},
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

func TestHTTPCodeAndMessageForKind(t *testing.T) {
	tests := []struct {
		kind        fault.Kind
		wantCode    string
		wantMessage string
	}{
		{kind: fault.KindNotFound, wantCode: "NOT_FOUND", wantMessage: "resource not found"},
		{kind: fault.KindInvalidInput, wantCode: "INVALID_ARGUMENT", wantMessage: "request validation failed"},
		{kind: fault.KindAlreadyExists, wantCode: "ALREADY_EXISTS", wantMessage: "resource already exists"},
		{kind: fault.KindPermissionDenied, wantCode: "PERMISSION_DENIED", wantMessage: "permission denied"},
		{kind: fault.KindUnauthenticated, wantCode: "UNAUTHENTICATED", wantMessage: "authentication required"},
		{kind: fault.KindFailedPrecondition, wantCode: "FAILED_PRECONDITION", wantMessage: "request cannot be completed in the current state"},
		{kind: fault.KindConflict, wantCode: "CONFLICT", wantMessage: "request conflicts with the current state"},
		{kind: fault.KindUnavailable, wantCode: "UNAVAILABLE", wantMessage: "service temporarily unavailable"},
		{kind: fault.KindResourceExhausted, wantCode: "RESOURCE_EXHAUSTED", wantMessage: "request limit exceeded"},
		{kind: fault.KindUnknown, wantCode: "INTERNAL", wantMessage: "internal server error"},
	}

	for _, tt := range tests {
		t.Run(tt.kind.String(), func(t *testing.T) {
			if got := HTTPCodeForKind(tt.kind); got != tt.wantCode {
				t.Errorf("HTTPCodeForKind(%s) = %q, want %q", tt.kind, got, tt.wantCode)
			}
			if got := HTTPMessageForKind(tt.kind); got != tt.wantMessage {
				t.Errorf("HTTPMessageForKind(%s) = %q, want %q", tt.kind, got, tt.wantMessage)
			}
		})
	}
}

func TestWriteHTTPStatusError(t *testing.T) {
	statuses := []struct {
		status   int
		wantCode string
		message  string
	}{
		{status: http.StatusBadRequest, wantCode: "INVALID_ARGUMENT", message: "invalid body"},
		{status: http.StatusUnauthorized, wantCode: "UNAUTHENTICATED", message: "token expired"},
		{status: http.StatusForbidden, wantCode: "PERMISSION_DENIED", message: "forbidden"},
		{status: http.StatusNotFound, wantCode: "NOT_FOUND", message: "item not found"},
		{status: http.StatusConflict, wantCode: "CONFLICT", message: "already booked"},
		{status: http.StatusPreconditionFailed, wantCode: "FAILED_PRECONDITION", message: "precondition not met"},
		{status: http.StatusTooManyRequests, wantCode: "RESOURCE_EXHAUSTED", message: "slow down"},
		{status: http.StatusServiceUnavailable, wantCode: "UNAVAILABLE", message: "server down"},
		{status: http.StatusInternalServerError, wantCode: "INTERNAL", message: "internal error"},
	}

	for _, tt := range statuses {
		t.Run(tt.wantCode, func(t *testing.T) {
			rec := httptest.NewRecorder()
			WriteHTTPStatusError(rec, tt.status, tt.message)

			if rec.Code != tt.status {
				t.Errorf("status = %d, want %d", rec.Code, tt.status)
			}
			var resp HTTPErrorResponse
			if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}
			if resp.Error.Code != tt.wantCode {
				t.Errorf("code = %q, want %q", resp.Error.Code, tt.wantCode)
			}
			if resp.Error.Message != tt.message {
				t.Errorf("message = %q, want %q", resp.Error.Message, tt.message)
			}
		})
	}
}
