# Issue 02: Plan Router Profile Sync

## Konteks

Satu `plan` (paket layanan, mis. "Home 10Mbps") di Polyglot bersifat **abstrak** — ia hanya mendeskripsikan bandwidth dan burst yang dijual, bukan objek nyata di perangkat. Supaya PPPoE/Hotspot benar-benar membatasi kecepatan pelanggan, paket itu harus diwujudkan sebagai profil `/ppp profile` (dengan `rate-limit`) yang **benar-benar ada di setiap router MikroTik** tempat paket dipakai. `DATABASE-SCHEMA.md` §4.3 (`plan_router_profiles`, migrasi 000006) adalah titik sinkron ini: satu baris per pasangan (plan, device), memisahkan "apa yang dijual" dari "sudah di-provision di router mana saja". Tanpa profil ini, Issue 03 (provisioning `/ppp secret`) tidak bisa menautkan pelanggan ke paket, karena `/ppp secret` mereferensi nama profil yang harus sudah ada di router.

`ANALISIS-PROVISIONING-REPO-REFERENSI.md` menunjukkan bahwa keempat repo referensi memperlakukan profil MikroTik sebagai katalog yang di-setup lebih dulu, lalu dipakai berulang oleh banyak pelanggan — bukan dibuat ulang per pelanggan. Issue ini membangun jalur CRUD + sinkronisasi untuk katalog itu: mengaitkan plan ke device, menamai profil, dan mendorong `/ppp profile add|set` ke router lewat pola sinkronisasi kanonik (K4) sehingga tetap tercatat di `command_audit_log`.

Ada satu ketegangan desain yang harus diselesaikan issue ini secara eksplisit: pola sinkronisasi kanonik memakai `provisioning_sync_log` (§6.3), tetapi tabel itu dirancang **per-subscription** (`subscription_id NOT NULL`, `target_type` belum punya nilai untuk profil `/ppp profile`). Profil paket adalah **setup katalog**, bukan aksi per-pelanggan. Keputusan penanganannya ada di bagian Migrasi Database & Task 4 di bawah.

## Prasyarat

- **Issue 01 (Provisioning Sync Engine)** — WAJIB selesai. Issue ini menulis baris `provisioning_sync_log` `pending` dan bergantung pada Sync Engine untuk menerjemahkannya jadi `command.Command`, memanggil `usecase/network.ExecuteCommand`, dan menautkan hasil ke `command_audit_log`. Tanpa Sync Engine, baris `pending` tidak akan pernah diproses.
- **Foundation Phase 4–5** (dari `docs/plan-foundation-first.md`): router Gin nyata dengan middleware `AuthRequired` + `RBACRequired`, `internal/adapter/http/dto/`, dan CRUD `plans` + `devices` sudah jalan.
- **Migrasi 000005 (`plans`)**, **000006 (`plan_router_profiles`)**, **000011 (`provisioning_sync_log`)** sudah ada. Domain `subscription.Subscription` dan `plan.Plan` sudah ada.
- **Driver MikroTik** (`internal/driver/mikrotik/`) sudah ada: `driver.go`, `connect.go` (`dialAndLogin`), `commands.go` (`operationMap`, `Classify`, `Translate`).
- **Registry** (`internal/registry/registry.go`) sudah memegang satu `port.DeviceDriver` per device.

## Ruang Lingkup

**In scope:**
- Domain `PlanRouterProfile` (entity + validasi + error sentinel).
- Kontrak `port.PlanRouterProfileRepository` + implementasi Postgres (memetakan tabel migrasi 000006 yang sudah ada).
- Usecase business (CRUD asosiasi plan↔device↔nama profil) dan usecase network (memicu sinkron ke router lewat `provisioning_sync_log`).
- Perluasan katalog command MikroTik untuk operasi `/ppp/profile/add` dan `/ppp/profile/set` di `internal/driver/mikrotik/commands.go`.
- Migrasi 000023: menambah nilai `target_type='mikrotik_ppp_profile'` pada CHECK `provisioning_sync_log` dan melonggarkan `subscription_id` menjadi nullable (lihat Migrasi Database).
- 4 endpoint REST + baris RBAC.

