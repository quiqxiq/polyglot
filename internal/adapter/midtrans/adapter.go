// Package midtrans mengimplementasikan port.PaymentGateway untuk Midtrans:
// Snap API untuk create transaction, Core API untuk check status, dan
// notification callback dengan signature SHA-512 (F4-3).
package midtrans

import (
	"bytes"
	"context"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	"github.com/quixiq/polyglot/internal/port"
)

// Name is the stable identifier for the Midtrans payment gateway.
const Name = "MIDTRANS"

// Config contains the Midtrans gateway configuration.
type Config struct {
	Endpoint  string // https://api.sandbox.midtrans.com | https://api.midtrans.com
	ServerKey string
	ClientKey string
	Channel   string
}

// ReadConfig reads Midtrans configuration values from the settings repository.
func ReadConfig(ctx context.Context, reader port.SettingReader) Config {
	endpoint := reader.GetValue(ctx, "gw.midtrans.endpoint", "https://api.sandbox.midtrans.com")
	return Config{
		Endpoint:  strings.TrimSuffix(endpoint, "/"),
		ServerKey: reader.GetValue(ctx, "gw.midtrans.server_key", ""),
		ClientKey: reader.GetValue(ctx, "gw.midtrans.client_key", ""),
		Channel:   reader.GetValue(ctx, "gw.midtrans.channel", ""),
	}
}

// Adapter implements port.PaymentGateway for Midtrans.
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
		"gw.midtrans.endpoint":   cfg.Endpoint,
		"gw.midtrans.server_key": cfg.ServerKey,
		"gw.midtrans.client_key": cfg.ClientKey,
		"gw.midtrans.channel":    cfg.Channel,
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

// Enabled reports whether Midtrans is aktif dan server key terisi.
func (a *Adapter) Enabled(ctx context.Context) bool {
	return strings.EqualFold(a.reader.GetValue(ctx, "gw.midtrans.enabled", "false"), "true") &&
		a.cfg(ctx).ServerKey != ""
}

func basicAuth(serverKey string) string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(serverKey+":"))
}

// ─── CreateCharge (Snap) ────────────────────────────────────────────────

type snapResponse struct {
	Token         string   `json:"token"`
	RedirectURL   string   `json:"redirect_url"`
	ErrorMessages []string `json:"error_messages"`
}

// CreateCharge membuat Snap transaction; ExternalID = order_id (nomor
// invoice) karena Snap tidak mengembalikan transaction_id dan notification
// selalu membawa order_id.
func (a *Adapter) CreateCharge(ctx context.Context, req port.ChargeRequest) (port.ChargeResult, error) {
	cfg := a.cfg(ctx)
	if cfg.ServerKey == "" {
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
	orderID := req.InvoiceNumber

	body := map[string]any{
		"transaction_details": map[string]any{
			"order_id":     orderID,
			"gross_amount": req.Amount,
		},
		"customer_details": map[string]any{
			"first_name": req.CustomerName,
			"phone":      req.CustomerPhone,
			"email":      orEmail(req.CustomerEmail),
		},
	}
	if channel != "" {
		body["enabled_payments"] = []string{strings.ToLower(channel)}
	}
	payload, _ := json.Marshal(body)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		cfg.Endpoint+"/snap/v1/transactions", bytes.NewReader(payload))
	if err != nil {
		return port.ChargeResult{}, fmt.Errorf("build request: %w", err)
	}
	httpReq.Header.Set("Authorization", basicAuth(cfg.ServerKey))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return port.ChargeResult{}, fmt.Errorf("http midtrans: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return port.ChargeResult{}, fmt.Errorf("read midtrans response: %w", err)
	}

	var out snapResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return port.ChargeResult{}, fmt.Errorf("parse response: %w", err)
	}
	if resp.StatusCode >= 400 || out.RedirectURL == "" {
		return port.ChargeResult{}, fmt.Errorf("midtrans reject (http %d): %s",
			resp.StatusCode, strings.Join(out.ErrorMessages, "; "))
	}

	return port.ChargeResult{
		ExternalID:  orderID,
		MerchantRef: orderID,
		PaymentURL:  out.RedirectURL,
		Channel:     channel,
		Status:      domainBilling.GatewayStatusPending,
		ExpiresAt:   time.Now().Add(time.Duration(expire) * time.Minute),
		RawResponse: raw,
	}, nil
}

// ─── Webhook ────────────────────────────────────────────────────────────

