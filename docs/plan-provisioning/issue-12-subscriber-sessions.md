# Issue 12: Subscriber Session Tracking

## Konteks

Tabel `subscriber_sessions` (DATABASE-SCHEMA.md §6.4, dibuat oleh migrasi 000012) dirancang untuk merekam kapan seorang pelanggan **online** dan **offline** di jaringan — satu baris per sesi PPPoE, berisi username, IP yang di-assign, MAC, waktu mulai/berhenti, byte counter, dan penyebab putus. Temuan di `ANALISIS-PROVISIONING-REPO-REFERENSI.md` menegaskan bahwa repo referensi mengisi data sejenis dengan **polling** `/ppp active print` berkala. Pendekatan itu **dilarang** di Polyglot untuk protokol yang mendukung push (RouterOS) sesuai ADR 0003 — polling boros koneksi, telat, dan tidak scalable untuk ratusan router.

Isu ini membangun jalur pengisian `subscriber_sessions` dari **event stream real-time**: memanfaatkan `port.StreamingDeviceDriver` dan `internal/driver/mikrotik/stream.go` yang sudah ada (`Stream()` untuk `/log follow`). Dengan menyaring `topic ppp,info`, event connect/disconnect PPPoE terekam saat kejadian, bukan saat di-poll. Consumer jangka-panjang berjalan per device MikroTik aktif, dipasang dari `main.go` di bawah supervisor stream yang sudah ada, dan reconnect otomatis bila koneksi putus.

**Peringatan domain yang mudah tertukar:** `subscriber_sessions` adalah *status pelanggan online*, sama sekali berbeda dari `internal/domain/session` yang berarti *koneksi driver KE perangkat*. Penamaan tabel sengaja `subscriber_sessions`, bukan `sessions`. Semua tipe/berkas di isu ini memakai awalan/nama `subscribersession` agar tidak bertabrakan dengan `domain/session`. Jangan pernah menyatukan keduanya.

## Prasyarat

Harus sudah ada / dipakai sebagai foundation:

- Migrasi **000012** (`subscriber_sessions`) sudah ada — isu ini tidak mengubah skema.
- Migrasi **000021** (`device_stream_subscriptions`) sudah ada — dipakai untuk mencatat stream per device yang aktif.
- Migrasi **000009** (`subscriptions`) dengan kolom `pppoe_username` + `device_id` — kunci pemetaan username → subscription.
- Driver MikroTik lengkap: `internal/driver/mikrotik/driver.go`, `connect.go` (`dialAndLogin`), `stream.go` (implementasi `port.StreamingDeviceDriver.Stream()` untuk `/log follow`).
- `internal/port/streaming_driver.go` (`port.StreamingDeviceDriver`) dan `internal/registry/registry.go` (`Get(ctx, deviceID)` mengembalikan `port.DeviceDriver` yang di-cache per device).
- Supervisor stream device + pola errgroup/cancel yang sudah dipakai fondasi `stream_output.go` dan `main.go` (reconnect via supervisor yang sudah ada).
- RBAC Casbin (`configs/rbac_policy.csv`) dan middleware auth/rbac (`internal/adapter/http/middleware/`).

Tidak bergantung pada Issue 01 (Sync Engine): isu ini hanya **membaca** event dari perangkat dan menulis ke `subscriber_sessions`, tidak menulis `provisioning_sync_log` maupun mengeksekusi command destruktif — sehingga alur sinkronisasi K4 tidak berlaku di sini.

## Ruang Lingkup

**In scope**

- Domain `subscribersession` (entity sesi + tipe event connect/disconnect).
- Kontrak `port.SubscriberSessionRepository` + implementasi Postgres (map ke tabel 000012).
- Kontrak `port.SubscriberSessionStreamer` (port baru — diusulkan eksplisit di Task 3) + implementasinya di driver MikroTik dengan parser log. **Tiga topik log** didukung: `ppp,info` (PPPoE), `hotspot,info` (hotspot subscription mode A `mac_login`, Issue 13), dan `netwatch` (hotspot subscription mode B `ip_binding`, host up/down atas IP statik yang dipasang `/tool netwatch`). Ketiganya menghasilkan `subscribersession.Event` connect/disconnect yang sama — hanya parser per-topik yang berbeda, satu mesin monitoring.
- Usecase network consumer jangka-panjang: konsumsi event stream → map identitas (username PPPoE / kode hotspot / IP netwatch) → subscription → tulis `subscriber_sessions`.
- Supervisor per-device di `main.go` (goroutine errgroup/cancel, reconnect, catat `device_stream_subscriptions`), dengan bootstrap snapshot saat tiap stream connect + reconciliation pass periodik ber-interval longgar untuk menutup gap saat stream putus.
- Usecase + handler REST read-only: riwayat sesi per subscription, daftar pelanggan online (global dan per device).
- ADR 0011 mendokumentasikan keputusan event-stream + bootstrap snapshot + tiga topik log (ppp/hotspot/netwatch), ditautkan dari `README.md` root.

