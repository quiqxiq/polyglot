# MIKHMON v4 → POLYGLOT (CONNECTRPC) — DOKUMEN MASTER

> **Hasil reverse-engineering** terhadap source code Mikhmon v4 (PHP + RouterOS API) untuk diimplementasikan ulang di dalam **proyek Polyglot** (Go, Clean Architecture).
>
> Semua klaim diverifikasi langsung terhadap file source: `index.php`, `core/*`, `config/*`, `get/*.php`, `post/*.php`, `view/*.php`, dan `template/*.txt`.
>
> **Catatan transport:** dokumen ini awalnya ditulis untuk target **REST API**. Karena proyek Polyglot memakai **ConnectRPC** (`connectrpc.com/connect` + `net/http.ServeMux` Go 1.22+, JSON codec — lihat ADR `docs/adr/0005-migrasi-dari-gin-ke-net-http-servemux.md`), seluruh "endpoint REST" pada modul sudah dipetakan ulang ke **prosedur ConnectRPC** dengan path file aktual dan status implementasi.

---

## Cara menggunakan dokumen ini

Spesifikasi dipecah **per-modul**. Setiap file modul berisi: pemetaan legacy → prosedur ConnectRPC, request/response (proto), DTO/tipe aktual, status implementasi, dan logika khusus yang wajib dipertahankan.

| # | Modul | File | Status | Konten utama |
| :-- | :-- | :-- | :-- | :-- |
| 1 | Auth & Admin | [`01-auth-admin.md`](01-auth-admin.md) | ✅ | Login/logout JWT, `me`, refresh, ganti credential, user & RBAC |
| 2 | Device / Router Instance | [`02-router-instance.md`](02-router-instance.md) | ✅ | CRUD device, test koneksi, upload logo, config legacy + enkripsi |
| 3 | Dashboard & Monitoring | [`03-dashboard-monitoring.md`](03-dashboard-monitoring.md) | 🟡 | System resource, hotspot summary, traffic, log |
| 4 | Hotspot User | [`04-hotspot-user.md`](04-hotspot-user.md) | ✅ | CRUD user, reset counter, logika comment `vc-`/`up-` |
| 5 | Hotspot Profile | [`05-hotspot-profile.md`](05-hotspot-profile.md) | ✅ | CRUD profile + generator/parser on-login script |
| 6 | Active / Host / Server | [`06-active-host-server.md`](06-active-host-server.md) | ✅ | List active, kick, host, server hotspot, DHCP lease |
| 7 | Voucher Generator | [`07-voucher-generator.md`](07-voucher-generator.md) | ✅ | Generate batch (vc/up, char-type), query batch, formatter |
| 8 | Sales Report | [`08-sales-report.md`](08-sales-report.md) | ✅ | ListReports (filter day/month/year, summary_only), DeleteReport, fix filter gateway |
| 9 | Expire Monitor | [`09-expire-monitor.md`](09-expire-monitor.md) | ✅ | Status (kompatibel 2 bentuk scheduler), Setup (idempoten), Disable, Remove |
| 10 | Template & Print | [`10-template-print.md`](10-template-print.md) | ✅ | ListTemplates, GetTemplateSection (read-only), RenderVouchers (single/batch/preview, QR login URL) |
| 11 | Resources & Theme | [`11-resources-theme.md`](11-resources-theme.md) | 🟡 | Interface, IP pool, parent queue, NAT, theme |

> **📋 Rencana kerja gap modul 04–10 (termasuk perluasan proto `Device`):** lihat [`IMPLEMENTATION-PLAN-04-10.md`](IMPLEMENTATION-PLAN-04-10.md) — berisi fase per modul, definisi proto lengkap, file yang dibuat/diubah, logika wajib 1:1 dengan legacy, dan acceptance criteria.

Legend: ✅ sudah lengkap · 🟡 sebagian (gateway/usecase ada, prosedur Connect belum diekspos) · 🔴 rencana.

---

## 1. Ringkasan Eksekutif

Mikhmon v4 adalah aplikasi **MikroTik Hotspot Monitor** berbasis PHP monolitik yang:

