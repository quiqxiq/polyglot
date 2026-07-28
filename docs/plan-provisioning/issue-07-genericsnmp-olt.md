# Issue 07: Driver OLT SNMP Multi-Vendor (genericsnmp)

## Konteks

Analisis pasar RT/RW-Net dan ISP kecil di `ANALISIS-PROVISIONING-REPO-REFERENSI.md` (Temuan #6) menunjukkan bahwa pelanggan Polyglot memakai OLT dari banyak merk sekaligus — ZTE (C300/C320), Huawei (MA5608T/MA5680T), VSOL, Hioso, HSGQ, Fiberhome, BDCOM, dan CDATA — sering bercampur dalam satu jaringan. Untuk keperluan monitoring optik dan discovery ONU, perbedaan antar-merk ini hampir seluruhnya adalah perbedaan **OID SNMP**, bukan perbedaan logika. Nilai yang dibaca (status ONU, serial number/SN, RX power, jarak/distance, daftar ONU belum terdaftar) semuanya bersifat read-only dan konseptual sama; yang beda hanya di batang OID mana angka itu berada dan skala unitnya (mis. RX power dalam satuan 0.01 dBm).

Membuat satu paket Go per merk (`internal/driver/vsol/`, `internal/driver/hioso/`, dst.) akan menghasilkan delapan-plus paket yang isinya 95% identik dan hanya beda tabel OID — persis anti-pattern yang sudah dihindari untuk CLI generik lewat `genericcli.Catalog` di `docs/adr/0004-generic-cli-driver-scrapligo.md`. Karena itu issue ini membuat **satu** driver generik `internal/driver/genericsnmp/` dengan **katalog OID sebagai data**, bukan sebagai kode per-paket. Keputusan arsitektur ini didokumentasikan sebagai ADR baru `0008` yang mengikuti pola ADR `0004`.

**Katalog OID berstruktur per-profile (model), bukan satu set per-brand** (field-tested di 5 repo referensi, sesuai K13 — lihat README §Konvensi Bersama). Satu merk punya banyak profile MIB yang tidak kompatibel: ZTE saja mencakup C300/C600 dan varian OLD, masing-masing dengan batang OID berbeda untuk objek yang secara konseptual sama. Karena itu unit katalog adalah **profile**, bukan merk, dan profile yang tepat untuk satu OLT **dideteksi otomatis via `probe_oid`** (profile yang `probe_oid`-nya merespons walk pertama menang) — inilah yang membuat satu driver melayani banyak model tanpa operator harus memilih brand/model secara manual di config.

Data yang dibaca driver ini menjadi fondasi bagi Issue 05 (scan ONU belum terdaftar untuk auto-discovery) dan Issue 10 (monitoring RX power / jarak sebagai time-series). Issue ini **tidak** menyimpan time-series apa pun — hanya menyediakan operasi baca snapshot. Persistensi metrik ada di Issue 10. OLT SNMP di sini bersifat monitoring read-only; provisioning aktual (buat ONU, ganti profil) tetap lewat jalur telnet/CLI vendor yang sudah ada (`zteolt`, `huaweiolt`).

## Prasyarat

- **Issue 01 (Sync Engine)** — pola sinkronisasi `provisioning_sync_log` dan `usecase/network.ExecuteCommand` sebagai jalur kanonik ke perangkat (dipakai bila operasi SNMP suatu saat menulis; untuk issue ini murni baca).
- **Foundation registry** — `internal/registry/registry.go` yang memegang satu driver per device dan `Get(ctx, deviceID)`; driver `genericsnmp` harus bisa dikembalikan lewat factory registry berdasarkan tipe/merk device.
- **Domain device** — `internal/domain/device/` untuk `device.Target` (host, kredensial SNMP community/v3).
- Driver eksisting `internal/driver/zteolt/` dan `internal/driver/huaweiolt/` sebagai referensi OID yang sudah terbukti (nilai OID dipindahkan ke katalog, bukan ditulis ulang dari nol).
- ADR `0004` sebagai template pola "satu driver + katalog data".

## Ruang Lingkup

**In scope:**
- ADR `0008` yang menjelaskan keputusan satu driver + katalog OID **per-profile** sebagai data (deteksi profile via `probe_oid`), ditautkan dari `README.md` root.
- Paket `internal/driver/genericsnmp/` dengan `driver.go` (koneksi SNMP + implementasi kontrak port read-only) dan `catalog.go` (struktur katalog OID **per-profile**, satu objek data per model).
- Port baru khusus operasi baca optik/ONU (mis. `port.OpticalReader` / `port.ONULister`) di `internal/port/`.
- Katalog OID awal terisi untuk profile ZTE (C300/C600) dan Huawei; merk/profile lain disediakan sebagai slot data kosong yang terkurasi menyusul.
- Operasi: `WalkUnconfiguredONUs`, `ReadRxPower`, `ReadDistance`, `ListONUStatus`.
- **Auto-deteksi profile** via `probe_oid` saat driver dibuka + **SN candidate scanning** dengan cache OID hasil deteksi per-OLT (self-healing OID untuk field tak stabil).
- Integrasi ke registry/factory agar Issue 05 dan Issue 10 memanggil lewat abstraksi.
- Endpoint REST read-only untuk daftar merk/profile yang didukung dan snapshot status ONU per OLT.

**Out of scope:**
- Provisioning ONU (create/delete/change-profile) — tetap di driver telnet/CLI vendor (`zteolt`/`huaweiolt`).
- Penyimpanan time-series RX power/jarak — Issue 10.
- Refactor total `zteolt`/`huaweiolt` menjadi `genericsnmp` — hanya rekomendasi migrasi bertahap, bukan eksekusi penuh di issue ini.
- Tabel katalog OID di database — tidak ada; katalog OID adalah data di kode (`catalog.go`). Satu-satunya penambahan skema di issue ini adalah tabel cache OID hasil deteksi per-OLT (`resolved_oids`, lihat §Migrasi Database) — bukan katalog itu sendiri.

## REST API

Base path `/api/v1/`. Semua endpoint di issue ini read-only; tidak ada aksi menulis ke perangkat, jadi tidak ada respons 202/`sync_log` di sini (konvensi 202 + id `sync_log` berlaku hanya untuk aksi tulis ke perangkat, yang tidak ada dalam issue ini).

| Method | Path | Tujuan | Role minimum |
|---|---|---|---|
| GET | `/api/v1/olt-vendors` | Daftar merk/profile OLT yang didukung + status kelengkapan katalog OID per profile | staff |
| GET | `/api/v1/devices/:id/onu-status` | Snapshot status ONU pada satu OLT via SNMP (opsional) | teknisi |

### GET `/api/v1/olt-vendors`

- **Request:** tidak ada body. Query opsional `only_complete` (boolean) untuk memfilter hanya profile yang katalog OID-nya lengkap.
- **Response 200:** array objek **profile** (bukan sekadar merk), tiap objek berisi: `vendor` (kode merk, mis. `zte`, `huawei`), `profile` (kode model/MIB, mis. `C300`, `C600`), `catalog_status` (nilai `complete` / `partial` / `stub`), flag ketersediaan tiap tabel OID: `has_status_table`, `has_name_table`, `has_sn_table`, `has_tx_power_table`, `has_rx_power_table`, `has_distance_table`, `has_unauth_sn_table`, `has_offline_reason_table`, `has_firmware_table`, `has_uptime_table`; plus metadata: `distance_tenths_meter` (bool), `online_values` (himpunan nilai/label yang berarti online), `rx_power_scale_hint`, dan `probe_oid` (batang OID pendeteksi profile). Nilai `rx_power_dbm` di endpoint lain berasal dari decode per-brand, bukan pembagian tetap (lihat Task 4/Issue 10).
- **Gagal:** 500 bila katalog gagal dimuat.

### GET `/api/v1/devices/:id/onu-status` (opsional)

- **Request:** path param `id` = device_id OLT. Query opsional `pon_port` untuk membatasi ke satu port PON.
- **Response 200:** objek berisi `device_id`, `vendor`, `profile` (profile yang terdeteksi via `probe_oid`), `polled_at` (timestamp saat SNMP di-walk), dan `onus` = array per-ONU: `pon_port`, `onu_id`, `serial_number`, `oper_status` (dipetakan dari raw via `online_values` profile — mis. `online`/`offline`/`los`/`unknown`), `rx_power_dbm` (nullable, hasil decode per-brand dari raw, bukan divider tetap), `distance_m` (nullable, memperhitungkan `distance_tenths_meter`). Nilai yang tidak tersedia di katalog profile tersebut dikembalikan null, bukan mengarang.
- **Gagal:** 404 bila device tidak ada; 400 bila device bukan tipe OLT atau vendornya tidak punya katalog SNMP; 502 bila SNMP walk gagal/timeout ke OLT.

## Tasks

**Task 1: ADR 0008 — satu driver genericsnmp + katalog OID per-profile sebagai data**

**Description:** Tulis ADR yang menjelaskan alasan memilih satu driver SNMP generik dengan katalog OID **per-profile (model)** sebagai data, alih-alih satu paket Go per merk, mengikuti pola ADR 0004.

**Acceptance criteria:**
- [ ] File `docs/adr/0008-genericsnmp-olt-katalog-oid.md` dibuat dengan status `Accepted`.
- [ ] ADR menyebut konteks Temuan #6 (banyak merk OLT), keputusan (satu driver + katalog data **per-profile**, deteksi profile via `probe_oid` — walk pertama yang merespons menang), konsekuensi (menambah model = menambah satu objek profile OID, bukan paket baru), dan alternatif yang ditolak (satu paket per merk, dan katalog per-brand tunggal yang gagal karena satu merk punya banyak MIB tak kompatibel spt ZTE C300/C600/OLD).
- [ ] ADR menjelaskan bahwa unit katalog adalah profile (bukan brand), memuat set field per profile (`status_table`, `name_table`, `sn_table`, `tx_power_table`, `rx_power_table`, `distance_table` + `distance_tenths_meter`, `unauth_sn_table`, `unauth_type_table`, `offline_reason_table`, `firmware_table`, `uptime_table`, `probe_oid`, `online_values`), dan merujuk K13 (README §Konvensi Bersama) tanpa menyalin ulang isinya.
- [ ] ADR menyebut hubungan dengan ADR 0004 (pola serupa `genericcli.Catalog`) dan batas: SNMP hanya read-only monitoring, provisioning tetap di driver CLI vendor.
- [ ] Link ke ADR ini ditambahkan di `README.md` root pada commit yang sama.

**Files likely touched:** `docs/adr/0008-genericsnmp-olt-katalog-oid.md`, `README.md`

**Dependencies:** Tidak ada.

**Estimated scope:** Small

---

**Task 2: Definisikan port read-only optik/ONU**

**Description:** Buat kontrak interface baru di `internal/port/` untuk operasi baca SNMP OLT, karena OLT SNMP tidak memenuhi `port.DeviceDriver` penuh (tidak Execute command destruktif).

**Acceptance criteria:**
- [ ] File `internal/port/optical_reader.go` (atau nama serupa) berisi interface `OpticalReader` dan/atau `ONULister` dengan method: `WalkUnconfiguredONUs`, `ReadRxPower`, `ReadDistance`, `ListONUStatus` — semua menerima `context.Context` sebagai parameter pertama dan mengembalikan `error` sebagai return terakhir.
- [ ] Tipe hasil (mis. `ONUStatus`, `UnconfiguredONU`, `OpticalReading`) didefinisikan di domain yang sesuai (`internal/domain/device/` atau domain ONU/optik baru bila diperlukan) — bukan di `port/`.
- [ ] Doc comment tiap identifier exported dimulai dengan nama identifier itu sendiri.
- [ ] Tidak ada import framework/adapter dari `port/`.

**Files likely touched:** `internal/port/optical_reader.go`, `internal/domain/device/` (atau domain optik baru)

**Dependencies:** Task 1 (arah desain).

**Estimated scope:** Medium

---

**Task 3: Struktur katalog OID per-profile (catalog.go)**

**Description:** Definisikan struktur data katalog OID **per-profile (model)** di `internal/driver/genericsnmp/catalog.go`, mengikuti pola `genericcli.Catalog` dan K13 (README §Konvensi Bersama).

**Acceptance criteria:**
- [ ] Unit katalog adalah **profile** (satu objek data per model MIB), bukan satu set per-brand. Tiap profile memuat field OID: `status_table`, `name_table`, `sn_table`, `tx_power_table`, `rx_power_table`, `distance_table`, `unauth_sn_table`, `unauth_type_table`, `offline_reason_table`, `firmware_table`, `uptime_table`, `probe_oid`; plus metadata: `distance_tenths_meter` (bool — jarak dalam satuan 0.1 m atau m), `online_values` (himpunan nilai/label online, numerik **atau** string, varian EPON vs GPON — jangan hardcode `status==1`), dan `scale_hint` untuk RX/TX power (bukan divider tetap — decode aktual per-brand di driver, lihat Task 5/Issue 10).
- [ ] `sn_table` (dan field tak stabil lain) disimpan sebagai **daftar kandidat OID**, bukan satu OID tunggal — mendukung SN candidate scanning di Task 5.
- [ ] Ada registry katalog per-profile (mis. map `vendor/profile -> Profile`) dan fungsi lookup, plus fungsi pemilihan profile berbasis hasil probe (`probe_oid`).
- [ ] Posture fail-safe: katalog zero value / profile tak dikenal membuat operasi baca mengembalikan error eksplisit "profile belum dikurasi", bukan diam-diam mengembalikan data kosong seolah sukses (sesuai fail-safe genericcli).
- [ ] Ada penanda kelengkapan katalog (`complete`/`partial`/`stub`) per profile yang bisa dibaca endpoint `olt-vendors`.

**Files likely touched:** `internal/driver/genericsnmp/catalog.go`

**Dependencies:** Task 2.

**Estimated scope:** Medium

---

**Task 4: Isi katalog OID profile ZTE/Huawei + decode power per-brand**

**Description:** Isi katalog OID untuk profile ZTE (C300/C600) dan Huawei dengan nilai OID yang sudah terbukti (dipindahkan dari driver `zteolt`/`huaweiolt`), dan implementasikan decode RX/TX power sebagai **fungsi per-brand**, bukan divider tetap.

**Acceptance criteria:**
- [ ] Profile ZTE C300, ZTE C600, dan Huawei terisi untuk minimal `status_table`, `sn_table` (daftar kandidat), `rx_power_table`, `unauth_sn_table`, `distance_table` (+ `distance_tenths_meter`), `online_values`, dan `probe_oid` masing-masing.
- [ ] Nilai OID diambil/diverifikasi dari implementasi `zteolt`/`huaweiolt` yang sudah ada, bukan ditulis ulang dari ingatan.
- [ ] Decode RX/TX power = **fungsi per-brand** yang auto-scale by magnitude, menangani signed-16 pada unsigned, dan membuang sentinel (`0`/`65535`) → dBm plausibel (~ -60..+10). Katalog hanya menyimpan `scale_hint`; aritmetika decode tidak ada di usecase (sesuai K13, silang ke Issue 10).
- [ ] Slot data untuk merk/profile lain (VSOL, Hioso, HSGQ, Fiberhome, BDCOM, CDATA) hadir sebagai entri `stub`/`partial` yang jujur menandai OID belum diisi.
- [ ] OID ditulis sebagai konstanta/data string, bukan logic bercabang per merk.

**Files likely touched:** `internal/driver/genericsnmp/catalog.go`

**Dependencies:** Task 3.

**Estimated scope:** Medium

---

**Task 5: Implementasi driver SNMP (driver.go)**

**Description:** Implementasikan koneksi SNMP dan operasi baca di `internal/driver/genericsnmp/driver.go`, memenuhi port dari Task 2 dengan compile-time assertion.

**Acceptance criteria:**
- [ ] `driver.go` membuka koneksi SNMP (v2c community / v3) dari `device.Target`, lalu **mendeteksi profile via `probe_oid`** (walk tiap kandidat profile untuk vendor device; profile pertama yang merespons menang) dan menyimpan profile terpilih pada driver. Profile hasil deteksi di-cache per-OLT (tabel `resolved_oids`, lihat §Migrasi Database) supaya walk deteksi tidak diulang tiap panggilan.
- [ ] Method `WalkUnconfiguredONUs`, `ReadRxPower`, `ReadDistance`, `ListONUStatus` diimplementasikan dengan SNMP walk/get berdasarkan OID profile, menerapkan decode power per-brand (bukan divider tetap), pemetaan status lewat `online_values`, dan `distance_tenths_meter` untuk jarak.
- [ ] **SN candidate scanning (self-healing OID):** untuk field tak stabil (khususnya SN), driver mencoba daftar kandidat OID, memakai heuristik validasi sampel (buang string yang tampak status/offline/power) untuk memilih OID yang benar, lalu meng-cache hasil deteksi per-OLT di `resolved_oids`. Panggilan berikutnya memakai OID hasil cache dulu, re-scan hanya bila cache miss/invalid.
- [ ] Ada compile-time assertion bahwa `*Driver` memenuhi interface port yang relevan.
- [ ] `context.Context` parameter pertama; error di-wrap dengan `%w` + konteks operasi; timeout SNMP dihormati via context.
- [ ] Profile tak dikenal / tanpa katalog lengkap mengembalikan error eksplisit, dan nilai OID yang tak tersedia dikembalikan sebagai null/unset, bukan nol palsu.
- [ ] `Close()` melepas koneksi SNMP (best-effort, dengan komentar bila error diabaikan).

**Files likely touched:** `internal/driver/genericsnmp/driver.go`, `migrations/000035_create_resolved_oids_table.up.sql` + `.down.sql`, `internal/adapter/postgres/` (repo cache `resolved_oids`), `DATABASE-SCHEMA.md`

**Dependencies:** Task 2, Task 3, Task 4.

**Estimated scope:** Large

---

**Task 6: Integrasi registry/factory**

**Description:** Sambungkan `genericsnmp` ke factory registry agar bisa dikembalikan lewat `registry.Get` berdasarkan tipe/merk device OLT, sehingga Issue 05 dan Issue 10 memanggilnya lewat abstraksi port.

**Acceptance criteria:**
- [ ] Factory registry mengenali device tipe OLT dan mengembalikan driver `genericsnmp` (dengan katalog merk yang sesuai) di balik interface port optik.
- [ ] Seleksi merk berasal dari atribut device (vendor/model) dan tidak di-hardcode di pemanggil.
- [ ] Pemanggil (Issue 05/10) hanya bergantung pada port, tidak pada paket `genericsnmp` konkret.

**Files likely touched:** `internal/registry/registry.go`

**Dependencies:** Task 5.

**Estimated scope:** Medium

---

**Task 7: Endpoint GET /api/v1/olt-vendors**

**Description:** Tambahkan handler REST yang mengembalikan daftar merk OLT yang didukung beserta kelengkapan katalog OID, dibaca dari registry katalog `genericsnmp`.

**Acceptance criteria:**
- [ ] Handler `internal/adapter/http/olt_vendor_handler.go` terdaftar di router pada path `/api/v1/olt-vendors`, method GET.
- [ ] Response memuat per merk: kode vendor, model dikenali, `catalog_status`, flag ketersediaan tiap tabel OID, dan skala RX power — sesuai bagian REST API.
- [ ] Query `only_complete` memfilter merk lengkap.
- [ ] DTO response di `internal/adapter/http/dto/`.
- [ ] RBAC: role minimum `staff` (baca) via Casbin `configs/rbac_policy.csv`.

**Files likely touched:** `internal/adapter/http/olt_vendor_handler.go`, `internal/adapter/http/dto/`, `internal/adapter/http/router.go`, `configs/rbac_policy.csv`

**Dependencies:** Task 3.

**Estimated scope:** Small

---

**Task 8: Endpoint GET /api/v1/devices/:id/onu-status (opsional)**

**Description:** Tambahkan handler snapshot status ONU per OLT yang memanggil driver `genericsnmp` via registry dan mengembalikan hasil walk terkini.

**Acceptance criteria:**
- [ ] Handler pada path `/api/v1/devices/:id/onu-status`, method GET, mendukung query `pon_port`.
- [ ] Handler mengambil driver via registry (bukan memanggil paket driver langsung), memanggil `ListONUStatus`, dan memetakan hasil ke DTO response sesuai bagian REST API.
- [ ] Status code: 200 sukses; 404 device tak ada; 400 device bukan OLT / vendor tanpa katalog SNMP; 502 SNMP gagal/timeout.
- [ ] Nilai OID yang tak tersedia di katalog merk dikembalikan null.
- [ ] RBAC: role minimum `teknisi`.

**Files likely touched:** `internal/adapter/http/device_handler.go` (atau handler ONU terpisah), `internal/adapter/http/dto/`, `internal/adapter/http/router.go`, `configs/rbac_policy.csv`

**Dependencies:** Task 5, Task 6.

**Estimated scope:** Medium

---

**Task 9: Rekomendasi migrasi bertahap zteolt/huaweiolt**

**Description:** Dokumentasikan (dalam ADR 0008 atau catatan terkait) keputusan migrasi bertahap: OID pindah ke katalog `genericsnmp`, sementara telnet/provisioning tetap di driver vendor.

**Acceptance criteria:**
- [ ] ADR 0008 memuat bagian "Migrasi Bertahap" yang menyatakan: OID monitoring dipindah/dirujuk ke katalog `genericsnmp`, jalur telnet/CLI provisioning tetap di `zteolt`/`huaweiolt`, dan tidak ada penghapusan paket vendor di issue ini.
- [ ] Disebutkan kriteria kapan katalog dianggap menggantikan OID inline vendor (mis. setelah endpoint dan test hijau).
- [ ] Tidak ada perubahan destruktif ke `zteolt`/`huaweiolt` dalam issue ini.

**Files likely touched:** `docs/adr/0008-genericsnmp-olt-katalog-oid.md`

**Dependencies:** Task 1.

**Estimated scope:** Small

---

**Task 10: Test**

**Description:** Tulis unit test table-driven untuk katalog per-profile dan decode/status, serta test handler dengan httptest; integrasi SNMP ke OLT nyata ditandai untuk lingkungan lab.

**Acceptance criteria:**
- [ ] Test table-driven untuk lookup katalog per profile, deteksi profile via `probe_oid` (kandidat yang merespons pertama menang), dan fail-safe profile tak dikenal.
- [ ] Test table-driven untuk decode power per-brand (auto-scale magnitude, signed-16, buang sentinel `0`/`65535` → dBm plausibel) dan pemetaan status mentah → `oper_status` via `online_values` (numerik & string, EPON & GPON).
- [ ] Test heuristik SN candidate scanning: memilih OID kandidat yang benar dan membuang sampel yang tampak status/offline/power; verifikasi hasil ter-cache di `resolved_oids` dan re-scan hanya saat cache miss/invalid.
- [ ] Test handler `olt-vendors` dan `onu-status` pakai `httptest`, termససuk kasus 400/404/502.
- [ ] Test integrasi SNMP ke OLT nyata (ZTE/Huawei di lab) ditandai/skippable di CI tanpa perangkat, bukan di-mock sebagai pengganti logika.
- [ ] Semua test berada di folder yang sama dengan file yang diuji (kecuali integrasi di `test/integration/`).

**Files likely touched:** `internal/driver/genericsnmp/catalog_test.go`, `internal/driver/genericsnmp/driver_test.go`, `internal/adapter/http/olt_vendor_handler_test.go`, `test/integration/genericsnmp_test.go`

**Dependencies:** Task 4, Task 5, Task 7, Task 8.

**Estimated scope:** Medium

## Migrasi Database

Katalog OID itu sendiri adalah data di kode (`internal/driver/genericsnmp/catalog.go`), bukan tabel database — tidak ada migrasi untuk katalog. Data time-series RX power/jarak juga tidak disimpan di issue ini — lihat Issue 10.

Satu migrasi baru diperlukan untuk **cache OID hasil deteksi per-OLT** (self-healing OID): tabel `resolved_oids` yang menyimpan, per `device_id` + profile terdeteksi, OID mana yang menang untuk field tak stabil (khususnya SN) beserta timestamp deteksi terakhir. Ini bukan katalog (katalog tetap di kode) — hanya cache hasil scanning agar walk deteksi tidak diulang tiap panggilan.

- Nomor migrasi: `000035_create_resolved_oids_table` (`.up.sql` + `.down.sql`), sudah direservasi untuk issue ini di tabel reservasi README §K6; cerminkan skemanya di `DATABASE-SCHEMA.md` pada PR yang sama.
- Kolom minimal: `device_id` (FK OLT), `profile` (kode profile terdeteksi), `field` (mis. `sn`), `resolved_oid` (OID yang menang), `detected_at`. Unik per (`device_id`, `field`).

## Verification

- [ ] `go build ./...` sukses.
- [ ] `go test ./internal/driver/genericsnmp/...` hijau (katalog per-profile + deteksi `probe_oid` + decode power per-brand + status mapping + SN candidate scanning).
- [ ] `go test ./internal/adapter/http/...` hijau (handler `olt-vendors` dan `onu-status`).
- [ ] `go test ./internal/port/... ./internal/registry/...` hijau.
- [ ] `make lint` bersih (golangci-lint, termasuk staticcheck untuk konsistensi akronim/godoc).
- [ ] Migrasi `000035_create_resolved_oids_table` (up/down) jalan bersih dan tercermin di `DATABASE-SCHEMA.md` + tabel reservasi README §K6 pada PR yang sama.
- [ ] Smoke test manual: curl GET `/api/v1/olt-vendors` (Bearer token role staff) mengembalikan array profile dengan `catalog_status` ZTE/Huawei = `complete`/`partial`.
- [ ] Smoke test manual: curl GET `/api/v1/devices/<id-olt>/onu-status` (Bearer token role teknisi) terhadap OLT lab mengembalikan snapshot ONU (profile terdeteksi via `probe_oid`); verifikasi 400 untuk device non-OLT dan 404 untuk id tak dikenal.
- [ ] ADR 0008 tertaut dari `README.md` root (cek grep link).

## Definition of Done

- [ ] ADR `0008` dibuat, status Accepted, dan tertaut dari `README.md` root pada commit yang sama.
- [ ] Port read-only optik/ONU (`OpticalReader`/`ONULister`) terdefinisi di `internal/port/` dengan tipe hasil di domain.
- [ ] `internal/driver/genericsnmp/` (driver.go + catalog.go) mengimplementasikan operasi `WalkUnconfiguredONUs`, `ReadRxPower`, `ReadDistance`, `ListONUStatus` dengan compile-time assertion port, katalog **per-profile** + deteksi `probe_oid`, decode power per-brand, SN candidate scanning, dan fail-safe untuk profile tak dikenal.
- [ ] Katalog OID profile ZTE (C300/C600) dan Huawei terisi; merk/profile lain hadir sebagai slot `stub`/`partial` yang jujur.
- [ ] Driver tersambung ke registry/factory sehingga Issue 05 dan Issue 10 memakainya lewat port.
- [ ] Endpoint `GET /api/v1/olt-vendors` (staff+) dan opsional `GET /api/v1/devices/:id/onu-status` (teknisi+) berfungsi dengan RBAC benar.
- [ ] Rekomendasi migrasi bertahap `zteolt`/`huaweiolt` terdokumentasi; tidak ada perubahan destruktif ke paket vendor eksisting.
- [ ] Semua test hijau, `make lint` bersih; satu-satunya perubahan skema adalah tabel cache `000035_create_resolved_oids_table` (tercermin di `DATABASE-SCHEMA.md` + reservasi README §K6) — katalog OID tetap data di kode.
