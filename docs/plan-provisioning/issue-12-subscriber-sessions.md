# Issue 12: Subscriber Session Tracking

## Konteks

Tabel `subscriber_sessions` (DATABASE-SCHEMA.md §6.4, dibuat oleh migrasi 000012) dirancang untuk merekam kapan seorang pelanggan **online** dan **offline** di jaringan — satu baris per sesi PPPoE, berisi username, IP yang di-assign, MAC, waktu mulai/berhenti, byte counter, dan penyebab putus. Temuan di `ANALISIS-PROVISIONING-REPO-REFERENSI.md` menegaskan bahwa repo referensi mengisi data sejenis dengan **polling** `/ppp active print` berkala (tiap ~30 detik) lalu men-diff snapshot. Pendekatan polling-berkala itu **dilarang** di Polyglot sesuai ADR 0003 — boros koneksi, telat, dan tidak scalable untuk ratusan router.

Isu ini membangun jalur pengisian `subscriber_sessions` dari **event stream native RouterOS**. Kuncinya: di RouterOS API, **setiap command tree yang punya `print` mendukung `follow`/`follow-only`/`interval=`** — dan `print follow` mengubah tabel apa pun menjadi aliran streaming. Jadi **tidak perlu** membaca `/log` dan mem-parsing teks pesan log (`topic ppp,info`): cukup jalankan **`/ppp/active/print follow`**, yang langsung mem-push **record terstruktur** dari tabel active itu sendiri (name, address, caller-id, service, uptime, `.id`) saat sesi masuk/keluar. Ini lebih andal daripada parsing teks log yang formatnya bisa berubah, dan `Stream()` di `internal/driver/mikrotik/stream.go` yang sudah ada memang menerima print-follow apa pun (lihat `commands.go`: `follow`/`follow-only`/`interval` = streaming). Consumer jangka-panjang berjalan per device MikroTik aktif, dipasang dari `main.go` di bawah supervisor stream yang sudah ada, dan reconnect otomatis bila koneksi putus.

**Dua sifat penting `print follow` yang membentuk desain isu ini:**
- **Snapshot awal gratis.** `follow` (bukan `follow-only`) meng-emit **semua entry yang sedang ada lebih dulu** sebagai `!re`, lalu terus streaming perubahan. Jadi state awal (sesi yang sudah online sebelum stream hidup) ikut terkirim — tidak perlu panggilan `/ppp active print` snapshot terpisah. `follow-only` melewati state awal; itulah kenapa kita pakai `follow`.
- **Disconnect = record ber-`.dead=yes`.** Saat sebuah entry active hilang, RouterOS mengirim `!re` dengan `.dead=yes` untuk `.id` itu. Parser memetakan `.id` → sesi terbuka dan menutupnya. Tidak ada pesan log yang perlu di-parse untuk mendeteksi putus.

**Peringatan domain yang mudah tertukar:** `subscriber_sessions` adalah *status pelanggan online*, sama sekali berbeda dari `internal/domain/session` yang berarti *koneksi driver KE perangkat*. Penamaan tabel sengaja `subscriber_sessions`, bukan `sessions`. Semua tipe/berkas di isu ini memakai awalan/nama `subscribersession` agar tidak bertabrakan dengan `domain/session`. Jangan pernah menyatukan keduanya.

## Prasyarat

Harus sudah ada / dipakai sebagai foundation:

- Migrasi **000012** (`subscriber_sessions`) sudah ada — isu ini tidak mengubah skema.
- Migrasi **000021** (`device_stream_subscriptions`) sudah ada — dipakai untuk mencatat stream per device yang aktif.
- Migrasi **000009** (`subscriptions`) dengan kolom `pppoe_username` + `device_id` — kunci pemetaan username → subscription.
- Driver MikroTik lengkap: `internal/driver/mikrotik/driver.go`, `connect.go` (`dialAndLogin`), `stream.go` (implementasi `port.StreamingDeviceDriver.Stream()` — menerima **print-follow apa pun**, mis. `/ppp/active/print follow`, bukan hanya `/log`).
- `internal/port/streaming_driver.go` (`port.StreamingDeviceDriver`) dan `internal/registry/registry.go` (`Get(ctx, deviceID)` mengembalikan `port.DeviceDriver` yang di-cache per device).
- Supervisor stream device + pola errgroup/cancel yang sudah dipakai fondasi `stream_output.go` dan `main.go` (reconnect via supervisor yang sudah ada).
- RBAC Casbin (`configs/rbac_policy.csv`) dan middleware auth/rbac (`internal/adapter/http/middleware/`).