**Out of scope**

- Polling `/ppp active print` berulang **sebagai mekanisme monitoring utama** (dilarang ADR 0003 — inilah yang dilakukan repo referensi tiap 30s). Yang DIIZINKAN dan bukan pelanggaran: **satu** snapshot saat tiap stream connect (seed), dan reconciliation pass ber-interval longgar (default 5 menit) sebagai jaring pengaman — keduanya bukan loop monitoring real-time.
- Byte counter live untuk sesi yang masih aktif (butuh polling — tidak dilakukan; byte diisi saat event disconnect).
- Vendor selain MikroTik (Cisco/OLT/GenieACS) — isu ini hanya MikroTik (PPPoE + hotspot mode A/B); vendor lain menyusul di isu terpisah dengan pola port yang sama.
- Perubahan skema tabel apa pun.

## REST API

Semua endpoint di bawah `base path /api/v1/` dan bersifat **read-only** terhadap database (bukan aksi ke perangkat), sehingga mengembalikan **200 OK** langsung — bukan 202 Accepted (tidak ada baris `sync_log`).

| Method | Path | Tujuan | Role minimum |
|---|---|---|---|
| GET | /api/v1/subscriptions/:id/sessions | Riwayat sesi online/offline satu pelanggan | staff |
| GET | /api/v1/sessions/active | Daftar semua pelanggan yang sedang online | staff |
| GET | /api/v1/devices/:id/sessions/active | Daftar pelanggan online pada satu router | teknisi |

### GET /api/v1/subscriptions/:id/sessions

- **Request:** path param `:id` = `subscription_id`. Query opsional: `limit` (default 50, maks 200), `offset`, `from`/`to` (filter rentang `started_at`, format RFC3339), `active_only` (boolean).
- **Response (200):** objek berisi `data` (array sesi) + `pagination` (`limit`, `offset`, `total`). Tiap item: `id`, `subscription_id`, `device_id`, `pppoe_username`, `framed_ip`, `mac_address`, `started_at`, `stopped_at` (null bila masih online), `bytes_in`, `bytes_out` (null/0 selama sesi aktif — lihat catatan byte di Task 4), `terminate_cause` (null bila masih online). Diurutkan `started_at` menurun.
- **Gagal:** 404 bila `subscription_id` tidak ditemukan; 400 bila query param tidak valid; 401/403 sesuai auth/RBAC.

### GET /api/v1/sessions/active

- **Request:** query opsional `device_id`, `plan_id`, `limit`, `offset`, dan `q` (pencarian sebagian pada `pppoe_username`).
- **Response (200):** `data` (array sesi dengan `stopped_at IS NULL`) + `pagination`. Field per item sama dengan endpoint riwayat, ditambah `uptime_seconds` (diturunkan dari `now - started_at`, dihitung di sisi server). `bytes_in`/`bytes_out` untuk sesi aktif **selalu null/0** (tidak di-poll).
- **Gagal:** 400 query tidak valid; 401/403 sesuai auth/RBAC.

### GET /api/v1/devices/:id/sessions/active

- **Request:** path param `:id` = `device_id`. Query opsional `limit`, `offset`, `q`.
- **Response (200):** identik dengan `/sessions/active` tetapi difilter satu device; menyertakan ringkasan `count` jumlah pelanggan online pada router itu.
- **Gagal:** 404 bila `device_id` tidak ditemukan; 400 query tidak valid; 401/403 sesuai auth/RBAC.

## Tasks

**Task 1: Definisikan domain `subscribersession`**

