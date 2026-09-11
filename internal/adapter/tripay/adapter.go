// Package tripay mengimplementasikan port.PaymentGateway untuk Tripay
// (QRIS/VA) — kredensial dibaca dari system_settings, HTTP client dapat
// disuntik untuk testing.
package tripay

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	"github.com/quixiq/polyglot/internal/port"
)

// Name is the stable identifier for the Tripay payment gateway.
const Name = "TRIPAY"

// Config contains the Tripay gateway configuration.
type Config struct {
	Endpoint       string // https://tripay.co.id/api | api-sandbox
	MerchantCode   string
	APIKey         string
	PrivateKey     string
	Channel        string
	CallbackAction func(ctx context.Context, event port.WebhookEvent) error
}

// ReadConfig reads Tripay configuration values from the settings repository.
func ReadConfig(ctx context.Context, reader port.SettingReader) Config {
	endpoint := reader.GetValue(ctx, "gw.tripay.endpoint", "https://tripay.co.id/api")
	return Config{
		Endpoint:     strings.TrimSuffix(endpoint, "/"),
		MerchantCode: reader.GetValue(ctx, "gw.tripay.merchant_code", ""),
		APIKey:       reader.GetValue(ctx, "gw.tripay.api_key", ""),
		PrivateKey:   reader.GetValue(ctx, "gw.tripay.private_key", ""),
		Channel:      reader.GetValue(ctx, "gw.tripay.channel", "QRIS"),
	}
}

type Adapter struct {
	reader port.SettingReader
	client *http.Client
}

var _ port.PaymentGateway = (*Adapter)(nil)

func NewAdapter(reader port.SettingReader) *Adapter {
	return &Adapter{reader: reader, client: &http.Client{Timeout: 20 * time.Second}}
}

// NewAdapterWithClient untuk testing (httptest server + custom config).
func NewAdapterWithClient(cfg Config, client *http.Client) *Adapter {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &Adapter{reader: staticReader(cfg), client: client}
}

func staticReader(cfg Config) port.SettingReader {
	return settingMap{
		"gw.tripay.endpoint": cfg.Endpoint, "gw.tripay.merchant_code": cfg.MerchantCode,
		"gw.tripay.api_key": cfg.APIKey, "gw.tripay.private_key": cfg.PrivateKey,
		"gw.tripay.channel": cfg.Channel,
	}
}

type settingMap map[string]string

func (m settingMap) GetValue(_ context.Context, key, fallback string) string {
	if v, ok := m[key]; ok && v != "" {
		return v
	}
	return fallback
}

func (a *Adapter) cfg(ctx context.Context) Config { return ReadConfig(ctx, a.reader) }

func (a *Adapter) Name() string { return Name }

func (a *Adapter) Enabled(ctx context.Context) bool {
	return strings.EqualFold(a.reader.GetValue(ctx, "gw.tripay.enabled", "false"), "true") &&
		a.cfg(ctx).PrivateKey != ""
}

// ─── CreateCharge ───────────────────────────────────────────────────────

type createResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Reference   string  `json:"reference"`
		MerchantRef string  `json:"merchant_ref"`
		PaymentURL  string  `json:"checkout_url"`
		QRString    string  `json:"qr_string"`
		PayCode     string  `json:"pay_code"`
		FeeMerchant float64 `json:"fee_merchant"`
		Status      string  `json:"status"`
	} `json:"data"`
	Message string `json:"message"`
}

func (a *Adapter) CreateCharge(ctx context.Context, req port.ChargeRequest) (port.ChargeResult, error) {
	cfg := a.cfg(ctx)
	channel := req.Channel
	if channel == "" {
		channel = cfg.Channel
	}
	expire := req.ExpireMinutes
	if expire <= 0 {
		expire = 60
	}
	merchantRef := req.InvoiceNumber
	signature := signHMAC(cfg.PrivateKey, cfg.MerchantCode+merchantRef+fmt.Sprintf("%.0f", req.Amount))
	expiresAt := time.Now().Add(time.Duration(expire) * time.Minute)

	body := map[string]any{
		"method":         channel,
		"merchant_code":  cfg.MerchantCode,
		"merchant_ref":   merchantRef,
		"amount":         req.Amount,
		"customer_name":  req.CustomerName,
		"customer_phone": req.CustomerPhone,
		"customer_email": orEmail(req.CustomerEmail),
		"signature":      signature,
		"expired_time":   expiresAt.Unix(),
	}
	payload, _ := json.Marshal(body)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		cfg.Endpoint+"/transaction/create", bytes.NewReader(payload))
	if err != nil {
		return port.ChargeResult{}, fmt.Errorf("build request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return port.ChargeResult{}, fmt.Errorf("http tripay: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return port.ChargeResult{}, fmt.Errorf("read tripay response: %w", err)
	}

	var out createResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return port.ChargeResult{}, fmt.Errorf("parse response: %w", err)
	}
	if !out.Success || resp.StatusCode >= 400 {
		return port.ChargeResult{}, fmt.Errorf("tripay reject (http %d): %s", resp.StatusCode, out.Message)
	}

	result := port.ChargeResult{
		ExternalID:  out.Data.Reference,
		MerchantRef: merchantRef,
		PaymentURL:  out.Data.PaymentURL,
		QRString:    out.Data.QRString,
		VANumber:    out.Data.PayCode,
		Channel:     channel,
		FeeAmount:   out.Data.FeeMerchant,
		Status:      domainBilling.GatewayStatusPending,
		ExpiresAt:   expiresAt,
		RawResponse: raw,
	}
	return result, nil
}

