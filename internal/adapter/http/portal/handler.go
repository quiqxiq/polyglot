// Package portal menyediakan endpoint plain-HTTP untuk portal mandiri
// pelanggan (login OTP WhatsApp, profil, tagihan, riwayat).
package portal

import (
	"encoding/json"
	"net/http"
	"strings"

	domainCustomer "github.com/quixiq/polyglot/internal/domain/customer"
	uc "github.com/quixiq/polyglot/internal/usecase/portal"
)

type Handler struct {
	usecase *uc.UseCase
}

func NewHandler(u *uc.UseCase) *Handler {
	return &Handler{usecase: u}
}

// RegisterPublic mounts login/OTP endpoints (tanpa auth).
func (h *Handler) RegisterPublic(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/portal/otp/request", h.requestOTP)
	mux.HandleFunc("POST /api/portal/login", h.login)
}

// RegisterAuthenticated mounts data-mandiri endpoints (bearer token portal).
func (h *Handler) RegisterAuthenticated(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/portal/me", h.auth(h.me))
	mux.HandleFunc("GET /api/portal/invoices", h.auth(h.invoices))
	mux.HandleFunc("GET /api/portal/payments", h.auth(h.payments))
	mux.HandleFunc("POST /api/portal/logout", h.auth(h.logout))
}

func bearer(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(strings.ToLower(auth), "bearer ") {
		return strings.TrimSpace(auth[7:])
	}
	return ""
}

func (h *Handler) auth(next func(w http.ResponseWriter, r *http.Request, cust domainCustomer.Customer)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cust, err := h.usecase.Authenticate(r.Context(), bearer(r))
		if err != nil {
			writeErr(w, http.StatusUnauthorized, "sesi tidak valid atau kedaluwarsa")
			return
		}
		next(w, r, cust)
	}
}

func (h *Handler) requestOTP(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Identifier string `json:"identifier"` // kode portal ATAU no. HP
	}
	if !decodeBody(w, r, &req) {
		return
	}
	masked, err := h.usecase.RequestOTP(r.Context(), req.Identifier)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "gagal mengirim OTP: periksa kode portal / nomor HP dan pastikan WhatsApp aktif")
		return
	}
	writeJSON(w, map[string]string{"message": "OTP dikirim via WhatsApp ke " + masked})
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Identifier string `json:"identifier"`
		OTP        string `json:"otp"`
	}
	if !decodeBody(w, r, &req) {
		return
	}
	token, cust, err := h.usecase.Login(r.Context(), req.Identifier, req.OTP)
	if err != nil {
		status := http.StatusUnauthorized
		if strings.Contains(err.Error(), "locked") {
			status = http.StatusTooManyRequests
		} else if strings.Contains(err.Error(), "expired") || strings.Contains(err.Error(), "not found") {
			status = http.StatusBadRequest
		}
		writeErr(w, status, err.Error())
		return
	}
	writeJSON(w, map[string]any{
		"token":    token,
		"customer": publicCustomer(cust),
	})
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request, cust domainCustomer.Customer) {
	ov, err := h.usecase.Overview(r.Context(), cust.ID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, ov)
}

func (h *Handler) invoices(w http.ResponseWriter, r *http.Request, cust domainCustomer.Customer) {
	invoices, err := h.usecase.Invoices(r.Context(), cust.ID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, map[string]any{"invoices": invoices})
}

func (h *Handler) payments(w http.ResponseWriter, r *http.Request, cust domainCustomer.Customer) {
	payments, err := h.usecase.Payments(r.Context(), cust.ID, 50)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, map[string]any{"payments": payments})
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request, _ domainCustomer.Customer) {
	if err := h.usecase.Logout(r.Context(), bearer(r)); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, map[string]string{"message": "logged out"})
}

func publicCustomer(c domainCustomer.Customer) map[string]any {
	return map[string]any{
		"id": c.ID, "name": c.Name, "status": c.Status,
		"customer_code": c.CustomerCode,
	}
}

func decodeBody(w http.ResponseWriter, r *http.Request, dst any) bool {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		writeErr(w, http.StatusBadRequest, "body JSON tidak valid")
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
