package auth

// ProcedurePermissions maps a ConnectRPC procedure (URL path) to a Casbin
// object in the form "resource:action" (e.g. "knowledge:write"). Enforcement
// uses this object against the RBAC policies; unknown procedures are denied
// (fail closed). Keep this registry in sync with the service handlers.
var ProcedurePermissions = map[string]string{
	// DeviceService
	"/polyglot.v1.DeviceService/ListDevices":                  "device:read",
	"/polyglot.v1.DeviceService/GetDevice":                    "device:read",
	"/polyglot.v1.DeviceService/UpdateDevice":                 "device:manage",
	"/polyglot.v1.DeviceService/DeleteDevice":                 "device:manage",
	"/polyglot.v1.DeviceService/TestDeviceConnection":         "device:diagnostic:exec",
	"/polyglot.v1.DeviceService/StreamDeviceStatus":           "device:read",
	"/polyglot.v1.DeviceService/StreamPing":                   "device:diagnostic:exec",
	"/polyglot.v1.DeviceService/StreamInterfaceTraffic":       "device:read",
	"/polyglot.v1.DeviceService/StreamTerminal":               "device:terminal:exec",
	"/polyglot.v1.DeviceService/GetDevicePingConfig":          "device:read",
	"/polyglot.v1.DeviceService/UpdateDevicePingConfig":       "device:manage",
	"/polyglot.v1.DeviceService/QueryDevicePingMetrics":       "device:read",
	"/polyglot.v1.DeviceService/GetIsolationStatus":           "device:read",
	"/polyglot.v1.DeviceService/CreateIsolationProfile":       "device:manage",
	"/polyglot.v1.DeviceService/UpdateIsolationProfile":       "device:manage",
	"/polyglot.v1.DeviceService/DeleteIsolationProfile":       "device:manage",
	"/polyglot.v1.DeviceService/GetRouterIntegrationScript":   "device:read",
	"/polyglot.v1.DeviceService/ApplyRouterIntegrationScript": "device:manage",

	// ProbeService
	"/polyglot.v1.ProbeService/ReportStatus":    "probe:read",
	"/polyglot.v1.ProbeService/StreamTelemetry": "probe:read",

	// CustomerService
	"/polyglot.v1.CustomerService/ListCustomers":      "customer:read",
	"/polyglot.v1.CustomerService/GetCustomer":        "customer:read",
	"/polyglot.v1.CustomerService/CreateCustomer":     "customer:manage",
	"/polyglot.v1.CustomerService/UpdateCustomer":     "customer:manage",
	"/polyglot.v1.CustomerService/DeleteCustomer":     "customer:manage",
	"/polyglot.v1.CustomerService/FindByPhone":        "customer:read",
	"/polyglot.v1.CustomerService/FindByCustomerCode": "customer:read",
	"/polyglot.v1.CustomerService/FindByPortalCode":   "customer:read",
	"/polyglot.v1.CustomerService/ListSubscriptions":  "customer:read",

	// WhatsAppService
	"/polyglot.v1.WhatsAppService/ListSessions":     "whatsapp:read",
	"/polyglot.v1.WhatsAppService/CreateSession":    "whatsapp:manage",
	"/polyglot.v1.WhatsAppService/GetQRCode":        "whatsapp:manage",
	"/polyglot.v1.WhatsAppService/GetPairingCode":   "whatsapp:manage",
	"/polyglot.v1.WhatsAppService/ToggleBot":        "whatsapp:manage",
	"/polyglot.v1.WhatsAppService/ReconnectSession": "whatsapp:manage",
	"/polyglot.v1.WhatsAppService/LogoutSession":    "whatsapp:manage",
	"/polyglot.v1.WhatsAppService/PurgeSession":     "whatsapp:manage",
	"/polyglot.v1.WhatsAppService/SendTextMessage":  "whatsapp:message",
	"/polyglot.v1.WhatsAppService/ListChats":        "whatsapp:read",
	"/polyglot.v1.WhatsAppService/GetChatMessages":  "whatsapp:read",
	"/polyglot.v1.WhatsAppService/MarkChatRead":     "whatsapp:write",
	"/polyglot.v1.WhatsAppService/ToggleChatBot":    "whatsapp:write",

	// BotService (conversations & rate limits & technicians)
	"/polyglot.v1.BotService/ListConversations":      "conversation:read",
	"/polyglot.v1.BotService/GetConversation":        "conversation:read",
	"/polyglot.v1.BotService/GetConversationContext": "conversation:read",
	"/polyglot.v1.BotService/TakeOverConversation":   "conversation:write",
	"/polyglot.v1.BotService/ResetConversationBot":   "conversation:write",
	"/polyglot.v1.BotService/CloseConversation":      "conversation:write",
	"/polyglot.v1.BotService/ResetRateLimit":         "conversation:write",
	"/polyglot.v1.BotService/GetRateLimitStatus":     "conversation:read",
	"/polyglot.v1.BotService/ListTechnicians":        "technician:read",
	"/polyglot.v1.BotService/CreateTechnician":       "technician:manage",
	"/polyglot.v1.BotService/UpdateTechnician":       "technician:manage",
	"/polyglot.v1.BotService/ToggleTechnicianActive": "technician:manage",
	"/polyglot.v1.BotService/DeleteTechnician":       "technician:manage",

	// SkillService (LocalAI Standard)
	"/polyglot.v1.SkillService/ListSkills":       "skill:read",
	"/polyglot.v1.SkillService/GetSkill":         "skill:read",
	"/polyglot.v1.SkillService/CreateSkill":      "skill:manage",
	"/polyglot.v1.SkillService/UpdateSkill":      "skill:manage",
	"/polyglot.v1.SkillService/DeleteSkill":      "skill:manage",
	"/polyglot.v1.SkillService/ExportSkill":      "skill:read",
	"/polyglot.v1.SkillService/ImportSkill":      "skill:manage",
	"/polyglot.v1.SkillService/ListResources":    "skill:read",
	"/polyglot.v1.SkillService/GetResource":      "skill:read",
	"/polyglot.v1.SkillService/SaveResource":     "skill:manage",
	"/polyglot.v1.SkillService/DeleteResource":   "skill:manage",
	"/polyglot.v1.SkillService/ListGitRepos":     "skill:read",
	"/polyglot.v1.SkillService/AddGitRepo":       "skill:manage",
	"/polyglot.v1.SkillService/UpdateGitRepo":    "skill:manage",
	"/polyglot.v1.SkillService/DeleteGitRepo":    "skill:manage",
	"/polyglot.v1.SkillService/SyncGitRepo":      "skill:manage",
	"/polyglot.v1.SkillService/ToggleGitRepo":    "skill:manage",
	"/polyglot.v1.SkillService/ToggleSkill":      "skill:manage",
	"/polyglot.v1.SkillService/GetGlobalPrompt":  "skill:read",
	"/polyglot.v1.SkillService/SaveGlobalPrompt": "skill:manage",

	// LLMConfigService
	"/polyglot.v1.LLMConfigService/ListLLMConfigs":    "llmconfig:read",
	"/polyglot.v1.LLMConfigService/CreateLLMConfig":   "llmconfig:manage",
	"/polyglot.v1.LLMConfigService/UpdateLLMConfig":   "llmconfig:manage",
	"/polyglot.v1.LLMConfigService/ActivateLLMConfig": "llmconfig:manage",
	"/polyglot.v1.LLMConfigService/TestLLMConfig":     "llmconfig:manage",
	"/polyglot.v1.LLMConfigService/DeleteLLMConfig":   "llmconfig:manage",

	// HotspotService (Mikhmon)
	"/polyglot.v1.HotspotService/ListProfiles":           "hotspot:profile:read",
	"/polyglot.v1.HotspotService/CreateProfile":          "hotspot:profile:write",
	"/polyglot.v1.HotspotService/UpdateProfile":          "hotspot:profile:write",
	"/polyglot.v1.HotspotService/DeleteProfile":          "hotspot:profile:write",
	"/polyglot.v1.HotspotService/ListUsers":              "hotspot:user:read",
	"/polyglot.v1.HotspotService/GetUser":                "hotspot:user:read",
	"/polyglot.v1.HotspotService/CreateUser":             "hotspot:user:write",
	"/polyglot.v1.HotspotService/UpdateUser":             "hotspot:user:write",
	"/polyglot.v1.HotspotService/ResetUserCounters":      "hotspot:user:write",
	"/polyglot.v1.HotspotService/DeleteUser":             "hotspot:user:write",
	"/polyglot.v1.HotspotService/DeleteHotspotUsers":     "hotspot:user:write",
	"/polyglot.v1.HotspotService/ListActiveSessions":     "hotspot:active:read",
	"/polyglot.v1.HotspotService/KickActiveSession":      "hotspot:active:kick",
	"/polyglot.v1.HotspotService/GenerateVouchers":       "hotspot:voucher:generate",
	"/polyglot.v1.HotspotService/GetVoucherBatch":        "hotspot:voucher:generate",
	"/polyglot.v1.HotspotService/CheckVoucherStatus":     "hotspot:voucher:generate",
	"/polyglot.v1.HotspotService/ListHosts":              "hotspot:host:read",
	"/polyglot.v1.HotspotService/RemoveHost":             "hotspot:host:manage",
	"/polyglot.v1.HotspotService/ListHotspotServers":     "hotspot:binding:read",
	"/polyglot.v1.HotspotService/ListHotspotIPBindings":  "hotspot:binding:read",
	"/polyglot.v1.HotspotService/CreateHotspotIPBinding": "hotspot:binding:manage",
	"/polyglot.v1.HotspotService/UpdateHotspotIPBinding": "hotspot:binding:manage",
	"/polyglot.v1.HotspotService/DeleteHotspotIPBinding": "hotspot:binding:manage",
	"/polyglot.v1.HotspotService/ListHotspotCookies":     "hotspot:binding:read",
	"/polyglot.v1.HotspotService/DeleteHotspotCookie":    "hotspot:binding:manage",
	"/polyglot.v1.HotspotService/ListReports":            "hotspot:report:read",
	"/polyglot.v1.HotspotService/DeleteReport":           "hotspot:report:manage",
	"/polyglot.v1.HotspotService/GetExpireMonitorStatus": "hotspot:expire:read",
	"/polyglot.v1.HotspotService/SetupExpireMonitor":     "hotspot:expire:manage",
	"/polyglot.v1.HotspotService/DisableExpireMonitor":   "hotspot:expire:manage",
	"/polyglot.v1.HotspotService/RemoveExpireMonitor":    "hotspot:expire:manage",
	"/polyglot.v1.HotspotService/ListTemplates":          "hotspot:template:read",
	"/polyglot.v1.HotspotService/GetTemplateSection":     "hotspot:template:read",
	"/polyglot.v1.HotspotService/RenderVouchers":         "hotspot:voucher:generate",

	// NetworkService (DHCP Leases, Parent Queues, IP Pools)
	"/polyglot.v1.NetworkService/ListDHCPLeases":   "network:read",
	"/polyglot.v1.NetworkService/BlockDHCPLease":   "network:manage",
	"/polyglot.v1.NetworkService/ListParentQueues": "network:read",
	"/polyglot.v1.NetworkService/ListIPPools":      "network:read",

	// NetworkMonitorService (Real-Time Streams)
	"/polyglot.v1.NetworkMonitorService/StreamTraffic":           "monitor:read",
	"/polyglot.v1.NetworkMonitorService/StreamResource":          "monitor:read",
	"/polyglot.v1.NetworkMonitorService/StreamActiveSessions":    "monitor:read",
	"/polyglot.v1.NetworkMonitorService/StreamActiveStats":       "monitor:read",
	"/polyglot.v1.NetworkMonitorService/StreamSystemSnapshot":    "monitor:read",
	"/polyglot.v1.NetworkMonitorService/StreamInterfaceEthernet": "monitor:read",
	"/polyglot.v1.NetworkMonitorService/StreamQueueStats":        "monitor:read",
	"/polyglot.v1.NetworkMonitorService/StreamLogs":              "log:read",
	"/polyglot.v1.NetworkMonitorService/StreamHotspotInactive":   "monitor:read",
	"/polyglot.v1.NetworkMonitorService/StreamPPPActive":         "monitor:read",
	"/polyglot.v1.NetworkMonitorService/StreamPPPInactive":       "monitor:read",

	// PPPService — manajemen PPPoE / PPP secrets, profiles, active/inactive sessions.
	"/polyglot.v1.PPPService/ListSecrets":           "ppp:secret:read",
	"/polyglot.v1.PPPService/GetSecret":             "ppp:secret:read",
	"/polyglot.v1.PPPService/CreateSecret":          "ppp:secret:write",
	"/polyglot.v1.PPPService/UpdateSecret":          "ppp:secret:write",
	"/polyglot.v1.PPPService/DeleteSecret":          "ppp:secret:write",
	"/polyglot.v1.PPPService/SetSecretDisabled":     "ppp:secret:write",
	"/polyglot.v1.PPPService/ListProfiles":          "ppp:profile:read",
	"/polyglot.v1.PPPService/GetProfile":            "ppp:profile:read",
	"/polyglot.v1.PPPService/CreateProfile":         "ppp:profile:write",
	"/polyglot.v1.PPPService/UpdateProfile":         "ppp:profile:write",
	"/polyglot.v1.PPPService/DeleteProfile":         "ppp:profile:write",
	"/polyglot.v1.PPPService/ListActiveSessions":    "ppp:active:read",
	"/polyglot.v1.PPPService/KickActiveSession":     "ppp:active:kick",
	"/polyglot.v1.PPPService/KickActiveSessions":    "ppp:active:kick",
	"/polyglot.v1.PPPService/ListInactiveSecrets":   "ppp:inactive:read",
	"/polyglot.v1.PPPService/StreamActiveSessions":  "ppp:active:read",
	"/polyglot.v1.PPPService/StreamInactiveSecrets": "ppp:inactive:read",

	// UserService — manajemen user (CRUD, reset password, aktif/nonaktif, device assignment).
	"/polyglot.v1.UserService/ListUsers":                 "user:read",
	"/polyglot.v1.UserService/CreateUser":                "user:manage",
	"/polyglot.v1.UserService/UpdateUser":                "user:manage",
	"/polyglot.v1.UserService/ResetPassword":             "user:manage",
	"/polyglot.v1.UserService/ToggleActive":              "user:manage",
	"/polyglot.v1.UserService/DeleteUser":                "user:manage",
	"/polyglot.v1.UserService/AssignDevicesToUser":       "user:manage",
	"/polyglot.v1.UserService/ListUserAccessibleDevices": "user:read",

	// RBACService — manajemen policy & role assignment, owner-only (rbac:manage)
	"/polyglot.v1.RBACService/ListPolicies":        "rbac:manage",
	"/polyglot.v1.RBACService/AddPolicy":           "rbac:manage",
	"/polyglot.v1.RBACService/RemovePolicy":        "rbac:manage",
	"/polyglot.v1.RBACService/ListRoleAssignments": "rbac:manage",
	"/polyglot.v1.RBACService/AssignRole":          "rbac:manage",
	"/polyglot.v1.RBACService/UnassignRole":        "rbac:manage",
	"/polyglot.v1.RBACService/SyncRolePermissions": "rbac:manage",
	"/polyglot.v1.RBACService/DeleteRole":          "rbac:manage",

	// AuthService — profil & ganti password akun sendiri
	"/polyglot.v1.AuthService/GetMe":          "profile:read",
	"/polyglot.v1.AuthService/UpdateMe":       "profile:write",
	"/polyglot.v1.AuthService/ChangePassword": "profile:write",

	// SettingService — konfigurasi sistem & bot dinamis
	"/polyglot.v1.SettingService/GetAllSettings":        "setting:read",
	"/polyglot.v1.SettingService/GetSettingsByCategory": "setting:read",
	"/polyglot.v1.SettingService/GetBotSettings":        "setting:read",
	"/polyglot.v1.SettingService/UpdateSetting":         "setting:manage",
	"/polyglot.v1.SettingService/BatchUpdateSettings":   "setting:manage",
	"/polyglot.v1.SettingService/UpdateBotSettings":     "setting:manage",

	// ─── PlanService ────────────────────────────────────────────────────────
	"/polyglot.v1.PlanService/ListPlans":  "plan:read",
	"/polyglot.v1.PlanService/GetPlan":    "plan:read",
	"/polyglot.v1.PlanService/CreatePlan": "plan:manage",
	"/polyglot.v1.PlanService/UpdatePlan": "plan:manage",
	"/polyglot.v1.PlanService/DeletePlan": "plan:manage",

	// ─── SubscriptionService ────────────────────────────────────────────────
	"/polyglot.v1.SubscriptionService/ListSubscriptions":     "subscription:read",
	"/polyglot.v1.SubscriptionService/GetSubscription":       "subscription:read",
	"/polyglot.v1.SubscriptionService/CreateSubscription":    "subscription:manage",
	"/polyglot.v1.SubscriptionService/UpdateSubscription":    "subscription:manage",
	"/polyglot.v1.SubscriptionService/DeleteSubscription":    "subscription:manage",
	"/polyglot.v1.SubscriptionService/ChangePlan":            "subscription:manage",
	"/polyglot.v1.SubscriptionService/SuspendSubscription":   "subscription:manage",
	"/polyglot.v1.SubscriptionService/ResumeSubscription":    "subscription:manage",
	"/polyglot.v1.SubscriptionService/TerminateSubscription": "subscription:manage",
	"/polyglot.v1.SubscriptionService/ActivateSubscription":  "subscription:manage",
	"/polyglot.v1.SubscriptionService/IsolateSubscription":   "subscription:manage",
	"/polyglot.v1.SubscriptionService/RestoreSubscription":   "subscription:manage",
	"/polyglot.v1.SubscriptionService/RetryProvision":        "subscription:manage",

	// ─── BillingService (Invoices & Cashier) ────────────────────────────────
	"/polyglot.v1.BillingService/ListInvoices":             "billing:read",
	"/polyglot.v1.BillingService/GetInvoice":               "billing:read",
	"/polyglot.v1.BillingService/GenerateInvoices":         "billing:manage",
	"/polyglot.v1.BillingService/CancelInvoice":            "billing:manage",
	"/polyglot.v1.BillingService/CashierResolve":           "billing:read",
	"/polyglot.v1.BillingService/CashierPay":               "billing:manage",
	"/polyglot.v1.BillingService/ListPaymentSubmissions":   "billing:read",
	"/polyglot.v1.BillingService/ApprovePaymentSubmission": "billing:manage",
	"/polyglot.v1.BillingService/RejectPaymentSubmission":  "billing:manage",
	"/polyglot.v1.BillingService/ReversePayment":           "billing:manage",
	"/polyglot.v1.BillingService/CreateInvoice":            "billing:manage",
	"/polyglot.v1.BillingService/ExportInvoices":           "billing:export",
	"/polyglot.v1.BillingService/ListPaymentMethods":       "billing:read",
	"/polyglot.v1.BillingService/SavePaymentMethod":        "billing:manage",
	"/polyglot.v1.BillingService/DeletePaymentMethod":      "billing:manage",
	"/polyglot.v1.BillingService/CreateMassPayment":        "billing:manage",
	"/polyglot.v1.BillingService/ListMassPayments":         "billing:read",
	"/polyglot.v1.BillingService/GetMassPayment":           "billing:read",
	"/polyglot.v1.BillingService/ListGatewayTransactions":  "billing:read",
	"/polyglot.v1.BillingService/SyncGatewayTransactions":  "billing:manage",
	"/polyglot.v1.BillingService/ListGatewaySettlements":   "billing:read",
	"/polyglot.v1.BillingService/SaveGatewaySettlement":    "billing:manage",

	// ─── DebtService (hutang-piutang) ───────────────────────────────────────
	"/polyglot.v1.DebtService/ListDebts":        "debt:read",
	"/polyglot.v1.DebtService/GetDebt":          "debt:read",
	"/polyglot.v1.DebtService/ListDebtPayments": "debt:read",
	"/polyglot.v1.DebtService/GetDebtSummary":   "debt:read",
	"/polyglot.v1.DebtService/CreateDebt":       "debt:manage",
	"/polyglot.v1.DebtService/UpdateDebt":       "debt:manage",
	"/polyglot.v1.DebtService/CancelDebt":       "debt:manage",
	"/polyglot.v1.DebtService/PayDebt":          "debt:manage",
	"/polyglot.v1.DebtService/ExportDebts":      "billing:export",

	// ─── RegistrationService ────────────────────────────────────────────────
	"/polyglot.v1.RegistrationService/ListRegistrations":      "registration:read",
	"/polyglot.v1.RegistrationService/GetRegistration":        "registration:read",
	"/polyglot.v1.RegistrationService/ApproveRegistration":    "registration:manage",
	"/polyglot.v1.RegistrationService/ScheduleInstall":        "registration:manage",
	"/polyglot.v1.RegistrationService/RecordSurvey":           "registration:install",
	"/polyglot.v1.RegistrationService/StartInstallation":      "registration:install",
	"/polyglot.v1.RegistrationService/MarkInstalled":          "registration:install",
	"/polyglot.v1.RegistrationService/RejectRegistration":     "registration:manage",
	"/polyglot.v1.RegistrationService/CancelRegistration":     "registration:manage",
	"/polyglot.v1.RegistrationService/ConvertRegistration":    "registration:manage",
	"/polyglot.v1.RegistrationService/ListRegistrationEvents": "registration:read",

	// Laporan operasional lanjutan (Fase 5)
	"/polyglot.v1.ReportService/ARAging":               "report:read",
	"/polyglot.v1.ReportService/GatewayReconciliation": "report:read",
	"/polyglot.v1.ReportService/PSBStats":              "report:read",
	"/polyglot.v1.ReportService/TaxReport":             "report:read",
	"/polyglot.v1.ReportService/SalesReport":           "report:read",
	"/polyglot.v1.ReportService/CollectionReport":      "report:read",
	"/polyglot.v1.ReportService/CashflowReport":        "report:read",
	"/polyglot.v1.ReportService/AgingDetail":           "report:read",
	"/polyglot.v1.ReportService/ClosingReport":         "report:read",
	"/polyglot.v1.ReportService/ExportReport":          "report:export",

	// ─── RouterJobService (antrean job router — Fase 4) ─────────────────────
	"/polyglot.v1.RouterJobService/ListRouterJobs":        "device:read",
	"/polyglot.v1.RouterJobService/RetryRouterJob":        "device:write",
	"/polyglot.v1.RouterJobService/ExportRouterJobScript": "device:read",

	// ─── OltService (inventori OLT/ONU HiOSO — PLAN-ACS-OLT-HIOSO.md) ───────
	"/polyglot.v1.OltService/ListOlts":             "olt:read",
	"/polyglot.v1.OltService/GetOlt":               "olt:read",
	"/polyglot.v1.OltService/TestOltConnection":    "olt:read",
	"/polyglot.v1.OltService/ListOnus":             "olt:read",
	"/polyglot.v1.OltService/GetOnu":               "olt:read",
	"/polyglot.v1.OltService/ListOnuEvents":        "olt:read",
	"/polyglot.v1.OltService/GetOnuOpticalHistory": "olt:read",
	"/polyglot.v1.OltService/CreateOlt":            "olt:manage",
	"/polyglot.v1.OltService/UpdateOlt":            "olt:manage",
	"/polyglot.v1.OltService/DeleteOlt":            "olt:manage",
	"/polyglot.v1.OltService/SyncOlt":              "olt:manage",
	"/polyglot.v1.OltService/SaveOltConfig":        "olt:manage",
	"/polyglot.v1.OltService/RenameOnu":            "ont:manage",
	"/polyglot.v1.OltService/RebootOnu":            "ont:manage",
	"/polyglot.v1.OltService/SetOnuAdminState":     "ont:manage",
	"/polyglot.v1.OltService/DeleteOnu":            "ont:manage",

	// ─── OntService (ACS TR-069 / ONT lifecycle — PLAN-ACS-OLT-HIOSO.md) ────
	"/polyglot.v1.OntService/GetAcsConfig":        "ont:read",
	"/polyglot.v1.OntService/ListAcsDevices":      "ont:read",
	"/polyglot.v1.OntService/SyncAcsDevices":      "ont:manage",
	"/polyglot.v1.OntService/RefreshOntTelemetry": "ont:manage",
	"/polyglot.v1.OntService/RebootOnt":           "ont:manage",
	"/polyglot.v1.OntService/SetOntWifi":          "ont:manage",
	"/polyglot.v1.OntService/PushOntPppoe":        "ont:manage",
	"/polyglot.v1.OntService/LinkOnt":             "ont:manage",
	"/polyglot.v1.OntService/UnlinkOnt":           "ont:manage",
	"/polyglot.v1.OntService/PinAcsDevice":        "ont:manage",
	"/polyglot.v1.OntService/UnpinAcsDevice":      "ont:manage",
	"/polyglot.v1.OntService/EnsureAcsPreset":     "ont:manage",

	// ─── CashbookService ────────────────────────────────────────────────────
	"/polyglot.v1.CashbookService/ListAccounts":     "cashbook:read",
	"/polyglot.v1.CashbookService/SaveAccount":      "cashbook:manage",
	"/polyglot.v1.CashbookService/ListCategories":   "cashbook:read",
	"/polyglot.v1.CashbookService/SaveCategory":     "cashbook:manage",
	"/polyglot.v1.CashbookService/AddTransaction":   "cashbook:manage",
	"/polyglot.v1.CashbookService/ListTransactions": "cashbook:read",
	"/polyglot.v1.CashbookService/Balances":         "cashbook:read",

	// ─── NotificationService ────────────────────────────────────────────────
	"/polyglot.v1.NotificationService/ListTemplates":          "notification:read",
	"/polyglot.v1.NotificationService/GetTemplate":            "notification:read",
	"/polyglot.v1.NotificationService/SaveTemplate":           "notification:manage",
	"/polyglot.v1.NotificationService/ListNotifications":      "notification:read",
	"/polyglot.v1.NotificationService/PendingCount":           "notification:read",
	"/polyglot.v1.NotificationService/MarkNotificationSent":   "notification:manage",
	"/polyglot.v1.NotificationService/MarkNotificationFailed": "notification:manage",
	"/polyglot.v1.NotificationService/TestSend":               "notification:manage",

	// ─── ReportService ──────────────────────────────────────────────────────
	"/polyglot.v1.ReportService/DailyReport":     "report:read",
	"/polyglot.v1.ReportService/MonthlyReport":   "report:read",
	"/polyglot.v1.ReportService/YearlyReport":    "report:read",
	"/polyglot.v1.ReportService/RefreshSnapshot": "report:manage",

	// ─── IspAdminService ────────────────────────────────────────────────────
	"/polyglot.v1.IspAdminService/ImportFile":          "ispadmin:manage",
	"/polyglot.v1.IspAdminService/ImportRouter":        "ispadmin:manage",
	"/polyglot.v1.IspAdminService/ExportCustomers":     "ispadmin:manage",
	"/polyglot.v1.IspAdminService/Reconcile":           "ispadmin:manage",
	"/polyglot.v1.IspAdminService/PullRouterPlans":     "ispadmin:manage",
	"/polyglot.v1.IspAdminService/CommitPlans":         "ispadmin:manage",
	"/polyglot.v1.IspAdminService/PullRouterCustomers": "ispadmin:manage",
	"/polyglot.v1.IspAdminService/CommitCustomers":     "ispadmin:manage",

	// ─── Plain-HTTP Staff Endpoints (JWT Guarded) ───────────────────────────
	"/api/cashier/charge":         "billing:manage",
	"/api/cashier/gateways":       "billing:read",
	"/api/reports/daily":          "report:read",
	"/api/reports/monthly":        "report:read",
	"/api/reports/yearly":         "report:read",
	"/api/admin/import":           "ispadmin:manage",
	"/api/admin/import-router":    "ispadmin:manage",
	"/api/admin/reconcile":        "ispadmin:manage",
	"/api/admin/export":           "ispadmin:manage",
	"/api/admin/snapshot/refresh": "ispadmin:manage",
}

