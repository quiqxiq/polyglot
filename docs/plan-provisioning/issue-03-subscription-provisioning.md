# Issue 03: Subscription Provisioning Lifecycle

## Konteks

Foundation sudah menyediakan CRUD subscription murni (`internal/usecase/business/manage_subscription.go` — create/read/update data di Postgres), tetapi membuat baris di tabel `subscriptions` **tidak** membuat pelanggan bisa terhubung: RouterOS belum tahu apa-apa tentang username PPPoE tersebut. Issue ini menutup celah itu — mengubah subscription dari `pending_install` menjadi `active` dengan cara membuat `/ppp secret` PPPoE nyata di router (username, password, profile, `service=pppoe`), lewat pola sinkronisasi kanonik (K4). Ini adalah realisasi baris pertama dari tabel pemetaan `DATABASE-SCHEMA.md` §7.2: peristiwa bisnis "Subscription baru (PPPoE)" → `provisioning_sync_log.action='create'`, dan "Ganti paket" → `action='change_profile'`.

`ANALISIS-PROVISIONING-REPO-REFERENSI.md` menegaskan bahwa di repo referensi (gembok-simple, billing-rtrw) provisioning PPPoE dilakukan dengan menembak command RouterOS langsung dari service layer — jalur yang persis dilarang oleh K4 di sini. Polyglot menempuh jalur yang berbeda dan lebih dapat diaudit: handler REST hanya menulis "niat" ke `provisioning_sync_log` status `pending`; Sync Engine (Issue 01) yang menerjemahkannya jadi `command.Command` dan mengeksekusinya lewat `usecase/network.ExecuteCommand`, sehingga setiap `/ppp secret/add` berakhir sebagai satu baris `command_audit_log` — sama seperti kalau AI agent lewat MCP menjalankan command yang sama.

Aspek kredensial adalah inti issue ini. `subscriptions.pppoe_password_encrypted` disimpan terenkripsi AES lewat `port.CredentialVault` (mekanisme yang sama dengan `internal/adapter/vault`). Password plaintext **hanya** didekripsi di titik terakhir — saat Sync Engine menyusun `command.Command` untuk `/ppp secret/add` — dan tidak pernah dilewatkan ke lapisan yang bisa dibaca AI/LLM. Ini konsekuensi langsung dari prinsip `port.CredentialVault` (`DATABASE-SCHEMA.md` §1, poin kredensial terenkripsi).

## Prasyarat

- **Issue 01 (Provisioning Sync Engine)** — wajib. Tanpa Sync Engine, baris `provisioning_sync_log` `pending` tidak akan pernah diproses jadi command; endpoint di sini hanya akan menumpuk baris `pending`.
- **Issue 02 (Plan Router Profile Sync)** — wajib. Provision `/ppp secret` butuh nama profil MikroTik yang valid, yang berasal dari `plan_router_profiles.mikrotik_profile_name` untuk pasangan (plan, device) subscription. Kalau profil belum tersinkron ke router, secret akan merujuk profil yang tidak ada.
- **Foundation Phase 4–5** — router Gin nyata, middleware `AuthRequired` + `RBACRequired`, `internal/adapter/http/dto/`, wiring `cmd/server/main.go`, serta `port.CredentialVault` + adapter vault sudah aktif.
- Tabel yang dipakai sudah ada: `subscriptions` (migrasi 000009), `subscription_status_history` (000010), `provisioning_sync_log` (000011), `plan_router_profiles` (000006).

## Ruang Lingkup

**In scope:**
- Perluasan `manage_subscription.go` (business) dengan aksi **provision** dan **activate**, plus **change-plan** (opsional).
- Penulisan baris `provisioning_sync_log` `target_type='mikrotik_ppp_secret'` dengan `action` `create`/`change_profile`.
- Guard status transition (`pending_install`/`suspended` → provision/activate) + penulisan `subscription_status_history`.
- Validasi keberadaan `plan_router_profiles` untuk (plan, device) sebelum provision (409 bila tidak ada).
- Dekripsi password PPPoE **di Sync Engine** (kontrak, bukan enkripsi kedua).
- 2 endpoint wajib (`/provision`, `/activate`) + 1 opsional (`/change-plan`) + baris RBAC.

