# Issue 09: GenieACS CPE Operations

## Konteks

Temuan C di `ANALISIS-PROVISIONING-REPO-REFERENSI.md` §4 dan `DATABASE-SCHEMA.md`
§7.3 menetapkan bahwa GenieACS/TR-069 diperlakukan sebagai **satu `target_type`
di `provisioning_sync_log`** (`genieacs_tr069`), dengan `external_reference`
menyimpan **task ID dari GenieACS NBI**. Issue 08 sudah membangun cache lokal
`acs_devices` (mirror device GenieACS) dan menambah kolom `genieacs_device_id` di
`subscriptions`. Yang belum ada adalah **operasi aktif ke CPE/ONT**: reboot,
enable/disable WAN, ganti SSID/password WiFi, dan refresh parameter. Issue ini
mengisi gap itu.

Semua operasi mengikuti konvensi K4: handler REST TIDAK memanggil
`port.DeviceDriver` langsung. Alur: handler → usecase menulis baris
`provisioning_sync_log` `target_type='genieacs_tr069'` status `pending` → Sync
Engine (Issue 01) menerjemahkannya jadi `command.Command` → `usecase/network`.
`ExecuteCommand` (Classify → Decide → HITL bila destruktif → Execute ke driver
`genieacs`) → `command_audit_log` terisi → `sync_log` jadi `success`/`failed`.
Pengetahuan protokol TR-069 (task type, path parameter) SELALU hidup di
`internal/driver/genieacs/` (K1), tidak pernah bocor ke `usecase/`/`domain/`.

**Invariant satu baris = satu operasi CPE (validasi K4):** satu operasi CPE logis
= satu baris `provisioning_sync_log` `genieacs_tr069`, yang di-realisasikan menjadi
**satu task NBI primer** (`{name, parameterValues}` untuk set, `{name, objectName}`
untuk refresh, `{name}` untuk reboot) yang di-`POST` ke device `_id` GenieACS
(`device_id` baris = `genieacs_device_id` = `_id` GenieACS). Task ID primer inilah
yang mengisi `external_reference` (§7.3). Task `refreshObject` susulan yang
dirangkai driver setelah `setParameterValues` (Task 2) adalah **detail internal
driver**, bukan baris `sync_log` kedua. `reboot` dan `factoryReset` bersifat
destruktif → wajib lewat HITL (K4); `factoryReset` sendiri di luar scope issue
ini.

Karena GenieACS asinkron by design (ADR 0005 — CPE menghubungi ACS secara
berkala lewat "Inform", ACS memicu lebih awal lewat "Connection Request", lalu
NBI di-poll untuk hasilnya), hasil final sebuah task bisa datang belakangan.
Driver `genieacs` yang sudah ada (`client.go`) menangani polling ke
`GET /tasks/{id}` sampai task keluar antrian; `sync_log` menjadi `success` saat
task keluar antrian **tanpa fault** dan `failed` bila ada fault/timeout. Issue ini
juga melengkapi Issue 04: cascade suspend menulis baris `genieacs_tr069`
`action=disable` (disable WAN), dan issue inilah yang membuat baris itu
benar-benar tereksekusi ke CPE.

## Prasyarat

- **Issue 01 (Provisioning Sync Engine)** — wajib. Menyediakan pemrosesan baris
  `provisioning_sync_log` `pending`: dispatch per `target_type`, penerjemahan ke
  `command.Command`, pemanggilan `usecase/network.ExecuteCommand`, penautan
  `command_audit_log_id`, dan pengisian `external_reference` dari hasil eksekusi.
  Issue ini hanya MENULIS baris `pending` dan MENYEDIAKAN penerjemah
  genieacs-spesifik yang dipanggil Sync Engine.
- **Issue 08 (GenieACS Device Cache)** — wajib. Menyediakan tabel `acs_devices`
  (sumber `params`/`connection_request_url`/`last_inform` untuk endpoint baca) dan
  kolom `subscriptions.genieacs_device_id` (penunjuk CPE mana yang dioperasikan).
- **Foundation Phase 4/5** — router Gin nyata, middleware `AuthRequired` +
  `RBACRequired`, `internal/adapter/http/dto/`, registry aktif (`registry.Get`
  mengembalikan `port.DeviceDriver` untuk `genieacs_device_id`).
- **Driver `genieacs` yang sudah ada** — `client.go` (Driver + polling NBI),
  `commands.go` (`Classify`/`Translate` + `buildTaskBody`), `errors.go`. Issue ini
  MEMPERLUAS `commands.go`/`client.go`, bukan menulis ulang.

## Ruang Lingkup

