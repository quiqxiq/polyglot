# Analisis Provisioning/Sinkronisasi — 4 Repo Referensi vs Polyglot

> **Status:** v1 — hasil analisis LANGSUNG terhadap source code (clone git,
> bukan cuma baca README), dengan kutipan kode nyata sebagai bukti setiap
> klaim.
> **Repo dianalisis:** `alijayanet/gembok-simple` (PHP), `alijayanet/billing-rtrw`
> (Node.js), `alijayanet/gembok-bill` (Node.js), `alijayanet/mikhmon-agent` (PHP)
> **Fokus:** provisioning & sinkronisasi ke MikroTik, GenieACS, OLT, ONU
> **Melengkapi:** `docs/adr/0003-mikrotik-dual-connection-streaming.md`,
> `Polyglot-Architecture.md`

---

## 0. Metode

Keempat repo di-`git clone` langsung ke sandbox dan dibaca source code-nya
— bukan cuma README atau deskripsi GitHub. Fokus baca: file yang benar-benar
menyentuh koneksi ke MikroTik, GenieACS, dan OLT/ONU (bukan UI/routing/
WhatsApp gateway, yang di keempat repo ini porsinya jauh lebih besar dari
bagian provisioning-nya).

| Repo | Bahasa | Bagian yang dibaca |
|---|---|---|
| `gembok-simple` | PHP | `includes/mikrotik_api.php`, `api/onu_locations.php`, `api/genieacs.php` |
| `billing-rtrw` | Node.js | `services/oltService.js` (2197 baris), `services/onuProvisionService.js`, `services/mikrotikService.js` (1708 baris), `config/genieacs.js`, `config/database.js` |
| `gembok-bill` | Node.js | `config/serviceSuspension.js`, `config/staticIPSuspension.js`, `config/rxPowerMonitor.js`, `config/mikrotik.js` |
| `mikhmon-agent` | PHP | `lib/routeros_api.class.php`, `process/psecret.php`, `genieacs/api.php` |

Keempatnya adalah sistem billing ISP RT/RW-Net Indonesia yang benar-benar
dipakai di produksi (bukan contoh/tutorial) — jadi pola yang ditemukan di
sini adalah pola yang **sudah teruji di lapangan**, bukan teori.

---

## 1. Ringkasan Singkat Tiap Repo

- **gembok-simple** (PHP prosedural) — billing RT/RW-Net paling ringkas dari
  keempatnya: pelanggan, invoice, voucher, integrasi MikroTik + GenieACS,
  ada peta lokasi ONU (`onu_locations`) dan topologi ODP.
- **billing-rtrw** (Node.js/Express) — paling lengkap untuk sisi jaringan:
  py `oltService.js` mendukung **8 merk OLT** lewat SNMP, py
  `onuProvisionService.js` provisioning ONU lewat SSH mentah per-vendor, py
  banyak service terpisah (agent, teknisi, kolektor, payroll, inventory).
- **gembok-bill** — fokus ke integrasi WhatsApp/Telegram sebagai kanal
  utama (registrasi pelanggan, notifikasi, command admin lewat chat) di atas
  fondasi MikroTik+GenieACS yang mirip billing-rtrw; py monitoring RX power
  optik dengan notifikasi threshold.
- **mikhmon-agent** — fork dari **Mikhmon** (tool monitoring hotspot MikroTik
  yang sangat umum dipakai admin RT/RW-Net Indonesia), ditambah lapisan
  billing/agent di atasnya. `lib/routeros_api.class.php`-nya adalah kelas API
  RouterOS PHP klasik yang jadi rujukan banyak proyek PHP MikroTik lainnya.

---

## 2. Temuan A — Pola Koneksi MikroTik: Connect-per-Call, Bukan Persisten

Kutipan dari `billing-rtrw/services/mikrotikService.js`:

```javascript
async function getPppoeProfiles(routerId = null) {
  // ...
  const conn = await getConnection(routerId);   // dial baru
  // ... satu operasi ...
  conn.api.close();                              // langsung tutup
}
```