type notificationPayload struct {
	TransactionStatus string `json:"transaction_status"`
	FraudStatus       string `json:"fraud_status"`
	OrderID           string `json:"order_id"`
	TransactionID     string `json:"transaction_id"`
	GrossAmount       string `json:"gross_amount"`
	StatusCode        string `json:"status_code"`
	SignatureKey      string `json:"signature_key"`
	PaymentType       string `json:"payment_type"`
}

// ParseWebhook memvalidasi signature_key = SHA512(order_id + status_code +
// gross_amount + server_key) lalu memetakan status Midtrans.
func (a *Adapter) ParseWebhook(ctx context.Context, body []byte, _ string) (port.WebhookEvent, error) {
	cfg := a.cfg(ctx)
	var p notificationPayload
	if err := json.Unmarshal(body, &p); err != nil {
		return port.WebhookEvent{}, fmt.Errorf("parse callback: %w", err)
	}
	if cfg.ServerKey == "" {
		return port.WebhookEvent{}, domainBilling.ErrGatewayBadSign
	}
	expect := sha512Hex(p.OrderID + p.StatusCode + p.GrossAmount + cfg.ServerKey)
	if !strings.EqualFold(expect, p.SignatureKey) {
		return port.WebhookEvent{}, domainBilling.ErrGatewayBadSign
	}
	paid, _ := strconv.ParseFloat(p.GrossAmount, 64)
	return port.WebhookEvent{
		ExternalID:  p.OrderID,
		MerchantRef: p.OrderID,
		Status:      mapStatus(p.TransactionStatus, p.FraudStatus),
		PaidAmount:  paid,
		Raw:         body,
	}, nil
}

// CheckStatus queries Core API transaction status by order_id.
func (a *Adapter) CheckStatus(ctx context.Context, externalID string) (port.WebhookEvent, error) {
	cfg := a.cfg(ctx)
	if !a.Enabled(ctx) || cfg.ServerKey == "" {
		return port.WebhookEvent{}, domainBilling.ErrGatewayDisabled
	}
	reqURL := fmt.Sprintf("%s/v2/%s/status", cfg.Endpoint, url.PathEscape(externalID))
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return port.WebhookEvent{}, fmt.Errorf("build request: %w", err)
	}
	httpReq.Header.Set("Authorization", basicAuth(cfg.ServerKey))
	httpReq.Header.Set("Accept", "application/json")

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return port.WebhookEvent{}, fmt.Errorf("http midtrans: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return port.WebhookEvent{}, fmt.Errorf("read midtrans: %w", err)
	}

	var out struct {
		TransactionStatus string `json:"transaction_status"`
		FraudStatus       string `json:"fraud_status"`
		OrderID           string `json:"order_id"`
		GrossAmount       string `json:"gross_amount"`
		StatusCode        string `json:"status_code"`
		StatusMessage     string `json:"status_message"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return port.WebhookEvent{}, fmt.Errorf("parse response: %w", err)
	}
	if resp.StatusCode >= 400 {
		return port.WebhookEvent{}, fmt.Errorf("midtrans reject (http %d): %s", resp.StatusCode, out.StatusMessage)
	}
	paid, _ := strconv.ParseFloat(out.GrossAmount, 64)
	orderID := out.OrderID
	if orderID == "" {
		orderID = externalID
	}
	return port.WebhookEvent{
		ExternalID:  orderID,
		MerchantRef: orderID,
		Status:      mapStatus(out.TransactionStatus, out.FraudStatus),
		PaidAmount:  paid,
		Raw:         raw,
	}, nil
}

// mapStatus memetakan transaction_status Midtrans ke status gateway internal.
func mapStatus(txStatus, fraudStatus string) string {
	switch strings.ToLower(txStatus) {
	case "capture":
		if strings.EqualFold(fraudStatus, "challenge") {
			return domainBilling.GatewayStatusPending
		}
		return domainBilling.GatewayStatusSettled
	case "settlement":
		return domainBilling.GatewayStatusSettled
	case "pending":
		return domainBilling.GatewayStatusPending
	case "expire":
		return domainBilling.GatewayStatusExpired
	case "deny", "cancel", "failure", "refund", "partial_refund":
		return domainBilling.GatewayStatusFailed
	default:
		return domainBilling.GatewayStatusPending
	}
}

func sha512Hex(s string) string {
	sum := sha512.Sum512([]byte(s))
	return hex.EncodeToString(sum[:])
}

func orEmail(e string) string {
	if e == "" {
		return "pelanggan@polyglot.local"
	}
	return e
}
