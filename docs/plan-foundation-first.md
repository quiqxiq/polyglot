# Implementation Plan: Foundation First

## Overview

Polyglot saat ini berjalan sebagai MCP-only demo dengan in-memory repository dan vault. Rencana ini membuat backend Go benar-benar *production-ready* dengan PostgreSQL, GORM, AES-GCM vault, JWT/Casbin auth, dan Gin REST server. Setelah fase ini selesai, semua CRUD admin (device, customer, subscription, plan, invoice) dan operasi MCP akan membaca/menulis dari database nyata, bukan dari memory.

Target akhir fase ini: `make db-up && make migrate-up && make run` menjalankan satu binary yang:
- Menyajikan MCP server (HTTP atau stdio).
- Menyajikan REST API Gin di port terpisah.
- Membaca device + credentials dari PostgreSQL.
- Mengautentikasi admin via JWT dan mengotorisasi via Casbin RBAC.
- Mengenkripsi kredensial device dengan AES-GCM.

## Architecture Decisions

1. **GORM v2 + pgx** untuk PostgreSQL — sesuai `TECH-STACK-DAN-PERSIAPAN.md` §3.
2. **AES-GCM 256-bit** untuk credential vault — kunci dari `POLYGLOT_VAULT_KEY` (32 byte base64). Ciphertext dan nonce disimpan di tabel `credentials` per migration `000001`.
3. **JWT v5 + Casbin v3** untuk auth/RBAC — sesuai stack; role user dari tabel `users` (`superadmin`, `owner`, `admin`, `staff`, `teknisi`).
4. **Gin v1.12.0** untuk REST — satu `http.Server` terpisah dari MCP HTTP handler.
5. **In-memory Casbin policy cukup untuk fase ini** — file `configs/rbac_policy.csv` dan `configs/rbac_model.conf`, bukan policy DB. Policy loader tetap bisa baca dari env/file.
6. **Tidak membuat migration baru** — schema sudah lengkap (21 migrations). Hanya implementasi GORM model yang map ke table existing.
7. **Repository minimal dulu** — hanya method yang dibutuhkan MCP (`FindByID` device) dan REST CRUD dasar. Method lanjutan (search, pagination, filter) masuk fase berikutnya.
8. **Tidak mengubah domain layer** — `domain.Device`, `domain.customer.Customer`, dll. tetap seperti adanya; adapter hanya memetakan ke/from database.

## Task List

### Phase 1: Dependencies & Configuration

#### Task 1: Add Go dependencies

**Description:** Tambahkan library yang direncanakan ke `go.mod`.

**Acceptance criteria:**
- [x] `go.mod` berisi: `github.com/gin-gonic/gin v1.12.0`, `gorm.io/gorm`, `gorm.io/driver/postgres`, `github.com/golang-jwt/jwt/v5`, `github.com/casbin/casbin/v3`, `github.com/joho/godotenv` (atau loader env pilihan), `golang.org/x/crypto`.
- [x] `go mod tidy` berhasil tanpa error.
- [x] `go build ./...` masih berhasil (belum ada import baru yang digunakan).

**Verification:**
- [x] Tests pass: `go test ./...`
- [x] Build succeeds: `go build ./...`

**Dependencies:** None
**Files likely touched:** `go.mod`, `go.sum`
**Estimated scope:** Small

---

#### Task 2: Implement config loader

**Description:** Implementasi `internal/config/config.go` untuk membaca env var dan `.env` file. Output berupa struct `Config` yang bisa di-pass ke adapter constructor.

**Acceptance criteria:**
- [x] `config.Load(ctx)` membaca `.env` (jika ada) via `godotenv` lalu mengambil env var.
- [x] `Config` berisi field: `DatabaseURL`, `VaultKey`, `JWTSecret`, `JWTExpiry`, `MCPTransport`, `MCPHTTPAddr`, `RESTAddr`, `Environment`.
- [x] Validasi: `DatabaseURL`, `VaultKey`, `JWTSecret` wajib terisi di non-`test` environment.
- [x] Unit test untuk `Load` dengan env var sementara.

