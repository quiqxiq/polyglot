# Issue 14: Auto-Suspend & Auto-Restore Scheduler (Billing-Driven)

## Konteks

Issue 04 menyediakan usecase suspend/resume/terminate yang **dipanggil manual**
lewat REST. Tetapi di operasi ISP nyata, isolir pelanggan telat bayar terjadi
**otomatis dan terjadwal**, bukan diklik satu per satu. Kedua repo referensi
punya cron harian untuk ini (bukti: `REFERENCES.md` §B, "Auto-suspend trigger 2
model"):

- **`gembok-bill`** — cron ambil `getOverdueInvoices()`, hitung `daysOverdue`,
  suspend bila `daysOverdue >= grace_period_days` (default 7), skip bila sudah
  `suspended` atau `auto_suspension=0`. Model **grace-period**: N hari sejak
  tanggal jatuh tempo invoice.
- **`billing-rtrw`** — cron harian jam 02:00, suspend bila `today >=
  customer.isolate_day` (**tanggal-isolir tetap** per-pelanggan, default 10) dan
  `auto_isolate` aktif. Model **tanggal tetap**: pada tanggal-X tiap bulan.

Keduanya memakai **flag per-pelanggan** untuk opt-out, dan **idempoten** (skip
yang sudah suspended). Simetrisnya: saat pelanggan **lunas**, layanan
**dipulihkan otomatis** — `billing-rtrw` hanya restore bila **tidak ada invoice
unpaid tersisa** (bukan sekadar satu pembayaran masuk).

Issue ini membangun scheduler itu di Polyglot: sebuah usecase bisnis yang
menentukan **SIAPA** yang harus di-isolir/dipulihkan berdasarkan status billing,
lalu memanggil **usecase Issue 04** dengan `actor_type='system_scheduled'`. Issue
ini **tidak** mengeksekusi command perangkat sendiri — seluruh cascade
provisioning (kill sesi, isolir profil, address-list) tetap milik Issue 04 →
Sync Engine (Issue 01) → `command_audit_log` (K4). Dengan begitu isolir manual
dan otomatis memakai jalur audit yang sama persis, hanya beda `actor_type`.

**Dua model trigger didukung, dipilih per-subscription** (bukan salah satu
dipaksakan): default global **grace-period** (dari config), dan override
per-subscription **tanggal-isolir tetap** (`isolir_day`) bila diisi. Ini
menutup paritas kedua repo tanpa memaksa operator memilih satu gaya.

## Prasyarat

- **Issue 04 (Suspend/Resume/Terminate Cascade)** — wajib. Issue ini adalah
  *pemicu terjadwal* yang memanggil usecase suspend & resume Issue 04; seluruh
  logika cascade + kill sesi + audit ada di sana. Tanpa Issue 04, tak ada yang
  dipanggil.
- **Foundation billing** — repository `invoices` (`internal/adapter/postgres/
  invoice_repository.go`, migrasi 000014) dan domain `billing.Invoice` sudah ada;
  dipakai untuk query invoice `overdue`/unpaid.
- **Foundation subscription** — repo + domain `subscription.Subscription`
  (migrasi 000009).
- **Payment confirmation** — untuk auto-restore jalur event (§Task 5). Bila
  mekanisme konfirmasi pembayaran belum ada, jalur event dilewati dan hanya
  sweep terjadwal yang jalan; nyatakan sebagai dependency lunak.
- **Mekanisme scheduler in-process** — lihat §Keputusan Scheduler; konsisten
  dengan pilihan Issue 01 (ticker + errgroup di `main.go`), bukan library baru
  kecuali diputuskan lain di Open Questions README.
- **Foundation Phase 4/5** — router Gin, middleware `AuthRequired`/`RBACRequired`.

## Keputusan Scheduler

Konsisten dengan Open Question #1 README (ticker sederhana dulu, bukan library
cron), auto-suspend & auto-restore dijalankan sebagai **sweep harian** dari
ticker in-process di `cmd/server/main.go`, pada jam yang dapat dikonfigurasi
(default 02:00 waktu lokal — meniru `billing-rtrw`). Alasan sweep harian (bukan
event-driven murni): keanggotaan "overdue" berubah seiring **berlalunya waktu**
(tanggal jatuh tempo terlewati), bukan hanya saat ada aksi — jadi harus ada
proses yang bangun tiap hari mengevaluasi ulang. Auto-restore punya **dua
pemicu**: event (saat pembayaran dikonfirmasi — latensi rendah) **dan** sweep
harian sebagai jaring pengaman bila event terlewat. Bila kelak job terjadwal
makin banyak (GenieACS sync Issue 08, RX poll Issue 10, voucher expiry Issue 13),
pertimbangkan mengangkat scheduler jadi komponen bersama — dicatat sebagai Open
Question, **bukan** dibangun diam-diam di issue ini.

