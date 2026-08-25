package plan

import "errors"

var (
	ErrNotFound           = errors.New("plan not found")
	ErrNameRequired       = errors.New("plan name is required")
	ErrInvalidServiceType = errors.New("service type must be PPPOE or HOTSPOT")
	ErrInvalidRate        = errors.New("rate down/up kbps must be greater than zero")
	ErrInvalidPrice       = errors.New("price must not be negative")
	ErrInvalidBurst       = errors.New("burst must be all-or-none: burst rate >= base rate and threshold between base and burst")
	ErrAlreadyExists      = errors.New("plan name already exists")
)
