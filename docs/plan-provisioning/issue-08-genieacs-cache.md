# Issue 08: GenieACS Device Cache

## Konteks

Temuan C di `ANALISIS-PROVISIONING-REPO-REFERENSI.md` mencatat bahwa dashboard status ONU (perangkat CPE via TR-069) saat ini harus memanggil NBI (Northbound Interface) GenieACS secara langsung setiap kali halaman dibuka. Ini punya dua masalah: (1) latensi tinggi karena NBI GenieACS memproxy ke device tree yang besar, dan (2) ketergantungan keras — begitu GenieACS down atau lambat, seluruh tampilan status pelanggan TR-069 ikut mati. Padahal data yang benar-benar dipakai UI hanya sekumpulan kecil parameter (serial, manufacturer, product class, waktu inform terakhir, beberapa parameter status).

Solusinya adalah cache lokal berbentuk tabel baru `acs_devices` — mirror lokal dari device yang dikenal GenieACS, hanya menyimpan subset parameter yang dipakai UI (bukan seluruh device tree). Skema cache ini divalidasi lapangan dari built-in ACS `billing-rtrw` (lihat REFERENCES.md §C): **PK = `_id` GenieACS asli** (bukan surrogate lokal), kolom `serial_number`/`manufacturer`/`product_class`/`oui`/`params`/`tags`/`last_inform`. Satu konsekuensi penting: **status "online" BUKAN kolom tersimpan** — ia turunan dari `last_inform` + threshold (mis. 5 menit sejak inform terakhir = online). Menyimpan kolom `status` boolean akan cepat basi antar siklus polling; hitung on-the-fly saat baca. Cache ini disinkronkan oleh job polling terjadwal yang menarik daftar device dari NBI GenieACS. Polling di sini **sah** dan sesuai ADR 0005: TR-069 memang berbasis inform periodik, sehingga tidak ada mekanisme push seperti RouterOS `print follow`, dan mirror-by-polling adalah pola yang benar untuk kelas perangkat ini. Tabel task NBI GenieACS (`acs_tasks(device_id,name,payload,status)` di referensi) adalah analog langsung dari `provisioning_sync_log` Polyglot — issue ini TIDAK membuat tabel task terpisah; eksekusi aksi TR-069 tetap lewat sync-log → command_audit_log (K4), lihat Issue 01/09.

Selain cache, issue ini juga menutup celah linking. Di referensi, resolusi "subscription mana untuk device GenieACS mana" bukan satu kunci melainkan **3 strategi berjenjang** (REFERENCES.md §C: resolve by tag `{_tags}` → PPPoE username `$or` banyak path → serial ternormalisasi). Issue ini memodelkan keduanya: (a) kolom `genieacs_device_id` pada `subscriptions` sebagai **hasil resolusi yang di-cache** / tautan eksplisit (FK ke `acs_devices.id`) supaya dashboard tidak menebak tiap kali, dan (b) resolver berjenjang + usecase set-tag yang menanam tag `pppoe:<username>` ke GenieACS saat aktivasi sehingga strategi (a) tag berhasil untuk device baru (cross-link ke Issue 03/12). Ini melengkapi bagian DATABASE-SCHEMA.md §6.1 (subscriptions) dan menambah §7 (acs cache) yang baru.

## Prasyarat

- **Issue 01 (Sync Engine & Scheduler)** — issue ini memakai keputusan scheduler yang sama (ticker vs cron, lihat Open Question di Issue 01). Job sinkron `acs_devices` didaftarkan ke scheduler yang sama, bukan membuat mekanisme penjadwalan baru.
- **Driver `genieacs` yang sudah ada** (`internal/driver/genieacs/client.go` — NBI polling, `commands.go`, `errors.go`). Issue ini memperluas client-nya untuk mengambil daftar device + subset parameter, bukan membuat client TR-069 baru.
- Foundation domain `subscription.Subscription` dan repo Postgres subscription yang sudah ada (dari migrasi 000009) — dipakai untuk menambah field `genieacs_device_id` dan method linking.
- **Issue 03 (Subscription Provisioning) & Issue 12 (Subscriber Sessions)** — usecase set-tag di issue ini menanam tag `pppoe:<username>` di titik aktivasi PPPoE; username kanoniknya berasal dari Issue 03. Kalau Issue 03 belum ada, set-tag boleh dipicu terpisah tapi harus memakai username yang sama persis.
- Konvensi bersama K1, K3, K4, K6, K7 dari README folder ini; ditambah **K13** (identitas/katalog vendor = multi-path/multi-value, milik driver) untuk resolver PPPoE-username lintas banyak path TR-069.