Pola ini konsisten di **hampir setiap fungsi** di file itu (1708 baris):
`getConnection()` di awal fungsi, satu operasi, `close()` di `finally`.
`mikhmon-agent`'s `routeros_api.class.php` juga sama — satu `connect()` per
request PHP (wajar untuk PHP, karena PHP memang stateless per-request).

**Ini BERBEDA dari keputusan `docs/adr/0003-mikrotik-dual-connection-streaming.md`**,
di mana `netops-engine` sengaja mempertahankan **2 koneksi persisten**
(exec + stream) seumur hidup `Driver`.

### Kenapa perbedaan ini BENAR, bukan sinyal bahwa desain kita salah

Keempat repo referensi ini **tidak melakukan streaming sama sekali** ke
MikroTik — semua operasinya request-response sekali jalan (`getPppoeProfiles`,
`addPppoeSecret`, dst). Untuk beban kerja seperti itu, connect-per-call itu
valid dan lebih sederhana (tidak perlu reconnect supervisor, tidak perlu
semaphore serialisasi wire). **Kebutuhan kita berbeda secara struktural**:
`netops-engine` punya `Stream()` (ping, monitor-traffic, log follow) yang
harus tetap mengalir SAMBIL command lain dieksekusi — persis masalah yang
mendorong ADR 0003 dan yang TIDAK pernah dihadapi keempat repo ini karena
mereka tidak mencobanya.

**Kesimpulan**: keputusan dual-connection persisten kita tetap tepat untuk
kebutuhan yang lebih luas (mendukung streaming). Tidak ada yang perlu
diubah di `internal/driver/mikrotik` berdasarkan temuan ini — justru temuan
ini memperkuat kenapa desain kita perlu lebih rumit dari referensi umum.

### Yang MEMANG layak diadopsi dari pola mereka

`getConnection()` di `billing-rtrw` punya **cache probe koneksi** (baris
~294-310): kalau TCP connect ke suatu host baru saja gagal, catat
`failUntil` selama 5 detik supaya percobaan berikutnya tidak menunggu
timeout penuh lagi. Ini pola kecil dan murah untuk ditiru di
`internal/driver/mikrotik/connect.go` — supervisor kita SUDAH punya backoff
untuk reconnect loop, tapi belum ada "circuit breaker" jangka pendek untuk
`NewDriver()` pertama kali. **Rekomendasi**: tambahkan cache singkat serupa
di `dialAndLogin` agar percobaan `NewDriver` yang gagal berturut-turut
(mis. dipanggil `internal/registry` berkali-kali sebelum device benar-benar
online) tidak masing-masing menunggu TCP timeout penuh.

---

## 3. Temuan B — OLT: 8 Profil SNMP Vendor + Discovery ONU Belum Terkonfigurasi

`billing-rtrw/services/oltService.js` (2197 baris) punya tabel OID SNMP
lengkap untuk **8 merk OLT**: ZTE, Huawei, VSOL, Hioso, HSGQ, Fiberhome,
BDCOM, CDATA. Contoh untuk ZTE:

```javascript
zte: [{
  name: 'ZTE_GPON_C300',
  status_table: '1.3.6.1.4.1.3902.1082.500.10.2.3.3.1.9',
  sn_table:     '1.3.6.1.4.1.3902.1082.500.10.2.3.3.1.6',
  rx_power_table: '1.3.6.1.4.1.3902.1015.1010.11.2.1.2', // 0.01 dBm
  unauth_sn_table: '1.3.6.1.4.1.3902.1012.3.13.3.1.2',    // ONU belum dikonfigurasi
  distance_table: '1.3.6.1.4.1.3902.1015.1010.11.2.1.4',  // jarak fiber
}]
```

### 3.1 Temuan penting: alur "ONU belum terkonfigurasi" (unauthorized ONU discovery)

`unauth_sn_table` mendeteksi ONU yang **secara fisik sudah tersambung ke
port PON tapi belum di-otorisasi** di OLT. `onuProvisionService.js`
punya fungsi eksplisit untuk ini:

