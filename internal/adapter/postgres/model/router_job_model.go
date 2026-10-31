package model

import (
	"encoding/json"
	"time"

	"github.com/quixiq/polyglot/internal/domain/routerjob"
)

// RouterJobModel is the GORM model for the asynchronous router job queue
// (planning v2 §3.5 — router_jobs).
type RouterJobModel struct {
	ID       string `gorm:"column:id;primaryKey"`
	TenantID string `gorm:"column:tenant_id;type:text;not null;default:tenant-default"`
	DeviceID string `gorm:"column:device_id;type:uuid;not null;index"`

	JobType string `gorm:"column:job_type;type:varchar(50);not null"`
	Payload string `gorm:"column:payload;type:jsonb;not null;default:'{}'"`

	IdempotencyKey string `gorm:"column:idempotency_key;type:varchar(100);unique;not null"`

	Status      string `gorm:"column:status;type:varchar(20);not null;default:QUEUED;index"`
	Attempts    int    `gorm:"column:attempts;not null;default:0"`
	MaxAttempts int    `gorm:"column:max_attempts;not null;default:5"`

	NextRunAt time.Time `gorm:"column:next_run_at;not null"`

	LastError string `gorm:"column:last_error;type:text"`

	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

// TableName returns the database table name for router jobs.
func (RouterJobModel) TableName() string {
	return "router_jobs"
}

// ToDomain converts a router job model to its domain representation.
func (m *RouterJobModel) ToDomain() routerjob.Job {
	if m == nil {
		return routerjob.Job{}
	}
	var payload json.RawMessage
	if m.Payload != "" {
		payload = json.RawMessage(m.Payload)
	}
	return routerjob.Job{
		ID:             m.ID,
		TenantID:       m.TenantID,
		DeviceID:       m.DeviceID,
		JobType:        routerjob.Type(m.JobType),
		Payload:        payload,
		IdempotencyKey: m.IdempotencyKey,
		Status:         routerjob.Status(m.Status),
		Attempts:       m.Attempts,
		MaxAttempts:    m.MaxAttempts,
		NextRunAt:      m.NextRunAt,
		LastError:      m.LastError,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
}

// RouterJobModelFromDomain converts a router job domain entity to a model.
func RouterJobModelFromDomain(j routerjob.Job) *RouterJobModel {
	payload := "{}"
	if len(j.Payload) > 0 {
		payload = string(j.Payload)
	}
	return &RouterJobModel{
		ID:             j.ID,
		TenantID:       j.TenantID,
		DeviceID:       j.DeviceID,
		JobType:        string(j.JobType),
		Payload:        payload,
		IdempotencyKey: j.IdempotencyKey,
		Status:         string(j.Status),
		Attempts:       j.Attempts,
		MaxAttempts:    j.MaxAttempts,
		NextRunAt:      j.NextRunAt,
		LastError:      j.LastError,
		CreatedAt:      j.CreatedAt,
		UpdatedAt:      j.UpdatedAt,
	}
}