## Ruang Lingkup

**In scope:**
- Tabel `acs_devices` (cache lokal) + migrasi up/down.
- Kolom `subscriptions.genieacs_device_id` (FK nullable) + migrasi up/down + update domain/repo/schema.
- Domain kecil `acsdevice` untuk merepresentasikan entri cache.
- Kontrak `port.ACSDeviceRepository` + implementasi Postgres.
- Usecase network `SyncGenieACSDevices` (upsert cache dari NBI, dibatasi ke daftar parameter yang dipakai UI).
- Resolver linking berjenjang (tag → PPPoE username → serial ternormalisasi) untuk menautkan subscription ↔ device GenieACS, dan usecase set-tag (menulis `_tags` penuh secara read-modify-write) yang menanam tag `pppoe:<username>` saat aktivasi.
- Pendaftaran job sinkron ke scheduler Issue 01 + endpoint pemicu manual.
- Handler REST list/detail cache, pemicu sync, dan linking subscription↔acs-device.

**Out of scope:**
- Mengubah alur provisioning TR-069 aktual (set-parameter, reboot ONU) — itu di issue provisioning GenieACS terpisah.
- Menyimpan seluruh device tree GenieACS — sengaja dibatasi subset parameter.
- Real-time push dari GenieACS — tidak ada; tetap polling (ADR 0005).
- UI/dashboard front-end — hanya menyediakan endpoint pembacanya.

## REST API

Base path: `/api/v1/`. Aksi yang memicu komunikasi ke GenieACS (sync) mengikuti konvensi aksi perangkat: balas **202 Accepted** dan referensikan pekerjaan sinkron. Endpoint baca murni dari cache lokal balas **200**.

| Method | Path | Tujuan | Role minimum |
|---|---|---|---|
| GET | `/api/v1/acs-devices` | List device dari cache lokal, dengan filter | staff |
| GET | `/api/v1/acs-devices/:id` | Detail satu device dari cache lokal | staff |
| POST | `/api/v1/acs-devices/sync` | Picu sinkron cache dari NBI GenieACS | admin |
| GET | `/api/v1/subscriptions/:id/acs-device` | Device GenieACS yang tertaut ke subscription | staff |
| PUT | `/api/v1/subscriptions/:id/acs-device` | Set/ubah `genieacs_device_id` subscription (opsional) | teknisi |
| POST | `/api/v1/subscriptions/:id/acs-device/resolve` | Jalankan resolver berjenjang (tag→PPPoE user→serial) & tautkan hasilnya | teknisi |

**GET `/api/v1/acs-devices`**
- Query param opsional: `manufacturer`, `product_class`, `serial` (cocok persis atau prefix — nyatakan di dokumentasi handler; default cocok persis; pencocokan serial dilakukan pada bentuk **ternormalisasi**, lihat Migrasi Database), `online` (boolean — filter turunan `last_inform` + threshold, bukan kolom tersimpan), plus paginasi `page`/`page_size`.
- Response 200: array objek cache berisi `id`, `serial_number`, `manufacturer`, `product_class`, `oui`, `connection_request_url`, `last_inform`, `online` (turunan boolean), `synced_at`, dan ringkasan `tags`. Bungkus dengan metadata paginasi (total, page, page_size) sesuai konvensi list handler lain.
- Gagal: 400 bila query param tidak valid.

**GET `/api/v1/acs-devices/:id`**
- Path param `:id` = device id GenieACS (juga PK cache).
- Response 200: objek cache lengkap termasuk `params` (subset parameter yang dipakai UI), `tags`, dan `online` (turunan dari `last_inform` + threshold, dihitung saat baca — bukan kolom).
- Gagal: 404 bila id tidak ada di cache.

