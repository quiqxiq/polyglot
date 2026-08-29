package billing

import "github.com/quixiq/polyglot/pkg/fault"

// Sentinel errors for the billing domain: invoices, subscriptions, plans,
// payments, and payment gateway callbacks.
var (
	// ErrNotFound indicates the requested billing entity does not exist.
	ErrNotFound = fault.New(fault.KindNotFound, "billing: not found")
	// ErrInvalidInput indicates the request fails billing validation rules.
	ErrInvalidInput = fault.New(fault.KindInvalidInput, "billing: validation failed")
	// ErrInvalidTransition indicates an illegal subscription status transition.
	ErrInvalidTransition = fault.New(fault.KindFailedPrecondition, "billing: invalid subscription transition")
	// ErrPlanInUse indicates a plan cannot be deleted while active
	// subscriptions reference it.
	ErrPlanInUse = fault.New(fault.KindFailedPrecondition, "billing: plan in use by active subscriptions")
	// ErrGatewayTxMissingInvoice indicates a settled gateway transaction has
	// no linked invoice — data integrity issue on the provider callback path.
	ErrGatewayTxMissingInvoice = fault.New(fault.KindFailedPrecondition, "billing: gateway transaction has no invoice")

	// Payment processing errors.
	ErrInvoiceAlreadyPaid = fault.New(fault.KindConflict, "billing: invoice already paid")
	ErrInvoiceCancelled   = fault.New(fault.KindConflict, "billing: invoice cancelled")
	ErrOverpayment        = fault.New(fault.KindInvalidInput, "billing: payment amount exceeds outstanding balance")

	// Payment gateway callback errors.
	ErrGatewayDisabled   = fault.New(fault.KindFailedPrecondition, "billing: payment gateway disabled")
	ErrGatewayBadSign    = fault.New(fault.KindPermissionDenied, "billing: invalid callback signature")
	ErrGatewayUnknownRef = fault.New(fault.KindNotFound, "billing: unknown external_id")
	// ErrRepositoryUnavailable indicates a missing billing repository dependency.
	ErrRepositoryUnavailable = fault.New(fault.KindUnavailable, "billing: repository unavailable")
)