- Menyimpan konfigurasi router di **file teks** `config/config.php` dengan pemisah string khusus (13 separator).
- Berkomunikasi dengan router MikroTik via **RouterOS API socket TCP port 8728/8729** (bukan HTTP).
- Tidak memiliki REST API — "API"-nya adalah **sekumpulan file PHP** (`get/*.php` dan `post/*.php`) yang dipanggil via AJAX dengan `$_GET` / `$_POST`, sebagian besar response di-obfuscate **XOR 25 (`jsEncode`)**.
- Menyimpan laporan penjualan **di dalam MikroTik** melalui `/system/script` (bukan database).
- Menyematkan metadata bisnis (harga, masa aktif, mode expire) ke dalam **on-login script** User Profile MikroTik.

**Target transformasi (Polyglot):** prosedur ConnectRPC stateless dengan JWT, inventaris device di **PostgreSQL** + credential di **vault AES**, koneksi RouterOS per-request via `port.DeviceDriver` dari `registry`, dan kompatibilitas penuh dengan skrip MikroTik yang sudah terpasang di router (on-login, scheduler `Mikhmon-Expire-Monitor`, log script penjualan).

**Status saat ini:** mayoritas logika Mikhmon sudah hidup di `internal/driver/mikrotik/hotspot/` (gateway + comment + profile + voucher + expire + report), diorkestrasi `internal/usecase/hotspot/`, dan sebagian diekspos via `HotspotService` ConnectRPC di `internal/adapter/connect/hotspot/`. Gap utama: ekspos prosedur yang masih di gateway/usecase, template print, dan migrasi legacy (lihat §7 Roadmap).

---

## 2. Arsitektur Aplikasi Legacy (PHP)

### 2.1 Routing URL

`index.php` mem-parsing `REQUEST_URI` menjadi `?<m_user>/<page>` atau `?admin/<page>`:

```
?<m_user>            → dashboard router (view/dashboard.php)
?<m_user>/<page>     → page router (lihat tabel inventaris)
?admin/<page>        → page admin (settings, template_editor, about, login, set_theme, vpreview)
```

Route map didefinisikan di `config/page.php` (`$m_user_page` dan `$admin_page`), di-validasi oleh `core/route.php` (e404 bila page tidak dikenal, maksimal 3 segmen path).

**Poin penting:** seluruh "endpoint" legacy adalah **operasi CRUD terhadap RouterOS** yang dibungkus file PHP. Tidak ada autentikasi per-request selain session `$_SESSION["mikhmon"]` (login admin). Di Polyglot, setiap operasi menjadi prosedur ConnectRPC yang di-auth JWT + RBAC (Casbin).

### 2.2 Format File Konfigurasi

`config/config.php` berisi satu baris per router dengan separator unik:

```php
$data['mikhmon']  = array('1' => 'mikhmon<|<mikhmon', 'mikhmon>|>aWNlbA==');
$data['router01'] = array('1'=>'router01!192.168.88.1:8728',
                          '2'=>'router01@|@api_user',
                          '3'=>'router01#|#cGFzcw==',
                          '4'=>'router01%WIFI BERKAH',
                          '5'=>'router01^wifi.net',
                          '6'=>'router01&Rp',
                          '7'=>'router01*08123456789',
                          '8'=>'router01(admin@wifi.net',
                          '9'=>'router01)546572696d61204b61736968',
                          '10'=>'router01=30',
                          '11'=>'router01@!@enable',
                          '12'=>'router01#!#');
```

| Separator | Makna | Contoh |
| :--- | :--- | :--- |
| `<|<` / `>|>` | Username / Password admin (baris `mikhmon`) | `mikhmon<|<mikhmon`, `mikhmon>|>aWNlbA==` |
| `!` | IP:port MikroTik | `router01!192.168.88.1:8728` |
| `@\|@` | Username API MikroTik | `router01@|@api_user` |
| `#\|#` | Password API MikroTik (terenkripsi) | `router01#|#cGFzcw==` |
| `%` | Hotspot Name (branding voucher) | `router01%WIFI BERKAH` |
| `^` | DNS Name (hostname hotspot) | `router01^wifi.net` |
| `&` | Currency / mata uang | `router01&Rp` |
| `*` | Phone (kontak di voucher) | `router01*08123456789` |
| `(` | Email admin | `router01(admin@wifi.net` |
| `)` | Info LP (info login page, **hex-encoded**) | `router01)546572696d61204b61736968` |
| `=` | Idle timeout (menit / `disable`) | `router01=30` |
| `@!@` | Live report (`enable` / `disable`) | `router01@!@enable` |
| `#!#` | Token (cadangan) | `router01#!#` |

