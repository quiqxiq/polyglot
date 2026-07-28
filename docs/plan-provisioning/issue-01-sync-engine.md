# Issue 01: Provisioning Sync Engine

## Konteks

Semua alur provisioning di Polyglot — mulai dari aktivasi PPPoE secret di MikroTik, pembuatan hotspot user, sampai push konfigurasi TR-069 ke GenieACS — bergantung pada satu pola tunggal yang sudah ditetapkan sebagai konvensi bersama (K4): handler REST/usecase bisnis **tidak pernah** memanggil `port.DeviceDriver` secara langsung, melainkan menulis satu atau lebih baris di tabel `provisioning_sync_log` berstatus `pending`. Sebuah mesin sinkronisasi (Sync Engine) kemudian membaca baris-baris tersebut, menerjemahkannya menjadi `command.Command` yang sesuai untuk `target_type`-nya, mengeksekusinya lewat `usecase/network.ExecuteCommand`, dan menautkan hasilnya kembali ke `command_audit_log` sambil meng-update status `sync_log` menjadi `success`/`failed`. Selama mesin ini belum ada, seluruh temuan provisioning di `ANALISIS-PROVISIONING-REPO-REFERENSI.md` menggantung — tabel `provisioning_sync_log` (migrasi 000011) sudah dibuat, tetapi tidak ada satu pun proses yang mengonsumsinya. Pola ini bukan sekadar teori: repo referensi `billing-rtrw` memakai tabel `acs_tasks(device_id, name, payload, status)` yang bergerak `pending → done` lewat worker terpisah — validasi lapangan bahwa arsitektur `provisioning_sync_log(pending) → execute → audit` di Polyglot adalah pola yang sudah terbukti di produksi (lihat REFERENCES.md).

Issue ini membangun inti mesin tersebut: domain baru `internal/domain/provisioning/` yang konstantanya mencerminkan persis CHECK constraint DB (`DATABASE-SCHEMA.md §6.3`), kontrak `port.ProvisioningSyncRepository`, implementasi Postgres-nya, usecase `ProcessPendingSyncs`, serta runner yang memicunya. Penerjemahan `target_type`+`action` menjadi command native **tetap** menjadi tanggung jawab driver vendor via `Translate`/katalog (K1) — usecase hanya mengorkestrasi, tidak mengetahui detail RouterOS/GenieACS.

Tanpa Issue 01 tidak ada issue provisioning lain (aktivasi PPPoE, hotspot, suspend/unsuspend, change profile, TR-069) yang bisa berjalan. Ini adalah fondasi yang harus selesai lebih dulu.

## Prasyarat

- **Foundation domain `command`** sudah ada: `command.Command{Raw, Args}`, `command.Operation`, `command.Class`, `command.Decision`, serta `Execute/Classify/Translate/Close` di driver.
- **`usecase/network.ExecuteCommand`** (stub/parsial di `internal/usecase/network/execute_command.go`) sudah ada dan menjadi titik eksekusi tunggal (Classify → Decide → HITL bila destruktif → Execute → tulis `command_audit_log`). Issue ini mengonsumsinya; bila ada gap kecil pada `ExecuteCommand` (mis. cara mengembalikan `command_audit_log_id`), disebutkan sebagai dependency antar-Task di bawah.
- **`internal/registry/registry.go`** (`Get(ctx, deviceID) (port.DeviceDriver, error)`) sudah ada untuk resolusi driver per device.
- **Tabel `provisioning_sync_log`** (migrasi 000011) dan **`command_audit_log`** (migrasi 000017) sudah ada di skema.
- **Driver MikroTik** (`internal/driver/mikrotik/`) sudah lengkap; circuit-breaker (Task 8) menyentuh `connect.go` yang sudah ada.

Tidak ada issue lain yang menjadi prasyarat — Issue 01 adalah akar dependency.

## Ruang Lingkup

