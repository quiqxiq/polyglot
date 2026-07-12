# Issue 10: RX Power / ONU Optical Health Monitoring

## Konteks

Temuan #5 di `ANALISIS-PROVISIONING-REPO-REFERENSI.md` menyoroti bahwa metrik kesehatan optik ONU — **RX power (dBm)**, **jarak fiber (distance)**, dan **temperature** — adalah data **time-series** yang di-sample berulang kali sepanjang waktu. Menyimpannya di Postgres (baris per sampel, per ONU, tiap beberapa menit) akan meledakkan ukuran tabel, mempersulit retensi/downsampling, dan memaksa query agregasi yang bukan kekuatan RDBMS. Karena itu data ini **WAJIB** ditulis ke InfluxDB (time-series DB), bukan Postgres. Postgres hanya boleh menyimpan **konfigurasi threshold** (data kecil, jarang berubah) — itu bukan time-series.

Sumber data ada **dua jalur, keduanya wajib didukung**: (1) **SNMP ke OLT** lewat driver `genericsnmp` yang dibangun di Issue 07 — OLT mengekspos `rx_power_table` dan `distance_table` per ONU; (2) **GenieACS/TR-069** untuk ONU yang report parameter optik lewat TR-069 (jalur ini SAH untuk polling menurut ADR 0005). Kedua jalur memberi RX power untuk populasi ONU yang berbeda (OLT melihat semua ONU tergabung; GenieACS hanya CPE yang inform TR-069), jadi bukan pilihan salah-satu — usecase menggabung keduanya. Karena optik dan TR-069 tidak mendukung push, **job polling terjadwal itu SAH** di sini — ini bukan pelanggaran ADR 0003 (yang hanya melarang polling untuk protokol yang mendukung push, seperti RouterOS `print follow`).

**Decode RX/TX power BUKAN divider tetap `÷100`** (koreksi field-tested, lihat K13 di README §Konvensi Bersama dan bukti di REFERENCES.md §D). Nilai mentah SNMP/TR-069 antar-merk memakai skala berbeda (raw dBm, per-sepuluh, per-seratus) dan sering unsigned padahal semestinya negatif. Decode karena itu **fungsi per-brand** yang: auto-scale by magnitude (>500 → ÷100, >50 → ÷10), interpretasi signed-16 untuk field unsigned, buang sentinel `0`/`65535` sebagai "tidak ada bacaan", dan pilih skala yang menghasilkan dBm **plausibel** (~ -60..+10). Katalog vendor cukup menyimpan `scale_hint` (`raw`|`tenth_dbm`|`centi_dbm`) sebagai petunjuk, bukan pembagi mutlak. Tidak ada aritmetika skala di `usecase/` atau `domain/`.

Nilai RX power yang turun menandakan degradasi link fiber (bend, konektor kotor, jarak berlebih, splitter bermasalah). Sistem harus mengevaluasi tiap sampel terhadap **threshold** (mis. warning -25 dBm, critical -27 dBm) dan menghasilkan **alert** yang bisa dibaca operator lewat endpoint. Notifikasi eksternal (WhatsApp/Telegram) berada di luar cakupan issue ini. Referensi skema terkait: `DATABASE-SCHEMA.md` §6.1 (`subscriptions`, khususnya `onu_serial_number`, `device_id`, dan kolom ONU yang ditambah Issue 08), §3.3 (`odps` untuk relasi `pon_port`).

## Prasyarat

- **Issue 07 (SNMP OLT / driver `genericsnmp`)** — menyediakan pembacaan `rx_power_table` dan `distance_table` per ONU dari OLT via SNMP. Ini sumber data utama untuk ONU yang tidak report TR-069.
- **Issue 08 (data ONU / integrasi GenieACS)** — menyediakan kolom identitas ONU pada `subscriptions` (`onu_pon_port`, `onu_id`, `genieacs_device_id`) dan jalur baca parameter TR-069 lewat driver `genieacs`. Tanpa ini, tag metric (pon_port, ACS device id) dan jalur baca TR-069 tidak lengkap.
- **Foundation yang dipakai:**
  - `internal/registry/registry.go` — untuk mengambil `port.DeviceDriver` per OLT device (SNMP) dan device GenieACS.
  - Domain `command` (`command.Command`, `command.Operation`, `Classify/Translate`) — polling optik adalah operasi **read-only**, memakai `command.OpGetStatus`/operasi baca yang diklasifikasikan `ClassReadOnly` sehingga tidak butuh HITL.
  - `internal/config/config.go` — untuk memuat konfigurasi InfluxDB (URL, token, org, bucket) dan default threshold.
