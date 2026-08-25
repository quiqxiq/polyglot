package registration

import "errors"

var (
	ErrNotFound         = errors.New("registration not found")
	ErrPlanRequired     = errors.New("plan_id is required")
	ErrNameRequired     = errors.New("full name is required")
	ErrPhoneRequired    = errors.New("phone is required")
	ErrAddressRequired  = errors.New("address is required")
	ErrAlreadyPending   = errors.New("an active registration already exists for this phone")
	ErrInvalidTransition = errors.New("invalid registration status transition")
	ErrDeviceRequired   = errors.New("device_id is required to mark installation")
)
