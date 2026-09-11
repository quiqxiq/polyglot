// Package xendit mengimplementasikan port.PaymentGateway untuk Xendit:
// Invoice API untuk create + check status, dan callback dengan verifikasi
// header x-callback-token (F4-4).
package xendit

import (
	"bytes"
	"context"
	"crypto/subtle"
	"encoding/base64"
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

// Name is the stable identifier for the Xendit payment gateway.
const Name = "XENDIT"

// Config contains the Xendit gateway configuration.
type Config struct {
	Endpoint      string // https://api.xendit.co
	SecretKey     string
	CallbackToken string
	Channel       string
}

// ReadConfig reads Xendit configuration values from the settings repository.
func ReadConfig(ctx context.Context, reader port.SettingReader) Config {
	endpoint := reader.GetValue(ctx, "gw.xendit.endpoint", "https://api.xendit.co")
	return Config{
		Endpoint:      strings.TrimSuffix(endpoint, "/"),
		SecretKey:     reader.GetValue(ctx, "gw.xendit.secret_key", ""),
		CallbackToken: reader.GetValue(ctx, "gw.xendit.callback_token", ""),
		Channel:       reader.GetValue(ctx, "gw.xendit.channel", ""),
	}
}

// Adapter implements port.PaymentGateway for Xendit.
type Adapter struct {
	reader port.SettingReader
	client *http.Client
}

var _ port.PaymentGateway = (*Adapter)(nil)

// NewAdapter constructs an adapter reading settings via reader.
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
		"gw.xendit.endpoint":       cfg.Endpoint,
		"gw.xendit.secret_key":     cfg.SecretKey,
		"gw.xendit.callback_token": cfg.CallbackToken,
		"gw.xendit.channel":        cfg.Channel,
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

// Name returns the gateway identifier.
func (a *Adapter) Name() string { return Name }

// Enabled reports whether Xendit aktif dan secret key terisi.
func (a *Adapter) Enabled(ctx context.Context) bool {
	return strings.EqualFold(a.reader.GetValue(ctx, "gw.xendit.enabled", "false"), "true") &&
		a.cfg(ctx).SecretKey != ""
}

func basicAuth(secretKey string) string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(secretKey+":"))
}

// ─── CreateCharge (Invoice API) ─────────────────────────────────────────

type invoiceResponse struct {
	ID         string  `json:"id"`
	ExternalID string  `json:"external_id"`
	InvoiceURL string  `json:"invoice_url"`
	Amount     float64 `json:"amount"`
	Status     string  `json:"status"`
	ExpiryDate string  `json:"expiry_date"`
	ErrorCode  string  `json:"error_code"`
	Message    string  `json:"message"`
}

// CreateCharge membuat invoice; ExternalID = id invoice Xendit (stabil di
// create response, callback, dan check-status).
func (a *Adapter) CreateCharge(ctx context.Context, req port.ChargeRequest) (port.ChargeResult, error) {
	cfg := a.cfg(ctx)
	if cfg.SecretKey == "" {
		return port.ChargeResult{}, domainBilling.ErrGatewayDisabled
	}
	expire := req.ExpireMinutes
	if expire <= 0 {
		expire = 60
	}
	channel := req.Channel
	if channel == "" {
		channel = cfg.Channel
	}

	body := map[string]any{
		"external_id":      req.InvoiceNumber,
		"amount":           req.Amount,
		"payer_email":      orEmail(req.CustomerEmail),
		"description":      "Tagihan " + req.InvoiceNumber,
		"invoice_duration": expire * 60,
		"customer": map[string]any{
			"given_names":   req.CustomerName,
			"mobile_number": req.CustomerPhone,
		},
	}
	if channel != "" {
		body["payment_methods"] = []string{strings.ToUpper(channel)}
	}
	payload, _ := json.Marshal(body)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		cfg.Endpoint+"/v2/invoices", bytes.NewReader(payload))
	if err != nil {
		return port.ChargeResult{}, fmt.Errorf("build request: %w", err)
	}
	httpReq.Header.Set("Authorization", basicAuth(cfg.SecretKey))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return port.ChargeResult{}, fmt.Errorf("http xendit: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return port.ChargeResult{}, fmt.Errorf("read xendit response: %w", err)
	}

	var out invoiceResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return port.ChargeResult{}, fmt.Errorf("parse response: %w", err)
	}
	if resp.StatusCode >= 400 || out.InvoiceURL == "" {
		return port.ChargeResult{}, fmt.Errorf("xendit reject (http %d): %s %s",
			resp.StatusCode, out.ErrorCode, out.Message)
	}

	res := port.ChargeResult{
		ExternalID:  out.ID,
		MerchantRef: out.ExternalID,
		PaymentURL:  out.InvoiceURL,
		Channel:     channel,
		Status:      domainBilling.GatewayStatusPending,
		RawResponse: raw,
	}
	if ts, perr := time.Parse(time.RFC3339, out.ExpiryDate); perr == nil {
		res.ExpiresAt = ts
	} else {
		res.ExpiresAt = time.Now().Add(time.Duration(expire) * time.Minute)
	}
	return res, nil
}

