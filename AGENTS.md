<!-- code-review-graph MCP tools -->
## MCP Tools: code-review-graph

**IMPORTANT: This project has a knowledge graph. ALWAYS use the
code-review-graph MCP tools BEFORE using Grep/Glob/Read to explore
the codebase.** The graph is faster, cheaper (fewer tokens), and gives
you structural context (callers, dependents, test coverage) that file
scanning cannot.

### When to use graph tools FIRST

- **Exploring code**: `semantic_search_nodes_tool` or `query_graph_tool` instead of Grep
- **Understanding impact**: `get_impact_radius_tool` instead of manually tracing imports
- **Code review**: `detect_changes_tool` + `get_review_context_tool` instead of reading entire files
- **Finding relationships**: `query_graph_tool` with callers_of/callees_of/imports_of/tests_for
- **Architecture questions**: `get_architecture_overview_tool` + `list_communities_tool`

Fall back to Grep/Glob/Read **only** when the graph doesn't cover what you need.

### Key Tools

| Tool | Use when |
| ------ | ---------- |
| `detect_changes_tool` | Reviewing code changes — gives risk-scored analysis |
| `get_review_context_tool` | Need source snippets for review — token-efficient |
| `get_impact_radius_tool` | Understanding blast radius of a change |
| `get_affected_flows_tool` | Finding which execution paths are impacted |
| `query_graph_tool` | Tracing callers, callees, imports, tests, dependencies |
| `semantic_search_nodes_tool` | Finding functions/classes by name or keyword |
| `get_architecture_overview_tool` | Understanding high-level codebase structure |
| `refactor_tool` | Planning renames, finding dead code |

### Workflow

1. The graph auto-updates on file changes (via hooks).
2. Use `detect_changes_tool` for code review.
3. Use `get_affected_flows_tool` to understand impact.
4. Use `query_graph_tool` pattern="tests_for" to check coverage.

---

# AGENTS.md — NetOps Engine (Go Backend)

**Dokumen ini adalah instruksi operasional untuk AI agent (Claude Code, Antigravity, atau agent lain) yang menulis kode di repo ini. Ini bukan sekadar panduan gaya untuk kontributor manusia.** Baca ulang dokumen ini di awal setiap task.

Referensi wajib:
- [DEVELOPMENT-GUIDELINES.md](DEVELOPMENT-GUIDELINES.md) — **Panduan & Standar Pengembangan Definitif Proyek** (Arsitektur, Naming, Interfaces, Structs, Logging, Error Mapping).
- `Polyglot-Architecture.md` (alasan arsitektur)
- `TECH-STACK-DAN-PERSIAPAN.md` (versi library)
- `docs/adr/0005-migrasi-dari-gin-ke-net-http-servemux.md` (ADR Web Server Standar Go 1.22+)

---

## 0. Prinsip Non-Negotiable

1. **Boundary layer adalah hukum.** 
   - `domain` tidak pernah impor `adapter`/`driver`/framework eksternal/proto.
   - `usecase` hanya bergantung ke `domain` dan `port`.
   - Kalau sebuah perubahan memaksa boundary ini dilanggar, itu tanda desainnya salah — stop, jangan dipaksakan.
2. **Pemisahan Kontrak Protobuf vs Model Domain**:
   - Generated code Protobuf (`api/gen/v1`) hanya untuk transport wire serialization.
   - Domain Model (`internal/domain/`) murni Go struct.
   - Konversi bolak-balik **hanya ada di `mapper.go`** pada masing-masing subfolder `internal/adapter/connect/<domain>/`.
3. **Router Standar Go 1.22+ `net/http.ServeMux` (Gin Dihapus Penuh)**:
   - Tidak ada Gin (`gin.Engine`, `gin.Context`, `gin.WrapH`). Semua endpoint dimount ke `*http.ServeMux`.
   - Middleware menggunakan signature standar Go `func(http.Handler) http.Handler` dan dirangkai via `middleware.Chain`.
4. **Logging Terpusat via `pkg/logger` (Logrus)**:
   - Dilarang keras memakai `log.Printf`, `log.Println`, `fmt.Println` di kode production.
   - Gunakan `logger.WithComponent("<Name>").WithFields(...)`.
