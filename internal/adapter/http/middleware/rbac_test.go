package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// fakeEnforcer implements PolicyEnforcer with configurable behavior.
type fakeEnforcer struct {
	roles   map[string][]string // user -> roles
	allow   map[string]bool     // "obj|role" -> allowed
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
	gin.SetMode(gin.TestMode)

	// admin user 1 boleh knowledge:write; user 2 tidak punya role.
	ce := &fakeEnforcer{
		roles: map[string][]string{"1": {"admin"}},
		allow: map[string]bool{"knowledge:write|admin": true},
	}

	newRouter := func(userID uint, fallbackRole string) *gin.Engine {
		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set("user_id", userID)
			c.Set("user_role", fallbackRole)
			c.Next()
		})
		r.Use(AuthorizeProcedure(ce))
		r.POST("/polyglot.v1.KnowledgeService/CreateKnowledge", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"ok": true})
		})
		return r
	}

	t.Run("allowed admin", func(t *testing.T) {
		r := newRouter(1, "admin")
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/polyglot.v1.KnowledgeService/CreateKnowledge", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200 (body: %s)", w.Code, w.Body.String())
		}
	})

	t.Run("fallback to JWT role when casbin has none", func(t *testing.T) {
		// user 2 tidak punya assignment Casbin; fallback ke klaim JWT "admin".
		r := newRouter(2, "admin")
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/polyglot.v1.KnowledgeService/CreateKnowledge", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200 via JWT fallback (body: %s)", w.Code, w.Body.String())
		}
	})

	t.Run("no roles at all -> 401", func(t *testing.T) {
		r := newRouter(3, "")
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/polyglot.v1.KnowledgeService/CreateKnowledge", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401 (body: %s)", w.Code, w.Body.String())
		}
	})

	t.Run("role present but policy denies", func(t *testing.T) {
		// user 4 punya role "agent" (via Casbin) tapi tidak boleh knowledge:write.
		ce.roles["4"] = []string{"agent"}
		r := newRouter(4, "admin")
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/polyglot.v1.KnowledgeService/CreateKnowledge", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want 403 (body: %s)", w.Code, w.Body.String())
		}
	})

	t.Run("unknown procedure fail closed", func(t *testing.T) {
		r := newRouter(1, "admin")
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/polyglot.v1.KnowledgeService/NotRegistered", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want 403 (body: %s)", w.Code, w.Body.String())
		}
	})
}

func TestAuthorizeProcedureNilEnforcer(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", uint(1))
		c.Set("user_role", "admin")
		c.Next()
	})
	r.Use(AuthorizeProcedure(nil))
	r.POST("/polyglot.v1.KnowledgeService/CreateKnowledge", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/polyglot.v1.KnowledgeService/CreateKnowledge", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("nil enforcer status = %d, want 500 (body: %s)", w.Code, w.Body.String())
	}
}