> **Di Polyglot:** file ini **tidak dipakai** — inventaris device ada di tabel `devices` PostgreSQL (`internal/adapter/postgres/device_repository.go`) dengan credential di vault (`CredentialVault`). Parser separator tetap berguna untuk **migrasi** data legacy (modul 02 §4.3).

### 2.3 Skema Encoding / Obfuscation

**a) `decode()` — `core/page_route.php`** (input form router dari JS):
```
decode(data) = base64_decode( base64_decode( XOR(data, 10) ) )
```

**b) `enc_rypt()` / `dec_rypt()` — `core/routeros_api.class.php`** (password host & admin):
```
enc_rypt(str) = base64_encode( setiap char: ord(char) + ord(keychar) ), key = "128" (berulang)
dec_rypt(str) = kebalikannya (kurangi)
```

**c) `jsEncode` — `core/jsencode.class.php`** (obfuscate response JSON):
```
encodeString(str, 25) = XOR setiap char dengan 25
```

> **Di Polyglot:** response ConnectRPC dikirim **JSON murni** (tanpa obfuscation). Fungsi `decode`/`enc_rypt`/`dec_rypt` hanya diperlukan untuk alat **migrasi** config legacy (modul 02 §4.3) — belum diimplementasikan.

### 2.4 Client RouterOS API

`core/routeros_api.class.php` mengimplementasikan protokol RouterOS API:
- Handshake login `/login` (metode lama challenge `=response=00...md5` dan metode baru post-6.43).
- `comm($command, $params)` mengirim command + argument (`=key=value`), query (`?key=value`), regex (`~key=value`).
- `parseResponse()` mengubah wire protocol (`!re`, `!done`, `!trap`, `!fatal`) menjadi array.
- `!trap` → `!trap[0]['message']` berisi pesan error RouterOS.

> **Di Polyglot:** protokol RouterOS dihandle driver `internal/driver/mikrotik/` (Dial/command). Eksekusi command di-orkestrasi `usecase/network.ExecuteCommand` (policy gate: klasifikasi → approval destructive). Karena koneksi dibuat per-request via `port.DeviceDriver`, **tidak perlu connection pool persisten ala draft awal** — state dibawa oleh driver per request.

---

## 3. Inventaris API Surface Asli (Request/Response Aktual)

Acuan "dari mana" setiap prosedur ConnectRPC target berasal.

### 3.1 Tabel Lengkap Endpoint GET

URL pattern: `?<m_user>/<page>` → file di kolom *Handler*. Semua butuh session login.