```javascript
async zteGetUnconfiguredONUs(oltConfig, pon) {
  const commands = ['enable', `show gpon onu uncfg gpon-olt_${pon}`];
  // ...
}
```

**Ini adalah workflow nyata yang SAMA SEKALI BELUM ADA di
`DATABASE-SCHEMA.md`.** Alur produksi yang sebenarnya:

1. Teknisi pasang ONU baru di rumah pelanggan.
2. ONU otomatis terdeteksi OLT sebagai "unconfigured" (SN terlihat, belum
   ada di database OLT).
3. Admin/teknisi pilih SN dari daftar "unconfigured", masukkan: PON port,
   ONU-ID, profil bandwidth, VLAN.
4. Sistem kirim SATU RANGKAIAN command CLI ke OLT (contoh ZTE, dari
   `onuProvisionService.js`):

```javascript
const cmds = [
  'enable', 'configure terminal',
  `interface gpon-olt_${pon}`,
  `onu ${onuId} type ${onuType} sn ${sn}`,
  'exit',
  `interface gpon-onu_${pon}:${onuId}`,
  `tcont 1 profile ${bandwidth}`, 'gemport 1 tcont 1',
  // opsional: ssid/security kalau ONU juga router WiFi
  // opsional: tr069 acs url ... <- OLT MENDORONG URL GenieACS ke ONU!
  'write'
];
```

5. Kalau `tr069 acs url` disertakan, **OLT-lah yang mendorong konfigurasi
   TR-069 (URL GenieACS) ke ONU** — ONU lalu meng-*inform* ke GenieACS
   dengan sendirinya begitu online.

### 3.2 Kesesuaian dengan `netops-engine`

- `internal/driver/zteolt` project ini SUDAH punya desain SNMP (monitoring)
  + Telnet mentah (provisioning) — arahnya SUDAH BENAR, tapi baru untuk
  **satu merk** (ZTE). Cakupan riil pasar RT/RW-Net Indonesia jauh lebih
  luas (8 merk di atas, dan itu belum termasuk semua varian).
- **Workflow discovery ONU belum terkonfigurasi TIDAK ADA** di
  `DATABASE-SCHEMA.md` — `subscriptions.onu_serial_number` mengasumsikan
  ONU SUDAH terikat, padahal ada tahap "ditemukan tapi belum diotorisasi"
  yang perlu direpresentasikan.
- **RX power / jarak fiber tidak ada** di skema manapun yang sudah kita
  buat.

### 3.3 Rekomendasi Konkret

1. **Tabel baru: `onu_discovery_queue`** — menyimpan ONU yang terdeteksi
   "unconfigured" dari hasil SNMP walk OLT (device_id, pon_port, serial_number,
   detected_at, status: `pending`/`bound`/`ignored`, bound_subscription_id
   nullable). Teknisi memilih dari daftar ini saat instalasi baru, bukan
   mengetik serial number manual (mengurangi salah ketik SN yang jadi
   sumber masalah lapangan paling umum).
2. **`subscriptions` perlu kolom tambahan**: `onu_pon_port`, `onu_id`
   (nomor ONU-ID di OLT, bukan cuma serial number) — provisioning command
   di atas butuh KEDUANYA, bukan cuma serial number.
3. **Tabel baru (atau perluasan `provisioning_sync_log`): `onu_health_readings`**
   — RX power (dBm), jarak fiber, temperature. **Rekomendasi: taruh di
   InfluxDB, bukan PostgreSQL** — ini data time-series, dan Anda sudah
   punya infrastruktur ini dari `roskit`/`roslib` (`Sink` interface untuk
   InfluxDB sudah pernah dibangun). Jangan bangun mekanisme time-series
   kedua di Postgres kalau sudah ada yang teruji.