5. **Ukuran File Terkontrol**:
   - Maksimal **400–500 baris kode per file**. File monolitik wajib dipecah menjadi beberapa file fokus dalam folder yang sama.
6. **Setiap penyimpangan wajib ditandai komentar `// DEVIASI: <alasan>`** tepat di atas kode yang menyimpang.

---

## 1. Struktur Folder & Penempatan File

### 1.1 Struktur Lengkap — Definitif, Satu-Satunya yang Berlaku

```
polyglot/
├── cmd/
│   ├── server/
│   │   └── main.go
│   └── probe/
│       └── main.go
│
├── internal/
│   ├── app/
│   │   ├── app.go
│   │   └── shutdown_test.go
│   │
│   ├── config/
│   │   └── config.go
│   │
│   ├── domain/
│   │   ├── device/
│   │   │   ├── device.go
│   │   │   └── errors.go
│   │   ├── command/
│   │   │   ├── command.go
│   │   │   └── policy.go
│   │   ├── customer/
│   │   │   ├── customer.go
│   │   │   └── errors.go
│   │   ├── subscription/
│   │   │   └── subscription.go
│   │   ├── plan/
│   │   │   └── plan.go
│   │   ├── billing/
│   │   │   ├── invoice.go
│   │   │   └── payment.go
│   │   ├── session/
│   │   │   └── session.go
│   │   ├── knowledge/
│   │   │   └── knowledge.go
│   │   └── llm/
│   │       └── config.go
│   │
│   ├── port/
│   │   ├── device_driver.go
│   │   ├── device_repository.go
│   │   ├── customer_repository.go
│   │   ├── user_repository.go
│   │   ├── subscription_repository.go
│   │   ├── invoice_repository.go
│   │   ├── knowledge_repository.go
│   │   ├── auth_service.go
│   │   ├── cache_store.go
│   │   ├── credential_vault.go
│   │   └── audit_writer.go
│   │
│   ├── usecase/
│   │   ├── auth/
│   │   │   ├── login.go
│   │   │   └── refresh_token.go
│   │   ├── user/
│   │   │   └── manage_user.go
│   │   ├── network/
│   │   │   ├── execute_command.go
│   │   │   ├── get_device_status.go
│   │   │   ├── get_active_sessions.go
│   │   │   ├── push_config.go
│   │   │   └── stream_terminal.go
│   │   ├── device/
│   │   │   └── manage_device.go
│   │   ├── customer/
│   │   │   └── manage_customer.go
│   │   ├── billing/
│   │   │   ├── manage_invoice.go
│   │   │   └── manage_subscription.go
│   │   ├── hotspot/
│   │   │   └── hotspot_usecase.go
│   │   ├── bot/
│   │   │   └── engine.go
│   │   ├── conversation/
│   │   │   └── manage_conversation.go
│   │   ├── chat/
│   │   │   └── chat_service.go
│   │   └── knowledge/
│   │       ├── document_manager.go
│   │       └── retriever.go
│   │
│   ├── adapter/
│   │   ├── connect/
│   │   │   ├── device/
│   │   │   │   ├── device_handler.go
│   │   │   │   ├── stream_handler.go
│   │   │   │   ├── probe_handler.go
│   │   │   │   ├── mapper.go
│   │   │   │   └── router.go
│   │   │   ├── auth/
│   │   │   │   ├── auth_handler.go
│   │   │   │   ├── user_handler.go
│   │   │   │   ├── rbac_handler.go
│   │   │   │   ├── mapper.go
│   │   │   │   └── router.go
│   │   │   ├── customer/
│   │   │   │   ├── customer_handler.go
│   │   │   │   ├── mapper.go
│   │   │   │   └── router.go
│   │   │   ├── billing/
│   │   │   │   ├── billing_handler.go
│   │   │   │   ├── mapper.go
│   │   │   │   └── router.go
│   │   │   ├── hotspot/
│   │   │   │   ├── hotspot_handler.go
│   │   │   │   ├── profile_user_handler.go
│   │   │   │   ├── session_handler.go
│   │   │   │   ├── system_report_handler.go
│   │   │   │   ├── mapper.go
│   │   │   │   └── router.go
│   │   │   └── bot/
│   │   │       ├── whatsapp_session_handler.go
│   │   │       ├── whatsapp_chat_handler.go
│   │   │       ├── bot_conversation_handler.go
│   │   │       ├── knowledge_handler.go
│   │   │       ├── llm_config_handler.go
│   │   │       ├── mapper.go
│   │   │       └── router.go
│   │   │
│   │   ├── http/
│   │   │   └── middleware/
│   │   │       ├── chain.go
│   │   │       ├── cors.go
│   │   │       ├── auth.go
│   │   │       ├── rbac.go
│   │   │       ├── logger.go
│   │   │       └── recovery.go
│   │   │
│   │   ├── ws/
│   │   │   ├── sse_hub.go
│   │   │   ├── device_stream_handler.go
│   │   │   └── router.go
│   │   │
│   │   ├── mcp/
│   │   │   ├── server.go
│   │   │   ├── tool_get_device_status.go
│   │   │   ├── tool_run_command.go
│   │   │   └── tool_push_config.go
│   │   │
│   │   ├── postgres/
│   │   │   ├── store.go
│   │   │   ├── device_repository.go
│   │   │   ├── customer_repository.go
│   │   │   ├── user_repository.go
│   │   │   ├── subscription_repository.go
│   │   │   ├── invoice_repository.go
│   │   │   └── knowledge_repository.go
│   │   │
│   │   ├── auth/
│   │   │   ├── jwt.go
│   │   │   ├── refresh_token.go
│   │   │   └── casbin.go
│   │   │
│   │   ├── vault/
│   │   │   └── aes_vault.go
│   │   ├── knowledge/
│   │   │   ├── retriever.go
│   │   │   ├── manager.go
│   │   │   └── chat.go
│   │   ├── llm/
│   │   │   └── factory.go
│   │   └── redis/
│   │       └── store.go
│   │
│   ├── driver/
│   │   ├── mikrotik/
│   │   │   ├── driver.go
│   │   │   ├── commands.go
│   │   │   ├── system_resource.go
│   │   │   ├── system_ping.go
│   │   │   ├── system_scheduler.go
│   │   │   ├── system_script.go
│   │   │   ├── system_log.go
│   │   │   ├── system_identity.go
│   │   │   ├── dhcp.go
│   │   │   ├── traffic.go
│   │   │   ├── ppp.go
│   │   │   └── hotspot/
│   │   ├── cisco/
│   │   │   ├── driver.go
│   │   │   └── commands.go
│   │   ├── whatsapp/
│   │   │   ├── client.go
│   │   │   ├── qr.go
│   │   │   ├── sender.go
│   │   │   ├── events.go
│   │   │   └── mirror.go
│   │   ├── genericcli/
│   │   │   └── session.go
│   │   ├── genericssh/
│   │   │   ├── driver.go
│   │   │   └── commands.go
│   │   ├── generictelnet/
│   │   │   ├── driver.go
│   │   │   └── commands.go
│   │   ├── netconf/
│   │   │   ├── driver.go
│   │   │   └── commands.go
│   │   ├── zteolt/
│   │   │   ├── snmp.go
│   │   │   ├── telnet.go
│   │   │   └── commands.go
│   │   ├── huaweiolt/
│   │   │   ├── driver.go
│   │   │   └── commands.go
│   │   └── genieacs/
│   │       ├── client.go
│   │       └── commands.go
│   │
│   ├── platformdef/
│   │   ├── *.yaml
│   │   └── README.md
│   │
│   ├── registry/
│   │   └── registry.go
│   ├── audit/
│   │   └── writer.go
│   └── templates/
│
├── pkg/
│   ├── logger/
│   │   └── logger.go
│   ├── ping/
│   │   └── parser.go
│   ├── response/
│   │   └── errors.go
│   ├── retry/
│   │   └── retry.go
│   └── voucher/
│       └── generator.go
│
├── api/
│   ├── proto/v1/
│   │   ├── auth.proto
│   │   ├── users.proto
│   │   ├── device.proto
│   │   ├── hotspot.proto
│   │   ├── customer.proto
│   │   ├── billing.proto
│   │   └── bot.proto
│   └── gen/v1/
│       ├── *.pb.go
│       └── *_connect.pb.go
│
├── migrations/
│   ├── 000001_create_devices_table.up.sql
│   └── 000001_create_devices_table.down.sql
│
├── docs/
│   └── adr/
│       ├── 0001-pilih-gin-daripada-echo.md (Superseded)
│       ├── 0002-devicedriver-tanpa-session-terpisah.md
│       ├── 0003-mikrotik-dual-connection-streaming.md
│       ├── 0004-generic-cli-driver-scrapligo.md
│       └── 0005-migrasi-dari-gin-ke-net-http-servemux.md
│
├── DEVELOPMENT-GUIDELINES.md
├── Polyglot-Architecture.md
├── SYSTEM-STRUCTURE-AND-ARCHITECTURE.md
├── TECH-STACK-DAN-PERSIAPAN.md
├── README.md
├── Makefile
├── go.mod
└── go.sum
```