**POST `/api/v1/acs-devices/sync`**
- Body kosong atau opsional `{ "device_id": "<id>" }` untuk sinkron satu device saja (tanpa itu = sinkron seluruh daftar).
- Response **202 Accepted**: kembalikan identifier pekerjaan sinkron (id/handle yang bisa dipakai untuk menelusuri hasil, konsisten dengan pola sync_log/job pada issue lain) + `status: "pending"`.
- Gagal: 502/503 bila NBI GenieACS tidak reachable saat pemicu sinkron langsung diproses; 403 bila role di bawah admin.

**GET `/api/v1/subscriptions/:id/acs-device`**
- Path param `:id` = subscription id.
- Response 200: objek cache device yang tertaut (mengikuti `subscriptions.genieacs_device_id`).
- Gagal: 404 bila subscription tidak ada, atau bila `genieacs_device_id` null / device belum ada di cache (nyatakan pesan yang membedakan "subscription tak ada" vs "belum tertaut").

**PUT `/api/v1/subscriptions/:id/acs-device`** (opsional)
- Body: `{ "genieacs_device_id": "<id | null>" }`. Nilai null = melepas tautan.
- Validasi: bila non-null, `genieacs_device_id` harus sudah ada di `acs_devices` (FK).
- Response 200: subscription setelah update (atau 204 tanpa body — pilih konsisten dengan handler subscription lain).
- Gagal: 404 subscription tak ada; 422 bila device id yang di-set tidak ada di cache.

**POST `/api/v1/subscriptions/:id/acs-device/resolve`**
- Body kosong. Menjalankan resolver berjenjang (Task 12): (a) cari device dengan tag `pppoe:<username subscription>`; (b) fallback cocokkan PPPoE username terhadap subset param TR-069 yang di-cache; (c) fallback serial ternormalisasi. Strategi pertama yang menghasilkan tepat satu device menang, dan `genieacs_device_id` di-set ke device itu.
- Response 200: objek berisi device yang di-resolusi + strategi mana yang berhasil (`matched_by`: `tag`/`pppoe`/`serial`).
- Gagal: 404 bila subscription tak ada; 409/422 bila resolusi ambigu (lebih dari satu kandidat, terutama pada serial pendek — lihat catatan false-positive di Migrasi Database) atau tidak ada kandidat sama sekali (biarkan tak tertaut, jangan tebak).

## Tasks

**Task 1: Migrasi tabel `acs_devices`**
**Description:** Buat tabel cache lokal mirror device GenieACS beserta migrasi turunnya.
**Acceptance criteria:**
- [ ] Migrasi baru bernomor lanjut dari 000021 (lihat bagian Migrasi Database) dengan pasangan up/down.
- [ ] Tabel `acs_devices` memuat kolom sesuai daftar di bagian Migrasi Database (PK `id` text = device id/`_id` GenieACS asli, **bukan** UUID/surrogate lokal).
- [ ] Kolom `oui` disertakan (dipakai referensi untuk identifikasi vendor), dan `serial_number` disimpan dalam bentuk **ternormalisasi** (uppercase, buang non-alfanumerik) + diindeks untuk lookup/join OLT↔ACS.
- [ ] **Tidak ada** kolom `status`/`online` tersimpan — status online adalah turunan `last_inform` + threshold, dihitung saat baca (lihat Konteks & Migrasi Database).
- [ ] `params` dan `tags` bertipe jsonb; `last_inform` dan `synced_at` timestamptz nullable sesuai ketentuan.
- [ ] File down menghapus tabel bersih (drop).
- [ ] DATABASE-SCHEMA.md ditambah §7 yang mendeskripsikan tabel ini pada PR yang sama.
**Files likely touched:** `migrations/0000NN_create_acs_devices_table.up.sql`, `migrations/0000NN_create_acs_devices_table.down.sql`, `DATABASE-SCHEMA.md`.
**Dependencies:** —
**Estimated scope:** Small

---