4. **`internal/driver/` perlu strategi untuk banyak merk OLT SNMP**, bukan
   cuma ZTE. Karena OID per-merk itu murni DATA (bukan logic beda), ini
   cocok dengan pola yang SUDAH kita pakai untuk `genericssh`/`generictelnet`
   (ADR 0004): satu driver generik `internal/driver/genericsnmp` (baru) +
   katalog OID per merk sebagai data (mirip `genericcli.Catalog`), BUKAN
   satu paket Go per merk OLT. Kalau nanti dikerjakan, ini idealnya jadi
   ADR baru (0005) mengikuti pola yang sudah ditetapkan.

---

## 4. Temuan C — GenieACS: Cache Lokal + Tag-Based Linking + Model Polling

`billing-rtrw/config/database.js` punya tabel `acs_devices` — **cache
lokal** dari device GenieACS:

```sql
CREATE TABLE acs_devices (
  id TEXT PRIMARY KEY,               -- ID device asli dari GenieACS
  serial_number TEXT, manufacturer TEXT, product_class TEXT,
  connection_request_url TEXT,        -- untuk "colek" device on-demand
  tags TEXT DEFAULT '[]',             -- JSON array, termasuk tag customer
  params TEXT DEFAULT '{}',           -- cache SEMUA parameter TR-069
  last_inform DATETIME
);

CREATE TABLE acs_tasks (
  device_id TEXT NOT NULL, name TEXT NOT NULL, payload TEXT DEFAULT '{}',
  status TEXT DEFAULT 'pending', retry_count INTEGER DEFAULT 0
);
```

Dan customer terhubung ke device GenieACS lewat **tag**, bukan foreign key:

```sql
CREATE TABLE customers (
  -- ...
  genieacs_tag TEXT DEFAULT ''   -- dicocokkan ke acs_devices.tags
);
```

### 4.1 Kenapa pola tag ini penting

GenieACS sendiri secara native mendukung "tag" per device (bisa di-set
lewat NBI-nya) — jadi keempat repo referensi konsisten memakai mekanisme
BAWAAN GenieACS untuk linking, bukan membuat sistem linking sendiri.

### 4.2 Polling, bukan streaming — dan ini SAH, bukan pelanggaran prinsip kita

```javascript
// config/genieacs.js
function scheduleMonitoring() {
  setInterval(async () => { /* cek device GenieACS berkala */ }, ...);
}
```

Dan `gembok-bill/config/rxPowerMonitor.js`:

```javascript
const response = await axios.get(`${genieacsUrl}/devices`, ...);
const warningThreshold = getSetting('rx_power_warning', -25);   // dBm
const criticalThreshold = getSetting('rx_power_critical', -27); // dBm
```

**Ini polling, dan itu benar untuk TR-069/GenieACS** — bukan pengecualian
yang melemahkan prinsip "tidak ada polling" yang sudah ditetapkan untuk
`netops-engine` (lihat ADR 0003). Prinsip itu berlaku untuk protokol yang
GENUINELY mendukung push/streaming (RouterOS API, yang punya `Listen`).
TR-069 secara desain protokol adalah **device yang menghubungi ACS secara
berkala** ("Inform") — ACS tidak bisa "subscribe" ke perubahan device
secara real-time selain menunggu Inform berikutnya atau mengirim
"Connection Request" untuk memicu Inform lebih awal. Polling periodik ke
NBI GenieACS untuk data seperti RX power adalah **satu-satunya cara yang
masuk akal**, bukan kemalasan desain.

### 4.3 Kesesuaian dengan `netops-engine`

`DATABASE-SCHEMA.md` saat ini memperlakukan GenieACS HANYA sebagai satu
`target_type` di `provisioning_sync_log` — tidak ada tabel cache device
GenieACS sendiri. Ini **gap nyata**: tanpa cache lokal, setiap halaman
dashboard yang menampilkan status ONU pelanggan harus memanggil API
GenieACS secara langsung (lambat, dan gagal total kalau GenieACS sedang
down meski datanya cuma perlu ditampilkan, bukan diubah).

### 4.4 Rekomendasi Konkret