**Description:** Buat entity murni untuk satu sesi pelanggan dan tipe event connect/disconnect yang menjadi bahasa domain antar-layer, tanpa I/O maupun import eksternal.

**Acceptance criteria:**
- [ ] Folder `internal/domain/subscribersession/` dibuat; nama package `subscribersession`.
- [ ] Tipe entity (usulan `subscribersession.Session`) memuat field: id, subscriptionID (opsional/nullable untuk username tak dikenal), deviceID, pppoeUsername, framedIP, macAddress, startedAt, stoppedAt (opsional), bytesIn, bytesOut, terminateCause.
- [ ] Tipe event (usulan `subscribersession.Event`) memuat: Kind (enum `EventConnect`/`EventDisconnect`), pppoeUsername, framedIP, macAddress, terminateCause, bytesIn/bytesOut (opsional), occurredAt.
- [ ] Ada helper klasifikasi/validasi ringan bila perlu (mis. `Event.IsConnect()`), tanpa logika I/O.
- [ ] Tidak ada import ke `adapter`/`driver`/framework eksternal (boundary domain terjaga).
- [ ] Setiap identifier exported punya doc comment diawali namanya sendiri.

**Files likely touched:** `internal/domain/subscribersession/subscribersession.go`, `internal/domain/subscribersession/event.go`.

**Dependencies:** —

**Estimated scope:** Small

---

**Task 2: Kontrak `port.SubscriberSessionRepository` + implementasi Postgres**

**Description:** Definisikan interface repository di `port/` dan implementasi Postgres yang memetakan `subscribersession.Session` ke/dari tabel `subscriber_sessions` (000012).

**Acceptance criteria:**
- [ ] `internal/port/subscriber_session_repository.go` mendefinisikan `SubscriberSessionRepository` dengan method minimal: `OpenSession(ctx, Session)` (insert baris baru saat connect), `CloseSession(ctx, ...)` (update `stopped_at`, `bytes_in/out`, `terminate_cause` untuk sesi terbuka yang cocok), `FindOpenByUsername(ctx, username, deviceID)` (cari sesi `stopped_at IS NULL` untuk mencocokkan event disconnect), `ListBySubscription(ctx, subscriptionID, filter)`, `ListActive(ctx, filter)`, `ListActiveByDevice(ctx, deviceID, filter)`.
- [ ] `context.Context` adalah parameter pertama pada semua method; error return terakhir.
- [ ] Implementasi di `internal/adapter/postgres/subscriber_session_repository.go` memakai pola repo Postgres yang sudah ada (GORM/db.go).
- [ ] Insert connect mengisi `started_at`, `pppoe_username`, `framed_ip`, `mac_address`, `device_id`, dan `subscription_id` bila resolusi berhasil (boleh null bila tak dikenal).
- [ ] `CloseSession` idempoten aman: bila tidak ada sesi terbuka yang cocok, tidak error fatal — kembalikan sentinel (usulan `ErrNoOpenSession`) yang bisa di-`errors.Is` caller, atau no-op ter-log.
- [ ] Query list mendukung filter pagination (`limit`/`offset`) dan filter opsional (`device_id`, `plan_id`, pencarian username) via join ke `subscriptions`/`plans` seperlunya.
- [ ] Error dibungkus `%w` dengan konteks operasi; `sql.ErrNoRows` dipetakan ke sentinel domain, bukan dibandingkan string.

**Files likely touched:** `internal/port/subscriber_session_repository.go`, `internal/adapter/postgres/subscriber_session_repository.go`, (opsional) `internal/domain/subscribersession/errors.go`.

**Dependencies:** Task 1.

**Estimated scope:** Medium

---

**Task 3: Port `SubscriberSessionStreamer` + parser log (ppp,info / hotspot,info / netwatch) di driver MikroTik**

**Description:** Usulkan **eksplisit** satu port baru yang mengubah stream mentah `/log follow` menjadi aliran `subscribersession.Event` domain, dan implementasikan parser format log RouterOS di dalam driver MikroTik (pengetahuan vendor wajib di `internal/driver/mikrotik/`, bukan di usecase). Mendukung tiga topik log untuk PPPoE dan hotspot subscription (Issue 13 mode A/B).

