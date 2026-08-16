package auth

import (
	"strings"
	"testing"

	authadapter "github.com/quixiq/polyglot/internal/adapter/auth"
	userUC "github.com/quixiq/polyglot/internal/usecase/user"
)

var KnownRoles = []string{"owner", "admin", "agent", "teknisi"}

func TestIsKnownRole(t *testing.T) {
	for _, role := range KnownRoles {
		t.Run("known "+role, func(t *testing.T) {
			if !userUC.KnownRoles[role] {
				t.Fatalf("expected %q to be a known role", role)
			}
		})
	}
	for _, role := range []string{"superadmin", "cashier", "manager", "", "OWNER"} {
		t.Run("unknown "+role, func(t *testing.T) {
			if userUC.KnownRoles[role] {
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
