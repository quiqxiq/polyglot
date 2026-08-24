package reports

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/quixiq/polyglot/internal/domain/reporting"
	"github.com/quixiq/polyglot/internal/port"
)

// Handler menyajikan laporan finansial harian/bulanan/tahunan dari
// daily_financial_snapshots (plain HTTP, JSON).
type Handler struct {
	repo port.ReportingRepository
}

func NewHandler(repo port.ReportingRepository) *Handler {
	return &Handler{repo: repo}
}

// Register mounts laporan routes ke mux.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/reports/daily", h.daily)
	mux.HandleFunc("GET /api/reports/monthly", h.monthly)
	mux.HandleFunc("GET /api/reports/yearly", h.yearly)
}

type summary struct {
	Period              string          `json:"period"`
	InvoiceCount        int             `json:"invoice_count"`
	InvoiceTotal        float64         `json:"invoice_total"`
	PaymentCount        int             `json:"payment_count"`
	PaymentTotal        float64         `json:"payment_total"`
	OutstandingTotal    float64         `json:"outstanding_total"`
	ExpenseTotal        float64         `json:"expense_total"`
	ActiveSubscriptions int             `json:"active_subscriptions"`
	CashBalances        json.RawMessage `json:"cash_balances,omitempty"`
}

const defaultTenant = "tenant-default"

func (h *Handler) daily(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("date")
	var day time.Time
	if q == "" {
		day = time.Now()
	} else {
		d, err := time.Parse("2006-01-02", q)
		if err != nil {
			httpError(w, http.StatusBadRequest, "date harus format YYYY-MM-DD")
			return
		}
		day = d
	}
	snap, err := h.repo.GetByDate(r.Context(), defaultTenant, day)
	if err != nil {
		httpError(w, http.StatusNotFound, "snapshot tidak ditemukan untuk tanggal tersebut")
		return
	}
	writeJSON(w, toSummary(day.Format("2006-01-02"), snap))
}

func (h *Handler) monthly(w http.ResponseWriter, r *http.Request) {
	month := r.URL.Query().Get("month")
	if month == "" {
		month = time.Now().Format("2006-01")
	}
	from, err := time.Parse("2006-01", month)
	if err != nil {
		httpError(w, http.StatusBadRequest, "month harus format YYYY-MM")
		return
	}
	to := from.AddDate(0, 1, -1)
	h.rangeSummary(w, r, month, from, to)
}

func (h *Handler) yearly(w http.ResponseWriter, r *http.Request) {
	yearStr := r.URL.Query().Get("year")
	if yearStr == "" {
		yearStr = strconv.Itoa(time.Now().Year())
	}
	year, err := strconv.Atoi(yearStr)
	if err != nil || year < 2000 || year > 2200 {
		httpError(w, http.StatusBadRequest, "year tidak valid")
		return
	}
	from := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(1, 0, -1)
	h.rangeSummary(w, r, yearStr, from, to)
}

func (h *Handler) rangeSummary(w http.ResponseWriter, r *http.Request, period string, from, to time.Time) {
	snaps, err := h.repo.ListRange(r.Context(), defaultTenant, from, to)
	if err != nil {
		httpError(w, http.StatusInternalServerError, err.Error())
		return
	}
	sum := summary{Period: period}
	balances := map[string]float64{}
	for _, s := range snaps {
		sum.InvoiceCount += s.InvoiceCount
		sum.InvoiceTotal += s.InvoiceTotal
		sum.PaymentCount += s.PaymentCount
		sum.PaymentTotal += s.PaymentTotal
		sum.OutstandingTotal = s.OutstandingTotal // nilai terakhir dalam periode
		sum.ExpenseTotal += s.ExpenseTotal
		sum.ActiveSubscriptions = s.ActiveSubscriptions
		if len(s.CashBalanceJSON) > 0 {
			_ = json.Unmarshal(s.CashBalanceJSON, &balances)
		}
	}
	if len(balances) > 0 {
		if b, err := json.Marshal(balances); err == nil {
			sum.CashBalances = b
		}
	}
	writeJSON(w, sum)
}

func toSummary(period string, s reporting.DailyFinancialSnapshot) summary {
	return summary{
		Period: period, InvoiceCount: s.InvoiceCount, InvoiceTotal: s.InvoiceTotal,
		PaymentCount: s.PaymentCount, PaymentTotal: s.PaymentTotal,
		OutstandingTotal: s.OutstandingTotal, ExpenseTotal: s.ExpenseTotal,
		ActiveSubscriptions: s.ActiveSubscriptions, CashBalances: s.CashBalanceJSON,
	}
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func httpError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
