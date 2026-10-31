package billing

import (
	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	"github.com/quixiq/polyglot/internal/port"
	billingUC "github.com/quixiq/polyglot/internal/usecase/billing"
)

// BillingConnectHandler implements the billing & invoice ConnectRPC service.
//
//nolint:revive // Explicit transport role is part of the project naming convention.
type BillingConnectHandler struct {
	invoiceUC      *billingUC.InvoiceUseCase
	checkoutUC     *billingUC.CheckoutUseCase
	runBilling     *billingUC.RunBillingUseCase
	payMethodUC    *billingUC.PaymentMethodUseCase
	massPayUC      *billingUC.MassPaymentUseCase
	chargeUC       *billingUC.GatewayChargeUseCase
	settlementRepo port.GatewaySettlementRepository
	gwTxRepo       port.GatewayTransactionRepository
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

// WithPaymentMethods configures payment method usecase.
func (h *BillingConnectHandler) WithPaymentMethods(uc *billingUC.PaymentMethodUseCase) *BillingConnectHandler {
	h.payMethodUC = uc
	return h
}

// WithMassPayment configures mass payment usecase.
func (h *BillingConnectHandler) WithMassPayment(uc *billingUC.MassPaymentUseCase) *BillingConnectHandler {
	h.massPayUC = uc
	return h
}

// WithGateway configures gateway charge and settlement operations.
func (h *BillingConnectHandler) WithGateway(chargeUC *billingUC.GatewayChargeUseCase, settlements port.GatewaySettlementRepository, gwTx port.GatewayTransactionRepository) *BillingConnectHandler {
	h.chargeUC = chargeUC
	h.settlementRepo = settlements
	h.gwTxRepo = gwTx
	return h
}

type domainBillingInvoiceAlias = domainBilling.Invoice