**Out of scope:**
- Provisioning `/ppp secret` per pelanggan (Issue 03).
- Sinkron Hotspot `/ip hotspot user profile` (paket Hotspot/voucher — issue terpisah, walau polanya identik dan bisa mengikuti issue ini kelak).
- Penghapusan profil nyata di router saat plan dihapus di sisi bisnis (cascade delete perangkat) — issue ini hanya menghapus **asosiasi**; opsi kirim `/ppp/profile/remove` dicatat sebagai enhancement, tidak wajib.
- Perhitungan `rate-limit`/burst final MikroTik yang rumit di luar pemetaan langsung dari kolom `plans` (lihat catatan Task 5).

## REST API

Semua endpoint di bawah `/api/v1/`. Aksi yang menyentuh perangkat mengembalikan **202 Accepted** + id `sync_log` (K4/K5). Format error mengikuti foundation: `{ "error": { "code": "...", "message": "..." } }`.

| Method | Path | Tujuan | Role minimum |
|---|---|---|---|
| GET | `/api/v1/plans/:id/router-profiles` | List semua asosiasi profil router untuk satu plan | staff |
| POST | `/api/v1/plans/:id/router-profiles` | Kaitkan plan ke satu device + nama profil (`plan_router_profiles` baris baru, belum sinkron) | admin |
| POST | `/api/v1/plans/:id/router-profiles/:profileId/sync` | Picu sinkron profil ke router MikroTik (tulis `provisioning_sync_log` `pending`) | admin |
| DELETE | `/api/v1/plans/:id/router-profiles/:profileId` | Hapus asosiasi profil (tidak menghapus profil nyata di router) | admin |

### GET `/api/v1/plans/:id/router-profiles`
- **Request:** path param `id` (plan UUID). Tidak ada body. Query opsional `device_id` untuk filter.
- **Response 200:** array objek profil, tiap objek berisi: `id`, `plan_id`, `device_id`, `mikrotik_profile_name`, `sync_status` (`pending`/`synced`/`error`), `last_synced_at` (nullable), `sync_error_message` (nullable). Bila plan tidak punya profil, kembalikan array kosong `[]`, bukan 404.
- **Gagal:** 404 bila plan `:id` tidak ada; 403 bila role di bawah staff; 401 bila token tidak valid.

### POST `/api/v1/plans/:id/router-profiles`
- **Request (field penting):** `device_id` (UUID device MikroTik target, wajib), `mikrotik_profile_name` (string nama `/ppp profile`, wajib; bila kosong, usecase boleh menurunkan default dari nama plan — lihat Task 3). 
- **Response 201:** objek profil yang baru dibuat (bentuk sama dengan GET), `sync_status='pending'`, `last_synced_at=null`. Membuat baris di sini **tidak** langsung memicu sinkron ke router — hanya mencatat asosiasi. Sinkron dilakukan lewat endpoint `/sync`.
- **Gagal:** 400 bila `device_id`/`mikrotik_profile_name` kosong atau `device_id` bukan device bertipe MikroTik; 404 bila plan/device tidak ada; **409** bila pasangan (plan, device) sudah ada (melanggar `UNIQUE(plan_id, device_id)`); 403/401 sesuai RBAC/auth.

### POST `/api/v1/plans/:id/router-profiles/:profileId/sync`
- **Request:** path param `id` (plan) + `profileId`. Tidak ada body wajib; body opsional `action` (`create` untuk `/ppp/profile/add`, `update` untuk `/ppp/profile/set`) — default: `create` bila `sync_status` masih `pending` & belum pernah `synced`, `update` bila sudah pernah `synced`.
- **Response 202 Accepted:** `{ "sync_log_id": "<uuid>", "status": "pending" }`. Handler → usecase network menulis satu baris `provisioning_sync_log` (`target_type='mikrotik_ppp_profile'`, `action` sesuai, `device_id` = device profil, `status='pending'`). Sync Engine (Issue 01) yang mengeksekusi ke router; hasil dipolling lewat `GET /api/v1/sync-logs/:id` (Issue 01).
- **Gagal:** 404 bila plan/profil tidak ada atau `profileId` bukan milik plan `:id`; 409 bila sudah ada baris sync `pending` untuk profil yang sama (hindari duplikasi in-flight); 403/401 sesuai RBAC/auth. Kegagalan perangkat sendiri **tidak** muncul di sini (asinkron) — statusnya jadi `error`/`failed` di `sync_log` belakangan.

