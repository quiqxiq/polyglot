# Analisis API GenieACS (NBI — Northbound Interface)

**Sumber**: Dokumentasi resmi GenieACS (docs.genieacs.com — versi `latest`/1.3.0-dev), dicek langsung per Juli 2026, disilangkan dengan GenieACS Forum resmi untuk bagian autentikasi.

**Catatan penting di awal**: Dokumen ini secara eksplisit membedakan mana yang **terdokumentasi resmi** dan mana yang **belum/tidak terdokumentasi** (ditandai TODO oleh GenieACS sendiri), karena ini berpengaruh besar pada bagian autentikasi.

---

## 1. Gambaran Arsitektur

GenieACS bukan satu proses monolitik, melainkan 4 komponen terpisah yang saling independen:

| Komponen | Port default | Fungsi | Protokol |
|---|---|---|---|
| `genieacs-cwmp` | 7547 | Server TR-069/CWMP — ini yang diajak bicara oleh CPE (ONU/modem/router) | SOAP/XML via HTTP(S) |
| `genieacs-nbi` | 7557 | **Northbound Interface** — REST API yang jadi objek analisis ini | HTTP + JSON |
| `genieacs-fs` | 7567 | File server — tempat CPE mengunduh firmware/config file | HTTP |
| `genieacs-ui` | 3000 | Web GUI (frontend admin) | HTTP + JWT cookie |

Yang dimaksud "API endpoint" pada dokumentasi resmi GenieACS **hanya merujuk pada `genieacs-nbi`** (port 7557). CWMP (7547) bukan REST API — itu protokol TR-069 yang dipakai CPE, formatnya SOAP/XML, bukan untuk diakses developer secara langsung.

---

## 2. Endpoint NBI — Daftar Lengkap

Base URL contoh: `http://localhost:7557` (default, non-TLS, sesuai port default `NBI_PORT=7557`).

### 2.1 Search / Query (GET)

```
GET /<collection>/?query=<query>
```

- **Fungsi**: mencari record di database berdasarkan query bergaya MongoDB.
- **`<collection>`** yang didukung: `devices`, `tasks`, `presets`, `files` (juga disebut `objects` di versi wiki lama — pada dokumentasi rst saat ini istilah resminya adalah "devices, tasks, presets, files").
- **`query`**: string JSON, di-URL-encode, mengikuti sintaks query MongoDB (`$lt`, `$gt`, `$gte`, `$lte`, `$ne`, dsb.).
- **Parameter tambahan**: `projection` — daftar nama parameter dipisah koma, untuk membatasi field yang dikembalikan (hemat bandwidth/parsing).
- **Response**: JSON array dari seluruh item yang cocok.

Contoh pola query resmi:
- Cari device by ID: `{"_id": "202BC1-BM632w-000000"}`
- Cari device by MAC address (path parameter TR-069 penuh sebagai key)
- Cari device yang belum inform > 7 hari: `{"_lastInform": {"$lt": "<timestamp>"}}`
- Cari task pending milik device tertentu: `GET /tasks/?query={"device":"<device_id>"}`

Endpoint turunan yang eksplisit disebut di dokumentasi:
```
GET /files/                              → semua file yang pernah diupload
GET /files/?query={"filename":"<nama>"}  → cari file spesifik
```

### 2.2 Devices

| Method | Endpoint | Fungsi |
|---|---|---|
| GET | `/devices/?query=<query>` | Cari/list device |
| POST | `/devices/<device_id>/tasks?[connection_request]` | Enqueue satu/lebih task ke device, opsional langsung trigger connection request |
| DELETE | `/devices/<device_id>` | Hapus device dari database (device akan register ulang otomatis saat inform berikutnya — ini **bukan** cara "memutus" device secara permanen) |
| POST | `/devices/<device_id>/tags/<tag>` | Tempelkan tag ke device (no-op jika tag sudah ada) |
| DELETE | `/devices/<device_id>/tags/<tag>` | Lepas tag dari device |

