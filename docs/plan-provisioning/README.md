# Implementation Plan: Provisioning REST API

## Overview

Dokumen ini adalah rencana implementasi **lapisan provisioning & sinkronisasi
perangkat** Polyglot, dipecah menjadi issue-issue berurutan. Fokusnya adalah
REST API yang menghubungkan data bisnis (customer, subscription, plan) dengan
perangkat nyata di lapangan: **MikroTik** (PPPoE/Hotspot), **OLT/ONU** (FTTH),
dan **GenieACS/TR-069** (CPE).

Rencana ini **melanjutkan** `docs/plan-foundation-first.md`. Foundation membuat
backend production-ready (Postgres, GORM, vault, JWT/Casbin, Gin) dan CRUD dasar.
Rencana provisioning ini mengasumsikan foundation **sudah selesai sampai Phase 5**
(router Gin nyata, middleware auth/RBAC, CRUD device/customer/subscription/plan
sudah jalan). Lihat §Prasyarat di bawah.

Sumber analisis rencana ini:
- `REFERENCES.md` (folder ini) — analisis source langsung 5 repo produksi
  (billing-rtrw, gembok-bill, gembok-simple, mikhmonv3, mikhmon-agent); basis
  bukti untuk konvensi field-tested K9–K15.
- `ANALISIS-PROVISIONING-REPO-REFERENSI.md` — 9 temuan actionable dari 4 repo
  billing RT/RW-Net produksi (gembok-simple, billing-rtrw, gembok-bill,
  mikhmon-agent).
- `DATABASE-SCHEMA.md` — 27 tabel, khususnya §6 (subscription & sinkronisasi)
  dan §7 (mekanisme sinkron ke MikroTik).
- `docs/adr/0003-mikrotik-dual-connection-streaming.md`,
  `docs/adr/0004-generic-cli-driver-scrapligo.md`,
  `docs/adr/0005-genieacs-polling.md`.

## Cara Memakai Dokumen Ini (Flow: Issue → Planning → Implementasi)

Setiap issue di rencana ini dirancang untuk mengikuti alur kerja:

1. **Submit issue** — buat GitHub issue baru dengan judul dan isi yang diambil
   dari file issue terkait di folder ini. Body issue = ringkasan "Konteks",
   "Tasks", dan "REST API" dari file tersebut.
2. **Planning** — sebelum menulis kode, baca ulang file issue lengkap +
   dokumen yang dirujuknya (AGENTS.md untuk struktur, DATABASE-SCHEMA.md untuk
   kolom, ADR terkait). Konfirmasi asumsi penempatan file lewat AGENTS.md §1.
3. **Implementasi** — kerjakan Tasks berurutan, penuhi Acceptance Criteria,
   jalankan Verification. Satu issue = satu PR.

**Aturan wajib untuk implementer:** baca `AGENTS.md` di root sebelum membuat file
apa pun. Rencana ini menyebut path file yang *disarankan*, tetapi AGENTS.md §1
adalah otoritas final soal struktur folder. Kalau ada konflik, AGENTS.md menang.

## Prasyarat (harus selesai sebelum issue mana pun di sini)

Dari `docs/plan-foundation-first.md`:
- Phase 4 (REST API): `internal/adapter/http/router.go` sudah membangun
  `*gin.Engine` nyata dengan `/health`, `/api/v1/login`, group `/api/v1`
  ber-middleware `AuthRequired` + `RBACRequired`, dan `internal/adapter/http/dto/`
  untuk request/response.
- Phase 5 (Wire): `cmd/server/main.go` sudah menyalakan REST server + MCP server
  + Postgres + vault + registry.

Kalau prasyarat ini belum ada, selesaikan foundation dulu — jangan mulai issue
provisioning di atas router yang masih stub.

## Peta Issue & Urutan

Issue dikelompokkan dalam 5 fase. Panah = ketergantungan (harus selesai dulu).