**In scope:**
- Domain `internal/domain/provisioning/` (tipe `SyncLog`, `TargetType`, `SyncAction`, `SyncStatus` sebagai konstanta yang mencerminkan CHECK constraint DB).
- Kontrak `port.ProvisioningSyncRepository`.
- Implementasi repo Postgres `internal/adapter/postgres/provisioning_sync_repository.go`.
- Usecase `internal/usecase/network/sync_provisioning.go` (`ProcessPendingSyncs`), termasuk resolusi device via registry dan delegasi penerjemahan ke driver vendor.
- Runner/scheduler (ticker sederhana di `main.go`) + in-process trigger opsional.
- Circuit-breaker singkat pada `mikrotik/connect.go` (temuan #7).
- Endpoint REST: list, detail, retry.
- Retry baris `failed` (endpoint + logika backoff sederhana/manual).
- Index opsional pada kolom `status` untuk query pending.

**Out of scope:**
- Logika bisnis penulis baris `pending` per fitur (aktivasi PPPoE, hotspot, TR-069) — itu issue-issue turunan yang memakai Sync Engine ini.
- Penambahan kolom baru pada `subscriptions` (`onu_pon_port`, `genieacs_device_id`, dll) — issue terpisah.
- Katalog command vendor baru di luar yang sudah ada — hanya dipakai, tidak diperluas di sini kecuali gap minimal untuk `Translate`.
- Antarmuka HITL/approval UI — Issue 01 hanya memanggil `ExecuteCommand`; keputusan approval sudah menjadi tanggung jawab layer itu.

## REST API

Base path: `/api/v1/`. Aksi ke perangkat mengikuti konvensi 202 Accepted + id `sync_log` (K4). Endpoint di bawah adalah endpoint pendukung untuk observabilitas dan retry Sync Engine; endpoint bisnis yang *menghasilkan* baris sync ada di issue turunan.

| Method | Path | Tujuan | Role minimum |
|---|---|---|---|
| GET | `/api/v1/sync-logs` | List baris sync dengan filter | staff |
| GET | `/api/v1/sync-logs/:id` | Detail satu baris sync + audit log tertaut | staff |
| POST | `/api/v1/sync-logs/:id/retry` | Set ulang baris `failed` menjadi `pending` | admin |

**GET `/api/v1/sync-logs`**
- Query param penting (semua opsional, dikombinasikan dengan AND): `subscription_id` (UUID), `status` (`pending`/`success`/`failed`), `target_type` (salah satu enum `target_type`), plus paginasi standar `limit`/`offset` (default `limit=50`, cap maksimum mis. 200) dan `order` by `requested_at desc`.
- Response: objek berisi array `data` dari ringkasan sync log (field: `id`, `subscription_id`, `target_type`, `device_id`, `action`, `status`, `requested_at`, `completed_at`, `error_message`, `external_reference`, `command_audit_log_id`) + metadata paginasi (`total`, `limit`, `offset`).
- Status: `200 OK` sukses; `400 Bad Request` bila nilai filter enum tidak valid; `401/403` bila auth/role gagal.

**GET `/api/v1/sync-logs/:id`**
- Path param: `id` (UUID baris sync).
- Response: objek detail sync log (semua kolom `provisioning_sync_log`) plus objek `command_audit_log` tertaut (di-embed bila `command_audit_log_id` tidak null) berisi field ringkas dari `command_audit_log` (`actor_type`, `source`, `classification`, `decision`, `command_raw`, `command_args`, `success`, `result_summary`, `error_message`).
- Status: `200 OK`; `404 Not Found` bila id tidak ada; `401/403` sesuai auth/role.

**POST `/api/v1/sync-logs/:id/retry`**
- Path param: `id` (UUID). Body kosong (atau opsional `{ "reset_error": true }` untuk membersihkan `error_message` lama).
- Perilaku: hanya baris berstatus `failed` yang boleh di-retry; baris di-set ulang ke `pending`, `error_message`/`completed_at`/`command_audit_log_id` dibersihkan sesuai kebutuhan, lalu (bila in-process trigger aktif) memicu `ProcessPendingSyncs`. Ini aksi yang berujung ke perangkat, maka mengikuti konvensi 202.
- Response: `202 Accepted` dengan body berisi `id` baris sync yang kini `pending` dan `status: "pending"`.
- Status: `202 Accepted` sukses; `409 Conflict` bila status baris bukan `failed` (mis. sedang `pending`/`success`); `404 Not Found` bila id tidak ada; `401/403` sesuai auth/role (role di bawah `admin` ditolak).

## Tasks

**Task 1: Domain provisioning (tipe & konstanta cermin DB)**

**Description:** Buat domain baru `internal/domain/provisioning/` berisi entity `SyncLog` dan tipe enum `TargetType`, `SyncAction`, `SyncStatus` yang nilai konstantanya persis sama dengan CHECK constraint di `provisioning_sync_log`. Tidak ada I/O, tidak impor library eksternal.

**Acceptance criteria:**
- [ ] File `internal/domain/provisioning/provisioning.go` berisi struct `SyncLog` dengan field yang memetakan seluruh kolom `provisioning_sync_log` (`ID`, `SubscriptionID`, `TargetType`, `DeviceID`, `Action`, `Status`, `RequestedAt`, `CompletedAt` nullable, `ErrorMessage` nullable, `CommandAuditLogID` nullable, `ExternalReference` nullable).
- [ ] Konstanta `TargetType` mencakup persis: `mikrotik_ppp_secret`, `mikrotik_hotspot_user`, `mikrotik_address_list`, `freeradius`, `genieacs_tr069` — tidak menambah/mengurangi.
- [ ] Konstanta `SyncAction` mencakup persis: `create`, `update`, `disable`, `enable`, `delete`, `change_profile`.
- [ ] Konstanta `SyncStatus` mencakup persis: `pending`, `success`, `failed`.
- [ ] Ada helper validasi (mis. `TargetType.Valid()`, `SyncAction.Valid()`, `SyncStatus.Valid()`) yang menolak nilai di luar enum.
- [ ] Setiap identifier exported punya doc comment diawali namanya (AGENTS.md §7).
- [ ] Tidak ada import ke `adapter`/`driver`/framework.

**Files likely touched:** `internal/domain/provisioning/provisioning.go`, `internal/domain/provisioning/errors.go` (bila perlu sentinel `ErrInvalidTargetType` dsb).

**Dependencies:** —

**Estimated scope:** Small

---

**Task 2: Kontrak port.ProvisioningSyncRepository**

**Description:** Definisikan interface repository di `internal/port/provisioning_sync_repository.go` sebagai kontrak tunggal untuk membaca/menulis baris `provisioning_sync_log`.

**Acceptance criteria:**
- [ ] File `internal/port/provisioning_sync_repository.go` mendefinisikan interface `ProvisioningSyncRepository`.
- [ ] Method minimal: `FindPending(ctx, limit) ([]provisioning.SyncLog, error)`, `FindByID(ctx, id) (provisioning.SyncLog, error)`, `FindBySubscription(ctx, subscriptionID, filter) ([]provisioning.SyncLog, error)`, `Create(ctx, log) (provisioning.SyncLog, error)`, `MarkSuccess(ctx, id, commandAuditLogID, externalReference) error`, `MarkFailed(ctx, id, errMessage string) error`, dan `ResetToPending(ctx, id) error` (dipakai retry).
- [ ] Ada method list dengan filter untuk endpoint GET (mis. `List(ctx, filter, page) ([]provisioning.SyncLog, int, error)` yang mengembalikan total untuk paginasi).
- [ ] `context.Context` selalu parameter pertama; `error` selalu return terakhir.
- [ ] Interface hanya bergantung ke `internal/domain/provisioning` (dan stdlib), tidak ke Postgres/GORM.
- [ ] Setiap method punya doc comment.

**Files likely touched:** `internal/port/provisioning_sync_repository.go`.

**Dependencies:** Task 1.

**Estimated scope:** Small

---

**Task 3: Implementasi repo Postgres**

**Description:** Implementasikan `ProvisioningSyncRepository` di `internal/adapter/postgres/provisioning_sync_repository.go`, memetakan ke tabel `provisioning_sync_log` yang sudah ada (migrasi 000011).

**Acceptance criteria:**
- [ ] Struct implementasi (mis. `SyncRepo`) dengan compile-time assertion bahwa ia memenuhi `port.ProvisioningSyncRepository`.
- [ ] `FindPending` mengambil baris `status='pending'` diurut `requested_at asc`, dibatasi `limit`, dan aman terhadap pemrosesan konkuren (lihat Task 4 untuk strategi klaim — repo menyediakan mekanisme, mis. `SELECT ... FOR UPDATE SKIP LOCKED` atau update status ke penanda in-progress; bila enum status DB hanya `pending/success/failed`, klaim dilakukan lewat locking transaksi, bukan menambah status baru).
- [ ] `MarkSuccess` mengisi `status='success'`, `completed_at=now()`, `command_audit_log_id`, dan `external_reference` bila ada.
- [ ] `MarkFailed` mengisi `status='failed'`, `completed_at=now()`, `error_message`.
- [ ] `ResetToPending` hanya berhasil bila status saat ini `failed`; mengembalikan error sentinel yang bisa dipetakan ke 409 bila status bukan `failed`.
- [ ] `List`/`FindBySubscription` mendukung filter `subscription_id`/`status`/`target_type` + paginasi, mengembalikan total count.
- [ ] Mapping model DB ↔ domain tidak membocorkan tipe GORM ke domain.
- [ ] Error `sql.ErrNoRows`/tidak ditemukan dipetakan ke sentinel domain via `errors.Is`.

**Files likely touched:** `internal/adapter/postgres/provisioning_sync_repository.go`, `internal/adapter/postgres/models.go` (bila perlu model row).

**Dependencies:** Task 1, Task 2.

**Estimated scope:** Medium

---

**Task 4: Usecase ProcessPendingSyncs (orkestrasi inti)**

**Description:** Buat `internal/usecase/network/sync_provisioning.go` dengan `ProcessPendingSyncs(ctx)` yang mengambil baris pending, meresolusi device via registry, mendelegasikan penerjemahan `target_type`+`action` ke driver vendor, memanggil `ExecuteCommand`, lalu menulis hasil ke `sync_log`. Perhatikan bahwa satu operasi bisnis bisa menerjemah menjadi **sekuens command berurutan** (mis. `change_profile`/`disable` PPPoE yang harus diikuti kill sesi aktif — K9, lihat README §Konvensi Bersama); Sync Engine harus mendukung "satu baris `sync_log` → satu sekuens command" tanpa mengetahui isi sekuensnya (itu tanggung jawab driver).

**Acceptance criteria:**
- [ ] `ProcessPendingSyncs(ctx)` mengambil batch pending via `repo.FindPending`.
- [ ] Untuk tiap baris: resolusi driver via `registry.Get(ctx, deviceID)`; bila device tidak ditemukan/unreachable → `MarkFailed` dengan pesan jelas, lanjut ke baris berikutnya (satu baris gagal tidak menghentikan batch).
- [ ] Penerjemahan `target_type`+`action` menjadi `command.Operation` lalu `command.Command` dilakukan lewat `driver.Translate` / katalog vendor — **tidak ada** string command MikroTik/RouterOS yang di-hardcode di file usecase ini (K1). Pemetaan `(target_type, action)` → `command.Operation` boleh berada di usecase sebagai tabel abstrak, tetapi bentuk `Raw` native harus datang dari driver.
- [ ] Sync Engine mendukung satu baris `sync_log` yang menerjemah menjadi **sekuens command berurutan** dari driver (`driver.Translate` boleh mengembalikan lebih dari satu `command.Command`, mis. set profile → `/ppp active remove` per K9). Usecase mengeksekusi sekuens itu **berurutan** dan berhenti pada command pertama yang gagal (baris ditandai `failed`, sisa sekuens tidak dijalankan). Isi/urutan sekuens tidak diketahui usecase — murni kontrak driver.
- [ ] **Pilihan granularitas audit dinyatakan eksplisit:** satu baris `sync_log` dengan sekuens N command menghasilkan **N** baris `command_audit_log` (satu per command native yang benar-benar dieksekusi ke perangkat, K4); `command_audit_log_id` yang disimpan di baris `sync_log` adalah id command **terakhir** yang sukses dalam sekuens. Perilaku ini didokumentasikan di ADR (Task 10).
- [ ] **Idempotensi dijaga di kontrak driver (K10), bukan di usecase:** karena Sync Engine me-retry baris `failed`, penerjemahan/eksekusi tiap aksi harus aman diulang — driver melakukan cek-sebelum-tulis (set+kick profil hanya bila berubah, jangan double-kick). Usecase tidak menambah dedup sendiri; ini ditegaskan sebagai acceptance kontrak driver di issue-issue turunan dan diuji di sana. Retry baris yang sama tidak boleh menghasilkan efek ganda di perangkat.
- [ ] Command dieksekusi via `usecase/network.ExecuteCommand` (Classify → Decide → HITL → Execute), bukan `driver.Execute` langsung.
- [ ] Hasil sukses → `repo.MarkSuccess` dengan `command_audit_log_id` yang dikembalikan `ExecuteCommand`; hasil gagal atau butuh approval yang ditolak → `repo.MarkFailed`.
- [ ] Bila `ExecuteCommand` mengembalikan status "menunggu approval" (destruktif, HITL), baris `sync_log` **tidak** ditandai `failed` secara keliru — didefinisikan perilaku eksplisit (mis. tetap `pending` dan didokumentasikan) dan disebut sebagai Open Question bila `command.Decision` belum menyediakan status pending yang jelas.
- [ ] Pemrosesan antar-baris konkuren dibatasi dengan `errgroup` + batas paralelisme; tidak ada goroutine fire-and-forget (AGENTS.md §5).
- [ ] Test table-driven mencakup: baris sukses, device tidak ditemukan, `Translate` gagal (target belum didukung), `ExecuteCommand` error.

**Files likely touched:** `internal/usecase/network/sync_provisioning.go`, kemungkinan penyesuaian kecil di `internal/usecase/network/execute_command.go` agar mengembalikan `command_audit_log_id`.

**Dependencies:** Task 2, Task 3; foundation `ExecuteCommand` dan `registry`.

**Estimated scope:** Large

---

**Task 5: Runner/scheduler (ticker) + in-process trigger**

**Description:** Sediakan mekanisme yang memicu `ProcessPendingSyncs` secara periodik (ticker sederhana di `main.go` dengan `errgroup`/graceful shutdown) sebagai jaring pengaman, plus jalur in-process trigger agar usecase bisnis dapat memicu pemrosesan segera setelah menulis baris pending (latensi rendah).

**Acceptance criteria:**
- [ ] Ticker berjalan pada interval yang dikonfigurasi (mis. `SYNC_ENGINE_INTERVAL`, default beberapa detik) dan berhenti bersih saat context `main` dibatalkan.
- [ ] Ticker dijalankan dalam goroutine yang ditunggu (errgroup), bukan fire-and-forget.
- [ ] Ada mekanisme trigger in-process (mis. channel/notifier yang di-inject ke usecase bisnis) sehingga penulisan baris pending dapat "membangunkan" engine tanpa menunggu tick berikutnya; scheduler tetap sebagai fallback untuk baris yang gagal terpicu.
- [ ] Konfigurasi interval dibaca via `internal/config/config.go` (bukan hardcode).
- [ ] Dokumentasikan pilihan **ticker vs library cron** sebagai Open Question; rekomendasikan ticker sederhana untuk MVP, cron hanya bila kebutuhan penjadwalan bertambah.
- [ ] Tidak ada panic di jalur runtime (panic hanya di startup init bila config wajib hilang).

**Files likely touched:** `cmd/server/main.go`, `internal/config/config.go`, mungkin `internal/usecase/network/sync_provisioning.go` (untuk API notifier).

**Dependencies:** Task 4.

**Estimated scope:** Medium

---

**Task 6: Handler REST + DTO (list & detail)**

**Description:** Buat handler `internal/adapter/http/sync_log_handler.go` untuk `GET /api/v1/sync-logs` dan `GET /api/v1/sync-logs/:id`, dengan DTO request/response di `internal/adapter/http/dto/`.

**Acceptance criteria:**
- [ ] `GET /api/v1/sync-logs` memparsing filter `subscription_id`/`status`/`target_type` + paginasi, memvalidasi nilai enum (400 bila invalid), memanggil `repo.List`, mengembalikan `200` dengan array + metadata paginasi.
- [ ] `GET /api/v1/sync-logs/:id` memanggil `repo.FindByID`, meng-embed `command_audit_log` tertaut bila `command_audit_log_id` tidak null, mengembalikan `200`; `404` bila tidak ada.
- [ ] DTO response tidak membocorkan tipe internal domain langsung; mapping domain → DTO eksplisit.
- [ ] Route didaftarkan di `internal/adapter/http/router.go`.
- [ ] Kedua endpoint dilindungi middleware auth; role minimum `staff` (baca) via RBAC Casbin (Task 9).
- [ ] Test handler pakai `httptest` mencakup: filter valid, filter enum invalid (400), detail ditemukan, detail tidak ditemukan (404).

**Files likely touched:** `internal/adapter/http/sync_log_handler.go`, `internal/adapter/http/dto/sync_log.go`, `internal/adapter/http/router.go`.

**Dependencies:** Task 3.

**Estimated scope:** Medium

---

**Task 7: Endpoint retry (POST /sync-logs/:id/retry)**

**Description:** Tambahkan handler untuk `POST /api/v1/sync-logs/:id/retry` yang mereset baris `failed` menjadi `pending`, memicu engine, dan mengembalikan 202.

**Acceptance criteria:**
- [ ] Endpoint memanggil `repo.ResetToPending`; bila status baris bukan `failed`, mengembalikan `409 Conflict` dengan pesan jelas.
- [ ] Bila in-process trigger aktif (Task 5), retry membangunkan engine segera.
- [ ] Response `202 Accepted` berisi `id` dan `status: "pending"`.
- [ ] `404` bila id tidak ada; `403` bila role di bawah `admin`.
- [ ] Logika backoff/retry sederhana didokumentasikan: retry bersifat manual via endpoint ini (tidak auto-retry tak terbatas); bila kelak ingin auto-retry, disebut sebagai Open Question dengan batas percobaan.
- [ ] Test handler pakai `httptest`: retry baris `failed` (202), retry baris `pending`/`success` (409), id tidak ada (404).

**Files likely touched:** `internal/adapter/http/sync_log_handler.go`, `internal/adapter/http/router.go`.

**Dependencies:** Task 3, Task 6.

**Estimated scope:** Small

---

**Task 8: Circuit-breaker singkat pada mikrotik/connect.go (temuan #7)**

**Description:** Cache kegagalan TCP connect per host selama ~5 detik di `dialAndLogin` agar `NewDriver` berturut-turut ke host yang sedang down tidak menunggu timeout penuh berulang kali.

**Acceptance criteria:**
- [ ] Di `internal/driver/mikrotik/connect.go`, `dialAndLogin` memeriksa cache kegagalan per host sebelum mencoba dial; bila host baru saja gagal connect dalam jendela ~5 detik, langsung kembalikan error tanpa menunggu timeout.
- [ ] Jendela waktu (mis. 5 detik) dapat dikonfigurasi atau berupa konstanta bernama (`PascalCase`, bukan `ALL_CAPS`).
- [ ] Cache thread-safe (mutex sebagai field, tidak di-embed) dan tidak bocor tak terbatas (entri kedaluwarsa dibersihkan/di-overwrite).
- [ ] Keberhasilan connect membersihkan/mereset entri kegagalan host tersebut.
- [ ] Interval waktu tidak menggunakan `Date.now`-style non-deterministik di test; test menggunakan clock yang dapat di-inject atau toleransi waktu wajar.
- [ ] Perilaku baru tidak mengubah signature publik driver; hanya optimasi internal.

**Files likely touched:** `internal/driver/mikrotik/connect.go`, `internal/driver/mikrotik/connect_test.go`.

**Dependencies:** — (independen dari Task 1-7, dapat dikerjakan paralel).

**Estimated scope:** Medium

---

**Task 9: RBAC policy untuk endpoint sync-logs**

**Description:** Daftarkan aturan Casbin agar endpoint sync-logs mengikuti K3: baca (`staff`+), retry (`admin`+).

**Acceptance criteria:**
- [ ] `configs/rbac_policy.csv` menambah aturan: `GET /api/v1/sync-logs` dan `GET /api/v1/sync-logs/:id` untuk `staff`, `teknisi`, `admin`, `owner`, `superadmin`.
- [ ] `POST /api/v1/sync-logs/:id/retry` hanya untuk `admin`, `owner`, `superadmin`.
- [ ] Middleware `rbac.go` menegakkan aturan pada route yang didaftarkan.
- [ ] Test middleware/handler membuktikan `staff` ditolak retry (403) tetapi boleh list/detail.

**Files likely touched:** `configs/rbac_policy.csv`, `internal/adapter/http/middleware/rbac.go` (bila perlu penyesuaian mapping).

**Dependencies:** Task 6, Task 7.

**Estimated scope:** Small

---

**Task 10: ADR + dokumentasi arsitektur Sync Engine**

**Description:** Tulis ADR yang mendokumentasikan keputusan desain Sync Engine (pola pull dari `provisioning_sync_log`, delegasi Translate ke driver, ticker + in-process trigger, strategi klaim baris konkuren) dan tautkan dari `README.md` root.

**Acceptance criteria:**
- [ ] File `docs/adr/0006-provisioning-sync-engine.md` dibuat (nomor lanjut dari 0005).
- [ ] ADR menjelaskan: alasan pola sync-log-driven (K4), mengapa penerjemahan di driver bukan usecase, pilihan ticker vs cron, strategi konkurensi (locking vs status in-progress), perilaku baris saat menunggu HITL approval, granularitas audit untuk satu baris sync yang menerjemah ke sekuens command (N command → N `command_audit_log`, K9), dan kontrak idempotensi driver yang membuat retry aman (K10).
- [ ] ADR ditautkan dari `README.md` root pada commit yang sama (AGENTS.md §1.5).
- [ ] `DATABASE-SCHEMA.md` diperbarui bila Task 11 (index) dijalankan.

**Files likely touched:** `docs/adr/0006-provisioning-sync-engine.md`, `README.md`, `docs/plan-provisioning/` (link balik bila diperlukan).

**Dependencies:** Task 4, Task 5.

**Estimated scope:** Small

---

**Task 11 (opsional): Migrasi index pada status**

**Description:** Bila profiling menunjukkan `FindPending` lambat, tambahkan index pada kolom `status` (atau composite `(status, requested_at)`) di `provisioning_sync_log`.

**Acceptance criteria:**
- [ ] Migrasi baru `000022_add_index_provisioning_sync_log_status.up.sql` + `.down.sql` (nomor lanjut dari 000021).
- [ ] `up` membuat index pada `(status, requested_at)`; `down` menghapusnya.
- [ ] Tidak ada perubahan kolom/constraint lain.
- [ ] `DATABASE-SCHEMA.md §6.3` diperbarui menyebut index baru pada commit yang sama.

**Files likely touched:** `migrations/000022_add_index_provisioning_sync_log_status.up.sql`, `migrations/000022_add_index_provisioning_sync_log_status.down.sql`, `DATABASE-SCHEMA.md`.

**Dependencies:** Task 3 (untuk menilai kebutuhan lewat query nyata).

**Estimated scope:** Small

## Migrasi Database

Tidak ada tabel baru — `provisioning_sync_log` sudah dibuat migrasi 000011 dan `command_audit_log` migrasi 000017. Issue ini memakainya apa adanya.

**Migrasi kecil opsional (Task 11):** menambah index untuk mempercepat query baris pending.
- Nomor migrasi: **000022**, lanjut dari 000021.
- File: `migrations/000022_add_index_provisioning_sync_log_status.up.sql` dan pasangan `.down.sql`.
- Isi yang diperlukan (dijelaskan sebagai teks): buat index pada kolom `status`, sebaiknya composite `(status, requested_at)` agar sekaligus mendukung urutan `requested_at asc` pada `FindPending`; `down` menghapus index tersebut. Satu perubahan skema per pasang file, tidak menyentuh kolom lain.
- Cerminkan penambahan index ini ke `DATABASE-SCHEMA.md §6.3` pada PR/commit yang sama (K6).

## Verification

- [ ] `go build ./...` sukses tanpa error.
- [ ] `go vet ./...` dan `gofumpt`/`goimports` bersih.
- [ ] `go test ./internal/domain/provisioning/...` — validasi enum & helper.
- [ ] `go test ./internal/usecase/network/...` — table-driven `ProcessPendingSyncs` (sukses, device tak ditemukan, Translate gagal, ExecuteCommand error).
- [ ] `go test ./internal/adapter/postgres/...` — repo dengan `testcontainers-go` (Postgres asli), termasuk `FindPending`, `MarkSuccess`, `MarkFailed`, `ResetToPending` (termasuk kasus status bukan `failed`).
- [ ] `go test ./internal/adapter/http/...` — handler `httptest` untuk list/detail/retry (200/400/404/409/403).
- [ ] `go test ./internal/driver/mikrotik/...` — circuit-breaker `dialAndLogin`.
- [ ] `make lint` lolos (golangci-lint, staticcheck untuk doc comment & akronim).
- [ ] Smoke test manual (sebut sebagai curl): `curl` GET `/api/v1/sync-logs?status=pending` dengan token `staff` → 200; `curl` POST `/api/v1/sync-logs/<id>/retry` dengan token `staff` → 403, dengan token `admin` pada baris `failed` → 202, pada baris `pending` → 409.
- [ ] Smoke test integrasi: sisipkan satu baris `pending` (mis. via seed/skrip) menargetkan MikroTik CHR di GNS3, jalankan engine, verifikasi baris menjadi `success` dan `command_audit_log` terisi.

## Definition of Done

- [ ] Domain `provisioning`, `port.ProvisioningSyncRepository`, dan implementasi Postgres selesai dan teruji.
- [ ] `ProcessPendingSyncs` memproses baris pending end-to-end via `ExecuteCommand`, tanpa hardcode command vendor di usecase.
- [ ] Runner ticker + in-process trigger aktif dengan graceful shutdown.
- [ ] Tiga endpoint (`GET list`, `GET detail`, `POST retry`) berfungsi dengan role RBAC benar.
- [ ] Circuit-breaker MikroTik (temuan #7) terpasang dan teruji.
- [ ] ADR 0006 ditulis dan ditautkan dari `README.md`; `DATABASE-SCHEMA.md` diperbarui bila index ditambahkan.
- [ ] Semua perintah pada bagian Verification hijau.
- [ ] Open Questions (ticker vs cron, perilaku baris saat menunggu HITL, auto-retry) didokumentasikan di ADR, bukan dibiarkan implisit.
