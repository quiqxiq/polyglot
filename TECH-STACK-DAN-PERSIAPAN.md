# Persiapan & Tech Stack — NetOps Engine (Go)

> **Status:** v2 — keputusan tegas, semua versi diverifikasi langsung (bukan dari memori pelatihan)
> **Melengkapi:** `Polyglot-Architecture.md`
> **Model tenancy:** Single-tenant per deployment

Versi v1 dokumen ini menyajikan beberapa pilihan sebagai "opsi A atau B" — itu keliru. Di bawah ini satu keputusan per kategori, dengan versi yang sudah saya cek langsung (bukan tebakan), plus alasan singkat kalau ada penyimpangan dari yang Anda sebutkan.

---

## ⚠️ Perhatikan Ini Dulu Sebelum Mulai: Perubahan Besar Spesifikasi MCP (28 Juli 2026)

Ini temuan paling penting dari riset saya dan berlaku terlepas dari keputusan tech stack lainnya: **spesifikasi protokol MCP akan mengalami revisi besar yang final pada 28 Juli 2026** — sekitar 3-4 minggu dari sekarang. Perubahan utamanya: protokol menjadi *stateless* (menghilangkan initialize handshake dan session di level protokol), supaya server MCP bisa di-scale horizontal dengan load balancer round-robin biasa tanpa sticky session.

SDK resmi Go, Python, TypeScript, dan C# sudah punya rilis beta untuk spec baru ini (`go get github.com/modelcontextprotocol/go-sdk@v1.7.0-pre.1`), tapi **instalasi normal tanpa menyebut pre-release tetap mendapat versi stabil saat ini**, dan client lama otomatis fallback ke handshake versi sebelumnya. Implikasi konkret untuk Anda:
- Mulai development sekarang dengan **versi stabil saat ini** (bukan pre-release) — jangan tunda proyek menunggu spec final.
- Tapi desain `internal/adapter/mcp/` Anda sebaiknya **tidak mengasumsikan session-per-koneksi yang persisten sebagai satu-satunya model** — begitu spec final rilis akhir Juli, ada kemungkinan kuat Anda akan upgrade dalam waktu dekat, dan model stateless ini justru **kemungkinan menyederhanakan** sebagian kompleksitas connection-state yang kita bahas sebelumnya di `Polyglot-Architecture.md`.
- Jadwalkan review ulang bagian MCP setelah 28 Juli 2026.

---

## Ringkasan Tegas — Satu Baris Per Kategori

| Kategori | Keputusan Final | Import/Install |
|---|---|---|
| Bahasa | **Go 1.26** (rilis 10 Feb 2026; patch terbaru 1.26.4) | `go 1.26` di `go.mod` |
| Web framework | **Gin v1.12.0** | `github.com/gin-gonic/gin` |
| ORM | **GORM** | `gorm.io/gorm` |
| RBAC | **Casbin v3.10.0** | `github.com/casbin/casbin/v3` |
| Adapter Casbin↔GORM | **gorm-adapter v3** | `github.com/casbin/gorm-adapter/v3` |
| JWT | **golang-jwt v5** | `github.com/golang-jwt/jwt/v5` |
| MCP SDK | **SDK resmi Go (Anthropic + Google)**, versi stabil saat ini (bukan pre-release) | `github.com/modelcontextprotocol/go-sdk/mcp` |
| Mikrotik | **go-routeros v3** | `github.com/go-routeros/routeros/v3` |
| Cisco/SSH-CLI umum/NETCONF | **scrapligo v1.3.3** | `github.com/scrapli/scrapligo` |
| OLT ZTE | **gosnmp** (monitoring) + Telnet client mentah (provisioning) | `github.com/gosnmp/gosnmp` |
| ACS/TR-069 | REST client ke GenieACS NBI | `net/http` standar, tidak perlu library |
| WebSocket | **coder/websocket** | `github.com/coder/websocket` |
| Logging | **zerolog** | `github.com/rs/zerolog` |
| Migrasi | **golang-migrate** | `github.com/golang-migrate/migrate/v4` |
| Testing | **testify** + **testcontainers-go** | `github.com/stretchr/testify`, `github.com/testcontainers/testcontainers-go` |
| Linting | **golangci-lint** + **gofumpt** | terpisah dari go.mod, alat CLI |
| Dokumentasi API | **swaggo/swag** | `github.com/swaggo/swag` |

Detail dan alasan setiap baris ada di bagian bawah. Cek `go list -m -versions <module>` saat instalasi untuk patch terbaru — tabel di atas benar per periset saya hari ini, tapi patch rilis terus berjalan.

---

## Koreksi dari Sebutan Anda — Ditegaskan, Bukan Didiamkan

Anda menyebut beberapa nama spesifik. Dua di antaranya saya ubah, dengan alasan konkret:

