// Package gateway menyediakan endpoint plain-HTTP: webhook provider
// (publik) dan pembuatan tagihan online oleh kasir (terproteksi JWT).
package gateway

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	uc "github.com/quixiq/polyglot/internal/usecase/billing"
	"github.com/quixiq/polyglot/pkg/fault"
	"github.com/quixiq/polyglot/pkg/response"
)

type Handler struct {
	usecase *uc.GatewayChargeUseCase
}

func NewHandler(u *uc.GatewayChargeUseCase) *Handler {
	return &Handler{usecase: u}
}

// RegisterPublic mounts webhook callback per gateway (tanpa auth; tiap
// adapter memverifikasi signature/token-nya sendiri). Pembuatan tagihan
// portal pindah ke endpoint bersesi pelanggan (/api/portal/charge).
func (h *Handler) RegisterPublic(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/webhook/tripay", h.webhook("TRIPAY"))
	mux.HandleFunc("POST /api/webhook/midtrans", h.webhook("MIDTRANS"))
	mux.HandleFunc("POST /api/webhook/xendit", h.webhook("XENDIT"))
}

// RegisterProtected mounts kasir endpoints (di balik middleware JWT).
func (h *Handler) RegisterProtected(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/cashier/charge", h.charge)
	mux.HandleFunc("POST /api/cashier/check-charge", h.checkCharge)
}

// webhookAuthHeader mengambil kredensial callback sesuai gateway:
// Xendit memakai x-callback-token, Tripay memakai X-Callback-Signature, dan
// Midtrans menandatangani body (signature_key) sehingga header tidak dipakai.
func webhookAuthHeader(gatewayName string, r *http.Request) string {
	switch strings.ToUpper(gatewayName) {
	case "XENDIT":
		return r.Header.Get("X-Callback-Token")
	case "MIDTRANS":
		return ""
	default:
		return r.Header.Get("X-Callback-Signature")
	}
}

func (h *Handler) webhook(gatewayName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
		if err != nil {
			respond(w, http.StatusBadRequest, false, "body tidak terbaca")
			return
		}
		invoiceID, settled, err := h.usecase.HandleWebhookFor(r.Context(), gatewayName, body, webhookAuthHeader(gatewayName, r))
		if err != nil {
			switch {
			case errors.Is(err, domainBilling.ErrGatewayBadSign):
				response.WriteHTTPStatusError(w, http.StatusUnauthorized, "callback signature is invalid")
			case errors.Is(err, domainBilling.ErrGatewayUnknownRef):
				response.WriteHTTPStatusError(w, http.StatusNotFound, "transaction not found")
			case errors.Is(err, domainBilling.ErrGatewayNotFound):
				response.WriteHTTPStatusError(w, http.StatusNotFound, "unknown payment gateway")
			default:
				writeMappedError(w, err)
			}
			return
		}
		payload := map[string]any{"success": true}
		if settled && invoiceID != "" {
			payload["invoice_id"] = invoiceID
		}
		writeJSON(w, http.StatusOK, payload)
	}
}

func (h *Handler) charge(w http.ResponseWriter, r *http.Request) {
	var req struct {
		InvoiceID     string `json:"invoice_id"`
		Gateway       string `json:"gateway,omitempty"`
		Channel       string `json:"channel,omitempty"`
		ExpireMinutes int    `json:"expire_minutes,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond(w, http.StatusBadRequest, false, "body JSON tidak valid")
		return
	}
	res, tx, err := h.usecase.CreateForInvoiceVia(r.Context(), req.InvoiceID, req.Gateway, req.Channel, req.ExpireMinutes)
	if err != nil {
		writeMappedError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"gateway":     tx.Gateway,
		"external_id": res.ExternalID,
		"payment_url": res.PaymentURL,
		"qr_string":   res.QRString,
		"va_number":   res.VANumber,
		"status":      tx.Status,
		"amount":      tx.Amount,
	})
}

func (h *Handler) checkCharge(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ExternalID string `json:"external_id"`
		Gateway    string `json:"gateway,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond(w, http.StatusBadRequest, false, "body JSON tidak valid")
		return
	}
	if req.ExternalID == "" {
		respond(w, http.StatusBadRequest, false, "external_id wajib diisi")
		return
	}
	invoiceID, settled, status, err := h.usecase.CheckPaymentStatusVia(r.Context(), req.Gateway, req.ExternalID)
	if err != nil {
		writeMappedError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"external_id": req.ExternalID,
		"status":      status,
		"settled":     settled,
		"invoice_id":  invoiceID,
	})
}

func respond(w http.ResponseWriter, code int, success bool, errMsg string) {
	if !success {
		message := errMsg
		if code >= http.StatusInternalServerError {
			message = "internal server error"
		}
		response.WriteHTTPStatusError(w, code, message)
		return
	}
	writeJSON(w, code, map[string]any{"success": true})
}

func writeMappedError(w http.ResponseWriter, err error) {
	if fault.KindOf(err) != fault.KindUnknown {
		response.WriteHTTPError(w, err)
		return
	}
	response.WriteHTTPStatusError(w, http.StatusInternalServerError, "internal server error")
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