**Out of scope:**
- Suspend/resume/terminate cascade → Issue 04.
- Hotspot user provisioning (`mikrotik_hotspot_user`) → issue terpisah.
- Penerjemahan `command.Command` RouterOS itu sendiri + eksekusi ke device → sudah milik Issue 01 + `internal/driver/mikrotik`.
- Pembuatan/penyimpanan awal `pppoe_username`/`pppoe_password_encrypted` (bagian dari CRUD create di foundation) — issue ini hanya **memakai**-nya.

## REST API

Semua di bawah `/api/v1/`. Aksi yang menyentuh perangkat mengembalikan **202 Accepted** + id `provisioning_sync_log` sesuai K5 (bukan 200 seolah sudah tereksekusi di router).

| Method | Path | Tujuan | Role minimum |
|---|---|---|---|
| POST | `/api/v1/subscriptions/:id/provision` | Buat `/ppp secret` PPPoE di router (tulis sync_log `create`) | `teknisi` |
| POST | `/api/v1/subscriptions/:id/activate` | Tandai subscription `active` setelah secret terpasang/terverifikasi | `teknisi` |
| POST | `/api/v1/subscriptions/:id/change-plan` | Ganti paket → `change_profile` di router (opsional) | `admin` |

### POST `/api/v1/subscriptions/:id/provision`
- **Request:** body kosong atau opsional `{ "reason": "<alasan/catatan operator>" }`. `:id` = UUID subscription.
- **Perilaku:** validasi subscription ada, `service_type='pppoe'`, status ∈ {`pending_install`,`suspended`}, `pppoe_username` + `pppoe_password_encrypted` terisi, dan ada `plan_router_profiles` untuk (`plan_id`,`device_id`) dengan `sync_status='synced'`. Bila lolos, tulis **satu** baris `provisioning_sync_log` (`target_type='mikrotik_ppp_secret'`, `action='create'`, `status='pending'`, `subscription_id`, `device_id`) dalam transaksi bisnis. Tidak mengubah status subscription (status berubah di endpoint activate atau setelah sync sukses, sesuai kebijakan — lihat Task 4).
- **Field `/ppp secret` (kanonik):** hanya `name` (=`pppoe_username`), `password`, `service=pppoe`, `profile` (=`plan_router_profiles.mikrotik_profile_name`), plus **opsional** `remote-address` (IP statik bila plan mengalokasikannya). **Tidak** ada `rate-limit` di `/ppp secret` — pengaturan kecepatan hidup di `/ppp profile` dan sudah milik Issue 02. Menaruh rate-limit di secret adalah anti-pattern referensi yang tidak ditiru di sini.
- **Response sukses:** `202 Accepted`, body `{ "sync_log_id": "<uuid>", "status": "pending" }`.
- **Response gagal:** `404` subscription tak ada; `400` `service_type` bukan pppoe atau kredensial belum lengkap; `409` status tidak mengizinkan provision **atau** `plan_router_profiles` belum ada/belum `synced` (pesan jelas, mis. `"router profile belum disinkron untuk device ini — jalankan Issue 02 dulu"`).