| Page (URL) | Handler (file) | Parameter | Command RouterOS | Format Response Asli |
| :--- | :--- | :--- | :--- | :--- |
| `users` | `get/get_users.php` | `prof` (profile), `f` (force), `c` (cache/count) | `/ip/hotspot/user/print ?profile=<prof>` | jsEncode25 dari array user; jika `c` ada → **angka** (count) |
| `user` | `get/get_user.php` | `id` (`.id`) **atau** `name` | `/ip/hotspot/user/print ?.id=<id>` / `?name=<name>` | jsEncode25 dari array |
| `profiles` | `get/get_profiles.php` | `f` | `/ip/hotspot/user/profile/print` | jsEncode25 dari array |
| `profile` | `get/get_profile.php` | `id` **atau** `name` | `/ip/hotspot/user/profile/print ?.id` / `?name` | **JSON polos** |
| `get_sys_resource` | `get/get_dashboard.php` | – | `/system/clock/print`, `/system/resource/print`, `/system/routerboard/print`, `/system/identity/print`, `/system/health/print` | jsEncode25 dari `{systime, resource, syshealth, model, identity}` |
| `get_hotspotinfo` | `get/get_dashboard.php` | – | `/ip/hotspot/user/print =count-only`, `/ip/hotspot/active/print =count-only` | jsEncode25 dari `{hotspot_users: count-1, hotspot_active: count}` |
| `get_traffic` | `get/get_dashboard.php` | `iface` | `/interface/monitor-traffic =interface=<iface> =once` | **JSON polos** `{tx, rx}` (bits/sec) |
| `get_log` | `get/get_dashboard.php` | `f` (force) | pastikan logging prefix `->` (kalau belum: `/system/logging/add =action=disk =prefix=-> =topics=hotspot,info,debug`); lalu `/log/print ?topics=hotspot, info, debug` (di-reverse) | jsEncode25 dari array (di-reverse) |
| `get_livereport` | `get/get_dashboard.php` | `day`, `month`, `f` | `/system/script/print ?owner=<month>` (+ `count-only`) | jsEncode25 dari array |
| `get_report` | `get/get_report.php` | `day`, `f` | `/system/script/print ?source=<day>` (+ `count-only`) | jsEncode25 dari array |
| `get_hotspot_server` | `get/get_hotspot_server.php` | `f` | `/ip/hotspot/print` | jsEncode25 dari array |
| `get_hotspot_active` | `get/get_hotspot_active.php` | – | `/ip/hotspot/active/print` | jsEncode25 dari array |
| `get_hosts` | `get/get_hosts.php` | – | `/ip/hotspot/host/print` | jsEncode25 dari array |
| `get_expire_mon` | `get/get_expire_mon.php` | – | `/system/scheduler/print ?name=Mikhmon-Expire-Monitor ?disabled=false` | **JSON polos** `{"expire_monitor":"ok"}` / `{"expire_monitor":"not ready"}` |
| `get_tot_users` | `get/get_tot_users.php` | – | `/ip/hotspot/user/print =count-only` | **JSON polos** `{"users": count-1}` |
| `get_interface` | `get/get_interface.php` | – | `/interface/print` | jsEncode25 dari array |
| `get_addr_pool` | `get/get_addr_pool.php` | `f` | `/ip/pool/print` | **JSON polos** |
| `get_parent_queue` | `get/get_parent_queue.php` | `f` | `/queue/simple/print ?dynamic=false` | **JSON polos** |
| `get_nat` | `get/get_nat.php` | – | `/ip/firewall/nat/print` | **JSON polos** |
| `connect` | `get/get_connect.php` | – | koneksi + handshake login (debug mode) | **Teks**: `Connected,` / `Invalid username or password,` / `Error,` |
| `set_theme` (admin) | `config/settheme.php` | `theme` ∈ dark/light/blue/green/pink | – (menulis `config/theme.php`) | HTML/redirect |

### 3.2 Tabel Lengkap Endpoint POST

Semua POST butuh `$_SESSION["mikhmon"]`. Hampir semua membawa `sessname` berisi `?<m_user>` (contoh: `?router01`).

