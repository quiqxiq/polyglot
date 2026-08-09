# Polyglot Backend Audit & Refactor Plan

## Executive Summary

Proyek Polyglot adalah backend Go untuk NetOps/ISP operations yang mengikuti Clean Architecture/Ports-and-Adapters. Backend memiliki fondasi arsitektur yang baik, tetapi masih terdapat banyak sekali masalah: data dummy/hardcoded di production code, stub domain/usecase yang belum diimplementasi, deployment yang rusak, dan ketidakkonsistenan migrasi skema. Rencana ini menangani semua kategori sekaligus secara iteratif, dengan prioritas keamanan dan stabilitas sistem terlebih dahulu.

## Critical Issues Summary

| Kategori | Severity | Contoh File |
|----------|----------|--------------|
| Secrets & default insecure | Tinggi | `internal/config/config.go`, `.env.example`, `cmd/seed/main.go` |
| Repositori in-memory bypass DB | Tinggi | `internal/adapter/memory/*.go`, `internal/app/app.go` |
| Mock data di handler produksi | Tinggi | `internal/adapter/connect/billing/billing_handler.go` |
| Skema DB tidak cocok antara SQL migrations dan GORM models | Tinggi | `migrations/000002_*.sql`, `internal/adapter/postgres/models/*.go` |
| Deployment & Docker rusak | Tinggi | `deployments/docker/Dockerfile`, `deployments/docker-compose.yml`, `Makefile` |
| Stub usecase/domain belum implementasi | Menengah | `internal/usecase/billing/*.go`, `internal/domain/billing/*.go` |
| Package naming tidak konsisten | Menengah | `internal/adapter/connect/errors.go`, `api/gen/v1/*.go` |
| `context.Background()` di driver | Menengah | `internal/driver/whatsapp/*.go`, `internal/driver/mikrotik/*.go` |
| Dokumentasi & OpenAPI stale | Rendah-Menengah | `api/openapi.yaml`, `api/mcp-tools.md`, `README.md` |

## Decision Points (dari user)

1. **Prioritas**: Semua kategori dikerjakan secara iteratif.
2. **In-memory repositories**: Dihapus. Server harus gagal startup jika Postgres tidak tersedia.
3. **Skema DB**: Hybrid. SQL migrations (`golang-migrate`) adalah sumber kebenaran untuk production; `AutoMigrate` tetap ada untuk development mode.

## Phased Approach

### Phase 1 — Security & Safety (Week 1)

Tujuan: Menghilangkan bahaya langsung sebelum sentuhan fitur lain.

#### 1.1 Hapus/Hindari insecure defaults
- `internal/config/config.go`
  - Hapus fallback value untuk `JWTSecret`, `EncryptionKey`, dan secret lainnya.
  - Jika env var wajib tidak ada, server harus `log.Fatalf` dengan pesan jelas.
  - Pastikan `ENCRYPTION_KEY` panjangnya tepat 32 byte (validasi di runtime).
- `.env.example`
  - Hapus placeholder secrets. Gunakan `<CHANGEME>` atau kosongkan.
  - Tambahkan komentar per variabel.
- `cmd/seed/main.go` dan `scripts/seed.go`
  - Hapus seed user dengan password hardcoded (`admin123`, `agent123`).
  - Seed user hanya dibuat jika flag/environment eksplisit diberikan (mis. `SEED_ADMIN_PASSWORD` dari CLI).
  - Validasi password strength sebelum hashing.

#### 1.2 Hapus hardcoded credentials dari production code
- `internal/usecase/network/open_terminal.go`
  - Hapus fallback `Username: "admin", Password: "r00t"`.
  - Return error jika credential vault tidak punya kredensial.
- `internal/driver/genericssh/pty.go`
  - Hapus fallback `user = "admin"`.
  - Hindari `ssh.InsecureIgnoreHostKey()` secara default; tambahkan konfigurasi `HostKeyCallback` (bisa diasah/disetujui).
- `scripts/grpc_client/main.go`
  - Hapus hardcoded username/password. Terima dari argumen atau env.

