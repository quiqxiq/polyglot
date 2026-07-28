# Issue 06: ONU Authorization ke OLT

## Konteks

Issue 05 (ONU Discovery Queue) menutup separuh pertama alur FTTH: ONU yang
"unconfigured" hasil SNMP walk OLT ditampung di `onu_discovery_queue`, teknisi
memilih satu dari daftar, dan mengikatnya ke subscription (mengisi
`subscriptions.onu_pon_port`, `onu_id`, `onu_serial_number`). Tetapi mengikat ONU
ke baris subscription **tidak** membuat pelanggan bisa online — OLT belum tahu
apa-apa tentang ONU itu: belum ada `onu <id> type ... sn ...` di interface
`gpon-olt`, belum ada tcont/gemport (profil bandwidth), belum ada VLAN. Selama
otorisasi ini belum dijalankan, ONU tetap berstatus fisik-tersambung-tapi-tidak
dilayani. Issue ini menutup separuh kedua: mengirim **rangkaian command CLI** ke
OLT untuk mengotorisasi ONU, sesuai temuan §3.1 di
`ANALISIS-PROVISIONING-REPO-REFERENSI.md`.

`ANALISIS-PROVISIONING-REPO-REFERENSI.md` §3.1 mendokumentasikan urutan command
nyata dari repo referensi (`onuProvisionService.js`), dan urutan ini
**tervalidasi lapangan** (bukti: `REFERENCES.md` §D "Authorize ONU sequence +
TR-069 push, shell prompt-driven"): masuk `enable` → `configure terminal` →
`interface gpon-olt_<pon>` → `onu <onuId> type <type> sn <sn>` → keluar →
`interface gpon-onu_<pon>:<onuId>` → `name <label>` → `tcont 1 profile
<bandwidth>` → `gemport 1 tcont 1` → `pon-onu-mng gpon-onu_<pon>:<onuId>` →
`service <name> gemport 1 vlan <vlan>` → **opsional** `tr069 acs url <url>` →
`exit` → `write`. Dua poin kunci temuan itu:

- Kalau `tr069 acs url` disertakan, **OLT-lah yang mendorong konfigurasi TR-069
  (URL GenieACS) ke ONU**, sehingga ONU auto-*inform* ke GenieACS begitu online —
  tanpa perlu menyentuh ONU langsung. Push ACS URL ini adalah **bagian dari batch
  authorize yang sama** (satu sesi shell, bukan langkah/sync_log terpisah).
- **Syntax `tr069 acs url` bukan tunggal — ada 4 varian per-vendor** (`tr069 acs
  url` / `ont tr069-server url` / `ont cwmp acs-url` / `cwmp acs url`). Karena itu
  bentuk baris ACS adalah **template per-vendor** yang hidup di `commands.go`
  masing-masing driver OLT, bukan string tunggal yang diasumsikan berlaku lintas
  merk.

**Command terakhir urutan otorisasi WAJIB `write`/`save`** — tanpa persist,
konfigurasi hilang saat OLT reboot. Ini jadi bagian acceptance Task 2.

Seluruh pengetahuan urutan command ini (termasuk template ACS per-vendor) adalah
**pengetahuan vendor OLT** dan karenanya hidup di
`internal/driver/zteolt/commands.go` (dan `huaweiolt`), bukan di `usecase/`
(K1, AGENTS.md §1.2).

Provisioning OLT dilakukan lewat **shell CLI interaktif yang prompt-driven**
(bukan SNMP — SNMP di `zteolt` hanya untuk monitoring/discovery), dan transport
shell itu **bisa SSH atau Telnet, ditentukan per-OLT** — sebagian OLT lab hanya
membuka Telnet, sebagian menuntut SSH (bukti lapangan: `REFERENCES.md` §D, shell
interaktif kirim-saat-prompt). Karena itu kontrak driver OLT adalah: pilih
transport dari field `transport` (`ssh`/`telnet`) pada device, kirim tiap baris
command **setelah prompt device (`#`|`>`) muncul**, dengan **timeout per-sequence**,
dan bila OLT menuntut mode privileged, naikkan lewat `enable` memakai
`enable_password` device. Model kontrak ini ditegaskan sebagai kontrak driver OLT
di Task 2 (bukan diam-diam mengasumsikan Telnet mentah saja). Apa pun transportnya,
eksekusi tetap mengikuti pola sinkronisasi kanonik (K4): handler REST tidak
memanggil driver langsung; ia menyuruh usecase menulis baris
`provisioning_sync_log` `pending`, dan Sync Engine (Issue 01) yang menerjemahkan
baris itu jadi `command.Command` lalu mengeksekusinya lewat
`usecase/network.ExecuteCommand` (dengan jejak `command_audit_log`, K4). Karena
target sinkronisasi ini adalah OLT/ONU — bukan salah satu dari lima `target_type`
yang sudah ada — issue ini menambah nilai `target_type='olt_onu'` lewat migrasi
ALTER CHECK (000026).

## Prasyarat

- **Issue 01 (Provisioning Sync Engine)** — wajib. Baris `provisioning_sync_log`
  `pending` yang ditulis di sini hanya akan menumpuk tanpa Sync Engine yang
  memprosesnya jadi command dan mengeksekusinya ke driver OLT.
- **Issue 05 (ONU Discovery Queue)** — wajib. Otorisasi butuh `subscriptions`
  yang sudah punya `onu_pon_port`, `onu_id`, `onu_serial_number` terisi (kolom
  yang ditambahkan Issue 05). Tanpa itu, tidak ada parameter untuk menyusun
  command otorisasi.
- **Foundation Phase 4–5** — router Gin nyata, middleware `AuthRequired` +
  `RBACRequired`, `internal/adapter/http/dto/`, wiring `cmd/server/main.go`,
  serta `internal/config` (untuk sumber `acs_url` GenieACS).
- Driver OLT yang dipakai sudah ada kerangkanya: `internal/driver/zteolt`
  (`telnet.go` untuk CLI mentah, `commands.go`, `snmp.go`) dan
  `internal/driver/huaweiolt` (`driver.go`, `commands.go`).
- Tabel yang dipakai sudah ada: `subscriptions` (000009, + kolom ONU dari Issue
  05), `provisioning_sync_log` (000011), `plans` (000005), `odps` (000003).

## Ruang Lingkup

**In scope:**
- Operasi abstrak baru **authorize ONU** (mis. `command.OpAuthorizeONU`) dan
  **de-authorize ONU** (`command.OpDeauthorizeONU`) di domain `command`.
- Perluasan `internal/driver/zteolt/commands.go` (dan `huaweiolt`) dengan
  `Translate` yang memetakan operasi abstrak itu → **urutan** command native OLT
  (multi-line lewat shell prompt-driven), plus `Classify` yang menandai keduanya
  destruktif (butuh HITL). Termasuk **template baris `tr069 acs url` per-vendor**
  (4 varian syntax) di masing-masing `commands.go`.
- **Mapping plan → profil OLT** untuk vendor yang butuh id profil eksplisit:
  Huawei/Fiberhome menuntut `ont-lineprofile-id` & `ont-srvprofile-id` di urutan
  authorize (ZTE cukup nama profil tcont). Mapping ini **analog
  `plan_router_profiles`** (Issue 02) tetapi untuk OLT — hidup sebagai data yang
  bisa dipetakan dari `plan_id`+`device_id` (OLT) ke id profil vendor. Untuk merk
  yang belum dikurasi, posture fail-safe (Translate error), bukan menebak id
  profil.
- Usecase network baru `authorize_onu.go` yang membaca parameter dari
  subscription + plan + config, lalu menulis satu baris `provisioning_sync_log`
  (`target_type='olt_onu'`, `action='create'` untuk authorize / `'delete'` untuk
  de-authorize).
- Migrasi 000026: ALTER CHECK `provisioning_sync_log.target_type` menambah
  `'olt_onu'`, + cermin ke `DATABASE-SCHEMA.md` §6.3.
- Endpoint `POST .../onu/authorize` (wajib) + `DELETE .../onu` (opsional) + baris
  RBAC.
- Penanganan idempotensi & error parsial (OLT setengah terkonfigurasi) minimal:
  catat error jelas + biarkan teknisi re-run.

**Out of scope:**
- SNMP walk / pengisian `onu_discovery_queue` → Issue 05.
- Driver SNMP generik multi-vendor (`genericsnmp`) + katalog OID → Issue 07
  (melengkapi issue ini untuk merk di luar ZTE/Huawei).
- Pembacaan RX power / optical health → Issue 10.
- Konfigurasi TR-069 lanjutan dari sisi GenieACS (reboot/WAN/WiFi) → Issue 09.
  Issue ini hanya mendorong **ACS URL** lewat OLT (opsional), bukan operasi CPE.
- Penerjemahan `command.Command` → eksekusi Telnet ke device: kerangka eksekusi
  milik Issue 01 + `zteolt/telnet.go`; issue ini menyediakan katalog command-nya.

## REST API

Semua di bawah `/api/v1/`. Aksi yang menyentuh perangkat mengembalikan **202
Accepted** + id `provisioning_sync_log` (K5) — bukan 200 seolah ONU sudah pasti
terotorisasi di OLT.

| Method | Path | Tujuan | Role minimum |
|---|---|---|---|
| POST | `/api/v1/subscriptions/:id/onu/authorize` | Tulis sync_log `olt_onu`/`create` untuk mengotorisasi ONU di OLT | `teknisi` |
| DELETE | `/api/v1/subscriptions/:id/onu` | Tulis sync_log `olt_onu`/`delete` untuk men-de-authorize ONU dari OLT (opsional) | `admin` |

### POST `/api/v1/subscriptions/:id/onu/authorize`
- **Request:** `:id` = UUID subscription. Body **opsional**:
  - `vlan` (integer, opsional) — override VLAN. Jika kosong, usecase memakai VLAN
    default dari konfigurasi/plan (didokumentasikan di Task; jangan diam-diam 0).
  - `push_acs` (boolean, opsional, default `false`) — bila `true`, urutan command
    menyertakan `tr069 acs url <acs_url>` sehingga OLT mendorong URL GenieACS ke
    ONU. `acs_url` **tidak** diterima dari body (bukan input user); diambil dari
    `internal/config` (GenieACS base URL). Bila `push_acs=true` tetapi `acs_url`
    kosong di config → 400.
- **Perilaku:** validasi subscription ada; `service_type` FTTH/GPON (bukan
  pppoe-only tanpa ONU); `onu_pon_port`, `onu_id`, `onu_serial_number` semuanya
  terisi (kalau salah satu kosong → 409, arahkan ke Issue 05 bind dulu);
  `device_id` menunjuk OLT yang valid; plan punya `bandwidth_down_kbps`/`up`
  untuk memilih profil tcont. Bila lolos, tulis **satu** baris
  `provisioning_sync_log` (`target_type='olt_onu'`, `action='create'`,
  `status='pending'`, `subscription_id`, `device_id`=OLT). Parameter otorisasi
  (pon_port, onu_id, sn, bandwidth profile, vlan, push_acs) dibawa ke Sync Engine
  — lihat Task 3 soal cara Sync Engine mengambilnya kembali dari subscription +
  plan + config (bukan dari kolom bebas di sync_log).
- **Response sukses:** `202 Accepted`, body `{ "sync_log_id": "<uuid>", "status":
  "pending" }`.
- **Response gagal:** `404` subscription tak ada; `400` `push_acs=true` tapi
  `acs_url` tak terkonfigurasi, atau `vlan` di luar rentang valid; `409`
  parameter ONU belum lengkap (belum di-bind lewat Issue 05) atau `device_id`
  bukan OLT; `403` role kurang.

### DELETE `/api/v1/subscriptions/:id/onu` (opsional)
- **Request:** body kosong. `:id` = UUID subscription.
- **Perilaku:** validasi subscription ada + parameter ONU terisi; tulis **satu**
  baris `provisioning_sync_log` (`target_type='olt_onu'`, `action='delete'`,
  `status='pending'`). Sync Engine menerjemahkan jadi urutan command yang
  menghapus ONU dari interface OLT (`no onu <onuId>` pada `gpon-olt_<pon>`).
  Tidak menghapus baris subscription dan tidak mengosongkan kolom ONU (itu
  keputusan bisnis terpisah) — hanya melepas otorisasi di OLT.
- **Response sukses:** `202 Accepted` + `{ "sync_log_id": "<uuid>", "status":
  "pending" }`.
- **Response gagal:** `404` tak ada; `409` parameter ONU kosong (tak ada yang
  bisa di-de-authorize); `403` role kurang (butuh `admin`+).

## Tasks

**Task 1: Operasi abstrak authorize/deauthorize ONU di domain `command`**

**Description:** Tambah dua nilai `command.Operation` baru — otorisasi dan
de-otorisasi ONU — sebagai kosakata abstrak yang akan diterjemahkan tiap driver
OLT jadi urutan command native.

**Acceptance criteria:**
- [ ] `command.OpAuthorizeONU` dan `command.OpDeauthorizeONU` terdefinisi di
      domain `command` (mis. `internal/domain/command/command.go`), konsisten
      gaya dengan `OpGetStatus`/`OpReboot` yang sudah ada.
- [ ] Parameter otorisasi (pon_port, onu_id, serial_number, bandwidth profile,
      vlan, acs_url opsional, dan — untuk Huawei/Fiberhome — lineprofile/srvprofile
      id) dibawa lewat `command.Command.Args`
      (`map[string]string`) — tidak menambah field baru ke struct `Command`.
      Kunci Args distandarkan sebagai konstanta bernama di domain `command` (mis.
      `ArgPONPort`, `ArgONUID`, `ArgSerialNumber`, `ArgBandwidthProfile`,
      `ArgVLAN`, `ArgACSURL`, `ArgLineProfileID`, `ArgSrvProfileID`) supaya driver
      dan Sync Engine memakai kunci yang sama, bukan string literal tersebar.
- [ ] Domain `command` tetap tidak mengimpor driver/adapter apa pun (boundary).

**Files likely touched:** `internal/domain/command/command.go`.

**Dependencies:** —

**Estimated scope:** Small.

---

**Task 2: Katalog urutan command OLT di zteolt (dan huaweiolt) — Classify + Translate**

**Description:** Perluas `commands.go` tiap driver OLT agar `Translate`
mengembalikan urutan command native untuk authorize/deauthorize ONU, dan
`Classify` menandai keduanya destruktif.

**Acceptance criteria:**
- [ ] `internal/driver/zteolt/commands.go` `Translate(OpAuthorizeONU)`
      menghasilkan representasi **urutan** command native ZTE sesuai urutan
      tervalidasi lapangan (`ANALISIS-PROVISIONING-REPO-REFERENSI.md` §3.1 +
      `REFERENCES.md` §D): `enable` → `configure terminal` → `interface
      gpon-olt_<pon>` → `onu <onuId> type <type> sn <sn>` → `exit` → `interface
      gpon-onu_<pon>:<onuId>` → `name <label>` → `tcont 1 profile <bandwidth>` →
      `gemport 1 tcont 1` → `pon-onu-mng gpon-onu_<pon>:<onuId>` → `service <name>
      gemport 1 vlan <vlan>` → **opsional** baris ACS bila `ArgACSURL` ada →
      `exit` → `write`.
- [ ] **Baris ACS = template per-vendor, bukan string tunggal.** Ada 4 varian
      syntax (`tr069 acs url` / `ont tr069-server url` / `ont cwmp acs-url` /
      `cwmp acs url`); bentuk yang dipakai ditentukan di `commands.go` driver OLT
      yang bersangkutan, tidak diasumsikan sama lintas merk.
- [ ] **Command terakhir urutan WAJIB `write`/`save`.** Tanpa persist, konfigurasi
      hilang saat OLT reboot — Translate tidak boleh menghasilkan urutan yang
      berhenti sebelum baris persist.
- [ ] Nilai pon_port, onu_id, sn, bandwidth profile, vlan diambil dari
      `command.Command.Args` dengan kunci konstanta Task 1; bila ada kunci wajib
      yang kosong, `Translate` mengembalikan error jelas (mis. "zteolt: authorize
      onu butuh pon_port & onu_id").
- [ ] `Translate(OpDeauthorizeONU)` menghasilkan urutan penghapusan (`interface
      gpon-olt_<pon>` → `no onu <onuId>` → `exit` → `write`).
- [ ] `Classify(OpAuthorizeONU)` dan `Classify(OpDeauthorizeONU)` mengembalikan
      `command.ClassDestructive` (keduanya mengubah state OLT → wajib HITL).
- [ ] `huaweiolt/commands.go` menyediakan `Translate`/`Classify` setara dengan
      dialek command Huawei (bentuk berbeda, kontrak sama). **Huawei/Fiberhome
      menuntut `ont-lineprofile-id` & `ont-srvprofile-id`** di urutan authorize
      (lihat mapping plan→profil OLT, Task 3) — bukan sekadar nama profil tcont
      gaya ZTE. Bila dialek Huawei belum dikurasi penuh, posture fail-safe:
      `Translate` error dan `Classify` tetap `ClassDestructive` — tidak pernah
      diam-diam auto-approve.
- [ ] **Kontrak transport shell prompt-driven** dipatuhi: driver OLT memilih
      transport dari field device (`transport` = `ssh`/`telnet`), mengirim tiap
      baris **setelah prompt (`#`|`>`) muncul** dengan **timeout per-sequence**,
      dan menaikkan privileged via `enable`+`enable_password` bila OLT memintanya.
      Memanfaatkan jalur eksekusi shell yang sudah ada di `zteolt` (SSH maupun
      Telnet) — tidak membuat mekanisme koneksi baru.
- [ ] Pengetahuan command ini **hanya** ada di folder driver vendor; tidak ada
      string command OLT yang bocor ke `usecase/` atau `domain/` (K1).

**Files likely touched:** `internal/driver/zteolt/commands.go`,
`internal/driver/zteolt/telnet.go` dan/atau transport SSH `zteolt` (bila butuh
helper kirim-urutan prompt-driven), `internal/driver/huaweiolt/commands.go`.

**Dependencies:** Task 1.

**Estimated scope:** Large.

---

**Task 3: Usecase network `authorize_onu.go` — tulis sync_log olt_onu**

**Description:** Tambah orkestrasi (mis. `AuthorizeONU` dan `DeauthorizeONU`)
yang memvalidasi kelengkapan parameter ONU dan menulis satu baris
`provisioning_sync_log` tanpa menyentuh driver (K4).

**Acceptance criteria:**
- [ ] `AuthorizeONU` menerima `ctx`, `subscriptionID`, opsi (`vlan` opsional,
      `pushACS` bool), dan konteks aktor; mengembalikan id `sync_log` + error.
- [ ] Validasi: subscription ada; `onu_pon_port`, `onu_id`, `onu_serial_number`
      non-kosong (kalau tidak → sentinel error `ErrONUNotBound` → 409); `device_id`
      merujuk OLT; plan punya bandwidth (untuk profil tcont).
- [ ] Bila `pushACS=true`, `acs_url` dibaca dari `internal/config` (GenieACS base
      URL); bila kosong → sentinel error `ErrACSURLNotConfigured` → 400.
- [ ] Menulis tepat **satu** baris `provisioning_sync_log`
      (`target_type='olt_onu'`, `action='create'`, `status='pending'`,
      `subscription_id`, `device_id`=OLT) — bukan memanggil `port.DeviceDriver`.
- [ ] `DeauthorizeONU` menulis satu baris `olt_onu`/`delete` dengan validasi
      parameter ONU terisi.
- [ ] **Kontrak dengan Sync Engine (Issue 01):** saat memproses baris
      `target_type='olt_onu'`, Sync Engine mengambil parameter otorisasi dengan
      me-`join` `subscriptions` (pon_port, onu_id, sn, plan_id, device_id) + `plans`
      (bandwidth → nama profil tcont) + `internal/config` (acs_url bila diminta),
      lalu menyusun `command.Command` `OpAuthorizeONU`/`OpDeauthorizeONU` dengan
      `Args` sesuai konstanta Task 1, dan memanggil
      `usecase/network.ExecuteCommand` (Classify → Decide → HITL → Execute).
      Karena `provisioning_sync_log` tidak punya kolom bebas untuk vlan/push_acs,
      dokumentasikan di sini bagaimana kedua opsi itu direkonstruksi Sync Engine:
      VLAN dari default config/plan, `push_acs` — bila butuh dibedakan per
      permintaan — dititipkan lewat `external_reference` (mis. token
      `push_acs=1;vlan=100`) yang di-parse Sync Engine. Nyatakan pilihan ini
      eksplisit; jangan diam-diam mengabaikan opsi request.
- [ ] **Mapping plan → profil OLT (Huawei/Fiberhome).** ZTE cukup nama profil
      tcont dari bandwidth `plans`, tetapi Huawei/Fiberhome menuntut
      `ont-lineprofile-id` & `ont-srvprofile-id`. Sediakan pemetaan
      (`plan_id`+`device_id` OLT → id lineprofile/srvprofile) yang **analog
      `plan_router_profiles`** (Issue 02) tetapi untuk OLT, dan sertakan id itu ke
      `Args` (mis. `ArgLineProfileID`, `ArgSrvProfileID`) saat vendor OLT-nya
      Huawei/Fiberhome. Bila mapping tak ada untuk pasangan plan+OLT, gagal
      eksplisit (jangan menebak id profil) — konsisten posture fail-safe driver.
      Nyatakan apakah mapping ini kolom di `plans`, tabel pivot baru, atau config;
      bila tabel/kolom baru dibutuhkan, ambil nomor migrasi bebas berikutnya
      **setelah 000036** dan tambahkan barisnya ke tabel reservasi `README.md` §K6
      pada PR yang sama (jangan reuse 000026 yang hanya untuk ALTER CHECK).
- [ ] Sentinel error terdefinisi di domain terkait (mis.
      `internal/domain/subscription/errors.go` atau `internal/domain/command/errors.go`):
      `ErrONUNotBound`, `ErrACSURLNotConfigured`, `ErrDeviceNotOLT`.

**Files likely touched:** `internal/usecase/network/authorize_onu.go`,
`internal/domain/subscription/errors.go`, `internal/port/subscription_repository.go`
(bila butuh query join/insert sync_log), `internal/config/config.go` (baca
`acs_url` bila belum ada).

**Dependencies:** Task 1, Task 2, Issue 01, Issue 05.

**Estimated scope:** Large.

---

**Task 4: Handler REST + DTO + routing + RBAC**

**Description:** Tambah handler `POST .../onu/authorize` dan `DELETE .../onu`,
DTO request/response, daftarkan route, tambah baris Casbin.

**Acceptance criteria:**
- [ ] Handler di `internal/adapter/http/subscription_handler.go` memanggil
      usecase network `AuthorizeONU`/`DeauthorizeONU`, tidak menyentuh driver (K4).
- [ ] DTO request/response di `internal/adapter/http/dto/` (mis.
      `AuthorizeONURequest{ vlan *int, push_acs bool }`, `AuthorizeONUResponse{
      sync_log_id, status }`).
- [ ] Route terdaftar di `internal/adapter/http/router.go` dalam group
      ber-middleware `AuthRequired` + `RBACRequired`.
- [ ] Pemetaan error → status sesuai bagian REST API: 404/400/409/403/202;
      format error `{ "error": { "code", "message" } }`.
- [ ] Baris RBAC ditambah ke `configs/rbac_policy.csv`: `teknisi` boleh
      `authorize`; `admin`+ boleh `authorize` dan `de-authorize` (DELETE);
      `superadmin`/`owner` semua; `staff` tidak diberi akses kedua aksi ini.

**Files likely touched:** `internal/adapter/http/subscription_handler.go`,
`internal/adapter/http/dto/subscription.go`, `internal/adapter/http/router.go`,
`configs/rbac_policy.csv`.

**Dependencies:** Task 3.

**Estimated scope:** Medium.

---

**Task 5: Idempotensi & penanganan error parsial (OLT setengah terkonfigurasi)**

**Description:** Definisikan perilaku saat sebagian command dalam urutan
otorisasi gagal (mis. `onu ... sn ...` sukses tapi `tcont` gagal), sehingga OLT
bisa berada dalam keadaan setengah-terkonfigurasi.

**Acceptance criteria:**
- [ ] Bila salah satu command dalam urutan gagal, `sync_log` baris itu jadi
      `failed` dengan `error_message` yang menyebut **command mana** yang gagal
      (bukan pesan generik) — cukup untuk teknisi tahu di titik mana berhenti.
- [ ] Otorisasi ulang (re-run endpoint `authorize` pada subscription yang sama)
      tidak menduplikasi ONU secara berbahaya: didokumentasikan bahwa OLT ZTE
      menolak/menimpa `onu <id>` yang sudah ada, dan bila OLT mengembalikan pesan
      "already exists" itu **tidak** dianggap kegagalan fatal (diperlakukan sesuai
      `genericcli.Catalog.FailedWhenContains` yang relevan — pesan idempotent
      tidak masuk daftar gagal).
- [ ] Tidak ada rollback otomatis yang mencoba "membersihkan" konfigurasi
      sebagian secara diam-diam (bisa memperburuk keadaan). Kebijakan eksplisit:
      catat error jelas, biarkan baris `failed`, dan sediakan jalur re-run
      (endpoint authorize lagi, atau retry sync_log lewat endpoint retry Issue 01).
- [ ] Perilaku ini didokumentasikan singkat sebagai komentar `// DEVIASI`/catatan
      bila menyimpang dari pola sync lain, atau ditulis sebagai bagian di file
      issue/ADR bila penjelasannya panjang (>5 baris komentar → rujuk dokumen).

**Files likely touched:** `internal/usecase/network/authorize_onu.go` (dokumentasi
kontrak), `internal/driver/zteolt/commands.go`/`telnet.go` (klasifikasi pesan
idempotent), catatan di file issue ini.

**Dependencies:** Task 2, Task 3.

**Estimated scope:** Medium.

---

**Task 6: Test — usecase (table-driven), driver Translate/Classify, handler (httptest)**

**Description:** Uji validasi parameter ONU, penulisan sync_log, urutan command
yang dihasilkan driver, dan kontrak HTTP.

**Acceptance criteria:**
- [ ] Test usecase table-driven: authorize sukses (parameter ONU lengkap),
      parameter ONU tak lengkap (409 `ErrONUNotBound`), `push_acs=true` tanpa
      `acs_url` config (400 `ErrACSURLNotConfigured`), `device_id` bukan OLT (409),
      subscription tak ada (404). Verifikasi tepat satu baris `provisioning_sync_log`
      `olt_onu`/`create` tertulis dan **tidak** ada pemanggilan driver.
- [ ] Test driver `zteolt` (table-driven, tanpa perangkat): `Translate(OpAuthorizeONU)`
      dengan Args lengkap menghasilkan urutan command yang benar dan berurutan;
      dengan Args wajib kosong menghasilkan error; `push_acs` menamb/menghapus baris
      `tr069 acs url`; `Classify` = `ClassDestructive` untuk kedua operasi.
- [ ] Test repo Postgres pakai `testcontainers-go` (bukan mock): CHECK baru
      menerima `target_type='olt_onu'`, dan baris tersimpan atomik.
- [ ] Test handler `httptest`: status HTTP + bentuk body (`sync_log_id`) sesuai
      kontrak; RBAC menolak `staff`; DELETE menolak `teknisi` (butuh `admin`+).
- [ ] Test integrasi ke OLT nyata (ZTE C320/CHR-equivalent di lab, bila tersedia)
      ditaruh di `test/integration/` — bukan mock `port.DeviceDriver` untuk
      membuktikan end-to-end (K7). Boleh di-skip di CI tanpa lab.

**Files likely touched:** `internal/usecase/network/authorize_onu_test.go`,
`internal/driver/zteolt/commands_test.go`,
`internal/adapter/postgres/subscription_repository_test.go`,
`internal/adapter/http/subscription_handler_test.go`,
`test/integration/zteolt_authorize_test.go`.

**Dependencies:** Task 2–5.

**Estimated scope:** Large.

## Migrasi Database

Ada satu perubahan skema (bukan tabel baru — hanya perluasan CHECK constraint):

- **Nomor migrasi: 000026** (direservasi untuk issue ini di tabel reservasi
  `docs/plan-provisioning/README.md` §K6 — periksa folder `migrations/` sebelum
  menetapkan nomor, jangan reuse).
- **File:** `migrations/000026_add_olt_onu_target_type.up.sql` +
  `migrations/000026_add_olt_onu_target_type.down.sql` (berpasangan, K6).
- **Isi `up`:** ubah CHECK constraint pada kolom `provisioning_sync_log.target_type`
  agar menerima nilai `'olt_onu'` di samping lima nilai lama
  (`mikrotik_ppp_secret`, `mikrotik_hotspot_user`, `mikrotik_address_list`,
  `freeradius`, `genieacs_tr069`). Karena Postgres tidak bisa "menambah nilai" ke
  CHECK secara langsung, langkahnya: DROP constraint lama (sebut nama constraint
  eksplisit sesuai yang dibuat migrasi 000011), lalu ADD constraint baru dengan
  daftar enam nilai. Tidak ada perubahan data baris lama.
- **Isi `down`:** kebalikannya — DROP constraint enam-nilai, ADD ulang constraint
  lima-nilai lama. Catatan di `down`: bila sudah ada baris `target_type='olt_onu'`
  saat rollback, ADD constraint akan gagal; `down` mengasumsikan tidak ada baris
  `olt_onu` (atau operator menghapusnya dulu) — dokumentasikan asumsi ini di
  komentar migrasi.
- **Cermin ke `DATABASE-SCHEMA.md` §6.3 pada PR yang sama:** perbarui daftar nilai
  `target_type` menjadi enam, tambahkan satu paragraf yang menjelaskan
  `target_type='olt_onu'` menunjuk ke otorisasi ONU di OLT lewat urutan command
  CLI (rujuk §3.1 `ANALISIS-PROVISIONING-REPO-REFERENSI.md`), analog dengan
  penjelasan `genieacs_tr069` yang sudah ada.

## Verification

- [ ] `go build ./...` sukses.
- [ ] `go test ./internal/usecase/network/...` hijau (table-driven authorize/deauthorize).
- [ ] `go test ./internal/driver/zteolt/...` hijau (Translate menghasilkan urutan
      command benar; Classify destruktif; error saat Args kurang).
- [ ] `go test ./internal/adapter/http/...` hijau (httptest handler + RBAC).
- [ ] `go test ./internal/adapter/postgres/...` hijau (testcontainers menerima
      `target_type='olt_onu'`).
- [ ] `make lint` bersih — perhatikan boundary: `usecase/network` tidak mengimpor
      `driver/zteolt` (pengetahuan command hanya di driver, dipanggil Sync Engine
      lewat registry + ExecuteCommand).
- [ ] Migrasi naik-turun bersih: jalankan `migrate up` lalu `migrate down` satu
      langkah pada DB test, pastikan constraint kembali ke lima nilai.
- [ ] Smoke test manual (butuh Issue 01 + Issue 05 aktif + OLT lab): bind ONU dari
      discovery queue ke subscription (Issue 05) → `curl -X POST
      .../subscriptions/:id/onu/authorize` body `{"push_acs":true}` dengan JWT role
      `teknisi`, harap **202** + `sync_log_id` → poll `curl .../sync-logs/:id`
      (Issue 01) sampai `success`/`failed` → verifikasi di OLT bahwa `show gpon onu
      state gpon-olt_<pon>` memunculkan ONU sebagai `working` dan (bila push_acs)
      ONU muncul di GenieACS.
- [ ] Smoke test negatif: `curl` authorize pada subscription yang parameter ONU-nya
      kosong → **409**; `curl` authorize dengan JWT role `staff` → **403**; `curl -X
      DELETE .../onu` dengan role `teknisi` → **403**.

## Definition of Done

- [ ] Operasi abstrak `OpAuthorizeONU`/`OpDeauthorizeONU` ada di domain `command`;
      pengetahuan urutan command OLT **hanya** di `internal/driver/zteolt`
      (`huaweiolt`), tidak bocor ke usecase/domain (K1).
- [ ] `AuthorizeONU`/`DeauthorizeONU` menulis tepat satu baris
      `provisioning_sync_log` `olt_onu` dan **tidak** memanggil driver langsung (K4).
- [ ] `push_acs` mendorong baris ACS (dari config) ke urutan command **dalam batch
      authorize yang sama**; syntax baris ACS = template per-vendor (4 varian) di
      `commands.go`, bukan string tunggal; `acs_url` tidak pernah datang dari body
      request.
- [ ] Command terakhir urutan authorize adalah `write`/`save` (persist), dan urutan
      ZTE menyertakan `name` + `pon-onu-mng`/`service … gemport … vlan` sesuai
      urutan tervalidasi lapangan (`REFERENCES.md` §D).
- [ ] Kontrak driver OLT prompt-driven dipatuhi: transport per-OLT (`ssh`/`telnet`),
      kirim-saat-prompt dengan timeout per-sequence, `enable`+`enable_password`
      bila perlu; semua eksekusi tetap lewat sync_log → `command_audit_log` (K4).
- [ ] Huawei/Fiberhome memakai `ont-lineprofile-id`/`ont-srvprofile-id` hasil
      mapping plan→profil OLT (analog `plan_router_profiles`); tak ada mapping →
      gagal eksplisit, bukan menebak.
- [ ] Migrasi 000026 menambah `target_type='olt_onu'` naik-turun bersih dan
      dicerminkan di `DATABASE-SCHEMA.md` §6.3 pada PR yang sama.
- [ ] Endpoint mengikuti konvensi (202 + `sync_log_id`) dan RBAC diperbarui di
      `configs/rbac_policy.csv` (`teknisi` authorize; `admin`+ de-authorize).
- [ ] Error parsial OLT ditangani dengan pesan jelas + jalur re-run, tanpa rollback
      otomatis yang berisiko.
- [ ] Semua item Verification hijau; satu issue = satu PR.