1. **Tabel baru: `acs_devices`** (cache lokal, mirror dari GenieACS) — id
   (device ID asli GenieACS), serial_number, manufacturer, product_class,
   connection_request_url, tags (JSONB), params (JSONB, TERBATAS ke
   parameter yang benar-benar dipakai UI kita, bukan seluruh device tree
   TR-069 yang bisa sangat besar), last_inform, synced_at.
2. **`subscriptions` cukup simpan `genieacs_device_id` (referensi ke
   `acs_devices.id`)** — BUKAN duplikasi `genieacs_tag` sebagai mekanisme
   linking terpisah. Karena `subscriptions` di skema kita SUDAH jadi
   sumber kebenaran utama (beda dari keempat repo referensi yang
   `customers`-nya langsung pegang tag), linking lewat FK eksplisit lebih
   konsisten dengan prinsip desain kita sendiri di §1 `DATABASE-SCHEMA.md`.
3. **Job sinkron `acs_devices` dari GenieACS: polling terjadwal**, taruh
   di `internal/usecase/network` sebagai usecase baru (mis.
   `SyncGenieACSDevices`), dipanggil dari scheduler (`internal/config`
   sudah punya placeholder, `scripts/seed.go` bisa jadi rujukan pola cron
   job kalau belum ada mekanisme scheduler lain).
4. **`acs_tasks` = pola yang SAMA dengan `provisioning_sync_log`** —
   jangan bikin tabel task queue GenieACS terpisah. Cukup pakai
   `provisioning_sync_log` yang sudah ada dengan `target_type='genieacs_tr069'`
   dan `external_reference` diisi task ID dari GenieACS NBI (sudah
   direncanakan di `DATABASE-SCHEMA.md` §7.3 — validasi ini dari repo
   referensi MENGONFIRMASI keputusan itu sudah tepat, tidak perlu diubah).

---

## 5. Temuan D — Suspend/Isolir: Cascade ke Banyak Target Sekaligus, Bukan Satu

`gembok-bill/config/serviceSuspension.js`:

```javascript
async suspendCustomerService(customer, reason) {
  // 1. Prioritas: PPPoE -> ganti profile ke 'isolir'
  if (customer.pppoe_username) { /* set profile isolir di Mikrotik */ }
  // 2. Kalau tidak ada PPPoE: static IP -> address-list suspension
  else if (customer.static_ip) { await staticIPSuspension.suspendStaticIPCustomer(...); }
  // 3. TAMBAHAN, bukan pengganti: suspend via GenieACS (disable WAN di CPE)
  if (genieacsDevice) { /* disable WAN connection di parameter TR-069 */ }
  // 4. Baru update status billing
  await billingManager.setCustomerStatusById(customer.id, 'suspended');
}
```

**Poin kunci**: satu event bisnis ("suspend pelanggan X") bisa memicu
**LEBIH DARI SATU** target provisioning sekaligus (MikroTik DAN GenieACS),
tidak selalu cuma satu.

### Kesesuaian dengan `netops-engine`

Ini justru **mengonfirmasi** desain `provisioning_sync_log` di
`DATABASE-SCHEMA.md` §6.3 sudah benar: satu `subscription_id` bisa punya
BANYAK baris `provisioning_sync_log` untuk SATU event (satu untuk
`target_type='mikrotik_ppp_secret'`, satu lagi untuk
`target_type='genieacs_tr069'`, keduanya `action='disable'`, ditulis
dalam transaksi bisnis yang sama). **Tidak perlu perubahan skema** — ini
kasus penggunaan yang MEMANG sudah didesain untuk itu, cuma perlu
dipastikan usecase Go-nya nanti benar-benar menulis multi-baris saat
suspend, bukan cuma satu.

---

## 6. Temuan E — Topologi ODP: Graf, Bukan Cuma Hierarki Satu Arah

`gembok-simple` py tabel `odp_links`:

```sql
-- from_odp_id, to_odp_id — merepresentasikan kabel fiber ANTAR ODP,
-- bukan cuma ODP -> OLT satu arah
SELECT * FROM odp_links WHERE (from_odp_id = ? AND to_odp_id = ?) OR ...
```

