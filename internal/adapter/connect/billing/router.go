package billing

import (
	"net/http"

	"connectrpc.com/connect"

	iconnect "github.com/quixiq/polyglot/internal/adapter/connect"
	billingUC "github.com/quixiq/polyglot/internal/usecase/billing"
)

// NewBillingServiceHandler mounts BillingService Connect handlers (Invoice & Cashier).
func NewBillingServiceHandler(
	handler *BillingConnectHandler,
	paymentHandler *PaymentConnectHandler,
) (string, http.Handler) {
	mux := http.NewServeMux()
	opts := iconnect.DefaultHandlerOptions()

	serviceName := "polyglot.v1.BillingService"

	mux.Handle("/"+serviceName+"/ListInvoices", connect.NewUnaryHandler("/"+serviceName+"/ListInvoices", handler.ListInvoices, opts...))
	mux.Handle("/"+serviceName+"/GetInvoice", connect.NewUnaryHandler("/"+serviceName+"/GetInvoice", handler.GetInvoice, opts...))
	mux.Handle("/"+serviceName+"/CreateInvoice", connect.NewUnaryHandler("/"+serviceName+"/CreateInvoice", handler.CreateInvoice, opts...))
	mux.Handle("/"+serviceName+"/CancelInvoice", connect.NewUnaryHandler("/"+serviceName+"/CancelInvoice", handler.CancelInvoice, opts...))
	mux.Handle("/"+serviceName+"/ApplyInvoiceDiscount", connect.NewUnaryHandler("/"+serviceName+"/ApplyInvoiceDiscount", handler.ApplyInvoiceDiscount, opts...))
	mux.Handle("/"+serviceName+"/GenerateInvoices", connect.NewUnaryHandler("/"+serviceName+"/GenerateInvoices", handler.GenerateInvoices, opts...))
	mux.Handle("/"+serviceName+"/ExportInvoices", connect.NewUnaryHandler("/"+serviceName+"/ExportInvoices", handler.ExportInvoices, opts...))

	mux.Handle("/"+serviceName+"/CashierResolve", connect.NewUnaryHandler("/"+serviceName+"/CashierResolve", handler.CashierResolve, opts...))
	mux.Handle("/"+serviceName+"/CashierPay", connect.NewUnaryHandler("/"+serviceName+"/CashierPay", handler.CashierPay, opts...))

	mux.Handle("/"+serviceName+"/ListPaymentSubmissions", connect.NewUnaryHandler("/"+serviceName+"/ListPaymentSubmissions", paymentHandler.ListPaymentSubmissions, opts...))
	mux.Handle("/"+serviceName+"/ApprovePaymentSubmission", connect.NewUnaryHandler("/"+serviceName+"/ApprovePaymentSubmission", paymentHandler.ApprovePaymentSubmission, opts...))
	mux.Handle("/"+serviceName+"/RejectPaymentSubmission", connect.NewUnaryHandler("/"+serviceName+"/RejectPaymentSubmission", paymentHandler.RejectPaymentSubmission, opts...))
	mux.Handle("/"+serviceName+"/ReversePayment", connect.NewUnaryHandler("/"+serviceName+"/ReversePayment", paymentHandler.ReversePayment, opts...))

	mux.Handle("/"+serviceName+"/ListPaymentMethods", connect.NewUnaryHandler("/"+serviceName+"/ListPaymentMethods", handler.ListPaymentMethods, opts...))
	mux.Handle("/"+serviceName+"/SavePaymentMethod", connect.NewUnaryHandler("/"+serviceName+"/SavePaymentMethod", handler.SavePaymentMethod, opts...))
	mux.Handle("/"+serviceName+"/DeletePaymentMethod", connect.NewUnaryHandler("/"+serviceName+"/DeletePaymentMethod", handler.DeletePaymentMethod, opts...))

	mux.Handle("/"+serviceName+"/CreateMassPayment", connect.NewUnaryHandler("/"+serviceName+"/CreateMassPayment", handler.CreateMassPayment, opts...))
	mux.Handle("/"+serviceName+"/ListMassPayments", connect.NewUnaryHandler("/"+serviceName+"/ListMassPayments", handler.ListMassPayments, opts...))
	mux.Handle("/"+serviceName+"/GetMassPayment", connect.NewUnaryHandler("/"+serviceName+"/GetMassPayment", handler.GetMassPayment, opts...))

	mux.Handle("/"+serviceName+"/ListGatewayTransactions", connect.NewUnaryHandler("/"+serviceName+"/ListGatewayTransactions", handler.ListGatewayTransactions, opts...))
	mux.Handle("/"+serviceName+"/SyncGatewayTransactions", connect.NewUnaryHandler("/"+serviceName+"/SyncGatewayTransactions", handler.SyncGatewayTransactions, opts...))
	mux.Handle("/"+serviceName+"/ListGatewaySettlements", connect.NewUnaryHandler("/"+serviceName+"/ListGatewaySettlements", handler.ListGatewaySettlements, opts...))
	mux.Handle("/"+serviceName+"/SaveGatewaySettlement", connect.NewUnaryHandler("/"+serviceName+"/SaveGatewaySettlement", handler.SaveGatewaySettlement, opts...))

	return "/" + serviceName + "/", mux
}

// RegisterPublic mounts SubmitManualPayment and ResolveManualPayment directly on mux
// (allowing unauthenticated public access without conflicting with protected BillingService routes).
func RegisterPublic(mux *http.ServeMux, submissions *billingUC.PaymentSubmissionUseCase) {
	handler := NewPaymentConnectHandler(submissions, nil)
	opts := iconnect.DefaultHandlerOptions()
	serviceName := "polyglot.v1.BillingService"
	mux.Handle("/"+serviceName+"/ResolveManualPayment", connect.NewUnaryHandler(
		"/"+serviceName+"/ResolveManualPayment", handler.ResolveManualPayment, opts...))
	mux.Handle("/"+serviceName+"/SubmitManualPayment", connect.NewUnaryHandler(
		"/"+serviceName+"/SubmitManualPayment", handler.SubmitManualPayment, opts...))
}

// NewPublicBillingHandler mounts SubmitManualPayment secara publik (pelanggan
// mengunggah bukti transfer tanpa login, memakai kode bayar manual).
//
// Deprecated: use RegisterPublic directly on rootMux to avoid ServeMux subtree pattern conflicts.
func NewPublicBillingHandler(submissions *billingUC.PaymentSubmissionUseCase) (string, http.Handler) {
	mux := http.NewServeMux()
	RegisterPublic(mux, submissions)
	return "/polyglot.v1.BillingService/", mux
}