**In scope:**
- Perluasan `internal/driver/genieacs/commands.go`: penerjemahan operasi abstrak
  CPE (reboot, set WAN enable/disable, set WiFi SSID/password, refresh/get
  parameter) menjadi task GenieACS NBI dengan **nama task NBI persis** —
  `setParameterValues`, `refreshObject`, `reboot`, `factoryReset`,
  `getParameterValues` — plus klasifikasi risiko yang benar.
- Peta path TR-069 per `product_class` sebagai **data** di dalam driver,
  bersifat **multi-path shotgun** (satu operasi abstrak → beberapa path native,
  lihat K13 di README §Konvensi Bersama), fail-safe untuk model yang belum
  dikurasi. Katalog path per operasi = pengetahuan vendor milik driver, tidak
  pernah bocor ke usecase/domain (K1).
- Setelah `setParameterValues` (set WiFi/WAN), driver **otomatis** merangkai task
  `refreshObject` lanjutan agar cache lokal `acs_devices` (Issue 08) tidak basi
  dari nilai yang baru ditulis.
- Perluasan `internal/driver/genieacs/client.go` (`Execute`) agar mengembalikan
  **task ID GenieACS** di `command.Result` supaya Sync Engine bisa mengisi
  `external_reference`.
- Empat usecase di `internal/usecase/network/`: reboot CPE, set WAN, set WiFi
  (ketiganya menulis baris `provisioning_sync_log` `pending`), dan baca parameter
  CPE (membaca cache `acs_devices`, opsional memicu refresh).
- Empat endpoint REST + DTO + wiring route ber-RBAC.
- Satu kolom baru `payload` (jsonb, nullable) di `provisioning_sync_log` untuk
  membawa argumen aksi (SSID/password WiFi, flag `enabled` WAN) sampai Sync
  Engine memprosesnya.

**Out of scope:**
- Pemrosesan baris sync ke perangkat nyata (dispatch, HITL, audit) — milik Issue
  01. Issue ini hanya menulis baris `pending` + menyediakan penerjemah driver.
- Pembuatan tabel `acs_devices` dan kolom `genieacs_device_id` — milik Issue 08.
- Pembacaan RX power / optical health ke InfluxDB — milik Issue 10.
- Penulisan baris cascade suspend/resume/terminate — milik Issue 04 (issue ini
  hanya membuat baris `genieacs_tr069` milik Issue 04 bisa dieksekusi).
- Penambahan operasi TR-069 di luar empat kategori di atas (firmware
  `download`/`factoryReset`, addObject/deleteObject) — bila dibutuhkan nanti,
  issue tersendiri.

## REST API

Base path: `/api/v1`. Semua aksi ke CPE mengikuti konvensi K4: handler menulis
baris `provisioning_sync_log` `pending`, lalu response **202 Accepted**
mengembalikan `sync_log_id` (bukan hasil eksekusi CPE — hasil final menyusul
asinkron via Sync Engine). Endpoint baca parameter adalah pengecualian: ia
membaca cache lokal `acs_devices` dan mengembalikan **200 OK** sinkron.

| Method | Path | Tujuan | Role minimum |
|---|---|---|---|
| POST | `/api/v1/subscriptions/:id/cpe/reboot` | Reboot CPE/ONT via task `reboot` GenieACS | teknisi |
| POST | `/api/v1/subscriptions/:id/cpe/wan` | Enable/disable WAN CPE (dipakai cascade suspend Issue 04) | admin |
| PUT | `/api/v1/subscriptions/:id/cpe/wifi` | Ganti SSID + password WiFi CPE | teknisi |
| GET | `/api/v1/subscriptions/:id/cpe/parameters` | Baca parameter TR-069 dari cache `acs_devices` (opsional refresh) | staff |

Precondition umum untuk keempat endpoint: subscription harus ada (`404` bila
tidak) dan punya `genieacs_device_id` non-kosong yang mereferensikan baris
`acs_devices` valid (`409` bila subscription tidak terhubung ke CPE GenieACS).

### POST `/api/v1/subscriptions/:id/cpe/reboot`

- **Request:** body kosong (atau opsional `reason` string untuk audit).
- **Sync log yang ditulis:** satu baris `target_type='genieacs_tr069'`,
  `action='update'` (lihat catatan pemetaan aksi di bawah), `device_id` =
  `genieacs_device_id`, `status='pending'`, `payload=NULL` (reboot tak butuh
  argumen).
- **Response sukses (202):** objek berisi `sync_log_id` (UUID), `subscription_id`,
  `target_type` (`genieacs_tr069`), `action` (`update`), dan `operation`
  (`reboot`) agar pemanggil tahu apa yang dijadwalkan.
- **Response gagal:** `404` subscription tidak ada; `409` tanpa
  `genieacs_device_id`; `403` role kurang; `500` transaksi tulis gagal.