| Handler (file) | Trigger field | Field Request | Command RouterOS | Response Asli |
| :--- | :--- | :--- | :--- | :--- |
| `post/post_a_router.php` | `router_` + `do=add` | `router_` (diawali `sess_`) | – | `{"message":"Success","sesname":"sessionXXX<rand>"}` atau pesan error write |
| `post/post_a_router.php` | `do=remove` | `router_` (session) | – | `{"message":"Success"}` + hapus `assets/img/logo-<session>.png` |
| `post/post_a_router.php` | `do=save` | `router_`, `session`, `ipmik`(encode), `usermik`(encode), `passmik`(encode), `hotspotname`, `dnsname`, `currency`, `email`(encode), `infolp`, `idleto`, `phone`(encode), `report` | – | `{"message":"Success","sess":"<newname>"}` atau pesan session duplikat/diperbaiki |
| `post/post_a_router.php` | `do=saveAdmin` | `username`(encode), `password`(encode) | – | `{"message":"Success"}` |
| `post/post_add_user.php` | `name` | `sessname`, `server`, `name`, `password`, `profile`, `macaddr`, `timelimit`, `datalimit`, `comment` | `/ip/hotspot/user/add` (comment otomatis `vc-`/`up-`) | `{"message":"success","data":<user>}` atau `{"message":"error","data":{"error":<trap msg>}}` |
| `post/post_update_user.php` | `name` | `sessname`, `uid`, `reset` (`yes`/`no`), `server`, `name`, `password`, `profile`, `macaddr`, `timelimit`, `datalimit`, `comment`, `expdate`, `ucode` | `/ip/hotspot/user/reset-counters` (jika reset=yes) → `/ip/hotspot/user/set` → `/ip/hotspot/user/print ?.id` | `{"message":"success","data":<user>}` / error |
| `post/post_add_userprofile.php` | `name` | `sessname`, `name`, `addresspool`, `sharedusers`, `ratelimit`, `parentqueue`, `expmode`, `validity`, `price`, `sellingprice`, `lockuser`, `lockserver` | `/ip/hotspot/user/profile/add` (dengan on-login script) | `{"message":"success","data":<profile>}` / error |
| `post/post_update_userprofile.php` | `name` | + `upid` (`.id` profile) | `/ip/hotspot/user/profile/set` → print | `{"message":"success","data":<profile>}` / error |
| `post/post_generate_voucher.php` | `qty` | `sessname`, `qty`, `server`, `user` (`vc`/`up`), `userl`, `prefix`, `char`, `profile`, `timelimit`, `datalimit`, `gcomment`, `gencode` | `/ip/hotspot/user/add` × qty → `/ip/hotspot/user/print ?comment=<commt>` | `{"message":"success","data":{"count":N,"comment":<commt>,"profile":...}}` / error |
| `post/post_cache_voucher.php` | `qty` | `sessname`, `user`, `gcomment`, `gencode` | `/ip/hotspot/user/print ?comment=<commt> ?uptime=0s` | `{"message":"success","data":{"count":N,"comment":<commt>}}` |
| `post/post_hotspot_remove.php` | `sessname` | `sessname`, `where` ∈ `user_`/`profile_`/`active_`/`host_`, `id` | `/ip/hotspot/user/remove` / `user/profile/remove` / `active/remove` / `host/remove` | `{"message":"success"}` atau `{"message":"error"}` (gagal konek) |
| `post/post_expire_monitor.php` | `sessname` | `sessname`, `expmon` (source script) | `/system/scheduler/print ?name=Mikhmon-Expire-Monitor` → `/system/scheduler/add` (interval `00:01:00`, start `00:00:00`, disabled no) atau `/system/scheduler/set` bila disabled | `{"message":"success"}` atau `{"message":"<name>"}` |
| `post/post_template.php` | `router_` | `do=saveTemplate`, `_template` (isi), `file_` (path template) | – (menulis `template/<file_>`) | `{"message":"Saved"}` / pesan error |
| `post/post_logout.php` | `logout` | `logout` | – | echo nilai `logout` (teks) |

### 3.3 Alur Login & Logout (Session)

- **Login** — `view/login.php` (route `?admin/login`): form POST `user`, `pass`, `login`. Validasi: `user === $useradm && pass === dec_rypt($passadm)` → `$_SESSION["mikhmon"] = $user` → redirect `?admin/settings`.
- **Logout** — `post/post_logout.php` → `session_destroy()`.
- **Otentikasi halaman** — `index.php`: bila `!isset($_SESSION["mikhmon"])` → route ke `admin/login`.

> **Di Polyglot:** diganti `AuthService` (JWT access + refresh token cookie + logout revoke) dan `UserService` (multi-user + RBAC Casbin). Default credential legacy (username `mikhmon`, password `mikhmon`) hanya relevan untuk migrasi (modul 01 §4).

---

## 4. Format Standar Response & Error

Proyek ini memakai protobuf + ConnectRPC, bukan envelope REST. Lihat pemetaan lengkap di [`ANALISIS_REST_API_MIKHMON_GOLANG.md`](../../ANALISIS_REST_API_MIKHMON_GOLANG.md) §"Standar Response & Error".

**Ringkasnya:**
- Sukses → proto `*Response` (JSON via `JSONCodec()`).
- Error → `connect.Error` dengan kode standar; semua handler wajib `response.MapDomainError(err)` (`pkg/response/errors.go`).
- Kode lama (`ROUTEROS_CONNECTION_FAILED`, `ROUTEROS_TRAP`, `UNAUTHORIZED`, dst.) dipetakan ke `connect.Code` (Internal, Unauthenticated, dst.).