**Acceptance criteria:**
- [ ] **Proposal layer dinyatakan:** `internal/port/subscriber_session_streamer.go` mendefinisikan `SubscriberSessionStreamer` dengan method (usulan) `StreamSessions(ctx) (<-chan subscribersession.Event, error)`. Ini port baru (bukan layer baru) — dinyatakan sebagai proposal sesuai AGENTS.md §0.4, bukan pola diam-diam.
- [ ] `subscribersession.Event` memuat `Kind` sumber (`pppoe`/`hotspot`/`netwatch`) + identitas yang relevan (username PPPoE / kode hotspot+MAC / IP netwatch) agar usecase (Task 4) bisa memetakan ke subscription lewat kolom yang tepat.
- [ ] Driver MikroTik menambah berkas `internal/driver/mikrotik/sessions.go` (atau `ppp_events.go`) berisi: (a) parser per-topik dari satu baris log menjadi `subscribersession.Event` connect/disconnect untuk `ppp,info`, `hotspot,info`, dan `netwatch`; (b) implementasi `StreamSessions` yang membuka `Stream()` `/log follow` terfilter ketiga topik (satu stream, filter `topic` mencakup ppp/hotspot/netwatch), membaca chunk, memparser sesuai topik, lalu mengirim event ke channel.
- [ ] Parser `ppp,info`: ekstrak `pppoe_username`, `framed_ip`, `mac_address` (login); `terminate_cause`, `bytes_in/out` bila ada (logout). Serap juga field native RouterOS yang tersedia di baris log/tabel active bila ada: `caller-id` (MAC/identitas pemanggil), `service` (pppoe/l2tp/pptp/…), `address` (IP ter-assign, kadang berbeda dari yang di-parse dari teks), `uptime` (durasi sesi saat logout, dipakai lintas-cek `stopped_at - started_at`), dan `last-logged-out` (timestamp putus dari sisi router). Petakan ke field domain yang setara; simpan yang tidak punya kolom sebagai konteks event best-effort (jangan mengarang bila absen).
- [ ] Parser `hotspot,info`: ekstrak username/kode voucher, `framed_ip`, `mac_address` dari event login/logout hotspot. Serap juga field native hotspot bila tersedia: `mac-address` (identitas klien, kunci utama mode A `mac_login`), `login-by` (metode: `cookie`/`http-chap`/`mac`/`trial` — penting untuk deteksi mac-cookie, lihat catatan mac-cookie di Task 4), `session-time-left` (sisa kuota waktu), dan `bytes-in`/`bytes-out` (byte counter hotspot saat logout). Field tanpa kolom domain disimpan sebagai konteks event best-effort.
- [ ] Parser `netwatch`: ekstrak host IP + status up/down → connect/disconnect (mode B tidak punya username; identitas = IP statik yang dipetakan ke subscription oleh Task 4).
- [ ] Field yang tidak tersedia di log dibiarkan null (best-effort, jangan mengarang). Baris tak relevan di-skip tanpa error.
- [ ] `var _ port.SubscriberSessionStreamer = (*Driver)(nil)` sebagai bukti compile-time bahwa `*mikrotik.Driver` memenuhi kontrak.
- [ ] Channel ditutup rapi saat `ctx` dibatalkan atau stream berakhir; tidak ada goroutine bocor.
- [ ] Tidak ada polling (`/ppp active print`, `/ip hotspot active print`, atau ping berkala dari Golang) di jalur ini — mode B memakai event `netwatch` router, bukan ping dari backend (ADR 0003).

**Files likely touched:** `internal/port/subscriber_session_streamer.go`, `internal/driver/mikrotik/sessions.go`, (mungkin sedikit) `internal/driver/mikrotik/driver.go`.

**Dependencies:** Task 1.

**Estimated scope:** Large

---

**Task 4: Usecase consumer `TrackSubscriberSessions`**

**Description:** Buat orkestrasi network jangka-panjang untuk satu device: konsumsi `subscribersession.Event`, resolusi username→subscription, dan tulis ke repository.