**Verification:**
- [x] Tests pass: `go test ./internal/config/...`
- [x] Build succeeds: `go build ./...`

**Dependencies:** Task 1
**Files likely touched:** `internal/config/config.go`, `internal/config/config_test.go`
**Estimated scope:** Small

---

#### Task 3: Setup PostgreSQL connection + GORM

**Description:** Implementasi `internal/adapter/postgres/db.go` untuk membuka koneksi GORM ke PostgreSQL dengan `DATABASE_URL`.

**Acceptance criteria:**
- [x] `NewDB(ctx, databaseURL)` mengembalikan `*gorm.DB` atau error.
- [x] Gunakan `gorm.io/driver/postgres` dengan `pgx`.
- [x] Connection pool di-configure: `SetMaxOpenConns`, `SetMaxIdleConns`, `SetConnMaxLifetime` (ambil dari Config/defaults).
- [x] Health check method `Ping(ctx) error`.
- [x] `Close()` graceful.

**Verification:**
- [x] Tests pass: `go test ./internal/adapter/postgres/...` (gunakan testcontainers untuk integration test, skip jika `TESTCONTAINERS` tidak tersedia).
- [x] Build succeeds: `go build ./...`

**Dependencies:** Task 2
**Files likely touched:** `internal/adapter/postgres/db.go`, `internal/adapter/postgres/db_test.go`
**Estimated scope:** Medium

---

### Phase 2: Persistence Layer

#### Task 4: Device repository

**Description:** Implementasi `internal/adapter/postgres/device_repository.go` dengan GORM model map ke tabel `devices`.

**Acceptance criteria:**
- [x] `NewDeviceRepository(db) port.DeviceRepository`.
- [x] Method: `FindByID(ctx, id)`, `FindAll(ctx)`, `Create(ctx, device)`, `Update(ctx, device)`, `Delete(ctx, id)`.
- [x] GORM model mencakup semua kolom migration `000001` + `000002` (site_name, is_active, dll).
- [x] `FindByID` mengembalikan `device.ErrNotFound` jika tidak ada.
- [x] Unit test dengan in-memory SQLite (atau testcontainers Postgres).

**Verification:**
- [x] Tests pass: `go test ./internal/adapter/postgres/... -run Device`
- [x] Build succeeds: `go build ./...`

**Dependencies:** Task 3
**Files likely touched:** `internal/adapter/postgres/device_repository.go`, `internal/adapter/postgres/device_repository_test.go`, `internal/adapter/postgres/models.go`
**Estimated scope:** Medium

---

#### Task 5: AES-GCM credential vault

**Description:** Implementasi `internal/adapter/vault/aes_vault.go` untuk enkripsi/dekripsi blob kredensial.

**Acceptance criteria:**
- [x] `NewAESVault(key []byte) port.CredentialVault`.
- [x] `Get(ctx, deviceID)` query row `credentials`, dekripsi `ciphertext` dengan `nonce`, return `device.Credentials`.
- [x] `Store(ctx, deviceID, creds)` (opsional tapi sangat disarankan) untuk mengenkripsi dan upsert ke tabel `credentials`.
- [x] Gunakan AES-GCM 256-bit dengan nonce 12 byte random per enkripsi.
- [x] `Get` mengembalikan `device.ErrNotFound` jika row tidak ada.
- [x] Unit test round-trip encrypt/decrypt.

**Verification:**
- [x] Tests pass: `go test ./internal/adapter/vault/...`
- [x] Build succeeds: `go build ./...`

**Dependencies:** Task 3
**Files likely touched:** `internal/adapter/vault/aes_vault.go`, `internal/adapter/vault/aes_vault_test.go`
**Estimated scope:** Medium

---

#### Task 6: User repository