Tidak bergantung pada Issue 01 (Sync Engine): isu ini hanya **membaca** event dari perangkat dan menulis ke `subscriber_sessions`, tidak menulis `provisioning_sync_log` maupun mengeksekusi command destruktif — sehingga alur sinkronisasi K4 tidak berlaku di sini.

## Ruang Lingkup

**In scope**

- Domain `subscribersession` (entity sesi + tipe event connect/disconnect).
- Kontrak `port.SubscriberSessionRepository` + implementasi Postgres (map ke tabel 000012).
- Kontrak `port.SubscriberSessionStreamer` (port baru — diusulkan eksplisit di Task 3) + implementasinya di driver MikroTik. **Tiga sumber stream native**, satu per jenis identitas, semuanya `print follow` (bukan `/log`): `/ppp/active/print follow` (PPPoE), `/ip/hotspot/active/print follow` (hotspot subscription mode A `mac_login`, Issue 13), dan `/tool/netwatch/print follow` (hotspot subscription mode B `ip_binding` — status host up/down atas IP statik yang dipasang `/tool netwatch`). Ketiganya menghasilkan `subscribersession.Event` connect/disconnect yang sama (tambah entry / entry `.dead=yes`) — hanya nama command + pemetaan field yang berbeda, satu mesin monitoring.
- Usecase network consumer jangka-panjang: konsumsi event stream → map identitas (username PPPoE / kode hotspot / IP netwatch) → subscription → tulis `subscriber_sessions`.
- Supervisor per-device di `main.go` (goroutine errgroup/cancel, reconnect, catat `device_stream_subscriptions`). Snapshot state awal **inheren** dari `print follow` (`follow` meng-emit entry yang ada lebih dulu, lalu streaming) — tidak perlu panggilan snapshot terpisah. Reconciliation pass periodik ber-interval longgar hanya sebagai jaring pengaman untuk menutup gap saat stream putus di antara reconnect.
- Usecase + handler REST read-only: riwayat sesi per subscription, daftar pelanggan online (global dan per device).
- ADR 0011 mendokumentasikan keputusan event-stream via **`print follow` native** (bukan `/log` topic-parsing), snapshot inheren dari `follow`, reconciliation pass, dan tiga sumber stream (ppp/hotspot/netwatch active), ditautkan dari `README.md` root.

**Out of scope**

- Polling `/ppp active print` berulang **sebagai mekanisme monitoring utama** (dilarang ADR 0003 — inilah yang dilakukan repo referensi tiap 30s). Di Polyglot sumber kebenaran real-time adalah `/ppp/active/print follow` (satu stream persisten per device, bukan loop). Yang DIIZINKAN dan bukan pelanggaran: reconciliation pass ber-interval longgar (default 5 menit) sebagai jaring pengaman koreksi drift — bukan loop monitoring real-time. (Snapshot state awal tidak lagi butuh panggilan terpisah: sudah inheren dari `follow`.)
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

**Task 3: Port `SubscriberSessionStreamer` + parser record `print follow` di driver MikroTik**

**Description:** Usulkan **eksplisit** satu port baru yang mengubah stream `print follow` dari tabel active menjadi aliran `subscribersession.Event` domain, dan implementasikan parser record RouterOS di dalam driver MikroTik (pengetahuan vendor wajib di `internal/driver/mikrotik/`, bukan di usecase). Sumbernya adalah tabel active native (`/ppp/active`, `/ip/hotspot/active`, `/tool/netwatch`) via `print follow` — **bukan** `/log` topic-parsing.

