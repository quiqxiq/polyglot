# 0001 — Pilih Gin daripada Echo

## Status
Digantikan oleh [0005 — Migrasi dari Gin ke net/http.ServeMux Standar Go 1.22+](file:///c:/Users/g0str/projects/polyground/polyglot/docs/adr/0005-migrasi-dari-gin-ke-net-http-servemux.md)

## Konteks
Butuh web framework untuk REST API. Kandidat: Gin, Echo, chi.

## Keputusan
Gin v1.12.0 — konsisten dengan proyek Go lain yang sudah ada (roskit),
ekosistem middleware matang. Lihat TECH-STACK-DAN-PERSIAPAN.md §2.

## Konsekuensi
Semua handler REST baru mengikuti pola Gin (`gin.Context`), bukan `net/http`
murni. (Catatan: Gin telah dihapus dan digantikan oleh `net/http.ServeMux` di ADR 0005).