### DELETE `/api/v1/plans/:id/router-profiles/:profileId`
- **Request:** path param `id` + `profileId`. Tidak ada body.
- **Response 204 No Content** (atau 200 dengan objek terhapus). Menghapus baris `plan_router_profiles`. **Tidak** mengirim command ke router (out of scope) — dokumentasikan di response/godoc bahwa profil `/ppp profile` di router tidak ikut terhapus.
- **Gagal:** 404 bila plan/profil tidak ada atau mismatch; 409 bila masih ada subscription aktif yang memakai profil ini (bila keterkaitan itu bisa dideteksi — kalau tidak, lewati cek ini dan catat sebagai known limitation); 403/401 sesuai RBAC/auth.

## Tasks

**Task 1: Domain `PlanRouterProfile`**

**Description:** Definisikan entity katalog profil router sebagai domain murni (tanpa I/O, tanpa import library eksternal) beserta validasi dan error sentinel.

**Acceptance criteria:**
- [ ] Tipe `PlanRouterProfile` punya field: `ID`, `PlanID`, `DeviceID`, `MikrotikProfileName`, `SyncStatus`, `LastSyncedAt` (pointer/nullable time), `SyncErrorMessage`.
- [ ] `SyncStatus` direpresentasikan sebagai tipe tertutup (mis. konstanta `SyncStatusPending`/`SyncStatusSynced`/`SyncStatusError`) — bukan string bebas — dengan doc comment per konstanta.
- [ ] Ada konstruktor `New(...)` (pakai struct params bila argumen >4) yang memvalidasi `PlanID`/`DeviceID`/`MikrotikProfileName` tidak kosong dan mengembalikan `SyncStatusPending` sebagai zero-state.
- [ ] Error sentinel di `errors.go`: mis. `ErrProfileNameRequired`, `ErrProfileNotFound`, `ErrProfileAlreadyExists`, masing-masing dengan doc comment dimulai nama identifier.
- [ ] Tidak ada import ke `adapter`/`driver`/framework (boundary §0).

**Files likely touched:** `internal/domain/plan/router_profile.go`, `internal/domain/plan/errors.go` (perluas bila sudah ada).

**Dependencies:** —

**Estimated scope:** Small

---

**Task 2: Port `PlanRouterProfileRepository` + implementasi Postgres**

**Description:** Definisikan kontrak repository di `internal/port/` dan implementasikan di `internal/adapter/postgres/` yang memetakan tabel `plan_router_profiles` (migrasi 000006 yang sudah ada — tidak buat migrasi baru untuk tabel ini).

**Acceptance criteria:**
- [ ] Interface `PlanRouterProfileRepository` di `internal/port/plan_router_profile_repository.go` dengan method minimal: `Create(ctx, profile) error`, `FindByID(ctx, id) (PlanRouterProfile, error)`, `ListByPlan(ctx, planID) ([]PlanRouterProfile, error)`, `FindByPlanAndDevice(ctx, planID, deviceID) (PlanRouterProfile, error)`, `UpdateSyncState(ctx, id, status, syncedAt, errMsg) error`, `Delete(ctx, id) error`. Semua method `ctx` parameter pertama, `error` return terakhir.
- [ ] Doc comment pada interface & tiap method dimulai dengan nama identifier.
- [ ] Implementasi Postgres di `internal/adapter/postgres/plan_router_profile_repository.go`, model GORM di `models.go` memetakan kolom `id, plan_id, device_id, mikrotik_profile_name, sync_status, last_synced_at, sync_error_message, created_at`.
- [ ] `FindByPlanAndDevice`/`FindByID` memetakan `sql.ErrNoRows` (atau `gorm.ErrRecordNotFound`) → `ErrProfileNotFound` via `errors.Is`.
- [ ] Pelanggaran `UNIQUE(plan_id, device_id)` dipetakan → `ErrProfileAlreadyExists` (cek kode error unik Postgres via `errors.As` pada `*pgconn.PgError`, bukan string-match).
- [ ] Compile-time assertion `var _ port.PlanRouterProfileRepository = (*...)(nil)`.