### POST `/api/v1/subscriptions/:id/activate`
- **Request:** body opsional `{ "reason": "<catatan>" }`.
- **Perilaku:** guard status (`pending_install`/`suspended` → `active`); set `subscriptions.status='active'`, isi `activated_at`; tulis baris `subscription_status_history` (`old_status`, `new_status='active'`, `changed_by_user` dari klaim JWT, `changed_by_actor_type='human'`, `reason`). Endpoint ini adalah transisi **status bisnis** — tidak selalu menyentuh perangkat; kalau kebijakan mengharuskan verifikasi secret aktif di router lebih dulu, itu dibaca dari `provisioning_sync_log` terakhir (`success`).
- **Response sukses:** `200 OK` bila murni transisi status (tanpa sync perangkat baru), body subscription ringkas; `202 Accepted` + `sync_log_id` bila implementasi memilih memicu verifikasi/aktivasi lewat sync. Pilih satu dan konsisten (rekomendasi: `200` — activate = transisi status murni; secret sudah dibuat di `/provision`).
- **Response gagal:** `404` tak ada; `409` status tidak mengizinkan aktivasi (mis. sudah `active`/`terminated`, atau belum ada sync `create` yang `success`).

### POST `/api/v1/subscriptions/:id/change-plan` (opsional)
- **Request:** `{ "plan_id": "<uuid plan baru>", "reason": "<alasan>" }`.
- **Perilaku:** validasi plan baru ada, ada `plan_router_profiles` `synced` untuk (`plan_id` baru,`device_id`). Update `subscriptions.plan_id`; tulis satu baris `provisioning_sync_log` (`action='change_profile'`, `target_type='mikrotik_ppp_secret'`) supaya Sync Engine mengeluarkan set profil baru. **Wajib K9 (lihat README §Konvensi Bersama):** `set profile=` pada `/ppp secret` **tidak** berlaku ke sesi PPPoE yang sedang online — profil baru baru aktif saat dial ulang. Karena itu satu operasi abstrak `change_profile` menerjemah jadi **sekuens** command di driver (`commands.go`): set profil → `/ppp active remove [find name=<pppoe_username>]`. Kedua langkah teraudit sebagai baris `command_audit_log` di bawah satu baris `provisioning_sync_log` (K4). Catat perubahan (opsional) ke `subscription_status_history` bila status ikut berubah (umumnya tidak).
- **Response sukses:** `202 Accepted` + `sync_log_id`.
- **Response gagal:** `404` subscription/plan tak ada; `409` profil paket baru belum tersinkron untuk device; `400` `plan_id` kosong/invalid.

## Tasks

**Task 1: Kontrak dekripsi kredensial di Sync Engine (bukan enkripsi kedua)**

**Description:** Pastikan Sync Engine (Issue 01) mendekripsi `pppoe_password_encrypted` lewat `port.CredentialVault` tepat saat menyusun `command.Command` untuk `action='create'`/`change_profile`, sehingga plaintext tidak pernah keluar dari batas Sync Engine dan tidak pernah tersimpan/di-log.

**Acceptance criteria:**
- [ ] Sync Engine mengambil `pppoe_username` + `pppoe_password_encrypted` dari `subscriptions` saat memproses baris `mikrotik_ppp_secret`/`create`.
- [ ] Dekripsi memakai `port.CredentialVault` yang sama (AES) — **tidak** ada mekanisme enkripsi/dekripsi baru diperkenalkan.
- [ ] Password plaintext hanya ada di dalam `command.Args` yang diteruskan ke driver; tidak masuk `command_audit_log.command_args` mentah (di-mask/redact) dan tidak masuk log aplikasi.
- [ ] Bila dekripsi gagal, baris sync jadi `failed` dengan `error_message` yang tidak membocorkan ciphertext.
- [ ] `Translate` untuk `create` menghasilkan `/ppp secret/add` dengan **hanya** field kanonik (`name`/`password`/`service=pppoe`/`profile`, opsional `remote-address`) — tanpa `rate-limit`.
- [ ] `Translate` untuk `change_profile` menghasilkan **sekuens** [set profil] → [`/ppp active remove [find name=<user>]`], keduanya teraudit (K9 — lihat README §Konvensi Bersama); pengetahuan sekuens ini hidup di `internal/driver/mikrotik/commands.go`, bukan usecase.

