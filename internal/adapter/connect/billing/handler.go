package billing

import (
	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	billingUC "github.com/quixiq/polyglot/internal/usecase/billing"
)

// BillingConnectHandler implements the billing ConnectRPC service.
//
//nolint:revive // Explicit transport role is part of the project naming convention.
type BillingConnectHandler struct {
	invoiceUC   *billingUC.InvoiceUseCase
	checkoutUC  *billingUC.CheckoutUseCase
	subUC       *billingUC.SubscriptionUseCase
	lifecycleUC *billingUC.SubscriptionLifecycleUseCase
	planUC      *billingUC.PlanUseCase
	runBilling  *billingUC.RunBillingUseCase
	manageSubUC *billingUC.ManageSubscriptionUseCase
}

// NewBillingConnectHandler constructs a billing ConnectRPC handler.
func NewBillingConnectHandler(
	invUC *billingUC.InvoiceUseCase,
	checkoutUC *billingUC.CheckoutUseCase,
	subUC *billingUC.SubscriptionUseCase,
	lifecycleUC *billingUC.SubscriptionLifecycleUseCase,
	planUC *billingUC.PlanUseCase,
	runBilling *billingUC.RunBillingUseCase,
	manageSubUC *billingUC.ManageSubscriptionUseCase,
) *BillingConnectHandler {
	return &BillingConnectHandler{
		invoiceUC:   invUC,
		checkoutUC:  checkoutUC,
		subUC:       subUC,
		lifecycleUC: lifecycleUC,
		planUC:      planUC,
		runBilling:  runBilling,
		manageSubUC: manageSubUC,
	}
}

type domainBillingInvoiceAlias = domainBilling.Invoice
