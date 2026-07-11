# Issue 11: ODP Topology Graph

## Konteks

Temuan E di `ANALISIS-PROVISIONING-REPO-REFERENSI.md` menyoroti bahwa skema infrastruktur saat ini (`odcs`/`odps`, migrasi `000003`, `DATABASE-SCHEMA.md §3`) hanya memodelkan hierarki **satu arah menuju OLT**: setiap ODP menunjuk ke satu `olt_device_id` dan opsional satu `odc_id`, dengan `UNIQUE(olt_device_id, pon_port)`. Model ini benar untuk topologi *flat* (satu ODP ditarik langsung dari satu port PON OLT, atau lewat satu ODC pasif). Tetapi pada jaringan riil yang bertingkat — OLT → ODC → ODP besar (splitter level-1) → ODP kecil (splitter level-2) — kabel fiber **antar ODP** tidak punya tempat untuk direpresentasikan. Akibatnya visualisasi topologi tidak bisa menggambar rantai splitter, dan penelusuran "ODP kecil ini bersumber dari splitter besar yang mana" tidak bisa dilakukan dari data.

Issue ini menambahkan tabel penghubung `odp_links` (satu baris = satu segmen kabel fiber antar dua ODP) sehingga topologi bisa direpresentasikan sebagai graf berarah: node = OLT/ODC/ODP, edge = relasi hierarki (`odps.olt_device_id`, `odps.odc_id`) ditambah `odp_links` untuk sambungan ODP→ODP. Di atas data itu dibangun satu endpoint `GET /api/v1/topology` yang merakit graf lengkap untuk konsumsi frontend visualisasi.

**Issue ini OPSIONAL.** Kalau topologi ISP dijamin selalu flat (setiap ODP terhubung langsung ke OLT/ODC tanpa splitter bertingkat antar-ODP), `odp_links` tidak akan pernah terisi dan issue ini boleh **dilewati sepenuhnya** — cukup andalkan graf hierarki dari kolom yang sudah ada. Keputusan lewat/tidaknya harus diambil di awal (lihat Open Question di §Ruang Lingkup) sebelum menulis migrasi apa pun.

## Prasyarat

- **Foundation infrastruktur** — tabel `odcs`/`odps` (migrasi `000003`) sudah ada. Domain `odp`/`odc` dan repo Postgress-nya: bila **belum** ada handler/CRUD-nya di foundation, issue ini menambahkannya (Task 4–5). Bila sudah ada, Task 4–5 diskip dan cukup dirujuk sebagai prasyarat.
- **Foundation HTTP** — router (`internal/adapter/http/router.go`), middleware auth + RBAC (`internal/adapter/http/middleware/`), dan konvensi DTO (`internal/adapter/http/dto/`) sudah tersedia.
- **RBAC Casbin** (K3) — `configs/rbac_policy.csv` sudah termuat; issue ini menambah baris policy untuk resource `odps`, `odp_links`, dan `topology`.
- Tidak bergantung pada Sync Engine (Issue 01) maupun `provisioning_sync_log` — issue ini murni data infrastruktur/topologi, **tidak ada aksi ke perangkat**.

## Ruang Lingkup

**In scope:**
- Tabel baru `odp_links` (edge kabel fiber antar ODP) + migrasi up/down + cermin ke `DATABASE-SCHEMA.md §3`.
- Domain `odplink` + kontrak port repository + implementasi Postgres.
- CRUD REST untuk `odp_links` (list, create, delete).
- Endpoint `GET /api/v1/topology` yang merakit graf OLT→ODC→ODP + `odp_links`.
- CRUD dasar ODC/ODP (Task 4–5) **hanya bila belum ada** di foundation.

**Out of scope:**
- Perhitungan optical budget / loss redaman per segmen (hanya menyimpan `cable_length_meters` sebagai metadata, tidak dihitung).
- Auto-discovery topologi dari perangkat.
- Visualisasi frontend (issue ini hanya menyediakan data graf).
- Perubahan pada relasi `subscriptions.odp_id`.

