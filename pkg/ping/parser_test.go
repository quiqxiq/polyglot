package ping_test

import (
	"testing"

	"github.com/quixiq/polyglot/pkg/ping"
)

func TestParsePingLatency(t *testing.T) {
	tests := []struct {
		name         string
		row          map[string]string
		wantLatency  int64
		wantStatus   string
	}{
		{
			name:        "timeout status",
			row:         map[string]string{"status": "timeout"},
			wantLatency: 0,
			wantStatus:  "timeout",
		},
		{
			name:        "duration string 15ms",
			row:         map[string]string{"time": "15ms"},
			wantLatency: 15,
			wantStatus:  "connected",
		},
		{
			name:        "duration sub-millisecond 230us",
			row:         map[string]string{"time": "230us"},
			wantLatency: 1,
			wantStatus:  "connected",
		},
		{
			name:        "timestamp format 00:00:00.025000",
			row:         map[string]string{"time": "00:00:00.025000"},
			wantLatency: 25,
			wantStatus:  "connected",
		},
		{
			name:        "avg-rtt fallback",
			row:         map[string]string{"avg-rtt": "42ms"},
			wantLatency: 42,
			wantStatus:  "connected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLat, gotStat := ping.ParsePingLatency(tt.row)
			if gotLat != tt.wantLatency {
				t.Errorf("ParsePingLatency() latency = %v, want %v", gotLat, tt.wantLatency)
			}
			if gotStat != tt.wantStatus {
				t.Errorf("ParsePingLatency() status = %v, want %v", gotStat, tt.wantStatus)
			}
		})
	}
}

func TestParseDurationMs(t *testing.T) {
	tests := []struct {
		input string
		want  int64
	}{
		{"23ms", 23},
		{"1ms200us", 1},
		{"00:00:00.042000", 0},
		{"42ms", 42},
		{"", 0},
	}

	for _, tt := range tests {
		got := ping.ParseDurationMs(tt.input)
		if got != tt.want {
			t.Errorf("ParseDurationMs(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestParsePacketLoss(t *testing.T) {
	tests := []struct {
		input string
		want  int32
	}{
		{"0%", 0},
		{"5%", 5},
		{"100%", 100},
		{"15", 15},
		{"", 0},
	}

	for _, tt := range tests {
		got := ping.ParsePacketLoss(tt.input)
		if got != tt.want {
			t.Errorf("ParsePacketLoss(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}