**Files likely touched:** `internal/usecase/network/` (perluasan Sync Engine dari Issue 01, mis. penerjemah `mikrotik_ppp_secret`), `internal/port/credential_vault.go` (dipakai, tidak diubah), `internal/driver/mikrotik/commands.go` (Translate `OpCreatePPPSecret` bila belum ada).

**Dependencies:** Issue 01.

**Estimated scope:** Medium.

---

**Task 2: Aksi Provision di manage_subscription.go**

**Description:** Tambah fungsi orkestrasi (mis. `ProvisionSubscription`) yang memvalidasi prasyarat dan menulis satu baris `provisioning_sync_log` `create` dalam satu transaksi bisnis.

**Acceptance criteria:**
- [ ] Fungsi menerima `ctx`, `subscriptionID`, dan konteks aktor (user id + actor_type), mengembalikan id `sync_log` + error.
- [ ] Validasi: subscription ada, `service_type='pppoe'`, status ∈ {`pending_install`,`suspended`}, `pppoe_username`/`pppoe_password_encrypted` non-kosong.
- [ ] Validasi profil: ada `plan_router_profiles` untuk (`plan_id`,`device_id`) dengan `sync_status='synced'` — jika tidak, kembalikan sentinel error yang dipetakan ke 409.
- [ ] Menulis tepat **satu** baris `provisioning_sync_log` (`target_type='mikrotik_ppp_secret'`, `action='create'`, `status='pending'`, `subscription_id`, `device_id`) — bukan langsung memanggil `port.DeviceDriver` (K4).
- [ ] Sentinel error terdefinisi di domain terkait (mis. `ErrProfileNotSynced`, `ErrInvalidStatusTransition`, `ErrCredentialsMissing`) di `internal/domain/subscription/errors.go`.

**Files likely touched:** `internal/usecase/business/manage_subscription.go`, `internal/domain/subscription/errors.go`, `internal/port/subscription_repository.go` (bila butuh query profil/insert sync_log — atau lewat repo yang sudah ada).

**Dependencies:** Task 1, Issue 02.

**Estimated scope:** Medium.

---

**Task 3: Aksi Activate + subscription_status_history**

**Description:** Tambah fungsi (mis. `ActivateSubscription`) yang melakukan transisi status ke `active` dengan guard dan mencatat riwayat status.

**Acceptance criteria:**
- [ ] Guard: hanya status `pending_install`/`suspended` yang boleh → `active`; selain itu kembalikan `ErrInvalidStatusTransition` (→409).
- [ ] Update `subscriptions.status='active'` + isi `activated_at`, dalam satu transaksi dengan insert `subscription_status_history`.
- [ ] `subscription_status_history` terisi: `old_status`, `new_status='active'`, `changed_by_user` (dari klaim JWT), `changed_by_actor_type` sesuai pemicu (`human` untuk REST manual; `system_scheduled` bila dipanggil job; `ai_agent` bila lewat MCP).
- [ ] (Bila kebijakan verifikasi dipakai) menolak aktivasi jika belum ada baris `provisioning_sync_log` `create` berstatus `success` untuk subscription itu.

**Files likely touched:** `internal/usecase/business/manage_subscription.go`, `internal/domain/subscription/` (helper transisi status bila ada), `internal/port/subscription_repository.go` (metode insert status history bila belum ada).

**Dependencies:** Task 2.

**Estimated scope:** Medium.

---

**Task 4: Aksi Change-Plan (opsional) → change_profile**

**Description:** Tambah fungsi (mis. `ChangeSubscriptionPlan`) yang mengganti `plan_id` dan menulis baris `provisioning_sync_log` `change_profile`.