**Task 2: Migrasi ALTER `subscriptions.genieacs_device_id`**
**Description:** Tambah kolom FK nullable dari subscriptions ke acs_devices sebagai linking eksplisit.
**Acceptance criteria:**
- [ ] Migrasi berpasangan (up menambah kolom + FK ke `acs_devices(id)` `ON DELETE SET NULL`, down menghapus kolom/constraint) bernomor tepat setelah migrasi Task 1.
- [ ] Kolom `genieacs_device_id` text nullable, dengan index untuk lookup balik.
- [ ] DATABASE-SCHEMA.md §6.1 (subscriptions) diperbarui menambah kolom ini pada PR yang sama.
**Files likely touched:** `migrations/0000NN_add_genieacs_device_id_to_subscriptions.up.sql`, `.down.sql`, `DATABASE-SCHEMA.md`.
**Dependencies:** Task 1 (FK menunjuk `acs_devices`).
**Estimated scope:** Small

---

**Task 3: Domain `acsdevice`**
**Description:** Definisikan entity domain murni untuk satu entri cache device GenieACS.
**Acceptance criteria:**
- [ ] File `internal/domain/acsdevice/acsdevice.go` berisi tipe `Device` (atau nama setara non-stutter, mis. `acsdevice.Device`) dengan field: id, serial number (ternormalisasi), manufacturer, product class, oui, connection request URL, tags (map/tipe domain), params (map subset), last inform, synced at.
- [ ] Method domain murni `IsOnline(now, threshold) bool` menghitung status dari `last_inform` (tanpa I/O) — status online tidak pernah disimpan sebagai field/kolom.
- [ ] Helper normalisasi serial (uppercase + buang non-alfanumerik) tersedia sebagai fungsi domain agar dipakai konsisten oleh resolver & repo (hindari duplikasi aturan di adapter).
- [ ] Tidak ada import adapter/driver/library eksternal — domain murni (AGENTS.md §0/§2.4).
- [ ] Doc comment exported dimulai dengan nama identifier.
- [ ] `errors.go` bila perlu sentinel (mis. `ErrACSDeviceNotFound`, `ErrACSLinkAmbiguous`).
**Files likely touched:** `internal/domain/acsdevice/acsdevice.go`, `internal/domain/acsdevice/errors.go`.
**Dependencies:** —
**Estimated scope:** Small

---

**Task 4: Kontrak `port.ACSDeviceRepository`**
**Description:** Definisikan interface repository cache di layer port.
**Acceptance criteria:**
- [ ] File `internal/port/acs_device_repository.go` mendefinisikan interface dengan minimal: `Upsert(ctx, device)` (satu device), `UpsertBatch(ctx, devices)` opsional untuk sinkron massal, `List(ctx, filter)` (dengan filter manufacturer/product_class/serial + paginasi), `FindByID(ctx, id)`, `FindBySerialNormalized(ctx, serial)` (dipakai resolver strategi serial — kembalikan semua kandidat agar caller bisa menolak yang ambigu, jangan diam-diam ambil satu).
- [ ] Interface berada di `internal/port/`, bukan di adapter/driver (AGENTS.md §2.4).
- [ ] `context.Context` parameter pertama; error return terakhir.
- [ ] Doc comment exported dimulai dengan nama identifier.
**Files likely touched:** `internal/port/acs_device_repository.go`.
**Dependencies:** Task 3.
**Estimated scope:** Small

---

**Task 5: Implementasi Postgres `ACSDeviceRepository`**
**Description:** Implementasi konkret repository cache di adapter Postgres.
**Acceptance criteria:**
- [ ] File `internal/adapter/postgres/acs_device_repository.go` mengimplementasikan `port.ACSDeviceRepository` (dengan compile-time assertion `var _ port.ACSDeviceRepository = (*...)(nil)`).
- [ ] Upsert idempotent berbasis PK `id` (insert-or-update; hanya kolom yang disinkronkan, `synced_at` selalu diperbarui).
- [ ] Serialisasi `tags`/`params` ke jsonb; mapping model ↔ domain di `models.go`.
- [ ] `List` mendukung filter manufacturer/product_class/serial dan paginasi.
- [ ] Kesalahan "no rows" dipetakan ke sentinel domain via `errors.Is` (bukan bandingkan string).
**Files likely touched:** `internal/adapter/postgres/acs_device_repository.go`, `internal/adapter/postgres/models.go`.
**Dependencies:** Task 1, Task 3, Task 4.
**Estimated scope:** Medium