**Acceptance criteria:**
- [ ] **Proposal layer dinyatakan:** `internal/port/subscriber_session_streamer.go` mendefinisikan `SubscriberSessionStreamer` dengan method (usulan) `StreamSessions(ctx) (<-chan subscribersession.Event, error)`. Ini port baru (bukan layer baru) — dinyatakan sebagai proposal sesuai AGENTS.md §0.4, bukan pola diam-diam.
- [ ] `subscribersession.Event` memuat `Kind` sumber (`pppoe`/`hotspot`/`netwatch`) + identitas yang relevan (username PPPoE / kode hotspot+MAC / IP netwatch) agar usecase (Task 4) bisa memetakan ke subscription lewat kolom yang tepat.
- [ ] Driver MikroTik menambah berkas `internal/driver/mikrotik/sessions.go` (atau `active_events.go`) berisi: (a) parser per-sumber dari satu **record `!re`** menjadi `subscribersession.Event` connect/disconnect; (b) implementasi `StreamSessions` yang membuka **tiga** `Stream()` — `/ppp/active/print follow`, `/ip/hotspot/active/print follow`, `/tool/netwatch/print follow` — (hotspot & netwatch hanya bila device melayani jenis itu; boleh selalu buka ketiganya karena murah), membaca record, memparser sesuai sumber, lalu mengirim event ke satu channel gabungan.
- [ ] **Semantik record `print follow` (kunci parser):** pakai `follow` (bukan `follow-only`) → record entry yang **sudah ada** dikirim lebih dulu (snapshot awal inheren) lalu perubahan. Record penambahan/awal = **connect**; record dengan atribut **`.dead=yes` pada `.id`** = **disconnect** (entry hilang dari tabel). Parser memelihara pemetaan `.id → identitas` dalam-memori per stream agar record `.dead=yes` (yang hanya membawa `.id`) bisa dipetakan balik ke sesi yang dibuka. Tidak ada teks log yang di-parse.
- [ ] Parser `/ppp/active`: ambil field record `name` (→ `pppoe_username`), `address` (→ `framed_ip`), `caller-id` (→ `mac_address`/identitas pemanggil), `service` (pppoe/l2tp/pptp/…), `uptime`, `.id`. Byte counter tidak tersedia di tabel active PPPoE per-sesi secara andal → biarkan null (lihat catatan byte Task 4). `terminate_cause` tidak ada di record `.dead` → set sintetik/null.
- [ ] Parser `/ip/hotspot/active`: ambil `user` (→ username/kode voucher), `address` (→ `framed_ip`), `mac-address` (→ `mac_address`, kunci utama mode A `mac_login`), `login-by` (`cookie`/`http-chap`/`mac`/`trial` — penting deteksi mac-cookie, lihat catatan mac-cookie Task 4), `session-time-left`, `bytes-in`/`bytes-out`, `.id`.
- [ ] Parser `/tool/netwatch`: record memuat `host` (→ IP yang dipantau) dan `status` (`up`/`down`). Perubahan `status=up` → connect, `status=down` → disconnect (mode B tidak punya username; identitas = IP statik yang dipetakan ke subscription oleh Task 4). Netwatch memakai perubahan field `status`, bukan `.dead` (entry netwatch persisten).
- [ ] Field yang tidak tersedia di record dibiarkan null (best-effort, jangan mengarang). Record tak relevan / sentinel (`!done`, `!re` kosong) di-skip tanpa error.
- [ ] `var _ port.SubscriberSessionStreamer = (*Driver)(nil)` sebagai bukti compile-time bahwa `*mikrotik.Driver` memenuhi kontrak.
- [ ] Channel ditutup rapi saat `ctx` dibatalkan atau stream berakhir; tidak ada goroutine bocor (tiga stream + fan-in di-cancel bersama).
- [ ] Tidak ada polling (loop `/ppp active print` / `/ip hotspot active print` / ping berkala dari Golang) di jalur ini — semua via `print follow` persisten (ADR 0003).

**Files likely touched:** `internal/port/subscriber_session_streamer.go`, `internal/driver/mikrotik/sessions.go`, (mungkin sedikit) `internal/driver/mikrotik/driver.go`.

**Dependencies:** Task 1.

**Estimated scope:** Large

---

**Task 4: Usecase consumer `TrackSubscriberSessions`**

**Description:** Buat orkestrasi network jangka-panjang untuk satu device: konsumsi `subscribersession.Event`, resolusi username→subscription, dan tulis ke repository.