### POST `/api/v1/subscriptions/:id/cpe/wan`

- **Request:** field `enabled` (bool, wajib) — `false` = disable WAN (isolir),
  `true` = enable WAN.
- **Sync log yang ditulis:** satu baris `genieacs_tr069`, `action='disable'` bila
  `enabled=false` atau `action='enable'` bila `enabled=true`, `payload` jsonb
  `{"enabled": <bool>}`.
- **Response sukses (202):** `sync_log_id`, `subscription_id`, `action`
  (`enable`/`disable`), `operation` (`set_wan`), `enabled`.
- **Response gagal:** `400` bila `enabled` tidak ada di body; `404`; `409`; `403`;
  `500`.

### PUT `/api/v1/subscriptions/:id/cpe/wifi`

- **Request:** field `ssid` (string, wajib, 1–32 karakter sesuai batas SSID) dan
  `password` (string, wajib, minimal 8 karakter sesuai WPA2). Validasi panjang di
  DTO.
- **Sync log yang ditulis:** satu baris `genieacs_tr069`, `action='update'`,
  `payload` jsonb `{"ssid": "...", "password": "..."}`.
- **Response sukses (202):** `sync_log_id`, `subscription_id`, `action`
  (`update`), `operation` (`set_wifi`), `ssid`. **`password` TIDAK pernah
  dikembalikan di response** maupun ditulis ke log aplikasi.
- **Response gagal:** `400` bila `ssid`/`password` kosong atau di luar batas
  panjang; `404`; `409`; `403`; `500`.

### GET `/api/v1/subscriptions/:id/cpe/parameters`

- **Request:** query opsional `refresh` (bool, default `false`). Bila `true`,
  usecase juga menulis baris `genieacs_tr069` `action='update'` (task
  `refreshObject`, read-only) untuk memicu pembaruan cache, lalu tetap
  mengembalikan snapshot cache saat ini.
- **Response sukses (200):** objek berisi `genieacs_device_id`, `serial_number`,
  `manufacturer`, `product_class`, `last_inform` (timestamp Inform terakhir),
  `synced_at`, dan `params` (subset parameter TR-069 yang di-cache Issue 08 —
  mis. status WAN, SSID, RX power ringkas). Bila `refresh=true`, sertakan juga
  `refresh_sync_log_id` (UUID baris refresh yang dijadwalkan).
- **Response gagal:** `404` subscription tidak ada; `409` tanpa
  `genieacs_device_id` atau baris `acs_devices` belum pernah terisi (belum ada
  Inform pertama); `403` role kurang; `500`.

**Catatan pemetaan aksi (`provisioning_sync_log.action`):** kolom `action` hanya
menerima nilai `create|update|disable|enable|delete|change_profile`. Untuk
`genieacs_tr069` dipakai konvensi berikut dan HARUS konsisten dengan Issue
01/04: `enable`/`disable` khusus untuk WAN CPE (memetakan ke `setParameterValues`
yang menyalakan/mematikan WAN), dan `update` untuk semua operasi lain yang
mengubah/menyegarkan state CPE tanpa semantik enable/disable — yaitu **reboot**,
**set WiFi**, dan **refresh parameter**. Semantik operasi sebenarnya (reboot vs
wifi vs refresh) dibedakan oleh isi `payload` + penerjemah driver, bukan oleh
`action`. Pilihan ini disengaja: menghindari menambah nilai enum `action` baru
(mis. `reboot`) yang akan memaksa migrasi ulang constraint dan menyentuh Issue
01/04. Nyatakan ulang keputusan ini di deskripsi PR (AGENTS.md §0.3).

## Tasks

**Task 1: Perluas katalog & klasifikasi operasi CPE di driver `genieacs`**

**Description:** Tambahkan penerjemahan operasi abstrak CPE (reboot, set WAN, set
WiFi, refresh/get parameter) menjadi task GenieACS NBI di
`internal/driver/genieacs/commands.go`, tanpa membocorkan pengetahuan TR-069 ke
luar package driver.

**Acceptance criteria:**
- [ ] Operasi baru dapat diterjemahkan ke `command.Command` dengan `Raw` berupa
      **nama task GenieACS NBI persis** (case-sensitive, sesuai NBI): reboot →
      `reboot`; set WAN dan set WiFi → `setParameterValues`; refresh →
      `refreshObject`; baca parameter → `getParameterValues`. Nama-nama ini adalah
      `name` pada body task `POST /devices/{id}/tasks` — jangan disingkat/diubah
      kapitalisasinya. `factoryReset` termasuk kosakata NBI yang sama tapi di luar
      scope issue ini (lihat Out of scope).