- **Blocker potensial (lihat Open Question di bawah):** ketersediaan instance InfluxDB. Jika belum ada infrastruktur InfluxDB yang disepakati, bagian penulisan metric diblok sampai keputusan infra diambil.

## Ruang Lingkup

**In scope:**
- Kontrak `port.MetricSink` untuk menulis titik metric time-series.
- Adapter InfluxDB di `internal/adapter/influx/` yang mengimplementasikan `port.MetricSink`.
- Usecase `PollOpticalHealth` di `internal/usecase/network/` — iterasi OLT/ONU aktif, baca rx power + distance (+ temperature bila tersedia), tulis ke `MetricSink` dengan tag lengkap, evaluasi threshold, hasilkan alert.
- Domain kecil untuk representasi sampel optik + hasil evaluasi threshold (nilai + status warning/critical).
- Konfigurasi threshold: default dari config/env, opsional dapat di-override per plan/OLT lewat tabel kecil `optical_thresholds` (Postgres, **non** time-series).
- Endpoint REST: baca kesehatan optik terkini + tren singkat per subscription, daftar alert, dan pemicu polling manual.
- Job scheduler untuk memanggil `PollOpticalHealth` secara berkala.

**Out of scope:**
- Menyimpan time-series di Postgres (dilarang keras).
- Notifikasi eksternal (WhatsApp/Telegram/email) — hanya expose lewat endpoint.
- Auto-remediation (mis. reboot ONU otomatis saat RX drop).
- Dashboard/grafik front-end (cukup ekspos data JSON; visualisasi konsumen API).
- Downsampling/retention policy lanjutan InfluxDB (cukup tulis raw; retensi diatur di sisi InfluxDB oleh ops).

## REST API

Base path: `/api/v1/`. Semua endpoint di bawah proteksi middleware auth + RBAC (K3). Aksi baca tidak mengubah perangkat sehingga status sukses 200; pemicu polling ke banyak perangkat mengikuti konvensi aksi-ke-perangkat = **202 Accepted**.

| Method | Path | Tujuan | Role minimum |
|---|---|---|---|
| GET | `/api/v1/subscriptions/:id/optical-health` | Bacaan optik terkini + tren singkat satu subscription | staff |
| GET | `/api/v1/optical-health/alerts` | Daftar ONU yang saat ini di bawah threshold | teknisi |
| POST | `/api/v1/optical-health/poll` | Picu polling optik manual (seluruh/subset OLT-ONU) | admin |

**GET `/api/v1/subscriptions/:id/optical-health`**
- **Request:** path param `id` (subscription id, UUID). Query opsional: `window` (rentang tren, mis. `1h`/`24h`/`7d`, default `24h`), `points` (jumlah titik maksimum untuk tren singkat, default kecil mis. 50).
- **Response (200):** objek berisi identitas ONU (`subscription_id`, `device_id` OLT, `onu_serial_number`, `pon_port`), `latest` (objek: `rx_power_dbm`, `distance_m`, `temperature_c` bila ada, `measured_at`, `status` ∈ `ok`/`warning`/`critical`/`unknown`), dan `trend` (array titik `{measured_at, rx_power_dbm}` hasil query InfluxDB, sudah didownsample sesuai `points`).
- **Gagal:** 404 bila subscription tidak ada atau tidak punya ONU teridentifikasi; 200 dengan `latest.status = "unknown"` dan `trend` kosong bila subscription valid tetapi belum ada data di InfluxDB; 503 bila InfluxDB tidak dapat dihubungi; 403 bila role di bawah staff.

