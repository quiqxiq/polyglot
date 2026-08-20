# CLAUDE.md — Polyglot NetOps Engine (Go Backend)

**Dokumen ini adalah instruksi operasional untuk AI agent (Claude Code atau agent lain) yang menulis kode di repo ini.** Baca ulang dokumen ini di awal setiap task.

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

## 2. Penamaan Identifier & Aturan Koding

### 2.1 Penamaan
- **Package:** Huruf kecil, satu kata, tanpa plural (`device`, `plan`, `customer`).
- **Struct / Interface Suffix:** `*UseCase`, `*Repository`, `*ConnectHandler`, `*Driver`.
- **Method / Function:** Awalan kata kerja `Get*`, `List*`, `Create*`, `Update*`, `Delete*`, `ToProto*`.
- **Akronim:** Konsisten kapitalisasi (`deviceID`, `userID`, `httpServer`, `macAddress`).
- **Konstanta:** `PascalCase` (tidak pernah `ALL_CAPS`).
- **Receiver:** 1–2 huruf konsisten (`func (u *ManageUserUseCase)`). Dilarang `self`/`this`.

### 2.2 Control Flow
- Guard clause / early return wajib. Dilarang nesting bertingkat.
- Tidak ada `else` setelah cabang `if` yang sudah `return`/`panic`/`continue`.
- Maksimal 2 `else if` (3 cabang total); selebihnya wajib `switch`.

### 2.3 Error & Logging
- Wrap error dengan `%w` dan konteks: `fmt.Errorf("find user %d: %w", id, err)`.
- Logging wajib melalui `pkg/logger`: `logger.WithComponent("Name").WithFields(...)`.
- Dilarang `log.Printf` dan `fmt.Println` di kode production.

### 2.4 Checklist Selesai Task
- [ ] `go build -v ./cmd/server` (sukses).
- [ ] `go test -v ./...` (semua test lolos).
- [ ] File size tidak melebihi 400–500 baris.
- [ ] Boundary domain bersih (tidak impor proto/adapter).