**Description:** Implementasi repository untuk tabel `users` dengan password hashing (bcrypt).

**Acceptance criteria:**
- [x] `NewUserRepository(db) port.UserRepository` — perlu tambahkan interface `port.UserRepository`.
- [x] Method: `FindByUsername(ctx, username)`, `Create(ctx, user)`, `UpdatePassword(ctx, id, passwordHash)`, `UpdateLastLogin(ctx, id)`.
- [x] Password di-hash dengan `bcrypt` (cost default 10) — helper di `internal/adapter/auth/password.go`.
- [x] GORM model map ke tabel `users` dari migration `000004`.

**Verification:**
- [x] Tests pass: `go test ./internal/adapter/postgres/... -run User`
- [x] Build succeeds: `go build ./...`

**Dependencies:** Task 3
**Files likely touched:** `internal/adapter/postgres/user_repository.go`, `internal/adapter/postgres/user_repository_test.go`, `internal/port/user_repository.go`, `internal/adapter/auth/password.go`
**Estimated scope:** Medium

---

#### Task 7: Customer repository

**Description:** Implementasi `internal/adapter/postgres/customer_repository.go`.

**Acceptance criteria:**
- [x] `NewCustomerRepository(db) port.CustomerRepository` dengan signature yang benar: `FindByID(ctx, id string) (customer.Customer, error)` — perlu perbaiki interface yang saat ini `FindByID(ctx, id) error`.
- [x] Method: `FindByID`, `FindAll`, `Create`, `Update`, `Delete`.
- [x] GORM model map ke tabel `customers` + `customer_documents`.
- [x] `customer.Customer` domain diperluas dulu (struct field minimal: ID, FullName, Phone, Email, Address, Status, CustomerType, RegisteredAt).

**Verification:**
- [x] Tests pass: `go test ./internal/adapter/postgres/... -run Customer`
- [x] Build succeeds: `go build ./...`

**Dependencies:** Task 6
**Files likely touched:** `internal/adapter/postgres/customer_repository.go`, `internal/domain/customer/customer.go`, `internal/port/customer_repository.go`, `internal/adapter/postgres/customer_repository_test.go`
**Estimated scope:** Medium

---

#### Task 8: Plan repository

**Description:** Implementasi repository untuk tabel `plans`.

**Acceptance criteria:**
- [x] `NewPlanRepository(db) port.PlanRepository` — tambahkan interface.
- [x] Method: CRUD + `FindActive`.
- [x] Domain `plan.Plan` didefinisikan.
- [x] GORM model map ke tabel `plans` dari migration `000005`.

**Verification:**
- [x] Tests pass: `go test ./internal/adapter/postgres/... -run Plan`
- [x] Build succeeds: `go build ./...`

**Dependencies:** Task 7
**Files likely touched:** `internal/adapter/postgres/plan_repository.go`, `internal/domain/plan/plan.go`, `internal/port/plan_repository.go`, `internal/adapter/postgres/plan_repository_test.go`
**Estimated scope:** Medium

---

#### Task 9: Subscription repository

**Description:** Implementasi repository untuk tabel `subscriptions`.

**Acceptance criteria:**
- [x] `NewSubscriptionRepository(db) port.SubscriptionRepository` dengan signature benar: `FindByID(ctx, id string) (subscription.Subscription, error)`.
- [x] Method: CRUD + `FindByCustomer`, `FindByDevice`.
- [x] Domain `subscription.Subscription` didefinisikan.
- [x] GORM model map ke tabel `subscriptions` dari migration `000009`.

**Verification:**
- [x] Tests pass: `go test ./internal/adapter/postgres/... -run Subscription`
- [x] Build succeeds: `go build ./...`

**Dependencies:** Task 8
**Files likely touched:** `internal/adapter/postgres/subscription_repository.go`, `internal/domain/subscription/subscription.go`, `internal/port/subscription_repository.go`, `internal/adapter/postgres/subscription_repository_test.go`
**Estimated scope:** Medium

