# polyglot

NetOps + ISP management backend — standalone Go service exposing MCP, REST,
dan WebSocket/SSE untuk multi-vendor network automation.

## Dokumen Wajib Dibaca

- `CLAUDE.md` — instruksi struktur folder, penamaan, dan gaya kode (untuk AI agent). **Satu-satunya sumber kebenaran struktur folder** (§1.1).
- `Polyglot-Architecture.md` — arsitektur dan alur kerja
- `SYSTEM-STRUCTURE-AND-ARCHITECTURE.md` — dokumentasi struktur folder definitif dan arsitektur sistem
- `TECH-STACK-DAN-PERSIAPAN.md` — pemilihan teknologi dan versi library, termasuk peringatan revisi spec MCP 28 Juli 2026

Dokumen disalin otomatis ke root repo ini oleh `scaffold.sh` kalau
ditemukan di folder yang sama dengan script saat dijalankan. Kalau tidak
ditemukan, salin manual sebelum menulis baris kode pertama.

## ADR

- `docs/adr/0001-pilih-gin-daripada-echo.md`
- `docs/adr/0002-devicedriver-tanpa-session-terpisah.md` — termasuk catatan deviasi eksplisit dari `Polyglot-Architecture.md` §5.3
- `docs/adr/0003-mikrotik-dual-connection-streaming.md` — dua koneksi persisten (exec/stream) di driver Mikrotik, dan jebakan context cancellation di go-routeros v3
- `docs/adr/0004-generic-cli-driver-scrapligo.md` — genericssh & generictelnet berbagi satu mesin scrapligo; vendor jadi data (platform YAML + Catalog), bukan kode

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

Test integrasi ke device fisik (tidak ikut `make test`, butuh device asli
yang bisa dijangkau):

```
# Mikrotik
MIKROTIK_TEST_HOST=192.168.88.1 \
MIKROTIK_TEST_USER=admin \
MIKROTIK_TEST_PASS=secret \
make test-integration

# Generic SSH (vendor apa pun — tentukan lewat *_PLATFORM)
GENERICSSH_TEST_HOST=192.168.1.1 \
GENERICSSH_TEST_USER=admin \
GENERICSSH_TEST_PASS=secret \
GENERICSSH_TEST_PLATFORM=cisco_iosxe \
GENERICSSH_TEST_COMMAND="show version" \
make test-integration

# Generic Telnet (sama pola, prefix GENERICTELNET_TEST_*)
GENERICTELNET_TEST_HOST=192.168.1.2 \
GENERICTELNET_TEST_USER=admin \
GENERICTELNET_TEST_PASS=secret \
GENERICTELNET_TEST_PLATFORM=cisco_iosxe \
make test-integration
```

## Dependensi Baru (genericssh/generictelnet)

```
go get github.com/scrapli/scrapligo@v1.3.3
```

`golang.org/x/crypto`, `golang.org/x/net`, `golang.org/x/sys`,
`golang.org/x/text` ikut ter-resolve otomatis sebagai dependensi transitif
— tidak perlu langkah tambahan di lingkungan dengan akses internet normal.

# Platform Definitions

Custom scrapligo YAML platform definitions for vendors without built-in
support. See `TECH-STACK-DAN-PERSIAPAN.md` §7 and
`docs/adr/0004-generic-cli-driver-scrapligo.md` before adding a new file
here.

Naming: `<vendor>_<model>.yaml` (e.g. `zte_c320.yaml`).