---

### 1.2 Algoritma Keputusan Penempatan File

Sebelum membuat file baru apa pun, ikuti urutan berikut:

1. **Apakah ini entity/aturan bisnis murni tanpa I/O & tanpa lib eksternal?** → `internal/domain/<nama_domain>/`
2. **Apakah ini orkestrasi satu use case (memanggil port, independen dari transport/DB)?** → `internal/usecase/<area>/<verb>_<noun>.go`
3. **Apakah ini kontrak Go interface yang diimplementasikan adapter/driver?** → `internal/port/`
4. **Apakah ini handler ConnectRPC / transport inbound?** → `internal/adapter/connect/<domain>/` (bagi menjadi `handler.go`, `mapper.go`, `router.go`)
5. **Apakah ini protokol komunikasi hardware perangkat jaringan?** → `internal/driver/<vendor>/` (dipecah menjadi `driver.go`, `commands.go`, dan submodul fitur)
6. **Apakah ini implementasi storage/infra (Postgres, Redis, Vault, Auth)?** → `internal/adapter/<postgres|redis|vault|auth>/`
7. **Apakah ini utilitas generik mandiri tanpa ketergantungan domain?** → `pkg/<nama>/`
8. **Tidak cocok satu pun di atas** → JANGAN buat file. Diskusikan dan buat proposal penempatan terlebih dahulu.