**Acceptance criteria:**
- [ ] Berkas `internal/usecase/network/track_subscriber_sessions.go` berisi `func TrackSubscriberSessions(...)` yang menerima `ctx`, satu `port.SubscriberSessionStreamer` (diperoleh via type-assert dari `registry.Get`), `port.SubscriberSessionRepository`, dan `port.SubscriptionRepository`.
- [ ] Loop membaca channel event sampai `ctx` selesai; setiap event `EventConnect` → resolusi `subscription_id` dari identitas (lihat bawah) lalu `OpenSession`; setiap `EventDisconnect` → temukan sesi terbuka yang cocok lalu `CloseSession` mengisi `stopped_at`/byte/cause. Karena record `.dead=yes` hanya membawa `.id`, pemetaan `.id → sesi terbuka` dijaga oleh streamer (Task 3) sehingga event disconnect yang sampai ke usecase sudah membawa identitas sesi yang benar (bukan hanya `.id` mentah).
- [ ] Resolusi identitas → subscription tergantung `Event.Kind`: `pppoe`/`hotspot` via username (usulan tambah `FindByPPPoEUsername`/`FindByHotspotIdentity(ctx, code, deviceID)` bila belum ada — nyatakan sebagai penambahan method); `netwatch` (mode B `ip_binding`) via `FindByStaticIP(ctx, ip, deviceID)` yang mencocokkan `subscriptions.static_ip` + `hotspot_access_mode='ip_binding'`. Identitas tak dikenal tetap dicatat dengan `subscription_id` null + log peringatan, tidak menggagalkan consumer.
- [ ] Usecase **tidak** meng-import package driver mana pun secara langsung (hanya `port` + `domain`) — boundary K1 terjaga; deteksi vendor terjadi di supervisor/registry, bukan di sini.
- [ ] Kegagalan menulis satu event di-log dan di-skip; consumer tidak mati karena satu error tulis (kecuali `ctx` dibatalkan).
- [ ] **Catatan byte:** dokumentasikan (komentar singkat merujuk ADR 0011) bahwa byte counter berbeda per sumber: tabel `/ppp/active` **tidak** mengekspos `bytes-in/out` per-sesi secara andal → PPPoE `bytes_in/out` dibiarkan **null** (byte akurat butuh `/interface` atau `/queue`, di luar lingkup dan akan jadi polling). Tabel `/ip/hotspot/active` **memuat** `bytes-in/out` → hotspot boleh mengisinya dari record terakhir yang terlihat (best-effort, snapshot terakhir sebelum `.dead`, bukan real-time). Untuk mode B (`netwatch`) byte counter memang tidak ada → null. Tidak ada polling byte berkala di jalur mana pun.
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
- [ ] **Snapshot awal inheren dari `follow` (bukan panggilan terpisah):** karena `/ppp/active/print follow` (dan hotspot/netwatch setara) memakai `follow` yang meng-emit entry yang **sudah ada** lebih dulu lalu streaming, state awal (sesi yang online sebelum stream hidup) ikut terkirim sebagai record connect **tanpa** panggilan `/ppp active print` terpisah. Consumer merekonsiliasi record awal ini dengan sesi terbuka di DB agar tidak menduplikasi baris (sesi yang sudah tercatat open tidak di-insert ulang). Ini yang membuat pendekatan `print follow` superior atas polling+diff 30s ala referensi (lihat REFERENCES.md): satu stream memberi snapshot **dan** delta sekaligus.
- [ ] **Reconciliation pass periodik (jaring pengaman, opsional):** setiap reconnect `print follow` sudah **meng-emit ulang seluruh state** (sifat `follow`), sehingga gap saat stream putus umumnya tertutup otomatis begitu reconnect — reconcile terjadi natural di awal tiap follow. Sebagai lapis tambahan untuk skenario stream "hidup tapi diam-diam basi" (jarang), sediakan reconciliation pass **opsional** ber-interval **rendah** (usulan default 5 menit, default MATI karena reconnect sudah menutup mayoritas gap; dikontrol flag config) yang menjalankan `/ppp active print` **satu kali** untuk mendeteksi (a) sesi yang **hilang** dari router tetapi masih `stopped_at IS NULL` di DB → tutup dengan `terminate_cause` sintetik (mis. `reconciliation`); (b) sesi yang **ada** di router tetapi tidak ada baris open di DB → buka baris baru. Ini **bukan** polling monitoring 30s ala referensi — interval jauh lebih longgar dan opsional; sumber kebenaran real-time tetap `print follow`. Dokumentasikan trade-off di ADR 0011.
- [ ] Reconciliation dikontrol flag config (default: MATI — reconnect menutup gap; nyalakan bila operator ingin lapis ekstra) di `internal/config/config.go`.
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
- [ ] `docs/adr/0011-subscriber-session-via-event-stream.md` dibuat (empat digit, lanjut dari 0005) menjelaskan: alasan menolak polling monitoring 30s ala referensi (ADR 0003), pemilihan **`print follow` native pada tabel active** (`/ppp/active/print follow`, `/ip/hotspot/active/print follow`, `/tool/netwatch/print follow`) alih-alih `/log` topic-parsing (record terstruktur + `.dead=yes`, lebih andal dari teks log), snapshot inheren dari `follow`, port `SubscriberSessionStreamer`, reconciliation pass opsional (mengapa bukan pelanggaran "no polling"), dan batasan byte/cause (PPPoE active tak punya byte/cause; hotspot active punya byte).
- [ ] `README.md` root ditautkan ke ADR 0011 **pada commit yang sama** (AGENTS.md §1.5).
- [ ] ADR merujuk `subscriber_sessions` (§6.4) dan menegaskan perbedaan dengan `domain/session`.