**Acceptance criteria:**
- [ ] Berkas `internal/usecase/network/track_subscriber_sessions.go` berisi `func TrackSubscriberSessions(...)` yang menerima `ctx`, satu `port.SubscriberSessionStreamer` (diperoleh via type-assert dari `registry.Get`), `port.SubscriberSessionRepository`, dan `port.SubscriptionRepository`.
- [ ] Loop membaca channel event sampai `ctx` selesai; setiap event `EventConnect` → resolusi `subscription_id` dari `pppoe_username`+`device_id` lalu `OpenSession`; setiap `EventDisconnect` → `FindOpenByUsername` lalu `CloseSession` mengisi `stopped_at`/byte/cause.
- [ ] Resolusi identitas → subscription tergantung `Event.Kind`: `pppoe`/`hotspot` via username (usulan tambah `FindByPPPoEUsername`/`FindByHotspotIdentity(ctx, code, deviceID)` bila belum ada — nyatakan sebagai penambahan method); `netwatch` (mode B `ip_binding`) via `FindByStaticIP(ctx, ip, deviceID)` yang mencocokkan `subscriptions.static_ip` + `hotspot_access_mode='ip_binding'`. Identitas tak dikenal tetap dicatat dengan `subscription_id` null + log peringatan, tidak menggagalkan consumer.
- [ ] Usecase **tidak** meng-import package driver mana pun secara langsung (hanya `port` + `domain`) — boundary K1 terjaga; deteksi vendor terjadi di supervisor/registry, bukan di sini.
- [ ] Kegagalan menulis satu event di-log dan di-skip; consumer tidak mati karena satu error tulis (kecuali `ctx` dibatalkan).
- [ ] **Catatan byte:** dokumentasikan (komentar singkat merujuk ADR 0011) bahwa `bytes_in/out` hanya terisi saat disconnect; sesi aktif tidak di-update byte-nya karena itu memerlukan polling terlarang. Untuk mode B (`netwatch`) byte counter memang tidak tersedia (netwatch hanya up/down) — biarkan null.
- [ ] **Catatan mac-cookie 'hantu' (interaksi dengan K9):** pada hotspot mode A (mac-cookie aktif, K13), cookie yang bertahan di router bisa memicu **auto-relogin tanpa kredensial** — saat consumer monitoring hidup, ini muncul sebagai event connect hotspot "hantu" untuk MAC yang sebenarnya sudah tidak seharusnya online (mis. setelah subscription di-suspend). Consumer **tidak** menyembunyikan/menyaring event ini (ia mencatat apa yang benar-benar terjadi di jaringan), tetapi dokumentasikan (komentar merujuk ADR 0011 + K9) bahwa sumber sesi hantu adalah cookie yang belum di-`/ip hotspot cookie remove`. Penghapusan cookie adalah tanggung jawab langkah kill sesi aktif **sesuai K9 (lihat README §Konvensi Bersama, K9 → cookie remove saat putus)** di jalur provisioning (Issue 13/Sync Engine), bukan di consumer read-only ini — consumer hanya perlu tahu bahwa sesi auto-relogin ber-`login-by=cookie` adalah gejala cookie yang belum dibersihkan, bukan bug parser.

**Files likely touched:** `internal/usecase/network/track_subscriber_sessions.go`, (mungkin) `internal/port/subscription_repository.go`, `internal/adapter/postgres/subscription_repository.go`.

**Dependencies:** Task 1, Task 2, Task 3.

**Estimated scope:** Medium

---

**Task 5: Supervisor per-device di `main.go` + bootstrap snapshot & reconciliation pass**

**Description:** Pasang `TrackSubscriberSessions` sebagai goroutine per device MikroTik aktif di startup, di bawah errgroup/cancel dan supervisor reconnect yang sudah ada, serta catat stream ke `device_stream_subscriptions`. Sertakan bootstrap snapshot saat stream connect dan reconciliation pass periodik agar tidak ada gap state saat stream putus.

