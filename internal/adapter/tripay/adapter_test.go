// Package tripay test suite (Fase 0 — PLAN-ISP-CORE-HARDENING.md).
//
// Kontrak yang dikunci:
//   - CreateCharge menandatangani (HMAC-SHA256) merchant_code+merchant_ref+amount
//     dan mengirim Authorization Bearer api_key.
//   - ParseWebhook memvalidasi signature callback dan memetakan status provider.
//   - CheckStatus menanyakan detail transaksi via ?reference= ke API provider.
//
// Test yang ditandai RED TEST mengunci perilaku benar yang BELUM diperbaiki;
// aktifkan dengan menghapus t.Skip setelah task fase 1 terkait selesai.
package tripay

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	domainCustomer "github.com/quixiq/polyglot/internal/domain/customer"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/internal/port/mocktest"
	billingUC "github.com/quixiq/polyglot/internal/usecase/billing"
)

const (
	testMerchantCode = "T12345"
	testAPIKey       = "API-KEY"
	testPrivateKey   = "PRIVATE-KEY"
)

func newTestAdapter(endpoint string) *Adapter {
	return &Adapter{
		reader: settingMap{
			"gw.tripay.enabled":       "true",
			"gw.tripay.endpoint":      endpoint,
			"gw.tripay.merchant_code": testMerchantCode,
			"gw.tripay.api_key":       testAPIKey,
			"gw.tripay.private_key":   testPrivateKey,
			"gw.tripay.channel":       "QRIS",
		},
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

// ─── CreateCharge ───────────────────────────────────────────────────────

func TestCreateCharge_SendsSignedRequestAndMapsResult(t *testing.T) {
	var captured map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/transaction/create", r.URL.Path)
		assert.Equal(t, "Bearer "+testAPIKey, r.Header.Get("Authorization"))
		require.NoError(t, json.NewDecoder(r.Body).Decode(&captured))
		_, _ = w.Write([]byte(`{"success":true,"data":{` +
			`"reference":"TREF-1","merchant_ref":"INV-1",` +
			`"checkout_url":"https://pay/x","qr_string":"QRDATA","pay_code":"PAYCODE",` +
			`"fee_merchant":700,"status":"UNPAID"}}`))
	}))
	defer srv.Close()

	a := newTestAdapter(srv.URL)
	res, err := a.CreateCharge(context.Background(), port.ChargeRequest{
		InvoiceNumber: "INV-1", Amount: 110000,
		CustomerName: "Budi", CustomerPhone: "081200000000",
	})
	require.NoError(t, err)

	assert.Equal(t, testMerchantCode, captured["merchant_code"])
	assert.Equal(t, "INV-1", captured["merchant_ref"])
	assert.Equal(t, "QRIS", captured["method"])
	// Signature create resmi: HMAC_SHA256(private_key, merchant_code+merchant_ref+amount).
	assert.Equal(t, signHMAC(testPrivateKey, testMerchantCode+"INV-1"+"110000"), captured["signature"])

	assert.Equal(t, "https://pay/x", res.PaymentURL)
	assert.Equal(t, "QRDATA", res.QRString)
	assert.Equal(t, "PAYCODE", res.VANumber)
	assert.InDelta(t, 700, res.FeeAmount, 0.01)
	assert.Equal(t, domainBilling.GatewayStatusPending, res.Status)
}

func TestCreateCharge_ExternalIDIsProviderReference(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"success":true,"data":{` +
			`"reference":"TREF-1","merchant_ref":"INV-1",` +
			`"checkout_url":"https://pay/x","status":"UNPAID"}}`))
	}))
	defer srv.Close()

	a := newTestAdapter(srv.URL)
	res, err := a.CreateCharge(context.Background(), port.ChargeRequest{InvoiceNumber: "INV-1", Amount: 110000})
	require.NoError(t, err)
	assert.Equal(t, "TREF-1", res.ExternalID)
}

func TestCreateCharge_RejectedByProvider(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"success":false,"message":"invalid merchant"}`))
	}))
	defer srv.Close()

	a := newTestAdapter(srv.URL)
	_, err := a.CreateCharge(context.Background(), port.ChargeRequest{InvoiceNumber: "INV-1", Amount: 1000})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "tripay reject")
}