**Files likely touched:** `internal/port/plan_router_profile_repository.go`, `internal/adapter/postgres/plan_router_profile_repository.go`, `internal/adapter/postgres/models.go`.

**Dependencies:** Task 1.

**Estimated scope:** Medium

---

**Task 3: Usecase business — CRUD asosiasi profil**

**Description:** Orkestrasi CRUD katalog: membuat asosiasi (plan↔device↔nama profil), list per plan, dan hapus asosiasi. Tidak menyentuh driver perangkat (itu Task 4).

**Acceptance criteria:**
- [ ] Fungsi/usecase di `internal/usecase/business/manage_plan_router_profile.go`: `CreatePlanRouterProfile`, `ListPlanRouterProfiles`, `DeletePlanRouterProfile`.
- [ ] `CreatePlanRouterProfile` memvalidasi plan & device ada (lewat repo plan & device), device bertipe MikroTik; bila `mikrotik_profile_name` kosong, turunkan default deterministik dari nama/slug plan (mis. `plan-<slug>`), dan dokumentasikan aturan default itu di godoc.
- [ ] Guard clause / early return dipakai (tidak ada nesting berlapis; tidak ada `else` setelah `return`).
- [ ] Error domain dipropagasi apa adanya (`ErrProfileAlreadyExists`, `ErrProfileNotFound`) dengan wrap `%w` + konteks operasi.
- [ ] Table-driven unit test mencakup: sukses, nama kosong→default, plan tidak ada, device bukan MikroTik, duplikasi (`UNIQUE`).

**Files likely touched:** `internal/usecase/business/manage_plan_router_profile.go` (+ `_test.go`).

**Dependencies:** Task 1, Task 2.

**Estimated scope:** Medium

---

**Task 4: Usecase network — picu sinkron via `provisioning_sync_log`**

**Description:** Implementasi endpoint `/sync`: usecase network menulis satu baris `provisioning_sync_log` `pending` dengan `target_type='mikrotik_ppp_profile'`, mengikuti pola kanonik K4 (bukan memanggil driver langsung). Menyelesaikan ketegangan desain "profil bukan per-subscription".

**Acceptance criteria:**
- [ ] Fungsi di `internal/usecase/network/sync_plan_router_profile.go`: `SyncPlanRouterProfile(ctx, planID, profileID, action)`.
- [ ] Usecase memuat profil (Task 2), memvalidasi `profileId` milik `planId`, menolak bila sudah ada baris `sync_log` `pending` untuk profil yang sama (kembalikan error yang dipetakan handler ke 409).
- [ ] Menulis SATU baris `provisioning_sync_log`: `target_type='mikrotik_ppp_profile'`, `action` = `create`/`update` (dipetakan dari state `sync_status`), `device_id` = device profil, `subscription_id = NULL` (lihat Migrasi), `status='pending'`, `external_reference` = id `plan_router_profiles` (supaya Sync Engine bisa menautkan balik ke katalog profil).
- [ ] Mengembalikan id `sync_log` yang baru dibuat (untuk response 202).
- [ ] **Tidak** memanggil `port.DeviceDriver` sama sekali di usecase ini (K4) — Sync Engine yang mengeksekusi.
- [ ] Sync Engine (Issue 01) mampu memetakan `target_type='mikrotik_ppp_profile'` + `external_reference` menjadi operasi `Translate` MikroTik yang benar — koordinasikan kontrak ini dengan Issue 01 (dokumentasikan mapping di godoc usecase ini). Bila Issue 01 memakai `subscription_id` untuk resolusi device, pastikan jalur `subscription_id NULL` + `external_reference` ditangani.
- [ ] Table-driven unit test: sukses (create), sukses (update setelah synced), profil tidak ada, mismatch plan↔profil, duplikasi pending→409-mappable.