---

**Task 6: Perluas driver `genieacs` untuk daftar device + subset params**
**Description:** Tambah kemampuan mengambil daftar device GenieACS beserta parameter terpilih dari NBI, menulis tag, serta menyediakan katalog path untuk resolver.
**Acceptance criteria:**
- [ ] Client `internal/driver/genieacs/client.go` mendapat method untuk query daftar device NBI (mengembalikan id, serial, manufacturer, product class, oui, connection request url, tags, last inform) dan mengambil subset parameter yang diminta.
- [ ] Daftar parameter yang diambil didefinisikan sebagai katalog konstanta di package `genieacs` (mis. di `commands.go`) — pengetahuan "parameter TR-069 mana yang dipakai UI" hidup di driver, bukan di usecase/domain (AGENTS.md §1.2 poin 4; K1).
- [ ] **Katalog path PPPoE-username multi-path** (untuk resolver strategi (b)) hidup di driver sebagai data: satu daftar path native TR-069 yang bisa memuat PPPoE username (`InternetGatewayDevice.WANDevice.*.WANConnectionDevice.*.WANPPPConnection.*.Username` + `Device.*` + `VirtualParameters.pppoeUsername`/setara), di-query lewat `$or` — ini persis pola multi-value milik driver (K13), bukan hardcode satu path di usecase.
- [ ] Method **set-tag** (`AddTag`/`SetTags`) menulis `_tags` penuh secara **read-modify-write** (baca tag lama, tambah/ubah, tulis balik) — hindari lost-update; NBI GenieACS meng-overwrite set tag, jadi jangan tulis parsial.
- [ ] Query NBI hanya memproyeksikan field yang dibutuhkan (bukan seluruh device tree) demi hemat payload.
- [ ] Error koneksi/timeout NBI dibungkus `%w` dengan konteks, memakai `errors.go` package.
**Files likely touched:** `internal/driver/genieacs/client.go`, `internal/driver/genieacs/commands.go`, `internal/driver/genieacs/errors.go`.
**Dependencies:** Task 3.
**Estimated scope:** Medium

---

**Task 7: Usecase network `SyncGenieACSDevices`**
**Description:** Orkestrasi tarik daftar device dari NBI (via driver genieacs) lalu upsert ke cache lokal, dibatasi subset parameter.
**Acceptance criteria:**
- [ ] File `internal/usecase/network/sync_genieacs_devices.go` berisi fungsi `SyncGenieACSDevices(ctx, ...)` yang: memanggil driver genieacs untuk daftar device, memetakan ke domain `acsdevice`, lalu `Upsert`/`UpsertBatch` ke `port.ACSDeviceRepository`.
- [ ] Mendukung dua mode: sinkron seluruh daftar, dan sinkron satu `device_id` tertentu.
- [ ] Usecase TIDAK tahu detail parameter TR-069 mana yang diambil (itu di driver) — usecase hanya orkestrasi (K1/K4 boundary).
- [ ] Kegagalan sebagian (satu device gagal) tidak membatalkan seluruh sinkron; kumpulkan dan laporkan ringkasan (jumlah sukses/gagal) tanpa membuang error diam-diam.
- [ ] Table-driven test untuk mapping + penanganan kegagalan sebagian (K7).
**Files likely touched:** `internal/usecase/network/sync_genieacs_devices.go`, `internal/usecase/network/sync_genieacs_devices_test.go`.
**Dependencies:** Task 4, Task 5, Task 6.
**Estimated scope:** Medium

---

**Task 8: Daftarkan job sinkron ke scheduler (Issue 01)**
**Description:** Jadwalkan `SyncGenieACSDevices` sebagai job polling berkala memakai mekanisme scheduler Issue 01.
**Acceptance criteria:**
- [ ] Job sinkron didaftarkan ke scheduler yang sama dengan Issue 01 (mengikuti keputusan ticker vs cron dari Open Question Issue 01 — tidak membuat scheduler baru).
- [ ] Interval polling dapat dikonfigurasi lewat `internal/config/config.go` (nilai default wajar, mis. beberapa menit) — tidak hard-code.
- [ ] Kegagalan satu siklus polling dicatat (log/audit) dan tidak menghentikan siklus berikutnya.
- [ ] Polling ini dijustifikasi merujuk ADR 0005 (TR-069 memang polling) — bila perlu, tambahkan catatan singkat, bukan ADR baru.
**Files likely touched:** file registrasi scheduler dari Issue 01, `internal/config/config.go`.
**Dependencies:** Task 7, Issue 01.
**Estimated scope:** Small

