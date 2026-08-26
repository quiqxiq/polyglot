package device

import (
	"time"
)

// DevicePingConfig defines per-device settings for continuous streaming ping telemetry.
type DevicePingConfig struct {
	Enabled       bool   `json:"enabled"`
	Target        string `json:"target"`
	RetentionDays int    `json:"retention_days"`
}

// DefaultPingConfig returns default fallback ping configuration.
func DefaultPingConfig() DevicePingConfig {
	return DevicePingConfig{
		Enabled:       false,
		Target:        "8.8.8.8",
		RetentionDays: 7,
	}
}

// PingMetricPoint represents a single recorded ping measurement row.
type PingMetricPoint struct {
	RecordedAt time.Time `json:"recorded_at"`
	DeviceID   string    `json:"device_id"`
	Target     string    `json:"target"`
	Seq        int       `json:"seq"`
	Size       int       `json:"size"`
	TTL        int       `json:"ttl"`
	RTTMS      float32   `json:"rtt_ms"`
	Status     string    `json:"status"`
	Sent       int       `json:"sent"`
	Received   int       `json:"received"`
	PacketLoss int       `json:"packet_loss"`
	MinRTTMS   *float32  `json:"min_rtt_ms,omitempty"`
	AvgRTTMS   *float32  `json:"avg_rtt_ms,omitempty"`
	MaxRTTMS   *float32  `json:"max_rtt_ms,omitempty"`
}

// PingSummary provides aggregate statistical metrics over a queried time window.
type PingSummary struct {
	MinRTT        float32 `json:"min_rtt"`
	AvgRTT        float32 `json:"avg_rtt"`
	MaxRTT        float32 `json:"max_rtt"`
	PacketLossPct float32 `json:"packet_loss_pct"`
	TotalSamples  int64   `json:"total_samples"`
}

// PingMetricsFilter specifies criteria for historical time-series queries.
type PingMetricsFilter struct {
	DeviceID       string
	StartTime      time.Time
	EndTime        time.Time
	BucketInterval string // e.g. "raw", "1m", "5m", "1h"
}
