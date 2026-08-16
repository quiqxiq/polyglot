# polyglot

NetOps + ISP management backend — standalone Go service exposing MCP, ConnectRPC,
dan WebSocket/SSE untuk multi-vendor network automation.

## Dokumen Wajib Dibaca

- [DEVELOPMENT-GUIDELINES.md](DEVELOPMENT-GUIDELINES.md) — **Panduan & Standar Pengembangan Definitif** (naming, interface, struct, penempatan file, logging, error handling, checklist).
- `CLAUDE.md` / `AGENTS.md` — instruksi struktur folder, penamaan, dan gaya kode operasional AI agent.
- `Polyglot-Architecture.md` — arsitektur dan alur kerja sistem.
- `SYSTEM-STRUCTURE-AND-ARCHITECTURE.md` — dokumentasi struktur folder definitif dan diagram arsitektur sistem.
- `TECH-STACK-DAN-PERSIAPAN.md` — pemilihan teknologi dan versi library.

## ADR (Architectural Decision Records)

- `docs/adr/0001-pilih-gin-daripada-echo.md` (Digantikan oleh ADR 0005)
- `docs/adr/0002-devicedriver-tanpa-session-terpisah.md` — catatan deviasi eksplisit dari `Polyglot-Architecture.md` §5.3
- `docs/adr/0003-mikrotik-dual-connection-streaming.md` — dua koneksi persisten (exec/stream) di driver Mikrotik
- `docs/adr/0004-generic-cli-driver-scrapligo.md` — genericssh & generictelnet berbagi satu mesin scrapligo
- `docs/adr/0005-migrasi-dari-gin-ke-net-http-servemux.md` — migrasi total dari Gin ke `net/http.ServeMux` Go 1.22+

## Menjalankan

```bash
make build
make run
```

## Lint & Test

```bash
make lint
make test
```

Test integrasi ke device fisik:

```bash
# Mikrotik
MIKROTIK_TEST_HOST=192.168.88.1 \
MIKROTIK_TEST_USER=admin \
MIKROTIK_TEST_PASS=secret \
make test-integration
```
