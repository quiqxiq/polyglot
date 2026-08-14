package auth

import (
	"errors"
	"strings"
	"testing"

	"connectrpc.com/connect"

	authadapter "github.com/quixiq/polyglot/internal/adapter/auth"
)

func TestValidateCreateUser(t *testing.T) {
	tests := []struct {
		name     string
		username string
		email    string
		password string
		role     string
		wantCode connect.Code
	}{
		{"valid", "teknisi1", "teknisi1@example.com", "rahasia123", "teknisi", 0},
		{"valid owner", "owner", "owner@example.com", "ownerpass1", "owner", 0},
		{"empty username", "", "a@example.com", "rahasia123", "agent", connect.CodeInvalidArgument},
		{"empty email", "user", "", "rahasia123", "agent", connect.CodeInvalidArgument},
		{"invalid email", "user", "not-an-email", "rahasia123", "agent", connect.CodeInvalidArgument},
		{"short password", "user", "a@example.com", "short", "agent", connect.CodeInvalidArgument},
		{"empty password", "user", "a@example.com", "", "agent", connect.CodeInvalidArgument},
		{"unknown role", "user", "a@example.com", "rahasia123", "superadmin", connect.CodeInvalidArgument},
		{"empty role", "user", "a@example.com", "rahasia123", "", connect.CodeInvalidArgument},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCreateUser(tt.username, tt.email, tt.password, tt.role)
			if tt.wantCode == 0 {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			var connectErr *connect.Error
			if !errors.As(err, &connectErr) {
				t.Fatalf("expected *connect.Error, got %T: %v", err, err)
			}
			if connectErr.Code() != tt.wantCode {
				t.Fatalf("expected code %v, got %v (%v)", tt.wantCode, connectErr.Code(), connectErr)
			}
		})
	}
}

func TestIsKnownRole(t *testing.T) {
	for _, role := range KnownRoles {
		t.Run("known "+role, func(t *testing.T) {
			if !isKnownRole(role) {
				t.Fatalf("expected %q to be a known role", role)
			}
		})
	}
	for _, role := range []string{"superadmin", "cashier", "manager", "", "OWNER"} {
		t.Run("unknown "+role, func(t *testing.T) {
			if isKnownRole(role) {
				t.Fatalf("expected %q to be rejected", role)
			}
		})
	}
}

// TestUserPermissionRegistry memastikan semua procedure UserService terpetakan
// ke user:read / user:manage — tanpa ini enforcement fail-closed menolak semua.
func TestUserPermissionRegistry(t *testing.T) {
	expectations := map[string]string{
		"/polyglot.v1.UserService/ListUsers":     "user:read",
		"/polyglot.v1.UserService/CreateUser":    "user:manage",
		"/polyglot.v1.UserService/UpdateUser":    "user:manage",
		"/polyglot.v1.UserService/ResetPassword": "user:manage",
		"/polyglot.v1.UserService/ToggleActive":  "user:manage",
		"/polyglot.v1.UserService/DeleteUser":    "user:manage",
	}
	for proc, want := range expectations {
		got, ok := authadapter.PermissionFor(proc)
		if !ok {
			t.Fatalf("procedure %q missing from registry — akan di-deny (fail closed)", proc)
		}
		if got != want {
			t.Fatalf("procedure %q: expected %q, got %q", proc, want, got)
		}
	}
}

// TestKnownRolesSinkronSeeder memastikan katalog role handler tidak melenceng
// dari role yang punya policy di seeder (owner/admin/agent/teknisi).
func TestKnownRolesSinkronSeeder(t *testing.T) {
	joined := strings.Join(KnownRoles, ",")
	for _, r := range []string{"owner", "admin", "agent", "teknisi"} {
		if !strings.Contains(joined, r) {
			t.Fatalf("known role %q tidak ada di katalog handler", r)
		}
	}
}