**Files likely touched:** `internal/usecase/network/sync_plan_router_profile.go` (+ `_test.go`), koordinasi dengan `internal/usecase/network/execute_command.go` & kode Sync Engine Issue 01.

**Dependencies:** Task 2, Issue 01, migrasi 000023 (Migrasi Database).

**Estimated scope:** Large

---

**Task 5: Perluas katalog command MikroTik untuk `/ppp profile`**

**Description:** Tambah pengetahuan operasi profil `/ppp/profile/add` dan `/ppp/profile/set` ke driver MikroTik — pemetaan bandwidth/burst `plans` → `rate-limit` ada DI SINI, bukan di usecase (AGENTS.md §1.2).

**Acceptance criteria:**
- [ ] `internal/driver/mikrotik/commands.go` diperluas: `operationMap`/katalog menambah operasi profil (mis. `OpCreatePPPProfile`, `OpUpdatePPPProfile`) → command native `/ppp/profile/add` & `/ppp/profile/set` dengan `Args` (`name`, `rate-limit`).
- [ ] Nilai `rate-limit` dibentuk dari kolom `plans`: `bandwidth_up_kbps`/`bandwidth_down_kbps` (rx/tx) plus burst `burst_up`/`burst_down`/`burst_threshold_*`/`burst_time_*` sesuai format RouterOS (`rx-rate/tx-rate [rx-burst-rate/tx-burst-rate rx-burst-threshold/tx-burst-threshold rx-burst-time/tx-burst-time]`). Format string RouterOS dikomentari singkat (bila butuh >5 baris penjelasan, buat ADR — lihat catatan).
- [ ] `Classify` untuk operasi profil: `/ppp/profile/add` & `/ppp/profile/set` diklasifikasikan `ClassReadOnly` **hanya jika** kebijakan menganggapnya tidak destruktif — namun karena mengubah rate-limit pelanggan berdampak layanan, klasifikasikan sebagai `ClassDestructive` (butuh HITL) kecuali ada keputusan eksplisit sebaliknya; nyatakan keputusan ini sebagai `// DEVIASI` bila menyimpang, atau ADR bila argumennya panjang.
- [ ] Enum operasi baru (`command.Operation`) ditambahkan di `internal/domain/command/` bila belum ada, dengan doc comment.
- [ ] Unit test `commands_test.go` untuk `Translate` (operasi profil → command + args benar) dan `Classify` (klasifikasi sesuai keputusan).
- [ ] Tidak ada pengetahuan RouterOS yang bocor ke `usecase/` atau `domain/`.

**Files likely touched:** `internal/driver/mikrotik/commands.go` (+ `commands_test.go`), `internal/domain/command/command.go` (enum operasi baru bila perlu).

**Dependencies:** Task 4 (kontrak operasi), grounding kolom `plans` §4.2.

**Estimated scope:** Medium

---

**Task 6: Handler REST + DTO + RBAC**

**Description:** Empat endpoint di `internal/adapter/http/`, DTO request/response di `internal/adapter/http/dto/`, dan baris RBAC di `configs/rbac_policy.csv`.

**Acceptance criteria:**
- [ ] Handler di `internal/adapter/http/plan_router_profile_handler.go`: list (200), create (201), sync (202 + `sync_log_id`), delete (204). Handler **tidak** memanggil driver — hanya usecase (K4).
- [ ] DTO di `internal/adapter/http/dto/`: request create (`device_id`, `mikrotik_profile_name`), request sync (opsional `action`), response profil, response sync (`sync_log_id`, `status`).
- [ ] Route didaftarkan di `internal/adapter/http/router.go` di bawah group `/api/v1` ber-middleware `AuthRequired` + `RBACRequired`.
- [ ] Pemetaan error → status: `ErrProfileNotFound`/plan-not-found→404, validasi→400, `ErrProfileAlreadyExists`→409, sync pending duplikat→409, RBAC→403.
- [ ] Baris RBAC ditambahkan ke `configs/rbac_policy.csv`: GET → staff+; POST create/sync & DELETE → admin+ (superadmin/owner/admin). Konsisten dengan matrix K3.
- [ ] Handler test dengan `httptest` mencakup tiap status code sukses & minimal 404/409/403.

