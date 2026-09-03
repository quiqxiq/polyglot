package auth

import (
	"testing"
)

func TestPermissionFor(t *testing.T) {
	tests := []struct {
		procedure string
		want      string
		ok        bool
	}{
		{"/polyglot.v1.DeviceService/ListDevices", "device:read", true},
		{"/polyglot.v1.DeviceService/UpdateDevice", "device:manage", true},
		{"/polyglot.v1.WhatsAppService/SendTextMessage", "whatsapp:message", true},
		{"/polyglot.v1.BotService/ListConversations", "conversation:read", true},
		{"/polyglot.v1.SkillService/ListSkills", "skill:read", true},
		{"/polyglot.v1.SkillService/CreateSkill", "skill:manage", true},
		{"/polyglot.v1.SkillService/SaveResource", "skill:manage", true},
		{"/polyglot.v1.SkillService/SyncGitRepo", "skill:manage", true},
		{"/polyglot.v1.LLMConfigService/ListLLMConfigs", "llmconfig:read", true},
		{"/polyglot.v1.LLMConfigService/ActivateLLMConfig", "llmconfig:manage", true},
		{"/polyglot.v1.BotService/ListTechnicians", "technician:read", true},
		{"/polyglot.v1.HotspotService/ListProfiles", "hotspot:profile:read", true},
		{"/polyglot.v1.HotspotService/GenerateVouchers", "hotspot:voucher:generate", true},
		{"/polyglot.v1.NetworkService/ListDHCPLeases", "network:read", true},
		{"/polyglot.v1.NetworkMonitorService/StreamTraffic", "monitor:read", true},
		{"/polyglot.v1.PlanService/ListPlans", "plan:read", true},
		{"/polyglot.v1.SubscriptionService/ListSubscriptions", "subscription:read", true},
		{"/polyglot.v1.BillingService/ListInvoices", "billing:read", true},
		{"/polyglot.v1.PPPService/ListSecrets", "ppp:secret:read", true},
		{"/polyglot.v1.ProbeService/StreamTelemetry", "probe:read", true},
		{"/polyglot.v1.UserService/ListUsers", "user:read", true},
		{"/polyglot.v1.RBACService/ListPolicies", "rbac:manage", true},
		{"/polyglot.v1.UnknownService/UnknownMethod", "", false},
		{"", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.procedure, func(t *testing.T) {
			got, ok := PermissionFor(tt.procedure)
			if ok != tt.ok {
				t.Fatalf("PermissionFor(%q) ok = %v, want %v", tt.procedure, ok, tt.ok)
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
		"/polyglot.v1.BillingService/GenerateInvoices",
		"/polyglot.v1.SubscriptionService/ListSubscriptions",
		"/polyglot.v1.SubscriptionService/GetSubscription",
		"/polyglot.v1.SubscriptionService/CreateSubscription",
		"/polyglot.v1.SubscriptionService/UpdateSubscription",
		"/polyglot.v1.SubscriptionService/DeleteSubscription",
		"/polyglot.v1.SubscriptionService/ChangePlan",
		"/polyglot.v1.SubscriptionService/SuspendSubscription",
		"/polyglot.v1.SubscriptionService/ResumeSubscription",
		"/polyglot.v1.SubscriptionService/TerminateSubscription",
		"/polyglot.v1.SubscriptionService/ActivateSubscription",
		"/polyglot.v1.SubscriptionService/IsolateSubscription",
		"/polyglot.v1.SubscriptionService/RestoreSubscription",
		"/polyglot.v1.PlanService/ListPlans",
		"/polyglot.v1.PlanService/GetPlan",
		"/polyglot.v1.PlanService/CreatePlan",
		"/polyglot.v1.PlanService/UpdatePlan",
		"/polyglot.v1.PlanService/DeletePlan",
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
		"/polyglot.v1.DeviceService/GetIsolationStatus",
		"/polyglot.v1.DeviceService/CreateIsolationProfile",
		"/polyglot.v1.DeviceService/UpdateIsolationProfile",
		"/polyglot.v1.DeviceService/DeleteIsolationProfile",
		"/polyglot.v1.DeviceService/GetRouterIntegrationScript",
		"/polyglot.v1.DeviceService/ApplyRouterIntegrationScript",
	}

	for _, proc := range required {
		perm, ok := PermissionFor(proc)
		if !ok || perm == "" {
			t.Errorf("procedure %q tidak terdaftar di ProcedurePermissions — akan fail-closed 403", proc)
		}
	}
}