**Acceptance criteria:**
- [ ] Validasi plan baru ada + ada `plan_router_profiles` `synced` untuk (`plan_id` baru,`device_id`) — jika tidak, 409.
- [ ] Update `subscriptions.plan_id` dan tulis satu baris `provisioning_sync_log` (`action='change_profile'`, `target_type='mikrotik_ppp_secret'`, `status='pending'`) dalam satu transaksi.
- [ ] **K9 (field-tested, lihat README §Konvensi Bersama):** operasi `change_profile` **wajib** memutus sesi PPPoE aktif sebagai langkah kedua — `set profile=` di `/ppp secret` tidak berlaku ke sesi online sampai dial ulang. Ini kontrak driver: `Translate(change_profile)` mengembalikan sekuens [set profil] → [`/ppp active remove [find name=<pppoe_username>]`], keduanya berakhir sebagai `command_audit_log` di bawah baris sync yang sama. Test membuktikan langkah `active remove` hadir (referensi lapangan sering melewatkannya).
- [ ] Tidak mengubah `status` subscription (ganti paket ≠ ganti status), kecuali kebijakan ISP menyatakan lain (didokumentasikan bila iya).
- [ ] Idempotency ringan: mengganti ke plan yang sama dengan plan saat ini ditolak/no-op (400 atau 200 tanpa sync baru — pilih dan konsisten).

**Files likely touched:** `internal/usecase/business/manage_subscription.go`, `internal/domain/subscription/errors.go`.

**Dependencies:** Task 2.

**Estimated scope:** Small.

---

**Task 5: Handler REST + DTO + routing + RBAC**

**Description:** Tambah handler untuk `/provision`, `/activate`, dan `/change-plan`, DTO request/response, daftarkan route di router, tambah baris Casbin.

**Acceptance criteria:**
- [ ] Handler di `internal/adapter/http/subscription_handler.go` memanggil usecase business, tidak menyentuh driver (K4).
- [ ] DTO request/response di `internal/adapter/http/dto/` (mis. `ChangePlanRequest`, `ProvisionResponse` berisi `sync_log_id` + `status`).
- [ ] Route terdaftar di `internal/adapter/http/router.go` dalam group ber-middleware `AuthRequired` + `RBACRequired`.
- [ ] Pemetaan error → status sesuai K5: 404/400/409/202/200 seperti bagian REST API di atas; format error `{ "error": { "code", "message" } }`.
- [ ] `/provision` & `/activate` mengembalikan sesuai kontrak (202 + `sync_log_id` untuk provision; 200 untuk activate murni status); `/change-plan` → 202 + `sync_log_id`.
- [ ] Baris RBAC ditambahkan ke `configs/rbac_policy.csv`: `teknisi` (provision, activate), `admin`+ (change-plan), `superadmin`/`owner` semua; `staff` tidak diberi akses ketiga aksi ini.

**Files likely touched:** `internal/adapter/http/subscription_handler.go`, `internal/adapter/http/dto/subscription.go`, `internal/adapter/http/router.go`, `configs/rbac_policy.csv`.

**Dependencies:** Task 2, Task 3, (Task 4 bila change-plan diikutkan).

**Estimated scope:** Medium.

---

**Task 6: Test — usecase (table-driven), repo (testcontainers), handler (httptest)**

**Description:** Uji guard status, validasi profil, dan penulisan sync_log/status_history di ketiga lapisan.

**Acceptance criteria:**
- [ ] Test usecase table-driven mencakup: provision dari `pending_install` (sukses), dari `suspended` (sukses), dari `active`/`terminated` (409), tanpa `plan_router_profiles` `synced` (409), kredensial kosong (400), `service_type` non-pppoe (400).
- [ ] Test usecase activate: transisi valid/invalid + baris `subscription_status_history` benar-benar tertulis dengan `changed_by_actor_type` yang tepat.
- [ ] Test repo Postgres pakai `testcontainers-go` (bukan mock) untuk memverifikasi baris `provisioning_sync_log` + `subscription_status_history` tersimpan dan transaksi atomic (rollback saat error).
- [ ] Test handler pakai `httptest`: kode status HTTP dan bentuk body (`sync_log_id`) sesuai kontrak; RBAC menolak `staff`.
- [ ] Test driver `commands.go`: `Translate(change_profile)` mengembalikan sekuens yang menyertakan `/ppp active remove [find name=<user>]` setelah `set profile=` (bukti langkah kill K9 hadir); `Translate(create)` tidak menyertakan `rate-limit`.
- [ ] Tidak ada test yang menuliskan/log-kan password plaintext.