---

### 1.3 Tabel Penempatan Cepat (Lookup Reference)

| Jenis Perubahan | Path | Pola Nama File | Contoh |
|---|---|---|---|
| Entity domain baru | `internal/domain/<domain>/` | `<domain>.go` | `internal/domain/plan/plan.go` |
| Error khusus domain | folder sama dengan entity | `errors.go` | `internal/domain/device/errors.go` |
| Use case baru | `internal/usecase/<area>/` | `<verb>_<noun>.go` | `internal/usecase/auth/login.go` |
| Interface port baru | `internal/port/` | `<noun>_repository.go` / `<noun>.go` | `internal/port/user_repository.go` |
| Handler ConnectRPC baru | `internal/adapter/connect/<domain>/` | `<domain>_handler.go` | `internal/adapter/connect/device/device_handler.go` |
| DTO Mapper ConnectRPC | `internal/adapter/connect/<domain>/` | `mapper.go` | `internal/adapter/connect/device/mapper.go` |
| Router ConnectRPC | `internal/adapter/connect/<domain>/` | `router.go` | `internal/adapter/connect/device/router.go` |
| Middleware HTTP baru | `internal/adapter/http/middleware/` | `<nama>.go` | `internal/adapter/http/middleware/logger.go` |
| Tool MCP baru | `internal/adapter/mcp/` | `tool_<nama_tool_snake_case>.go` | `internal/adapter/mcp/tool_run_command.go` |
| Repository Postgres baru | `internal/adapter/postgres/` | `<resource>_repository.go` | `internal/adapter/postgres/user_repository.go` |
| Driver vendor baru | `internal/driver/<vendor>/` | `driver.go` + `commands.go` | `internal/driver/huaweiolt/driver.go` |
| Utilitas generik baru | `pkg/<name>/` | `<name>.go` | `pkg/logger/logger.go` |
| Migrasi DB baru | `migrations/` | `NNNNNN_<name>.up.sql` + `.down.sql` | `migrations/000008_create_users_table.up.sql` |
| Unit test | folder sama dengan file diuji | `<nama_file>_test.go` | `internal/usecase/auth/login_test.go` |
| ADR baru | `docs/adr/` | `NNNN-<slug-kebab-case>.md` | `docs/adr/0005-migrasi-dari-gin-ke-net-http-servemux.md` |

