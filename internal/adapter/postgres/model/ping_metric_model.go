package model

import (
	"time"

	"github.com/quixiq/polyglot/internal/domain/device"
)

// DevicePingMetricModel maps directly to the `device_ping_metrics` TimescaleDB hypertable.
type DevicePingMetricModel struct {
	RecordedAt time.Time `gorm:"column:recorded_at;not null"`
	DeviceID   string    `gorm:"column:device_id;type:uuid;not null;index:idx_device_ping_metrics_device_time,priority:1"`
	Target     string    `gorm:"column:target;type:varchar(255);not null"`
	Seq        int       `gorm:"column:seq;not null;default:0"`
	Size       int       `gorm:"column:size;not null;default:56"`
	TTL        int       `gorm:"column:ttl;not null;default:0"`
	RTTMS      float32   `gorm:"column:rtt_ms;not null;default:0"`
	Status     string    `gorm:"column:status;type:varchar(32);not null;default:'connected'"`
	Sent       int       `gorm:"column:sent;not null;default:1"`
	Received   int       `gorm:"column:received;not null;default:1"`
	PacketLoss int       `gorm:"column:packet_loss;type:smallint;not null;default:0"`
	MinRTTMS   *float32  `gorm:"column:min_rtt_ms"`
	AvgRTTMS   *float32  `gorm:"column:avg_rtt_ms"`
	MaxRTTMS   *float32  `gorm:"column:max_rtt_ms"`
}

// TableName explicitly maps to `device_ping_metrics`.
func (DevicePingMetricModel) TableName() string {
	return "device_ping_metrics"
}

// ToDomain maps database model to domain PingMetricPoint.
func (m DevicePingMetricModel) ToDomain() device.PingMetricPoint {
	return device.PingMetricPoint{
		RecordedAt: m.RecordedAt,
		DeviceID:   m.DeviceID,
		Target:     m.Target,
		Seq:        m.Seq,
		Size:       m.Size,
		TTL:        m.TTL,
		RTTMS:      m.RTTMS,
		Status:     m.Status,
		Sent:       m.Sent,
		Received:   m.Received,
		PacketLoss: m.PacketLoss,
		MinRTTMS:   m.MinRTTMS,
		AvgRTTMS:   m.AvgRTTMS,
		MaxRTTMS:   m.MaxRTTMS,
	}
}

// PingMetricModelFromDomain maps domain PingMetricPoint to GORM database model.
func PingMetricModelFromDomain(p device.PingMetricPoint) DevicePingMetricModel {
	if p.RecordedAt.IsZero() {
		p.RecordedAt = time.Now().UTC()
	}
	return DevicePingMetricModel{
		RecordedAt: p.RecordedAt,
		DeviceID:   p.DeviceID,
		Target:     p.Target,
		Seq:        p.Seq,
		Size:       p.Size,
		TTL:        p.TTL,
		RTTMS:      p.RTTMS,
		Status:     p.Status,
		Sent:       p.Sent,
		Received:   p.Received,
		PacketLoss: p.PacketLoss,
		MinRTTMS:   p.MinRTTMS,
		AvgRTTMS:   p.AvgRTTMS,
		MaxRTTMS:   p.MaxRTTMS,
	}
}
