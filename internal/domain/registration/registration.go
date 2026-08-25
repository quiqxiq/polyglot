package registration

import "time"

// Status flow: PENDING → APPROVED → INSTALLED (provisioning triggered here)
// → ACTIVE; or terminal REJECTED / CANCELLED.
const (
	StatusPending   = "PENDING"
	StatusApproved  = "APPROVED"
	StatusInstalled = "INSTALLED"
	StatusActive    = "ACTIVE"
	StatusRejected  = "REJECTED"
	StatusCancelled = "CANCELLED"
)

// Registration is a new-subscriber application form (ringkas, tanpa NIK/KTP).
// On MarkInstalled the usecase layer converts it into a Customer +
// Subscription and provisions the account on the chosen Mikrotik router.
type Registration struct {
	ID                    string     `json:"id"`
	TenantID              string     `json:"tenant_id"`
	RegistrationNo        string     `json:"registration_no"`
	PlanID                string     `json:"plan_id"`
	FullName              string     `json:"full_name"`
	Phone                 string     `json:"phone"`
	Address               string     `json:"address"`
	Latitude              float64    `json:"latitude,omitempty"`
	Longitude             float64    `json:"longitude,omitempty"`
	Notes                 string     `json:"notes,omitempty"`
	Status                string     `json:"status"`
	ReviewedBy            *int64     `json:"reviewed_by,omitempty"`
	ReviewedAt            *time.Time `json:"reviewed_at,omitempty"`
	AdminNotes            string     `json:"admin_notes,omitempty"`
	ScheduledInstallDate  *time.Time `json:"scheduled_install_date,omitempty"`
	AssignedTechnicianID  *int64     `json:"assigned_technician_id,omitempty"`
	DeviceID              string     `json:"device_id,omitempty"`
	InstalledAt           *time.Time `json:"installed_at,omitempty"`
	TechnicianNotes       string     `json:"technician_notes,omitempty"`
	CustomerID            string     `json:"customer_id,omitempty"`
	SubscriptionID        string     `json:"subscription_id,omitempty"`
	RejectedAt            *time.Time `json:"rejected_at,omitempty"`
	RejectedReason        string     `json:"rejected_reason,omitempty"`
	CancelledAt           *time.Time `json:"cancelled_at,omitempty"`
	CancelReason          string     `json:"cancel_reason,omitempty"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

// Validate enforces the concise intake rules.
func (r Registration) Validate() error {
	switch {
	case r.PlanID == "":
		return ErrPlanRequired
	case r.FullName == "":
		return ErrNameRequired
	case r.Phone == "":
		return ErrPhoneRequired
	case r.Address == "":
		return ErrAddressRequired
	default:
		return nil
	}
}
