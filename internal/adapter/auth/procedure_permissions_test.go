package auth

import "testing"

func TestPermissionFor(t *testing.T) {
	tests := []struct {
		name      string
		procedure string
		want      string
		wantOK    bool
	}{
		{"skill create", "/polyglot.v1.BotService/CreateSkill", "skill:manage", true},
		{"device stream terminal", "/polyglot.v1.DeviceService/StreamTerminal", "device:command", true},
		{"rbac manage", "/polyglot.v1.RBACService/AssignRole", "rbac:manage", true},
		{"hotspot stream", "/polyglot.v1.HotspotService/StreamTraffic", "hotspot:read", true},
		{"hotspot stream logs", "/polyglot.v1.HotspotService/StreamLogs", "log:read", true},
		{"hotspot create user", "/polyglot.v1.HotspotService/CreateUser", "hotspot:manage", true},
		{"hotspot create profile", "/polyglot.v1.HotspotService/CreateProfile", "hotspot:manage", true},
		{"hotspot list hosts", "/polyglot.v1.HotspotService/ListHosts", "hotspot:read", true},
		{"hotspot remove host", "/polyglot.v1.HotspotService/RemoveHost", "hotspot:manage", true},
		{"hotspot voucher batch", "/polyglot.v1.HotspotService/GetVoucherBatch", "hotspot:read", true},
		{"hotspot list reports", "/polyglot.v1.HotspotService/ListReports", "hotspot:read", true},
		{"hotspot delete report", "/polyglot.v1.HotspotService/DeleteReport", "hotspot:manage", true},
		{"hotspot expire status", "/polyglot.v1.HotspotService/GetExpireMonitorStatus", "hotspot:read", true},
		{"hotspot expire setup", "/polyglot.v1.HotspotService/SetupExpireMonitor", "hotspot:manage", true},
		{"hotspot expire disable", "/polyglot.v1.HotspotService/DisableExpireMonitor", "hotspot:manage", true},
		{"hotspot expire remove", "/polyglot.v1.HotspotService/RemoveExpireMonitor", "hotspot:manage", true},
		{"hotspot list templates", "/polyglot.v1.HotspotService/ListTemplates", "hotspot:read", true},
		{"hotspot get template section", "/polyglot.v1.HotspotService/GetTemplateSection", "hotspot:read", true},
		{"hotspot render vouchers", "/polyglot.v1.HotspotService/RenderVouchers", "hotspot:read", true},
		{"ppp list secrets", "/polyglot.v1.PPPService/ListSecrets", "ppp:read", true},
		{"ppp create secret", "/polyglot.v1.PPPService/CreateSecret", "ppp:manage", true},
		{"ppp kick active", "/polyglot.v1.PPPService/KickActiveSession", "ppp:manage", true},
		{"ppp list profiles", "/polyglot.v1.PPPService/ListProfiles", "ppp:read", true},
		{"probe stream", "/polyglot.v1.ProbeService/StreamTelemetry", "probe:read", true},
		{"unknown procedure", "/polyglot.v1.KnowledgeService/SomeNewRPC", "", false},
		{"empty", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := PermissionFor(tt.procedure)
			if ok != tt.wantOK {
				t.Fatalf("PermissionFor(%q) ok = %v, want %v", tt.procedure, ok, tt.wantOK)
			}
			if got != tt.want {
				t.Fatalf("PermissionFor(%q) = %q, want %q", tt.procedure, got, tt.want)
			}
		})
	}
}

func TestIspProceduresMapped(t *testing.T) {
	required := []string{
		"/polyglot.v1.BillingService/ListInvoices",
		"/polyglot.v1.BillingService/GetInvoice",
		"/polyglot.v1.BillingService/CashierResolve",
		"/polyglot.v1.BillingService/CashierPay",
		"/polyglot.v1.BillingService/ListSubscriptions",
		"/polyglot.v1.BillingService/GetSubscription",
		"/polyglot.v1.BillingService/ChangePlan",
		"/polyglot.v1.BillingService/SuspendSubscription",
		"/polyglot.v1.BillingService/ResumeSubscription",
		"/polyglot.v1.BillingService/TerminateSubscription",
		"/polyglot.v1.BillingService/ActivateSubscription",
		"/polyglot.v1.BillingService/ListPlans",
		"/polyglot.v1.BillingService/GetPlan",
		"/polyglot.v1.BillingService/CreatePlan",
		"/polyglot.v1.BillingService/UpdatePlan",
		"/polyglot.v1.BillingService/DeletePlan",
		"/polyglot.v1.BillingService/GenerateInvoices",
		"/polyglot.v1.CustomerService/FindByPhone",
		"/polyglot.v1.CustomerService/FindByCustomerCode",
		"/polyglot.v1.CustomerService/FindByPortalCode",
		"/polyglot.v1.RegistrationService/ListRegistrations",
		"/polyglot.v1.RegistrationService/GetRegistration",
		"/polyglot.v1.RegistrationService/ApproveRegistration",
		"/polyglot.v1.RegistrationService/ScheduleInstall",
		"/polyglot.v1.RegistrationService/MarkInstalled",
		"/polyglot.v1.RegistrationService/RejectRegistration",
		"/polyglot.v1.RegistrationService/CancelRegistration",
		"/polyglot.v1.RegistrationService/ConvertRegistration",
		"/polyglot.v1.CashbookService/ListAccounts",
		"/polyglot.v1.CashbookService/SaveAccount",
		"/polyglot.v1.CashbookService/ListCategories",
		"/polyglot.v1.CashbookService/SaveCategory",
		"/polyglot.v1.CashbookService/AddTransaction",
		"/polyglot.v1.CashbookService/ListTransactions",
		"/polyglot.v1.CashbookService/Balances",
		"/polyglot.v1.NotificationService/ListTemplates",
		"/polyglot.v1.NotificationService/GetTemplate",
		"/polyglot.v1.NotificationService/SaveTemplate",
		"/polyglot.v1.NotificationService/ListNotifications",
		"/polyglot.v1.NotificationService/PendingCount",
		"/polyglot.v1.NotificationService/MarkNotificationSent",
		"/polyglot.v1.NotificationService/MarkNotificationFailed",
		"/polyglot.v1.NotificationService/TestSend",
		"/polyglot.v1.ReportService/DailyReport",
		"/polyglot.v1.ReportService/MonthlyReport",
		"/polyglot.v1.ReportService/YearlyReport",
		"/polyglot.v1.ReportService/RefreshSnapshot",
		"/polyglot.v1.IspAdminService/ImportFile",
		"/polyglot.v1.IspAdminService/ImportRouter",
		"/polyglot.v1.IspAdminService/ExportCustomers",
		"/polyglot.v1.IspAdminService/Reconcile",
	}
	for _, proc := range required {
		perm, ok := PermissionFor(proc)
		if !ok || perm == "" {
			t.Errorf("procedure %q tidak terdaftar di ProcedurePermissions — akan fail-closed 403", proc)
		}
	}
}