**Open Question (WAJIB dijawab sebelum implementasi):** Apakah topologi ISP ini pernah bertingkat (splitter ODP→ODP)? Jika **tidak pernah**, tutup issue ini sebagai "won't do" dan dokumentasikan alasannya di `ANALISIS-PROVISIONING-REPO-REFERENSI.md` (Temuan E). Jika **ya/mungkin**, lanjutkan.

## REST API

Base path `/api/v1/`. Semua endpoint di sini adalah data infrastruktur murni (bukan aksi ke perangkat), sehingga respons sukses memakai status sinkron biasa (`200`/`201`/`204`), **bukan** `202 Accepted` + `sync_log`.

| Method | Path | Tujuan | Role minimum |
|---|---|---|---|
| GET | `/api/v1/odps` | List ODP (opsional filter `olt_device_id`, `odc_id`) | staff |
| POST | `/api/v1/odps` | Buat ODP baru | admin |
| GET | `/api/v1/odps/:id` | Detail satu ODP | staff |
| PUT | `/api/v1/odps/:id` | Ubah ODP | admin |
| DELETE | `/api/v1/odps/:id` | Hapus ODP | admin |
| GET | `/api/v1/odp-links` | List sambungan kabel antar-ODP | staff |
| POST | `/api/v1/odp-links` | Buat sambungan `from_odp_id`→`to_odp_id` | admin |
| DELETE | `/api/v1/odp-links/:id` | Hapus sambungan | admin |
| GET | `/api/v1/topology` | Graf lengkap OLT→ODC→ODP + links | staff |

> Baris `/api/v1/odps` (5 endpoint pertama) **hanya didaftarkan bila CRUD ODP belum ada di foundation**. Bila sudah ada, jangan duplikasi rute — cukup pastikan role dan bentuk respons konsisten dengan tabel ini.

**GET `/api/v1/odps`**
- Request: query opsional `olt_device_id` (uuid), `odc_id` (uuid), plus paginasi standar (`page`, `page_size`) sesuai konvensi list existing.
- Response `200`: array objek ODP berisi `id`, `odc_id` (nullable), `olt_device_id`, `pon_port`, `name`, `capacity_ports`, `lat`, `lng`.
- Gagal: `401` (tanpa token), `403` (role kurang).

**POST `/api/v1/odps`**
- Request field penting: `olt_device_id` (wajib), `pon_port` (wajib), `name` (wajib), `odc_id` (opsional), `capacity_ports`, `lat`, `lng`.
- Response `201`: objek ODP yang baru dibuat (termasuk `id`).
- Gagal: `400` (validasi/field wajib kosong), `409` (langgar `UNIQUE(olt_device_id,pon_port)`), `403`.

**GET `/api/v1/odps/:id`**
- Response `200`: satu objek ODP. Gagal: `404` (tak ditemukan), `403`.

**PUT `/api/v1/odps/:id`**
- Request: field ODP yang boleh diubah (`name`, `odc_id`, `capacity_ports`, `lat`, `lng`; perubahan `olt_device_id`/`pon_port` boleh ditolak bila melanggar keunikan).
- Response `200`: objek ODP setelah update. Gagal: `400`, `404`, `409`, `403`.

**DELETE `/api/v1/odps/:id`**
- Response `204` tanpa body. Gagal: `404`, `409` (masih dirujuk `subscriptions.odp_id` atau `odp_links` — tolak hapus dengan pesan jelas), `403`.

**GET `/api/v1/odp-links`**
- Request: query opsional `odp_id` (kembalikan semua link yang menyentuh ODP tsb, sebagai `from` maupun `to`).
- Response `200`: array `{ id, from_odp_id, to_odp_id, cable_length_meters (nullable) }`.
- Gagal: `401`, `403`.