**GET `/api/v1/optical-health/alerts`**
- **Request:** query opsional `severity` (`warning`/`critical`, default semua), `olt_device_id` (filter per OLT), `limit`/`offset` untuk paginasi.
- **Response (200):** array alert `{subscription_id, device_id, onu_serial_number, pon_port, rx_power_dbm, threshold_dbm, severity, measured_at}`, diurutkan dari yang paling parah (rx power terendah relatif ke threshold) lebih dulu. Sertakan `total` untuk paginasi.
- **Catatan sumber:** alert diturunkan dari sampel terbaru per ONU. Implementasi boleh membaca "latest per ONU" dari InfluxDB, atau dari cache/tabel ringkasan yang diperbarui `PollOpticalHealth` (lihat Task 6). Tidak ada tabel time-series Postgres.
- **Gagal:** 503 bila sumber data tidak tersedia; 403 bila role di bawah teknisi.

**POST `/api/v1/optical-health/poll`**
- **Request (body opsional):** `olt_device_id` (batasi polling ke satu OLT), atau `subscription_id` (batasi ke satu ONU). Body kosong = poll semua OLT/ONU aktif.
- **Response (202):** objek `{job_id, scheduled_targets}` — polling dijalankan asinkron; `job_id` untuk korelasi log. Karena ini operasi baca (read-only) massal, tidak menulis `provisioning_sync_log`; namun tiap pembacaan perangkat tetap tercatat di `command_audit_log` sesuai K4 (classification `read_only`, decision `auto_approved`, source `rest_api`).
- **Gagal:** 400 bila `olt_device_id`/`subscription_id` tidak valid; 403 bila role di bawah admin; 409 opsional bila polling manual sedang berjalan dan implementasi memilih menolak overlap.

## Tasks

**Task 1: Kontrak `port.MetricSink`**
**Description:** Definisikan interface untuk menulis satu titik metric time-series (measurement, tags, fields, timestamp) sehingga usecase tidak bergantung pada implementasi InfluxDB konkret.
**Acceptance criteria:**
- [ ] File `internal/port/metric_sink.go` berisi interface `MetricSink` dengan satu method tulis (mis. `Write(ctx, point)`), `context.Context` sebagai parameter pertama, `error` sebagai return terakhir.
- [ ] Tipe titik metric didefinisikan sebagai struct netral (bukan tipe library Influx) — memuat nama measurement, map tags (string→string), map fields (nama→nilai numerik/float), dan waktu ukur.
- [ ] Interface punya doc comment diawali nama identifier (`MetricSink ...`) sesuai AGENTS.md §7.
- [ ] Tidak ada import library eksternal Influx di file port ini.
**Files likely touched:** `internal/port/metric_sink.go`.
**Dependencies:** —
**Estimated scope:** Small

---

**Task 2: Adapter InfluxDB (`internal/adapter/influx/`)**
**Description:** Implementasikan `port.MetricSink` menggunakan client InfluxDB (write) dan sediakan juga fungsi query untuk membaca "latest per ONU" dan "tren singkat" yang dipakai handler.
**Acceptance criteria:**
- [ ] File `internal/adapter/influx/client.go` membuka koneksi ke InfluxDB dari config (URL, token, org, bucket) dan mengimplementasikan `port.MetricSink` (tulis batch/single point).
- [ ] Ada compile-time assertion bahwa tipe adapter memenuhi `port.MetricSink`.
- [ ] File query terpisah (mis. `internal/adapter/influx/query.go`) menyediakan fungsi baca tren (rentang waktu + downsample ke N titik) dan baca sampel terbaru per ONU untuk kebutuhan alerts.
- [ ] Kegagalan koneksi/timeout dibungkus dengan `%w` + konteks operasi; tidak ada error dibuang tanpa komentar.
- [ ] Nama measurement dan skema tag disepakati dan didokumentasikan singkat di doc comment (measurement mis. `onu_optical`; tags: `subscription_id`, `device_id`, `onu_serial`, `pon_port`, `pppoe_username`; fields: `rx_power_dbm`, `distance_m`, `temperature_c`). Tag `subscription_id` adalah kunci indeks utama (bukan `genieacs_device_id`) supaya metrik dari kedua sumber (SNMP OLT & TR-069 ONU) menyatu pada satu subscription; lihat Task 5 untuk resolusinya.
- [ ] Tidak ada tabel Postgres time-series dibuat di task ini.
**Files likely touched:** `internal/adapter/influx/client.go`, `internal/adapter/influx/query.go`, `internal/config/config.go` (menambah bagian konfigurasi InfluxDB).
**Dependencies:** Task 1. **Blok oleh Open Question ketersediaan InfluxDB.**
**Estimated scope:** Medium

