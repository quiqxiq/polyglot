package customer

import "github.com/quixiq/polyglot/pkg/fault"

// Sentinel errors for the customer domain: system users (owner, admin, agent,
// teknisi), customers, and portal authentication. Message convention:
// "<entity>: <description>", lowercase English.
var (
	// User account errors.
	ErrUserNotFound              = fault.New(fault.KindNotFound, "user: not found")
	ErrUserAlreadyExists         = fault.New(fault.KindAlreadyExists, "user: username or email already taken")
	ErrUsernameRequired          = fault.New(fault.KindInvalidInput, "user: username is required")
	ErrPasswordTooShort          = fault.New(fault.KindInvalidInput, "user: password must be at least 8 characters")
	ErrInvalidRole               = fault.New(fault.KindInvalidInput, "user: invalid role")
	ErrSelfOperation             = fault.New(fault.KindPermissionDenied, "user: cannot perform this operation on your own account")
	ErrCannotModifyOwner         = fault.New(fault.KindPermissionDenied, "user: owner account can only be modified by itself")
	ErrCannotModifyAdmin         = fault.New(fault.KindPermissionDenied, "user: admin account can only be modified by itself or owner")
	ErrAdminCannotCreateAdmin    = fault.New(fault.KindPermissionDenied, "user: admin can only create accounts with role agent or teknisi")
	ErrAdminCannotAssignElevated = fault.New(fault.KindPermissionDenied, "user: admin can only assign role agent or teknisi")
	ErrCannotAssignOwnerRole     = fault.New(fault.KindPermissionDenied, "user: only owner can assign owner role")
	ErrLastOwnerDemotion         = fault.New(fault.KindFailedPrecondition, "user: system requires at least one active owner")
	ErrUnauthorizedDeviceAssign  = fault.New(fault.KindPermissionDenied, "user: cannot assign device outside your assigned devices")

	// Authentication errors (login, password change).
	ErrInvalidCredentials = fault.New(fault.KindUnauthenticated, "auth: invalid username or password")
	ErrAccountInactive    = fault.New(fault.KindPermissionDenied, "auth: account is disabled")
	ErrTooManyAttempts    = fault.New(fault.KindResourceExhausted, "auth: too many login attempts, try again later")
	ErrEmailAlreadyUsed   = fault.New(fault.KindAlreadyExists, "user: email already in use by another account")
	ErrWrongPassword      = fault.New(fault.KindInvalidInput, "auth: current password is incorrect")

	// Customer and portal errors.
	ErrCustomerNotFound     = fault.New(fault.KindNotFound, "customer: not found")
	ErrInvalidInput         = fault.New(fault.KindInvalidInput, "customer: invalid input")
	ErrPortalBadCredentials = fault.New(fault.KindUnauthenticated, "portal: invalid portal code, phone number, or OTP")
	ErrOTPLocked            = fault.New(fault.KindResourceExhausted, "portal: otp locked: too many failed attempts")
	ErrOTPExpired           = fault.New(fault.KindUnauthenticated, "portal: otp expired")
	ErrOTPNotFound          = fault.New(fault.KindNotFound, "portal: otp not found or already used")

	// Customer import errors (CSV/XLSX bulk import).
	ErrImportFileEmpty = fault.New(fault.KindInvalidInput, "customer: import file empty or header only")
	ErrImportHeader    = fault.New(fault.KindInvalidInput, "customer: unrecognized import header, download the export template as reference")
	ErrImportNoSheet   = fault.New(fault.KindInvalidInput, "customer: import xlsx has no sheet")
)
