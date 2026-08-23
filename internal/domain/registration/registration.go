package registration

import "time"

// Status alur pendaftaran:
// PENDING -> APPROVED -> INSTALLED -> ACTIVE, atau REJECTED / CANCELLED.
const (
	StatusPending   = "PENDING"
	StatusApproved  = "APPROVED"
	StatusInstalled = "INSTALLED"
	StatusActive    = "ACTIVE"
	StatusRejected  = "REJECTED"
	StatusCancelled = "CANCELLED"
)

// Registration represents a new-customer signup flow (ringkas tanpa NIK/KTP)
// with admin review, technician scheduling and installation results
// (DATABASE-SCHEMA-ISP.md §2.3 — registrations).
type Registration struct {
	ID                   string     `json:"id"`
	TenantID             string     `json:"tenant_id"`
	RegistrationNo       string     `json:"registration_no"` // "REG-202608-0001"
	PlanID               string     `json:"plan_id"`
	FullName             string     `json:"full_name"`
	Phone                string     `json:"phone"` // nomor WhatsApp aktif
	Email                string     `json:"email,omitempty"`
	Address              string     `json:"address"`
	Latitude             *float64   `json:"latitude,omitempty"`
	Longitude            *float64   `json:"longitude,omitempty"`
	Notes                string     `json:"notes,omitempty"`
	Status               string     `json:"status"`
	ReviewedBy           *uint      `json:"reviewed_by,omitempty"` // users.id
	ReviewedAt           *time.Time `json:"reviewed_at,omitempty"`
	AdminNotes           string     `json:"admin_notes,omitempty"`
	ScheduledInstallDate *time.Time `json:"scheduled_install_date,omitempty"`
	// Jam pemasangan (TIME). Fix review: bukan VARCHAR(20).
	ScheduledInstallTime  *time.Time `json:"scheduled_install_time,omitempty"`
	AssignedTechnicianID  *uint      `json:"assigned_technician_id,omitempty"` // users.id teknisi
	InstalledAt           *time.Time `json:"installed_at,omitempty"`
	TechnicianNotes       string     `json:"technician_notes,omitempty"`
	CustomerID            string     `json:"customer_id,omitempty"` // hasil konversi
	SubscriptionID        string     `json:"subscription_id,omitempty"`
	InvoiceID             string     `json:"invoice_id,omitempty"`
	RejectedAt            *time.Time `json:"rejected_at,omitempty"`
	RejectedReason        string     `json:"rejected_reason,omitempty"`
	CancelledAt           *time.Time `json:"cancelled_at,omitempty"`
	CancelReason          string     `json:"cancel_reason,omitempty"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}