func TestCreateCharge_TransportFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	url := srv.URL
	srv.Close()

	a := newTestAdapter(url)
	_, err := a.CreateCharge(context.Background(), port.ChargeRequest{InvoiceNumber: "INV-1", Amount: 1000})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "http tripay")
}

// ─── ParseWebhook ───────────────────────────────────────────────────────

func TestParseWebhook_RejectsBadSignature(t *testing.T) {
	a := newTestAdapter("https://tripay.invalid")
	body := []byte(`{"reference":"TREF-1","merchant_ref":"INV-1","status":"PAID","total_amount":110000}`)

	_, err := a.ParseWebhook(context.Background(), body, "deadbeef")
	assert.ErrorIs(t, err, domainBilling.ErrGatewayBadSign)
}

func TestParseWebhook_ValidatesRawBodySignature(t *testing.T) {
	a := newTestAdapter("https://tripay.invalid")
	body := []byte(`{"reference":"TREF-1","merchant_ref":"INV-1","status":"PAID",` +
		`"total_amount":110000,"amount_received":110000}`)

	ev, err := a.ParseWebhook(context.Background(), body, signHMAC(testPrivateKey, string(body)))
	require.NoError(t, err)
	assert.Equal(t, "TREF-1", ev.ExternalID)
	assert.Equal(t, "INV-1", ev.MerchantRef)
	assert.Equal(t, domainBilling.GatewayStatusSettled, ev.Status)
}

func TestParseWebhook_MapsProviderStatus(t *testing.T) {
	a := newTestAdapter("https://tripay.invalid")
	tests := []struct {
		provider string
		want     string
	}{
		{"PAID", domainBilling.GatewayStatusSettled},
		{"SETTLED", domainBilling.GatewayStatusSettled},
		{"EXPIRED", domainBilling.GatewayStatusExpired},
		{"FAILED", domainBilling.GatewayStatusFailed},
		{"UNPAID", domainBilling.GatewayStatusPending},
	}
	for _, tt := range tests {
		t.Run(tt.provider, func(t *testing.T) {
			body := []byte(`{"reference":"TREF-1","merchant_ref":"INV-1","status":"` + tt.provider +
				`","total_amount":110000,"amount_received":110000}`)
			ev, err := a.ParseWebhook(context.Background(), body, signHMAC(testPrivateKey, string(body)))
			require.NoError(t, err)
			assert.Equal(t, tt.want, ev.Status)
		})
	}
}

func TestParseWebhook_PaidAmountUsesTotalAmount(t *testing.T) {
	a := newTestAdapter("https://tripay.invalid")
	body := []byte(`{"reference":"TREF-1","merchant_ref":"INV-1","status":"PAID",` +
		`"total_amount":110000,"amount_received":108500}`)

	ev, err := a.ParseWebhook(context.Background(), body, signHMAC(testPrivateKey, string(body)))
	require.NoError(t, err)
	assert.InDelta(t, 110000, ev.PaidAmount, 0.01)
}

// ─── CheckStatus ────────────────────────────────────────────────────────

func TestCheckStatus_QueriesByReferenceAndMapsSettled(t *testing.T) {
	var gotRef, gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/transaction/detail", r.URL.Path)
		gotRef = r.URL.Query().Get("reference")
		gotAuth = r.Header.Get("Authorization")
		_, _ = w.Write([]byte(`{"success":true,"data":{` +
			`"reference":"TREF-1","merchant_ref":"INV-1","status":"PAID","amount":110000}}`))
	}))
	defer srv.Close()

	a := newTestAdapter(srv.URL)
	ev, err := a.CheckStatus(context.Background(), "TREF-1")
	require.NoError(t, err)

	assert.Equal(t, "TREF-1", gotRef)
	assert.Equal(t, "Bearer "+testAPIKey, gotAuth)
	assert.Equal(t, "TREF-1", ev.ExternalID)
	assert.Equal(t, "INV-1", ev.MerchantRef)
	assert.Equal(t, domainBilling.GatewayStatusSettled, ev.Status)
	assert.InDelta(t, 110000, ev.PaidAmount, 0.01)
}