Detail parameter `POST /devices/<device_id>/tasks`:
- `connection_request` (query flag, tanpa value): jika disertakan, GenieACS akan langsung mengirim connection request ke CPE agar task dieksekusi saat itu juga (bukan menunggu periodic inform berikutnya).
- Body request = JSON representasi objek task (lihat §3 untuk semua tipe task).
- **Response code**:
  - `200` → task berhasil dieksekusi langsung (device merespons sebelum timeout).
  - `202` → task masuk antrean, akan dieksekusi pada inform berikutnya (device tidak merespons dalam batas timeout, atau tidak pakai `connection_request`).
- Response body = objek task persis seperti tersimpan di database, termasuk field `_id` yang bisa dipakai untuk lookup/retry/delete task tersebut nanti.
- Ada juga query param `timeout` (dalam milidetik) yang muncul di contoh curl resmi, mengatur berapa lama NBI menunggu respons device sebelum memutuskan 200 vs 202.

### 2.3 Tasks

| Method | Endpoint | Fungsi |
|---|---|---|
| GET | `/tasks/?query=<query>` | Cari/list task |
| POST | `/tasks/<task_id>/retry` | Retry task yang gagal (fault), akan dicoba lagi di inform berikutnya |
| DELETE | `/tasks/<task_id>` | Hapus task dari antrean/database |

`<task_id>` didapat dari field `_id` hasil `GET /tasks` atau dari response `POST /devices/<device_id>/tasks`.

### 2.4 Faults

| Method | Endpoint | Fungsi |
|---|---|---|
| DELETE | `/faults/<fault_id>` | Hapus record fault |

- Format `fault_id`: `<device_id>:<channel>` — contoh: `202BC1-BM632w-000000:default`.
- Tidak ada endpoint `POST /faults` yang terdokumentasi — fault dibuat otomatis oleh sistem saat task gagal, bukan dibuat manual via API.
- Endpoint `GET /faults/?query=<query>` tidak dijelaskan secara eksplisit dalam contoh, tapi mengikuti pola umum `GET /<collection>/?query=<query>` di §2.1 karena faults juga merupakan collection yang bisa di-query dengan cara sama.

### 2.5 Tags (via device)

Sudah tercakup di §2.2 — tag di GenieACS selalu beroperasi pada scope device tertentu (`/devices/<device_id>/tags/<tag>`), tidak ada endpoint tag berdiri sendiri (`/tags/`) yang terpisah dari device.

### 2.6 Presets

| Method | Endpoint | Fungsi |
|---|---|---|
| PUT | `/presets/<preset_name>` | Buat atau update preset (idempotent — nama sama akan overwrite) |
| DELETE | `/presets/<preset_name>` | Hapus preset |
| GET | `/presets/?query=<query>` | Cari/list preset (mengikuti pola §2.1) |

Body `PUT /presets/<preset_name>` adalah JSON dengan struktur:
```
weight            : integer (urutan prioritas eksekusi preset, makin kecil makin dulu)
precondition      : string JSON — filter kondisi device yang akan kena preset ini
configurations    : array of configuration object
```
Detail `precondition` dan `configurations` di §4.

### 2.7 Files

| Method | Endpoint | Fungsi |
|---|---|---|
| PUT | `/files/<file_name>` | Upload file baru atau overwrite file lama. Body request = **binary content file itu sendiri** (bukan JSON) |
| DELETE | `/files/<file_name>` | Hapus file yang sudah diupload |
| GET | `/files/` | List semua file |
| GET | `/files/?query={"filename":"<filename>"}` | Cari file spesifik |

Metadata file dikirim sebagai **HTTP header**, bukan JSON body (karena body dipakai untuk isi file):
- `fileType` — nilai standar: `"1 Firmware Upgrade Image"`, `"2 Web Content"`, `"3 Vendor Configuration File"`.
- `oui` — OUI (Organizationally Unique Identifier) dari model device yang dituju.
- `productClass` — product class device.
- `version` — versi firmware (jika `fileType` adalah firmware image).

### 2.8 Provisions

| Method | Endpoint | Fungsi |
|---|---|---|
| PUT | `/provisions/<provision_name>` | Buat/update provision script. Body request = **kode JavaScript mentah** (bukan JSON, plain text) |
| DELETE | `/provisions/<provision_name>` | Hapus provision |
| GET | `/provisions/` | List semua provision |