- **`nhooyr.io/websocket` → `github.com/coder/websocket`.** Ini **sama library, sama pengembang inti** — repo-nya resmi dipindahkan ke Coder, dan `nhooyr.io/websocket` sekarang **deprecated** (setiap fungsi di package itu memberi warning "coder now maintains this library"). Import path yang benar sekarang adalah `github.com/coder/websocket`. Tidak ada trade-off di sini — ini murni migrasi nama, pilih yang tidak deprecated.
- **Logging → `zerolog`, bukan logrus.** Saya asumsikan maksud Anda `logrus` (`sirupsen/logrus`). Saya tetap pada `zerolog`: alokasi memori jauh lebih rendah (relevan untuk service yang menangani banyak koneksi device concurrent), output JSON native tanpa reflection, dan lebih idiomatic untuk service jaringan yang log-heavy. `logrus` sendiri sudah dalam mode maintenance (pengembangnya sudah menyarankan `zerolog`/`slog` untuk proyek baru) — jadi ini juga bukan sekadar preferensi pribadi.
- **Casbin v3 — konfirmasi Anda benar.** Casbin memang sudah di major version v3 (sekarang di bawah payung Apache Casbin Incubating), rilis stabil terbaru v3.10.0, import path tetap `github.com/casbin/casbin/v3`. Ditambah: ada `github.com/casbin/gorm-adapter/v3` resmi untuk menyimpan policy langsung di GORM/Postgres Anda — cocok sekali dengan stack yang sudah dipilih.
- **MCP dari Google dan Anthropic — konfirmasi Anda benar.** `github.com/modelcontextprotocol/go-sdk` sekarang **stabil di v1.x** (bukan lagi eksperimental seperti saat saya cek pertama kali beberapa minggu lalu), dengan komitmen tidak ada breaking change dalam v1. Lihat peringatan spec baru di atas.

---

## 1. Bahasa: Go 1.26

Go 1.26 dirilis 10 Februari 2026. Dua perubahan bahasa yang relevan untuk gaya kode di `CLAUDE.md` proyek ini:
- `new(expr)` sekarang bisa menerima ekspresi untuk inisialisasi langsung (bukan cuma alokasi zero-value) — berguna untuk pointer ke value optional (JSON, dsb).
- Generic type sekarang boleh mereferensikan dirinya sendiri di parameter tipe — relevan kalau `port.Repository[T]` generic dipakai.
- Green Tea GC aktif default, overhead cgo turun ~30% — relevan karena scrapligo memuat `libscrapli` via `purego` (FFI, bukan cgo, jadi tidak kena biaya cgo yang berkurang ini, tapi baik untuk tahu).

Set `go 1.26` di `go.mod`. Gin 1.12.0 sendiri sudah mensyaratkan CI-nya jalan di Go 1.25+, jadi 1.26 aman dan konsisten dengan seluruh ekosistem yang dipakai.

## 2. Web Framework: Gin v1.12.0

Tidak berubah dari rekomendasi sebelumnya — konsisten dengan roskit. `Version = "v1.12.0"` adalah rilis stabil terbaru (28 Feb 2026), sudah termasuk perbaikan race condition mode gin dan dukungan `GetError`/`GetErrorSlice` yang berguna untuk middleware error handling terpusat.

## 3. RBAC: Casbin v3.10.0 + gorm-adapter v3

```go
import (
    "github.com/casbin/casbin/v3"
    gormadapter "github.com/casbin/gorm-adapter/v3"
)

a, _ := gormadapter.NewAdapterByDB(db) // db = *gorm.DB Anda yang sudah ada
e, _ := casbin.NewEnforcer("model.conf", a)
```

Model RBAC: `superadmin / owner / admin / staff / teknisi`, tanpa domain/tenant (lihat Bagian 0 di `Polyglot-Architecture.md`). Migrasi dari v2 ke v3 kalau Anda pernah punya kode contoh lama: cukup ganti `/v2` jadi `/v3` di semua import path, API-nya tidak berubah drastis.

## 4. Auth: golang-jwt v5

`github.com/golang-jwt/jwt/v5` — API dianggap stabil oleh maintainer ("should be very few backward-incompatible changes outside major version"). v5 memperbaiki desain `Claims` interface dan validasi `exp`/`iat`/`aud` dibanding v4. Jangan pakai v4 atau versi `dgrijalva/jwt-go` yang sudah tidak di-maintain.

## 5. MCP SDK: SDK Resmi Go, Versi Stabil Saat Ini

```go
import "github.com/modelcontextprotocol/go-sdk/mcp"

server := mcp.NewServer(&mcp.Implementation{Name: "polyglot", Version: "v1.0.0"}, nil)
mcp.AddTool(server, &mcp.Tool{Name: "get_device_status", Description: "..."}, GetDeviceStatusHandler)
```