## Ruang Lingkup

**In scope:**
- Kolom konfigurasi auto-suspend per-subscription (migrasi 000037).
- Konfigurasi default global (grace-period-days, jam sweep, model default) di
  `internal/config`.
- Query kandidat: `FindAutoSuspendCandidates(now)` dan
  `FindAutoRestoreCandidates(now)` (join `invoices` × `subscriptions`).
- Usecase `RunAutoSuspendSweep(ctx, now)` — iterasi kandidat, skip idempoten yang
  sudah `suspended`, panggil usecase suspend Issue 04 dengan
  `actor_type='system_scheduled'` + `reason` deskriptif.
- Usecase `RunAutoRestoreSweep(ctx, now)` + hook event pada konfirmasi pembayaran
  — pulihkan hanya bila **tidak ada invoice unpaid tersisa**.
- Penjadwalan sweep harian di `main.go` (ticker) + endpoint trigger manual.
- Endpoint observabilitas: kandidat (dry-run "siapa yang akan di-isolir hari
  ini") + kebijakan per-subscription.
- Audit: setiap transisi otomatis tercatat di `subscription_status_history`
  dengan `changed_by_actor_type='system_scheduled'` (kolom sudah ada, migrasi
  000010) — dan cascade perangkatnya di `command_audit_log` via Issue 04/01.

**Out of scope:**
- Eksekusi command perangkat (kill sesi, isolir profil, address-list) — **milik
  Issue 04**; issue ini hanya menentukan siapa + memanggil usecase-nya.
- Pembuatan invoice / billing run bulanan — domain billing terpisah.
- Gateway pembayaran & webhook-nya — issue ini hanya *meng-consume* event
  "pembayaran dikonfirmasi" bila tersedia (auto-restore), tidak membangunnya.
- Notifikasi WA/Telegram "layanan Anda diisolir" — out of scope semua issue
  provisioning.
- Denda/late-fee otomatis — domain billing.

## REST API

Base path `/api/v1/`. Endpoint baca = 200; trigger sweep manual = **202
Accepted** (memicu proses yang menulis banyak baris `sync_log` lewat Issue 04).
Perubahan kebijakan per-subscription = perubahan DB murni = 200.

| Method | Path | Tujuan | Role minimum |
|---|---|---|---|
| GET | `/api/v1/billing/suspension-policy` | Baca kebijakan global efektif (grace-period, jam sweep, model default) | staff |
| GET | `/api/v1/subscriptions/:id/suspension-policy` | Baca kebijakan auto-suspend satu subscription | staff |
| PUT | `/api/v1/subscriptions/:id/suspension-policy` | Set override per-subscription (`auto_suspend_enabled`, `isolir_day`, `grace_period_days`) | admin |
| GET | `/api/v1/billing/auto-suspend/candidates` | Dry-run: daftar subscription yang JATUH tempo untuk di-isolir per `as_of` | staff |
| POST | `/api/v1/billing/auto-suspend/run` | Jalankan sweep auto-suspend sekarang (manual) | admin |
| GET | `/api/v1/billing/auto-restore/candidates` | Dry-run: daftar subscription tersuspend yang sudah lunas & layak dipulihkan | staff |
| POST | `/api/v1/billing/auto-restore/run` | Jalankan sweep auto-restore sekarang (manual) | admin |

### GET `/api/v1/billing/suspension-policy`
- **Response (200):** objek kebijakan global efektif dari config: `grace_period_days`
  (int, default 7), `sweep_hour_local` (0–23, default 2), `default_trigger_model`
  (`grace_period`), plus catatan bahwa nilai per-subscription bisa meng-override.
- **Gagal:** 401/403.

### GET / PUT `/api/v1/subscriptions/:id/suspension-policy`
- **GET Response (200):** `auto_suspend_enabled` (bool), `isolir_day` (int|null,
  1–28), `grace_period_days` (int|null — override; null = pakai global),
  `effective_trigger_model` (`fixed_date` bila `isolir_day` terisi, else
  `grace_period`).
- **PUT Request:** subset field di atas. Validasi: `isolir_day` ∈ [1,28] (hindari
  29–31 yang tak ada di semua bulan — dokumentasikan pembulatan); `grace_period_days`
  ≥ 0. Set `isolir_day=null` untuk kembali ke model grace-period.
- **PUT Response (200):** kebijakan efektif setelah update.
- **Gagal:** 404 subscription tak ada; 400 nilai invalid; 401/403 (PUT butuh admin).

### GET `/api/v1/billing/auto-suspend/candidates`
- **Request:** query `as_of` (tanggal RFC3339, default hari ini — untuk simulasi),
  paginasi.
- **Response (200):** array subscription yang MEMENUHI syarat isolir per `as_of`
  (status `active`, `auto_suspend_enabled=true`, dan lolos aturan trigger di
  §Task 3), tiap item: `subscription_id`, `customer_id`, `trigger_model` yang
  memutuskan, invoice/tanggal pemicu, `days_overdue`. **Tidak** mengubah apa pun
  (dry-run). Berguna untuk preview sebelum `run`.
- **Gagal:** 400 `as_of` invalid; 401/403.

### POST `/api/v1/billing/auto-suspend/run`
- **Request:** body opsional `as_of` (default now), `dry_run` (bool, default
  false — bila true, sama dengan endpoint candidates tapi via POST).
- **Perilaku:** untuk tiap kandidat, panggil usecase suspend Issue 04
  (`actor_type='system_scheduled'`, `reason` mis. `auto-isolir: overdue N hari`).
  Idempoten: subscription yang sudah `suspended` dilewati. Kegagalan satu
  subscription tidak menggagalkan sweep (per-item, dicatat).
- **Response (202):** ringkasan `{evaluated, suspended, skipped, failed}` + daftar
  `subscription_id` yang di-suspend beserta `sync_log_ids`-nya (dari Issue 04).
- **Gagal:** 400; 401/403 (admin).

### GET / POST `/api/v1/billing/auto-restore/*`
- Simetris dengan auto-suspend. **Candidates**: subscription `suspended` yang
  penyebab suspend-nya billing (lihat §Task 5) dan **tidak punya invoice unpaid
  tersisa**. **Run**: panggil usecase resume Issue 04 (`system_scheduled`,
  `reason='auto-restore: lunas'`). Response 202 ringkasan serupa.

## Tasks

**Task 1: Migrasi 000037 — kolom auto-suspend per-subscription**

**Description:** Tambah kolom konfigurasi auto-suspend ke `subscriptions`.

**Acceptance criteria:**
- [ ] Migrasi `000037_add_auto_suspend_config_to_subscriptions.up.sql` + `.down.sql`
  (berpasangan) menambah: `auto_suspend_enabled` BOOLEAN NOT NULL DEFAULT true;
  `isolir_day` INTEGER nullable (dokumentasikan constraint 1–28 via CHECK);
  `grace_period_days` INTEGER nullable (override per-subscription; null = pakai
  global). Nomor 000037 sesuai tabel reservasi README §K6.
- [ ] `down.sql` men-drop ketiga kolom.
- [ ] Domain `subscription.Subscription` + repo mapping `subscription_repository.go`
  diperluas dengan field baru.
- [ ] `DATABASE-SCHEMA.md` §6.1 dicerminkan pada PR yang sama (K6).

**Files likely touched:** `migrations/000037_add_auto_suspend_config_to_subscriptions.up.sql`
+ `.down.sql`, `internal/domain/subscription/subscription.go`,
`internal/adapter/postgres/subscription_repository.go`, `DATABASE-SCHEMA.md`.

**Dependencies:** —

**Estimated scope:** Small

---

**Task 2: Konfigurasi default global**

**Description:** Tambah default global auto-suspend ke `internal/config/config.go`.

**Acceptance criteria:**
- [ ] `Config` menambah field: `AutoSuspendGracePeriodDays` (default 7),
  `AutoSuspendSweepHourLocal` (default 2), `AutoSuspendEnabled` (master switch,
  default true — bila false, sweep tak dijadwalkan sama sekali).
- [ ] Dibaca dari env dengan default aman; divalidasi (grace ≥ 0, hour 0–23).
- [ ] Unit test loader mencakup default & override.
- [ ] Global policy bersifat config (bukan tabel) untuk fase ini — bila runtime
  editing global diinginkan kelak, itu Open Question (tabel settings), jangan
  dibangun sekarang.

**Files likely touched:** `internal/config/config.go`, `internal/config/config_test.go`.

**Dependencies:** —

**Estimated scope:** Small

---

**Task 3: Query kandidat auto-suspend (repo)**

**Description:** Tambah method repo yang mengembalikan subscription yang memenuhi
syarat isolir per waktu acuan, dengan logika DUA model trigger.

**Acceptance criteria:**
- [ ] Method (mis. di `port.SubscriptionRepository` atau repo query khusus)
  `FindAutoSuspendCandidates(ctx, asOf time.Time) ([]subscription.Subscription, error)`
  mengembalikan subscription yang: `status='active'` **dan** `auto_suspend_enabled=true`
  **dan** memenuhi salah satu model:
  - **grace_period** (bila `isolir_day` null): ada invoice pelanggan berstatus
    unpaid (`status IN ('issued','overdue','partially_paid')`) dengan
    `due_date + effective_grace_days <= asOf`, di mana `effective_grace_days` =
    `subscriptions.grace_period_days` bila terisi, else global.
  - **fixed_date** (bila `isolir_day` terisi): `day_of_month(asOf) >= isolir_day`
    **dan** ada invoice unpaid untuk periode berjalan. Dokumentasikan penanganan
    `isolir_day` > jumlah hari bulan (pakai hari terakhir bulan).
- [ ] Query efisien (join + index pada `invoices.status`/`due_date`,
  `subscriptions.status`); sebut index yang diperlukan (boleh migrasi index kecil
  bila perlu — ambil nomor setelah 000037, daftarkan ke README §K6).
- [ ] `FindAutoRestoreCandidates(ctx, asOf)` mengembalikan subscription
  `status='suspended'` yang disebabkan billing (lihat Task 5) dan **tidak** punya
  invoice unpaid tersisa (semua `paid`/`void`).
- [ ] Uji dengan `testcontainers-go`: skenario grace-period (tepat di batas,
  sebelum, sesudah), fixed_date (tanggal 9 vs 10 vs 11 dengan isolir_day=10),
  bulan pendek (isolir_day=31 di Februari), opt-out `auto_suspend_enabled=false`,
  sudah suspended (tidak muncul), restore (lunas penuh vs sisa unpaid).

**Files likely touched:** `internal/port/subscription_repository.go` (atau file
query baru), `internal/adapter/postgres/subscription_repository.go`, test terkait.

**Dependencies:** Task 1, Task 2

**Estimated scope:** Medium

---

**Task 4: Usecase RunAutoSuspendSweep**

**Description:** Orkestrasi bisnis yang mengiterasi kandidat dan memanggil usecase
suspend Issue 04.

**Acceptance criteria:**
- [ ] `internal/usecase/business/auto_suspend.go` berisi
  `RunAutoSuspendSweep(ctx, asOf time.Time, dryRun bool)` yang mengambil kandidat
  (Task 3) dan, untuk tiap subscription, memanggil **usecase suspend Issue 04**
  dengan `actor_type='system_scheduled'` dan `reason` deskriptif (mis.
  `auto-isolir: overdue <N> hari` atau `auto-isolir: lewat tanggal <isolir_day>`).
- [ ] Idempoten: subscription yang sudah `suspended` dilewati (juga dijaga di
  Task 3 query, tapi cek ulang untuk race antar-hari).
- [ ] `asOf` disuntik sebagai parameter (deterministik untuk test); `dryRun=true`
  mengembalikan kandidat tanpa memanggil suspend.
- [ ] Ketahanan: kegagalan satu subscription di-log dan di-skip, sweep lanjut
  (tidak fatal) — kembalikan ringkasan `{evaluated, suspended, skipped, failed}`.
- [ ] Usecase ini **tidak** menyentuh `internal/driver/` — hanya memanggil usecase
  Issue 04 (yang menulis `sync_log`); boundary K1/K4 terjaga.
- [ ] Table-driven test dengan fake suspend-usecase + fake repo.

**Files likely touched:** `internal/usecase/business/auto_suspend.go`, test terkait.

**Dependencies:** Task 3, Issue 04

**Estimated scope:** Medium

---

**Task 5: Usecase RunAutoRestoreSweep + hook event pembayaran**

**Description:** Pemulihan otomatis saat lunas, dua pemicu (event + sweep).

**Acceptance criteria:**
- [ ] `internal/usecase/business/auto_restore.go`
  `RunAutoRestoreSweep(ctx, asOf)` mengambil kandidat restore (Task 3) dan
  memanggil **usecase resume Issue 04** (`actor_type='system_scheduled'`,
  `reason='auto-restore: lunas'`).
- [ ] Hook event: sediakan fungsi `OnPaymentConfirmed(ctx, customerID)` yang,
  saat dipanggil dari alur konfirmasi pembayaran (bila ada), mengevaluasi
  subscription pelanggan itu dan me-resume yang **tidak lagi punya invoice
  unpaid**. Bila alur pembayaran belum ada, hook ini tetap didefinisikan tapi
  belum dipanggil — sweep harian menutup gap.
- [ ] Syarat restore = **tidak ada invoice unpaid tersisa** (bukan sekadar satu
  pembayaran) — cegah restore prematur pelanggan yang masih nunggak sebagian.
- [ ] Hanya me-resume subscription yang suspend-nya **karena billing**. Cara
  membedakan: cek `subscription_status_history` entri terakhir `→ suspended`
  dengan `changed_by_actor_type='system_scheduled'` (atau reason ber-prefix
  `auto-isolir`), supaya isolir manual (mis. pelanggaran) TIDAK ikut ter-restore
  otomatis oleh pembayaran. Dokumentasikan aturan ini.
- [ ] Idempoten + ketahanan per-item seperti Task 4.
- [ ] Table-driven test: lunas penuh → restore; sisa unpaid → tidak; suspend
  manual (`human`) → tidak ikut restore.

**Files likely touched:** `internal/usecase/business/auto_restore.go`, test terkait.

**Dependencies:** Task 3, Issue 04

**Estimated scope:** Medium

---

**Task 6: Penjadwalan sweep harian + trigger manual**

**Description:** Pasang sweep harian di composition root, plus jalur pemicu manual.

**Acceptance criteria:**
- [ ] `cmd/server/main.go` menjadwalkan `RunAutoSuspendSweep` dan
  `RunAutoRestoreSweep` sekali sehari pada `AutoSuspendSweepHourLocal` (ticker
  in-process, konsisten mekanisme Issue 01; goroutine di bawah errgroup/cancel,
  variabel loop dilewatkan eksplisit). Bila `AutoSuspendEnabled=false`, tidak
  dijadwalkan.
- [ ] `asOf` = waktu jalan aktual (sweep) atau dari request (manual/simulasi).
- [ ] Sweep tahan-ulang: dua kali jalan di hari sama tidak menghasilkan efek
  ganda (idempotensi Task 4/5).
- [ ] Graceful shutdown membatalkan sweep yang sedang berjalan lewat `ctx`.

**Files likely touched:** `cmd/server/main.go`, (mungkin) `internal/config/config.go`.

**Dependencies:** Task 4, Task 5

**Estimated scope:** Small

---

**Task 7: Handler REST + RBAC**

**Description:** Endpoint di §REST API dengan DTO, validasi, error→status, policy.

**Acceptance criteria:**
- [ ] Semua endpoint §REST API terpasang dengan `AuthRequired`+`RBACRequired`
  sesuai role minimum; DTO di `internal/adapter/http/dto/`.
- [ ] `run` mengembalikan 202 + ringkasan; `candidates` 200 (dry-run, tak
  mengubah); PUT policy 200; validasi `isolir_day`/`grace_period_days` → 400.
- [ ] Baris policy ditambahkan ke `configs/rbac_policy.csv` (baca: staff; run &
  PUT policy: admin).
- [ ] Handler diuji dengan `httptest` + usecase di-stub.

**Files likely touched:** `internal/adapter/http/billing_handler.go` (atau
`auto_suspend_handler.go`), `internal/adapter/http/dto/suspension_policy.go`,
`internal/adapter/http/router.go`, `configs/rbac_policy.csv`, test terkait.

**Dependencies:** Task 4, Task 5

**Estimated scope:** Medium

---

## Migrasi Database

Satu migrasi baru (nomor dari tabel reservasi README §K6):

- **`000037_add_auto_suspend_config_to_subscriptions`** (up/down berpasangan) —
  ALTER `subscriptions` tambah `auto_suspend_enabled` (BOOLEAN NOT NULL DEFAULT
  true), `isolir_day` (INTEGER nullable, CHECK 1–28), `grace_period_days` (INTEGER
  nullable). Cermin ke `DATABASE-SCHEMA.md` §6.1.

Bila query kandidat (Task 3) menuntut index tambahan pada `invoices(status,
due_date)`, ambil nomor migrasi berikutnya **setelah 000037** dan tambahkan ke
tabel reservasi README §K6 pada PR yang sama. Kolom `subscription_status_history`
(actor_type) sudah ada (000010) — tidak ada perubahan skema audit.

## Verification

- [ ] `go build ./...` sukses.
- [ ] `go test ./internal/adapter/postgres/... -run AutoSuspend` (query kandidat,
  testcontainers) lulus — termasuk kasus batas tanggal & bulan pendek.
- [ ] `go test ./internal/usecase/business/... -run 'AutoSuspend|AutoRestore'`
  lulus (idempoten, per-item failure, manual-suspend tak ter-restore).
- [ ] `go test ./internal/adapter/http/... -run Suspension` lulus.
- [ ] `make lint` bersih; migrasi 000037 naik-turun bersih di DB kosong.
- [ ] Smoke test manual (curl, sebutkan sebagai perintah): buat subscription
  active + invoice overdue → `GET /billing/auto-suspend/candidates` menampilkannya
  → `POST /billing/auto-suspend/run` → subscription jadi `suspended`, ada baris
  `sync_log` (via Issue 04) dan `subscription_status_history`
  (`system_scheduled`) → tandai invoice `paid` → `GET /billing/auto-restore/
  candidates` menampilkannya → `POST .../auto-restore/run` → kembali `active`.

## Definition of Done

- [ ] Migrasi 000037 selesai & tercermin di `DATABASE-SCHEMA.md`; config default
  global ada.
- [ ] Dua model trigger (grace-period default + fixed-date override
  per-subscription) berfungsi dan teruji, termasuk kasus tanggal batas.
- [ ] Auto-suspend & auto-restore memanggil usecase Issue 04 dengan
  `actor_type='system_scheduled'` — tidak menyentuh driver langsung; seluruh
  cascade + audit lewat Issue 04/01 (K4).
- [ ] Idempoten (skip sudah-suspended / masih-nunggak); auto-restore hanya untuk
  isolir yang penyebabnya billing, bukan suspend manual.
- [ ] Sweep harian terjadwal di `main.go` + trigger manual + dry-run candidates.
- [ ] Endpoint REST + RBAC lengkap dan teruji; `go test ./...` & `make lint` hijau.