---

## 5. Struktur Proyek Aktual (Polyglot)

```text
polyglot/                                  # ← root repo (bukan mikhmon-backend-go/)
├── api/proto/v1/
│   ├── hotspot.proto                      # HotspotService (dashboard, users, profiles, active, voucher, streaming)
│   ├── device.proto                       # DeviceService (inventaris + test + streaming)
│   └── auth.proto / users.proto           # AuthService, UserService, RBACService
├── internal/
│   ├── domain/                            # entity murni (device, command, ...) — bebas proto/adapter
│   ├── port/
│   │   ├── hotspot_gateway.go             # ★ HotspotGateway (kontrak operasi hotspot/voucher)
│   │   ├── device_driver.go               # DeviceDriver / StreamingDeviceDriver
│   │   └── ...                            # repository, auth_service, credential_vault, dst
│   ├── usecase/
│   │   ├── hotspot/hotspot_usecase.go     # ★ orkestrasi hotspot & voucher (gateway saja)
│   │   ├── device/manage_device.go        # CRUD inventaris + vault
│   │   ├── auth/                          # login, refresh token
│   │   └── network/                       # ExecuteCommand (policy gate), ActiveSessions, OpenTerminal
│   ├── driver/
│   │   ├── mikrotik/
│   │   │   ├── hotspot/                   # ★★ Logika Mikhmon:
│   │   │   │   ├── gateway.go             #     implementasi port.HotspotGateway
│   │   │   │   ├── comment.go             #     parser comment vc-/up- (pre/post login)
│   │   │   │   ├── profile.go             #     builder/parser on-login script + expire mode
│   │   │   │   ├── voucher.go             #     generator batch + char-set + ParseDataLimit
│   │   │   │   ├── expire.go              #     BuildExpireMonitorScript + scheduler
│   │   │   │   └── report.go              #     parser transaksi "-|-"
│   │   │   └── ...                        #     driver RouterOS lain (system, dhcp, traffic, ...)
│   │   └── genieacs/ genericssh/ ...      # driver vendor lain
│   ├── adapter/
│   │   ├── connect/hotspot/               # ★★ Handler ConnectRPC HotspotService
│   │   │   ├── user_handler.go            #     GetUser, CreateUser, UpdateUser, ResetUserCounters, DeleteUser
│   │   │   ├── profile_handler.go         #     CreateProfile, UpdateProfile, DeleteProfile
│   │   │   ├── profile_user_handler.go    #     ListProfiles, ListUsers (filter comment/only_unused)
│   │   │   ├── host_server_handler.go     #     ListHosts, RemoveHost, ListHotspotServers
│   │   │   ├── voucher_handler.go         #     GenerateVouchers, GetVoucherBatch
│   │   │   ├── session_handler.go         #     ListActiveSessions, Kick, DHCP
│   │   │   ├── report_handler.go          #     ListReports, DeleteReport
│   │   │   ├── expire_monitor_handler.go  #     GetExpireMonitorStatus, Setup, Disable, Remove
│   │   │   ├── template_handler.go        #     ListTemplates, GetTemplateSection, RenderVouchers
│   │   │   ├── monitor_system_handler.go  #     StreamSystemSnapshot (5-stream), StreamResource
│   │   │   ├── monitor_interface_handler.go #   StreamTraffic, StreamInterfaceEthernet
│   │   │   ├── monitor_queue_handler.go   #     StreamQueueStats
│   │   │   ├── monitor_log_handler.go     #     StreamLogs
│   │   │   ├── monitor_session_handler.go #     StreamActiveSessions, StreamHotspotInactive
│   │   │   ├── monitor_ppp_handler.go     #     StreamPPPActive, StreamPPPInactive
│   │   │   ├── mapper*.go / router.go
│   │   ├── connect/device/                # Handler ConnectRPC DeviceService
│   │   ├── connect/auth/                  # AuthService/UserService/RBACService
│   │   ├── ws/                            # SSE hub (/events) + terminal WebSocket
│   │   ├── http/middleware/               # Chain, JWT, RBAC, CORS, Logger, Recovery
│   │   └── postgres/ redis/ auth/ vault/  # adapter infrastruktur
│   ├── registry/registry.go               # DriverFactory (mikrotik, genieacs, ...)
│   └── app/app.go                         # Composition root — mount semua service
├── internal/template/                     # header/row/footer × default/small/thermal (+ embed.go)
├── pkg/response/errors.go                 # MapDomainError (Connect)
├── migrations/                            # skema DB (devices, users, ...)
└── docs/                                  # dokumentasi arsitektur + database-schema.md
```