---

**Task 9: Handler REST cache (list/detail/sync)**
**Description:** Sediakan endpoint baca cache dan pemicu sinkron manual.
**Acceptance criteria:**
- [ ] Handler `internal/adapter/http/acs_device_handler.go`: `GET /api/v1/acs-devices` (list+filter+paginasi), `GET /api/v1/acs-devices/:id` (detail), `POST /api/v1/acs-devices/sync` (pemicu; balas **202** + identifier pekerjaan + status pending).
- [ ] DTO request/response di `internal/adapter/http/dto/` (tidak mengekspos model Postgres langsung).
- [ ] RBAC via middleware: list/detail = staff+, sync = admin+ (K3, `configs/rbac_policy.csv` diperbarui bila perlu).
- [ ] Route didaftarkan di `internal/adapter/http/router.go`.
- [ ] Status code sesuai bagian REST API (200/202/400/404/403/502).
- [ ] Test handler pakai `httptest` (K7).
**Files likely touched:** `internal/adapter/http/acs_device_handler.go`, `internal/adapter/http/dto/acs_device.go`, `internal/adapter/http/router.go`, `configs/rbac_policy.csv`, `internal/adapter/http/acs_device_handler_test.go`.
**Dependencies:** Task 5, Task 7.
**Estimated scope:** Medium

---

**Task 10: Linking subscription ↔ acs-device (domain + repo + handler)**
**Description:** Tambah field `genieacs_device_id` ke domain/repo subscription dan endpoint untuk membaca/menyetel tautan.
**Acceptance criteria:**
- [ ] Domain `subscription.Subscription` mendapat field `GenieACSDeviceID` (pointer/nullable) tanpa melanggar boundary domain.
- [ ] Repo Postgres subscription membaca/menulis kolom `genieacs_device_id`; mapping di `models.go`.
- [ ] `GET /api/v1/subscriptions/:id/acs-device` mengembalikan device cache tertaut (staff+); membedakan 404 "subscription tak ada" vs "belum tertaut".
- [ ] `PUT /api/v1/subscriptions/:id/acs-device` (opsional) menyetel/melepas tautan (teknisi+); validasi FK ke `acs_devices` → 422 bila device id tak ada di cache.
- [ ] Handler ini boleh menempel di `subscription_handler.go` yang ada atau file handler subscription terkait; route terdaftar di `router.go`.
- [ ] Test handler + repo (httptest + testcontainers, K7).
**Files likely touched:** `internal/domain/subscription/subscription.go`, `internal/adapter/postgres/subscription_repository.go`, `internal/adapter/postgres/models.go`, `internal/adapter/http/subscription_handler.go`, `internal/adapter/http/dto/subscription.go`, `internal/adapter/http/router.go`.
**Dependencies:** Task 2, Task 5.
**Estimated scope:** Medium

---

**Task 11: Update dokumentasi skema & README**
**Description:** Pastikan DATABASE-SCHEMA.md dan README root konsisten dengan perubahan.
**Acceptance criteria:**
- [ ] DATABASE-SCHEMA.md menambah §7 (acs_devices) dan memperbarui §6.1 (subscriptions.genieacs_device_id) — sudah tercakup di Task 1/2, verifikasi finalnya di sini.
- [ ] Bila ada dokumen baru di root/docs, ditautkan dari README.md root pada PR yang sama (AGENTS.md §1.5) — issue ini kemungkinan tidak menambah dokumen root baru; pastikan tidak ada tautan menggantung.
**Files likely touched:** `DATABASE-SCHEMA.md`, `README.md`.
**Dependencies:** Task 1, Task 2.
**Estimated scope:** Small

---

## Migrasi Database

