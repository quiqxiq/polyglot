package whatsapp

// SenderRole represents the resolved role of an inbound WhatsApp sender.
type SenderRole string

// Sender roles for inbound WhatsApp messages.
const (
	RoleAdmin      SenderRole = "admin"      // Full access (owner, superadmin, admin)
	RoleTechnician SenderRole = "technician" // Network ops, device diagnostics, ONT ops (teknisi)
	RoleCustomer   SenderRole = "customer"   // Self-service (invoices, subscriptions, tickets)
	RoleGuest      SenderRole = "guest"      // Unrecognized or public inquiry
)

// Standard permission scopes for sender roles.
const (
	ScopeAll            = "*"
	ScopeDeviceRead     = "device:read"
	ScopeDeviceDiagnose = "device:diagnostic:exec"
	ScopeOntOps         = "ont:ops"
	ScopeCustomerRead   = "customer:read"
	ScopeBillingRead    = "billing:read"
	ScopeSelfService    = "self:service"
	ScopeSelfBilling    = "self:billing"
	ScopeSelfTicket     = "self:ticket"
	ScopePublicInfo     = "public:info"
)

// DefaultScopesForRole returns the authorized permission scopes granted to a role.
func DefaultScopesForRole(role SenderRole) []string {
	switch role {
	case RoleAdmin:
		return []string{ScopeAll}
	case RoleTechnician:
		return []string{
			ScopeDeviceRead,
			ScopeDeviceDiagnose,
			ScopeOntOps,
			ScopeCustomerRead,
			ScopeBillingRead,
		}
	case RoleCustomer:
		return []string{
			ScopeSelfService,
			ScopeSelfBilling,
			ScopeSelfTicket,
		}
	default:
		return []string{ScopePublicInfo}
	}
}

// SenderIdentity encapsulates the resolved identity and access scopes of a WhatsApp sender.
type SenderIdentity struct {
	PhoneNumber   string     `json:"phone_number"` // Normalized international digits, e.g. 628123456789
	JID           string     `json:"jid"`          // WhatsApp JID, e.g. 628123456789@s.whatsapp.net
	Role          SenderRole `json:"role"`
	UserID        *uint      `json:"user_id,omitempty"`     // Associated internal staff user ID if staff
	CustomerID    string     `json:"customer_id,omitempty"` // Associated customer ID if subscriber
	DisplayName   string     `json:"display_name"`
	AllowedScopes []string   `json:"allowed_scopes"`
}

// HasScope checks whether the identity has the requested scope or the wildcard scope.
func (s SenderIdentity) HasScope(scope string) bool {
	for _, sc := range s.AllowedScopes {
		if sc == ScopeAll || sc == scope {
			return true
		}
	}
	return false
}