---

#### Task 10: Invoice repository

**Description:** Implementasi repository untuk tabel `invoices` + `invoice_items`.

**Acceptance criteria:**
- [x] `NewInvoiceRepository(db) port.InvoiceRepository` dengan signature benar: `FindByID(ctx, id string) (billing.Invoice, error)`.
- [x] Method: CRUD + `FindByCustomer`.
- [x] Domain `billing.Invoice` dan `billing.InvoiceItem` didefinisikan.
- [x] GORM model map ke tabel `invoices` dan `invoice_items` dari migration `000014`.

**Verification:**
- [x] Tests pass: `go test ./internal/adapter/postgres/... -run Invoice`
- [x] Build succeeds: `go build ./...`

**Dependencies:** Task 9
**Files likely touched:** `internal/adapter/postgres/invoice_repository.go`, `internal/domain/billing/invoice.go`, `internal/domain/billing/invoice_item.go`, `internal/port/invoice_repository.go`, `internal/adapter/postgres/invoice_repository_test.go`
**Estimated scope:** Medium

---

### Checkpoint: Persistence Layer

- [x] Semua repository test pass: `go test ./internal/adapter/postgres/...`
- [x] Vault test pass: `go test ./internal/adapter/vault/...`
- [x] Tidak ada `return nil` stub di persistence/vault.
- [x] Review dengan human sebelum lanjut ke auth.

---

### Phase 3: Authentication & Authorization

#### Task 11: JWT issuer/validator

**Description:** Implementasi `internal/adapter/auth/jwt.go` untuk sign dan verify JWT.

**Acceptance criteria:**
- [x] `NewJWT(secret []byte, expiry time.Duration)`.
- [x] `Issue(ctx, userID, username, role) (string, error)`.
- [x] `Validate(ctx, tokenString) (Claims, error)`.
- [x] Claims memuat `user_id`, `username`, `role`, `exp`.
- [x] Gunakan `golang-jwt/jwt/v5`.

**Verification:**
- [x] Tests pass: `go test ./internal/adapter/auth/... -run JWT`
- [x] Build succeeds: `go build ./...`

**Dependencies:** Task 6
**Files likely touched:** `internal/adapter/auth/jwt.go`, `internal/adapter/auth/jwt_test.go`
**Estimated scope:** Small

---

#### Task 12: Casbin RBAC enforcer

**Description:** Implementasi `internal/adapter/auth/casbin.go` dengan model RBAC sederhana.

**Acceptance criteria:**
- [x] `NewRBAC(modelPath, policyPath string)`.
- [x] Model RBAC dasar: `configs/rbac_model.conf`.
- [x] Policy CSV: `configs/rbac_policy.csv` mapping role → permission (method + path).
- [x] Method `Enforce(ctx, role, resource, action) (bool, error)`.
- [x] Contoh policy: `superadmin` bisa semua, `teknisi` hanya device read + run_command, `staff` hanya customer/subscription read.

**Verification:**
- [x] Tests pass: `go test ./internal/adapter/auth/... -run RBAC`
- [x] Build succeeds: `go build ./...`

**Dependencies:** Task 11
**Files likely touched:** `internal/adapter/auth/casbin.go`, `internal/adapter/auth/casbin_test.go`, `configs/rbac_model.conf`, `configs/rbac_policy.csv`
**Estimated scope:** Medium

---

#### Task 13: Auth + RBAC middleware

**Description:** Implementasi middleware Gin untuk validasi JWT dan otorisasi Casbin.

**Acceptance criteria:**
- [x] `internal/adapter/http/middleware/auth.go` — `AuthRequired(jwt)` mengekstrak token dari header `Authorization: Bearer <token>`, set `role` dan `user_id` di gin context.
- [x] `internal/adapter/http/middleware/rbac.go` — `RBACRequired(rbac)` mengecek `role, path, method` dengan Casbin; return 403 jika ditolak.
- [x] Middleware tidak crash jika token tidak ada (return 401).
- [x] Unit test dengan `httptest` + `gin.Test`.