---

**Task 3: Domain sampel optik + evaluasi threshold**
**Description:** Definisikan representasi domain untuk sampel optik satu ONU dan logika murni evaluasi threshold (rx power → status ok/warning/critical) tanpa I/O.
**Acceptance criteria:**
- [ ] File domain (mis. `internal/domain/optical/optical.go`) berisi struct sampel (`rx_power_dbm`, `distance_m`, `temperature_c` opsional, identitas ONU, waktu ukur) dan tipe status/severity (enum `ok`/`warning`/`critical`/`unknown`).
- [ ] Fungsi murni evaluasi threshold menerima nilai rx power + konfigurasi threshold (warning, critical) dan mengembalikan severity; ambang critical lebih rendah (lebih negatif) daripada warning; nilai tidak diketahui → `unknown`.
- [ ] Tidak ada import library eksternal / I/O di package domain ini (boundary §0).
- [ ] Test table-driven untuk fungsi evaluasi mencakup: di atas warning (ok), tepat di ambang, antara warning-critical, di bawah critical, nilai kosong/unknown.
**Files likely touched:** `internal/domain/optical/optical.go`, `internal/domain/optical/optical_test.go`.
**Dependencies:** —
**Estimated scope:** Small

---

**Task 4: Operasi baca optik di driver sumber data**
**Description:** Pastikan driver OLT SNMP (`genericsnmp`, Issue 07) dan driver GenieACS (Issue 08) dapat menerjemahkan operasi abstrak "baca optik" menjadi pembacaan native (SNMP `rx_power_table`/`distance_table`; parameter TR-069 optik berjenjang), decode ke dBm per-brand, dan mengklasifikasikannya read-only. RX power punya **dua sumber** — SNMP OLT DAN TR-069 ONU — keduanya diimplementasikan di sini.
**Acceptance criteria:**
- [ ] `Translate` di `commands.go` driver terkait memetakan operasi baca optik (mis. `command.OpGetStatus` atau operasi baca khusus) ke command native yang mengambil rx power + distance (+ temperature bila didukung).
- [ ] `Classify` mengembalikan `command.ClassReadOnly` untuk operasi ini (tidak pernah destruktif → tidak memicu HITL).
- [ ] **Decode RX/TX power dilakukan per-brand di driver, BUKAN divider tetap `÷100`** (K13, REFERENCES.md §D): auto-scale by magnitude (>500 → ÷100, >50 → ÷10), interpretasi signed-16 untuk field unsigned, buang sentinel `0`/`65535`, dan pilih skala yang menghasilkan dBm plausibel (~ -60..+10). Katalog vendor menyimpan `scale_hint` (`raw`|`tenth_dbm`|`centi_dbm`) sebagai petunjuk, bukan pembagi mutlak. Tidak ada aritmetika skala di `usecase/`/`domain/` (AGENTS.md §1).
- [ ] **Path TR-069 RX power adalah katalog berjenjang per-vendor (multi-path shotgun), sebagai DATA di driver `genieacs`** (K13, REFERENCES.md §C): satu operasi "baca RX ONU" mencoba beberapa path native berurutan dan memakai yang pertama merespons — minimal mencakup `VirtualParameters.RXPower`/`redaman`, `InternetGatewayDevice...WANPONInterfaceConfig.RXPower`, `InternetGatewayDevice...WANPPPConnection.1.X_ALU-COM_RxPower`, `Device.XPON.Interface.1.RxPower`, dan `Device.Optical.Interface.1.RxPower`. Daftar path disimpan sebagai katalog data driver (bukan `if` per-vendor di kode), sejalan dengan multi-path WiFi/WAN di Issue 09.
- [ ] ONU tanpa data optik (tidak report / index tak ada / semua path TR-069 kosong) menghasilkan hasil "unknown", bukan error fatal yang menghentikan seluruh polling.
- [ ] Pengetahuan tabel/OID SNMP, `scale_hint`, dan daftar path TR-069 tidak bocor ke `usecase/` atau `domain/`.
**Files likely touched:** `internal/driver/genericsnmp/commands.go`, `internal/driver/genericsnmp/driver.go`, `internal/driver/genieacs/commands.go`, `internal/driver/genieacs/client.go`.
**Dependencies:** Issue 07, Issue 08.
**Estimated scope:** Medium

