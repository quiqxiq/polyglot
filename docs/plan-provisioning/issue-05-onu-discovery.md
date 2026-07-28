# Issue 05: ONU Discovery Queue

## Konteks

Temuan A/B di `ANALISIS-PROVISIONING-REPO-REFERENSI.md` menyoroti satu celah operasional pada alur instalasi ISP berbasis GPON: ketika sebuah ONU/ONT dipasang di rumah pelanggan dan disambungkan ke port PON, perangkat itu **terlihat secara fisik** oleh OLT (muncul di tabel ONU yang belum terotorisasi / *unconfigured* / *auto-find*), tetapi **belum diprovisioning** — belum punya ONU ID, belum diikat ke profil layanan. Selama ini teknisi harus membaca serial number (SN) ONU secara manual dari stiker perangkat lalu mengetiknya kembali saat mengaktifkan langganan. Ini rawan salah ketik SN (karakter mirip seperti `0/O`, `1/I`), yang berujung provisioning gagal atau salah bind ke ONU lain di PON yang sama.

Issue ini menutup celah tersebut dengan sebuah **antrian penemuan ONU** (`onu_discovery_queue`): hasil deteksi ONU tak-terotorisasi milik OLT disimpan sebagai daftar kandidat. Teknisi memilih SN dari daftar ini (bukan mengetik ulang), lalu meng-*bind* baris antrian ke sebuah `subscription`.

Temuan field-tested dari 5 repo referensi (`REFERENCES.md`): deteksi ONU unconfigured **tidak** seragam SNMP. Ada **dua** mekanisme dan issue ini wajib mendukung keduanya:
- **(a) SNMP walk `unauth_sn_table`** — hanya untuk merk yang memiliki OID tabel unauth: ZTE (C300/C600), Huawei, Fiberhome. Flag ketersediaan per merk dibaca dari katalog Issue 07 (`has_unauth_sn_table`, lihat K13).
- **(b) Fallback CLI** — merk yang **tidak** punya OID unauth (Hioso EPON, HSGQ, VSOL, BDCOM, C-Data) **wajib** memakai command CLI, dan formatnya berbeda GPON vs EPON: `show gpon onu uncfg` (gaya ZTE), `display ont autofind` (gaya Huawei), atau `show onu unregister`. Pemilihan command spesifik vendor hidup di `internal/driver/<vendor>/commands.go` (`Translate`), bukan usecase.

Setiap baris antrian menyimpan **sumber deteksi** (`snmp` atau `cli`) plus `onu_type`/`onu_model` mentah apa adanya dari OLT, supaya jejak asal data jelas dan parsing merk campuran bisa diaudit. Selain itu, ditemukan bahwa provisioning OLT sebenarnya butuh **dua** identitas per ONU — `onu_pon_port` (port PON tempat ONU tersambung) dan `onu_id` (nomor urut ONU pada port itu setelah diotorisasi) — bukan cuma serial number. Karena itu kolom `subscriptions.onu_pon_port` dan `subscriptions.onu_id` ditambahkan (lihat `DATABASE-SCHEMA.md` §6.1) agar data instalasi lengkap sebelum langkah provisioning OLT (Issue terkait) dijalankan.

Perlu ditegaskan: antrian ini adalah **inventaris kandidat**, bukan jalur eksekusi ke perangkat. Pengikatan (bind) hanya mengisi kolom pada `subscriptions` dan menandai baris antrian — provisioning OLT yang sebenarnya (mengotorisasi ONU, menetapkan `onu_id`) tetap mengalir lewat pola sinkronisasi standar (K4) di issue provisioning OLT, bukan di sini.

## Prasyarat

