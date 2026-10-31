package model

import (
	"time"

	"github.com/quixiq/polyglot/internal/domain/audit"
)

// CommandLogModel is the GORM model for command_logs (migrasi 000041
// PLAN-AI-ROUTER-OPS v2): jejak append-only setiap percobaan perintah
// perangkat, termasuk yang ditolak.
//
// ID memakai GENERATED ALWAYS AS IDENTITY sehingga dikosongkan saat INSERT;
// GORM mengisi balik dari RETURNING.
type CommandLogModel struct {
	ID       int64   `gorm:"column:id;primaryKey;autoIncrement"`
	TenantID string  `gorm:"column:tenant_id;type:text;not null;default:tenant-default"`
	DeviceID *string `gorm:"column:device_id;type:uuid"`

	ActorType string  `gorm:"column:actor_type;type:text;not null"`
	ActorID   *string `gorm:"column:actor_id;type:text"`
	SessionID *string `gorm:"column:session_id;type:text"`

	Mode        string  `gorm:"column:mode;type:text;not null"`
	Command     string  `gorm:"column:command;type:text;not null"`
	CommandArgs *string `gorm:"column:command_args;type:jsonb"`

	ApprovalID     *string `gorm:"column:approval_id;type:text"`
	RestorePointID *string `gorm:"column:restore_point_id;type:text"`

	Status        string  `gorm:"column:status;type:text;not null"`
	OutputSummary *string `gorm:"column:output_summary;type:text"`
	DurationMS    int64   `gorm:"column:duration_ms"`

	CreatedAt time.Time `gorm:"column:created_at"`
}

// TableName returns the database table name for command logs.
func (CommandLogModel) TableName() string {
	return "command_logs"
}

// ToDomain converts a command log database row to its domain representation.
func (m *CommandLogModel) ToDomain() audit.CommandLog {
	if m == nil {
		return audit.CommandLog{}
	}
	return audit.CommandLog{
		ID:             m.ID,
		TenantID:       m.TenantID,
		DeviceID:       derefOrEmpty(m.DeviceID),
		ActorType:      m.ActorType,
		ActorID:        derefOrEmpty(m.ActorID),
		SessionID:      derefOrEmpty(m.SessionID),
		Mode:           m.Mode,
		Command:        m.Command,
		CommandArgs:    aiRawFromPtr(m.CommandArgs),
		ApprovalID:     derefOrEmpty(m.ApprovalID),
		RestorePointID: derefOrEmpty(m.RestorePointID),
		Status:         m.Status,
		OutputSummary:  derefOrEmpty(m.OutputSummary),
		DurationMS:     m.DurationMS,
		CreatedAt:      m.CreatedAt,
	}
}

// CommandLogModelFromDomain converts a command log domain entity to a database
// row. Kolom nullable diisi NULL untuk nilai string kosong.
func CommandLogModelFromDomain(e audit.CommandLog) *CommandLogModel {
	tenantID := e.TenantID
	if tenantID == "" {
		tenantID = aiDefaultTenantID
	}
	return &CommandLogModel{
		ID:             e.ID,
		TenantID:       tenantID,
		DeviceID:       aiNullableTextID(e.DeviceID),
		ActorType:      e.ActorType,
		ActorID:        ptrIfNotEmpty(e.ActorID),
		SessionID:      ptrIfNotEmpty(e.SessionID),
		Mode:           e.Mode,
		Command:        e.Command,
		CommandArgs:    jsonStringPtr(e.CommandArgs),
		ApprovalID:     ptrIfNotEmpty(e.ApprovalID),
		RestorePointID: ptrIfNotEmpty(e.RestorePointID),
		Status:         e.Status,
		OutputSummary:  ptrIfNotEmpty(e.OutputSummary),
		DurationMS:     e.DurationMS,
		CreatedAt:      e.CreatedAt,
	}
}