### Fase A — Inti Sinkronisasi
- **[Issue 01: Provisioning Sync Engine](issue-01-sync-engine.md)** — jantung
  semua provisioning: domain + repo + usecase yang membaca baris
  `provisioning_sync_log` `pending`, menerjemahkannya jadi `command.Command`,
  memanggil `usecase/network.ExecuteCommand`, dan menautkan hasilnya ke
  `command_audit_log`. Termasuk circuit-breaker singkat di `dialAndLogin`
  (temuan #7). **Tidak ada issue lain yang bisa jalan tanpa ini.**

### Fase B — Provisioning MikroTik (PPPoE/Hotspot)
- **[Issue 02: Plan Router Profile Sync](issue-02-plan-router-profile.md)** —
  sinkron paket ↔ `/ppp profile` MikroTik lewat `plan_router_profiles`.
  *(butuh 01)*
- **[Issue 03: Subscription Provisioning Lifecycle](issue-03-subscription-provisioning.md)**
  — create/activate subscription → provision `/ppp secret` PPPoE. *(butuh 01, 02)*
- **[Issue 04: Suspend / Resume / Terminate Cascade](issue-04-suspend-cascade.md)**
  — satu event bisnis → banyak baris sync (MikroTik + GenieACS), temuan D.
  *(butuh 03; integrasi GenieACS penuh butuh 08)*
- **[Issue 13: Hotspot & Voucher Lifecycle](issue-13-hotspot-voucher.md)** —
  provisioning hotspot (`/ip hotspot user`, bukan static IP), bulk generate
  voucher, validity per plan, expiry & pemutusan (disable/remove + kill sesi).
  Enforcement hybrid: scheduler Golang teraudit + script router `on-login`
  gaya Mikhmon v3 sebagai jaring pengaman offline. *(butuh 01, 02; capture
  login butuh 12; cascade pelanggan hotspot dipakai 04)*

### Fase C — OLT & ONU (FTTH)
- **[Issue 05: ONU Discovery Queue](issue-05-onu-discovery.md)** — tabel baru
  `onu_discovery_queue` untuk ONU "unconfigured" hasil SNMP walk OLT; kolom
  `onu_pon_port` + `onu_id` di `subscriptions` (temuan A/B). *(butuh 01)*
- **[Issue 06: ONU Authorization ke OLT](issue-06-onu-authorization.md)** —
  kirim rangkaian command CLI ke OLT untuk mengotorisasi ONU + push TR-069 ACS
  URL. *(butuh 05)*
- **[Issue 07: Driver OLT SNMP Multi-Vendor (`genericsnmp`)](issue-07-genericsnmp-olt.md)**
  — driver SNMP generik + katalog OID per merk sebagai data (temuan #6, pola
  ADR 0004). Butuh ADR baru. *(butuh 05; melengkapi 06)*

### Fase D — GenieACS / TR-069
- **[Issue 08: GenieACS Device Cache](issue-08-genieacs-cache.md)** — tabel baru
  `acs_devices` (mirror lokal GenieACS) + usecase polling sinkron; kolom
  `genieacs_device_id` di `subscriptions` (temuan C). *(butuh 01)*
- **[Issue 09: GenieACS CPE Operations](issue-09-genieacs-cpe-ops.md)** — reboot,
  enable/disable WAN, konfigurasi WiFi via TR-069 lewat `provisioning_sync_log`.
  *(butuh 08; melengkapi 04)*
- **[Issue 10: RX Power / ONU Optical Health Monitoring](issue-10-rxpower-monitoring.md)**
  — pembacaan RX power/jarak fiber ke **InfluxDB** (time-series), bukan Postgres
  (temuan #5). *(butuh 08; data OLT bisa dari 07)*

### Fase E — Topologi & Observabilitas
- **[Issue 11: ODP Topology Graph](issue-11-odp-topology.md)** — tabel `odp_links`
  untuk splitter bertingkat (temuan E, **opsional** tergantung topologi riil).
  *(butuh foundation)*
- **[Issue 12: Subscriber Session Tracking](issue-12-subscriber-sessions.md)** —
  isi `subscriber_sessions` dari event stream `/log follow` PPPoE (bukan
  polling), status online pelanggan. *(butuh 01; pakai `port.StreamingDeviceDriver`)*

### Fase F — Otomasi Billing-Driven
- **[Issue 14: Auto-Suspend & Auto-Restore Scheduler](issue-14-auto-suspend-scheduler.md)**
  — sweep terjadwal yang meng-isolir pelanggan telat bayar & memulihkan yang
  lunas, dua model trigger (grace-period vs tanggal-isolir per-pelanggan),
  memanggil usecase Issue 04 dengan `actor_type='system_scheduled'`. Tidak
  menyentuh perangkat langsung. *(butuh 04; pakai repo `invoices` foundation)*

```
                 ┌─────────────┐
                 │  Foundation │  (plan-foundation-first.md, Phase 1–5)
                 └──────┬──────┘
                        │
                 ┌──────▼──────┐
                 │  Issue 01   │  Sync Engine (inti)
                 └──┬───┬───┬──┘
        ┌───────────┘   │   └───────────┬───────────┐
   ┌────▼────┐     ┌────▼────┐     ┌────▼────┐  ┌────▼────┐
   │Issue 02 │     │Issue 05 │     │Issue 08 │  │Issue 12 │
   │plan prof│     │ONU disc │     │ACS cache│  │sessions │
   └────┬────┘     └────┬────┘     └──┬───┬──┘  └─────────┘
   ┌────▼────┐     ┌────▼────┐   ┌────▼┐ ┌▼──────┐
   │Issue 03 │     │Issue 06 │   │Iss09│ │Issue10│
   │sub prov │     │ONU auth │   │CPEop│ │rxpower│
   └────┬────┘     └────┬────┘   └─────┘ └───────┘
   ┌────▼────┐     ┌────▼────┐
   │Issue 04 │◄────│Issue 07 │  (07 melengkapi 06; 04 butuh 09 utk cascade ACS)
   │ cascade │     │genericsnmp│
   └─────────┘     └─────────┘

   Issue 11 (odp_links) — opsional, hanya butuh foundation.
   Issue 13 (hotspot/voucher) — Fase B, butuh 01+02; capture login butuh 12.
   Issue 14 (auto-suspend scheduler) — Fase F, butuh 04; pakai repo invoices.
```

## Konvensi Bersama (berlaku untuk SEMUA issue)

Konvensi ini tidak diulang di tiap file issue — implementer wajib menerapkannya
di setiap issue.

### K1. Struktur & penempatan file
- Ikuti `AGENTS.md` §1 tanpa kecuali. Domain di `internal/domain/<domain>/`,
  kontrak di `internal/port/`, orkestrasi di `internal/usecase/{network,business}/`,
  handler REST di `internal/adapter/http/`, implementasi vendor di
  `internal/driver/<vendor>/`.
- Provisioning ke perangkat = orkestrasi jaringan → `internal/usecase/network/`.
  CRUD entitas bisnis murni → `internal/usecase/business/`.
- Pengetahuan protokol/command vendor **selalu** di `internal/driver/<vendor>/`,
  tidak pernah bocor ke `usecase/` atau `domain/` (AGENTS.md §1.2).

### K2. Base path & versioning REST
- Semua endpoint di bawah `/api/v1/`.
- Resource jamak, kebab/snake konsisten dengan yang sudah ada di foundation
  (`/api/v1/devices`, `/api/v1/subscriptions`, `/api/v1/plans`).
- Sub-resource aksi provisioning pakai sub-path verb eksplisit, mis.
  `POST /api/v1/subscriptions/:id/suspend` — bukan mengubah status lewat
  `PUT` generik, supaya aksi yang memicu perangkat selalu eksplisit dan
  bisa diberi izin RBAC terpisah.

### K3. Autentikasi & RBAC
- Semua endpoint provisioning butuh JWT valid (`AuthRequired`) kecuali
  disebut publik.
- Aksi yang menyentuh perangkat (`/suspend`, `/provision`, `/authorize`,
  `/reboot`, dst) dibatasi role via Casbin. Default matrix:
  `superadmin`/`owner` = semua; `admin` = semua provisioning; `teknisi` =
  operasi lapangan (ONU discovery/authorize, reboot CPE, baca status);
  `staff` = hanya baca. Setiap issue menyebut role minimum per endpoint;
  tambahkan barisnya ke `configs/rbac_policy.csv`.

### K4. Pola sinkronisasi (WAJIB — jangan buat jalur kedua)
Setiap aksi yang mengubah perangkat **harus** lewat pola di `DATABASE-SCHEMA.md`
§7.1, bukan memanggil driver langsung dari handler:
1. Handler REST memvalidasi input dan memanggil usecase bisnis/network.
2. Usecase menulis satu/lebih baris `provisioning_sync_log` `status='pending'`
   (satu event bisa multi-target — lihat Issue 04) dalam transaksi bisnis.
3. Sync Engine (Issue 01) memproses baris `pending` → `command.Command` →
   `usecase/network.ExecuteCommand` (yang menjalankan `Classify` → `Decide` →
   HITL bila destruktif → `Execute`).
4. Command tereksekusi tercatat di `command_audit_log`; `sync_log` diupdate
   `success`/`failed` dan `command_audit_log_id` diisi.
- **Tidak ada handler REST yang memanggil `port.DeviceDriver` langsung.** Handler
  hanya menyentuh usecase; driver hanya disentuh Sync Engine lewat ExecuteCommand.

### K5. Error & respons
- Format error JSON konsisten dengan foundation (`internal/adapter/http/dto/`):
  `{ "error": { "code": "...", "message": "..." } }`. Jangan bocorkan error
  internal mentah ke klien.
- Domain error dipetakan ke HTTP status di handler: `ErrNotFound`→404,
  validasi→400, RBAC ditolak→403, konflik status→409, kegagalan perangkat
  yang tertunda→202 (accepted, sync pending) bukan 500.
- Aksi provisioning yang asinkron (menulis `sync_log` `pending`, hasil perangkat
  belakangan) mengembalikan **202 Accepted** + id `sync_log` untuk dipolling,
  bukan 200 seolah sudah tereksekusi di perangkat.

### K6. Migrasi
- Nomor migrasi lanjut dari yang terakhir ada (`000021`). Issue yang butuh tabel
  baru memakai nomor berikutnya, berpasangan `.up.sql` + `.down.sql`, satu
  perubahan skema per pasang (AGENTS.md §1.4).
- Setiap tabel/kolom baru **wajib** juga dicerminkan di `DATABASE-SCHEMA.md`
  pada PR yang sama, supaya dokumen desain tidak basi.
- **Nomor migrasi & ADR direservasi per issue** di tabel di bawah — jangan
  memilih nomor sendiri; pakai yang sudah dialokasikan supaya tidak bentrok
  antar-issue yang dikerjakan paralel.

#### Reservasi Nomor Migrasi (lanjutan dari 000021)

| Migrasi | Issue | Perubahan skema |
|---|---|---|
| `000022_add_index_provisioning_sync_log_status` | 01 | Index pada `provisioning_sync_log(status)` (opsional) |
| `000023_add_ppp_profile_target_type` | 02 | ALTER CHECK `target_type` + `mikrotik_ppp_profile` |
| `000024_create_onu_discovery_queue` | 05 | Tabel `onu_discovery_queue` |
| `000025_add_onu_pon_port_and_onu_id_to_subscriptions` | 05 | Kolom `onu_pon_port`, `onu_id` |
| `000026_add_olt_onu_target_type` | 06 | ALTER CHECK `target_type` + `olt_onu` |
| `000027_create_acs_devices_table` | 08 | Tabel `acs_devices` |
| `000028_add_genieacs_device_id_to_subscriptions` | 08 | Kolom `genieacs_device_id` |
| `000029_add_payload_to_provisioning_sync_log` | 09 | Kolom payload TR-069 di `provisioning_sync_log` |
| `000030_create_optical_thresholds_table` | 10 | Tabel `optical_thresholds` (non time-series) |
| `000031_create_odp_links_table` | 11 | Tabel `odp_links` (opsional) |
| `000032_add_hotspot_config_to_plans` | 13 | Kolom validity/hotspot per plan di `plans` |
| `000033_add_hotspot_target_types` | 13 | ALTER CHECK `target_type` + `mikrotik_hotspot_profile` & `mikrotik_ip_binding` |
| `000034_add_hotspot_access_mode_to_subscriptions` | 13 | Kolom `hotspot_access_mode` (`mac_login`/`ip_binding`) di `subscriptions` |
| `000035_create_resolved_oids_table` | 07 | Tabel cache OID hasil deteksi per-OLT (self-healing OID) |
| `000036_add_static_suspension_method_to_subscriptions` | 04 | Kolom metode isolir static-IP (opsional, hanya bila disimpan di kolom) |
| `000037_add_auto_suspend_config_to_subscriptions` | 14 | Kolom `auto_suspend_enabled`, `isolir_day`, `grace_period_days` di `subscriptions` |

Issue 03, 12 **tidak** membuat migrasi baru (memakai tabel yang sudah ada).
Migrasi 000035 (Issue 07) dan 000036 (Issue 04) muncul dari koreksi field-tested
(lihat `REFERENCES.md`); 000036 kondisional (hanya bila metode isolir static-IP
disimpan sebagai kolom). Kalau sebuah issue ternyata butuh migrasi tak terduga
lain, ambil nomor bebas berikutnya **setelah** 000037 dan tambahkan barisnya ke
tabel ini pada PR yang sama.

#### Reservasi Nomor ADR (lanjutan dari 0005)

| ADR | Issue | Topik |
|---|---|---|
| `0006-provisioning-sync-engine` | 01 | Pola sync-log → ExecuteCommand → audit |
| `0007-onu-discovery-queue` | 05 | Discovery ONU unconfigured via SNMP |
| `0008-genericsnmp-olt-katalog-oid` | 07 | Satu driver SNMP + katalog OID sebagai data |
| `0009-influxdb-untuk-metrik-optik-timeseries` | 10 | RX power ke InfluxDB, bukan Postgres |
| `0010-topologi-odp-bertingkat` | 11 | `odp_links` graf (opsional) |
| `0011-subscriber-session-via-event-stream` | 12 | Sesi pelanggan dari `/log follow`, bukan polling |
| `0012-hotspot-voucher-lifecycle-hybrid` | 13 | Validity per plan + enforcement hybrid Golang/script router |

Setiap ADR baru wajib ditautkan dari `README.md` root pada PR yang sama
(AGENTS.md §1.5).

### K7. Testing
- Table-driven test untuk usecase (AGENTS.md §9). Repo Postgres diuji dengan
  `testcontainers-go` (bukan mock). Handler diuji dengan `httptest`.
- Driver ke perangkat nyata (MikroTik CHR di GNS3, OLT/GenieACS lab) untuk
  test integrasi di `test/integration/` — jangan mock `port.DeviceDriver`
  untuk membuktikan alur end-to-end.

### K8. InfluxDB (khusus data time-series)
- RX power, traffic, dan metrik time-series **tidak** masuk Postgres (Issue 10).
  Pakai sink InfluxDB. Kalau belum ada port/adapter InfluxDB di repo, Issue 10
  mendefinisikannya sebagai `port.MetricSink` + adapter — bukan tabel Postgres.

### K9. Kill sesi aktif WAJIB menyertai perubahan yang memutus (field-tested)

Terverifikasi dari 5 repo referensi (`refrensi/`, lihat `REFERENCES.md`): di
RouterOS, mengubah `/ppp secret` (set profile / `disabled=yes`) **tidak** memutus
sesi yang sedang online — perubahan baru berlaku saat dial berikutnya. Demikian
juga `/ip hotspot user set disabled=yes` / `remove` tidak menendang sesi aktif.

Karena itu, **setiap** aksi yang bermaksud memutus atau mengubah kelas layanan
seketika HARUS diikuti kill sesi aktif sebagai langkah kedua yang **juga
teraudit**:
- PPPoE: `mikrotik_ppp_secret` `disable`/`change_profile` (dan `enable`/resume
  yang mengubah profil) → diikuti `/ppp active remove [find name=<user>]`.
- Hotspot: `mikrotik_hotspot_user` `disable`/`delete` → kumpulkan **SEMUA** `.id`
  dari `/ip hotspot active print ?user=<code>` (shared-users bisa >1) dan remove
  tiap sesi.
- Hotspot mode A (mac-cookie aktif, K13): tambahkan `/ip hotspot cookie remove`
  untuk MAC/user tsb, kalau tidak MAC ber-cookie auto-relogin tanpa kredensial.

Ini kontrak driver MikroTik (di `commands.go`, bukan usecase): satu operasi
abstrak "putus/ubah" menerjemah jadi sekuens command [set/disable] → [active
remove …] (→ [cookie remove]). Test membuktikan langkah kill hadir. Referensi
lapangan justru sering lupa langkah ini (bug diam-diam) — kita perbaiki, bukan
menirunya.

### K10. Idempotensi aksi perangkat (cek-sebelum-tulis)

Semua operasi provisioning/isolir di referensi idempoten: cek keberadaan dulu
(address-list/queue/firewall/profil) dan hanya set+kick **bila berubah**. Karena
Sync Engine (Issue 01) bisa me-retry baris `failed`, kontrak driver setiap aksi
harus idempoten: cek sebelum add/remove, set profil hanya bila berbeda, jangan
double-kick. Jadikan bagian acceptance kontrak driver di Issue 01.

### K11. Isolir = change_profile ke profil ISOLIR + infra prasyarat (bukan sekadar disable)

Pola isolir teruji: pindahkan `/ppp secret` ke **profil "ISOLIR"** (bukan cuma
`disabled=yes`) karena profil isolir bisa membawa redirect portal + rate-limit
0/0 + hook address-list. Konsekuensi yang wajib dimodelkan:
- **Infra one-time per router** (bukan per-pelanggan): firewall rule
  `src-address-list=LIST_ISOLIR action=drop` + NAT redirect portal + allow
  DNS/portal. Tanpa ini, menambah IP ke address-list tidak memblokir apa pun.
  Dipasang idempoten (by comment) saat registrasi device / setup (Issue 02),
  atau task "ensure isolir firewall" sebelum suspend pertama.
- **Profil isolir per-device sebagai entitas** (nama + apakah pakai hook
  on-up/on-down yang memasukkan IP dinamis PPPoE ke address-list saat login) —
  bukan satu string global. Untuk static IP, address-list ditulis langsung; untuk
  PPPoE IP-dinamis, keanggotaan address-list dipasang oleh hook profil saat login
  (IP sesi berikutnya belum tentu sama).
- **Ensure profil isolir ada sebelum suspend** (fail-safe): kalau profil isolir
  tak ada di device, suspend GAGAL eksplisit — jangan diam-diam pakai profil lain.
- **Nama profil normal saat resume** berasal dari plan/subscription (Issue 02/03),
  bukan literal `default`.

### K12. Urutan cascade: perangkat dulu (niat), status DB mengikuti — degradasi anggun

Referensi meng-update perangkat lebih dulu lalu set status DB, dan sengaja tidak
menggagalkan seluruh operasi bila satu target gagal (OR-semantics). Di Polyglot
(K4 asinkron): status subscription berubah saat baris `sync_log` berhasil
**ditulis** (bukan saat perangkat sukses); kegagalan per-target ditangani
per-baris oleh Sync Engine. Bila tak ada target perangkat valid (device
unreachable / belum terpasang), kebijakan eksplisit: **boleh** ubah status lokal
saja **dengan** menandai `status_history.reason='local-only, device unreachable'`
— jangan diam-diam sukses tanpa jejak, jangan pula kehilangan paritas fitur.

### K13. Identitas & katalog vendor = DATA milik driver, multi-path/multi-value

Terverifikasi bahwa pengetahuan vendor jauh lebih kaya dari "satu path/nilai":
- **OID OLT** = katalog **per-profile (model)**, bukan per-brand; satu merk punya
  banyak profile (ZTE C300/C600/…). Tiap profile: `status_table`, `name_table`,
  `sn_table`, `rx_power_table`, `tx_power_table`, `distance_table` +
  `distance_tenths_meter`, `unauth_sn_table`, `offline_reason_table`,
  `probe_oid`, `online_values`. Deteksi profile via `probe_oid` (walk pertama
  yang merespons menang). `online_values` = himpunan nilai/label (numerik **atau**
  string) yang berarti online — beda per merk & per EPON/GPON; jangan hardcode
  `status==1`.
- **Path TR-069** (WiFi SSID/password, WAN, RX power) = **multi-path shotgun**:
  satu operasi abstrak di-Translate ke beberapa path native
  (`InternetGatewayDevice.*` + `Device.*` + vendor `X_BROADCOM_*`/`X_ALU-COM_*`)
  karena vendor ONU beda menyimpan di path berbeda; index band 5G ditemukan
  dinamis, bukan hardcode.
- **Decode RX/TX power** bukan divider tetap: auto-scale by magnitude, signed-16
  untuk unsigned, buang sentinel 0/65535, pilih skala yang menghasilkan dBm
  plausibel (~ -60..+10). Katalog cukup simpan `scale_hint`; decode = fungsi
  per-brand di driver, bukan aritmetika di usecase.

Semua ini hidup di `internal/driver/<vendor>/` (AGENTS.md §1.2), bukan usecase.

### K14. Semua identitas voucher/status ditulis dengan comment ber-prefix kanonik

Model validity/enforcement referensi bergantung pada `comment` user di router.
Untuk mendukung pass rekonsiliasi (K15) dan on-login guard, setiap user hotspot
yang dibuat Polyglot ditulis dengan `comment` ber-prefix yang bisa di-parse balik
(mis. `poly:vc:<voucher_id>,exp:<ts>`). On-login script (bila dipakai) **wajib
men-guard pada prefix** ini supaya (a) idempoten — login ulang tidak mereset
expiry, dan (b) tidak menyentuh user subscription/manual. Format comment kanonik
didokumentasikan di ADR 0012.

### K15. Enforcement hybrid = deviasi audit yang wajib direkonsiliasi

Bila memakai script router (`on-login` + `/system scheduler`) sebagai jaring
pengaman offline (Issue 13; juga preseden PPPoE isolir di referensi), itu aksi
perangkat **di luar** `command_audit_log` (melanggar K4). Boleh dipakai **hanya
dengan**: (a) pemasangan script/scheduler-nya sendiri lewat sync-log (teraudit),
(b) pass rekonsiliasi Golang periodik yang membaca state router (comment
`AUTO-ISOLIR`/expiry, scheduler/script yatim) dan menuliskan divergensi ke audit,
(c) dokumentasi deviasi eksplisit di ADR. Rekomendasi desain Polyglot: pilih
**scheduler harian tunggal** (bukan objek per-user bernama=username) supaya tak
ada orphan scheduler/script — lebih bersih dari referensi.

## Ringkasan REST API Lintas Issue

Tabel indeks cepat; detail (request/response field, role) ada di file issue
masing-masing.

| Issue | Endpoint utama | Aksi |
|---|---|---|
| 01 | `GET /api/v1/sync-logs`, `GET /api/v1/sync-logs/:id`, `POST /api/v1/sync-logs/:id/retry` | Amati & retry sinkronisasi |
| 02 | `GET/POST /api/v1/plans/:id/router-profiles`, `POST .../:profileId/sync` | Sinkron paket ↔ profil MikroTik |
| 03 | `POST /api/v1/subscriptions/:id/provision`, `POST .../activate` | Provision PPPoE |
| 04 | `POST /api/v1/subscriptions/:id/suspend`, `.../resume`, `.../terminate` | Cascade suspend/resume/terminate |
| 05 | `GET /api/v1/onu-discovery`, `POST /api/v1/onu-discovery/scan`, `POST .../:id/bind`, `POST .../:id/ignore` | Discovery ONU unconfigured |
| 06 | `POST /api/v1/subscriptions/:id/onu/authorize` | Otorisasi ONU di OLT |
| 07 | `GET /api/v1/olt-vendors` | Katalog vendor OLT SNMP |
| 08 | `GET /api/v1/acs-devices`, `POST /api/v1/acs-devices/sync`, `GET /api/v1/subscriptions/:id/acs-device` | Cache device GenieACS |
| 09 | `POST /api/v1/subscriptions/:id/cpe/reboot`, `.../wan`, `.../wifi` | Operasi CPE TR-069 |
| 10 | `GET /api/v1/subscriptions/:id/optical-health`, `GET /api/v1/optical-health/alerts` | RX power monitoring |
| 11 | `GET/POST /api/v1/odps`, `GET/POST /api/v1/odp-links`, `GET /api/v1/topology` | Topologi ODP |
| 12 | `GET /api/v1/subscriptions/:id/sessions`, `GET /api/v1/sessions/active` | Sesi online pelanggan |
| 13 | `POST/GET /api/v1/voucher-batches`, `POST .../:id/revoke`, `GET /api/v1/vouchers`, `POST /api/v1/vouchers/:id/disable\|enable`, `DELETE /api/v1/vouchers/:id`, `POST /api/v1/subscriptions/:id/provision-hotspot` (mode `mac_login`/`ip_binding`) | Hotspot & voucher lifecycle + identitas modem |
| 14 | `GET/PUT /api/v1/subscriptions/:id/suspension-policy`, `GET/POST /api/v1/billing/auto-suspend/{candidates,run}`, `GET/POST /api/v1/billing/auto-restore/{candidates,run}` | Auto-suspend/restore terjadwal (billing-driven) |

## Open Questions (diputuskan sebelum/di awal fase terkait)

1. **Scheduler**: mekanisme cron untuk job terjadwal (GenieACS sync Issue 08, RX
   power Issue 10, OLT discovery Issue 05, voucher expiry Issue 13, auto-suspend
   Issue 14) belum ada di repo. Semua issue itu memakai pola yang sama: ticker
   in-process + errgroup di `main.go`. Rekomendasi: ticker sederhana dulu; angkat
   jadi komponen scheduler bersama hanya bila jumlah job terjadwal makin banyak
   (jangan dibangun prematur). Diputuskan di Issue 01, dipakai ulang Issue
   05/08/10/13/14.
2. **HITL untuk provisioning otomatis**: apakah provisioning terjadwal
   (mis. auto-suspend telat bayar) juga lewat HITL, atau di-`auto_approved`
   dengan `source='scheduled_job'`? Lihat `command.Decide`. Rekomendasi:
   destruktif tetap butuh approval kecuali ada kebijakan whitelist eksplisit.
3. **InfluxDB availability**: apakah instance InfluxDB sudah tersedia di
   deployment? Kalau belum, Issue 10 diblok sampai infra ada (tapi Issue 08/09
   tidak tergantung ini).
4. **Multi-vendor OLT scope**: mulai dari berapa merk? Rekomendasi: ZTE +
   Huawei dulu (yang drivernya sudah ada kerangkanya), sisanya lewat katalog
   OID Issue 07 seiring kebutuhan.
5. **ODP graph (Issue 11)**: apakah topologi riil ISP ini bertingkat
   (butuh `odp_links`) atau flat? Kalau flat, Issue 11 dilewati.