---

**Task 5: Usecase `PollOpticalHealth`**
**Description:** Orkestrasi polling: untuk tiap OLT/ONU aktif, ambil driver dari registry, baca rx power/distance/temperature, tulis ke `MetricSink` dengan tag lengkap, evaluasi threshold, dan kumpulkan alert.
**Acceptance criteria:**
- [ ] File `internal/usecase/network/poll_optical_health.go` berisi fungsi `PollOpticalHealth(ctx, params)` mengikuti pola `<verb>_<noun>` dan penamaan §1.4.
- [ ] Usecase hanya bergantung pada port (`MetricSink`, repository subscription/ONU, registry driver, penyedia threshold) — tidak mengimpor adapter/driver konkret.
- [ ] Iterasi memakai `errgroup` dengan `ctx` yang diteruskan dan batas konkurensi wajar; kegagalan satu ONU tidak membatalkan seluruh batch (kumpulkan error per-target, jangan fail-fast total) — sesuaikan dengan §5.
- [ ] Tiap titik ditulis ke `MetricSink` dengan tag: `subscription_id` (kunci indeks utama), `device_id` (OLT), `onu_serial`, `pon_port`, dan `pppoe_username` bila tersedia.
- [ ] **Resolusi indeks metrik ke subscription/PPPoE username, bukan hanya `genieacs_device_id`:** sampel dari SNMP OLT ditautkan ke subscription via `onu_serial`/`pon_port`+`onu_id`; sampel dari TR-069 ditautkan via `genieacs_device_id` lalu dipetakan ke subscription lewat tag `pppoe:` pada ACS device dan/atau `WANPPPConnection.Username` (bandingkan dengan PPPoE username subscription, lihat strategi linking Issue 08 / REFERENCES.md §C). Sampel yang tidak dapat ditautkan ke subscription mana pun di-log dan dilewati (bukan error fatal), agar dua sumber data menyatu pada `subscription_id` yang sama.
- [ ] Evaluasi threshold memakai fungsi murni domain (Task 3); alert (severity warning/critical) dikumpulkan dan dikembalikan/di-persist untuk endpoint alerts.
- [ ] Tiap pembacaan perangkat tercatat di `command_audit_log` (classification `read_only`, decision `auto_approved`, actor/source sesuai pemicu: `scheduled_job` untuk scheduler, `rest_api` untuk poll manual) sesuai K4.
- [ ] Test table-driven memakai fake `MetricSink` + fake driver: verifikasi jumlah titik yang ditulis, tag benar, dan alert dihasilkan hanya ketika di bawah threshold.
**Files likely touched:** `internal/usecase/network/poll_optical_health.go`, `internal/usecase/network/poll_optical_health_test.go`, `internal/port/*` (bila perlu port penyedia threshold / repo daftar ONU aktif).
**Dependencies:** Task 1, Task 3, Task 4.
**Estimated scope:** Large

---

**Task 6: Penyimpanan/ekspos alert terkini**
**Description:** Sediakan cara membaca "latest per ONU" untuk endpoint alerts. Pilih salah satu strategi dan dokumentasikan: (a) query "last per ONU" langsung dari InfluxDB, atau (b) simpan ringkasan status terbaru per ONU di tabel kecil Postgres yang diperbarui tiap polling.
**Acceptance criteria:**
- [ ] Keputusan strategi (a) atau (b) dinyatakan eksplisit di deskripsi PR; jika (b), tabel ringkasan **bukan** time-series (satu baris per ONU, di-upsert) dan bukan penyimpanan histori.
- [ ] Ada fungsi/port yang mengembalikan daftar ONU dengan severity `warning`/`critical` beserta nilai rx power, threshold, dan waktu ukur, terurut paling parah dulu.
- [ ] Filter `severity` dan `olt_device_id` didukung di lapisan baca.
- [ ] Tidak menduplikasi data time-series ke Postgres di luar satu-baris-ringkasan-per-ONU (bila strategi b dipilih).
**Files likely touched:** `internal/adapter/influx/query.go` (strategi a), atau `internal/adapter/postgres/optical_status_repository.go` + `internal/port/optical_status_repository.go` (strategi b).
**Dependencies:** Task 2, Task 5.
**Estimated scope:** Medium

