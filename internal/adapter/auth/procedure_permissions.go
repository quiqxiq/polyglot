package auth

// ProcedurePermissions maps a ConnectRPC procedure (URL path) to a Casbin
// object in the form "resource:action" (e.g. "knowledge:write"). Enforcement
// uses this object against the RBAC policies; unknown procedures are denied
// (fail closed). Keep this registry in sync with the service handlers.
var ProcedurePermissions = map[string]string{
	// DeviceService
	"/polyglot.v1.DeviceService/ListDevices":            "device:read",
	"/polyglot.v1.DeviceService/GetDevice":              "device:read",
	"/polyglot.v1.DeviceService/UpdateDevice":           "device:manage",
	"/polyglot.v1.DeviceService/DeleteDevice":           "device:manage",
	"/polyglot.v1.DeviceService/TestDeviceConnection":   "device:command",
	"/polyglot.v1.DeviceService/StreamDeviceStatus":     "device:read",
	"/polyglot.v1.DeviceService/StreamPing":             "device:command",
	"/polyglot.v1.DeviceService/StreamInterfaceTraffic": "device:read",
	"/polyglot.v1.DeviceService/StreamTerminal":         "device:command",

	// ProbeService
	"/polyglot.v1.ProbeService/ReportStatus":    "probe:read",
	"/polyglot.v1.ProbeService/StreamTelemetry": "probe:read",

	// CustomerService
	"/polyglot.v1.CustomerService/ListCustomers": "customer:read",
	"/polyglot.v1.CustomerService/GetCustomer":   "customer:read",

	// IsolationService (isolir infrastructure per device)
	"/polyglot.v1.IsolationService/SetupIsolation":     "isolir:manage",
	"/polyglot.v1.IsolationService/GetIsolationStatus": "isolir:read",
	"/polyglot.v1.IsolationService/RemoveIsolation":    "isolir:manage",

	// BillingService
	"/polyglot.v1.BillingService/ListInvoices":       "billing:read",
	"/polyglot.v1.BillingService/GetInvoice":         "billing:read",
	"/polyglot.v1.BillingService/CreateInvoice":      "billing:write",
	"/polyglot.v1.BillingService/PayInvoice":         "billing:write",
	"/polyglot.v1.BillingService/ListSubscriptions":  "billing:read",
	"/polyglot.v1.BillingService/CreateSubscription": "billing:write",
	"/polyglot.v1.BillingService/CancelSubscription": "billing:write",

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

	// BotService (conversations & rate limits)
	"/polyglot.v1.BotService/ListConversations":      "conversation:read",
	"/polyglot.v1.BotService/GetConversation":        "conversation:read",
	"/polyglot.v1.BotService/GetConversationContext": "conversation:read",
	"/polyglot.v1.BotService/TakeOverConversation":   "conversation:write",
	"/polyglot.v1.BotService/ResetConversationBot":   "conversation:write",
	"/polyglot.v1.BotService/CloseConversation":      "conversation:write",
	"/polyglot.v1.BotService/ResetRateLimit":         "conversation:write",
	"/polyglot.v1.BotService/GetRateLimitStatus":     "conversation:read",

	// Skills Management (LocalAI Standard)
	"/polyglot.v1.BotService/ListSkills":       "skill:read",
	"/polyglot.v1.BotService/GetSkill":         "skill:read",
	"/polyglot.v1.BotService/CreateSkill":      "skill:manage",
	"/polyglot.v1.BotService/UpdateSkill":      "skill:manage",
	"/polyglot.v1.BotService/DeleteSkill":      "skill:manage",
	"/polyglot.v1.BotService/ExportSkill":      "skill:read",
	"/polyglot.v1.BotService/ImportSkill":      "skill:manage",
	"/polyglot.v1.BotService/ListResources":    "skill:read",
	"/polyglot.v1.BotService/GetResource":      "skill:read",
	"/polyglot.v1.BotService/SaveResource":     "skill:manage",
	"/polyglot.v1.BotService/DeleteResource":   "skill:manage",
	"/polyglot.v1.BotService/ListGitRepos":     "skill:read",
	"/polyglot.v1.BotService/AddGitRepo":       "skill:manage",
	"/polyglot.v1.BotService/UpdateGitRepo":    "skill:manage",
	"/polyglot.v1.BotService/DeleteGitRepo":    "skill:manage",
	"/polyglot.v1.BotService/SyncGitRepo":      "skill:manage",
	"/polyglot.v1.BotService/ToggleGitRepo":    "skill:manage",
	"/polyglot.v1.BotService/ToggleSkill":      "skill:manage",
	"/polyglot.v1.BotService/GetGlobalPrompt":  "skill:read",
	"/polyglot.v1.BotService/SaveGlobalPrompt": "skill:manage",

	// LLM Configs Management (BotService)
	"/polyglot.v1.BotService/ListLLMConfigs":    "llmconfig:read",
	"/polyglot.v1.BotService/CreateLLMConfig":   "llmconfig:manage",
	"/polyglot.v1.BotService/UpdateLLMConfig":   "llmconfig:manage",
	"/polyglot.v1.BotService/ActivateLLMConfig": "llmconfig:manage",
	"/polyglot.v1.BotService/TestLLMConfig":     "llmconfig:manage",
	"/polyglot.v1.BotService/DeleteLLMConfig":   "llmconfig:manage",

	// Technicians Management (BotService)
	"/polyglot.v1.BotService/ListTechnicians":        "technician:read",
	"/polyglot.v1.BotService/CreateTechnician":       "technician:manage",
	"/polyglot.v1.BotService/UpdateTechnician":       "technician:manage",
	"/polyglot.v1.BotService/ToggleTechnicianActive": "technician:manage",
	"/polyglot.v1.BotService/DeleteTechnician":       "technician:manage",

	// HotspotService (Mikhmon)
	"/polyglot.v1.HotspotService/ListProfiles":            "hotspot:read",
	"/polyglot.v1.HotspotService/ListUsers":               "hotspot:read",
	"/polyglot.v1.HotspotService/ListActiveSessions":      "hotspot:read",
	"/polyglot.v1.HotspotService/KickActiveSession":       "hotspot:manage",
	"/polyglot.v1.HotspotService/ListDHCPLeases":          "hotspot:read",
	"/polyglot.v1.HotspotService/BlockDHCPLease":          "hotspot:manage",
	"/polyglot.v1.HotspotService/GenerateVouchers":        "hotspot:manage",
	"/polyglot.v1.HotspotService/GetVoucherBatch":         "hotspot:read",
	"/polyglot.v1.HotspotService/GetUser":                 "hotspot:read",
	"/polyglot.v1.HotspotService/CreateUser":              "hotspot:manage",
	"/polyglot.v1.HotspotService/UpdateUser":              "hotspot:manage",
	"/polyglot.v1.HotspotService/ResetUserCounters":       "hotspot:manage",
	"/polyglot.v1.HotspotService/DeleteUser":              "hotspot:manage",
	"/polyglot.v1.HotspotService/DeleteHotspotUsers":      "hotspot:manage",
	"/polyglot.v1.HotspotService/CreateProfile":           "hotspot:manage",
	"/polyglot.v1.HotspotService/UpdateProfile":           "hotspot:manage",
	"/polyglot.v1.HotspotService/DeleteProfile":           "hotspot:manage",
	"/polyglot.v1.HotspotService/ListHosts":               "hotspot:read",
	"/polyglot.v1.HotspotService/RemoveHost":              "hotspot:manage",
	"/polyglot.v1.HotspotService/ListHotspotServers":      "hotspot:read",
	"/polyglot.v1.HotspotService/ListHotspotIPBindings":   "hotspot:read",
	"/polyglot.v1.HotspotService/CreateHotspotIPBinding":  "hotspot:manage",
	"/polyglot.v1.HotspotService/UpdateHotspotIPBinding":  "hotspot:manage",
	"/polyglot.v1.HotspotService/DeleteHotspotIPBinding":  "hotspot:manage",
	"/polyglot.v1.HotspotService/ListHotspotCookies":      "hotspot:read",
	"/polyglot.v1.HotspotService/DeleteHotspotCookie":     "hotspot:manage",
	"/polyglot.v1.HotspotService/CheckVoucherStatus":      "hotspot:read",
	"/polyglot.v1.HotspotService/StreamTraffic":           "hotspot:read",
	"/polyglot.v1.HotspotService/StreamResource":          "hotspot:read",
	"/polyglot.v1.HotspotService/StreamActiveSessions":    "hotspot:read",
	"/polyglot.v1.HotspotService/StreamActiveStats":       "hotspot:read",
	"/polyglot.v1.HotspotService/StreamSystemSnapshot":    "hotspot:read",
	"/polyglot.v1.HotspotService/StreamInterfaceEthernet": "hotspot:read",
	"/polyglot.v1.HotspotService/StreamQueueStats":        "hotspot:read",
	"/polyglot.v1.HotspotService/StreamLogs":              "log:read",
	"/polyglot.v1.HotspotService/StreamHotspotInactive":   "hotspot:read",
	"/polyglot.v1.HotspotService/StreamPPPActive":         "hotspot:read",
	"/polyglot.v1.HotspotService/StreamPPPInactive":       "hotspot:read",
	"/polyglot.v1.HotspotService/ListReports":             "hotspot:read",
	"/polyglot.v1.HotspotService/DeleteReport":            "hotspot:manage",
	"/polyglot.v1.HotspotService/GetExpireMonitorStatus":  "hotspot:read",
	"/polyglot.v1.HotspotService/SetupExpireMonitor":      "hotspot:manage",
	"/polyglot.v1.HotspotService/DisableExpireMonitor":    "hotspot:manage",
	"/polyglot.v1.HotspotService/RemoveExpireMonitor":     "hotspot:manage",
	"/polyglot.v1.HotspotService/ListTemplates":           "hotspot:read",
	"/polyglot.v1.HotspotService/GetTemplateSection":      "hotspot:read",
	"/polyglot.v1.HotspotService/RenderVouchers":          "hotspot:read",

	// PPPService — manajemen PPPoE / PPP secrets, profiles, active/inactive sessions.
	"/polyglot.v1.PPPService/ListSecrets":           "ppp:read",
	"/polyglot.v1.PPPService/GetSecret":             "ppp:read",
	"/polyglot.v1.PPPService/CreateSecret":          "ppp:manage",
	"/polyglot.v1.PPPService/UpdateSecret":          "ppp:manage",
	"/polyglot.v1.PPPService/DeleteSecret":          "ppp:manage",
	"/polyglot.v1.PPPService/SetSecretDisabled":     "ppp:manage",
	"/polyglot.v1.PPPService/ListProfiles":          "ppp:read",
	"/polyglot.v1.PPPService/GetProfile":            "ppp:read",
	"/polyglot.v1.PPPService/CreateProfile":         "ppp:manage",
	"/polyglot.v1.PPPService/UpdateProfile":         "ppp:manage",
	"/polyglot.v1.PPPService/DeleteProfile":         "ppp:manage",
	"/polyglot.v1.PPPService/ListActiveSessions":    "ppp:read",
	"/polyglot.v1.PPPService/KickActiveSession":     "ppp:manage",
	"/polyglot.v1.PPPService/KickActiveSessions":    "ppp:manage",
	"/polyglot.v1.PPPService/ListInactiveSecrets":   "ppp:read",
	"/polyglot.v1.PPPService/StreamActiveSessions":  "ppp:read",
	"/polyglot.v1.PPPService/StreamActiveStats":     "ppp:read",
	"/polyglot.v1.PPPService/StreamInactiveSecrets": "ppp:read",

	// UserService — manajemen user (CRUD, reset password, aktif/nonaktif).
	// user:read untuk melihat daftar, user:manage untuk mutasi.
	"/polyglot.v1.UserService/ListUsers":     "user:read",
	"/polyglot.v1.UserService/CreateUser":    "user:manage",
	"/polyglot.v1.UserService/UpdateUser":    "user:manage",
	"/polyglot.v1.UserService/ResetPassword": "user:manage",
	"/polyglot.v1.UserService/ToggleActive":  "user:manage",
	"/polyglot.v1.UserService/DeleteUser":    "user:manage",

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
}

// PermissionFor returns the resource:action object for a ConnectRPC procedure.
// Unknown procedures return false — callers must deny (fail closed).
func PermissionFor(procedure string) (string, bool) {
	perm, ok := ProcedurePermissions[procedure]
	return perm, ok
}
