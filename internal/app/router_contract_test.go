package app

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	_ "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/internal/adapter/auth"
	"github.com/quixiq/polyglot/pkg/logger"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

func TestMain(m *testing.M) {
	logger.Init("error", "test", io.Discard)
	os.Exit(m.Run())
}

// publicProcedures is the explicit allowlist of ConnectRPC methods that are
// intentionally reachable without the JWT+RBAC middleware. Adding a public
// procedure requires updating this list, which makes the exposure reviewable.
var publicProcedures = map[string]bool{
	"/polyglot.v1.AuthService/Login":                      true,
	"/polyglot.v1.AuthService/GetMe":                      true,
	"/polyglot.v1.AuthService/UpdateMe":                   true,
	"/polyglot.v1.AuthService/ChangePassword":             true,
	"/polyglot.v1.AuthService/RefreshToken":               true,
	"/polyglot.v1.AuthService/Logout":                     true,
	"/polyglot.v1.RegistrationService/SubmitRegistration": true,
	"/polyglot.v1.RegistrationService/TrackRegistration":  true,
	"/polyglot.v1.RegistrationService/ListPublicPlans":    true,
	"/polyglot.v1.BillingService/ResolveManualPayment":    true,
	"/polyglot.v1.BillingService/SubmitManualPayment":     true,
	"/polyglot.v1.PortalService/RequestOTP":               true,
	"/polyglot.v1.PortalService/Login":                    true,
	"/polyglot.v1.PortalService/Overview":                 true,
	"/polyglot.v1.PortalService/MyInvoices":               true,
	"/polyglot.v1.PortalService/MyPayments":               true,
	"/polyglot.v1.PortalService/Logout":                   true,
}

func TestRouterProcedureContract(t *testing.T) {
	handler, err := buildRouter(routerDeps{})
	if err != nil {
		t.Fatalf("buildRouter() error = %v", err)
	}

	methods := protoMethods(t)
	if len(methods) == 0 {
		t.Fatal("no polyglot.v1 methods found in protoregistry")
	}

	for method := range methods {
		req := httptest.NewRequest(http.MethodPost, method, strings.NewReader("{}"))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Connect-Protocol-Version", "1")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code == http.StatusNotFound {
			t.Errorf("proto method %s is not registered on the router", method)
			continue
		}

		_, hasPermission := auth.PermissionsFor(method)
		if rec.Code == http.StatusUnauthorized {
			if !hasPermission {
				t.Errorf("protected proto method %s has no ProcedurePermissions entry", method)
			}
			continue
		}

		if !publicProcedures[method] {
			t.Errorf("proto method %s is reachable without JWT but is not in publicProcedures allowlist (status %d)", method, rec.Code)
		}
	}

	for procedure := range auth.ProcedurePermissions {
		if !strings.HasPrefix(procedure, "/polyglot.v1.") {
			continue
		}
		if _, ok := methods[procedure]; !ok {
			t.Errorf("ProcedurePermissions contains stale proto procedure %s", procedure)
		}
	}
}

func TestProtectedPlainPathsHavePermissions(t *testing.T) {
	for _, path := range protectedPlainPaths {
		if _, ok := auth.PermissionFor(path); !ok {
			t.Errorf("protected plain-HTTP path %s has no ProcedurePermissions entry", path)
		}
	}
}

func protoMethods(t *testing.T) map[string]bool {
	t.Helper()
	methods := make(map[string]bool)
	protoregistry.GlobalFiles.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		if string(fd.Package()) != "polyglot.v1" {
			return true
		}
		services := fd.Services()
		for i := 0; i < services.Len(); i++ {
			service := services.Get(i)
			serviceMethods := service.Methods()
			for j := 0; j < serviceMethods.Len(); j++ {
				methods["/"+string(service.FullName())+"/"+string(serviceMethods.Get(j).Name())] = true
			}
		}
		return true
	})
	return methods
}
