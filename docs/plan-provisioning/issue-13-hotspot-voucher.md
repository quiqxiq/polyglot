# Issue 13: Hotspot & Voucher Lifecycle

## Konteks

Issue 03 hanya menangani provisioning PPPoE (`/ppp secret`). Layanan **hotspot** —
tulang punggung model RT/RW-Net dan voucher — punya siklus hidup yang berbeda
total dan belum tercakup di rencana mana pun. Skema pendukungnya **sudah ada**:
`voucher_batches` + `hotspot_vouchers` (migrasi 000007) dan `target_type=
'mikrotik_hotspot_user'` di `provisioning_sync_log`, tetapi tidak ada satu pun
usecase yang mengisinya.

Keputusan mekanisme (dikonfirmasi pemilik produk):

1. **Hotspot user (`/ip hotspot user`), bukan static IP layer-3.** Layanan
   hotspot memakai `/ip hotspot user` + `/ip hotspot user profile`. `service_type=
   'static_ip'` yang terpisah adalah hal berbeda. (Catatan: "IP binding" di poin 5
   adalah soal identitas di sisi MODEM untuk monitoring, bukan pengganti hotspot
   user.)
2. **Masa aktif (validity) configurable per plan, granularitas sub-hari.** Tiga
   model didukung, dipilih per plan: (a) dari login pertama (`activated_at` + N
   **satuan** → `expires_at`), (b) kuota waktu (`limit-uptime` native MikroTik),
   (c) dari tanggal pembuatan voucher. Durasi **bukan** integer hari — paket riil
   RT/RW-Net dijual per **jam** dan ada voucher <1 hari (REFERENCES §E: `gembok-bill`
   `voucher_pricing.duration` in hours + `duration_type`). Karena itu validity
   disimpan sebagai pasangan value+unit (`minutes`/`hours`/`days`), bukan
   `_days`. Kolom `hotspot_vouchers.activated_at`/`expires_at`/`status` yang sudah
   ada memang didesain untuk ini.
3. **Enforcement hybrid.** Generate + provisioning voucher dilakukan lewat jalur
   Golang teraudit (`provisioning_sync_log` → Sync Engine → `command_audit_log`).
   Penegakan expiry & cleanup diserahkan **juga** ke `/system scheduler` +
   script `on-login` di router untuk ketahanan saat backend Golang mati. Karena
   router mengubah state di luar pipeline audit, issue ini **wajib** menyertakan
   pass rekonsiliasi (§Tasks) dan ADR yang mendokumentasikan deviasi sadar ini.
4. **Pemutusan = disable/remove user + selalu kill sesi aktif (sesuai K9).**
   Fakta MikroTik yang divalidasi 5 repo (REFERENCES §A/§E): `disable`/`remove`
   user hotspot **tidak** memutus sesi yang sedang online — kill adalah **command
   sentinel terpisah**, bukan efek samping. Sekuensnya **berurutan**: `[disable|
   remove]` **lalu** `[/ip hotspot active remove]`. Referensi justru sering
   melupakan langkah kedua (bug diam-diam) — kita perbaiki, bukan tiru. Suspend →
   `disable` + kill; terminate/expired → `remove` + kill. Untuk **mode A**
   (`mac_login`, mac-cookie), kill saja tidak cukup: MAC ber-cookie auto-relogin
   tanpa kredensial, jadi sekuens **wajib** ditutup `/ip hotspot cookie remove`
   untuk MAC/user tsb setelah `active/remove` (lihat Task 5, K9/K13).
5. **Identitas modem untuk monitoring online/offline — dukung dua mode
   per-subscription (A+B).** Masalah inti: berbeda dari PPPoE (identitas melekat
   pada sesi tunnel, terpantau lewat stream `/ppp/active/print follow`), hotspot
   subscription yang polos berjalan di atas DHCP tanpa identitas stabil sehingga
   **tidak bisa dipantau seperti PPPoE**. Supaya paritas monitoring tercapai,
   subscription hotspot **wajib** punya salah satu identitas modem berikut,
   disimpan di kolom baru `subscriptions.hotspot_access_mode`:
   - **Mode A — `mac_login`:** hotspot user dengan auto-login by MAC. Sesi
     persisten muncul di `/ip hotspot active`; connect/disconnect di-stream
     langsung via **`/ip/hotspot/active/print follow`** (record `!re` masuk =
     connect, `.dead=yes` = disconnect). Dipantau oleh Issue 12 dengan **mesin
     yang sama** seperti PPPoE — hanya beda command tabel active, bukan `/log`.
     Wajib set `keepalive-timeout` + `mac-cookie` di profile agar sesi stabil
     (paritas ~90% PPPoE; sesi hotspot lebih rentan drop idle daripada tunnel
     PPPoE).
   - **Mode B — `ip_binding`:** MAC di-bind ke IP tetap via
     `/ip hotspot ip-binding type=bypassed` (bypass portal). Karena tak pernah
     "login", ketiadaan sesi ditutup dengan memasang `/tool netwatch` pada IP
     tetap itu; status host up/down di-stream via **`/tool/netwatch/print
     follow`**. Jadi mode B **tetap event-driven, bukan polling dari Golang**
     (konsisten ADR 0003) — Issue 12 memparser stream netwatch alih-alih hotspot
     active.
   Kedua mode berujung ke satu mesin monitoring (Issue 12), sehingga tidak ada
   engine monitoring kedua. Mode dipilih per subscription saat provisioning.