#### 1.3 CORS & middleware
- `internal/adapter/http/middleware/cors.go`
  - Pindahkan allow localhost ke development-only build tag atau konfigurasi env.
  - Default production tidak boleh mengizinkan localhost.

### Phase 2 — Foundation: Config, Migrations, Remove In-Memory Repos (Week 1–2)

Tujuan: Membuat fondasi runtime stabil dan deterministik.

#### 2.1 Hapus in-memory repositories
- Hapus direktori `internal/adapter/memory/` (`device_repository.go`, `customer_repository.go`).
- `internal/app/app.go`
  - Hapus blok fallback ke memory.
  - Server harus langsung gagal startup (`log.Fatalf`) jika `postgres.NewStore` gagal.

#### 2.2 Perbaiki konfigurasi env
- `internal/config/config.go`
  - Hapus env var yang tidak dipakai (`POLYGLOT_DEMO_*`, `MCP_TRANSPORT`, `MCP_HTTP_ADDR`, `POLYGLOT_VAULT_KEY`, `GENIEACS_TEST_*`).
  - Tambahkan `APP_PORT` dan `PORT` ke struct Config.
  - Dukung `.env.local` secara eksplisit (load via godotenv jika ada).
- `.env.example`
  - Sinkronkan dengan `internal/config/config.go`.
  - Perbaiki `TESTCONTAINERS_POSTGRES_IMAGE` agar konsisten (`postgres:16-alpine` sesuai Docker Compose).
- `.gitignore`
  - Pastikan `.env` dan `.env.local` di-ignore.
- Periksa apakah `.env` sudah ter-commit; jika ya, hapus dari histori git.

#### 2.3 Selaraskan SQL migrations dengan GORM models
- `migrations/000002_create_bot_tables.up.sql`
  - Tambahkan kolom `username` dan `tenant_id` pada tabel `users` agar cocok dengan `UserModel`.
  - Sesuaikan constraint unique.
- `migrations/000001_create_devices_table.up.sql`
  - Tambahkan `tenant_id` jika memang diperlukan oleh `DeviceModel`.
  - Diskusikan apakah `id` harus UUID atau string; pilih satu dan sesuaikan GORM model.
- `internal/adapter/postgres/models/credential_model.go`
  - Sesuaikan skema `credentials` dengan migration: gunakan `ciphertext`/`nonce` BYTEA atau ubah migration agar cocok.
- Buat aturan: production selalu jalankan `migrate-up` terlebih dahulu; `AutoMigrate` hanya aktif ketika `ENV=development` dan dilindungi flag.

### Phase 3 — Code Quality & Consistency (Week 2)

Tujuan: Membersihkan pola kode, naming, dan mengurangi magic strings.

#### 3.1 Perbaiki package naming
- `internal/adapter/connect/errors.go`
  - Ubah `package connectadapter` menjadi `package connect` agar cocok direktori.
- `api/gen/v1/*.go`
  - Perbaiki `protoc` option/package sehingga tiap service punya nama paket sesuai domain (mis. `authpb`, `devicepb`, dll.), bukan semua `devicepb`.
- `internal/adapter/postgres/*_repository.go`
  - Hapus prefix `Postgres` (redundant di package `postgres`).
- `internal/adapter/memory/*` (sebelum dihapus)
  - Hapus prefix `Mem`.
- `internal/adapter/postgres/models/*_model.go`
  - Pertimbangkan menghapus suffix `Model` atau rename package dari `models` ke domain-specific.

#### 3.2 Magic strings → constants
- `internal/adapter/connect/billing/billing_handler.go`
  - Buat constants untuk status: `UNPAID`, `PAID`, `ACTIVE`, `CANCELLED`.
- `internal/adapter/connect/auth/auth_handler.go`
  - Gunakan `cfg.JWTExpiryHours` daripada hardcoded `24 * time.Hour`.
  - Constant `Bearer`.
- `internal/usecase/bot/engine.go`
  - Constant untuk sender type: `customer`, `bot`.
- `internal/usecase/bot/guardrail.go`
  - Pindahkan list keyword ke file konfigurasi atau constants.
- `internal/adapter/http/middleware/auth.go`
  - Constant `Bearer`.