- **Issue 01 (Sync Engine + `provisioning_sync_log`)** — pola sinkronisasi K4 dan konvensi audit dipakai sebagai referensi alur; issue ini sendiri belum menulis `sync_log` (bind hanya menyiapkan data), tapi endpoint scan bersifat aksi-ke-perangkat sehingga mengikuti konvensi 202.
- **Foundation devices/subscriptions** — tabel `devices` (OLT sebagai device) dan `subscriptions` (migrasi 000009) sudah ada.
- **Driver OLT** — `internal/driver/zteolt` (`snmp.go`) dan `internal/driver/huaweiolt` sudah ada; kemampuan SNMP walk generik disempurnakan di **Issue 07 (genericsnmp)**. Issue ini boleh mendarat lebih dulu dengan dukungan ZTE/Huawei dan menyediakan titik ekstensi untuk vendor lain.
- **Registry** — `internal/registry/registry.go` `Get(ctx, deviceID)` untuk memperoleh driver OLT dari `olt_device_id`.
- **Scheduler foundation** — mekanisme penjadwalan periodik (issue scheduler); scan periodik didaftarkan ke sana. Bila scheduler belum ada saat implementasi, endpoint scan manual tetap fungsional dan job periodik ditandai TODO terhubung.

## Ruang Lingkup

**In scope:**
- Migrasi 000024 (`onu_discovery_queue`) dan 000025 (kolom `subscriptions.onu_pon_port`, `onu_id`).
- Domain `DiscoveredONU` di `internal/domain/provisioning/`.
- Port + implementasi repo Postgres untuk `onu_discovery_queue`.
- Usecase network: `scan` (deteksi ONU unauth via SNMP **atau** fallback CLI → upsert queue), `bind`, `ignore`.
- Deteksi ONU unconfigured **dua jalur**: (a) SNMP walk `unauth_sn_table` untuk merk yang mendukungnya (ZTE C300/C600, Huawei, Fiberhome); (b) fallback CLI (`show gpon onu uncfg` / `display ont autofind` / `show onu unregister`, beda GPON vs EPON) untuk merk tanpa OID unauth (Hioso EPON, HSGQ, VSOL, BDCOM, C-Data). Pemilihan jalur per merk mengikuti flag katalog Issue 07.
- Handler REST + DTO untuk list/scan/bind/ignore.
- Registrasi job scan periodik ke scheduler + entri RBAC.

**Out of scope:**
- Provisioning OLT yang sebenarnya (otorisasi ONU, penetapan `onu_id` di perangkat) — issue provisioning OLT.
- Driver `genericsnmp` baru — Issue 07 (di sini hanya konsumen antarmuka SNMP walk yang sudah ada).
- UI/frontend pemilihan SN.

## REST API

Base path: `/api/v1/`. Aksi yang menyentuh perangkat (scan) mengembalikan **202 Accepted**. Operasi murni-database (list, bind, ignore) mengembalikan 200.

| Method | Path | Tujuan | Role minimum |
|---|---|---|---|
| GET | `/api/v1/onu-discovery` | Daftar ONU hasil discovery, dengan filter | teknisi |
| POST | `/api/v1/onu-discovery/scan` | Picu SNMP walk ONU unconfigured pada satu OLT | teknisi |
| POST | `/api/v1/onu-discovery/:id/bind` | Ikat baris antrian ke sebuah subscription | teknisi |
| POST | `/api/v1/onu-discovery/:id/ignore` | Tandai baris antrian diabaikan | teknisi |

**GET `/api/v1/onu-discovery`**
- Query params (semua opsional): `olt_device_id` (uuid), `status` (`pending`/`bound`/`ignored`), `pon_port` (text). Sertakan paginasi standar (`limit`, `offset`) sesuai konvensi list yang sudah dipakai handler lain.
- Response 200: array objek discovery — field `id`, `olt_device_id`, `pon_port`, `serial_number`, `onu_type` (nullable), `onu_model` (nullable, tipe/model mentah dari OLT), `detection_source` (`snmp`/`cli`), `detected_at`, `status`, `bound_subscription_id` (nullable). Bungkus dalam amplop list standar (data + meta paginasi).
- Gagal: 400 bila `status` di luar enum; 401/403 bila role kurang.