- [ ] Bentuk argumen task NBI dimodelkan benar di driver: `setParameterValues`
      membawa `parameterValues` = array tuple `[path, value]` **atau**
      `[path, value, xsd_type]` (mis. `xsd:string`, `xsd:boolean`,
      `xsd:unsignedInt`); `refreshObject` membawa `objectName` (nama sub-tree yang
      di-refresh); `reboot` tanpa argumen tambahan; `getParameterValues` membawa
      daftar nama parameter yang dibaca.
- [ ] Opsi eksekusi segera vs tunggu Inform dimodelkan sebagai flag driver yang
      memetakan ke query `?connection_request` pada NBI (memaksa ACS mengirim
      Connection Request agar CPE memproses task saat itu juga, bukan menunggu
      Inform periodik berikutnya). Default untuk aksi CPE interaktif (reboot/WAN/
      WiFi) = pakai `connection_request`; refresh pasif boleh tanpa.
- [ ] `Classify` mengembalikan `command.ClassDestructive` untuk `reboot` dan
      `setParameterValues` (WAN enable/disable dan WiFi termasuk destruktif →
      butuh HITL), dan `command.ClassReadOnly` untuk `getParameterValues`/
      `refreshObject`. `factoryReset` (bila kelak ditambah) WAJIB destruktif.
      (Kumpulan `destructiveTaskTypes` yang sudah ada sudah mencakup `reboot` +
      `setParameterValues` — verifikasi tidak berubah.)
- [ ] Argumen aksi yang bersifat LOGIS (mis. kunci `wan_enabled`, `wifi_ssid`,
      `wifi_password`) diterima lewat `command.Command.Args` (map string→string,
      nilai kompleks di-encode JSON sesuai konvensi `buildTaskBody` yang sudah
      ada) — BUKAN path TR-069 mentah. Caller (Sync Engine) tidak pernah menulis
      path TR-069.
- [ ] Ada peta path TR-069 per `product_class` sebagai **data** di dalam package
      driver (mis. `map[string]cpeParamPaths`), memetakan kunci logis → path
      TR-069 konkret secara **multi-path shotgun** (K13): satu kunci logis di-fan-out
      ke beberapa path native karena vendor ONU menyimpan di path berbeda —
      minimal SSID ke `InternetGatewayDevice.LANDevice.1.WLANConfiguration.1.SSID`
      **dan** `Device.WiFi.SSID.1.SSID`; password ke
      `...WLANConfiguration.1.PreSharedKey.1.KeyPassphrase` /
      `...WLANConfiguration.1.KeyPassphrase` /
      `...WLANConfiguration.1.PreSharedKey.1.PreSharedKey`. Index band 5G
      **ditemukan dinamis** dari parameter CPE (mis. dari cache `acs_devices`),
      **bukan** di-hardcode `.5`. Semua path ini pengetahuan vendor milik driver
      (K13, README §Konvensi Bersama).
- [ ] **Fail-safe:** bila `product_class` CPE tidak dikenal dan tidak ada
      pemetaan default TR-098/TR-181 standar yang cocok, penerjemahan
      mengembalikan error (mis. `ErrUnknownCPEModel`) — TIDAK menebak path yang
      bisa merusak konfigurasi CPE.
- [ ] Katalog WAN (kunci logis `wan_enabled` dan atribut WAN lain) menyimpan
      pilihan koneksi sebagai **data**, bukan hardcode: pilih
      `WANPPPConnection` vs `WANIPConnection` sesuai model, `ConnectionType`
      (`IP_Routed`/`IP_Bridged`), `VLANID`, serta parameter vendor bila perlu
      (mis. `X_BROADCOM_*Bind`). Semua ini entri katalog per-`product_class` di
      driver — usecase/Sync Engine tidak pernah menyebut path/parameter WAN
      spesifik ini. (Detail atribut WAN di luar enable/disable boleh menyusul saat
      dibutuhkan; struktur katalognya disiapkan sekarang agar tidak perlu bongkar
      ulang.)
      set WiFi (verifikasi fan-out multi-path SSID+password), refresh,
      product_class tak dikenal (error), verifikasi `Classify` untuk tiap kasus.

**Files likely touched:** `internal/driver/genieacs/commands.go`,
`internal/driver/genieacs/errors.go`, `internal/driver/genieacs/commands_test.go`.

**Dependencies:** —

**Estimated scope:** Medium

---

**Task 2: Driver `Execute` — bangun task body & surface task ID GenieACS**

**Description:** Pastikan `Execute` di `client.go` menyusun body task dari kunci
logis Task 1, mem-POST ke `POST /devices/{id}/tasks`, memoll sampai selesai, lalu
mengembalikan **task ID GenieACS** di `command.Result` supaya Sync Engine bisa
menyimpannya di `provisioning_sync_log.external_reference`.