#### 3.3 Context propagation
- `internal/driver/whatsapp/session_manager.go`
- `internal/driver/mikrotik/*.go`
  - Ganti `context.Background()` dengan context dari parameter caller. Jika memang butuh timeout, gunakan `context.WithTimeout(ctx, ...)` dari context yang masuk.

#### 3.4 Error handling & anti-patterns
- `internal/adapter/auth/jwt.go`
  - Gunakan `%w` untuk wrapping error.
  - Perbaiki `fmt.Errorf("unexpected signing method: %v", ...)`.
- `internal/adapter/postgres/models/*.go`
  - Jangan abaikan error dengan `_ = json.Marshal(...)`. Return error atau log.
- `internal/port/cache_store.go`, `internal/adapter/ws/sse_hub.go`
  - Ganti `interface{}` menjadi `any`.
- `internal/adapter/ws/hub.go`
  - Implementasi WebSocket hub atau hapus jika belum digunakan.

### Phase 4 — Functional Completion (Week 2–3)

Tujuan: Menghilangkan mock data dan menyelesaikan fitur kunci.

#### 4.1 Billing & Subscription
- `internal/domain/billing/invoice.go`, `payment.go`
- `internal/domain/subscription/subscription.go`
- `internal/domain/plan/plan.go`
- `internal/usecase/billing/manage_invoice.go`
- `internal/usecase/billing/manage_plan.go`
- `internal/usecase/billing/manage_subscription.go`
  - Implementasikan domain logic sesuai `Polyglot-Architecture.md`.
  - Hapus mock data dari `internal/adapter/connect/billing/billing_handler.go`.
  - Hubungkan handler ke use case nyata.

#### 4.2 Device drivers
- `internal/driver/cisco/`, `huaweiolt/`, `zteolt/`, `netconf/`, `genericssh/`, `generictelnet/`
  - Selesaikan stub `commands.go` dan `driver.go` per service.
  - Prioritaskan driver yang paling sering digunakan.
  - Pisahkan driver yang belum siap ke feature branch agar tidak memperburuk coverage.

#### 4.3 Probe agent
- `cmd/probe/main.go`
  - Ganti hardcoded target IPs (`1.1.1.1`, `8.8.8.8`) dan simulated latency dengan ICMP ping nyata atau konfigurasi dari server.
  - Dukung konfigurasi target via env/ConnectRPC.

### Phase 5 — Deployment & CI/CD (Week 3–4)

Tujuan: Backend dapat dibangun, dijalankan, dan di-deploy secara andal.

#### 5.1 Dockerfile
- `deployments/docker/Dockerfile`
  - Perbaiki `apk add --no-linux-headers` → hapus `--no-linux-headers`.
  - Pin base image ke `golang:1.26-alpine` (sesuai `go.mod`).
  - Gunakan multi-stage build.
  - Pastikan binary tidak berjalan sebagai root.

#### 5.2 Docker Compose
- `deployments/docker-compose.yml`
  - Perbaiki service `server` dan `web` yang masih di-comment atau update Makefile agar sinkron.
- `deployments/docker-compose.prod.yml`
  - Tambahkan service migrasi terpisah yang dijalankan sebelum `server`.
  - Gunakan healthcheck Postgres/Redis.

#### 5.3 Makefile
- Perbarui target `dev-up`, `dev-setup`, `setup`, `prod-up`, `prod-setup` agar sesuai dengan service yang ada.
- Tambahkan target: `migrate-up`, `migrate-down`, `seed`, `lint`, `test`, `security-scan`.

#### 5.4 CI/CD
- Buat `.github/workflows/ci.yml`
  - Build & test.
  - Lint dengan `golangci-lint`.
  - Security scan (`gosec`, `trivy`).
  - Validate migrations (`migrate` up/down).
  - Build Docker image.

#### 5.5 .dockerignore
- Buat `.dockerignore` root yang mengecualikan `.git`, `.env`, `docs`, `scripts`, `test`, dll.

### Phase 6 — Documentation (Week 4)

Tujuan: Dokumentasi akurat dan mudah dipahami.