**POST `/api/v1/onu-discovery/scan`**
- Request body penting: `olt_device_id` (uuid, wajib) — OLT yang akan di-walk. Opsional `pon_port` untuk membatasi walk ke satu port.
- Response **202 Accepted**: menyertakan ringkasan bahwa scan dijalankan; sertakan `id` referensi audit bila scan diproses lewat jalur sync/audit, atau ringkasan jumlah baris ter-*upsert* bila scan sinkron cepat. Ikuti konvensi 202 + identifier proses agar konsisten dengan aksi-ke-perangkat lain.
- Gagal: 404 bila `olt_device_id` tidak ditemukan / bukan device tipe OLT; 502/504 bila SNMP walk ke OLT gagal/timeout (dicatat, bukan panic); 403 bila role kurang.

**POST `/api/v1/onu-discovery/:id/bind`**
- Path param `:id` = id baris antrian.
- Request body penting: `subscription_id` (uuid, wajib), `onu_id` (text, wajib — nomor ONU yang akan/ sudah ditetapkan pada PON). `pon_port` dan `serial_number` diambil dari baris antrian, tidak dikirim ulang.
- Efek: set `subscriptions.onu_serial_number` = `serial_number` baris antrian, `subscriptions.onu_pon_port` = `pon_port` baris antrian, `subscriptions.onu_id` = `onu_id` dari body; set baris antrian `status='bound'` dan `bound_subscription_id=subscription_id`. Dilakukan dalam satu transaksi.
- Response 200 (atau 202 bila implementasi memilih langsung men-*trigger* `sync_log` provisioning OLT pada issue lanjutan — default issue ini **200**, murni update data). Kembalikan representasi baris antrian terbaru + ringkasan subscription yang diperbarui.
- Gagal: 404 bila `:id` atau `subscription_id` tidak ada; 409 bila baris antrian `status != 'pending'` atau `serial_number` sudah terikat ke subscription lain; 422 bila subscription sudah punya ONU lain terikat; 403 bila role kurang.

**POST `/api/v1/onu-discovery/:id/ignore`**
- Path param `:id` = id baris antrian. Tanpa body wajib (opsional alasan `reason` untuk audit).
- Efek: set `status='ignored'`. Baris `ignored` tidak muncul di daftar default `pending` dan tidak akan di-*upsert* ulang ke `pending` oleh scan berikutnya (lihat aturan upsert Task 4).
- Response 200: representasi baris antrian terbaru.
- Gagal: 404 bila `:id` tidak ada; 409 bila `status='bound'` (baris yang sudah terikat tidak boleh di-ignore); 403 bila role kurang.

## Tasks

**Task 1: Migrasi 000024 — tabel `onu_discovery_queue`**

**Description:** Buat tabel antrian penemuan ONU beserta constraint unik dan indeks pendukung filter.

**Acceptance criteria:**
- [ ] File `migrations/000024_create_onu_discovery_queue.up.sql` dan `.down.sql` berpasangan.
- [ ] Kolom: `id` uuid PK; `olt_device_id` uuid FK → `devices(id)`; `pon_port` text; `serial_number` text; `onu_type` text nullable; `onu_model` text nullable (tipe/model mentah dari OLT); `detection_source` text CHECK IN (`snmp`,`cli`) — menandai apakah baris ditemukan lewat SNMP walk unauth atau fallback CLI; `detected_at` timestamptz; `status` text CHECK IN (`pending`,`bound`,`ignored`) default `pending`; `bound_subscription_id` uuid FK → `subscriptions(id)` nullable.
- [ ] Constraint `UNIQUE(olt_device_id, serial_number)`.
- [ ] Indeks untuk kolom filter yang sering dipakai (`olt_device_id`, `status`, `pon_port`).
- [ ] `.down.sql` men-*drop* tabel secara bersih.
- [ ] Perilaku FK saat subscription/ device dihapus dinyatakan eksplisit (mis. `ON DELETE SET NULL` untuk `bound_subscription_id`, `ON DELETE CASCADE`/`RESTRICT` untuk `olt_device_id` sesuai konvensi tabel lain).

