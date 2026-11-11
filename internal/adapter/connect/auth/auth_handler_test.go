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

func TestAuthServiceLogoutClearsNetOpsCookie(t *testing.T) {
	_, handler := connectAuth.NewAuthServiceHandler(nil, nil, nil, false)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(
		http.MethodPost,
		"/polyglot.v1.AuthService/Logout",
		bytes.NewBufferString(`{}`),
	)
	req.Header.Set("Content-Type", "application/json")

	handler.ServeHTTP(recorder, req)

	if got, want := recorder.Code, http.StatusOK; got != want {
		t.Fatalf("Logout status = %d, want %d", got, want)
	}

	cookies := recorder.Result().Cookies()
	var foundRefreshToken, foundNetOps bool
	for _, c := range cookies {
		if c.Name == "polyglot_refresh" && (c.MaxAge <= 0) {
			foundRefreshToken = true
		}
		if c.Name == "polyglot_netops" && c.Path == "/netops" && (c.MaxAge <= 0) {
			foundNetOps = true
		}
	}

	if !foundRefreshToken {
		t.Errorf("expected expired refresh_token cookie")
	}
	if !foundNetOps {
		t.Errorf("expected expired polyglot_netops cookie with Path=/netops")
	}
}