**POST `/api/v1/odp-links`**
- Request field penting: `from_odp_id` (wajib, FK→odps), `to_odp_id` (wajib, FK→odps), `cable_length_meters` (opsional, integer/decimal ≥0).
- Validasi domain: `from_odp_id` ≠ `to_odp_id`; kedua ODP harus ada; tolak bila membentuk siklus langsung (A→B sudah ada, cegah B→A) dan idealnya cegah siklus tak-langsung (lihat Task 2).
- Response `201`: objek link baru. Gagal: `400` (self-link / cycle / field kosong), `404` (salah satu ODP tak ada), `409` (link duplikat `from`+`to` sudah ada), `403`.

**DELETE `/api/v1/odp-links/:id`**
- Response `204`. Gagal: `404`, `403`.

**GET `/api/v1/topology`**
- Request: query opsional `olt_device_id` untuk membatasi graf ke satu OLT.
- Response `200`: objek graf dengan dua koleksi — `nodes` (array, tiap node `{ id, type IN (olt,odc,odp), label, lat, lng }`) dan `edges` (array, tiap edge `{ from, to, kind IN (hierarchy,fiber_link), cable_length_meters (nullable, hanya utk fiber_link) }`). Edge `hierarchy` diturunkan dari `odps.olt_device_id`/`odps.odc_id`; edge `fiber_link` dari `odp_links`.
- Gagal: `401`, `403`.

## Tasks

**Task 1: Domain `odplink` + validasi**
**Description:** Definisikan entity link kabel antar-ODP sebagai domain murni tanpa I/O, beserta aturan validasi konstruksi.
**Acceptance criteria:**
- [ ] File `internal/domain/odplink/odplink.go` berisi tipe `OdpLink` dengan field `ID`, `FromOdpID`, `ToOdpID`, `CableLengthMeters` (pointer/nullable).
- [ ] Konstruktor `odplink.New(...)` (≤4 param posisional, atau `NewParams` struct bila lebih) memvalidasi: `FromOdpID` ≠ `ToOdpID`, keduanya non-kosong.
- [ ] File `internal/domain/odplink/errors.go` berisi sentinel error `ErrSelfLink`, `ErrEmptyEndpoint` (dan `ErrCycleDetected` bila deteksi siklus ditaruh di domain).
- [ ] Tidak mengimpor `adapter`/`driver`/framework eksternal.
- [ ] Doc comment tiap identifier exported diawali nama identifier (patuh `staticcheck`).

**Files likely touched:** `internal/domain/odplink/odplink.go`, `internal/domain/odplink/errors.go`, `internal/domain/odplink/odplink_test.go`.
**Dependencies:** —
**Estimated scope:** Small

---

**Task 2: Deteksi siklus graf ODP**
**Description:** Tambahkan aturan yang mencegah pembuatan link yang membentuk siklus (langsung A↔B maupun tak-langsung A→…→A), karena topologi splitter harus berupa DAG.
**Acceptance criteria:**
- [ ] Fungsi cek reachability/siklus tersedia (di domain `odplink` bila hanya butuh daftar edge, atau di usecase bila butuh query repo untuk merangkai adjacency).
- [ ] Sebelum insert link `from→to`, sistem memastikan `to` belum bisa menjangkau `from` lewat link yang sudah ada.
- [ ] Kasus self-link ditolak lebih dulu (`ErrSelfLink`).
- [ ] Table-driven test menutup: link valid, self-link, cycle langsung, cycle tak-langsung (rantai ≥3 node).

**Files likely touched:** `internal/domain/odplink/odplink.go` (atau `internal/usecase/business/manage_odp_link.go`), test di folder yang sama.
**Dependencies:** Task 1
**Estimated scope:** Medium

---