**Acceptance criteria:**
- [ ] `buildTaskBody` (atau helper setara) meresolusi kunci logis `wan_enabled`/
      `wifi_ssid`/`wifi_password` menjadi tuple `parameterValues`
      `[path, value, xsd_type]` memakai peta path Task 1 (fan-out multi-path,
      K13) sebelum marshalling.
- [ ] Setelah task `setParameterValues` (set WiFi/WAN) di-POST dan selesai, driver
      **otomatis merangkai task `refreshObject`** untuk sub-tree yang berubah
      (objectName WLANConfiguration / WANConnection terkait) supaya cache
      `acs_devices` (Issue 08) konsisten dengan nilai yang baru ditulis — bukan
      dibiarkan basi sampai Inform periodik berikutnya. Refresh susulan ini
      read-only dan tidak mengubah semantik HITL task set-nya.
- [ ] `command.Result` yang dikembalikan `Execute` membawa task ID GenieACS
      (mis. field yang sudah ada atau ditambah — nyatakan pilihan di PR) sehingga
      Sync Engine dapat mengisi `external_reference`. Bila operasi direct-GET
      (baca), `external_reference` boleh kosong.
- [ ] Perilaku polling yang sudah ada dipertahankan: `success` saat task keluar
      antrian tanpa fault; error dibungkus bila ada fault channel/timeout
      (`errors.go` yang sudah ada dipakai ulang, tidak dibuat sentinel duplikat).
- [ ] Password WiFi TIDAK pernah masuk log/`result_summary` — hanya nama
      parameter yang boleh dicatat, bukan nilainya.
- [ ] `context.Context` parameter pertama; error posisi terakhir; wrap `%w`.
- [ ] Test unit dengan `httptest` server men-stub NBI: verifikasi body task benar
      untuk WAN/WiFi, task ID ter-surface di `command.Result`.

**Files likely touched:** `internal/driver/genieacs/client.go`,
`internal/driver/genieacs/commands.go`, `internal/driver/genieacs/client_test.go`.

**Dependencies:** Task 1.

**Estimated scope:** Medium

---

**Task 3: Kolom `payload` di `provisioning_sync_log` + port batch writer**

**Description:** Tambahkan kolom `payload` jsonb (nullable) ke
`provisioning_sync_log` agar argumen aksi (SSID/password WiFi, flag WAN) bertahan
sampai Sync Engine memprosesnya, dan pastikan kontrak port penulisan sync
menerima payload.

**Acceptance criteria:**
- [ ] Migrasi 000029 menambah kolom `payload` jsonb nullable ke
      `provisioning_sync_log` (lihat §Migrasi Database).
- [ ] `internal/adapter/postgres/models.go` merefleksikan kolom baru
      (tipe jsonb → representasi Go yang sesuai, mis. `datatypes.JSON`).
- [ ] Kontrak port penulisan sync (`internal/port/provisioning_sync_repository.go`
      milik Issue 01, atau perluasan entri sync-nya) menerima field `Payload`
      opsional per baris; bila reuse dari Issue 01, cukup perluas struct
      entri-nya, jangan buat port baru.
- [ ] Sync Engine (Issue 01) membaca `payload` dan meneruskan isinya sebagai
      `command.Command.Args` (kunci logis) ke penerjemah driver Task 1 — catat
      titik integrasi ini di PR; bila Issue 01 belum mendukung `payload`, koordinasi
      lewat perluasan minimal, bukan menduplikasi logika dispatch.
- [ ] `context.Context` parameter pertama pada semua method port.

**Files likely touched:** `migrations/000029_add_payload_to_provisioning_sync_log.up.sql`,
`migrations/000029_add_payload_to_provisioning_sync_log.down.sql`,
`internal/adapter/postgres/models.go`,
`internal/port/provisioning_sync_repository.go`.

**Dependencies:** Issue 01 (kontrak sync writer), Issue 08 (kolom
`genieacs_device_id`).

**Estimated scope:** Small

---

**Task 4: Usecase network — reboot / set WAN / set WiFi (tulis sync log)**

**Description:** Buat tiga usecase orkestrasi yang memvalidasi subscription +
keterhubungan CPE, lalu menulis satu baris `provisioning_sync_log`
`target_type='genieacs_tr069'` `pending` dengan `action`/`payload` yang benar.
Usecase TIDAK menyentuh driver/HITL — itu milik Sync Engine (K4).

**Acceptance criteria:**
- [ ] `RebootCPE`, `SetCPEWAN`, `SetCPEWiFi` di `internal/usecase/network/`
      (file terpisah per verb sesuai §1.4, mis. `reboot_cpe.go`, `set_cpe_wan.go`,
      `set_cpe_wifi.go`).