func TestCheckStatus_DisabledGateway(t *testing.T) {
	a := &Adapter{
		reader: settingMap{"gw.tripay.private_key": ""},
		client: http.DefaultClient,
	}
	_, err := a.CheckStatus(context.Background(), "TREF-1")
	assert.ErrorIs(t, err, domainBilling.ErrGatewayDisabled)
}

// ─── Enabled ────────────────────────────────────────────────────────────

func TestEnabled_RequiresFlagAndPrivateKey(t *testing.T) {
	tests := []struct {
		name    string
		enabled string
		key     string
		want    bool
	}{
		{"flag aktif dan key ada", "true", testPrivateKey, true},
		{"flag nonaktif", "false", testPrivateKey, false},
		{"key kosong", "true", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Adapter{
				reader: settingMap{
					"gw.tripay.enabled":     tt.enabled,
					"gw.tripay.private_key": tt.key,
				},
				client: http.DefaultClient,
			}
			assert.Equal(t, tt.want, a.Enabled(context.Background()))
		})
	}
}

// ─── Integrasi adapter + GatewayChargeUseCase ───────────────────────────

type recordingProcessor struct {
	cmds []port.CashPaymentCommand
	pay  domainBilling.Payment
}

func (r *recordingProcessor) ProcessCashPayment(_ context.Context, cmd port.CashPaymentCommand) (domainBilling.Payment, error) {
	r.cmds = append(r.cmds, cmd)
	return r.pay, nil
}

// RED TEST — aktifkan setelah F1-1/F1-2/F1-3: alur produksi nyata
// create-charge → callback harus menemukan transaksi yang sama
// (ExternalID reference Tripay), dengan signature raw-body dan nominal penuh.
func TestAdapter_EndToEndWithGatewayUseCase_CreateThenWebhook(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"success":true,"data":{` +
			`"reference":"TREF-9","merchant_ref":"INV-E2E-9",` +
			`"checkout_url":"https://pay/e2e","qr_string":"QR9","pay_code":"CODE9",` +
			`"fee_merchant":700,"status":"UNPAID"}}`))
	}))
	defer srv.Close()

	a := newTestAdapter(srv.URL)
	invoices := mocktest.NewFakeInvoiceRepo()
	customers := mocktest.NewFakeCustomerRepo()
	gwt := mocktest.NewFakeGatewayTxRepo()
	proc := &recordingProcessor{pay: domainBilling.Payment{ID: "pay-e2e"}}
	usecase := billingUC.NewGatewayChargeUseCase(invoices, customers, gwt, a, proc, settingMap{})

	ctx := context.Background()
	require.NoError(t, invoices.Save(ctx, domainBilling.Invoice{
		ID: "inv-e2e", TenantID: "tenant-default", InvoiceNumber: "INV-E2E-9",
		CustomerID: "cust-e2e", Period: "2026-09", Total: 110000,
		Status:            domainBilling.StatusUnpaid,
		QRPayload:         "polyglot://invoice/inv-e2e",
		ManualPaymentCode: "PAY-E2E-9",
	}))
	require.NoError(t, customers.Save(ctx, domainCustomer.Customer{
		ID: "cust-e2e", TenantID: "tenant-default", CustomerCode: "CUST-E2E",
		Name: "Budi", Phone: "081200000000", Address: "Jl. E2E",
		Status: domainCustomer.StatusActive,
	}))

	res, _, err := usecase.CreateForInvoice(ctx, "inv-e2e", "", 60)
	require.NoError(t, err)
	require.Equal(t, "TREF-9", res.ExternalID)

	body := []byte(`{"reference":"TREF-9","merchant_ref":"INV-E2E-9","status":"PAID",` +
		`"total_amount":110000,"amount_received":108500}`)
	mac := hmac.New(sha256.New, []byte(testPrivateKey))
	mac.Write(body)
	sig := hex.EncodeToString(mac.Sum(nil))

	invoiceID, settled, err := usecase.HandleWebhook(ctx, body, sig)
	require.NoError(t, err)
	assert.True(t, settled)
	assert.Equal(t, "inv-e2e", invoiceID)
	require.Len(t, proc.cmds, 1)
	assert.InDelta(t, 110000, proc.cmds[0].Amount, 0.01)
}