**Files likely touched:** `migrations/000024_create_onu_discovery_queue.up.sql`, `migrations/000024_create_onu_discovery_queue.down.sql`.

**Dependencies:** —

**Estimated scope:** Small

---

**Task 2: Migrasi 000025 — kolom `subscriptions.onu_pon_port` dan `onu_id`**

**Description:** Tambah dua kolom identitas ONU pada `subscriptions` yang dibutuhkan provisioning OLT.

**Acceptance criteria:**
- [ ] File `migrations/000025_add_onu_pon_port_and_onu_id_to_subscriptions.up.sql` dan `.down.sql` berpasangan.
- [ ] `up.sql`: `ALTER TABLE subscriptions ADD COLUMN onu_pon_port text` dan `ADD COLUMN onu_id text` (keduanya nullable — belum terisi sampai bind).
- [ ] `.down.sql` men-*drop* kedua kolom.
- [ ] Tidak menyentuh kolom lain yang sudah ada.

**Files likely touched:** `migrations/000025_add_onu_pon_port_and_onu_id_to_subscriptions.up.sql`, `.down.sql`.

**Dependencies:** —

**Estimated scope:** Small

---

**Task 3: Domain `DiscoveredONU` + perluasan domain `subscription`**

**Description:** Definisikan entity domain untuk baris antrian dan tambahkan field `OnuPonPort`/`OnuID` ke domain subscription.

**Acceptance criteria:**
- [ ] Tipe `DiscoveredONU` di `internal/domain/provisioning/discovered_onu.go` dengan field yang mencerminkan kolom tabel (id, olt device id, pon port, serial number, onu type, onu model, detection source, detected at, status, bound subscription id) — tipe murni, tanpa I/O, tanpa import library eksternal (selain stdlib).
- [ ] Enum status sebagai konstanta/typed string domain (`StatusPending`, `StatusBound`, `StatusIgnored`) dengan penamaan sesuai §2.3 (bukan ALL_CAPS).
- [ ] Enum sumber deteksi sebagai typed string domain (`DetectionSourceSNMP`, `DetectionSourceCLI`) — akronim `SNMP`/`CLI` konsisten kapital penuh (§2.2).
- [ ] Method/validasi transisi status bila relevan (mis. hanya `pending`→`bound`/`ignored` yang sah) diletakkan di domain, bukan usecase.
- [ ] `subscription.Subscription` menambah field `OnuPonPort string` dan `OnuID string`; doc comment field mengikuti §7.
- [ ] Setiap identifier exported punya doc comment dimulai dengan namanya sendiri.

**Files likely touched:** `internal/domain/provisioning/discovered_onu.go`, `internal/domain/provisioning/errors.go` (bila ada sentinel error transisi), `internal/domain/subscription/subscription.go`.

**Dependencies:** —

**Estimated scope:** Small

---

**Task 4: Port + repo Postgres `onu_discovery_queue`**

**Description:** Kontrak repository di `port/` dan implementasinya di adapter Postgres, termasuk upsert idempoten untuk hasil scan.

**Acceptance criteria:**
- [ ] Interface `port.ONUDiscoveryRepository` di `internal/port/onu_discovery_repository.go` dengan operasi minimal: `List` (dengan filter olt/status/pon_port + paginasi), `GetByID`, `Upsert` (batch hasil scan), `MarkBound`, `MarkIgnored`.
- [ ] Implementasi di `internal/adapter/postgres/onu_discovery_repository.go` dengan mapping model ↔ domain.
- [ ] `Upsert` bersifat idempoten pada `UNIQUE(olt_device_id, serial_number)`: baris yang sudah ada dan berstatus `bound` **atau** `ignored` **tidak** direset ke `pending`; hanya baris baru yang di-*insert* sebagai `pending`, dan `detected_at`/`onu_type`/`onu_model`/`pon_port`/`detection_source` baris `pending` yang sudah ada boleh di-*refresh*.
- [ ] `context.Context` sebagai parameter pertama; error di-*wrap* `%w` dengan konteks operasi; `sql.ErrNoRows` dipetakan ke sentinel domain.
- [ ] `MarkBound` menerima subscription id + menyetel status transaksional (transaksi lintas-tabel diatur di usecase, lihat Task 5 — repo menyediakan primitifnya).

