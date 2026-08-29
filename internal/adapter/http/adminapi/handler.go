// Package adminapi menyediakan endpoint plain-HTTP untuk operasi ISP
// tingkat admin: import/export pelanggan, reconcile router, snapshot.
package adminapi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/internal/usecase/importer"
)

// Handler exposes administrative HTTP endpoints for imports and reports.
type Handler struct {
	upsert      *importer.UpsertUseCase
	routerSrc   *importer.RouterSource
	reconciler  *importer.Reconciler
	snapshotter port.SnapshotComputer
	exporter    *importer.ExportUseCase
	resolve     func(ctx context.Context, deviceID string) (port.DeviceDriver, bool)
}

// NewHandler constructs an administrative HTTP handler.
func NewHandler(
	upsert *importer.UpsertUseCase,
	routerSrc *importer.RouterSource,
	reconciler *importer.Reconciler,
	snapshotter port.SnapshotComputer,
	exporter *importer.ExportUseCase,
	driverResolver func(ctx context.Context, deviceID string) (port.DeviceDriver, bool),
) *Handler {
	return &Handler{
		upsert: upsert, routerSrc: routerSrc, reconciler: reconciler,
		snapshotter: snapshotter, exporter: exporter, resolve: driverResolver,
	}
}

// RegisterProtected mounts admin endpoints (di balik middleware JWT+RBAC).
func (h *Handler) RegisterProtected(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/admin/import", h.importFile)
	mux.HandleFunc("POST /api/admin/import-router", h.importRouter)
	mux.HandleFunc("GET /api/admin/reconcile", h.reconcile)
	mux.HandleFunc("GET /api/admin/export", h.export)
	mux.HandleFunc("POST /api/admin/snapshot/refresh", h.snapshotRefresh)
}

// ─── Import file ────────────────────────────────────────────────────────

func (h *Handler) importFile(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(20 << 20); err != nil { // maks 20 MB
		writeErr(w, http.StatusBadRequest, "multipart form diperlukan (field 'file')")
		return
	}
	file, hdr, err := r.FormFile("file")
	if err != nil {
		writeErr(w, http.StatusBadRequest, "field file wajib")
		return
	}
	defer func() { _ = file.Close() }()

	var rows []importer.Row
	switch {
	case hasSuffix(hdr.Filename, ".xlsx"):
		buf, rerr := io.ReadAll(file)
		if rerr != nil {
			writeErr(w, http.StatusBadRequest, "baca file gagal")
			return
		}
		rows, err = importer.ParseXLSX(buf)
	default:
		rows, err = importer.ParseCSV(file)
	}
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	if verrs := importer.ValidateRows(rows); len(verrs) > 0 {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]any{
			"error": "validasi gagal", "details": errMsgs(verrs),
		})
		return
	}
	res, uerr := h.upsert.Import(r.Context(), rows)
	if uerr != nil {
		writeErr(w, http.StatusInternalServerError, uerr.Error())
		return
	}
	writeJSON(w, http.StatusOK, res)
}

// ─── Import dari router live ────────────────────────────────────────────

func (h *Handler) importRouter(w http.ResponseWriter, r *http.Request) {
	deviceID := r.URL.Query().Get("device_id")
	if deviceID == "" {
		writeErr(w, http.StatusBadRequest, "query device_id wajib")
		return
	}
	driver, ok := h.resolveDriver(w, r, deviceID)
	if !ok {
		return // respons sudah ditulis
	}
	deviceName := orDefault(r.URL.Query().Get("device_name"), deviceID)
	rows, err := h.routerSrc.PullPPPoERows(r.Context(), driver, deviceName)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if verrs := importer.ValidateRows(rows); len(verrs) > 0 {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]any{
			"error":        "beberapa baris tidak lengkap — lengkapi lalu impor ulang",
			"details":      errMsgs(verrs),
			"preview_rows": previewRows(rows),
		})
		return
	}
	res, uerr := h.upsert.Import(r.Context(), rows)
	if uerr != nil {
		writeErr(w, http.StatusInternalServerError, uerr.Error())
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (h *Handler) reconcile(w http.ResponseWriter, r *http.Request) {
	deviceID := r.URL.Query().Get("device_id")
	if deviceID == "" {
		writeErr(w, http.StatusBadRequest, "query device_id wajib")
		return
	}
	driver, ok := h.resolveDriver(w, r, deviceID)
	if !ok {
		return
	}
	report, err := h.reconciler.Compare(r.Context(), deviceID, driver)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, report)
}

// ─── Snapshot manual trigger ────────────────────────────────────────────

func (h *Handler) snapshotRefresh(w http.ResponseWriter, r *http.Request) {
	day := todayUTC()
	if q := r.URL.Query().Get("date"); q != "" {
		d, err := time.Parse("2006-01-02", q)
		if err != nil {
			writeErr(w, http.StatusBadRequest, "date harus YYYY-MM-DD")
			return
		}
		day = d
	}
	if err := h.snapshotter.RecomputeDaily(r.Context(), tenantDefault(), day); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"message": "snapshot diperbarui", "date": day.Format("2006-01-02"),
	})
}

// ─── Export ─────────────────────────────────────────────────────────────

func (h *Handler) export(w http.ResponseWriter, r *http.Request) {
	format := orDefault(r.URL.Query().Get("format"), "csv")
	data, err := h.exporter.ExportAll(r.Context(), format)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if format == "xlsx" {
		w.Header().Set("Content-Type",
			"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		w.Header().Set("Content-Disposition", "attachment; filename=pelanggan.xlsx")
	} else {
		w.Header().Set("Content-Type", "text/csv; charset=utf-8")
		w.Header().Set("Content-Disposition", "attachment; filename=pelanggan.csv")
	}
	_, _ = w.Write(data)
}

// ─── shared ─────────────────────────────────────────────────────────────

func (h *Handler) resolveDriver(w http.ResponseWriter, r *http.Request, deviceID string) (port.DeviceDriver, bool) {
	drv, ok := h.resolve(r.Context(), deviceID)
	if !ok || drv == nil {
		writeErr(w, http.StatusNotFound, "device tidak ditemukan: "+deviceID)
		return nil, false
	}
	return drv, true
}

func hasSuffix(name, suffix string) bool {
	return strings.HasSuffix(strings.ToLower(name), suffix)
}

func orDefault(v, def string) string {
	if v == "" {
		return def
	}
	return v
}

func todayUTC() time.Time {
	n := time.Now().UTC()
	return time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, time.UTC)
}

func tenantDefault() string { return "tenant-default" }

func errMsgs(errs []error) []string {
	out := make([]string, len(errs))
	for i, e := range errs {
		out[i] = e.Error()
	}
	return out
}

func previewRows(rows []importer.Row) []string {
	limit := len(rows)
	if limit > 5 {
		limit = 5
	}
	out := make([]string, limit)
	for i := 0; i < limit; i++ {
		out[i] = rows[i].Username + "@" + rows[i].PlanName +
			" (" + strconv.Itoa(rows[i].RowNumber) + ")"
	}
	return out
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}