**Perbedaan dengan draft awal (`mikhmon-backend-go/`):**
- ❌ Tidak ada `internal/delivery/http/` + handler REST → ✅ `internal/adapter/connect/<domain>/` (ConnectRPC).
- ❌ Tidak ada `internal/repository/mikrotik/client_pool.go` → ✅ koneksi per-request via `registry` + `port.DeviceDriver`.
- ❌ Tidak ada `internal/repository/db/` terpisah → ✅ repository Postgres di `internal/adapter/postgres/`.
- ❌ Tidak ada `internal/dto/` → ✅ tipe di proto (`api/gen/v1`) + `mapper.go` per domain; domain model murni di `internal/domain/`.
- ✅ `templates/` → `internal/template/` (embed via `embed.go`).
- ✅ `pkg/generator`, `pkg/mikhmonscript`, `pkg/legacy` → logika Mikhmon di `internal/driver/mikrotik/hotspot/` (comment/profile/voucher/expire/report). Helper `pkg/voucher/` ada untuk generate.

---

## 6. Connection Pooling & RouterOS Client

Draft awal merencanakan `RouterPool` + `RouterClient` (satu socket = satu command). **Di Polyglot ini tidak dipakai** — keputusan arsitektur (ADR `0002-devicedriver-tanpa-session-terpisah.md`):

- Setiap request mendapat `port.DeviceDriver` segar dari `registry.Registry.Get(ctx, deviceID)`; credential di-dekripsi dari vault saat pembuatan driver.
- Command dieksekusi melalui `usecase/network.ExecuteCommand` (policy gate: klasifikasi command → approval untuk yang destruktif). Gateway hotspot menerima `port.CommandExecutor` (lihat `internal/driver/mikrotik/hotspot/gateway.go`).
- Streaming (traffic/resource/active/log/queue/system snapshot/inactive) memakai `port.StreamingDeviceDriver` + command streaming native RouterOS (`monitor-traffic`, `print follow`, `print interval=1s`) via ConnectRPC server-streaming. Tidak ada polling backend — RouterOS mengirim data berulang.

Masih relevan dari draft awal: cek `reply.Done` + `reply.Trap` untuk menerjemahkan error RouterOS, timeout koneksi ~3 detik, dan batch besar sebaiknya async.

---

## 7. Roadmap Implementasi (disesuaikan kondisi proyek)

### ✅ Sudah selesai
1. Auth JWT + refresh + logout + multi-user + RBAC Casbin (`AuthService`, `UserService`).
2. Inventaris device CRUD + test koneksi + streaming status/traffic/ping + terminal (`DeviceService` + WS).
3. Hotspot core driver: comment parser, on-login script builder, voucher generator, expire monitor script, report parser (`internal/driver/mikrotik/hotspot/`).
4. `HotspotService` (modul 03–10): ListProfiles, ListUsers (+filter comment/only_unused), ListActiveSessions, KickActiveSession, ListDHCPLeases, BlockDHCPLease, GenerateVouchers, GetVoucherBatch, GetUser, CreateUser, UpdateUser, ResetUserCounters, DeleteUser, CreateProfile, UpdateProfile, DeleteProfile, ListHosts, RemoveHost, ListHotspotServers, ListReports, DeleteReport, GetExpireMonitorStatus, SetupExpireMonitor, DisableExpireMonitor, RemoveExpireMonitor, ListTemplates, GetTemplateSection, RenderVouchers.
   - **Streaming terpisah per-area (modul 03, pengganti GetDashboard — dihapus total):** StreamSystemSnapshot, StreamResource, StreamTraffic, StreamInterfaceEthernet, StreamQueueStats, StreamLogs, StreamActiveSessions, StreamHotspotInactive, StreamPPPActive, StreamPPPInactive.