// AlternativeProcedurePermissions maps procedures that can also be authorized
// by alternative permissions (e.g. portal-specific permissions like dial:*).
var AlternativeProcedurePermissions = map[string][]string{
	// Device & Monitor
	"/polyglot.v1.DeviceService/ListDevices": {
		"dial:device:read",
		"dial:dashboard:read",
		"dial:ppp-active:read",
		"dial:ppp-non-active:read",
		"dial:ppp-secrets:read",
		"dial:ppp-profiles:read",
		"dial:logs:read",
		"dial:monitor:read",
		"dial:settings:read",
	},
	"/polyglot.v1.DeviceService/GetDevice": {
		"dial:device:read",
		"dial:dashboard:read",
		"dial:ppp-active:read",
		"dial:ppp-non-active:read",
		"dial:ppp-secrets:read",
		"dial:ppp-profiles:read",
		"dial:logs:read",
		"dial:monitor:read",
		"dial:settings:read",
	},
	"/polyglot.v1.DeviceService/TestDeviceConnection": {
		"dial:monitor:read",
		"dial:dashboard:read",
		"dial:monitor:diagnostic",
		"dial:dashboard:ping",
	},
	"/polyglot.v1.DeviceService/StreamDeviceStatus": {
		"dial:device:read",
		"dial:dashboard:read",
		"dial:ppp-active:read",
		"dial:ppp-non-active:read",
		"dial:ppp-secrets:read",
		"dial:ppp-profiles:read",
		"dial:logs:read",
		"dial:monitor:read",
		"dial:settings:read",
	},
	"/polyglot.v1.DeviceService/StreamPing": {
		"dial:dashboard:read",
		"dial:ppp-active:read",
		"dial:monitor:read",
		"dial:dashboard:ping",
		"dial:ppp-active:ping",
	},
	"/polyglot.v1.DeviceService/StreamInterfaceTraffic": {
		"dial:monitor:read",
		"dial:dashboard:read",
	},

	// NetworkMonitorService
	"/polyglot.v1.NetworkMonitorService/StreamTraffic": {
		"dial:monitor:read",
		"dial:dashboard:read",
	},
	"/polyglot.v1.NetworkMonitorService/StreamResource": {
		"dial:monitor:read",
		"dial:dashboard:read",
	},
	"/polyglot.v1.NetworkMonitorService/StreamSystemSnapshot": {
		"dial:monitor:read",
		"dial:dashboard:read",
	},
	"/polyglot.v1.NetworkMonitorService/StreamInterfaceEthernet": {
		"dial:monitor:read",
	},
	"/polyglot.v1.NetworkMonitorService/StreamQueueStats": {
		"dial:monitor:read",
	},
	"/polyglot.v1.NetworkMonitorService/StreamLogs": {
		"dial:logs:read",
	},
	"/polyglot.v1.NetworkMonitorService/StreamPPPActive": {
		"dial:ppp-active:read",
		"dial:dashboard:read",
	},
	"/polyglot.v1.NetworkMonitorService/StreamPPPInactive": {
		"dial:ppp-non-active:read",
		"dial:dashboard:read",
	},

	// PPPService
	"/polyglot.v1.PPPService/ListActiveSessions": {
		"dial:ppp-active:read",
		"dial:dashboard:read",
	},
	"/polyglot.v1.PPPService/StreamActiveSessions": {
		"dial:ppp-active:read",
		"dial:dashboard:read",
	},
	"/polyglot.v1.PPPService/KickActiveSession": {
		"dial:ppp-active:read",
		"dial:dashboard:read",
		"dial:ppp-active:kick",
	},
	"/polyglot.v1.PPPService/KickActiveSessions": {
		"dial:ppp-active:read",
		"dial:dashboard:read",
		"dial:ppp-active:kick",
	},
	"/polyglot.v1.PPPService/ListInactiveSecrets": {
		"dial:ppp-non-active:read",
		"dial:dashboard:read",
	},
	"/polyglot.v1.PPPService/StreamInactiveSecrets": {
		"dial:ppp-non-active:read",
		"dial:dashboard:read",
	},
	"/polyglot.v1.PPPService/ListSecrets": {
		"dial:ppp-secrets:read",
		"dial:dashboard:read",
	},
	"/polyglot.v1.PPPService/GetSecret": {
		"dial:ppp-secrets:read",
		"dial:dashboard:read",
	},
	"/polyglot.v1.PPPService/CreateSecret": {
		"dial:ppp-secrets:read",
		"dial:ppp-secrets:write",
	},
	"/polyglot.v1.PPPService/UpdateSecret": {
		"dial:ppp-secrets:read",
		"dial:ppp-secrets:write",
	},
	"/polyglot.v1.PPPService/DeleteSecret": {
		"dial:ppp-secrets:read",
		"dial:ppp-secrets:write",
	},
	"/polyglot.v1.PPPService/SetSecretDisabled": {
		"dial:ppp-non-active:read",
		"dial:ppp-secrets:read",
		"dial:dashboard:read",
		"dial:ppp-non-active:toggle",
		"dial:ppp-secrets:write",
	},
	"/polyglot.v1.PPPService/ListProfiles": {
		"dial:ppp-profiles:read",
		"dial:ppp-secrets:read",
		"dial:dashboard:read",
	},
	"/polyglot.v1.PPPService/GetProfile": {
		"dial:ppp-profiles:read",
		"dial:dashboard:read",
	},

	// LogService
	"/polyglot.v1.LogService/StreamLogs": {
		"dial:logs:read",
	},
	"/polyglot.v1.LogService/ListLogs": {
		"dial:logs:read",
	},

	// UserService
	"/polyglot.v1.UserService/ListUsers": {
		"dial:settings:read",
		"dial:settings:users:read",
	},
	"/polyglot.v1.UserService/CreateUser": {
		"dial:settings:read",
		"dial:settings:users:manage",
	},
	"/polyglot.v1.UserService/UpdateUser": {
		"dial:settings:read",
		"dial:settings:users:manage",
	},
	"/polyglot.v1.UserService/ResetPassword": {
		"dial:settings:read",
		"dial:settings:users:manage",
	},
	"/polyglot.v1.UserService/ToggleActive": {
		"dial:settings:read",
		"dial:settings:users:manage",
	},
	"/polyglot.v1.UserService/DeleteUser": {
		"dial:settings:read",
		"dial:settings:users:manage",
	},
	"/polyglot.v1.UserService/AssignDevicesToUser": {
		"dial:settings:read",
		"dial:settings:devices:manage",
	},
	"/polyglot.v1.UserService/ListUserAccessibleDevices": {
		"dial:settings:read",
		"dial:settings:devices:read",
	},
}

// PermissionsFor returns all accepted permission objects for a ConnectRPC procedure,
// including its primary permission and any alternative portal permissions (e.g. dial:*).
// Unknown procedures return false — callers must deny (fail closed).
func PermissionsFor(procedure string) ([]string, bool) {
	primary, ok := ProcedurePermissions[procedure]
	if !ok {
		return nil, false
	}
	alts := AlternativeProcedurePermissions[procedure]
	if len(alts) == 0 {
		return []string{primary}, true
	}
	res := make([]string, 0, 1+len(alts))
	res = append(res, primary)
	res = append(res, alts...)
	return res, true
}

// PermissionFor returns the primary resource:action object for a ConnectRPC procedure.
// Unknown procedures return false — callers must deny (fail closed).
func PermissionFor(procedure string) (string, bool) {
	perm, ok := ProcedurePermissions[procedure]
	return perm, ok
}