**Task 3: Port + repo Postgres untuk `odp_links`**
**Description:** Definisikan kontrak repository di `port/` dan implementasinya di adapter Postgres.
**Acceptance criteria:**
- [ ] File `internal/port/odp_link_repository.go` mendefinisikan interface dengan method minimal: `Create`, `Delete`, `List` (opsional filter by odp), dan `ListAll` (untuk perakit topologi) — semua menerima `context.Context` sebagai parameter pertama, error sebagai return terakhir.
- [ ] File `internal/adapter/postgres/odp_link_repository.go` mengimplementasikan interface tsb; ada compile-time assertion `var _ port.OdpLinkRepository = (*...)(nil)`.
- [ ] Mapping model DB ↔ domain (bukan bocorkan struct GORM ke usecase).
- [ ] Error `sql.ErrNoRows`/duplikat dipetakan ke sentinel domain yang sesuai (mis. unik `from`+`to`).
- [ ] Test repo pakai `testcontainers-go` (K7), bukan mock — verifikasi create/list/delete dan penolakan duplikat.

**Files likely touched:** `internal/port/odp_link_repository.go`, `internal/adapter/postgres/odp_link_repository.go`, `internal/adapter/postgres/models.go` (tambah model), test integrasi repo.
**Dependencies:** Task 1, Migrasi Database
**Estimated scope:** Medium

---

**Task 4: (Kondisional) Domain/port/repo ODC & ODP bila belum ada**
**Description:** Bila foundation belum menyediakan domain + repo untuk `odcs`/`odps`, tambahkan agar CRUD dan topologi bisa membacanya. Skip bila sudah ada.
**Acceptance criteria:**
- [ ] Cek dulu apakah `internal/domain/odp/` (dan `odc`) + repo Postgres sudah ada. Bila ada, task ini ditandai N/A di ringkasan pekerjaan dan tidak menyentuh apa pun.
- [ ] Bila belum: domain `odp` (field sesuai `DATABASE-SCHEMA.md §3.3`: `id`, `odc_id` nullable, `olt_device_id`, `pon_port`, `name`, `capacity_ports`, `lat`, `lng`) + domain `odc` (§3.2).
- [ ] Kontrak `internal/port/odp_repository.go` (+ `odc_repository.go` bila perlu) dengan `Create/Update/Delete/FindByID/List`.
- [ ] Implementasi Postgres + test `testcontainers-go`, menegakkan `UNIQUE(olt_device_id,pon_port)`.

**Files likely touched:** `internal/domain/odp/…`, `internal/domain/odc/…`, `internal/port/odp_repository.go`, `internal/adapter/postgres/odp_repository.go`, tests.
**Dependencies:** — (independen dari `odp_links`)
**Estimated scope:** Medium (Large bila keduanya benar-benar belum ada)

---

**Task 5: Usecase bisnis manage ODP & ODP-link**
**Description:** Orkestrasi CRUD ODP (bila baru dari Task 4) dan CRUD link, memanggil port repo tanpa tahu detail HTTP/DB.
**Acceptance criteria:**
- [ ] File `internal/usecase/business/manage_odp_link.go` berisi fungsi bergaya `VerbNoun` untuk create/delete/list link; memanggil validasi siklus (Task 2) sebelum create.
- [ ] Bila Task 4 aktif: `internal/usecase/business/manage_odp.go` untuk CRUD ODP; tolak delete ODP yang masih dirujuk (`subscriptions.odp_id` atau `odp_links`) dengan sentinel error yang bisa dipetakan ke `409`.
- [ ] Table-driven test usecase (K7) menutup happy path + penolakan (self-link, cycle, duplikat, delete-in-use).

**Files likely touched:** `internal/usecase/business/manage_odp_link.go`, `internal/usecase/business/manage_odp.go` (kondisional), tests.
**Dependencies:** Task 2, Task 3, Task 4 (kondisional)
**Estimated scope:** Medium

---