// ─── Webhook ────────────────────────────────────────────────────────────

type callbackPayload struct {
	ID         string  `json:"id"`
	ExternalID string  `json:"external_id"`
	Status     string  `json:"status"`
	Amount     float64 `json:"amount"`
	PaidAmount float64 `json:"paid_amount"`
}

// ParseWebhook memverifikasi header x-callback-token terhadap setting
// callback_token, lalu memetakan status invoice Xendit.
func (a *Adapter) ParseWebhook(ctx context.Context, body []byte, signatureHeader string) (port.WebhookEvent, error) {
	cfg := a.cfg(ctx)
	if cfg.CallbackToken == "" ||
		subtle.ConstantTimeCompare([]byte(cfg.CallbackToken), []byte(signatureHeader)) != 1 {
		return port.WebhookEvent{}, domainBilling.ErrGatewayBadSign
	}
	var p callbackPayload
	if err := json.Unmarshal(body, &p); err != nil {
		return port.WebhookEvent{}, fmt.Errorf("parse callback: %w", err)
	}
	paid := p.PaidAmount
	if paid == 0 {
		paid = p.Amount
	}
	return port.WebhookEvent{
		ExternalID:  p.ID,
		MerchantRef: p.ExternalID,
		Status:      mapStatus(p.Status),
		PaidAmount:  paid,
		Raw:         body,
	}, nil
}

// CheckStatus queries invoice detail by Xendit invoice id.
func (a *Adapter) CheckStatus(ctx context.Context, externalID string) (port.WebhookEvent, error) {
	cfg := a.cfg(ctx)
	if !a.Enabled(ctx) || cfg.SecretKey == "" {
		return port.WebhookEvent{}, domainBilling.ErrGatewayDisabled
	}
	reqURL := fmt.Sprintf("%s/v2/invoices/%s", cfg.Endpoint, url.PathEscape(externalID))
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return port.WebhookEvent{}, fmt.Errorf("build request: %w", err)
	}
	httpReq.Header.Set("Authorization", basicAuth(cfg.SecretKey))

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return port.WebhookEvent{}, fmt.Errorf("http xendit: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return port.WebhookEvent{}, fmt.Errorf("read xendit: %w", err)
	}

	var out invoiceResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return port.WebhookEvent{}, fmt.Errorf("parse response: %w", err)
	}
	if resp.StatusCode >= 400 {
		return port.WebhookEvent{}, fmt.Errorf("xendit reject (http %d): %s %s", resp.StatusCode, out.ErrorCode, out.Message)
	}
	return port.WebhookEvent{
		ExternalID:  out.ID,
		MerchantRef: out.ExternalID,
		Status:      mapStatus(out.Status),
		PaidAmount:  out.Amount,
		Raw:         raw,
	}, nil
}

// mapStatus memetakan status invoice Xendit ke status gateway internal.
func mapStatus(status string) string {
	switch strings.ToUpper(status) {
	case "PAID", "SETTLED":
		return domainBilling.GatewayStatusSettled
	case "EXPIRED":
		return domainBilling.GatewayStatusExpired
	case "FAILED":
		return domainBilling.GatewayStatusFailed
	default:
		return domainBilling.GatewayStatusPending
	}
}

func orEmail(e string) string {
	if e == "" {
		return "pelanggan@polyglot.local"
	}
	return e
}