**Files likely touched:** `internal/adapter/http/plan_router_profile_handler.go`, `internal/adapter/http/dto/plan_router_profile.go`, `internal/adapter/http/router.go`, `configs/rbac_policy.csv` (+ handler `_test.go`).

**Dependencies:** Task 3, Task 4.

**Estimated scope:** Medium

---

**Task 7: Wiring di `main.go`**

**Description:** Sambungkan repository, usecase, dan handler baru ke composition root.

**Acceptance criteria:**
- [ ] `cmd/server/main.go` menginstansiasi `PlanRouterProfileRepository` (Postgres), usecase business & network, dan mendaftarkan handler ke router.
- [ ] `go build ./...` sukses; server menyala tanpa panic; endpoint muncul di route list.
- [ ] Tidak ada dependensi yang di-`panic` di luar startup init (AGENTS.md §4).

**Files likely touched:** `cmd/server/main.go`.

**Dependencies:** Task 2, Task 3, Task 4, Task 6.

**Estimated scope:** Small

## Migrasi Database

Diperlukan **satu** migrasi baru: **000023**, berpasangan up/down.

- **File:** `migrations/000023_add_ppp_profile_target_type.up.sql` + `migrations/000023_add_ppp_profile_target_type.down.sql`.
- **Perubahan (dijelaskan sebagai teks, bukan SQL):**
  1. **Perluas CHECK `provisioning_sync_log.target_type`** untuk menambah nilai baru `mikrotik_ppp_profile` di samping lima nilai yang sudah ada (`mikrotik_ppp_secret`, `mikrotik_hotspot_user`, `mikrotik_address_list`, `freeradius`, `genieacs_tr069`). Ini berarti drop constraint CHECK lama lalu buat ulang dengan enam nilai (nama constraint harus stabil agar `down` bisa mengembalikannya persis ke lima nilai).
  2. **Longgarkan `provisioning_sync_log.subscription_id` menjadi nullable** (drop `NOT NULL`). **Alasan wajib dicatat di komentar migrasi:** sinkron profil `/ppp profile` adalah setup **katalog** (per pasangan plan↔device), bukan aksi per-pelanggan, sehingga tidak selalu punya `subscription_id`. FK ke `subscriptions` tetap dipertahankan (baris profil menyimpan `NULL`, bukan referensi menggantung). Konsekuensinya: kode Sync Engine (Issue 01) dan query mana pun yang mengasumsikan `subscription_id` selalu terisi harus menangani `NULL` — koordinasikan dengan Issue 01.
- **Down migration:** kembalikan CHECK ke lima nilai semula **dan** kembalikan `subscription_id` ke `NOT NULL`. Catatan: `down` akan gagal bila sudah ada baris `mikrotik_ppp_profile` atau baris `subscription_id NULL` — ini perilaku yang diterima (down hanya aman sebelum data profil ditulis); dokumentasikan di komentar file `.down.sql`.

**Keputusan desain (opsi a vs b):** Spesifikasi menawarkan dua jalur — (a) tambah `target_type='mikrotik_ppp_profile'` lewat migrasi ALTER CHECK dan tetap lewat `provisioning_sync_log`, atau (b) lewati `sync_log` dan panggil `ExecuteCommand` langsung karena ini setup katalog. **Issue ini memilih (a)** demi konsistensi audit: dengan (a), setiap perubahan profil router ikut tercatat di `command_audit_log` lewat jalur yang sama seperti semua provisioning lain (satu tabel audit, satu pola — DATABASE-SCHEMA §7). Kelemahan (a) — yaitu `subscription_id` yang secara semantik tidak pas — diselesaikan dengan melonggarkan kolom itu jadi nullable, bukan dengan membuat jalur kedua yang lepas dari audit. Opsi (b) ditolak karena memecah pola K4 dan meninggalkan aksi ke perangkat tanpa jejak `provisioning_sync_log`.