**Task 6: Usecase perakit topologi (graf)**
**Description:** Bangun graf gabungan dari edge hierarki (`odps.olt_device_id`/`odps.odc_id`) dan edge `odp_links`, siap diserialisasi jadi `nodes`+`edges`.
**Acceptance criteria:**
- [ ] File `internal/usecase/business/build_topology.go` (nama fungsi `BuildTopology`) mengembalikan struktur graf domain berisi koleksi node (`type` olt/odc/odp) dan edge (`kind` hierarchy/fiber_link).
- [ ] Node OLT diambil dari device bertipe OLT yang dirujuk `odps.olt_device_id`; node ODC dari `odcs`; node ODP dari `odps`.
- [ ] Edge hierarki: ODP→ODC (bila `odc_id` non-null) lalu ODC→OLT, atau ODP→OLT langsung bila `odc_id` null. Edge fiber_link: dari `odp_links` (bawa `cable_length_meters`).
- [ ] Mendukung filter opsional per `olt_device_id`.
- [ ] Table-driven test menutup: flat (tanpa ODC, tanpa link), bertingkat (ODC + rantai fiber_link), dan graf kosong.

**Files likely touched:** `internal/usecase/business/build_topology.go`, test di folder yang sama.
**Dependencies:** Task 3, Task 4 (data ODP/ODC), Task 5
**Estimated scope:** Medium

---

**Task 7: Handler REST + DTO + registrasi rute + RBAC**
**Description:** Ekspos endpoint `odp_links`, `topology`, dan (kondisional) CRUD ODP; daftarkan rute dan policy Casbin.
**Acceptance criteria:**
- [ ] File `internal/adapter/http/odp_link_handler.go` menangani list/create/delete link; `internal/adapter/http/topology_handler.go` menangani `GET /api/v1/topology`; (kondisional) `internal/adapter/http/odp_handler.go` untuk CRUD ODP.
- [ ] DTO request/response di `internal/adapter/http/dto/` (mis. `odp_link.go`, `topology.go`); handler tidak memanggil repo/port device langsung, hanya usecase.
- [ ] Rute didaftarkan di `internal/adapter/http/router.go` sesuai tabel §REST API; role ditegakkan lewat middleware RBAC.
- [ ] `configs/rbac_policy.csv` ditambah: tulis (`odps`, `odp_links`) untuk admin ke atas; baca (`odps`, `odp_links`, `topology`) untuk staff ke atas.
- [ ] Status code sesuai spesifikasi (`201`/`204`/`409`/`400`/`404`); tidak ada `202` (bukan aksi perangkat).
- [ ] Test handler pakai `httptest` (K7) menutup sukses + `403` + `409`.

**Files likely touched:** `internal/adapter/http/odp_link_handler.go`, `internal/adapter/http/topology_handler.go`, `internal/adapter/http/odp_handler.go` (kondisional), `internal/adapter/http/dto/*.go`, `internal/adapter/http/router.go`, `configs/rbac_policy.csv`, tests.
**Dependencies:** Task 5, Task 6
**Estimated scope:** Medium

---

**Task 8: Dokumentasi — ADR opsional + update DATABASE-SCHEMA.md + OpenAPI**
**Description:** Rekam keputusan topologi bertingkat dan cerminkan skema baru ke dokumen kanonik.
**Acceptance criteria:**
- [ ] `DATABASE-SCHEMA.md §3` diperbarui: tambah deskripsi tabel `odp_links` (kolom, FK, indeks/keunikan) pada PR yang sama dengan migrasi (K6).
- [ ] (Opsional tapi disarankan) ADR baru `docs/adr/0010-topologi-odp-bertingkat.md` menjelaskan alasan `odp_links` sebagai DAG antar-ODP dan mengapa graf dibangun di usecase, ditautkan dari `README.md` root pada commit yang sama (AGENTS.md §1.5).
- [ ] `api/openapi.yaml` ditambah path `/odps` (kondisional), `/odp-links`, `/topology`.
- [ ] Bila Open Question dijawab "flat selamanya", Task 8 hanya mencatat keputusan won't-do di `ANALISIS-PROVISIONING-REPO-REFERENSI.md` dan issue ditutup tanpa migrasi.

**Files likely touched:** `DATABASE-SCHEMA.md`, `docs/adr/0010-topologi-odp-bertingkat.md`, `README.md`, `api/openapi.yaml`, `ANALISIS-PROVISIONING-REPO-REFERENSI.md`.
**Dependencies:** Migrasi Database, Task 7
**Estimated scope:** Small

