package app

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBuildRouter_RouteConflictAndAuthSeparation(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("buildRouter panicked: %v", r)
		}
	}()

	handler, err := buildRouter(routerDeps{})
	if err != nil {
		t.Fatalf("buildRouter() error = %v", err)
	}
	if handler == nil {
		t.Fatal("expected non-nil handler from buildRouter")
	}

	// 1. Protected route: BillingService/ListInvoices without JWT should return 401 Unauthorized
	reqProt := httptest.NewRequest("POST", "/polyglot.v1.BillingService/ListInvoices", nil)
	recProt := httptest.NewRecorder()
	handler.ServeHTTP(recProt, reqProt)
	if recProt.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for protected billing route without JWT, got %d", recProt.Code)
	}

	// 2. Protected route: RegistrationService/ListRegistrations without JWT should return 401 Unauthorized
	reqRegProt := httptest.NewRequest("POST", "/polyglot.v1.RegistrationService/ListRegistrations", nil)
	recRegProt := httptest.NewRecorder()
	handler.ServeHTTP(recRegProt, reqRegProt)
	if recRegProt.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for protected registration route without JWT, got %d", recRegProt.Code)
	}

	// 3. Public route: BillingService/SubmitManualPayment without JWT should NOT return 401 Unauthorized
	reqPub := httptest.NewRequest("POST", "/polyglot.v1.BillingService/SubmitManualPayment", nil)
	recPub := httptest.NewRecorder()
	handler.ServeHTTP(recPub, reqPub)
	if recPub.Code == http.StatusUnauthorized {
		t.Errorf("public billing route should not require JWT, got 401")
	}

	// 4. Public route: RegistrationService/SubmitRegistration without JWT should NOT return 401 Unauthorized
	reqRegPub := httptest.NewRequest("POST", "/polyglot.v1.RegistrationService/SubmitRegistration", nil)
	recRegPub := httptest.NewRecorder()
	handler.ServeHTTP(recRegPub, reqRegPub)
	if recRegPub.Code == http.StatusUnauthorized {
		t.Errorf("public registration route should not require JWT, got 401")
	}

	// 5. Protected plain-HTTP route: GET /api/devices without JWT should return 401.
	reqDevices := httptest.NewRequest(http.MethodGet, "/api/devices", nil)
	recDevices := httptest.NewRecorder()
	handler.ServeHTTP(recDevices, reqDevices)
	if recDevices.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for /api/devices without JWT, got %d", recDevices.Code)
	}
}