- [ ] Setiap usecase: muat subscription → guard clause: `404`/error bila tidak
      ada, error khusus bila `genieacs_device_id` kosong → tulis satu baris sync
      `pending` (device_id = `genieacs_device_id`, action & payload sesuai tabel
      REST) → kembalikan `sync_log_id`.
- [ ] `RebootCPE` → `action='update'`, `payload=NULL`. `SetCPEWAN` →
      `action=enable|disable` sesuai flag, `payload={"enabled":bool}`.
      `SetCPEWiFi` → `action='update'`, `payload={"ssid","password"}`.
- [ ] Usecase TIDAK memanggil `port.DeviceDriver` langsung (K4).
- [ ] Nilai payload tidak divalidasi ulang untuk aturan TR-069 (itu urusan
      driver); usecase hanya memastikan field wajib ada (validasi format dasar
      sudah di DTO handler).
- [ ] `context.Context` parameter pertama; error posisi terakhir; wrap `%w`;
      error konektivitas CPE kosong pakai sentinel (mis.
      `ErrCPENotLinked`) di domain/usecase yang sesuai.
- [ ] Table-driven test (mock repository): happy path tiap usecase, subscription
      tidak ada, `genieacs_device_id` kosong, verifikasi baris sync yang ditulis
      (action + payload) benar.

**Files likely touched:** `internal/usecase/network/reboot_cpe.go`,
`internal/usecase/network/set_cpe_wan.go`,
`internal/usecase/network/set_cpe_wifi.go`, test di folder sama, dan sentinel
error di `internal/domain/subscription/errors.go` bila diperlukan.

**Dependencies:** Task 3.

**Estimated scope:** Medium

---

**Task 5: Usecase network — baca parameter CPE dari cache (opsional refresh)**

**Description:** Buat usecase baca yang mengambil parameter TR-069 dari cache
lokal `acs_devices` (Issue 08) dan, bila diminta, menulis baris `refreshObject`
`pending` untuk memicu pembaruan cache.

**Acceptance criteria:**
- [ ] `GetCPEParameters` di `internal/usecase/network/get_cpe_parameters.go`.
- [ ] Membaca baris `acs_devices` yang direferensikan `genieacs_device_id` lewat
      port repository Issue 08 (`acs_devices` repository) — TIDAK memanggil NBI
      GenieACS langsung (tahan banting saat GenieACS down; ini alasan cache ada,
      Temuan C §4.3).
- [ ] Bila subscription tak punya `genieacs_device_id` atau baris `acs_devices`
      belum pernah terisi → error yang dipetakan ke `409` di handler.
- [ ] Bila `refresh=true`, tulis satu baris `provisioning_sync_log`
      `genieacs_tr069` `action='update'` `payload={"task":"refreshObject"}`
      (read-only di driver) dan kembalikan `sync_log_id`-nya bersama snapshot
      cache.
- [ ] `context.Context` parameter pertama; error posisi terakhir; wrap `%w`.
- [ ] Table-driven test (mock): cache ada tanpa refresh, cache ada dengan
      refresh (baris sync ditulis), tanpa `genieacs_device_id` (error), cache
      kosong (error).

**Files likely touched:** `internal/usecase/network/get_cpe_parameters.go`, test
di folder sama.

**Dependencies:** Task 3, Issue 08 (repository `acs_devices`).

**Estimated scope:** Small

---

**Task 6: Handler REST + DTO + wiring route + RBAC**

**Description:** Ekspos empat usecase sebagai endpoint, dengan DTO
request/response, binding+validasi, mapping error→status HTTP, dan pendaftaran
route ber-RBAC.

**Acceptance criteria:**
- [ ] Handler ditambahkan di `internal/adapter/http/subscription_handler.go`
      (reuse file handler subscription yang sudah ada — tambah method, jangan buat
      file baru) atau, bila file sudah >~400 baris (§1.4), pecah ke
      `internal/adapter/http/cpe_handler.go` dan nyatakan alasannya di PR.
- [ ] DTO di `internal/adapter/http/dto/` (mis. `cpe_reboot.go`, `cpe_wan.go`,
      `cpe_wifi.go`, `cpe_parameters.go`): request sesuai tabel REST, validasi
      `enabled` wajib (WAN), `ssid` 1–32 & `password` ≥8 (WiFi). Response WiFi
      TIDAK menyertakan `password`.
- [ ] Reboot/WAN/WiFi mengembalikan **202 Accepted** + `sync_log_id`; GET
      parameters mengembalikan **200 OK** + snapshot cache.
- [ ] Route didaftarkan di `internal/adapter/http/router.go` di bawah group
      `/api/v1` ber-middleware `AuthRequired` + `RBACRequired`.