**Acceptance criteria:**
- [ ] `cmd/server/main.go` men-spawn satu consumer per device MikroTik ber-status aktif menggunakan errgroup dengan `ctx` yang bisa dibatalkan saat shutdown; setiap goroutine menerima variabel loop sebagai parameter eksplisit.
- [ ] Reconnect memakai supervisor stream yang sudah ada (backoff/reconnect) — bukan implementasi loop reconnect baru; bila device putus, stream dibuka ulang tanpa menduplikasi baris sesi.
- [ ] Baris `device_stream_subscriptions` (000021) dibuat/diupdate untuk menandai stream sesi aktif per device, dan dibersihkan/di-mark saat berhenti (sesuai skema tabel itu).
- [ ] **Bootstrap snapshot WAJIB saat stream connect (bukan opsional):** `/log follow` hanya mengirim **delta ke depan** — sesi yang sudah online sebelum stream hidup tidak akan pernah muncul sebagai event connect. Karena itu, **setiap kali** stream terbuka (start awal maupun setelah reconnect) consumer membaca `/ppp active print` **satu kali** (dan `/ip hotspot active print` satu kali bila topik hotspot dipantau) untuk menyeed state awal, lalu langsung beralih ke stream. Field-tested: kelima repo referensi justru memakai `/ppp active print` sebagai mekanisme monitoring utama (polling + diff tiap 30s); Polyglot memakai snapshot ini **hanya untuk seed**, bukan loop berkala — inilah yang membuat pendekatan event-stream superior tanpa kehilangan state awal (lihat REFERENCES.md, bukti K-provisioning). Snapshot ter-seed direkonsiliasi dengan sesi terbuka di DB agar tidak menduplikasi baris (sesi yang sudah tercatat open tidak di-insert ulang).
- [ ] **Reconciliation pass periodik (menutup gap saat stream putus):** jalankan `/ppp active print` (+ hotspot active bila relevan) berkala pada interval **rendah** (usulan default 5 menit, dikontrol flag config) untuk mendeteksi (a) sesi yang **hilang** dari router tetapi masih `stopped_at IS NULL` di DB → tutup dengan `terminate_cause` sintetik (mis. `reconciliation`) karena event disconnect-nya terlewat saat stream putus; (b) sesi yang **ada** di router tetapi tidak ada baris open di DB → buka baris baru. Ini **bukan** polling monitoring 30s ala referensi — ini jaring pengaman ber-interval jauh lebih longgar yang menutup gap saat stream reconnect; boundary "no polling" ADR 0003 tetap dijaga karena sumber kebenaran real-time tetap event stream, reconciliation hanya koreksi drift. Interval + on/off dikontrol `internal/config/config.go`; dokumentasikan trade-off di ADR 0011.
- [ ] Bootstrap + reconciliation dikontrol flag config (default: keduanya aktif) di `internal/config/config.go`.
- [ ] Shutdown bersih: pembatalan `ctx` menutup semua consumer dan channel tanpa goroutine bocor; `g.Wait()` di-propagate.

**Files likely touched:** `cmd/server/main.go`, `internal/config/config.go`, (mungkin) `internal/usecase/network/track_subscriber_sessions.go` untuk hook snapshot.

**Dependencies:** Task 4.

**Estimated scope:** Large

---

**Task 6: Usecase read-only riwayat & sesi aktif**

**Description:** Buat orkestrasi baca untuk tiga endpoint: riwayat per subscription, daftar aktif global, dan daftar aktif per device.

**Acceptance criteria:**
- [ ] `internal/usecase/network/list_subscriber_sessions.go` (`ListSubscriberSessions`) mengembalikan riwayat sesi satu subscription dengan filter pagination/rentang waktu; validasi keberadaan subscription (kembalikan sentinel not-found bila tidak ada).
- [ ] `internal/usecase/network/list_active_sessions.go` (`ListActiveSessions`) mengembalikan sesi `stopped_at IS NULL` dengan filter opsional device/plan/username; menghitung `uptime_seconds` di layer usecase/handler (bukan disimpan).
- [ ] Varian per-device dilayani `ListActiveSessions` dengan filter `device_id` atau method repo `ListActiveByDevice`; validasi keberadaan device.
- [ ] Semua fungsi memakai `context.Context` sebagai parameter pertama, error terakhir; hanya bergantung ke `port` + `domain`.

**Files likely touched:** `internal/usecase/network/list_subscriber_sessions.go`, `internal/usecase/network/list_active_sessions.go`.

**Dependencies:** Task 2.

**Estimated scope:** Medium

---

**Task 7: Handler REST + DTO + wiring router + RBAC**

**Description:** Ekspos ketiga endpoint melalui handler HTTP, dengan DTO respons, registrasi rute, dan penambahan policy Casbin.