**Verification:**
- [x] Tests pass: `go test ./internal/adapter/http/... -run Middleware`
- [x] Build succeeds: `go build ./...`

**Dependencies:** Task 12
**Files likely touched:** `internal/adapter/http/middleware/auth.go`, `internal/adapter/http/middleware/rbac.go`, `internal/adapter/http/middleware/middleware_test.go`
**Estimated scope:** Medium

---

### Phase 4: REST API

#### Task 14: Gin router + health check

**Description:** Implementasi `internal/adapter/http/router.go` sebagai entry point Gin.

**Acceptance criteria:**
- [ ] `NewRouter(cfg, deps)` mengembalikan `*gin.Engine` dengan middleware recovery, logger, dan CORS (opsional).
- [ ] Route `/health` return 200 dengan status db dan vault.
- [ ] Route `/api/v1/login` untuk admin login (public).
- [ ] Route group `/api/v1/` dengan `AuthRequired` + `RBACRequired`.
- [ ] `deps` berisi repository, usecase, jwt, rbac, vault.
- [ ] `internal/adapter/http/dto/` untuk request/response JSON (pisah dari domain).

**Verification:**
- [ ] Tests pass: `go test ./internal/adapter/http/... -run Router`
- [ ] Build succeeds: `go build ./...`
- [ ] Manual check: `curl http://localhost:8081/health`

**Dependencies:** Task 13
**Files likely touched:** `internal/adapter/http/router.go`, `internal/adapter/http/dto/common.go`, `internal/adapter/http/login_handler.go`, `internal/adapter/http/router_test.go`
**Estimated scope:** Medium

---

#### Task 15: Device REST handler

**Description:** Implementasi CRUD device di REST.

**Acceptance criteria:**
- [ ] `GET /api/v1/devices` — list devices.
- [ ] `GET /api/v1/devices/:id` — detail device.
- [ ] `POST /api/v1/devices` — create device + store credentials via vault.
- [ ] `PUT /api/v1/devices/:id` — update device.
- [ ] `DELETE /api/v1/devices/:id` — soft/hard delete (sesuai schema, hard delete dengan CASCADE ke credentials).
- [ ] Hanya role tertentu yang bisa create/update/delete.

**Verification:**
- [ ] Tests pass: `go test ./internal/adapter/http/... -run Device`
- [ ] Build succeeds: `go build ./...`
- [ ] Manual check: create device via curl, verify muncul di `GET /api/v1/devices`.

**Dependencies:** Task 14
**Files likely touched:** `internal/adapter/http/device_handler.go`, `internal/adapter/http/dto/device.go`, `internal/adapter/http/device_handler_test.go`
**Estimated scope:** Medium

---

#### Task 16: Customer REST handler

**Description:** Implementasi CRUD customer.

**Acceptance criteria:**
- [ ] `GET /api/v1/customers`, `GET /api/v1/customers/:id`, `POST /api/v1/customers`, `PUT /api/v1/customers/:id`, `DELETE /api/v1/customers/:id`.
- [ ] DTO validasi: `full_name`, `phone`, `address` wajib.
- [ ] Test handler dengan mock repository.

**Verification:**
- [ ] Tests pass: `go test ./internal/adapter/http/... -run Customer`
- [ ] Build succeeds: `go build ./...`

**Dependencies:** Task 15
**Files likely touched:** `internal/adapter/http/customer_handler.go`, `internal/adapter/http/dto/customer.go`, `internal/adapter/http/customer_handler_test.go`, `internal/usecase/business/manage_customer.go`
**Estimated scope:** Medium

---

#### Task 17: Subscription REST handler

**Description:** Implementasi CRUD subscription.