Provision adalah script JavaScript yang dijalankan dalam sandbox untuk mendeklarasikan konfigurasi device secara deklaratif (pakai fungsi `declare()`, `commit()`, `clear()`, `log()`, dsb — ini bahasa scripting internal GenieACS, bukan endpoint REST tambahan).

---

## 3. Task — Tipe dan Struktur Objeknya

Task adalah unit kerja yang dikirim ke CPE lewat `POST /devices/<device_id>/tasks`. Semua tipe berikut terdokumentasi resmi:

| Task name | Fungsi | Field wajib tambahan |
|---|---|---|
| `getParameterValues` | Ambil nilai satu/lebih parameter TR-069 dari device | `parameterNames`: array string |
| `refreshObject` | Paksa GenieACS membaca ulang seluruh sub-tree parameter dari device (refresh cache lokal) | `objectName`: string (bisa string kosong `""` untuk refresh root/semua) |
| `setParameterValues` | Set satu/lebih nilai parameter di device | `parameterValues`: array of array `[nama_parameter, value, xsd_type]` — bisa banyak sekaligus dalam satu task |
| `addObject` | Tambah instance object baru (mis. buat WAN connection baru) | `objectName`: string, path ke parent object |
| `deleteObject` | Hapus instance object | `objectName`: string, path ke object spesifik (termasuk index-nya, mis. `...WANPPPConnection.1`) |
| `reboot` | Reboot device | — (tidak perlu field tambahan selain `name`) |
| `factoryReset` | Factory reset device | — (tidak perlu field tambahan selain `name`) |
| `download` | Perintahkan device mengunduh file (firmware/config) dari `genieacs-fs` | `file`: nama file (harus sudah diupload lewat `PUT /files/<file_name>`) |

Catatan implementasi penting dari dokumentasi resmi:
- `setParameterValues` menerima **banyak parameter dalam satu task** — cukup tambahkan array baru ke dalam `parameterValues`, tidak perlu kirim task terpisah per parameter.
- Setelah task `getParameterValues` sukses, nilai yang didapat **tidak dikembalikan langsung di response POST task** — melainkan disimpan ke database device. Harus melakukan `GET /devices/?query={"_id":"<device_id>"}` (dengan `projection` untuk efisiensi) setelahnya untuk membaca nilai tersebut dari objek device.

---

## 4. Presets — Precondition & Configuration

Preset = aturan otomatis "jika device memenuhi kondisi X, terapkan konfigurasi Y", dievaluasi berdasarkan kondisi (precondition), jadwal (cron expression), dan event tertentu.

### 4.1 Precondition
String JSON yang berisi filter (mirip query MongoDB) untuk menentukan device mana yang kena preset:
```
{"param": "value"}
{"param": "value", "param2": {"$ne": "value2"}}
```
Operator yang didukung resmi: `$gt`, `$lt`, `$gte`, `$lte` (selain kesetaraan langsung dan `$ne` yang muncul di contoh).

### 4.2 Configuration
Array berisi satu atau lebih objek konfigurasi, dengan 4 tipe:

| type | Fungsi | Field |
|---|---|---|
| `value` | Set nilai parameter | `name` (path parameter), `value` |
| `add_object` | Tambah object saat precondition terpenuhi | `name` (parent), `object` (nama object baru) |
| `delete_object` | Hapus object saat precondition terpenuhi | `name` (parent), `object` (nama object yang dihapus) |
| `provision` | Jalankan provision script tertentu | `name` (nama provision yang sudah dibuat lewat `PUT /provisions/<name>`) |

Ini artinya preset bisa dipakai untuk 2 gaya konfigurasi: deklaratif sederhana (`value`/`add_object`/`delete_object`) atau delegasi penuh ke script JavaScript custom (`provision`) untuk logika yang lebih kompleks.

---

## 5. Autentikasi — Bagian Paling Kritis (Baca Sebelum Implementasi)