**Acceptance criteria:**
- [ ] `internal/adapter/http/subscriber_session_handler.go` mengimplementasikan tiga handler sesuai bagian REST API; nama file/handler jelas berbeda dari sesi-driver.
- [ ] DTO respons di `internal/adapter/http/dto/subscriber_session.go` (item sesi + wrapper pagination); mapping domain→DTO tidak membocorkan tipe internal.
- [ ] Rute didaftarkan di `internal/adapter/http/router.go`: `GET /api/v1/subscriptions/:id/sessions`, `GET /api/v1/sessions/active`, `GET /api/v1/devices/:id/sessions/active`.
- [ ] Middleware auth + RBAC dipasang; `configs/rbac_policy.csv` ditambah baris agar: `superadmin`/`owner`/`admin` boleh ketiganya, `teknisi` boleh ketiganya (termasuk device-scoped), `staff` boleh dua endpoint global/per-subscription. Endpoint device-scoped minimal `teknisi`.
- [ ] Status code sesuai spesifikasi: 200 sukses, 404 resource tak ada, 400 query invalid, 401/403 auth/RBAC.
- [ ] Pagination & filter query di-parse dan divalidasi (batas `limit` maksimum ditegakkan).

**Files likely touched:** `internal/adapter/http/subscriber_session_handler.go`, `internal/adapter/http/dto/subscriber_session.go`, `internal/adapter/http/router.go`, `configs/rbac_policy.csv`.

**Dependencies:** Task 6.

**Estimated scope:** Medium

---

**Task 8: ADR 0011 + tautan README**

**Description:** Dokumentasikan keputusan arsitektur "subscriber session via event stream + bootstrap snapshot per connect + reconciliation pass" dan port baru `SubscriberSessionStreamer`.

**Acceptance criteria:**
- [ ] `docs/adr/0011-subscriber-session-via-event-stream.md` dibuat (empat digit, lanjut dari 0005) menjelaskan: alasan menolak polling monitoring 30s ala referensi (ADR 0003), pemilihan `/log follow topic=ppp,info`, port `SubscriberSessionStreamer`, trade-off bootstrap snapshot saat tiap connect + reconciliation pass periodik ber-interval longgar (mengapa keduanya bukan pelanggaran "no polling"), dan batasan byte counter live.
- [ ] `README.md` root ditautkan ke ADR 0011 **pada commit yang sama** (AGENTS.md §1.5).
- [ ] ADR merujuk `subscriber_sessions` (§6.4) dan menegaskan perbedaan dengan `domain/session`.

**Files likely touched:** `docs/adr/0011-subscriber-session-via-event-stream.md`, `README.md`.

**Dependencies:** Task 3, Task 5 (agar keputusan final tercermin).

**Estimated scope:** Small

---

**Task 9: Test**

**Description:** Uji parser log, usecase consumer, repository Postgres, dan handler.

**Acceptance criteria:**
- [ ] Test parser MikroTik table-driven: input beberapa baris log `ppp,info` (connect/disconnect/irelevan) → `subscribersession.Event` yang benar, termasuk kasus field byte/cause absen. Berkas `internal/driver/mikrotik/sessions_test.go`.
- [ ] Test usecase `TrackSubscriberSessions` table-driven dengan fake streamer + fake repo: connect menghasilkan `OpenSession`, disconnect menghasilkan `CloseSession`, username tak dikenal → `subscription_id` null tanpa gagal. Berkas `internal/usecase/network/track_subscriber_sessions_test.go`.
- [ ] Test repository Postgres memakai `testcontainers-go` (bukan mock): open→close→list, `FindOpenByUsername`, filter aktif per device. Berkas `internal/adapter/postgres/subscriber_session_repository_test.go`.
- [ ] Test handler memakai `httptest`: 200 dengan body & pagination benar, 404 resource tak ada, 403 role tidak berizin. Berkas `internal/adapter/http/subscriber_session_handler_test.go`.
- [ ] `require` untuk precondition, `assert` untuk pengecekan independen.

**Files likely touched:** `internal/driver/mikrotik/sessions_test.go`, `internal/usecase/network/track_subscriber_sessions_test.go`, `internal/adapter/postgres/subscriber_session_repository_test.go`, `internal/adapter/http/subscriber_session_handler_test.go`.

**Dependencies:** Task 3, Task 4, Task 6, Task 7.

**Estimated scope:** Large

---

## Migrasi Database