**Files likely touched:** `docs/adr/0011-subscriber-session-via-event-stream.md`, `README.md`.

**Dependencies:** Task 3, Task 5 (agar keputusan final tercermin).

**Estimated scope:** Small

---

**Task 9: Test**

**Description:** Uji parser log, usecase consumer, repository Postgres, dan handler.

**Acceptance criteria:**
- [ ] Test parser MikroTik table-driven: input beberapa **record `!re`** dari `/ppp/active/print follow` (entry baru = connect, entry `.dead=yes` = disconnect, record irelevan/`!done`) → `subscribersession.Event` yang benar, termasuk kasus field byte/cause absen; ditambah kasus hotspot active (dengan `bytes-in/out`, `login-by`) dan netwatch (`status` up/down). Berkas `internal/driver/mikrotik/sessions_test.go`.
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
- [ ] Verifikasi tidak ada polling monitoring: pantau koneksi — tidak ada `/ppp active print` berulang pada interval pendek (mis. 30s ala referensi); yang boleh muncul hanya stream `print follow` persisten per device (`/ppp/active`, dan `/ip/hotspot/active`+`/tool/netwatch` bila relevan) plus — bila diaktifkan — reconciliation pass pada interval longgar (default MATI).
- [ ] Smoke test reconciliation/reconnect: putuskan stream sementara saat sebuah sesi PPPoE berakhir di sisi router, biarkan reconnect — karena `follow` meng-emit ulang state saat reconnect, sesi yang event `.dead`-nya terlewat harus terkoreksi (yang sudah tak ada di router ditutup `stopped_at`), tidak menggantung `stopped_at IS NULL` selamanya.

## Definition of Done

- [ ] Domain `subscribersession`, port repository + streamer, dan implementasi Postgres/MikroTik selesai sesuai boundary AGENTS.md §1 (pengetahuan log vendor hanya di `internal/driver/mikrotik/`).
- [ ] Consumer per-device berjalan dari `main.go` di bawah errgroup/supervisor, reconnect otomatis, mencatat `device_stream_subscriptions`, tanpa goroutine bocor saat shutdown.
- [ ] `subscriber_sessions` terisi real-time dari event stream; byte/cause terisi saat disconnect; tanpa polling berkala (ADR 0003 dipatuhi).
- [ ] Tiga endpoint REST berfungsi dengan RBAC benar dan status code sesuai konvensi.
- [ ] ADR 0011 dibuat dan ditautkan dari `README.md` root pada commit yang sama.
- [ ] Semua test (parser, usecase, repo testcontainers, handler httptest) hijau; `go build`, `go vet`, `make lint` bersih.
- [ ] Tidak ada perubahan skema; bila kolom kurang, ditangani lewat migrasi baru + update `DATABASE-SCHEMA.md`, bukan tambalan diam-diam.
