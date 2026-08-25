package subscription

import "errors"

var (
	ErrNotFound          = errors.New("subscription not found")
	ErrCustomerRequired  = errors.New("customer_id is required")
	ErrPlanRequired      = errors.New("plan_id is required")
	ErrInvalidServiceType = errors.New("service type must be PPPOE or HOTSPOT")
	ErrInvalidBillingDay = errors.New("billing day must be between 1 and 28")
)