- [ ] Mapping error→HTTP: `404` not found, `409` CPE tak terhubung/cache kosong,
      `400` validasi, `403` role kurang, `500` sisanya.
- [ ] Test handler `httptest` (K7): 202 untuk reboot/WAN/WiFi, 200 untuk GET,
      400 validasi (enabled/ssid/password), 409 tanpa `genieacs_device_id`.

**Files likely touched:** `internal/adapter/http/subscription_handler.go` (atau
`cpe_handler.go`), `internal/adapter/http/dto/cpe_*.go`,
`internal/adapter/http/router.go`.

**Dependencies:** Task 4, Task 5.

**Estimated scope:** Medium

---

**Task 7: RBAC policy + konfigurasi + wiring `main.go`**

**Description:** Tambahkan entri Casbin untuk aksi CPE dan wiring usecase/handler
di startup, tanpa membaca env langsung di dalam usecase.

**Acceptance criteria:**
- [ ] `configs/rbac_policy.csv` memuat entri: reboot & set WiFi untuk
      superadmin/owner/admin/teknisi; set WAN untuk superadmin/owner/admin (bukan
      teknisi); baca parameter untuk semua termasuk staff. (K3: staff hanya baca.)
- [ ] Objek/aksi RBAC baru konsisten dengan pola yang sudah ada (mis. objek
      `cpe`, aksi `reboot`/`set_wan`/`set_wifi`/`read`) — nyatakan penamaan di PR.
- [ ] Usecase & handler di-wire di `cmd/server/main.go` (registry, repository
      `acs_devices`, sync writer di-inject lewat konstruktor).
- [ ] Bila ada parameter tunable (mis. batas panjang SSID/password sudah di DTO;
      poll interval sudah di driver), tidak menambah config yang tidak dipakai.

**Files likely touched:** `configs/rbac_policy.csv`, `cmd/server/main.go`.

**Dependencies:** Task 6.

**Estimated scope:** Small

---

**Task 8: Dokumentasi API + DATABASE-SCHEMA**

**Description:** Perbarui kontrak API dan skema DB agar mencerminkan endpoint +
kolom baru.

**Acceptance criteria:**
- [ ] `api/openapi.yaml` memuat empat endpoint baru dengan request/response body
      dan kode status (202/200/400/403/404/409/500). Response WiFi tanpa
      `password`.
- [ ] `DATABASE-SCHEMA.md` §6.3 diperbarui: kolom `payload` jsonb pada
      `provisioning_sync_log` dijelaskan (fungsi: membawa argumen aksi seperti
      SSID/password WiFi & flag WAN; nullable), plus catatan bahwa
      `external_reference` diisi task ID GenieACS untuk `genieacs_tr069` (§7.3).
- [ ] Bila membuat dokumen baru di root/docs, ditautkan dari `README.md` root pada
      commit yang sama (AGENTS.md §1.5). Issue ini kemungkinan tidak butuh ADR
      baru (ADR 0005 sudah mencakup polling GenieACS); bila memutuskan menambah
      ADR (mis. keputusan pemetaan `action`), mulai dari `0006` dan tautkan.

**Files likely touched:** `api/openapi.yaml`, `DATABASE-SCHEMA.md`,
opsional `README.md` + `docs/adr/0006-*.md`.

**Dependencies:** Task 6.

**Estimated scope:** Small

---

## Migrasi Database

Ada satu perubahan skema (penambahan kolom, bukan tabel baru):

- **Migrasi 000029** — nomor lanjut dari 000021.
  - File up: `migrations/000029_add_payload_to_provisioning_sync_log.up.sql`
  - File down: `migrations/000029_add_payload_to_provisioning_sync_log.down.sql`
  - **Perubahan:** tambah satu kolom `payload` bertipe **jsonb**, **nullable**,
    default NULL, ke tabel `provisioning_sync_log`.
  - **Alasan:** `provisioning_sync_log` saat ini tidak punya field untuk membawa
    argumen aksi. Reboot dan disable WAN cukup dideskripsikan oleh `action`, tapi
    set WiFi (SSID/password) dan set WAN (`enabled`) butuh membawa nilai konkret
    dari REST sampai Sync Engine memprosesnya. `payload` menampung nilai-nilai
    logis itu (BUKAN path TR-069 — path tetap milik driver).
  - **Down:** drop kolom `payload`. Aman selama tidak ada baris `pending` yang
    bergantung padanya saat rollback.
  - **Isi payload per aksi:** reboot → NULL; set WAN → `{"enabled": <bool>}`; set
    WiFi → `{"ssid": <str>, "password": <str>}`; refresh → `{"task":
    "refreshObject"}`.
  - **Catatan keamanan:** password WiFi disimpan di `payload` untuk baris yang
    berumur pendek (dihapus/diselesaikan Sync Engine). Jangan meng-log isi
    `payload` untuk aksi WiFi; pertimbangkan penghapusan/redaksi nilai `password`
    dari baris setelah `status='success'` (opsional, catat keputusan di PR).