---

## 2. Penamaan Identifier (Package, Variabel, Fungsi, Konstanta, Interface)

### 2.1 Package
- Huruf kecil, satu kata, tanpa underscore, tanpa plural (`device`, bukan `devices`).
- Nama package **tidak diulang** di nama fungsi/struct (`device.New()`, bukan `device.NewDevice()`).
- Dilarang membuat package bernama `package` (gunakan `plan` untuk paket langganan).

### 2.2 Variabel & Fungsi
- **Exported:** `PascalCase`. **Unexported:** `camelCase`.
- **Akronim konsisten**: `deviceID`, `userID`, `httpServer`, `macAddress`, `ipAddress` (bukan `deviceId`, `userId`, `HttpServer`).
- **Awalan kata kerja untuk fungsi**: `Get*`, `List*`, `Create*`, `Update*`, `Delete*`, `Validate*`, `ToProto*`.

### 2.3 Konstanta & Enums
- `PascalCase` untuk exported, **tidak pernah `ALL_CAPS`**:
  - ✅ `const MaxRetryAttempts = 3`
  - ❌ `const MAX_RETRY_ATTEMPTS = 3`

### 2.4 Interface & Struct
- Interface single-method: akhiran `-er` (`Reader`, `Closer`, `Publisher`, `Authorizer`).
- Interface multi-method: nama benda peran (`DeviceDriver`, `UserRepository`, `TokenService`).
- **Selalu di `internal/port/`**, implementasi di `adapter/` atau `driver/`.
- Suffix peran: `*UseCase`, `*Repository`, `*ConnectHandler`, `*Driver`.

### 2.5 Receiver
- 1–2 huruf konsisten dalam satu package (`func (u *ManageUserUseCase)`, `func (d *Driver)`). Dilarang `self` atau `this`.

---

## 3. Pola Kode & Struktur Kontrol

### 3.1 Guard Clause & Early Return
- Selalu gunakan guard clause / early return alih-alih nesting `if-else`:
```go
// ✅ BOLEH
func (u *ManageUserUseCase) GetUser(ctx context.Context, id int64) (*customer.User, error) {
    if id <= 0 {
        return nil, ErrInvalidUserID
    }
    return u.repo.FindByID(ctx, id)
}
```
- Tidak ada `else` setelah cabang `if` yang sudah `return`/`continue`/`break`.

### 3.2 Error Handling & Logging
- Wrap error dengan `%w` dan konteks operasi: `fmt.Errorf("find user %d: %w", id, err)`.
- Jangan buang error dengan `_` tanpa komentar alasan.
- Gunakan `pkg/logger`:
```go
logger.WithComponent("DeviceConnect").WithFields(map[string]any{
    "device_id": req.Msg.DeviceId,
}).Info("fetching device status")
```
- Pemetaan error domain ke ConnectRPC code via `pkg/response/errors.go` (`response.MapDomainError(err)`).

### 3.3 Concurrency & Context
- `context.Context` **selalu** parameter pertama, **tidak pernah** field struct.
- Semua goroutine wajib punya mekanisme wait (`errgroup.Group`) atau cancellation (`ctx.Done()`).

---

## 4. Checklist Wajib Sebelum Selesai Task

- [ ] `go build -v ./cmd/server` (sukses tanpa error).
- [ ] `go test -v ./...` (semua test lolos 100%).
- [ ] Ukuran file tidak melebihi 400–500 baris.
- [ ] Tidak ada pelanggaran batas layer (domain bersih dari proto/adapter).
- [ ] Menggunakan `pkg/logger` (tidak ada `log.Printf`/`fmt.Println` di kode production).
- [ ] Menjalankan `go mod tidy` jika ada perubahan dependency.
