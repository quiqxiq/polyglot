package model

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/quixiq/polyglot/internal/domain/approval"
)

// aiDefaultTenantID adalah tenant tunggal deployment saat ini, konsisten
// dengan default SQL kolom tenant_id di migrasi 000041.
const aiDefaultTenantID = "tenant-default"

// ApprovalModel is the GORM model for approval_requests (migrasi 000041
// PLAN-AI-ROUTER-OPS v2): satu permintaan human-in-the-loop untuk setiap
// mutasi router yang diajukan AI/MCP.
//
// Args/Diff/Validation/ExecutionResult disimpan sebagai JSONB; model memakai
// string agar round-trip tetap lossless terhadap JSON mentah domain.
type ApprovalModel struct {
	ID        string  `gorm:"column:id;primaryKey"`
	TenantID  string  `gorm:"column:tenant_id;type:text;not null;default:tenant-default"`
	DeviceID  *string `gorm:"column:device_id;type:uuid"`
	SessionID string  `gorm:"column:session_id;type:text;not null;default:''"`

	ActorType   string `gorm:"column:actor_type;type:text;not null"`
	ActorID     string `gorm:"column:actor_id;type:text;not null;default:''"`
	RequestedBy string `gorm:"column:requested_by;type:text;not null;default:''"`

	Command       string  `gorm:"column:command;type:text;not null"`
	CommandArgs   string  `gorm:"column:command_args;type:jsonb;not null;default:'{}'"`
	Revision      int     `gorm:"column:revision;not null;default:1"`
	EditedCommand *string `gorm:"column:edited_command;type:text"`
	EditedArgs    *string `gorm:"column:edited_args;type:jsonb"`
	EditedBy      *uint   `gorm:"column:edited_by"`
	EditCount     int     `gorm:"column:edit_count;not null;default:0"`

	Diff           string  `gorm:"column:diff;type:jsonb;not null;default:'{}'"`
	Validation     *string `gorm:"column:validation;type:jsonb"`
	RestorePointID *string `gorm:"column:restore_point_id;type:text"`

	Status     string `gorm:"column:status;type:text;not null;default:PENDING"`
	Reason     string `gorm:"column:reason;type:text"`
	ApprovedBy *uint  `gorm:"column:approved_by"`
	DecidedBy  *uint  `gorm:"column:decided_by"`

	IdempotencyKey  *string `gorm:"column:idempotency_key;type:text"`
	ExecutionNonce  *string `gorm:"column:execution_nonce;type:text"`
	ExecutionResult *string `gorm:"column:execution_result;type:jsonb"`

	CreatedAt  time.Time  `gorm:"column:created_at"`
	UpdatedAt  time.Time  `gorm:"column:updated_at"`
	ExpiresAt  time.Time  `gorm:"column:expires_at;not null"`
	DecidedAt  *time.Time `gorm:"column:decided_at"`
	ExecutedAt *time.Time `gorm:"column:executed_at"`
}

// TableName returns the database table name for approval requests.
func (ApprovalModel) TableName() string {
	return "approval_requests"
}

// ToDomain converts an approval database row to its domain representation.
func (m *ApprovalModel) ToDomain() approval.Request {
	if m == nil {
		return approval.Request{}
	}
	return approval.Request{
		ID:              m.ID,
		TenantID:        m.TenantID,
		DeviceID:        derefOrEmpty(m.DeviceID),
		SessionID:       m.SessionID,
		ActorType:       approval.ActorType(m.ActorType),
		ActorID:         m.ActorID,
		RequestedBy:     m.RequestedBy,
		Command:         m.Command,
		Args:            aiDecodeStringMap(m.CommandArgs),
		Revision:        m.Revision,
		EditedCommand:   m.EditedCommand,
		EditedArgs:      aiDecodeStringMapPtr(m.EditedArgs),
		EditedBy:        m.EditedBy,
		EditCount:       m.EditCount,
		Diff:            aiRawFromText(m.Diff),
		Validation:      aiRawFromPtr(m.Validation),
		RestorePointID:  derefOrEmpty(m.RestorePointID),
		Status:          approval.Status(m.Status),
		Reason:          m.Reason,
		ApprovedBy:      m.ApprovedBy,
		DecidedBy:       m.DecidedBy,
		IdempotencyKey:  derefOrEmpty(m.IdempotencyKey),
		ExecutionNonce:  derefOrEmpty(m.ExecutionNonce),
		ExecutionResult: aiRawFromPtr(m.ExecutionResult),
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
		ExpiresAt:       m.ExpiresAt,
		DecidedAt:       m.DecidedAt,
		ExecutedAt:      m.ExecutedAt,
	}
}