---

## Migrasi Database

Nomor migrasi lanjut dari `000021` → gunakan **`000031`**.

- File up: `migrations/000031_create_odp_links_table.up.sql`
- File down: `migrations/000031_create_odp_links_table.down.sql`

Tabel `odp_links` (dijelaskan sebagai teks, bukan SQL):
- `id` — primary key (uuid, konsisten dengan tabel infrastruktur lain).
- `from_odp_id` — FK → `odps(id)`, `NOT NULL`. Sisi sumber (ODP splitter besar/hulu).
- `to_odp_id` — FK → `odps(id)`, `NOT NULL`. Sisi tujuan (ODP hilir).
- `cable_length_meters` — nullable (integer atau numeric ≥ 0), metadata panjang kabel segmen ini.
- `created_at` / `updated_at` — timestamp, konsisten konvensi tabel lain.
- **Constraint:** `CHECK(from_odp_id <> to_odp_id)` untuk cegah self-link di level DB; `UNIQUE(from_odp_id, to_odp_id)` untuk cegah duplikat segmen.
- **Behavior FK:** `ON DELETE RESTRICT` (atau tolak di usecase) agar ODP yang masih punya link tidak terhapus diam-diam — konsisten dengan aturan `409` di handler.
- **Indeks:** index pada `from_odp_id` dan pada `to_odp_id` (query topologi dan filter `odp_id` menyentuh kedua kolom).

File `.down.sql` melakukan drop tabel `odp_links` (beserta index/constraint-nya) secara idempoten aman.

Cerminkan seluruh definisi di atas ke `DATABASE-SCHEMA.md §3` pada PR yang sama (Task 8).

> Bila Open Question dijawab "topologi flat selamanya": **Tidak ada perubahan skema** — jangan buat migrasi `000031`, tutup issue.

## Verification

- [ ] `go build ./...` sukses.
- [ ] `go test ./internal/domain/odplink/...` hijau (validasi + deteksi siklus).
- [ ] `go test ./internal/usecase/business/...` hijau (manage link, build topology, dan manage odp bila Task 4 aktif).
- [ ] `go test ./internal/adapter/postgres/...` hijau via `testcontainers-go` (create/list/delete link, penolakan duplikat, penolakan delete ODP in-use).
- [ ] `go test ./internal/adapter/http/...` hijau (`httptest`: sukses, `403`, `409`).
- [ ] `make lint` bersih (`gofumpt`/`goimports`/`staticcheck`), termasuk doc comment exported.
- [ ] Migrasi up lalu down berjalan tanpa error pada DB test.
- [ ] Smoke test manual (sebagai teks curl): `POST /api/v1/odps` dua ODP → `POST /api/v1/odp-links` menghubungkan keduanya → `GET /api/v1/topology` menampilkan dua node ODP + satu edge `fiber_link`; ulangi `POST /api/v1/odp-links` yang membentuk cycle → harus `400`; `POST /api/v1/odp-links` sebagai role `staff` → harus `403`.

## Definition of Done

- [ ] Open Question terjawab dan tercatat (flat → won't-do dan issue ditutup; bertingkat → lanjut).
- [ ] Tabel `odp_links` + migrasi `000031` up/down berpasangan, tercermin di `DATABASE-SCHEMA.md §3`.
- [ ] Domain `odplink` + port + repo Postgres lengkap dengan deteksi siklus, semua boundary AGENTS.md §1 dipatuhi (domain tidak impor adapter/driver).
- [ ] Endpoint `odp-links` (list/create/delete) dan `topology` berfungsi dengan RBAC benar; CRUD ODP tersedia (dari foundation atau ditambah Task 4).
- [ ] Semua test (unit/usecase table-driven, repo testcontainers, handler httptest) hijau; `make lint` bersih.
- [ ] Dokumentasi (ADR opsional, OpenAPI, DATABASE-SCHEMA) diperbarui dan ditautkan dari `README.md` root pada commit yang sama.
