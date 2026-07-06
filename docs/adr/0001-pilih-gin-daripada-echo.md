# 0001 — Pilih Gin daripada Echo

## Status
Diterima

## Konteks
Butuh web framework untuk REST API. Kandidat: Gin, Echo, chi.

## Keputusan
Gin v1.12.0 — konsisten dengan proyek Go lain yang sudah ada (roskit),
ekosistem middleware matang. Lihat TECH-STACK-DAN-PERSIAPAN.md §2.

## Konsekuensi
Semua handler REST baru mengikuti pola Gin (`gin.Context`), bukan `net/http`
murni.