### ✅ Sudah selesai (batch pertama — modul 04–07)
5. Hotspot user: `GetUser`, `CreateUser`, `UpdateUser` (reset counter + logika comment `expdate`/`ucode`), `DeleteUser`, `ResetUserCounters` — `user_handler.go` + `mapper_user.go`.
6. Hotspot profile: `CreateProfile`, `UpdateProfile`, `DeleteProfile` + parser on-login terstruktur (`ParseOnLoginScript`) — `profile_handler.go` + `mapper_profile.go`.
7. Active/Host/Server: `ListHosts`, `RemoveHost`, `ListHotspotServers` — `host_server_handler.go` + `mapper_host_server.go`.
8. Voucher: `GenerateVouchers` diperluas (server/time_limit/data_limit/comment) + `GetVoucherBatch` (filter comment + `uptime=0s`) — `voucher_handler.go`.

### ✅ Sudah selesai (batch kedua — modul 08–10, Fase 5–7)
9. Sales report: `ListReports` (filter legacy day/month/year + `summary_only`, **fix bug filter gateway** `?owner=mikhmon-report` → `?comment=mikhmon`/`?source=`/`?owner=`), `DeleteReport` — `report_handler.go` + `mapper_report.go`.
10. Expire monitor: `GetExpireMonitorStatus` (kompatibel `Mikhmon-Expire-Monitor` & `mikhmon-expire-scheduler`, status `ok`/`not ready`), `SetupExpireMonitor` (idempoten, default `00:01:00`), `DisableExpireMonitor`, `RemoveExpireMonitor` — `expire_monitor_handler.go` + `internal/port/hotspot_expire.go`.
11. Template & print: `ListTemplates`, `GetTemplateSection` (read-only dari embed), `RenderVouchers` (single `.id` / batch comment+`uptime=0s` / preview dummy; QR mode **login URL**; metadata scope-down: identity router + server + on-login profile) — `template_handler.go` + `pkg/voucher` (`RenderWithOptions`, `QRContent`, `ListTemplates`).

### ✅ Sudah selesai (batch ketiga — modul 03, dashboard streaming terpisah)
12. Dashboard & monitoring: `GetDashboard` (endpoint agregasi) **dihapus total** → diganti **10 prosedur streaming terpisah per-area** (semuanya streaming MikroTik → backend → frontend, tanpa polling):
    - `StreamSystemSnapshot` — gabungan 5 stream native `interval=1s` (clock, resource, routerboard, identity, health) jadi 1 frame.
    - `StreamInterfaceEthernet` (`/interface/ethernet/print interval=1s`), `StreamQueueStats` (`/queue/simple/print stats interval=1s`), `StreamLogs` (`/log/print follow`).
    - `StreamHotspotInactive` (user + active `interval` → selisih), `StreamPPPActive` (follow), `StreamPPPInactive` (secret + active `interval` → selisih).
    - Handler: `monitor_system/interface/queue/log/session/ppp_handler.go`; driver baru: `system_health.go`, `system_routerboard.go` + varian `interval` di `hotspot_user/active.go`, `ppp.go`, `ppp_active.go`, `iface.go`.

### 🟡 Tahap berikutnya
12. Resources: `ListIPPools`, `ListParentQueues`, `ListNATRules`, `ListInterfaces` + theme/settings.
13. Migrasi legacy: parser `config/config.php` (`decode`/`enc_rypt`/`dec_rypt`) → tabel `devices`.
14. Test suite integrasi RouterOS mock + `web/src/gen` (TS) regenerate bila frontend memakai prosedur baru.

---

*Dokumen ini bersifat living document — setiap endpoint yang diimplementasikan sebaiknya diverifikasi ulang terhadap `get/*.php` dan `post/*.php` untuk menjaga kompatibilitas perilaku, dan terhadap file `.go` aktual di `internal/driver/mikrotik/hotspot/` serta `internal/adapter/connect/hotspot/` untuk menjaga sinkronisasi status.*
