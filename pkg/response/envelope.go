package response

import (
	"encoding/json"
	"net/http"
)

// Envelope is the standard JSON response format for Polyglot APIs.
type Envelope[T any] struct {
	Success bool       `json:"success"`
	Message string     `json:"message,omitempty"`
	Data    T          `json:"data,omitempty"`
	Error   *ErrorInfo `json:"error,omitempty"`
	Meta    *Meta      `json:"meta,omitempty"`
}

// ErrorInfo contains detailed error metadata.
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Meta holds pagination or transaction metadata.
type Meta struct {
	Page       int   `json:"page,omitempty"`
	PerPage    int   `json:"per_page,omitempty"`
	TotalCount int64 `json:"total_count,omitempty"`
	TotalPages int   `json:"total_pages,omitempty"`
}

// OK writes a successful JSON response.
func OK[T any](w http.ResponseWriter, statusCode int, message string, data T) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(Envelope[T]{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// JSON writes any struct as JSON with the given status code.
func JSON[T any](w http.ResponseWriter, statusCode int, env Envelope[T]) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(env)
}

// Fail writes an error JSON response.
func Fail(w http.ResponseWriter, statusCode int, code string, message string, details string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(Envelope[any]{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}