// ApprovalModelFromDomain converts an approval domain entity to a database
// row. Nilai wajib diberi default agar kolom NOT NULL + CHECK constraint
// selalu terpenuhi tanpa bergantung pada default SQL.
func ApprovalModelFromDomain(r approval.Request) *ApprovalModel {
	tenantID := r.TenantID
	if tenantID == "" {
		tenantID = aiDefaultTenantID
	}
	actorType := string(r.ActorType)
	if actorType == "" {
		actorType = string(approval.ActorAgent)
	}
	status := string(r.Status)
	if status == "" {
		status = string(approval.StatusPending)
	}
	revision := r.Revision
	if revision <= 0 {
		revision = 1
	}
	expiresAt := r.ExpiresAt
	if expiresAt.IsZero() {
		// Pengaman: expires_at NOT NULL; usecase selalu mengisi nilai nyata.
		expiresAt = r.CreatedAt.Add(15 * time.Minute)
		if r.CreatedAt.IsZero() {
			expiresAt = time.Now().UTC().Add(15 * time.Minute)
		}
	}
	return &ApprovalModel{
		ID:              r.ID,
		TenantID:        tenantID,
		DeviceID:        aiNullableTextID(r.DeviceID),
		SessionID:       r.SessionID,
		ActorType:       actorType,
		ActorID:         r.ActorID,
		RequestedBy:     r.RequestedBy,
		Command:         r.Command,
		CommandArgs:     aiEncodeStringMap(r.Args),
		Revision:        revision,
		EditedCommand:   r.EditedCommand,
		EditedArgs:      aiEncodeStringMapPtr(r.EditedArgs),
		EditedBy:        r.EditedBy,
		EditCount:       r.EditCount,
		Diff:            aiJSONTextOr(r.Diff, "{}"),
		Validation:      jsonStringPtr(r.Validation),
		RestorePointID:  ptrIfNotEmpty(r.RestorePointID),
		Status:          status,
		Reason:          r.Reason,
		ApprovedBy:      r.ApprovedBy,
		DecidedBy:       r.DecidedBy,
		IdempotencyKey:  ptrIfNotEmpty(r.IdempotencyKey),
		ExecutionNonce:  ptrIfNotEmpty(r.ExecutionNonce),
		ExecutionResult: jsonStringPtr(r.ExecutionResult),
		CreatedAt:       r.CreatedAt,
		UpdatedAt:       r.UpdatedAt,
		ExpiresAt:       expiresAt,
		DecidedAt:       r.DecidedAt,
		ExecutedAt:      r.ExecutedAt,
	}
}

// ─── Helper JSONB bersama model AI Router Ops ─────────────────────────────

// aiNullableTextID mengubah string kosong menjadi NULL (kolom UUID/FK).
func aiNullableTextID(id string) *string {
	if strings.TrimSpace(id) == "" {
		return nil
	}
	return &id
}

// aiJSONTextOr mengembalikan JSON mentah apa adanya, atau fallback bila kosong
// (dipakai kolom JSONB NOT NULL seperti diff/manifest/scope).
func aiJSONTextOr(raw json.RawMessage, fallback string) string {
	if len(raw) == 0 {
		return fallback
	}
	return string(raw)
}

// aiRawFromText mengubah teks JSONB kembali menjadi JSON mentah.
func aiRawFromText(text string) json.RawMessage {
	if text == "" {
		return nil
	}
	return json.RawMessage(text)
}

// aiRawFromPtr mengubah pointer teks JSONB nullable menjadi JSON mentah.
func aiRawFromPtr(text *string) json.RawMessage {
	if text == nil || *text == "" {
		return nil
	}
	return json.RawMessage(*text)
}

// aiEncodeStringMap menyerialkan map argumen perintah menjadi objek JSON.
func aiEncodeStringMap(values map[string]string) string {
	if len(values) == 0 {
		return "{}"
	}
	encoded, err := json.Marshal(values)
	if err != nil {
		return "{}"
	}
	return string(encoded)
}

// aiEncodeStringMapPtr menyerialkan map opsional: nil → NULL (kolom nullable).
func aiEncodeStringMapPtr(values map[string]string) *string {
	if values == nil {
		return nil
	}
	encoded := aiEncodeStringMap(values)
	return &encoded
}

// aiDecodeStringMap membaca objek JSONB menjadi map argumen.
func aiDecodeStringMap(text string) map[string]string {
	if strings.TrimSpace(text) == "" || text == "null" {
		return nil
	}
	var values map[string]string
	if err := json.Unmarshal([]byte(text), &values); err != nil {
		return nil
	}
	return values
}

// aiDecodeStringMapPtr membaca objek JSONB nullable menjadi map opsional.
func aiDecodeStringMapPtr(text *string) map[string]string {
	if text == nil {
		return nil
	}
	return aiDecodeStringMap(*text)
}

// aiEncodeStringSlice menyerialkan slice string menjadi array JSON.
func aiEncodeStringSlice(values []string) string {
	if len(values) == 0 {
		return "[]"
	}
	encoded, err := json.Marshal(values)
	if err != nil {
		return "[]"
	}
	return string(encoded)
}

// aiDecodeStringSlice membaca array JSONB menjadi slice string.
func aiDecodeStringSlice(text string) []string {
	if strings.TrimSpace(text) == "" || text == "null" {
		return nil
	}
	var values []string
	if err := json.Unmarshal([]byte(text), &values); err != nil {
		return nil
	}
	return values
}