**Cermin ke `DATABASE-SCHEMA.md` pada PR yang sama (K6):**
- **§6.3** — perbarui daftar `target_type` menjadi enam nilai (tambah `mikrotik_ppp_profile`) dan tandai `subscription_id` sebagai nullable dengan catatan alasannya (katalog vs per-subscription).
- **§4.3** — tambah catatan bahwa `plan_router_profiles.sync_status`/`last_synced_at`/`sync_error_message` diisi oleh jalur sinkron Issue 02, dan bahwa baris `provisioning_sync_log` dengan `target_type='mikrotik_ppp_profile'` menautkan balik ke `plan_router_profiles.id` lewat `external_reference`.

## Verification

- [ ] `go build ./...` sukses tanpa error.
- [ ] `go vet ./...` dan `make lint` (golangci) bersih.
- [ ] `gofumpt`/`goimports` tidak menghasilkan diff.
- [ ] `go test ./internal/domain/plan/... ./internal/usecase/business/... ./internal/usecase/network/... ./internal/driver/mikrotik/... ./internal/adapter/http/...` hijau (unit + handler `httptest`).
- [ ] Test repository Postgres dengan `testcontainers-go` (`go test ./internal/adapter/postgres/...`) hijau, termasuk skenario `UNIQUE(plan_id, device_id)` → `ErrProfileAlreadyExists`.
- [ ] Migrasi up/down 000023 dijalankan bolak-balik di DB test bersih tanpa error (up pada skema kosong; down sebelum ada data profil).
- [ ] Compile-time assertion `var _ port.PlanRouterProfileRepository = (*...)(nil)` ada dan lolos build.
- [ ] Smoke test manual (server lokal + token admin), sebutkan sebagai langkah `curl`:
  - `POST /api/v1/plans/:id/router-profiles` dengan body `device_id` + `mikrotik_profile_name` → harapkan 201 + objek profil `sync_status=pending`.
  - `POST /api/v1/plans/:id/router-profiles/:profileId/sync` → harapkan 202 + `sync_log_id`.
  - `GET /api/v1/sync-logs/:sync_log_id` (Issue 01) → verifikasi baris `provisioning_sync_log` `target_type=mikrotik_ppp_profile` muncul dan diproses Sync Engine.
  - `GET /api/v1/plans/:id/router-profiles` dengan token `staff` → 200; dengan `POST` token `staff` → 403.
  - `DELETE /api/v1/plans/:id/router-profiles/:profileId` token admin → 204.

## Definition of Done

- [ ] Domain `PlanRouterProfile` + error sentinel selesai, tanpa pelanggaran boundary.
- [ ] Port + implementasi Postgres selesai, memetakan migrasi 000006, dengan compile-time assertion.
- [ ] Usecase business (CRUD) dan usecase network (`/sync` via `provisioning_sync_log`) selesai; handler tidak pernah memanggil driver langsung (K4).
- [ ] Katalog command MikroTik `/ppp profile add|set` + pemetaan bandwidth/burst ada di `internal/driver/mikrotik/commands.go`, bukan di usecase.
- [ ] Migrasi 000023 (ALTER CHECK + nullable `subscription_id`) up/down berpasangan, dicerminkan ke `DATABASE-SCHEMA.md` §6.3 & §4.3 pada PR yang sama.
- [ ] 4 endpoint REST + DTO + baris RBAC di `configs/rbac_policy.csv` sesuai matrix K3; 202/201/204 sesuai konvensi.
- [ ] Semua item Verification lolos; satu issue = satu PR.
- [ ] Keputusan opsi (a) dan konsekuensi `subscription_id` nullable terdokumentasi (komentar migrasi + godoc usecase), koordinasi dengan Issue 01 dikonfirmasi.
