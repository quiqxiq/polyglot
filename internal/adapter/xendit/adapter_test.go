// Xendit adapter tests (F4-4): Invoice API create, callback token, dan
// check-status.
package xendit

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

const testSecretKey = "xnd_development_TEST"

func newTestAdapter(endpoint string) *Adapter {
	return &Adapter{
		reader: settingMap{
			"gw.xendit.enabled":        "true",
			"gw.xendit.endpoint":       endpoint,
			"gw.xendit.secret_key":     testSecretKey,
			"gw.xendit.callback_token": "CB-TOKEN-1",
			"gw.xendit.channel":        "",
		},
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

func TestCreateCharge_SendsBasicAuthAndMapsInvoice(t *testing.T) {
	var captured map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v2/invoices", r.URL.Path)
		wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(testSecretKey+":"))
		assert.Equal(t, wantAuth, r.Header.Get("Authorization"))
		require.NoError(t, json.NewDecoder(r.Body).Decode(&captured))
		_, _ = w.Write([]byte(`{"id":"xnd-inv-1","external_id":"INV-XD-1",` +
			`"invoice_url":"https://checkout.xendit.co/web/xnd-inv-1",` +
			`"amount":110000,"status":"PENDING","expiry_date":"2026-09-12T10:00:00.000Z"}`))
	}))
	defer srv.Close()

	a := newTestAdapter(srv.URL)
	res, err := a.CreateCharge(context.Background(), port.ChargeRequest{
		InvoiceNumber: "INV-XD-1", Amount: 110000,
		CustomerName: "Budi", CustomerPhone: "081200000000", ExpireMinutes: 60,
	})
	require.NoError(t, err)

	assert.Equal(t, "INV-XD-1", captured["external_id"])
	assert.InDelta(t, 110000, captured["amount"], 0.01)
	assert.InDelta(t, 3600, captured["invoice_duration"], 1)

	assert.Equal(t, "xnd-inv-1", res.ExternalID)
	assert.Equal(t, "INV-XD-1", res.MerchantRef)
	assert.Equal(t, "https://checkout.xendit.co/web/xnd-inv-1", res.PaymentURL)
	assert.Equal(t, domainBilling.GatewayStatusPending, res.Status)
	assert.Equal(t, 2026, res.ExpiresAt.Year())
}

func TestCreateCharge_RejectedByProvider(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error_code":"API_VALIDATION_ERROR","message":"amount is required"}`))
	}))
	defer srv.Close()

	a := newTestAdapter(srv.URL)
	_, err := a.CreateCharge(context.Background(), port.ChargeRequest{InvoiceNumber: "INV-1", Amount: 1000})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "xendit reject")
}

func TestParseWebhook_ValidTokenAndStatusMapping(t *testing.T) {
	a := newTestAdapter("https://xendit.invalid")
	tests := []struct {
		status string
		want   string
	}{
		{"PAID", domainBilling.GatewayStatusSettled},
		{"SETTLED", domainBilling.GatewayStatusSettled},
		{"PENDING", domainBilling.GatewayStatusPending},
		{"EXPIRED", domainBilling.GatewayStatusExpired},
	}
	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			body := []byte(`{"id":"xnd-1","external_id":"INV-XD-9","status":"` + tt.status +
				`","amount":110000,"paid_amount":110000}`)
			ev, err := a.ParseWebhook(context.Background(), body, "CB-TOKEN-1")
			require.NoError(t, err)
			assert.Equal(t, tt.want, ev.Status)
			assert.Equal(t, "xnd-1", ev.ExternalID)
			assert.Equal(t, "INV-XD-9", ev.MerchantRef)
			assert.InDelta(t, 110000, ev.PaidAmount, 0.01)
		})
	}
}

func TestParseWebhook_RejectsWrongToken(t *testing.T) {
	a := newTestAdapter("https://xendit.invalid")
	body := []byte(`{"id":"xnd-1","external_id":"INV-1","status":"PAID","amount":1000}`)
	_, err := a.ParseWebhook(context.Background(), body, "WRONG")
	assert.ErrorIs(t, err, domainBilling.ErrGatewayBadSign)

	_, err = a.ParseWebhook(context.Background(), body, "")
	assert.ErrorIs(t, err, domainBilling.ErrGatewayBadSign)
}

func TestCheckStatus_MapsSettled(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"id":"xnd-2","external_id":"INV-XD-2","status":"PAID","amount":110000}`))
	}))
	defer srv.Close()

	a := newTestAdapter(srv.URL)
	ev, err := a.CheckStatus(context.Background(), "xnd-2")
	require.NoError(t, err)
	assert.Equal(t, "/v2/invoices/xnd-2", gotPath)
	assert.Equal(t, domainBilling.GatewayStatusSettled, ev.Status)
	assert.InDelta(t, 110000, ev.PaidAmount, 0.01)
}

func TestCheckStatus_DisabledGateway(t *testing.T) {
	a := &Adapter{
		reader: settingMap{"gw.xendit.secret_key": ""},
		client: http.DefaultClient,
	}
	_, err := a.CheckStatus(context.Background(), "xnd-1")
	assert.ErrorIs(t, err, domainBilling.ErrGatewayDisabled)
}

func TestEnabled_RequiresFlagAndSecretKey(t *testing.T) {
	on := &Adapter{reader: settingMap{"gw.xendit.enabled": "true", "gw.xendit.secret_key": "SK"}, client: http.DefaultClient}
	assert.True(t, on.Enabled(context.Background()))

	off := &Adapter{reader: settingMap{"gw.xendit.enabled": "false", "gw.xendit.secret_key": "SK"}, client: http.DefaultClient}
	assert.False(t, off.Enabled(context.Background()))
}