**Files likely touched:** `internal/usecase/business/manage_subscription_test.go`, `internal/adapter/postgres/subscription_repository_test.go`, `internal/adapter/http/subscription_handler_test.go`, `internal/testutil/db.go` (bila butuh helper).

**Dependencies:** Task 2–5.

**Estimated scope:** Large.

## Migrasi Database

Tidak ada perubahan skema. Issue ini hanya memakai tabel yang sudah ada: `subscriptions` (000009), `subscription_status_history` (000010), `provisioning_sync_log` (000011), `plan_router_profiles` (000006). Tidak ada kolom baru — kolom `onu_pon_port`/`onu_id`/`genieacs_device_id` adalah milik issue lain (05/08) dan tidak dipakai di sini. Karena tidak ada perubahan skema, tidak ada update yang diperlukan ke `DATABASE-SCHEMA.md` selain memastikan §7.2 (peristiwa `create`/`change_profile`) tetap konsisten dengan implementasi.

## Open Questions

- **Auto-isolir (auto-suspend) kolom subscription.** Isolir otomatis saat invoice jatuh tempo bukan lingkup issue ini (transisi status otomatis + kill sesi = Issue 04, dipicu **Issue 14 — Auto-Suspend & Auto-Restore Scheduler**). Kolom terkait (`auto_suspend_enabled`, `isolir_day`, `grace_period_days`) **milik Issue 14** (migrasi 000037), bukan issue ini — dicatat di sini hanya sebagai silang-rujuk agar tidak ditambahkan diam-diam ke migrasi Issue 03.

## Verification

- [ ] `go build ./...` sukses.
- [ ] `go test ./internal/usecase/business/...` hijau (table-driven provision/activate/change-plan).
- [ ] `go test ./internal/adapter/http/...` hijau (httptest handler + RBAC).
- [ ] `go test ./internal/adapter/postgres/...` hijau (testcontainers, verifikasi sync_log + status_history).
- [ ] `make lint` (golangci-lint) bersih — perhatikan boundary: `usecase/business` tidak impor `driver/`.
- [ ] Smoke test manual (butuh Issue 01 aktif + MikroTik CHR di GNS3): buat subscription PPPoE via CRUD → `curl -X POST .../subscriptions/:id/provision` dengan JWT role `teknisi`, harap **202** + `sync_log_id` → poll `curl .../sync-logs/:id` (Issue 01) sampai `success` → verifikasi `/ppp secret print` di CHR memunculkan username tsb → `curl -X POST .../subscriptions/:id/activate` → subscription `active`.
- [ ] Smoke test negatif: `curl` provision tanpa `plan_router_profiles` `synced` → **409** dengan pesan jelas; `curl` provision dengan JWT role `staff` → **403**.

## Definition of Done

- [ ] Aksi provision menulis tepat satu baris `provisioning_sync_log` `mikrotik_ppp_secret`/`create` dan **tidak** memanggil driver langsung (K4 terpenuhi).
- [ ] Ganti paket menulis baris `change_profile` **dan** menghasilkan sekuens driver yang memutus sesi PPPoE aktif (`/ppp active remove`, K9); activate menulis `subscription_status_history` dengan `changed_by_actor_type` benar.
- [ ] Guard status ditegakkan (hanya `pending_install`/`suspended`), validasi profil menghasilkan 409 yang jelas.
- [ ] Password PPPoE hanya didekripsi di Sync Engine via `port.CredentialVault`; tidak ada plaintext di audit log/aplikasi log/test.
- [ ] Endpoint mengikuti konvensi status (202 + `sync_log_id` untuk aksi perangkat) dan RBAC (`configs/rbac_policy.csv` diperbarui).
- [ ] Semua item Verification hijau; satu issue = satu PR.