// ─── Webhook ────────────────────────────────────────────────────────────

type callbackPayload struct {
	Reference   string  `json:"reference"`
	MerchantRef string  `json:"merchant_ref"`
	Status      string  `json:"status"`
	TotalAmount float64 `json:"total_amount"`
	FeeAmount   float64 `json:"fee_merchant"`
	PaidAmount  float64 `json:"amount_received"`
	QRString    string  `json:"qr_string"`
	PayCode     string  `json:"pay_code"`
}

// ParseWebhook implements port.PaymentGateway: validasi X-Callback-Signature
// = HMAC_SHA256(private_key, raw JSON body) sesuai dokumentasi resmi Tripay.
func (a *Adapter) ParseWebhook(ctx context.Context, body []byte, signatureHeader string) (port.WebhookEvent, error) {
	cfg := a.cfg(ctx)
	var p callbackPayload
	if err := json.Unmarshal(body, &p); err != nil {
		return port.WebhookEvent{}, fmt.Errorf("parse callback: %w", err)
	}
	expect := signHMAC(cfg.PrivateKey, string(body))
	if !hmac.Equal([]byte(expect), []byte(strings.ToLower(signatureHeader))) {
		return port.WebhookEvent{}, domainBilling.ErrGatewayBadSign
	}
	status := domainBilling.GatewayStatusPending
	switch strings.ToUpper(p.Status) {
	case "PAID", "SETTLED":
		status = domainBilling.GatewayStatusSettled
	case "EXPIRED":
		status = domainBilling.GatewayStatusExpired
	case "FAILED":
		status = domainBilling.GatewayStatusFailed
	}
	// Pelunasan memakai total_amount (kewajiban penuh), bukan amount_received
	// yang sudah dipotong fee merchant — fee adalah biaya provider, bukan
	// diskon tagihan pelanggan.
	paid := p.TotalAmount
	if paid == 0 {
		paid = p.PaidAmount
	}
	return port.WebhookEvent{
		ExternalID:  p.Reference,
		MerchantRef: p.MerchantRef,
		Status:      status,
		PaidAmount:  paid,
		Raw:         body,
	}, nil
}

// CheckStatus queries transaction detail directly from Tripay API.
func (a *Adapter) CheckStatus(ctx context.Context, externalID string) (port.WebhookEvent, error) {
	cfg := a.cfg(ctx)
	if !a.Enabled(ctx) || cfg.APIKey == "" {
		return port.WebhookEvent{}, domainBilling.ErrGatewayDisabled
	}
	reqURL := fmt.Sprintf("%s/transaction/detail?reference=%s", cfg.Endpoint, url.QueryEscape(externalID))
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return port.WebhookEvent{}, fmt.Errorf("build request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	resp, err := a.client.Do(httpReq)
	if err != nil {
		return port.WebhookEvent{}, fmt.Errorf("http tripay: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return port.WebhookEvent{}, fmt.Errorf("read tripay: %w", err)
	}
	var out struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
		Data    struct {
			Reference   string  `json:"reference"`
			MerchantRef string  `json:"merchant_ref"`
			Status      string  `json:"status"`
			Amount      float64 `json:"amount"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return port.WebhookEvent{}, fmt.Errorf("parse response: %w", err)
	}
	if !out.Success || resp.StatusCode >= 400 {
		return port.WebhookEvent{}, fmt.Errorf("tripay reject (http %d): %s", resp.StatusCode, out.Message)
	}
	status := domainBilling.GatewayStatusPending
	switch strings.ToUpper(out.Data.Status) {
	case "PAID", "SETTLED":
		status = domainBilling.GatewayStatusSettled
	case "EXPIRED":
		status = domainBilling.GatewayStatusExpired
	case "FAILED":
		status = domainBilling.GatewayStatusFailed
	}
	return port.WebhookEvent{
		ExternalID:  out.Data.Reference,
		MerchantRef: out.Data.MerchantRef,
		Status:      status,
		PaidAmount:  out.Data.Amount,
		Raw:         raw,
	}, nil
}

func signHMAC(key, data string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil))
}

func orEmail(e string) string {
	if e == "" {
		return "pelanggan@polyglot.local"
	}
	return e
}