Dua migrasi berpasangan, nomor lanjut dari 000021:

**000027 — create `acs_devices`** (`migrations/000027_create_acs_devices_table.up.sql` + `.down.sql`):
- `id` text **PRIMARY KEY** — device id asli dari GenieACS (bukan UUID lokal baru).
- `serial_number` text — serial ONU/CPE; diindeks untuk filter/lookup.
- `manufacturer` text.
- `product_class` text.
- `connection_request_url` text nullable — URL connection request TR-069.
- `tags` jsonb — tag device dari GenieACS (default `'[]'`/`'{}'` sesuai bentuk NBI).
- `params` jsonb — **terbatas** ke subset parameter yang dipakai UI (bukan seluruh device tree); default objek kosong.
- `last_inform` timestamptz nullable — waktu inform terakhir menurut GenieACS.
- `synced_at` timestamptz — kapan baris ini terakhir disinkron ke cache (di-set tiap upsert).
- Index tambahan pada `manufacturer`, `product_class` untuk mendukung filter list.
- File down: drop tabel `acs_devices`.

**000028 — alter `subscriptions` tambah `genieacs_device_id`** (`migrations/000028_add_genieacs_device_id_to_subscriptions.up.sql` + `.down.sql`):
- `genieacs_device_id` text **nullable**, FK ke `acs_devices(id)` dengan `ON DELETE SET NULL` (hapus device cache tidak menghapus subscription, hanya melepas tautan).
- Index pada `genieacs_device_id` untuk lookup balik subscription→device.
- File down: hapus constraint FK + kolom.

Cerminkan kedua perubahan ke **DATABASE-SCHEMA.md** pada PR yang sama: tambah **§7** untuk `acs_devices` dan perbarui **§6.1** (subscriptions) menambahkan kolom `genieacs_device_id`.

## Verification

- [ ] `go build ./...` sukses tanpa error.
- [ ] `go test ./internal/usecase/network/...` — test `SyncGenieACSDevices` (mapping + kegagalan sebagian) hijau.
- [ ] `go test ./internal/adapter/postgres/...` — repo `acs_devices` & subscription linking via testcontainers-go hijau (K7).
- [ ] `go test ./internal/adapter/http/...` — handler list/detail/sync/link via httptest hijau, termasuk assertion status 202 pada sync dan 422 pada link ke device tak dikenal.
- [ ] `make lint` bersih (gofumpt/goimports/staticcheck — perhatikan kapitalisasi akronim `ACS`/`URL`/`ID`).
- [ ] Migrasi up lalu down berjalan bersih di DB test (tidak ada residu tabel/kolom).
- [ ] Smoke test manual (sebagai teks, bukan skrip): jalankan migrasi, panggil `curl` `POST /api/v1/acs-devices/sync` sebagai admin dan pastikan balasan 202 + identifier pekerjaan; setelah siklus sinkron, `curl` `GET /api/v1/acs-devices` sebagai staff dan pastikan device muncul dari cache; `curl` `GET /api/v1/subscriptions/:id/acs-device` untuk subscription yang sudah di-link dan pastikan device tertaut kembali; verifikasi endpoint sync ditolak 403 untuk role staff.

## Definition of Done

- [ ] Tabel `acs_devices` dan kolom `subscriptions.genieacs_device_id` ada via migrasi 000027 & 000028 (up/down teruji).
- [ ] Domain `acsdevice`, `port.ACSDeviceRepository`, dan repo Postgres-nya lengkap dengan compile-time assertion.
- [ ] Driver `genieacs` bisa menarik daftar device + subset parameter yang dipakai UI (katalog parameter di driver, bukan usecase).
- [ ] `SyncGenieACSDevices` men-upsert cache, terdaftar di scheduler Issue 01, dan bisa dipicu manual via `POST /api/v1/acs-devices/sync` (202).
- [ ] Endpoint list/detail/sync + linking subscription↔acs-device terpasang dengan RBAC benar (staff baca, admin sync, teknisi set link).
- [ ] DATABASE-SCHEMA.md §6.1 & §7 diperbarui; tidak ada dokumen root menggantung.
- [ ] `go build`, test paket terkait, dan `make lint` hijau.