Ini bagian yang paling sering disalahpahami, karena ada **tiga lapis autentikasi berbeda** di GenieACS yang sering tertukar. Saya pisahkan tegas per lapis, dan saya tandai jelas mana yang benar-benar terdokumentasi resmi vs. tidak.

### 5.1 Lapis 1 — CPE ke ACS (`genieacs-cwmp`, port 7547)

Ini autentikasi **antara perangkat (ONU/router) dan server GenieACS**, sama sekali beda dari REST API yang jadi topik utama analisis ini. Tapi tetap saya cakup karena Anda minta "tanpa terkecuali".

- **Default GenieACS**: menerima **semua** koneksi HTTP/HTTPS masuk tanpa autentikasi.
- **Cara mengaktifkan**: lewat GUI, menu `Admin → Config → New config`, buat key `cwmp.auth` dengan value boolean:
  - `true` → terima semua koneksi (perilaku default)
  - `false` → tolak semua koneksi masuk
- **Fungsi `AUTH(username, password)`**: dipakai di dalam expression `cwmp.auth` untuk mencocokkan kredensial yang dikirim CPE terhadap nilai yang ditentukan.
  - Bentuk statis: `AUTH("fixed-username", "fixed-password")`
  - Bentuk dinamis (baca dari parameter device itu sendiri): `AUTH(Device.ManagementServer.Username, Device.ManagementServer.Password)`
  - Parameter TR-069 terkait: `Device.ManagementServer.Username` / `InternetGatewayDevice.ManagementServer.Username` dan pasangan `Password`-nya (password selalu di-redact saat dibaca, tapi bisa di-set).
- **Fungsi `EXT(extension_name, function_name, ...args)`**: memanggil extension Node.js eksternal dari dalam expression auth, dipakai kalau kredensial disimpan di sumber luar (mis. database billing):
  ```
  AUTH(DeviceID.SerialNumber, EXT("authenticate", "getPassword", DeviceID.SerialNumber))
  ```

**Status dokumentasi bagian "ACS to CPE"** (yaitu kredensial yang dipakai GenieACS saat *mengirim* connection request ke CPE) — pada dokumentasi resmi versi saat ini statusnya masih **TODO**, belum ditulis.

### 5.2 Lapis 2 — Client ke NBI REST API (port 7557) — **INI YANG PALING PENTING UNTUK IMPLEMENTASI**