**Start from scrapligo's own template, don't write one from scratch.** The
scrapligo module ships a fully-commented example at
`assets/platforms/example.yaml` (in the `github.com/scrapli/scrapligo`
module — find it in your local module cache, e.g.
`$(go env GOMODCACHE)/github.com/scrapli/scrapligo@v1.3.3/assets/platforms/example.yaml`,
or view it directly in the module's GitHub repo) — copy that as a starting
point. Also look at a built-in platform close to your target vendor (e.g.
`cisco_iosxe.yaml` in the same folder) as a second reference.

Two things worth knowing before authoring one:
- `driver-type` in the YAML must be either `'generic'` or `'network'` —
  this project's generic drivers (`internal/driver/genericssh`,
  `internal/driver/generictelnet`) work with either automatically (see
  `internal/driver/genericcli/session.go`'s `resolveDriver`), but every
  built-in platform and scrapligo's own example template use `'network'`
  (privilege levels, on-open paging-disable sequences, etc.) — that's the
  right default to follow unless you have a specific reason not to.
- `failed-when-contains` in the YAML is scrapligo's OWN generic
  error-output detection — separate from this project's
  `genericcli.Catalog.FailedWhenContains`, which does the exact same thing
  but is set from Go code per device rather than baked into the YAML file.
  Prefer putting it in the YAML if it's truly a property of the vendor
  (most cases); use `Catalog.FailedWhenContains` only for something that
  varies per-device rather than per-vendor.

Note: this folder holds scrapligo *prompt-pattern* definitions only.
Command catalog/translation and risk classification is a `genericcli.Catalog`
value (see `internal/driver/genericcli/catalog.go`) for generic SSH/Telnet
vendors, or lives in that vendor's own `internal/driver/<vendor>/commands.go`
for vendors with their own dedicated driver package (mikrotik, cisco, ...).

# Platform Definitions

Custom scrapligo YAML platform definitions for vendors without built-in
support. See `TECH-STACK-DAN-PERSIAPAN.md` §7 and
`docs/adr/0004-generic-cli-driver-scrapligo.md` before adding a new file
here.

Naming: `<vendor>_<model>.yaml` (e.g. `zte_c320.yaml`).

**Start from scrapligo's own template, don't write one from scratch.** The
scrapligo module ships a fully-commented example at
`assets/platforms/example.yaml` (in the `github.com/scrapli/scrapligo`
module — find it in your local module cache, e.g.
`$(go env GOMODCACHE)/github.com/scrapli/scrapligo@v1.3.3/assets/platforms/example.yaml`,
or view it directly in the module's GitHub repo) — copy that as a starting
point. Also look at a built-in platform close to your target vendor (e.g.
`cisco_iosxe.yaml` in the same folder) as a second reference.

Two things worth knowing before authoring one:
- `driver-type` in the YAML must be either `'generic'` or `'network'` —
  this project's generic drivers (`internal/driver/genericssh`,
  `internal/driver/generictelnet`) work with either automatically (see
  `internal/driver/genericcli/session.go`'s `resolveDriver`), but every
  built-in platform and scrapligo's own example template use `'network'`
  (privilege levels, on-open paging-disable sequences, etc.) — that's the
  right default to follow unless you have a specific reason not to.
- `failed-when-contains` in the YAML is scrapligo's OWN generic
  error-output detection — separate from this project's
  `genericcli.Catalog.FailedWhenContains`, which does the exact same thing
  but is set from Go code per device rather than baked into the YAML file.
  Prefer putting it in the YAML if it's truly a property of the vendor
  (most cases); use `Catalog.FailedWhenContains` only for something that
  varies per-device rather than per-vendor.

Note: this folder holds scrapligo *prompt-pattern* definitions only.
Command catalog/translation and risk classification is a `genericcli.Catalog`
value (see `internal/driver/genericcli/catalog.go`) for generic SSH/Telnet
vendors, or lives in that vendor's own `internal/driver/<vendor>/commands.go`
for vendors with their own dedicated driver package (mikrotik, cisco, ...).

## Isi folder ini

- `mikrotik_routeros.yaml` — MikroTik RouterOS CLI over SSH, dipakai lewat
  `internal/driver/genericssh`. Ini jalur CADANGAN, bukan pengganti
  `internal/driver/mikrotik` (RouterOS API via go-routeros) — pakai kalau
  akses API tidak tersedia di device, atau untuk command yang belum ada di
  katalog driver API. **Baca komentar di dalam file itu sebelum
  memakainya** — ada langkah wajib (suffix `+cet512w` di username) yang
  tidak bisa diekspresikan lewat YAML dan harus diset di sisi pemanggil.
  Divalidasi sejauh ini: skema YAML valid (berhasil di-parse scrapligo,
  `driver-type: network` ter-resolve), regex prompt teruji cocok untuk
  prompt normal/identity-berspasi/safe-mode/submenu. BELUM divalidasi ke
  device fisik — lihat `test/integration/mikrotik_ssh_test.go`.