#### 6.1 API docs
- `api/openapi.yaml`
  - Isi dengan endpoints REST yang sebenarnya, atau tambahkan catatan bahwa API utama via ConnectRPC.
- `api/mcp-tools.md`
  - Perbarui daftar tool sesuai `internal/adapter/mcp/server.go`.

#### 6.2 README & arsitektur
- `README.md`
  - Hapus bagian duplikat.
  - Tambahkan link ke `PANDUAN.md`, `MIKROTIK-COMMAND.md`, `analisis-api-genieacs.md`.
- `Polyglot-Architecture.md`
  - Perbarui §5.3 sesuai ADR terbaru.
- `docs/BACKEND-MIGRATION-ROADMAP.md`
  - Sesuaikan status dengan realita; tandai item yang masih TODO.

#### 6.3 AGENTS.md
- Gabungkan `AGENTS.MD` dan `AGENTS.md` menjadi satu file.

## Detailed Task List (Execution Order)

1. **Security**
   - [ ] Hapus fallback secrets di `internal/config/config.go`
   - [ ] Hapus seed password hardcoded di `cmd/seed/main.go` dan `scripts/seed.go`
   - [ ] Hapus fallback credentials di `internal/usecase/network/open_terminal.go`
   - [ ] Perbaiki SSH host key handling di `internal/driver/genericssh/pty.go`
   - [ ] Restrict CORS localhost default

2. **Foundation**
   - [ ] Hapus `internal/adapter/memory/` dan fallback di `internal/app/app.go`
   - [ ] Sinkronkan `.env.example` dengan `internal/config/config.go`
   - [ ] Hapus env var tidak terpakai
   - [ ] Perbaiki migrations SQL agar cocok dengan GORM models
   - [ ] Tentukan mode `AutoMigrate` hanya untuk development

3. **Code Quality**
   - [ ] Perbaiki package naming (`connectadapter` → `connect`, proto packages)
   - [ ] Extract magic strings ke constants
   - [ ] Propagate context di driver
   - [ ] Perbaiki error wrapping dan penggunaan `any`

4. **Functional**
   - [ ] Implementasi billing/subscription domain & usecase
   - [ ] Hapus mock data dari billing handler
   - [ ] Selesaikan stub driver per prioritas
   - [ ] Perbaiki probe agent

5. **Deployment**
   - [ ] Perbaiki Dockerfile
   - [ ] Perbaiki docker-compose dan Makefile
   - [ ] Buat `.github/workflows/ci.yml`
   - [ ] Buat `.dockerignore`

6. **Docs**
   - [ ] Perbarui OpenAPI & MCP tools doc
   - [ ] Bersihkan README & arsitektur
   - [ ] Gabungkan AGENTS files

## Verification

Untuk setiap phase, jalankan:

- `go build ./...` dan `go test ./...`
- `golangci-lint run ./...`
- `go test -race ./...`
- `docker compose -f deployments/docker-compose.yml build`
- `make migrate-up` dan `make migrate-down`
- `go run ./cmd/server` harus gagal startup jika `DATABASE_URL` kosong/salah setelah Phase 2.
- `go run ./cmd/seed` hanya boleh berjalan dengan password admin dari CLI/env.

## Risiko & Mitigasi

| Risiko | Mitigasi |
|--------|----------|
| Perubahan skema DB merusak data existing | Backup DB sebelum migrasi; test migrate up/down |
| Penghapusan in-memory repos membuat dev sulit | Dokumentasikan setup Postgres via Docker Compose |
| Implementasi billing kompleks | Pecah jadi sub-task; mulai dari invoice dulu |
| Driver vendor banyak | Prioritaskan mikrotik; driver lain ke feature branch |
| Perubahan secrets membuat existing setup rusak | Berikan `.env.example` yang jelas dan script setup |

## Notes

- Generated code `api/gen/v1/` tidak perlu direview secara manual, tetapi perlu memastikan protoc plugin dan option `go_package` benar.
- Perubahan ini besar; disarankan membagi menjadi beberapa PR/track sesuai phase.
- Setiap phase harus lalu lint/test sebelum melanjutkan ke phase berikutnya.
