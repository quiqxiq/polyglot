package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/quixiq/polyglot/internal/adapter/auth"
	"github.com/quixiq/polyglot/internal/adapter/http/middleware"
)

type fakeEnforcer struct {
	roles   map[string][]string
	allow   map[string]bool
	denyAll bool
}

func (f *fakeEnforcer) Enforce(sub, obj, act string) (bool, error) {
	if f.denyAll {
		return false, nil
	}
	return f.allow[obj+"|"+sub], nil
}

func (f *fakeEnforcer) GetRolesForUser(user string) ([]string, error) {
	return f.roles[user], nil
}

func TestAuthorizeProcedureMiddleware(t *testing.T) {
	ce := &fakeEnforcer{
		roles: map[string][]string{"1": {"admin"}},
		allow: map[string]bool{"knowledge:write|admin": true},
	}

	newMux := func(userID uint, fallbackRoles []string) http.Handler {
		finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"ok":true}`))
		})

		rbacMW := middleware.AuthorizeProcedure(ce)
		identityInjector := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if userID != 0 {
					ctx := auth.WithIdentity(r.Context(), userID, fallbackRoles)
					r = r.WithContext(ctx)
				}
				next.ServeHTTP(w, r)
			})
		}

		return identityInjector(rbacMW(finalHandler))
	}

	t.Run("allowed admin", func(t *testing.T) {
		h := newMux(1, []string{"admin"})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/polyglot.v1.KnowledgeService/CreateKnowledge", nil)
		h.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200 (body: %s)", w.Code, w.Body.String())
		}
	})

	t.Run("fallback to JWT role when casbin has none", func(t *testing.T) {
		h := newMux(2, []string{"admin"})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/polyglot.v1.KnowledgeService/CreateKnowledge", nil)
		h.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200 via fallback (body: %s)", w.Code, w.Body.String())
		}
	})

	t.Run("no identity at all -> 401", func(t *testing.T) {
		h := newMux(0, nil)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/polyglot.v1.KnowledgeService/CreateKnowledge", nil)
		h.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401 (body: %s)", w.Code, w.Body.String())
		}
	})

	t.Run("role present but policy denies", func(t *testing.T) {
		ce.roles["4"] = []string{"agent"}
		h := newMux(4, []string{"agent"})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/polyglot.v1.KnowledgeService/CreateKnowledge", nil)
		h.ServeHTTP(w, req)
		if w.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want 403 (body: %s)", w.Code, w.Body.String())
		}
	})

	t.Run("unknown procedure fail closed", func(t *testing.T) {
		h := newMux(1, []string{"admin"})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/polyglot.v1.KnowledgeService/NotRegistered", nil)
		h.ServeHTTP(w, req)
		if w.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want 403 (body: %s)", w.Code, w.Body.String())
		}
	})
}