**Files likely touched:** `internal/port/onu_discovery_repository.go`, `internal/adapter/postgres/onu_discovery_repository.go`, `internal/adapter/postgres/models.go`.

**Dependencies:** Task 1, Task 3.

**Estimated scope:** Medium

---

**Task 5: Usecase network — scan, bind, ignore**

**Description:** Orkestrasi tiga operasi: memicu SNMP walk unauth via driver OLT lalu upsert antrian; mengikat baris ke subscription secara transaksional; menandai ignore.

**Acceptance criteria:**
- [ ] `internal/usecase/network/scan_onu_discovery.go` berisi `ScanONUDiscovery`: ambil driver OLT dari registry via `olt_device_id`, panggil satu operasi abstrak "temukan ONU unconfigured" pada driver — driver-lah yang memutuskan (dari katalog Issue 07) apakah memakai SNMP walk `unauth_sn_table` atau fallback CLI (`show gpon onu uncfg` / `display ont autofind` / `show onu unregister`, beda GPON vs EPON). Petakan hasil ke `[]DiscoveredONU` beserta `detection_source` yang dilaporkan driver, panggil repo `Upsert`. Kegagalan (SNMP atau CLI) di-*wrap* dan dikembalikan sebagai error (bukan panic).
- [ ] Pengetahuan OID/tabel unauth **maupun** command CLI unconfigured (uncfg/autofind/unregister) beserta pemilihan GPON vs EPON **tidak** ada di usecase — usecase hanya memanggil satu method abstrak driver; detail SNMP/CLI hidup di `internal/driver/<vendor>/commands.go` (`Translate`/`Classify`), sesuai K1/K13/§1.2.
- [ ] `internal/usecase/network/bind_onu.go` berisi `BindONU`: validasi baris antrian `status='pending'`, validasi subscription ada, dalam **satu transaksi** set `subscriptions.onu_serial_number`/`onu_pon_port`/`onu_id` dan set baris antrian `bound` + `bound_subscription_id`. Konflik (baris bukan `pending`, subscription sudah punya ONU) dikembalikan sebagai error yang dapat dipetakan ke 409/422.
- [ ] `internal/usecase/network/ignore_onu.go` berisi `IgnoreONU`: tolak bila `status='bound'`, selain itu set `ignored`.
- [ ] Guard clause / early return dipakai (§3.2); error selalu return terakhir.
- [ ] Table-driven test untuk ketiga usecase (skenario sukses + tiap konflik/validasi).

**Files likely touched:** `internal/usecase/network/scan_onu_discovery.go`, `internal/usecase/network/bind_onu.go`, `internal/usecase/network/ignore_onu.go`, `internal/usecase/network/*_test.go`.

**Dependencies:** Task 3, Task 4, dukungan SNMP walk unauth di driver OLT (Task 8).

**Estimated scope:** Large

---

**Task 6: Handler REST + DTO**

**Description:** Empat endpoint (list/scan/bind/ignore) beserta DTO request/response dan pemetaan status code sesuai kontrak REST API di atas.

**Acceptance criteria:**
- [ ] `internal/adapter/http/onu_discovery_handler.go` dengan handler untuk keempat rute; didaftarkan di `router.go` di bawah `/api/v1/onu-discovery`.
- [ ] DTO di `internal/adapter/http/dto/` untuk request scan/bind (dan response list/detail) — tidak memakai `map[string]interface{}`.
- [ ] Handler **tidak** memanggil `port.DeviceDriver` langsung — hanya memanggil usecase (K4).
- [ ] Mapping status: scan → 202; list/ignore → 200; bind → 200; error usecase dipetakan ke 400/404/409/422/502/504 sesuai kontrak; validasi query `status` di luar enum → 400.
- [ ] Handler test dengan `httptest` untuk tiap rute (sukses + minimal satu jalur error).