---

**Task 7: Konfigurasi threshold**
**Description:** Muat threshold default (warning/critical dBm) dari config/env, dengan opsi override per OLT/plan lewat tabel kecil `optical_thresholds` (non time-series). Jika override tidak ada, pakai default.
**Acceptance criteria:**
- [ ] Default warning/critical terbaca dari config (mis. `OPTICAL_RX_WARNING_DBM`, `OPTICAL_RX_CRITICAL_DBM`) dengan nilai contoh -25 dan -27.
- [ ] Ada port + repo untuk membaca threshold override (bila tabel `optical_thresholds` diadakan di Task 8); resolusi threshold: override spesifik > default global.
- [ ] Usecase menerima threshold lewat penyedia (tidak hardcode di usecase) sehingga tetap mudah diuji.
- [ ] Validasi: critical harus lebih negatif dari warning; konfigurasi tak valid ditolak saat startup dengan pesan jelas.
**Files likely touched:** `internal/config/config.go`, `internal/port/optical_threshold_repository.go` (opsional), `internal/adapter/postgres/optical_threshold_repository.go` (opsional).
**Dependencies:** Task 3.
**Estimated scope:** Small

---

**Task 8: Migrasi `optical_thresholds` (opsional)**
**Description:** Bila override per OLT/plan dibutuhkan, buat tabel kecil `optical_thresholds` di Postgres (konfigurasi, bukan time-series). Bila threshold cukup dari env saja, task ini dilewati dan dinyatakan di PR.
**Acceptance criteria:**
- [ ] Bila diadakan: migrasi `000030_create_optical_thresholds_table.up.sql` + `.down.sql` berpasangan, lanjut dari 000021 (K6).
- [ ] Kolom sesuai bagian "Migrasi Database" di bawah.
- [ ] `DATABASE-SCHEMA.md` diperbarui pada PR yang sama.
- [ ] Bila tidak diadakan: PR menyatakan eksplisit "tidak ada perubahan skema" dan threshold hanya dari env.
**Files likely touched:** `migrations/000030_create_optical_thresholds_table.up.sql`, `migrations/000030_create_optical_thresholds_table.down.sql`, `DATABASE-SCHEMA.md`.
**Dependencies:** Task 7.
**Estimated scope:** Small

---

**Task 9: Handler REST + DTO**
**Description:** Implementasikan tiga endpoint (optical-health per subscription, daftar alerts, poll manual) sebagai handler HTTP tipis yang memanggil usecase/lapisan baca; tidak memanggil `port.DeviceDriver` langsung (K4).
**Acceptance criteria:**
- [ ] Handler baru di `internal/adapter/http/` (mis. `optical_health_handler.go`) beserta DTO request/response di `internal/adapter/http/dto/`.
- [ ] Route didaftarkan di `internal/adapter/http/router.go` dengan RBAC: GET per-subscription = staff+, GET alerts = teknisi+, POST poll = admin+ (K3).
- [ ] GET per-subscription mengembalikan `latest` + `trend` dari InfluxDB; menangani kasus "belum ada data" (status `unknown`, trend kosong) dan 404 subscription/ONU tak ada.
- [ ] GET alerts mendukung filter `severity`/`olt_device_id` + paginasi, terurut paling parah dulu.
- [ ] POST poll menjalankan `PollOpticalHealth` asinkron, mengembalikan **202** + `job_id`; validasi `olt_device_id`/`subscription_id`.
- [ ] Test handler memakai `httptest` untuk tiap endpoint (sukses, tak ada data, forbidden role, target invalid).
**Files likely touched:** `internal/adapter/http/optical_health_handler.go`, `internal/adapter/http/dto/optical_health.go`, `internal/adapter/http/router.go`, `configs/rbac_policy.csv`.
**Dependencies:** Task 5, Task 6.
**Estimated scope:** Medium

---