**Acceptance criteria:**
- [ ] Endpoint CRUD `/api/v1/subscriptions`.
- [ ] Validasi: `customer_id`, `plan_id`, `device_id`, `service_type` wajib.
- [ ] Business usecase `manage_subscription.go` diisi minimal CRUD.

**Verification:**
- [ ] Tests pass: `go test ./internal/adapter/http/... -run Subscription`
- [ ] Build succeeds: `go build ./...`

**Dependencies:** Task 16
**Files likely touched:** `internal/adapter/http/subscription_handler.go`, `internal/adapter/http/dto/subscription.go`, `internal/usecase/business/manage_subscription.go`, `internal/adapter/http/subscription_handler_test.go`
**Estimated scope:** Medium

---

#### Task 18: Plan REST handler

**Description:** Implementasi CRUD plan.

**Acceptance criteria:**
- [ ] Endpoint CRUD `/api/v1/plans`.
- [ ] Validasi: `name`, `service_type`, `price`, `bandwidth_down_kbps`, `bandwidth_up_kbps` wajib.
- [ ] Business usecase `manage_plan.go` diisi.

**Verification:**
- [ ] Tests pass: `go test ./internal/adapter/http/... -run Plan`
- [ ] Build succeeds: `go build ./...`

**Dependencies:** Task 17
**Files likely touched:** `internal/adapter/http/plan_handler.go`, `internal/adapter/http/dto/plan.go`, `internal/usecase/business/manage_plan.go`, `internal/adapter/http/plan_handler_test.go`
**Estimated scope:** Medium

---

#### Task 19: Invoice REST handler

**Description:** Implementasi read-only invoice (create invoice masuk fase billing automation).

**Acceptance criteria:**
- [ ] `GET /api/v1/invoices`, `GET /api/v1/invoices/:id`.
- [ ] Filter by `customer_id` query param.
- [ ] Business usecase `manage_invoice.go` diisi minimal read.

**Verification:**
- [ ] Tests pass: `go test ./internal/adapter/http/... -run Invoice`
- [ ] Build succeeds: `go build ./...`

**Dependencies:** Task 18
**Files likely touched:** `internal/adapter/http/invoice_handler.go`, `internal/adapter/http/dto/invoice.go`, `internal/usecase/business/manage_invoice.go`, `internal/adapter/http/invoice_handler_test.go`
**Estimated scope:** Medium

---

#### Task 20: Login handler

**Description:** Implementasi login admin/staff yang return JWT.

**Acceptance criteria:**
- [ ] `POST /api/v1/login` dengan username + password.
- [ ] Verify password dengan bcrypt.
- [ ] Return JWT + user profile.
- [ ] Update `last_login_at`.
- [ ] Rate limit sederhana (in-memory) opsional.

**Verification:**
- [ ] Tests pass: `go test ./internal/adapter/http/... -run Login`
- [ ] Build succeeds: `go build ./...`
- [ ] Manual check: login via curl dapat token.

**Dependencies:** Task 11
**Files likely touched:** `internal/adapter/http/login_handler.go`, `internal/adapter/http/login_handler_test.go`
**Estimated scope:** Small

---

### Checkpoint: REST API

- [ ] `go test ./internal/adapter/http/...` pass.
- [ ] Semua CRUD endpoint bisa diuji dengan `curl` (manual smoke test).
- [ ] Auth/RBAC middleware mengembalikan 401/403 sesuai harapan.
- [ ] Review dengan human sebelum wire ke main.

---

### Phase 5: Wire & Smoke Test

#### Task 21: Wire real adapters into main.go

**Description:** Ganti `loadDemoDevices()` di `cmd/server/main.go` dengan inisialisasi PostgreSQL, vault, user repository, JWT, RBAC, dan router.