**Files likely touched:** `internal/adapter/http/onu_discovery_handler.go`, `internal/adapter/http/router.go`, `internal/adapter/http/dto/onu_discovery.go`, `internal/adapter/http/onu_discovery_handler_test.go`.

**Dependencies:** Task 5.

**Estimated scope:** Medium

---

**Task 7: RBAC + scheduler scan periodik**

**Description:** Tambahkan policy Casbin untuk rute baru dan daftarkan job scan periodik ke scheduler.

**Acceptance criteria:**
- [ ] Entri di `configs/rbac_policy.csv`: keempat rute dapat diakses `teknisi` ke atas (teknisi, admin, owner, superadmin); `staff` hanya `GET /api/v1/onu-discovery` (baca), tidak boleh scan/bind/ignore.
- [ ] Job scheduler periodik memanggil `ScanONUDiscovery` untuk tiap device bertipe OLT (interval dari config, mis. tiap N menit). Job memakai `context.Context` dengan pembatalan; kegagalan satu OLT tidak menghentikan OLT lain (di-*log*, lanjut).
- [ ] Bila foundation scheduler belum tersedia, titik registrasi ditandai TODO yang jelas dan endpoint scan manual tetap bekerja; ketiadaan scheduler tidak menyilentkan kegagalan (di-*log*).

**Files likely touched:** `configs/rbac_policy.csv`, file registrasi scheduler (mengikuti foundation scheduler), `internal/config/config.go` (interval scan).

**Dependencies:** Task 5, Task 6.

**Estimated scope:** Medium

---

**Task 8: Kemampuan deteksi ONU unconfigured pada driver OLT (SNMP unauth + fallback CLI)**

**Description:** Pastikan driver OLT mengekspos satu operasi abstrak "temukan ONU unconfigured" yang di-Translate ke SNMP walk `unauth_sn_table` untuk merk yang mendukungnya, atau ke command CLI uncfg/autofind/unregister untuk merk yang tidak punya OID unauth.

**Acceptance criteria:**
- [ ] Driver melaporkan hasil deteksi lewat satu operasi abstrak; implementasi memilih jalur per merk berdasarkan flag katalog Issue 07 (`has_unauth_sn_table`): **SNMP walk `unauth_sn_table`** untuk ZTE (C300/C600), Huawei, Fiberhome — OID spesifik vendor di driver, bukan usecase.
- [ ] **Fallback CLI** untuk merk tanpa OID unauth (Hioso EPON, HSGQ, VSOL, BDCOM, C-Data): command `show gpon onu uncfg` (gaya ZTE), `display ont autofind` (gaya Huawei), atau `show onu unregister` — pemilihan **beda GPON vs EPON** dan parsing outputnya hidup di `internal/driver/<vendor>/commands.go` (`Translate`), tidak di usecase. Untuk merk yang punya CLI-nya di driver telnet vendor (`zteolt`/`huaweiolt`) atau via `genericcli`, command dipilih di situ.
- [ ] Setiap baris hasil menandai `detection_source` (`snmp`/`cli`) sesuai jalur yang dipakai, dikembalikan ke usecase.
- [ ] Hasil dikembalikan sebagai struktur netral (pon_port, serial_number, onu_type, onu_model, detection_source) yang dipetakan usecase ke `DiscoveredONU` — driver tidak mengimpor domain sebagai dependensi baru bila melanggar boundary; gunakan tipe hasil di `port` atau tipe domain yang sudah disepakati.
- [ ] Vendor yang belum dikurasi mengikuti posture fail-safe: bila **baik** OID unauth **maupun** command CLI unconfigured tidak diketahui untuk merk itu, kembalikan error eksplisit "unsupported", bukan daftar kosong diam-diam (K13 — jangan hardcode, jangan diam-diam auto-pass).
- [ ] Bila deteksi ikut membaca status ONU (online/offline) dari OLT, jangan **hardcode** `status==1`/`1=online`: nilai/label "online" berbeda per merk dan antara EPON vs GPON, dan bisa **string**. Interpretasikan lewat himpunan `online_values` di katalog Issue 07 (K13), bukan literal di driver/usecase issue ini.
- [ ] Kontrak operasi ini dinyatakan di `internal/port/` (mis. perluasan port device/streaming atau port khusus) agar usecase tidak bergantung ke tipe konkret vendor maupun ke pilihan SNMP-vs-CLI.