**Task 10: Scheduler polling berkala**
**Description:** Jadwalkan pemanggilan `PollOpticalHealth` secara periodik (mis. tiap 5–15 menit, dari config) untuk seluruh OLT/ONU aktif. Polling terjadwal ini SAH untuk optik/TR-069 (bukan pelanggaran ADR 0003).
**Acceptance criteria:**
- [ ] Scheduler dijalankan dari wiring startup (`cmd/server/main.go`) memakai `ctx` aplikasi; berhenti bersih saat shutdown (context cancel), tidak fire-and-forget (§5).
- [ ] Interval polling terbaca dari config (mis. `OPTICAL_POLL_INTERVAL`); nilai tak valid ditolak saat startup.
- [ ] Pemanggilan terjadwal memakai actor_type `system_scheduled` / source `scheduled_job` pada audit log.
- [ ] Overlap run dicegah (skip bila run sebelumnya belum selesai) agar OLT tidak dibanjiri SNMP.
- [ ] Kegagalan satu siklus polling di-log dan tidak menghentikan scheduler.
**Files likely touched:** `cmd/server/main.go`, (bila ada paket scheduler internal, mis. `internal/usecase/network/poll_optical_health.go` menyediakan fungsi yang dipanggil scheduler).
**Dependencies:** Task 5.
**Estimated scope:** Medium

---

**Task 11: ADR keputusan InfluxDB untuk time-series optik**
**Description:** Tulis ADR yang mendokumentasikan keputusan menyimpan metrik optik di InfluxDB (bukan Postgres), pemilihan measurement/tag, dan alasan polling terjadwal SAH di sini.
**Acceptance criteria:**
- [ ] ADR baru `docs/adr/0009-influxdb-untuk-metrik-optik-timeseries.md` (lanjut dari 0005).
- [ ] Menjelaskan alasan tidak memakai Postgres untuk time-series, skema measurement/tag/field, dan hubungannya dengan ADR 0003 (no-polling) & ADR 0005 (TR-069 polling SAH).
- [ ] Ditautkan dari `README.md` root pada commit yang sama (AGENTS.md §1.5).
**Files likely touched:** `docs/adr/0009-influxdb-untuk-metrik-optik-timeseries.md`, `README.md`.
**Dependencies:** Task 2.
**Estimated scope:** Small

---

## Migrasi Database

**Tabel time-series → TIDAK ADA migrasi Postgres.** RX power, distance, dan temperature disimpan di InfluxDB (measurement `onu_optical`), bukan Postgres. Jangan membuat tabel Postgres untuk data time-series ini.

**Opsional — `optical_thresholds` (konfigurasi, bukan time-series):** hanya bila override threshold per OLT/plan diputuskan diperlukan (Task 7/Task 8). Jika threshold cukup dari env, **tidak ada perubahan skema** dan hal ini dinyatakan eksplisit di PR.

Bila diadakan:
- Nomor migrasi: lanjut dari 000021 → **000030**.
- Nama file: `migrations/000030_create_optical_thresholds_table.up.sql` dan pasangan `.down.sql`.
- Kolom yang diperlukan (dijelaskan sebagai teks, bukan SQL):
  - `id` — primary key (UUID).
  - `scope_type` — enum kecil menandai lingkup override: `global`, `olt_device`, atau `plan`.
  - `olt_device_id` — nullable, FK ke `devices`, diisi bila `scope_type = olt_device`.
  - `plan_id` — nullable, FK ke `plans`, diisi bila `scope_type = plan`.
  - `warning_dbm` — nilai ambang warning (numeric, satuan dBm, biasanya negatif).
  - `critical_dbm` — nilai ambang critical (numeric, dBm, lebih negatif dari warning).
  - `created_at`, `updated_at` — timestamp audit.
  - Constraint: `critical_dbm < warning_dbm`; keunikan untuk mencegah override ganda pada scope yang sama (mis. UNIQUE parsial per `olt_device_id`, per `plan_id`, dan satu baris `global`).
- `down.sql`: drop tabel `optical_thresholds`.
- **Perbarui `DATABASE-SCHEMA.md`** (tambahkan deskripsi tabel `optical_thresholds` dan tegaskan bahwa metrik optik disimpan di InfluxDB) pada PR yang sama.

