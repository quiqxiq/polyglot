package auth_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	connectAuth "github.com/quixiq/polyglot/internal/adapter/connect/auth"
)

func TestAuthServiceValidationRejectsShortPassword(t *testing.T) {
	_, handler := connectAuth.NewAuthServiceHandler(nil, nil, nil, false)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(
		http.MethodPost,
		"/polyglot.v1.AuthService/ChangePassword",
		bytes.NewBufferString(`{"old_password":"oldPassword123","new_password":"short"}`),
	)
	req.Header.Set("Content-Type", "application/json")

	handler.ServeHTTP(recorder, req)

	if got, want := recorder.Code, http.StatusBadRequest; got != want {
		t.Errorf("ChangePassword short password status = %d, want %d", got, want)
	}
}

func TestAuthServiceValidationRejectsEmptyOldPassword(t *testing.T) {
	_, handler := connectAuth.NewAuthServiceHandler(nil, nil, nil, false)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(
		http.MethodPost,
		"/polyglot.v1.AuthService/ChangePassword",
		bytes.NewBufferString(`{"old_password":"","new_password":"validPassword123"}`),
	)
	req.Header.Set("Content-Type", "application/json")

	handler.ServeHTTP(recorder, req)

	if got, want := recorder.Code, http.StatusBadRequest; got != want {
		t.Errorf("ChangePassword empty old_password status = %d, want %d", got, want)
	}
}