**Files likely touched:** `internal/port/` (kontrak deteksi unconfigured), `internal/driver/zteolt/snmp.go`, `internal/driver/zteolt/telnet.go`, `internal/driver/zteolt/commands.go`, `internal/driver/huaweiolt/driver.go`, `internal/driver/huaweiolt/commands.go`, dan katalog Issue 07 (`internal/driver/genericsnmp/catalog.go`) untuk flag `has_unauth_sn_table`.

**Dependencies:** Task 3. Interoperabel dengan Issue 07 (genericsnmp) bila mendarat belakangan.

**Estimated scope:** Large

---

**Task 9: Perbarui DATABASE-SCHEMA.md dan dokumentasi**

**Description:** Cerminkan skema baru ke dokumen kanonik dan tautkan dokumen bila perlu.

**Acceptance criteria:**
- [ ] `DATABASE-SCHEMA.md`: tambahkan bagian baru infrastruktur/ONU yang mendeskripsikan `onu_discovery_queue` (semua kolom termasuk `onu_model` dan `detection_source` + constraint + enum status/detection_source) dan perbarui §6.1 `subscriptions` dengan kolom `onu_pon_port` dan `onu_id`.
- [ ] Pertimbangkan ADR baru `docs/adr/0007-onu-discovery-queue.md` yang menjelaskan keputusan "antrian discovery vs ketik manual SN" dan pemisahan bind (data) vs provisioning (device); bila dibuat, tautkan dari `README.md` root pada commit yang sama (§1.5).
- [ ] Perubahan skema dan dokumen berada dalam PR yang sama dengan migrasi.

**Files likely touched:** `DATABASE-SCHEMA.md`, `docs/adr/0007-onu-discovery-queue.md` (opsional), `README.md`.

**Dependencies:** Task 1, Task 2.

**Estimated scope:** Small

---

## Migrasi Database

Dua migrasi baru, melanjutkan dari 000021:

**000024 — `create_onu_discovery_queue`**
- File: `migrations/000024_create_onu_discovery_queue.up.sql` + `migrations/000024_create_onu_discovery_queue.down.sql`.
- Tabel `onu_discovery_queue` dengan kolom:
  - `id` — uuid, primary key.
  - `olt_device_id` — uuid, FK → `devices(id)` (perilaku hapus: cascade/restrict sesuai konvensi tabel referensi device lain).
  - `pon_port` — text, port PON tempat ONU terdeteksi.
  - `serial_number` — text, SN ONU dari tabel unauth OLT atau output CLI unconfigured.
  - `onu_type` — text, nullable (tipe/model ONU bila OLT melaporkannya).
  - `onu_model` — text, nullable (tipe/model mentah apa adanya dari OLT, sebelum normalisasi).
  - `detection_source` — text, CHECK IN (`snmp`,`cli`), menandai jalur penemuan baris (SNMP walk `unauth_sn_table` vs fallback CLI uncfg/autofind/unregister).
  - `detected_at` — timestamptz, waktu deteksi (di-*refresh* saat scan menemukan ulang baris `pending`).
  - `status` — text, CHECK IN (`pending`,`bound`,`ignored`), default `pending`.
  - `bound_subscription_id` — uuid, nullable, FK → `subscriptions(id)` (ON DELETE SET NULL).
