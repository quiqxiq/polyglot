package audit

import "time"

// Tipe aktor yang melakukan aksi.
const (
	ActorUser   = "USER"
	ActorSystem = "SYSTEM"
	ActorPortal = "PORTAL"
)

// AuditLog records sensitive activity for traceability
// (DATABASE-SCHEMA-ISP.md §2.10 — audit_logs).
type AuditLog struct {
	ID          uint      `json:"id"`
	TenantID    string    `json:"tenant_id"`
	ActorType   string    `json:"actor_type"`
	ActorID     string    `json:"actor_id,omitempty"` // users.id / customer id / "-"
	Action      string    `json:"action"`             // 'CREATE_INVOICE', 'COLLECT_PAYMENT', ...
	EntityType  string    `json:"entity_type"`        // 'customer', 'invoice', 'payment', ...
	EntityID    string    `json:"entity_id"`
	Description string    `json:"description,omitempty"`
	IPAddress   string    `json:"ip_address,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}
