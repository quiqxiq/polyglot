package auth

import "testing"

func TestPermissionFor(t *testing.T) {
	tests := []struct {
		name      string
		procedure string
		want      string
		wantOK    bool
	}{
		{"knowledge create", "/polyglot.v1.KnowledgeService/CreateKnowledge", "knowledge:write", true},
		{"device stream terminal", "/polyglot.v1.DeviceService/StreamTerminal", "device:command", true},
		{"rbac manage", "/polyglot.v1.RBACService/AssignRole", "rbac:manage", true},
		{"hotspot stream", "/polyglot.v1.HotspotService/StreamTraffic", "hotspot:read", true},
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
