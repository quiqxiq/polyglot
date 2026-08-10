package device

import (
	"testing"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	iconnect "github.com/quixiq/polyglot/internal/adapter/connect"
)

func TestParsePingLatency(t *testing.T) {
	tests := []struct {
		name        string
		row         map[string]string
		wantLatency int64
		wantStatus  string
	}{
		{
			name:        "Standard ping response",
			row:         map[string]string{"seq": "0", "host": "8.8.8.8", "time": "23ms"},
			wantLatency: 23,
			wantStatus:  "connected",
		},
		{
			name:        "Sub-millisecond latency microsecond",
			row:         map[string]string{"seq": "0", "host": "8.8.8.8", "time": "230us"},
			wantLatency: 1,
			wantStatus:  "connected",
		},
		{
			name:        "Compound latency ms and us",
			row:         map[string]string{"seq": "0", "host": "8.8.8.8", "time": "1ms200us"},
			wantLatency: 1,
			wantStatus:  "connected",
		},
		{
			name:        "Prefix less than latency",
			row:         map[string]string{"seq": "0", "host": "8.8.8.8", "time": "<1ms"},
			wantLatency: 1,
			wantStatus:  "connected",
		},
		{
			name:        "Duration HH:MM:SS format",
			row:         map[string]string{"seq": "0", "host": "8.8.8.8", "time": "00:00:00.023000"},
			wantLatency: 23,
			wantStatus:  "connected",
		},
		{
			name:        "Avg-rtt fallback",
			row:         map[string]string{"avg-rtt": "15ms"},
			wantLatency: 15,
			wantStatus:  "connected",
		},
		{
			name:        "Timeout response",
			row:         map[string]string{"status": "timeout"},
			wantLatency: 0,
			wantStatus:  "timeout",
		},
		{
			name:        "Empty time string with seq",
			row:         map[string]string{"seq": "1", "host": "8.8.8.8"},
			wantLatency: 1,
			wantStatus:  "connected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLatency, gotStatus := parsePingLatency(tt.row)
			if gotLatency != tt.wantLatency {
				t.Errorf("parsePingLatency() latency = %v, want %v", gotLatency, tt.wantLatency)
			}
			if gotStatus != tt.wantStatus {
				t.Errorf("parsePingLatency() status = %v, want %v", gotStatus, tt.wantStatus)
			}
		})
	}
}

func TestConnectJSONCodec(t *testing.T) {
	codec := iconnect.JSONCodec()

	t.Run("StreamDeviceTrafficRequest camelCase", func(t *testing.T) {
		jsonData := []byte(`{"id":"dev123","interfaceName":"ether4"}`)
		var req devicepb.StreamDeviceTrafficRequest
		if err := codec.Unmarshal(jsonData, &req); err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}
		if req.Id != "dev123" {
			t.Errorf("got Id = %q, want %q", req.Id, "dev123")
		}
		if req.InterfaceName != "ether4" {
			t.Errorf("got InterfaceName = %q, want %q", req.InterfaceName, "ether4")
		}
	})

	t.Run("TestDeviceConnectionRequest camelCase", func(t *testing.T) {
		jsonData := []byte(`{"id":"dev123","selectedInterface":"ether2"}`)
		var req devicepb.TestDeviceConnectionRequest
		if err := codec.Unmarshal(jsonData, &req); err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}
		if req.SelectedInterface != "ether2" {
			t.Errorf("got SelectedInterface = %q, want %q", req.SelectedInterface, "ether2")
		}
	})
}
