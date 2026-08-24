// Package gateway menyediakan endpoint plain-HTTP: webhook provider
// (publik) dan pembuatan tagihan online oleh kasir (terproteksi JWT).
package gateway

import (
	"encoding/json"
	"io"
	"net/http"

	uc "github.com/quixiq/polyglot/internal/usecase/billing"
)

type Handler struct {
	usecase *uc.GatewayChargeUseCase
}

func NewHandler(u *uc.GatewayChargeUseCase) *Handler {
	return &Handler{usecase: u}
}

// RegisterPublic mounts webhook callback (tanpa auth; diverifikasi signature).
func (h *Handler) RegisterPublic(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/webhook/tripay", h.tripayWebhook)
}

// RegisterProtected mounts kasir endpoints (di balik middleware JWT).
func (h *Handler) RegisterProtected(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/cashier/charge", h.charge)
}

func (h *Handler) tripayWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		respond(w, http.StatusBadRequest, false, "body tidak terbaca")
		return
	}
	invoiceID, settled, err := h.usecase.HandleWebhook(r.Context(), body, r.Header.Get("X-Callback-Signature"))
	if err != nil {
		code := http.StatusUnauthorized
		if invoiceID != "" || settled {
			code = http.StatusInternalServerError
		}
		respond(w, code, false, err.Error())
		return
	}
	payload := map[string]any{"success": true}
	if settled && invoiceID != "" {
		payload["invoice_id"] = invoiceID
	}
	writeJSON(w, http.StatusOK, payload)
}

func (h *Handler) charge(w http.ResponseWriter, r *http.Request) {
	var req struct {
		InvoiceID     string `json:"invoice_id"`
		Channel       string `json:"channel,omitempty"`
		ExpireMinutes int    `json:"expire_minutes,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond(w, http.StatusBadRequest, false, "body JSON tidak valid")
		return
	}
	res, tx, err := h.usecase.CreateForInvoice(r.Context(), req.InvoiceID, req.Channel, req.ExpireMinutes)
	if err != nil {
		respond(w, http.StatusBadRequest, false, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"external_id": res.ExternalID,
		"payment_url": res.PaymentURL,
		"qr_string":   res.QRString,
		"va_number":   res.VANumber,
		"status":      tx.Status,
		"amount":      tx.Amount,
	})
}

func respond(w http.ResponseWriter, code int, success bool, errMsg string) {
	payload := map[string]any{"success": success}
	if errMsg != "" {
		payload["message"] = errMsg
	}
	writeJSON(w, code, payload)
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
