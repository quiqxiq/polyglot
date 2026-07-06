# polyglot

NetOps + ISP management backend — standalone Go service exposing MCP, REST,
dan WebSocket/SSE untuk multi-vendor network automation.

## Dokumen Wajib Dibaca

- `CLAUDE.md` — instruksi struktur folder, penamaan, dan gaya kode (untuk AI agent). **Satu-satunya sumber kebenaran struktur folder** (§1.1).
- `NetOps-Architecture.md` — arsitektur dan alur kerja
- `TECH-STACK-DAN-PERSIAPAN.md` — pemilihan teknologi dan versi library, termasuk peringatan revisi spec MCP 28 Juli 2026

Ketiganya disalin otomatis ke root repo ini oleh `scaffold.sh` kalau
ditemukan di folder yang sama dengan script saat dijalankan. Kalau tidak
ditemukan, salin manual sebelum menulis baris kode pertama.

## ADR

- `docs/adr/0001-pilih-gin-daripada-echo.md`
- `docs/adr/0002-devicedriver-tanpa-session-terpisah.md` — termasuk catatan deviasi eksplisit dari `Polyglot-Architecture.md` §5.3

## Menjalankan

```
make build
make run
```

## Lint & Test

```
make lint
make test
```
