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
