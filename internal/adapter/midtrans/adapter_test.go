// Midtrans adapter tests (F4-3): Snap create, notification signature, dan
// check-status Core API.
package midtrans

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	"github.com/quixiq/polyglot/internal/port"
)

const testServerKey = "SB-Mid-server-TEST"

func newTestAdapter(endpoint string) *Adapter {
	return &Adapter{
		reader: settingMap{
			"gw.midtrans.enabled":    "true",
			"gw.midtrans.endpoint":   endpoint,
			"gw.midtrans.server_key": testServerKey,
			"gw.midtrans.client_key": "SB-Mid-client-TEST",
			"gw.midtrans.channel":    "",
		},
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

func TestCreateCharge_SendsBasicAuthAndMapsSnapResponse(t *testing.T) {
	var captured map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/snap/v1/transactions", r.URL.Path)
		wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(testServerKey+":"))
		assert.Equal(t, wantAuth, r.Header.Get("Authorization"))
		require.NoError(t, json.NewDecoder(r.Body).Decode(&captured))
		_, _ = w.Write([]byte(`{"token":"snap-token","redirect_url":"https://app.sandbox.midtrans.com/snap/v1/redirect/abc"}`))
	}))
	defer srv.Close()

	a := newTestAdapter(srv.URL)
	res, err := a.CreateCharge(context.Background(), port.ChargeRequest{
		InvoiceNumber: "INV-MT-1", Amount: 110000,
		CustomerName: "Budi", CustomerPhone: "081200000000",
	})
	require.NoError(t, err)

	details := captured["transaction_details"].(map[string]any)
	assert.Equal(t, "INV-MT-1", details["order_id"])
	assert.InDelta(t, 110000, details["gross_amount"], 0.01)

	assert.Equal(t, "INV-MT-1", res.ExternalID)
	assert.Equal(t, "INV-MT-1", res.MerchantRef)
	assert.Equal(t, "https://app.sandbox.midtrans.com/snap/v1/redirect/abc", res.PaymentURL)
	assert.Equal(t, domainBilling.GatewayStatusPending, res.Status)
	assert.False(t, res.ExpiresAt.IsZero())
}

func TestCreateCharge_RejectedByProvider(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error_messages":["Access denied due to unauthorized transaction"]}`))
	}))
	defer srv.Close()

	a := newTestAdapter(srv.URL)
	_, err := a.CreateCharge(context.Background(), port.ChargeRequest{InvoiceNumber: "INV-1", Amount: 1000})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "midtrans reject")
}

func TestParseWebhook_ValidSignatureAndStatusMapping(t *testing.T) {
	a := newTestAdapter("https://midtrans.invalid")
	tests := []struct {
		name   string
		status string
		fraud  string
		want   string
	}{
		{"settlement", "settlement", "accept", domainBilling.GatewayStatusSettled},
		{"capture accept", "capture", "accept", domainBilling.GatewayStatusSettled},
		{"capture challenge", "capture", "challenge", domainBilling.GatewayStatusPending},
		{"pending", "pending", "", domainBilling.GatewayStatusPending},
		{"expire", "expire", "", domainBilling.GatewayStatusExpired},
		{"deny", "deny", "", domainBilling.GatewayStatusFailed},
		{"cancel", "cancel", "", domainBilling.GatewayStatusFailed},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gross := "110000.00"
			sig := sha512Hex("INV-MT-9" + "200" + gross + testServerKey)
			body := []byte(`{"transaction_status":"` + tt.status + `","fraud_status":"` + tt.fraud +
				`","order_id":"INV-MT-9","status_code":"200","gross_amount":"` + gross +
				`","signature_key":"` + sig + `","transaction_id":"tx-1","payment_type":"qris"}`)

			ev, err := a.ParseWebhook(context.Background(), body, "")
			require.NoError(t, err)
			assert.Equal(t, tt.want, ev.Status)
			assert.Equal(t, "INV-MT-9", ev.ExternalID)
			assert.Equal(t, "INV-MT-9", ev.MerchantRef)
			assert.InDelta(t, 110000, ev.PaidAmount, 0.01)
		})
	}
}

func TestParseWebhook_RejectsBadSignature(t *testing.T) {
	a := newTestAdapter("https://midtrans.invalid")
	body := []byte(`{"transaction_status":"settlement","order_id":"INV-1","status_code":"200",` +
		`"gross_amount":"110000.00","signature_key":"deadbeef"}`)
	_, err := a.ParseWebhook(context.Background(), body, "")
	assert.ErrorIs(t, err, domainBilling.ErrGatewayBadSign)
}

func TestCheckStatus_MapsSettled(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"transaction_status":"settlement","fraud_status":"accept",` +
			`"order_id":"INV-MT-2","gross_amount":"110000.00","status_code":"200"}`))
	}))
	defer srv.Close()

	a := newTestAdapter(srv.URL)
	ev, err := a.CheckStatus(context.Background(), "INV-MT-2")
	require.NoError(t, err)
	assert.Equal(t, "/v2/INV-MT-2/status", gotPath)
	assert.Equal(t, domainBilling.GatewayStatusSettled, ev.Status)
	assert.InDelta(t, 110000, ev.PaidAmount, 0.01)
}

func TestCheckStatus_DisabledGateway(t *testing.T) {
	a := &Adapter{
		reader: settingMap{"gw.midtrans.server_key": ""},
		client: http.DefaultClient,
	}
	_, err := a.CheckStatus(context.Background(), "INV-1")
	assert.ErrorIs(t, err, domainBilling.ErrGatewayDisabled)
}

func TestEnabled_RequiresFlagAndServerKey(t *testing.T) {
	on := &Adapter{reader: settingMap{"gw.midtrans.enabled": "true", "gw.midtrans.server_key": "SK"}, client: http.DefaultClient}
	assert.True(t, on.Enabled(context.Background()))

	off := &Adapter{reader: settingMap{"gw.midtrans.enabled": "false", "gw.midtrans.server_key": "SK"}, client: http.DefaultClient}
	assert.False(t, off.Enabled(context.Background()))
}