**Temuan tervalidasi langsung dari dokumentasi resmi**: halaman [`HTTPS`](https://docs.genieacs.com/en/latest/https.html) dan halaman [`Roles and Permissions`](https://docs.genieacs.com/en/latest/roles-and-permissions.html) pada dokumentasi resmi GenieACS **keduanya berstatus "TODO"** — isinya benar-benar kosong, belum ditulis oleh tim GenieACS. Ini bukan saya yang menyimpulkan; itu isi harfiah halamannya.

Konsekuensinya:
- **Tidak ada mekanisme autentikasi bawaan yang terdokumentasi resmi untuk NBI REST API.** Tidak ada Basic Auth, tidak ada API key, tidak ada OAuth/JWT untuk endpoint `/devices`, `/tasks`, `/presets`, `/files`, `/provisions`, dsb.
- Ini dikonfirmasi silang oleh thread resmi di GenieACS Forum ("Secure REST API endpoints of genieacs-nbi") tahun 2020, di mana pengembang inti GenieACS sendiri (user `akcoder`, kontributor resmi proyek) dan komunitas mengonfirmasi:
  - Endpoint seperti `/users`, `/config`, `/files`, `/objects` bisa diakses **tanpa autentikasi apa pun**.
  - Solusi yang direkomendasikan resmi oleh maintainer **bukan** fitur bawaan GenieACS, melainkan **mitigasi di luar aplikasi**:
    1. Pasang **reverse proxy** (nginx/Apache) di depan port 7557, tambahkan Basic Auth di level proxy.
    2. Batasi akses pakai **firewall/iptables** — hanya IP tertentu (server internal Anda) yang boleh mencapai port 7557.
  - Salah satu komentar eksplisit menyebut: *"Not only GenieACS does not have authentication, but doesn't have any kind of ACL"* — dengan kata lain, sampai saat itu tidak ada Access Control List sama sekali di level NBI.

**Soal `NBI_AUTHENTICATION_KEY` / header `x-api-key`** — ini poin yang butuh kehati-hatian ekstra karena beberapa sumber pihak ketiga (bukan dokumentasi resmi GenieACS) menyebutkannya seolah itu fitur resmi:
- Setelah ditelusuri ke sumber aslinya: ini berasal dari **Pull Request komunitas #374** ("Add HTTP request header authentication to NBI") yang diajukan oleh kontributor eksternal (`markabrahams`) ke repo GitHub GenieACS, sebagai **respons langsung** terhadap thread forum di atas.
- PR ini mengusulkan parameter konfigurasi opsional `NBI_AUTHENTICATION_KEY` — jika di-set, setiap request ke NBI wajib menyertakan header `x-api-key` yang cocok; jika tidak di-set, perilaku tetap seperti sebelumnya (tanpa auth).
- **Halaman resmi [`Environment Variables`](https://docs.genieacs.com/en/latest/environment-variables.html) yang saya cek langsung TIDAK mencantumkan `NBI_AUTHENTICATION_KEY`** di antara daftar variabel NBI yang ada (`NBI_WORKER_PROCESSES`, `NBI_PORT`, `NBI_INTERFACE`, `NBI_SSL_CERT`, `NBI_SSL_KEY`, `NBI_LOG_FILE`, `NBI_ACCESS_LOG_FILE` — hanya ini yang resmi terdaftar).
- **Kesimpulan yang jujur**: `NBI_AUTHENTICATION_KEY`/`x-api-key` **bukan bagian dari dokumentasi resmi GenieACS per saat analisis ini dibuat**, statusnya adalah kontribusi komunitas yang mungkin sudah/belum di-merge ke branch tertentu tergantung versi build yang dipakai. **Jangan asumsikan fitur ini otomatis ada** di instalasi GenieACS Anda — harus dicek langsung ke changelog/source code versi yang benar-benar terpasang (`bin/genieacs-nbi` di server Anda), bukan dipercaya begitu saja dari dokumentasi pihak ketiga.

**Implikasi langsung untuk implementasi MikMongo/mikhmon-api Anda:**
- Karena NBI tidak native ber-auth (atau minimal, auth-nya tidak terjamin ada tergantung versi), **NBI TIDAK BOLEH diekspos langsung ke internet atau ke frontend Anda**.
- Pola aman yang konsisten dengan rekomendasi resmi maintainer: NBI (port 7557) hanya boleh diakses oleh backend Go Anda sendiri (localhost/private network/firewall-restricted), lalu backend Andalah yang menyediakan lapisan auth (JWT/session) ke frontend/user akhir. Backend Go bertindak sebagai proxy tepercaya ke NBI — persis pola yang disebut salah satu user forum ("wrote a little daemon on the same machine that pulls tasks from our private API").
- Kalau versi GenieACS Anda memang mendukung `NBI_AUTHENTICATION_KEY`, itu bisa jadi **lapisan tambahan** (defense in depth), bukan pengganti firewall/reverse proxy.

### 5.3 Lapis 3 — Login ke Web GUI (`genieacs-ui`, port 3000)

Ini autentikasi untuk **pengguna manusia** yang login ke dashboard admin GenieACS, beda lagi dari NBI (Lapis 2) yang dipakai untuk integrasi mesin-ke-mesin.

- **Terdokumentasi resmi (dari halaman Environment Variables)**: ada variabel `UI_JWT_SECRET` — deskripsi resminya: "kunci yang dipakai untuk menandatangani (sign) token JWT yang disimpan di cookie browser", panjang string bisa sampai 64 karakter, default: unset.
- Ini mengonfirmasi bahwa GUI **memang** pakai JWT (disimpan sebagai cookie), tapi dokumentasi tidak menjelaskan detail flow login/endpoint auth-nya secara eksplisit di halaman mana pun yang saya temukan — sejalan dengan status "TODO" pada halaman `Roles and Permissions`.
- Halaman `Roles and Permissions` (yang harusnya menjelaskan role/permission user GUI) **statusnya TODO / kosong** pada dokumentasi resmi versi saat ini — artinya dokumentasi resmi belum menjelaskan secara tertulis bagaimana sistem role-based access control di GUI bekerja, walau secara praktik UI-nya sendiri (dari referensi wiki lama `config/roles.yml`, `config/users.yml`) memang punya konsep role — ini bukan bagian dari NBI REST API yang jadi topik inti Anda, jadi tidak saya bahas lebih dalam kecuali diminta terpisah.

---

## 6. Ringkasan Autentikasi per Komponen (Tabel Perbandingan)

| Komponen | Port | Auth terdokumentasi resmi? | Mekanisme | Status |
|---|---|---|---|---|
| `genieacs-cwmp` (CPE↔ACS) | 7547 | ✅ Ya | `cwmp.auth` config + `AUTH()`/`EXT()` function | Terdokumentasi, arah CPE→ACS. Arah ACS→CPE: TODO |
| `genieacs-nbi` (REST API) | 7557 | ❌ **Tidak ada** | — | Halaman HTTPS & Roles/Permissions = TODO. Mitigasi = firewall/reverse proxy (rekomendasi komunitas, bukan fitur GenieACS) |
| `genieacs-nbi` + `x-api-key` | 7557 | ⚠️ Tidak terkonfirmasi di docs resmi | Header `x-api-key` + env `NBI_AUTHENTICATION_KEY` | Berasal dari community PR #374, **cek versi terpasang sebelum mengandalkan ini** |
| `genieacs-ui` (Web GUI) | 3000 | 🟡 Sebagian | JWT via cookie, ditandatangani `UI_JWT_SECRET` | Mekanisme signing terkonfirmasi, detail flow & role/permission: TODO |
| Transport (TLS) semua komponen | semua | ✅ Ya | `*_SSL_CERT` + `*_SSL_KEY` per komponen (`CWMP_SSL_CERT/KEY`, `NBI_SSL_CERT/KEY`, `FS_SSL_CERT/KEY`, `UI_SSL_CERT/KEY`) | Terdokumentasi resmi di halaman Environment Variables — kalau path cert/key diisi → HTTPS aktif, kalau kosong → HTTP biasa |

---

## 7. Environment Variables — Relevan untuk Setup NBI

Diambil langsung dari halaman resmi `Environment Variables` (semua wajib prefix `GENIEACS_`):

| Variable | Fungsi | Default |
|---|---|---|
| `NBI_WORKER_PROCESSES` | Jumlah worker process untuk `genieacs-nbi`. 0 = sebanyak core CPU | `0` |
| `NBI_PORT` | Port TCP yang didengarkan `genieacs-nbi` | `7557` |
| `NBI_INTERFACE` | Network interface yang di-bind | `::` (semua interface) |
| `NBI_SSL_CERT` | Path file sertifikat TLS. Jika kosong → HTTP non-secure | unset |
| `NBI_SSL_KEY` | Path file key sertifikat TLS | unset |
| `NBI_LOG_FILE` | File log event proses NBI. Jika kosong → ke stderr | unset |
| `NBI_ACCESS_LOG_FILE` | File log request masuk ke NBI. Jika kosong → ke stdout | unset |
| `MONGODB_CONNECTION_URL` | Connection string MongoDB (dipakai semua komponen, termasuk NBI karena NBI baca/tulis langsung ke DB yang sama) | `mongodb://127.0.0.1/genieacs` |
| `LOG_FORMAT` | Format log (`simple`/`json`) untuk semua `*_LOG_FILE` termasuk NBI | `simple` |
| `ACCESS_LOG_FORMAT` | Format access log (`simple`/`json`) untuk semua `*_ACCESS_LOG_FILE` | `simple` |

**Tidak ada** `NBI_AUTHENTICATION_KEY` dalam daftar resmi ini (lihat pembahasan §5.2).

---

## 8. Catatan Implementasi Praktis

Poin-poin ini murni turunan langsung dari perilaku yang didokumentasikan resmi, disusun sebagai catatan implementasi:

1. **URL encoding wajib** — dokumentasi resmi secara eksplisit memberi warning: kegagalan paling umum adalah tidak meng-encode karakter khusus di `device_id` atau `query` pada URL. Device ID sering mengandung karakter seperti `-`, `%2D` (encoded dash), spasi, dsb.

2. **Pola 200 vs 202 wajib ditangani** — client (backend Go Anda) harus membedakan dua alur:
   - `200`: task selesai seketika, hasil (untuk `getParameterValues`) sudah tersimpan di DB, bisa langsung di-query.
   - `202`: task masih di antrean, harus polling `GET /tasks/?query={"_id":"<task_id>"}` atau tunggu inform berikutnya untuk tahu kapan selesai/gagal.

3. **`refreshObject` dengan `objectName: ""`** — cara resmi untuk refresh *seluruh* parameter device, dipakai di contoh dokumentasi untuk kasus "refresh all device parameters now."

4. **File dan Provision pakai body mentah, bukan JSON** — dua endpoint ini beda pola dari yang lain:
   - `PUT /files/<file_name>` → body = binary file, metadata = HTTP header.
   - `PUT /provisions/<provision_name>` → body = teks kode JavaScript mentah.
   
   Kalau backend Go Anda membungkus semua request NBI dengan asumsi "semua body itu JSON", dua endpoint ini akan butuh jalur berbeda (raw bytes / plain text, bukan `json.Marshal`).

5. **Delete device ≠ blokir device** — `DELETE /devices/<device_id>` hanya menghapus record dari database. Kalau device masih hidup dan melakukan periodic inform, ia akan otomatis terdaftar lagi. Untuk benar-benar mencegah device terdaftar, perlu mekanisme di luar NBI (mis. blokir di level CWMP/firewall).

6. **NBI baca-tulis langsung ke MongoDB yang sama dengan komponen lain** — karena `MONGODB_CONNECTION_URL` dipakai bersama oleh `cwmp`, `nbi`, dan `ui`, task yang di-enqueue lewat NBI akan langsung terlihat/dieksekusi oleh proses `cwmp` pada inform berikutnya tanpa perlu komunikasi langsung antar proses NBI↔CWMP — keduanya "bertemu" lewat database, bukan lewat API call satu sama lain.

7. **Tidak ada endpoint bulk/batch resmi** — tidak ada `POST /devices/tasks` (jamak) untuk kirim task yang sama ke banyak device sekaligus dalam satu call. Untuk broadcast ke banyak device, client harus loop `GET /devices/?query=...` lalu `POST /devices/<device_id>/tasks` satu per satu per device — ini penting untuk desain retry/rate-limiting di sisi backend Go Anda kalau mau kirim task massal (mis. ke seluruh pelanggan MikMongo).

---

## 9. Yang Belum/Tidak Dijelaskan Resmi (Gap di Dokumentasi)

Supaya tidak ada kesan "semua terjawab" padahal sebagian memang belum ditulis GenieACS sendiri:

- Halaman **HTTPS** (transport security untuk semua komponen) → isi halaman: **TODO**.
- Halaman **Roles and Permissions** (RBAC untuk siapa boleh akses apa) → isi halaman: **TODO**.
- Bagian **"ACS to CPE"** authentication (kredensial yang dipakai GenieACS saat menghubungi CPE) di halaman CPE Authentication → **TODO**.
- Detail endpoint auth/login untuk `genieacs-ui` (bagaimana persis proses dapat JWT cookie itu) → tidak dijelaskan eksplisit di halaman mana pun yang tersedia.
- Rate limiting, CORS bawaan, dan validasi input NBI → tidak disebutkan di dokumentasi resmi sama sekali (ada thread forum terpisah soal masalah CORS yang menyebutkan solusi manual, bukan fitur bawaan).

Untuk hal-hal di atas, satu-satunya sumber yang valid adalah membaca source code GenieACS versi yang benar-benar Anda pasang (`genieacs-nbi`, `genieacs-ui` di repo GitHub resminya), karena dokumentasi resminya sendiri mengakui belum menulisnya.