Konteks perbandingan referensi: pola script `on-login`/scheduler router ini
persis yang dipakai Mikhmon v3 yang teruji lapangan; kita mengadopsinya **hanya**
sebagai jaring pengaman redundan di atas jalur Golang, bukan sebagai sumber
kebenaran (lihat `ANALISIS-PROVISIONING-REPO-REFERENSI.md` untuk posisi "no
polling" dan jalur audit tunggal yang tetap dipertahankan).

## Prasyarat

- **Issue 01 (Sync Engine)** — semua provisioning voucher menulis
  `provisioning_sync_log` yang dikonsumsi Sync Engine. Akar dependency.
- **Issue 02 (Plan Router Profile Sync)** — profil hotspot (`/ip hotspot user
  profile`) disinkronkan lewat mekanisme `plan_router_profiles` yang sama dengan
  `/ppp profile`; issue ini menambah `target_type='mikrotik_hotspot_profile'`.
- **Issue 12 (Subscriber Session Tracking)** — dua peran: (1) capture login
  pertama voucher (model validity "dari login pertama") via stream
  `/ip/hotspot/active/print follow`; (2) monitoring online/offline subscription
  hotspot — `/ip/hotspot/active/print follow` untuk mode A (`mac_login`) dan
  `/tool/netwatch/print follow` untuk mode B (`ip_binding`). Issue 13 memperluas
  scope Issue 12 ke dua sumber stream ini (lihat catatan silang di Issue 12).
  Tanpa Issue 12, validity model (a) dan monitoring hotspot tidak jalan.
- **Foundation:** router Gin, middleware auth/RBAC, CRUD plan (`plans`,
  migrasi 000005) dan CRUD subscription sudah ada.
- **Skema:** `voucher_batches` + `hotspot_vouchers` (migrasi 000007) sudah ada.

## Ruang Lingkup

**In scope:**
- Domain voucher/hotspot di `internal/domain/` (entity `Voucher`, `VoucherBatch`,
  enum status yang mencerminkan CHECK constraint DB).
- Kontrak `port.VoucherRepository` + implementasi Postgres (map migrasi 000007).
- Bulk generate voucher (`voucher_batches` → N `hotspot_vouchers`).
- Provisioning voucher ke router (`/ip hotspot user add`) lewat sync-log.
- Provisioning hotspot **terikat subscription** (pelanggan terdaftar,
  `service_type='hotspot'`) — dua alur: anonim/bulk vs terikat subscription.
- Identitas modem per subscription hotspot: mode A (`mac_login`, auto-login by
  MAC) dan mode B (`ip_binding` + pasang `/tool netwatch` pada IP tetap) — supaya
  Issue 12 bisa memantau online/offline. Kolom baru `subscriptions.hotspot_access_mode`.
- Sinkron profil hotspot per device (mapping `plans` → `/ip hotspot user profile`),
  termasuk `keepalive-timeout` + `mac-cookie` untuk mode A.
- Konfigurasi validity per plan (3 model) + kolom baru di `plans`.
- Capture login pertama (via Issue 12) → set `activated_at`/`expires_at`.
- Expiry enforcement hybrid: scheduler Golang (primer, teraudit) + generator
  script router `on-login`/scheduler (jaring pengaman offline).
- Pass rekonsiliasi state router ↔ DB (menutup lubang audit akibat hybrid).
- Disable/enable/remove voucher, selalu dengan kill sesi aktif.
- Revoke batch (disable/remove seluruh batch sekaligus).

**Out of scope:**
- IP binding `blocked` (blokir device by MAC) — mode B di sini hanya memakai
  `bypassed` untuk identitas monitoring; blokir adalah aksi suspend terpisah.
- FreeRADIUS-backed hotspot — jalur `target_type='freeradius'` sudah dipetakan di
  DATABASE-SCHEMA §7.3, dikerjakan issue tersendiri.
- Penjualan/pembayaran voucher (billing) — domain billing terpisah.
- Notifikasi WA/Telegram kode voucher — out of scope semua issue provisioning.

## REST API

Base path `/api/v1/`. Aksi ke perangkat mengikuti konvensi 202 Accepted + id
`sync_log` (K4). Generate bulk mengembalikan 202 karena provisioning N user ke
router berjalan asinkron lewat Sync Engine.

| Method | Path | Tujuan | Role minimum |
|---|---|---|---|
| POST | `/api/v1/voucher-batches` | Generate bulk voucher + provision | admin |
| GET | `/api/v1/voucher-batches` | List batch | staff |
| GET | `/api/v1/voucher-batches/:id` | Detail batch + ringkasan status voucher | staff |
| POST | `/api/v1/voucher-batches/:id/revoke` | Disable/remove seluruh batch | owner |
| GET | `/api/v1/vouchers` | List voucher (filter batch/status/plan) | staff |
| GET | `/api/v1/vouchers/:id` | Detail satu voucher | staff |
| POST | `/api/v1/vouchers/:id/disable` | Suspend voucher (disable + kill sesi) | admin |
| POST | `/api/v1/vouchers/:id/enable` | Aktifkan kembali voucher | admin |
| DELETE | `/api/v1/vouchers/:id` | Hapus voucher (remove + kill sesi) | admin |
| POST | `/api/v1/subscriptions/:id/provision-hotspot` | Provision hotspot terikat pelanggan | admin |

**POST `/api/v1/voucher-batches`**
- Body penting: `plan_id` (UUID, wajib, plan `service_type='hotspot'`),
  `quantity` (int, wajib, cap maksimum mis. 5000 per batch — tolak di atas itu),
  `price_per_voucher` (numeric), `device_id` (UUID router target provisioning),
  opsional format kode: `code_length`, `code_prefix`, `charset`
  (`numeric`|`alpha`|`alphanumeric`, default alfanumerik tanpa karakter ambigu
  0/O/1/l), `uppercase` (bool), `password_same_as_username` (bool — voucher
  swipe-only di mana password = kode). Bila tak dikirim, default diambil dari
  konfigurasi generate kode di plan/setting (lihat Task 1/6). Terverifikasi di
  REFERENCES §E (`VoucherGenerator` setting `username_length`/`char_type`/`prefix`/
  `password_same`).
- Perilaku: buat satu baris `voucher_batches`, generate `quantity` baris
  `hotspot_vouchers` (`status='unused'`, `code` unik — keunikan dicek terhadap DB
  **dan** terhadap router via `/ip hotspot user print ?name` sebelum provisioning,
  bukan hanya UNIQUE DB, karena router bisa memuat user lintas-batch; regenerasi
  kode yang bentrok), lalu untuk tiap voucher tulis baris `provisioning_sync_log`
  (`target_type='mikrotik_hotspot_user'`, `action='create'`). Semua dalam satu
  transaksi bisnis; provisioning aktual ke router asinkron via Sync Engine.
- Response: `202 Accepted` — `batch_id`, `quantity_generated`, dan cara memantau
  progres (mis. `GET /voucher-batches/:id` menampilkan hitung `synced`/`pending`/
  `failed` dari `sync_log` terkait).
- Status: `202`; `400` bila `quantity` melebihi cap atau plan bukan hotspot;
  `404` bila plan/device tidak ada; `401/403`.

**GET `/api/v1/voucher-batches`** / **`/:id`**
- List: filter `plan_id`, rentang tanggal `generated_at`, paginasi standar.
  Detail: field batch (`quantity_generated`, `price_per_voucher`, `generated_by`,
  `generated_at`) + agregat status voucher anggotanya (unused/active/expired/used)
  dan agregat status provisioning (`pending`/`success`/`failed`).
- Status: `200`; `404` (detail); `401/403`.

**POST `/api/v1/voucher-batches/:id/revoke`**
- Body: `mode` (`disable` | `remove`, default `remove`), `reason`.
- Perilaku **by-status, jangan membabi-buta** (REFERENCES §E:
  `removehotspotuserbycomment.php` hanya hapus voucher `uptime=00:00:00`): pisahkan
  voucher batch per status sebelum menulis sync-log.
  - Voucher **`unused`** (`uptime=00:00:00`, belum pernah login) → `remove` aman,
    tak ada sesi/cookie untuk di-kill.
  - Voucher **`active`/`used`** → `disable`+kill sesi+cookie remove (mode A), atau
    `remove`+kill+cookie sesuai `mode` — tidak boleh langsung `remove` tanpa kill
    karena sesi online tetap jalan (K9).
  Tiap baris sync-log (`disable`/`delete`) dipasangkan kill sesi + cookie-remove
  oleh driver (K9). Update `hotspot_vouchers.status`.
- Response: `202` + daftar/`count` `sync_log` id. Status: `202`; `404`; `401/403`
  (role di bawah `owner` ditolak).

**GET `/api/v1/vouchers`** / **`/:id`**
- List: filter `batch_id`, `status`, `plan_id`, `code` (exact), `used_by_mac`,
  `used_by_subscription_id`, paginasi. Detail: seluruh kolom `hotspot_vouchers`
  + `sync_log` provisioning terakhir tertaut.
- Status: `200`; `404` (detail); `401/403`.

**POST `/api/v1/vouchers/:id/disable`** / **`/enable`**
- Body kosong (`disable` boleh `reason`). `disable` → sync-log `action='disable'`
  + kill sesi aktif; `enable` → `action='enable'`. Update `status`
  (`active`↔ ditandai non-aktif; jangan pakai nilai di luar enum
  `unused/active/expired/used`).
- Response: `202` + `sync_log` id. Status: `202`; `409` bila transisi status tak
  valid (mis. enable voucher `expired`/`used`); `404`; `401/403`.

**DELETE `/api/v1/vouchers/:id`**
- Sync-log `action='delete'` + kill sesi aktif; `status='used'`/dihapus sesuai
  kebijakan retensi. Response `202` + `sync_log` id. Status `202`; `404`;
  `401/403`.

**POST `/api/v1/subscriptions/:id/provision-hotspot`**
- Untuk pelanggan terdaftar dengan `service_type='hotspot'` (bukan voucher
  anonim). Body: `access_mode` (`mac_login` | `ip_binding`, wajib),
  `mac_address` (wajib untuk kedua mode — identitas modem), `static_ip` (wajib
  untuk `ip_binding`, diabaikan untuk `mac_login`), opsional `code`/`username`,
  `password`. Perilaku:
  - **`mac_login`:** buat/kaitkan `hotspot_vouchers` (`used_by_subscription_id`),
    tulis sync-log `mikrotik_hotspot_user` `action=create` (user auto-login by
    MAC, profile dengan `keepalive-timeout`+`mac-cookie`). Set
    `subscriptions.hotspot_access_mode='mac_login'` + `mac_address`.
  - **`ip_binding`:** tulis sync-log `mikrotik_ip_binding` `action=create`
    (entri `type=bypassed` MAC→`static_ip`) **dan** baris kedua untuk memasang
    `/tool netwatch` pada `static_ip` (target `mikrotik_ip_binding` juga, atau
    driver menyisipkan netwatch dalam sekuens command yang sama — lihat Task 5).
    Set `hotspot_access_mode='ip_binding'` + `mac_address` + `static_ip`.
- Selanjutnya suspend/resume/terminate mengikuti cascade Issue 04 (mendeteksi
  `service_type='hotspot'` + `hotspot_access_mode`).
- Response: `202` + daftar `sync_log` id. Status `202`; `400` bila `access_mode`
  tak valid atau `static_ip`/`mac_address` kurang; `409` bila profil hotspot
  device belum disinkron; `404`; `401/403`.

## Tasks

**Task 1: Domain voucher & batch**

**Description:** Buat domain hotspot voucher: entity `Voucher`, `VoucherBatch`,
enum status yang mencerminkan CHECK constraint DB, tanpa I/O.

**Acceptance criteria:**
- [ ] Struct `Voucher` memetakan kolom `hotspot_vouchers` (`ID`, `BatchID`
  nullable, `PlanID`, `Code`, `Status`, `UsedBySubscriptionID` nullable,
  `UsedByMAC` nullable, `ActivatedAt` nullable, `ExpiresAt` nullable, `CreatedAt`).
- [ ] Struct `VoucherBatch` memetakan `voucher_batches`.
- [ ] Konstanta `VoucherStatus` mencakup persis `unused`, `active`, `expired`,
  `used` — tidak menambah/mengurangi.
- [ ] Enum `ValidityModel` (`from_login`, `uptime_quota`, `from_creation`) +
  helper menghitung `expires_at` dari model + parameter plan berupa **value+unit**
  (`minutes`/`hours`/`days`), bukan integer hari — paket riil pakai jam & ada
  voucher <1 hari (REFERENCES §E). Fungsi murni, input waktu via parameter —
  jangan panggil `time.Now()` di domain bila mau deterministik untuk test.
- [ ] Helper transisi status valid (mis. `unused→active→expired`) menolak
  transisi ilegal.
- [ ] Konfigurasi generate kode sebagai value object domain (`CodeGenSpec`:
  `Length`, `Charset` = `numeric`/`alpha`/`alphanumeric`, `Prefix`, `Uppercase`,
  `PasswordSameAsUsername`) + generator kode murni yang deterministik bila sumber
  acak disuntik — dipakai usecase generate (Task 6). Field ini boleh berasal dari
  kolom plan/setting (lihat Task 2); domain hanya mendefinisikan bentuk & aturan
  validitasnya (charset non-ambigu default, panjang minimal). Bukti REFERENCES §E.
- [ ] Doc comment pada setiap identifier exported; tanpa import adapter/driver.

**Files likely touched:** `internal/domain/voucher/voucher.go`,
`internal/domain/voucher/batch.go`, `internal/domain/voucher/errors.go`.

**Dependencies:** —

**Estimated scope:** Medium

---

**Task 2: Kolom konfigurasi hotspot per plan (migrasi 000032)**

**Description:** Tambah kolom konfigurasi hotspot/validity ke tabel `plans` agar
model validity configurable per plan.

**Acceptance criteria:**
- [ ] Migrasi `000032_add_hotspot_config_to_plans.up.sql` + `.down.sql`
  (berpasangan) menambah kolom **nullable** ke `plans`: `hotspot_shared_users`
  (int), `hotspot_validity_model` (text, CHECK IN `from_login`/`uptime_quota`/
  `from_creation`), `hotspot_validity_value` (int) + `hotspot_validity_unit`
  (text, CHECK IN `minutes`/`hours`/`days`) — **bukan** `hotspot_validity_days`,
  karena paket riil pakai jam & voucher <1 hari (REFERENCES §E); domain
  mengonversi value+unit → durasi untuk `expires_at`/`session-timeout`/
  `limit-uptime`. Ditambah `hotspot_uptime_limit_seconds` (int),
  `hotspot_session_timeout_seconds` (int), dan kolom konfigurasi generate kode:
  `voucher_code_length` (int), `voucher_code_charset` (text, CHECK IN `numeric`/
  `alpha`/`alphanumeric`), `voucher_code_prefix` (text), `voucher_code_uppercase`
  (bool), `voucher_password_same_as_username` (bool). Semua nullable karena plan
  non-hotspot tak memakainya.
- [ ] `down.sql` men-drop kolom-kolom itu.
- [ ] Domain `plan.Plan` + repo mapping `internal/adapter/postgres/plan_repository.go`
  diperluas dengan field baru.
- [ ] `DATABASE-SCHEMA.md` §4.2 dicerminkan pada PR yang sama (K6).
- [ ] Nomor migrasi diambil dari tabel reservasi README §K6 (000032), bukan
  dipilih sendiri.

**Files likely touched:** `migrations/000032_add_hotspot_config_to_plans.up.sql`
+ `.down.sql`, `internal/domain/plan/plan.go`,
`internal/adapter/postgres/plan_repository.go`, `DATABASE-SCHEMA.md`.

**Dependencies:** Task 1

**Estimated scope:** Small

---

**Task 3: `target_type` hotspot/ip-binding + kolom `hotspot_access_mode` (migrasi 000033, 000034)**

**Description:** Perluas CHECK constraint `provisioning_sync_log.target_type`
dengan `mikrotik_hotspot_profile` dan `mikrotik_ip_binding`, serta tambah kolom
`subscriptions.hotspot_access_mode` untuk memilih identitas modem per subscription.

**Acceptance criteria:**
- [ ] Migrasi `000033_add_hotspot_target_types.up.sql` + `.down.sql` meng-`ALTER`
  CHECK `provisioning_sync_log.target_type` menambah `mikrotik_hotspot_profile`
  (sinkron `/ip hotspot user profile`) dan `mikrotik_ip_binding` (entri
  `type=bypassed` + netwatch mode B). `mikrotik_hotspot_user` sudah ada dari skema
  awal — jangan duplikasi.
- [ ] Migrasi `000034_add_hotspot_access_mode_to_subscriptions.up.sql` + `.down.sql`
  menambah kolom **nullable** `subscriptions.hotspot_access_mode` (text, CHECK IN
  `mac_login`/`ip_binding`) — nullable karena subscription non-hotspot tak
  memakainya. (`mac_address` dan `static_ip` sudah ada di skema `subscriptions`
  migrasi 000009 — pakai kolom itu, jangan tambah duplikat.)
- [ ] `down.sql` masing-masing mengembalikan skema (perhatikan: down CHECK hanya
  aman bila tak ada baris memakai nilai baru — dokumentasikan).
- [ ] Konstanta `TargetType` di domain provisioning (Issue 01) ditambah kedua
  nilai; domain `subscription.Subscription` + repo diperluas dengan
  `HotspotAccessMode`.
- [ ] `DATABASE-SCHEMA.md` §6.1 dan §6.3 dicerminkan.
- [ ] Nomor migrasi dari tabel reservasi README §K6 (000033, 000034).

**Files likely touched:**
`migrations/000033_add_hotspot_target_types.up.sql` + `.down.sql`,
`migrations/000034_add_hotspot_access_mode_to_subscriptions.up.sql` + `.down.sql`,
`internal/domain/provisioning/provisioning.go`,
`internal/domain/subscription/subscription.go`,
`internal/adapter/postgres/subscription_repository.go`, `DATABASE-SCHEMA.md`.

**Dependencies:** Issue 01 (domain provisioning), Issue 02 (pola ALTER target_type)

**Estimated scope:** Small

---

**Task 4: Repository voucher (Postgres)**

**Description:** Implementasi `port.VoucherRepository` + repo Postgres map ke
`hotspot_vouchers` & `voucher_batches` (migrasi 000007).

**Acceptance criteria:**
- [ ] `port.VoucherRepository` di `internal/port/voucher_repository.go` dengan
  method: `CreateBatch`, `CreateVouchers` (bulk insert), `FindVoucherByID`,
  `FindVoucherByCode`, `ListVouchers(filter,page)`, `ListBatches(filter,page)`,
  `FindBatchByID`, `MarkActivated(ctx, id, activatedAt, expiresAt)`,
  `UpdateStatus`, `FindBySubscription`, `FindExpiring(ctx, before)` (untuk
  scheduler).
- [ ] Compile-time assertion repo memenuhi interface.
- [ ] Bulk insert efisien (satu statement multi-row atau batch), aman terhadap
  tabrakan UNIQUE `code` (regenerasi kode yang bentrok).
- [ ] `context.Context` param pertama; `error` return terakhir; mapping tak
  membocorkan GORM ke domain.
- [ ] Uji dengan `testcontainers-go` (bukan mock) termasuk skenario bulk 1000
  voucher.

**Files likely touched:** `internal/port/voucher_repository.go`,
`internal/adapter/postgres/voucher_repository.go`,
`internal/adapter/postgres/voucher_repository_test.go`.

**Dependencies:** Task 1

**Estimated scope:** Medium

---

**Task 5: Penerjemahan command hotspot di driver MikroTik**

**Description:** Perluas `internal/driver/mikrotik/commands.go` dengan katalog
command hotspot: create/disable/enable/remove `/ip hotspot user`, sinkron
`/ip hotspot user profile`, dan **selalu** memasangkan `/ip hotspot active remove`
saat tujuan aksi memutus koneksi.

**Acceptance criteria:**
- [ ] `Translate` memetakan operasi abstrak hotspot → urutan command native:
  buat user (`/ip hotspot user add` dengan `name`, `password`, `profile`,
  opsional `limit-uptime` untuk model `uptime_quota`, dan `comment` ber-prefix
  kanonik K14 mis. `poly:vc:<id>,exp:<ts>` agar bisa di-parse balik oleh on-login
  guard & rekonsiliasi), disable, enable, remove.
- [ ] Aksi `disable`/`delete`/`suspend` menghasilkan urutan command **berurutan**
  `[disable|remove]` **lalu** kill sesi — bukan mengandalkan efek samping
  disable/remove (K9: disable/remove TIDAK memutus sesi online, kill = command
  sentinel terpisah). Kill **wajib loop SEMUA `.id`** hasil `/ip hotspot active
  print ?user=<code>` (bila `shared-users>1` satu kode bisa punya banyak sesi
  aktif), bukan hanya `.id` pertama — referensi punya bug ini (REFERENCES §A/§E:
  `mikrotikKickHotspotUser()` break setelah satu sesi). Dibuktikan di test: N sesi
  aktif → N baris `/ip hotspot active remove`, bukan cuma satu.
- [ ] **Mode A (`mac_login`, `add-mac-cookie`) — MUST-FIX:** sekuens suspend/
  terminate **wajib** ditutup `/ip hotspot cookie remove` untuk MAC/user tsb
  **setelah** `active/remove`. Tanpa itu MAC ber-cookie auto-relogin tanpa
  kredensial dan pemutusan tidak efektif (REFERENCES §A `hotspot-cookies.php`;
  K9/K13). Test membuktikan sekuens mode A memuat langkah cookie-remove.
- [ ] Sinkron profil hotspot: `/ip hotspot user profile add/set` dengan
  `rate-limit` dari `plans.bandwidth_*`/`burst_*`, `shared-users`,
  `session-timeout` dari kolom plan (Task 2), dan untuk mode A (`mac_login`)
  set `keepalive-timeout` + `add-mac-cookie`/`mac-cookie-timeout` agar sesi
  hotspot stabil dan pantauan online/offline tidak churn palsu.
- [ ] **Mode B (`ip_binding`)** — `Translate` untuk `mikrotik_ip_binding`
  `action=create` menghasilkan sekuens: tambah `/ip hotspot ip-binding`
  (`mac-address`→`address` static, `type=bypassed`) **lalu** pasang `/tool
  netwatch` pada IP itu (status host up/down, dipantau Issue 12 via
  `/tool/netwatch/print follow`). `action=delete`
  menghapus keduanya; `action=disable` (suspend mode B) mengubah binding jadi
  `type=blocked` (bukan hapus), `enable` mengembalikan ke `bypassed`.
- [ ] `Classify`: create/disable/enable read-vs-destructive sesuai konvensi vendor
  (remove & kill sesi = destruktif → jalur HITL bila kebijakan menuntut).
- [ ] **Keunikan kode vs router (correction §E):** `Execute` create user menangani
  `!trap`/`!re` RouterOS `already exists` sebagai sinyal tabrakan lintas-batch (user
  bisa sudah ada di router walau UNIQUE DB lolos) — dikembalikan sebagai error
  bertipe yang bisa dikenali usecase/Sync Engine untuk regenerate kode & retry,
  bukan ditelan diam-diam. Keunikan dicek ke router (`/ip hotspot user print
  ?name`) sebelum add.
- [ ] **Orphan cleanup pada remove (correction §E, K15):** operasi `remove` user
  hotspot untuk model `from_login` **wajib** membersihkan `/system scheduler remove
  [find name=<code>]` (dan `/system script` bila ada) **sebelum**/berbarengan
  menghapus user — voucher `from_login` bisa meninggalkan scheduler/script
  bernama=username yatim. (Preferensi desain scheduler harian tunggal di Task 9
  menghilangkan objek per-user; klausul ini menutup deployment yang terlanjur
  memakai pola per-user.)
- [ ] Pengetahuan RouterOS **tidak** bocor ke usecase (AGENTS.md §1.2) — semua di
  `commands.go`.
- [ ] Table-driven test di `commands_test.go` mencakup mode A (user+profile
  keepalive + sekuens cookie-remove), mode B (ip-binding+netwatch), sekuens
  kill-sesi **N sesi → N `active/remove`**, dan pembersihan scheduler yatim pada
  remove.

**Files likely touched:** `internal/driver/mikrotik/commands.go`,
`internal/driver/mikrotik/commands_test.go`.

**Dependencies:** Task 2

**Estimated scope:** Medium

---

**Task 6: Usecase generate bulk + provisioning voucher**

**Description:** Usecase yang membuat batch + N voucher + baris sync-log
provisioning, dipanggil handler `POST /voucher-batches`.

**Acceptance criteria:**
- [ ] Usecase (mis. `internal/usecase/business/manage_voucher.go` untuk generate
  + `internal/usecase/network/provision_voucher.go` untuk penulisan sync-log)
  membuat `voucher_batches`, `hotspot_vouchers`, dan satu baris
  `provisioning_sync_log` (`target_type='mikrotik_hotspot_user'`,
  `action='create'`) per voucher, dalam satu transaksi bisnis.
- [ ] Validasi: plan harus `service_type='hotspot'`; `quantity` ≤ cap; device ada.
- [ ] Kode voucher digenerate memakai `CodeGenSpec` domain (Task 1) yang bersumber
  dari kolom plan/setting (Task 2): charset (`numeric`/`alpha`/`alphanumeric`,
  default non-ambigu), panjang, prefix, uppercase, `password_same_as_username`.
- [ ] **Keunikan lintas DB & router:** selain UNIQUE DB, kode dicek ke **router**
  (`/ip hotspot user print ?name`) sebelum provisioning karena router bisa memuat
  user lintas-batch; bila driver mengembalikan error `already exists` (Task 5),
  usecase/Sync Engine me-regenerate kode & retry, bukan menganggap sukses
  (REFERENCES §E `VoucherGenerator.isUsernameExists`).
- [ ] Tidak memanggil `port.DeviceDriver` langsung (K4) — hanya menulis sync-log.
- [ ] Untuk model `from_creation`, `expires_at` diisi saat generate; untuk
  `from_login`/`uptime_quota`, `expires_at` null sampai login pertama / ditegakkan
  native.
- [ ] Table-driven test usecase.

**Files likely touched:** `internal/usecase/business/manage_voucher.go`,
`internal/usecase/network/provision_voucher.go`, test terkait.

**Dependencies:** Task 3, Task 4, Task 5

**Estimated scope:** Large

---

**Task 7: Capture login pertama (integrasi Issue 12)**

**Description:** Konsumen event connect hotspot dari stream Issue 12
(`/ip/hotspot/active/print follow`) menandai login pertama voucher: set
`activated_at`, hitung `expires_at` sesuai `ValidityModel` plan, ubah `status`
`unused→active`, dan isi `used_by_mac`.

**Acceptance criteria:**
- [ ] Event login hotspot (username = kode voucher, MAC, IP) dipetakan ke voucher
  via `FindVoucherByCode`.
- [ ] Bila `status='unused'`: set `activated_at` = waktu event, `expires_at` =
  hitung dari `ValidityModel` (untuk `from_login`: `activated_at + validity_days`;
  `uptime_quota`: null, ditegakkan native; `from_creation`: sudah diisi saat
  generate), `status='active'`, `used_by_mac` diisi.
- [ ] Idempoten: login berulang voucher yang sudah `active` tidak menimpa
  `activated_at`.
- [ ] Untuk voucher terikat subscription, isi/verifikasi `used_by_subscription_id`.
- [ ] Memakai `port.StreamingDeviceDriver` via Issue 12 — **tidak** polling
  `/ip hotspot active print` (ADR 0003).
- [ ] Test dengan event sintetis.

**Files likely touched:** `internal/usecase/network/` (konsumen event hotspot,
bisa memperluas konsumen Issue 12), test terkait.

**Dependencies:** Task 4, Issue 12

**Estimated scope:** Medium

---

**Task 8: Expiry enforcement — scheduler Golang (jalur primer, teraudit)**

**Description:** Job terjadwal menemukan voucher yang lewat `expires_at` (atau
kuota habis) dan menulis baris sync-log `disable`/`delete` (per kebijakan) agar
Sync Engine memutus + kill sesi. Ini jalur utama yang teraudit.

**Acceptance criteria:**
- [ ] Usecase (mis. `internal/usecase/network/expire_vouchers.go`)
  `ExpireDueVouchers(ctx, now)` memakai `FindExpiring` → untuk tiap voucher tulis
  sync-log (`action='disable'` untuk suspend sementara atau `action='delete'`
  untuk pembersihan, sesuai kebijakan plan/retensi), set `status='expired'`.
- [ ] Dipicu scheduler (mekanisme sama dengan Open Question Issue 01 — ticker
  sederhana dulu). Interval dokumentasi (mis. tiap 1 menit) + idempoten.
- [ ] Tidak memanggil driver langsung (K4).
- [ ] `now` disuntik sebagai parameter agar dapat diuji deterministik.
- [ ] Table-driven test.

**Files likely touched:** `internal/usecase/network/expire_vouchers.go`,
`cmd/server/main.go` (registrasi scheduler), test terkait.

**Dependencies:** Task 4, Task 5, Issue 01

**Estimated scope:** Medium

---

**Task 9: Generator script router `on-login` + scheduler (jaring pengaman offline)**

**Description:** Hasilkan dan pasang script `on-login` di `/ip hotspot user
profile` + entri `/system scheduler` di router sebagai enforcement redundan saat
backend Golang mati. Pemasangan script ini **sendiri** melalui sync-log (jadi
tetap teraudit sebagai satu perubahan konfigurasi), tetapi eksekusi harian script
oleh router berjalan mandiri.

**Acceptance criteria:**
- [ ] Isi script `on-login` di-generate per `ValidityModel`: `from_login` → set
  `comment` voucher = waktu expiry saat login pertama; scheduler harian menghapus
  user yang `comment`-nya sudah lewat. `uptime_quota` → tak perlu script (native
  `limit-uptime`). `from_creation` → `comment` expiry di-set saat create (Task 5),
  scheduler harian menghapus yang lewat.
- [ ] **Guard prefix comment (MUST-FIX, K14):** script `on-login` **wajib**
  memeriksa prefix comment kanonik Polyglot (mis. `poly:vc:`/`poly:up:`, dan
  memperlakukan comment kosong sebagai belum-terset) sebelum menulis expiry.
  Tujuannya (a) **idempoten** — login ulang voucher yang sudah aktif tidak mereset
  expiry, dan (b) script **tidak menyentuh** user subscription/manual yang
  comment-nya tidak berprefix Polyglot. Bukti REFERENCES §E
  (`generateHotspotExpiryScript` guard `vc`/`up`/kosong).
- [ ] **Varian ROS6 vs ROS7 (MUST-FIX):** sediakan dua template script karena
  parsing/format tanggal RouterOS v6 dan v7 berbeda; driver memilih varian sesuai
  versi RouterOS device. Test membuktikan kedua varian dihasilkan.
- [ ] Pemasangan script + scheduler ditulis sebagai baris sync-log
  (`target_type='mikrotik_hotspot_profile'` atau setara) sehingga tercatat di
  `command_audit_log` — perubahan **struktur** tetap teraudit; hanya eksekusi
  periodik oleh router yang mandiri.
- [ ] Script bersifat opsional per deployment (flag konfigurasi) — bila
  offline-autonomy tak diinginkan, Task 9 bisa dinonaktifkan tanpa mematahkan
  Task 8.
- [ ] **Scheduler harian tunggal, bukan per-user (K15):** desain memakai **satu**
  `/system scheduler` harian yang menyapu semua user expired by-comment, bukan
  membuat objek scheduler/script bernama=username per voucher. Ini menghilangkan
  akar orphan (referensi memakai per-user → yatim; REFERENCES §E). Keputusan ini
  dinyatakan eksplisit di ADR 0012.
- [ ] Isi script didokumentasikan di ADR 0012 (bukan sebagai blok kode panjang di
  komentar Go — AGENTS.md §7).

**Files likely touched:** `internal/driver/mikrotik/commands.go` (template script
sebagai data vendor), ADR `docs/adr/0012-hotspot-voucher-lifecycle-hybrid.md`.

**Dependencies:** Task 5, Task 8

**Estimated scope:** Medium

---

**Task 10: Pass rekonsiliasi router ↔ DB (menutup lubang audit hybrid)**

**Description:** Karena scheduler router (Task 9) dapat menghapus voucher tanpa
melewati pipeline audit, sebuah pass rekonsiliasi periodik membaca daftar
`/ip hotspot user` di router, membandingkan dengan DB, dan mencatat divergensi ke
`command_audit_log`/`provisioning_sync_log` agar audit tetap eventually-consistent.

**Acceptance criteria:**
- [ ] Usecase `ReconcileHotspotUsers(ctx, deviceID)` membaca daftar user hotspot
  aktual dari router (read-only, via driver) dan set voucher di DB yang sudah
  tak ada di router menjadi `status='expired'`/dihapus, sambil menulis catatan
  audit bertanda `source='scheduled_job'` yang menjelaskan penghapusan dilakukan
  router.
- [ ] Divergensi sebaliknya (voucher ada di DB `active` tapi hilang di router
  tanpa catatan) dilaporkan sebagai anomali (log + entri yang bisa dilihat lewat
  `GET /sync-logs` atau endpoint audit) — tidak diam-diam.
- [ ] **Deteksi objek yatim (K15):** rekonsiliasi juga membaca `/system scheduler`
  dan `/system script` ber-nama/berprefix Polyglot yang tak lagi punya user
  hotspot pasangannya (sisa pola per-user pra-migrasi ke scheduler harian tunggal,
  Task 9) dan menandainya untuk pembersihan teraudit — jangan biarkan menumpuk.
- [ ] Idempoten & dijadwalkan (interval lebih longgar dari Task 8, mis. tiap
  15 menit).
- [ ] Read hotspot users **tidak** dianggap polling status pelanggan (itu
  Issue 12) — ini rekonsiliasi katalog user, dijalankan jarang, bukan sumber data
  sesi real-time.
- [ ] Test dengan daftar router sintetis.

**Files likely touched:** `internal/usecase/network/reconcile_hotspot.go`,
`internal/driver/mikrotik/commands.go` (operasi list users), test terkait.

**Dependencies:** Task 8, Task 9

**Estimated scope:** Medium

---

**Task 11: Handler REST voucher & batch**

**Description:** Implementasi seluruh endpoint di §REST API dengan DTO terpisah
dari domain, validasi, dan pemetaan error→status.

**Acceptance criteria:**
- [ ] Semua endpoint di tabel §REST API terpasang di router dengan
  `AuthRequired` + `RBACRequired` sesuai role minimum.
- [ ] DTO di `internal/adapter/http/dto/` (request/response), tak membocorkan
  domain.
- [ ] Aksi ke perangkat mengembalikan `202` + `sync_log` id; transisi status
  ilegal → `409`; validasi → `400`.
- [ ] Baris policy ditambahkan ke `configs/rbac_policy.csv` (generate/revoke:
  admin/owner; baca: staff; disable/enable/remove: admin).
- [ ] Handler diuji dengan `httptest` + repo/usecase yang di-stub pada boundary
  yang tepat.

**Files likely touched:** `internal/adapter/http/voucher_handler.go`,
`internal/adapter/http/dto/voucher.go`, `internal/adapter/http/router.go`,
`configs/rbac_policy.csv`, `internal/adapter/http/voucher_handler_test.go`.

**Dependencies:** Task 6, Task 8

**Estimated scope:** Medium

---

**Task 12: ADR 0012 — hotspot voucher lifecycle hybrid**

**Description:** Dokumentasikan keputusan: model validity configurable per plan,
enforcement hybrid (Golang primer + script router jaring pengaman), deviasi sadar
dari jalur audit tunggal, dan bagaimana pass rekonsiliasi (Task 10) menutupnya.

**Acceptance criteria:**
- [ ] `docs/adr/0012-hotspot-voucher-lifecycle-hybrid.md` menjelaskan konteks,
  keputusan, konsekuensi, dan **eksplisit** menyatakan deviasi dari prinsip audit
  tunggal (AGENTS.md §0) beserta mitigasinya (rekonsiliasi + pemasangan script
  yang tetap teraudit).
- [ ] ADR **wajib** memuat, sebagai jebakan/keputusan terdokumentasi:
  - **Format comment prefix kanonik** (K14) yang di-parse balik on-login guard &
    rekonsiliasi.
  - **Guard on-login** (idempoten + tidak menyentuh user subscription/manual) dan
    **varian script ROS6 vs ROS7** (parsing tanggal berbeda).
  - **Scheduler harian tunggal** (bukan per-user) sebagai anti-orphan (K15).
  - **Cookie-remove mode A** wajib setelah `active/remove` (jebakan auto-relogin
    mac-cookie).
  - **Deviasi audit + pass rekonsiliasi** (K15) yang menutup lubang, termasuk
    deteksi scheduler/script yatim.
- [ ] Menyertakan isi/template script `on-login` per model validity (di ADR, bukan
  di komentar kode) — kedua varian ROS6/ROS7.
- [ ] Ditautkan dari `README.md` root pada PR yang sama (AGENTS.md §1.5).
- [ ] Nomor ADR 0012 sesuai tabel reservasi README §K6.

**Files likely touched:** `docs/adr/0012-hotspot-voucher-lifecycle-hybrid.md`,
`README.md`.

**Dependencies:** Task 9, Task 10

**Estimated scope:** Small

---

## Migrasi Database

Empat migrasi baru (nomor dari tabel reservasi README §K6):

- **`000032_add_hotspot_config_to_plans`** (up/down berpasangan) — ALTER `plans`
  tambah kolom nullable: `hotspot_shared_users`, `hotspot_validity_model`
  (CHECK `from_login`/`uptime_quota`/`from_creation`), `hotspot_validity_value` +
  `hotspot_validity_unit` (CHECK `minutes`/`hours`/`days`; menggantikan gagasan
  `_days` karena paket riil per-jam & voucher <1 hari), `hotspot_uptime_limit_seconds`,
  `hotspot_session_timeout_seconds`, serta konfigurasi generate kode
  (`voucher_code_length`, `voucher_code_charset` CHECK `numeric`/`alpha`/
  `alphanumeric`, `voucher_code_prefix`, `voucher_code_uppercase`,
  `voucher_password_same_as_username`). Cermin ke `DATABASE-SCHEMA.md` §4.2.
- **`000033_add_hotspot_target_types`** (up/down berpasangan) — ALTER CHECK
  `provisioning_sync_log.target_type` tambah `mikrotik_hotspot_profile` dan
  `mikrotik_ip_binding` (`mikrotik_hotspot_user` sudah ada). Cermin ke
  `DATABASE-SCHEMA.md` §6.3.
- **`000034_add_hotspot_access_mode_to_subscriptions`** (up/down berpasangan) —
  ALTER `subscriptions` tambah kolom nullable `hotspot_access_mode` (CHECK
  `mac_login`/`ip_binding`). Kolom `mac_address`/`static_ip` sudah ada (000009),
  dipakai ulang. Cermin ke `DATABASE-SCHEMA.md` §6.1.

Tabel `voucher_batches` & `hotspot_vouchers` **tidak** dibuat ulang — sudah ada di
migrasi 000007; issue ini hanya mengonsumsinya.

## Verification

- [ ] `go build ./...` sukses.
- [ ] `go test ./internal/domain/voucher/...`,
  `go test ./internal/adapter/postgres/... -run Voucher`,
  `go test ./internal/driver/mikrotik/... -run Hotspot`,
  `go test ./internal/usecase/... -run Voucher` — semua lulus.
- [ ] `go test ./internal/adapter/http/... -run Voucher` lulus.
- [ ] `make lint` bersih.
- [ ] Migrasi maju-mundur: `make migrate-up` lalu `make migrate-down` untuk
  000032/000033/000034 tidak error pada DB kosong.
- [ ] Smoke test manual (curl, sebutkan sebagai perintah): login → generate batch
  20 voucher via `POST /api/v1/voucher-batches` → cek 20 baris `sync_log` pending
  di `GET /api/v1/sync-logs?status=pending` → (dengan MikroTik CHR di GNS3)
  verifikasi user muncul di `/ip hotspot user` → simulasikan login satu kode →
  cek `activated_at`/`expires_at` terisi via `GET /api/v1/vouchers/:id` → tunggu
  expiry / picu scheduler → verifikasi user ter-disable/remove dan sesi aktif
  ter-kill.
- [ ] Smoke test dua mode identitas: `provision-hotspot` `access_mode=mac_login`
  → verifikasi login by MAC muncul di `/ip hotspot active` + record connect
  tertangkap Issue 12 via `/ip/hotspot/active/print follow`; `access_mode=ip_binding`
  → verifikasi entri
  `/ip hotspot ip-binding type=bypassed` + `/tool netwatch` terpasang, matikan IP
  → event `netwatch` down tertangkap Issue 12.
- [ ] Test integrasi hotspot di `test/integration/` terhadap MikroTik nyata,
  bukan mock `port.DeviceDriver`.

## Open Questions

1. **Jual-online voucher (generate-after-payment) = issue billing terpisah.**
   Referensi (REFERENCES §E: `gembok-bill voucher_pricing` berjenjang +
   `voucher-order.php` order publik `pending→paid→voucher`) menjual voucher lewat
   tabel `voucher_orders`/webhook pembayaran. Itu **di luar** scope issue ini
   (billing). Silang yang wajib dijaga: provisioning voucher online **hanya**
   dipicu **setelah** status order `paid`, dan tetap lewat jalur teraudit —
   webhook billing menulis `provisioning_sync_log` (K4), **tidak** memanggil
   driver langsung dan **tidak** generate voucher sebelum bayar. Kolom
   `code_length`/`charset`/`prefix` reuse konfigurasi plan (Task 2). Diputuskan di
   issue billing, bukan di sini.
2. **Retensi voucher `used`/`expired`** — dihapus dari router (Task 8/9) tetapi
   apakah baris `hotspot_vouchers` di-soft-delete atau dipertahankan untuk
   laporan penjualan? Diselaraskan dengan kebijakan billing.

## Definition of Done

- [ ] Domain voucher + repo + migrasi 000032/000033 selesai dan tercermin di
  `DATABASE-SCHEMA.md`.
- [ ] Bulk generate + provisioning voucher jalan end-to-end lewat sync-log
  (teraudit), bukan memanggil driver dari handler.
- [ ] Tiga model validity berfungsi; capture login pertama via Issue 12 mengisi
  `activated_at`/`expires_at` tanpa polling.
- [ ] Pemutusan (disable/remove) selalu mengikutkan kill sesi aktif.
- [ ] Enforcement hybrid: scheduler Golang primer + script router opsional; pass
  rekonsiliasi menutup lubang audit dan divergensi dilaporkan, tidak senyap.
- [ ] ADR 0012 mendokumentasikan deviasi sadar + mitigasi, ditautkan dari
  `README.md`.
- [ ] Endpoint REST + policy RBAC lengkap dan teruji.
- [ ] `go test ./...` dan `make lint` hijau.
