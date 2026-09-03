package billing

import (
	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	billingUC "github.com/quixiq/polyglot/internal/usecase/billing"
)

// BillingConnectHandler implements the billing & invoice ConnectRPC service.
//
//nolint:revive // Explicit transport role is part of the project naming convention.
type BillingConnectHandler struct {
	invoiceUC  *billingUC.InvoiceUseCase
	checkoutUC *billingUC.CheckoutUseCase
	runBilling *billingUC.RunBillingUseCase
}

// NewBillingConnectHandler constructs a billing ConnectRPC handler.
func NewBillingConnectHandler(
	invUC *billingUC.InvoiceUseCase,
	checkoutUC *billingUC.CheckoutUseCase,
	runBilling *billingUC.RunBillingUseCase,
) *BillingConnectHandler {
	return &BillingConnectHandler{
		invoiceUC:  invUC,
		checkoutUC: checkoutUC,
		runBilling: runBilling,
	}
}

type domainBillingInvoiceAlias = domainBilling.Invoice
