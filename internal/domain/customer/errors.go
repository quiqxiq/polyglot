package customer

import "errors"

var (
	ErrNotFound      = errors.New("customer not found")
	ErrNameRequired  = errors.New("customer name is required")
	ErrPhoneRequired = errors.New("customer phone is required")
	ErrAlreadyExists = errors.New("customer code already exists")
)