`DATABASE-SCHEMA.md` saat ini cuma punya `odcs`/`odps` dengan hierarki
satu arah ke `devices` (OLT) — tidak merepresentasikan kabel ODP-ke-ODP
(splitter bertingkat, umum di FTTH: OLT → ODC → ODP besar → ODP kecil).

### Rekomendasi

Tambahkan tabel `odp_links` (from_odp_id, to_odp_id, cable_length_meters
nullable, keduanya FK ke `odps`) kalau topologi jaringan Anda memang
bertingkat lebih dari satu level splitter. Kalau topologi Anda selalu
flat (satu ODP langsung ke OLT), lewati saja — jangan tambah tabel yang
tidak akan dipakai.

---

## 7. Yang SENGAJA Tidak Diadopsi (dan Alasannya)

- **Hand-rolled SSH expect-loop** (`onuProvisionService.js` pakai `ssh2`
  mentah + regex prompt manual) — `netops-engine` sudah punya
  `internal/driver/genericssh` di atas scrapligo (ADR 0004), yang
  menangani prompt/paging/privilege-level secara generik dan sudah
  divalidasi. Membangun expect-loop kedua secara terpisah untuk OLT
  hanya akan menduplikasi apa yang sudah scrapligo tangani — kalau ada
  merk OLT tanpa dukungan built-in scrapligo, jalurnya tetap
  `genericssh` + platformdef YAML custom (persis pola yang sudah dipakai
  untuk `mikrotik_routeros.yaml`), bukan implementasi SSH terpisah.
- **PHP session-per-request untuk MikroTik** (`mikhmon-agent`) — wajar
  untuk PHP, tidak relevan untuk Go yang sudah punya `internal/registry`
  sebagai pemegang koneksi persisten seumur proses.
- **SQLite sebagai database utama** (`billing-rtrw`, `gembok-bill`) —
  project ini sudah komit ke PostgreSQL (`TECH-STACK-DAN-PERSIAPAN.md`),
  keputusan itu tidak berubah oleh temuan ini.

---

## 8. Ringkasan Actionable

| # | Rekomendasi | Dampak ke Artefak yang Sudah Ada |
|---|---|---|
| 1 | Tabel `onu_discovery_queue` (ONU belum diotorisasi) | Tambahan migration baru |
| 2 | `subscriptions` +`onu_pon_port`, +`onu_id` | `DATABASE-SCHEMA.md` §6.1 + migration `000009` perlu kolom tambahan |
| 3 | Tabel `acs_devices` (cache lokal GenieACS) | Tambahan migration baru + `DATABASE-SCHEMA.md` §7 baru |
| 4 | `subscriptions` +`genieacs_device_id` (FK ke `acs_devices`) | Migration `000009` |
| 5 | RX power/jarak fiber → **InfluxDB**, bukan Postgres | Tidak ada migration baru; pakai `roskit`/`roslib` Sink yang sudah ada |
| 6 | Driver OLT SNMP generik multi-merk (`genericsnmp` + katalog OID) | ADR baru (0005) kalau dikerjakan — BELUM dikerjakan di percakapan ini |
| 7 | Circuit-breaker singkat di `dialAndLogin` (cache probe koneksi) | Perbaikan kecil di `internal/driver/mikrotik/connect.go` |
| 8 | `odp_links` (topologi ODP graf) — opsional, tergantung topologi riil Anda | Tambahan migration, kondisional |
| 9 | Multi-baris `provisioning_sync_log` per event suspend | Tidak ada perubahan skema — sudah didukung, perlu dipastikan usecase Go menulis multi-baris |

**Yang TIDAK saya kerjakan di dokumen ini**: saya belum menulis migration
SQL atau kode Go untuk item di atas — ini murni analisis + rekomendasi.
Kalau Anda mau saya lanjutkan salah satu (mis. migration `acs_devices`
atau `onu_discovery_queue`, atau driver `genericsnmp`), sebut saja yang
mana duluan.