## Verification

- [ ] `go build ./...` sukses.
- [ ] `go test ./internal/domain/optical/...` — fungsi evaluasi threshold (table-driven) hijau.
- [ ] `go test ./internal/usecase/network/...` — `PollOpticalHealth` dengan fake `MetricSink` + fake driver: titik tertulis dengan tag benar, alert muncul hanya di bawah threshold, satu ONU gagal tidak menghentikan batch.
- [ ] `go test ./internal/adapter/http/...` — handler `httptest` untuk tiga endpoint (sukses, no-data → `unknown`, role forbidden, target invalid, poll → 202).
- [ ] `go test ./internal/adapter/influx/...` — bila lingkungan CI menyediakan InfluxDB (atau container), verifikasi write + query latest/trend; bila tidak, tandai skip dengan alasan jelas.
- [ ] `go test ./internal/adapter/postgres/...` — bila tabel `optical_thresholds` diadakan: repo pakai `testcontainers-go` (K7), verifikasi resolusi override > default.
- [ ] `make lint` bersih (gofumpt/goimports/staticcheck), termasuk akronim konsisten (`RxPowerDBm`, bukan `RxPowerDbm`).
- [ ] Smoke test manual (sebut sebagai curl): `GET /api/v1/subscriptions/<id>/optical-health` mengembalikan `latest`+`trend`; `GET /api/v1/optical-health/alerts?severity=critical` mengembalikan daftar terurut; `POST /api/v1/optical-health/poll` mengembalikan 202 + `job_id`, lalu data baru terlihat di endpoint pertama setelah polling selesai.
- [ ] Verifikasi RBAC: staff dapat GET per-subscription tetapi ditolak (403) pada `POST /poll`; teknisi dapat GET alerts.

## Definition of Done

- [ ] `port.MetricSink` + adapter InfluxDB terimplementasi; tidak ada tabel Postgres time-series.
- [ ] `PollOpticalHealth` membaca rx power/distance (+ temperature bila ada) via driver SNMP/GenieACS **dari kedua sumber** (OLT SNMP dan TR-069 ONU), menulis ke InfluxDB dengan tag lengkap, dan mengevaluasi threshold.
- [ ] Decode RX/TX power per-brand (auto-scale by magnitude, signed-16, buang sentinel, pilih dBm plausibel) terjadi di driver, bukan usecase/domain; katalog hanya menyimpan `scale_hint` — bukan divider tetap `÷100`.
- [ ] Path TR-069 RX power berjenjang (multi-path shotgun) sebagai data katalog driver `genieacs`; metrik terindeks pada `subscription_id`/PPPoE username, bukan hanya `genieacs_device_id`.
- [ ] Alert dapat dibaca lewat `GET /api/v1/optical-health/alerts`; kesehatan optik per subscription lewat `GET /api/v1/subscriptions/:id/optical-health`; poll manual lewat `POST /api/v1/optical-health/poll` (202).
- [ ] Threshold default dari config; override opsional lewat `optical_thresholds` (bila diadakan) dengan migrasi 000030 + update `DATABASE-SCHEMA.md`.
- [ ] Scheduler polling berkala aktif, berhenti bersih saat shutdown, mencegah overlap.
- [ ] Pembacaan tercatat di `command_audit_log` (read_only, auto_approved) sesuai K4.
- [ ] ADR 0009 ditulis dan ditautkan dari `README.md`.
- [ ] `go build`, unit test terkait, dan `make lint` hijau.

---

> **Open Question (blocker) — Ketersediaan InfluxDB.** Issue ini mengasumsikan ada instance InfluxDB yang dapat dipakai (URL/token/org/bucket). Jika infrastruktur InfluxDB **belum** disepakati/tersedia di lingkungan target, Task 2, Task 6 (strategi a), dan penulisan metric **diblok** sampai keputusan infra diambil. Alternatif sementara yang perlu dikonfirmasi ke owner sebelum eksekusi: (1) sediakan InfluxDB via `deployments/docker-compose.yml`; atau (2) tunda penulisan metric dan hanya jalankan evaluasi threshold in-memory untuk alert. Nyatakan keputusan ini eksplisit di PR sebelum mulai (AGENTS.md §0 poin 3).