Perbarui `DATABASE-SCHEMA.md` §6.3 di PR yang sama (kolom `payload` + peran
`external_reference` untuk `genieacs_tr069`).

## Verification

- [ ] `go build ./...` sukses.
- [ ] `go test ./internal/driver/genieacs/...` — penerjemahan reboot/WAN/WiFi/
      refresh, klasifikasi risiko, fail-safe product_class tak dikenal, body task
      benar, task ID ter-surface — hijau.
- [ ] `go test ./internal/usecase/network/...` — table-driven reboot/set WAN/set
      WiFi/get parameters + kasus error — hijau.
- [ ] `go test ./internal/adapter/postgres/...` — migrasi 000029 up/down bersih,
      round-trip `payload` jsonb (testcontainers, K7) — hijau.
- [ ] `go test ./internal/adapter/http/...` — handler httptest 202/200/400/409 —
      hijau.
- [ ] `make lint` bersih (gofumpt, staticcheck, boundary import: driver TIDAK
      diimpor domain/usecase; pengetahuan TR-069 hanya di `internal/driver/genieacs/`).
- [ ] Smoke test manual (sebut sebagai curl): `POST` ke
      `/api/v1/subscriptions/:id/cpe/reboot` dengan `Authorization: Bearer <token
      teknisi>` → 202 + `sync_log_id`; query `provisioning_sync_log` → satu baris
      `genieacs_tr069` `action=update` `pending`. `POST .../cpe/wan` body
      `{"enabled":false}` (token admin) → 202, baris `action=disable`
      `payload={"enabled":false}`. `PUT .../cpe/wifi` body `{"ssid","password"}`
      (token teknisi) → 202, response TANPA password. `GET .../cpe/parameters`
      (token staff) → 200 + snapshot cache; `?refresh=true` → 200 +
      `refresh_sync_log_id`.
- [ ] Verifikasi RBAC: `set WAN` dengan token teknisi → 403; `reboot` dengan token
      staff → 403; `GET parameters` dengan token staff → 200.
- [ ] Verifikasi integrasi cascade Issue 04: baris `genieacs_tr069`
      `action=disable` yang ditulis suspend cascade benar-benar diproses Sync
      Engine menjadi task `setParameterValues` (disable WAN) dan `sync_log`
      menjadi `success`/`failed` dengan `external_reference` terisi task ID
      GenieACS.

## Definition of Done

- [ ] Empat endpoint CPE (`reboot`, `wan`, `wifi`, `parameters`) jalan dengan
      kode status & role sesuai tabel REST.
- [ ] Aksi CPE menulis baris `provisioning_sync_log` `genieacs_tr069` `pending`
      (K4); handler tidak memanggil `port.DeviceDriver` langsung.
- [ ] Pengetahuan TR-069 (task type + path parameter) hanya di
      `internal/driver/genieacs/`; fail-safe untuk `product_class` tak dikenal.
- [ ] Nama task NBI persis (`setParameterValues`/`refreshObject`/`reboot`/
      `getParameterValues`), set WiFi/WAN fan-out multi-path (K13, index 5G
      dinamis), dan opsi `?connection_request` dimodelkan di driver.
- [ ] Set WiFi/WAN diikuti `refreshObject` otomatis di driver agar cache
      `acs_devices` konsisten; satu operasi CPE = satu baris `sync_log` (task NBI
      primer mengisi `external_reference`).
- [ ] `Classify` benar: reboot & `setParameterValues` destruktif (HITL),
      `getParameterValues`/`refreshObject` read-only.
- [ ] `command.Result` men-surface task ID GenieACS → Sync Engine mengisi
      `external_reference` (§7.3).
- [ ] Kolom `payload` jsonb ada (migrasi 000029), membawa argumen WiFi/WAN;
      `DATABASE-SCHEMA.md` §6.3 diperbarui.
- [ ] Endpoint baca parameter membaca cache `acs_devices` (Issue 08), tahan saat
      GenieACS down; refresh opsional menulis baris `refreshObject`.
- [ ] Cascade suspend Issue 04 (`genieacs_tr069` disable WAN) kini tereksekusi
      end-to-end.
- [ ] Password WiFi tidak pernah muncul di response/log.
- [ ] RBAC benar (K3): staff hanya baca, teknisi tidak boleh set WAN.
- [ ] Semua test hijau, `make lint` bersih, `api/openapi.yaml` diperbarui.