**Acceptance criteria:**
- [ ] `main.go` memanggil `config.Load(ctx)`.
- [ ] `NewDB` dibuka, repository device/customer/subscription/plan/invoice dibuat dengan DB nyata.
- [ ] Vault dibuat dengan key dari config.
- [ ] Registry mengambil device dari `postgres.DeviceRepository` dan `vault.CredentialVault`.
- [ ] Jika `POLYGLOT_DEMO_*` di-set, device demo di-create via repository (bukan in-memory) saat startup.
- [ ] Gin HTTP server di-start pada `RESTAddr` (default `:8081`).
- [ ] MCP server tetap di-start pada `MCP_HTTP_ADDR` (default `:8080`).
- [ ] Graceful shutdown menutup kedua server + DB + registry.

**Verification:**
- [ ] Build succeeds: `go build ./...`
- [ ] Lint clean: `make lint`
- [ ] Manual check: `make run` tidak crash, `/health` respons.

**Dependencies:** Tasks 14-21
**Files likely touched:** `cmd/server/main.go`
**Estimated scope:** Medium

---

#### Task 22: End-to-end smoke test

**Description:** Dokumentasikan dan jalankan smoke test manual dari nol.

**Acceptance criteria:**
- [ ] `make db-up && make migrate-up` berhasil.
- [ ] `make run` berhasil dan tidak error.
- [ ] `curl http://localhost:8081/health` return 200.
- [ ] Login dapat token JWT.
- [ ] Create device via REST, lalu `GET /api/v1/devices/:id` return data.
- [ ] MCP tool `get_device_status` untuk device yang baru dibuat berhasil (jika device reachable).
- [ ] Update `README.md` dengan instruksi run terbaru.

**Verification:**
- [ ] Manual smoke test script atau curl commands tercatat di `docs/smoke-test-foundation.md`.
- [ ] CI masih pass: `make ci` (atau `go test ./...`, `make lint`).

**Dependencies:** Task 21
**Files likely touched:** `README.md`, `docs/smoke-test-foundation.md`
**Estimated scope:** Medium

---

### Checkpoint: Foundation Complete

- [ ] `make run` menjalankan satu binary dengan MCP + REST + Postgres + vault + auth.
- [ ] Semua test pass: `go test ./...`.
- [ ] Lint pass: `make lint`.
- [ ] Dokumentasi run terbaru.
- [ ] Review dengan human sebelum lanjut fase berikutnya (provisioning/customer portal).

---

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Migration GORM model tidak cocok dengan schema existing | High | Buat GORM model field-by-field berdasarkan migration, lalu integration test dengan testcontainers. |
| Casbin v3 masih alpha/unstable | Medium | Pin ke versi stable yang tercatat di `TECH-STACK-DAN-PERSIAPAN.md`; siap fallback ke policy in-memory sederhana. |
| JWT secret / vault key tidak aman di `.env` | High | `.env` sudah gitignored; jangan pernah commit key nyata; dokumentasikan cara generate. |
| GORM v2 tidak cocok dengan `pgx` driver tertentu | Low | Gunakan `gorm.io/driver/postgres` v5+ yang built on pgx. |
| Dual server (MCP + REST) di satu binary membuat shutdown rumit | Medium | Gunakan `errgroup` atau `sync.WaitGroup` dengan graceful shutdown terpisah per server. |
| Port interface `FindByID` masih salah signature | Medium | Perbaiki di Task 7 sebelum repository diimplementasikan. |

## Open Questions

1. **Apakah Casbin policy harus di database atau file CSV?** Untuk fase ini file CSV cukup; DB policy bisa fase berikutnya.
2. **Port REST default berapa?** Saat ini MCP default `:8080`. REST disarankan `:8081` agar tidak tabrakan.
3. **Apakah perlu seed user admin default?** Sangat disarankan — nanti tidak bisa login jika tabel `users` kosong. Tambahkan seed CLI atau migration seed.
4. **Bagaimana menangani device credentials saat update?** Vault `Store` perlu ada; saat create device sekalian insert credentials, saat update pilih update credential atau tidak.
5. **Apakah customer portal masuk fase ini?** Tidak. Fase ini hanya admin REST. Customer portal masuk fase berikutnya.