SDK ini **sudah stabil di v1.x** dengan komitmen no-breaking-change dalam major version — bukan lagi status eksperimental. Fitur yang relevan untuk arsitektur kita: `ToolAnnotations` mendukung hint `ReadOnly`/`Destructive`/`Idempotent` per tool — **pasang ini di setiap tool** (`get_device_status` → `ReadOnly: true`, `push_config` → `Destructive: true`) supaya kebijakan HITL di LibreChat punya sinyal eksplisit dari sisi MCP server, bukan hanya dari `command_policy` internal Anda.

**Jangan install versi pre-release** (`@v1.7.0-pre.1` dst) untuk pekerjaan produksi sampai spec 28 Juli 2026 final dan Anda sudah sengaja memutuskan untuk upgrade.

## 6. Mikrotik: go-routeros v3

```go
import "github.com/go-routeros/routeros/v3"

c, err := routeros.DialContext(ctx, address, username, password)
```

Tidak berubah dari keputusan sebelumnya — protokol API native, bukan CLI-scraping. v3 menambahkan varian `*Context` untuk semua fungsi dial/listen (`DialContext`, `DialTLSContext`, `ListenArgsContext`) — **selalu pakai varian `*Context`**, konsisten dengan aturan "context.Context selalu ada" di `CLAUDE.md`.

## 7. Cisco/SSH-CLI/NETCONF: scrapligo v1.3.3

Koreksi dari riset saya sebelumnya: yang stabil dan jadi rilis "Latest" resmi adalah **v1.3.3** di jalur utama `github.com/scrapli/scrapligo` (bukan `/v2`, yang statusnya belum jelas kematangannya). Kabar baiknya, arsitektur YAML-platform-definition yang saya jelaskan di v1 dokumen ini **sudah ada di v1.3.3** — bukan eksklusif ke v2 seperti asumsi saya sebelumnya. Bonus temuan: **dukungan Huawei VRP CLI sudah di-merge ke v1.3.3** (kontribusi komunitas, PR #170), jadi tidak perlu Anda tulis sendiri untuk Huawei VRP — cukup untuk SmartAX OLT kalau belum tercakup, validasi saat implementasi.

```go
import (
    "github.com/scrapli/scrapligo/driver/options"
    "github.com/scrapli/scrapligo/platform"
)

p, err := platform.NewPlatform("cisco_iosxe", host, options.WithAuthUsername(user), options.WithAuthPassword(pass))
```

## 8. OLT ZTE: gosnmp + Telnet Mentah

Tidak berubah — tidak ada platform definition scrapligo untuk ZTE, pendekatan `gosnmp` (monitoring) + klien Telnet langsung (provisioning) tetap jadi satu-satunya jalan realistis, sesuai proyek Go serupa di komunitas ISP Indonesia yang saya temukan sebelumnya.

## 9. WebSocket: coder/websocket

```go
import "github.com/coder/websocket"

c, _, err := websocket.Accept(w, r, nil)
```

Lihat koreksi di atas — ini adalah `nhooyr.io/websocket` dengan nama baru, bukan library berbeda.

## 10. Logging: zerolog

```go
import "github.com/rs/zerolog"

log.Info().Str("device_id", id).Str("command", cmd).Msg("command executed")
```

Alasan detail ada di bagian koreksi di atas. Konfigurasi output JSON untuk agregasi log (Loki, dst), gunakan `zerolog.ConsoleWriter` hanya untuk local development agar terbaca manusia.

## 11. Testing: testify + testcontainers-go

Tidak berubah. `require` untuk assertion yang harus menghentikan test kalau gagal (mis. setup gagal), `assert` untuk assertion yang boleh lanjut mengecek hal lain di test yang sama. Selaras filosofi "real logic over mocks" yang sudah Anda internalisasi dari LibreChat.

## 12. Migrasi & Dokumentasi API

- `golang-migrate/migrate/v4` — SQL murni, `up`/`down` per file.
- `swaggo/swag` — generate OpenAPI dari comment annotation di handler Gin, supaya kontrak REST selalu sinkron dengan kode, bukan dokumen terpisah yang basi.

---

## Checklist Persiapan (Diperbarui)

- [ ] Go 1.26 terpasang (`go version` menunjukkan `go1.26.x`)
- [ ] `golangci-lint` + `gofumpt` terkonfigurasi sesuai `CLAUDE.md`
- [ ] Docker + Docker Compose untuk Postgres, GenieACS
- [ ] Validasi `go-routeros/v3` connect ke Mikrotik CHR di GNS3 Anda — langkah teknis pertama yang paling murah untuk divalidasi
- [ ] **Baca ulang bagian MCP di atas tanggal 28 Juli 2026** sebelum menganggap desain MCP adapter final
- [ ] `CLAUDE.md` (lihat file terpisah) dibaca oleh siapa pun yang mulai menulis kode, sebelum baris kode pertama