- Constraint: `UNIQUE(olt_device_id, serial_number)`.
- Indeks: pada `olt_device_id`, `status`, `pon_port` untuk mendukung filter list.

**000025 — `add_onu_pon_port_and_onu_id_to_subscriptions`**
- File: `migrations/000025_add_onu_pon_port_and_onu_id_to_subscriptions.up.sql` + `.down.sql`.
- `ALTER TABLE subscriptions` menambah kolom `onu_pon_port` (text, nullable) dan `onu_id` (text, nullable). `.down.sql` men-*drop* keduanya.

Kedua perubahan skema dicerminkan ke `DATABASE-SCHEMA.md` (bagian baru infrastruktur/ONU untuk `onu_discovery_queue`; §6.1 untuk kolom subscriptions) dalam PR yang sama.

## Verification

- [ ] `go build ./...` sukses.
- [ ] `go test ./internal/domain/provisioning/... ./internal/domain/subscription/...` — validasi transisi status dan field baru.
- [ ] `go test ./internal/usecase/network/...` — table-driven untuk scan/bind/ignore (sukses + konflik 409/422 + deteksi gagal), termasuk kasus scan jalur SNMP (merk ber-`unauth_sn_table`) dan jalur fallback CLI (merk tanpa OID unauth) dengan `detection_source` yang benar per baris.
- [ ] `go test ./internal/adapter/postgres/...` — repo `onu_discovery_queue` dengan `testcontainers-go` (Postgres asli), khususnya idempotensi `Upsert` terhadap baris `bound`/`ignored`.
- [ ] `go test ./internal/adapter/http/...` — handler `httptest` untuk keempat rute.
- [ ] `make lint` bersih (gofumpt/goimports/staticcheck), termasuk akronim `ONU`/`ID` konsisten.
- [ ] Migrasi up lalu down bersih: jalankan `migrate up` ke 000025 lalu `migrate down` dua langkah tanpa error.
- [ ] Smoke test manual (curl, sebutan): `POST /api/v1/onu-discovery/scan` dengan `olt_device_id` valid → 202; `GET /api/v1/onu-discovery?status=pending` menampilkan hasil; `POST /api/v1/onu-discovery/:id/bind` dengan `subscription_id`+`onu_id` → 200 dan verifikasi kolom subscriptions terisi; ulangi scan → baris yang sudah `bound` tidak kembali ke `pending`; `POST /api/v1/onu-discovery/:id/ignore` pada baris `bound` → 409.

## Definition of Done

- [ ] Migrasi 000024 & 000025 berpasangan up/down dan idempoten-aman; `DATABASE-SCHEMA.md` diperbarui di PR yang sama.
- [ ] Domain `DiscoveredONU` + field `OnuPonPort`/`OnuID` pada subscription selesai dengan doc comment sesuai §7.
- [ ] Port + repo Postgres dengan `Upsert` idempoten dan mapping error benar (`%w`, sentinel).
- [ ] Usecase scan/bind/ignore mengikuti boundary (pengetahuan SNMP **dan** command CLI unconfigured di driver, bukan usecase) dan pola transaksi bind.
- [ ] Handler REST tidak memanggil driver langsung (K4), status code sesuai kontrak, RBAC teknisi+ (staff read-only) di `rbac_policy.csv`.
- [ ] Driver OLT mengekspos deteksi ONU unconfigured dua jalur (SNMP `unauth_sn_table` untuk ZTE/Huawei/Fiberhome + fallback CLI uncfg/autofind/unregister untuk Hioso/HSGQ/VSOL/BDCOM/C-Data, beda GPON/EPON) dengan posture fail-safe untuk vendor belum dikurasi; `detection_source` tercatat per baris.
- [ ] Job scan periodik terdaftar (atau TODO terhubung bila scheduler belum ada) tanpa menyilentkan kegagalan.
- [ ] Semua item Verification hijau.