**Tidak ada perubahan skema.** Tabel `subscriber_sessions` sudah dibuat migrasi **000012** dan `device_stream_subscriptions` sudah dibuat migrasi **000021**; isu ini hanya membaca/menulis baris, tidak menambah tabel/kolom. Tidak menambah nomor migrasi di isu ini (lihat tabel reservasi README §K6 untuk nomor yang sudah dialokasikan issue lain).

Kolom `subscriber_sessions` (§6.4) yang dipakai isu ini sebagai referensi baca: `id`, `subscription_id` (FK, nullable untuk username tak dikenal), `device_id` (FK), `pppoe_username`, `framed_ip`, `mac_address`, `started_at`, `stopped_at` (nullable), `bytes_in`, `bytes_out`, `terminate_cause`. Bila saat implementasi ditemukan salah satu kolom di atas belum ada di 000012, **jangan** menambal diam-diam — hentikan, catat ketidakcocokan sebagai temuan, dan usulkan migrasi baru bernomor bebas berikutnya sesuai tabel reservasi README §K6 (up/down berpasangan) beserta pembaruan `DATABASE-SCHEMA.md` §6.4 pada PR yang sama.

## Verification

- [ ] `go build ./...` sukses tanpa error.
- [ ] `go vet ./...` bersih; assertion compile-time `var _ port.SubscriberSessionStreamer = (*mikrotik.Driver)(nil)` lolos.
- [ ] `go test ./internal/domain/subscribersession/... ./internal/driver/mikrotik/... ./internal/usecase/network/... ./internal/adapter/postgres/... ./internal/adapter/http/...` hijau (repo & driver test butuh Docker untuk testcontainers).
- [ ] `make lint` (golangci-lint) bersih — cek penamaan akronim (`framedIP`, `macAddress`), doc comment exported, dan boundary import.
- [ ] Smoke test manual: jalankan server terhubung ke MikroTik CHR (GNS3), lakukan dial/putus PPPoE dari sisi client, lalu `curl` `GET /api/v1/sessions/active` (dengan token role staff) — pelanggan muncul saat connect dan hilang dari daftar aktif saat disconnect; `GET /api/v1/subscriptions/:id/sessions` menampilkan baris tertutup dengan `stopped_at` terisi.
- [ ] `curl` `GET /api/v1/devices/:id/sessions/active` dengan token `teknisi` = 200; dengan token `staff` sesuai keputusan RBAC (403 bila device-scoped dibatasi teknisi+).
- [ ] Verifikasi tidak ada polling monitoring: pantau log/koneksi — tidak ada `/ppp active print` berulang pada interval pendek (mis. 30s ala referensi); yang boleh muncul hanya satu `/log follow` per device, satu snapshot per stream-connect, dan reconciliation pass pada interval longgar (default 5 menit).
- [ ] Smoke test reconciliation: matikan stream sementara (atau simulasikan putus) saat sebuah sesi PPPoE putus di sisi router, biarkan reconnect + reconciliation berjalan, lalu pastikan sesi yang event disconnect-nya terlewat akhirnya tertutup di DB (`stopped_at` terisi, `terminate_cause` sintetik) — tidak menggantung `stopped_at IS NULL` selamanya.

## Definition of Done

- [ ] Domain `subscribersession`, port repository + streamer, dan implementasi Postgres/MikroTik selesai sesuai boundary AGENTS.md §1 (pengetahuan log vendor hanya di `internal/driver/mikrotik/`).
- [ ] Consumer per-device berjalan dari `main.go` di bawah errgroup/supervisor, reconnect otomatis, mencatat `device_stream_subscriptions`, tanpa goroutine bocor saat shutdown.
- [ ] `subscriber_sessions` terisi real-time dari event stream; byte/cause terisi saat disconnect; tanpa polling berkala (ADR 0003 dipatuhi).
- [ ] Tiga endpoint REST berfungsi dengan RBAC benar dan status code sesuai konvensi.
- [ ] ADR 0011 dibuat dan ditautkan dari `README.md` root pada commit yang sama.
- [ ] Semua test (parser, usecase, repo testcontainers, handler httptest) hijau; `go build`, `go vet`, `make lint` bersih.
- [ ] Tidak ada perubahan skema; bila kolom kurang, ditangani lewat migrasi baru + update `DATABASE-SCHEMA.md`, bukan tambalan diam-diam.
