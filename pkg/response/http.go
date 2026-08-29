package response

import (
	"encoding/json"
	"net/http"

	"github.com/quixiq/polyglot/pkg/fault"
)

// HTTPErrorResponse is the stable JSON envelope returned by plain HTTP APIs.
type HTTPErrorResponse struct {
	Error HTTPError `json:"error"`
}

// HTTPError contains a machine-readable code and safe public message.
type HTTPError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

// HTTPStatusForKind maps a fault classification to an HTTP status code.
func HTTPStatusForKind(kind fault.Kind) int {
	switch kind {
	case fault.KindNotFound:
		return http.StatusNotFound
	case fault.KindInvalidInput:
		return http.StatusBadRequest
	case fault.KindAlreadyExists, fault.KindConflict:
		return http.StatusConflict
	case fault.KindPermissionDenied:
		return http.StatusForbidden
	case fault.KindUnauthenticated:
		return http.StatusUnauthorized
	case fault.KindFailedPrecondition:
		return http.StatusPreconditionFailed
	case fault.KindUnavailable:
		return http.StatusServiceUnavailable
	case fault.KindResourceExhausted:
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}

// HTTPCodeForKind returns the stable public code for a fault classification.
func HTTPCodeForKind(kind fault.Kind) string {
	switch kind {
	case fault.KindNotFound:
		return "NOT_FOUND"
	case fault.KindInvalidInput:
		return "INVALID_ARGUMENT"
	case fault.KindAlreadyExists:
		return "ALREADY_EXISTS"
	case fault.KindPermissionDenied:
		return "PERMISSION_DENIED"
	case fault.KindUnauthenticated:
		return "UNAUTHENTICATED"
	case fault.KindFailedPrecondition:
		return "FAILED_PRECONDITION"
	case fault.KindConflict:
		return "CONFLICT"
	case fault.KindUnavailable:
		return "UNAVAILABLE"
	case fault.KindResourceExhausted:
		return "RESOURCE_EXHAUSTED"
	default:
		return "INTERNAL"
	}
}

// HTTPMessageForKind returns a safe message that does not expose internal error details.
func HTTPMessageForKind(kind fault.Kind) string {
	switch kind {
	case fault.KindNotFound:
		return "resource not found"
	case fault.KindInvalidInput:
		return "request validation failed"
	case fault.KindAlreadyExists:
		return "resource already exists"
	case fault.KindPermissionDenied:
		return "permission denied"
	case fault.KindUnauthenticated:
		return "authentication required"
	case fault.KindFailedPrecondition:
		return "request cannot be completed in the current state"
	case fault.KindConflict:
		return "request conflicts with the current state"
	case fault.KindUnavailable:
		return "service temporarily unavailable"
	case fault.KindResourceExhausted:
		return "request limit exceeded"
	default:
		return "internal server error"
	}
}

// WriteHTTPError writes a safe, stable error envelope for a plain HTTP endpoint.
func WriteHTTPError(w http.ResponseWriter, err error) {
	kind := fault.KindOf(err)
	status := HTTPStatusForKind(kind)
	payload := HTTPErrorResponse{Error: HTTPError{
		Code:    HTTPCodeForKind(kind),
		Message: HTTPMessageForKind(kind),
	}}
	writeHTTPJSON(w, status, payload)
}

func writeHTTPJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
