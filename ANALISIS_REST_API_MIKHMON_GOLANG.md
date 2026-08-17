# ANALISIS MENDALAM & PEMETAAN MIKHMON v4 KE POLYGLOT (CONNECTRPC)

> **Spesifikasi lengkap** hasil reverse-engineering source code Mikhmon v4 (PHP + RouterOS API) untuk diimplementasikan ulang di dalam **proyek Polyglot** (Go, Clean Architecture).
>
> **Catatan penting:** dokumen ini awalnya berjudul *"ANALISIS REST API MIKHMON GOLANG"* karena target awal adalah REST API. Setelah disesuaikan dengan kondisi proyek saat ini, target transport adalah **ConnectRPC** (`connectrpc.com/connect`) yang dimount ke `net/http.ServeMux` standar Go 1.22+ dengan **JSON codec** (lihat ADR `docs/adr/0005-migrasi-dari-gin-ke-net-http-servemux.md` dan `DEVELOPMENT-GUIDELINES.md`). Seluruh pemetaan endpoint REST di bawah sudah diganti dengan prosedur ConnectRPC + path file aktual.
>
> Dokumen ini telah **dipecah menjadi file per-modul** di direktori [`docs/mikhmon/`](docs/mikhmon/README.md). Setiap modul berisi pemetaan legacy → prosedur ConnectRPC, request/response (proto), dan logika khusus yang wajib dipertahankan.

---

## Ringkasan

Mikhmon v4 adalah aplikasi **MikroTik Hotspot Monitor** berbasis PHP monolitik yang:

- Menyimpan konfigurasi router di **file teks** `config/config.php` dengan 13 separator unik.
- Berkomunikasi dengan router via **RouterOS API socket TCP 8728/8729** (bukan HTTP).
- Tidak memiliki REST API — "API"-nya adalah file PHP (`get/*.php` & `post/*.php`) yang dipanggil via AJAX, sebagian besar response di-obfuscate **XOR 25 (`jsEncode`)**.
- Menyimpan laporan penjualan **di dalam MikroTik** (`/system/script`), dan metadata bisnis (harga, masa aktif, mode expire) di **on-login script** User Profile.

**Target transformasi (Polyglot):** prosedur ConnectRPC stateless dengan JWT, inventaris device di PostgreSQL + credential di vault, connection RouterOS per-request via `port.DeviceDriver`, dan kompatibilitas penuh dengan skrip MikroTik yang sudah terpasang.

**Kondisi saat ini:** sebagian besar logika sudah diimplementasikan sebagai driver + gateway + usecase + handler ConnectRPC (lihat §Status Implementasi). Sebagian prosedur masih perlu diekspos/ditambahkan (lihat §Gap & Pekerjaan Berikutnya).

---

## Status Implementasi per Modul

| # | Modul | Status | Implementasi aktual |
| :-- | :-- | :-- | :-- |
| 1 | Auth & Admin | ✅ **Terverifikasi** | `AuthService` (Login/GetMe/RefreshToken/Logout) + `UserService` + RBAC Casbin |
| 2 | Device / Router Instance | ✅ **Terverifikasi** (bentuk berbeda) | `DeviceService` (CRUD + TestDeviceConnection + streaming) |
| 3 | Dashboard & Monitoring | ✅ **Terverifikasi** | `GetDashboard` **dihapus total** → 10 prosedur streaming terpisah: `StreamSystemSnapshot` (5 print `interval=1s` digabung), `StreamInterfaceEthernet`, `StreamQueueStats`, `StreamLogs`, `StreamTraffic`, `StreamResource`, `StreamActiveSessions`, `StreamHotspotInactive`, `StreamPPPActive`, `StreamPPPInactive` |
| 4 | Hotspot User | ✅ **Terverifikasi** | `ListUsers` (+filter comment/only_unused), `GetUser`, `CreateUser`, `UpdateUser` (reset + comment legacy), `ResetUserCounters`, `DeleteUser` |
| 5 | Hotspot Profile | ✅ **Terverifikasi** | `ListProfiles` (parser on-login terstruktur), `CreateProfile`, `UpdateProfile`, `DeleteProfile` |
| 6 | Active / Host / Server | ✅ **Terverifikasi** | `ListActiveSessions`/`KickActiveSession` (+ DHCP lease), `ListHosts`, `RemoveHost`, `ListHotspotServers` |
| 7 | Voucher Generator | ✅ **Terverifikasi** | `GenerateVouchers` (server/time_limit/data_limit/comment) + `GetVoucherBatch` (cache voucher `uptime=0s`) |
| 8 | Sales Report | ✅ **Terverifikasi** | `ListReports` (filter day/month/year + `summary_only`, **bug filter gateway diperbaiki**), `DeleteReport` |
| 9 | Expire Monitor | ✅ **Terverifikasi** | `GetExpireMonitorStatus` (kompatibel 2 bentuk scheduler), `SetupExpireMonitor` (idempoten, `00:01:00`), `DisableExpireMonitor`, `RemoveExpireMonitor` |
| 10 | Template & Print | ✅ **Terverifikasi** | `ListTemplates`, `GetTemplateSection` (read-only), `RenderVouchers` (single/batch/preview, QR login URL); `SaveTemplateSection` = lanjutan |
| 11 | Resources & Theme | 🟡 **Sebagian** | Gateway `ListIPPools`/`ListParentQueues`/`ListNATRules` ada; belum diekspos; theme belum |

Legend: ✅ sudah lengkap · 🟡 sebagian (gateway/usecase ada, prosedur Connect belum) · 🔴 rencana.

---

## Daftar Modul (referensi ke file detail)

| # | Modul | File | Isi |
| :-- | :-- | :-- | :-- |
| – | Master (arsitektur legacy, inventaris API, standar response, struktur proyek, roadmap) | [`docs/mikhmon/README.md`](docs/mikhmon/README.md) | Seluruh info lintas-modul |
| 1 | Auth & Admin | [`docs/mikhmon/01-auth-admin.md`](docs/mikhmon/01-auth-admin.md) | Login/logout JWT, `me`, ganti credential |
| 2 | Device / Router Instance | [`docs/mikhmon/02-router-instance.md`](docs/mikhmon/02-router-instance.md) | CRUD device, ping, upload logo, config legacy + enkripsi |
| 3 | Dashboard & Monitoring | [`docs/mikhmon/03-dashboard-monitoring.md`](docs/mikhmon/03-dashboard-monitoring.md) | Streaming terpisah per-area: system snapshot, interface, queue, log, traffic, active/inactive hotspot & PPP |
| 4 | Hotspot User | [`docs/mikhmon/04-hotspot-user.md`](docs/mikhmon/04-hotspot-user.md) | CRUD user, reset counter, logika comment `vc-`/`up-` |
| 5 | Hotspot Profile | [`docs/mikhmon/05-hotspot-profile.md`](docs/mikhmon/05-hotspot-profile.md) | CRUD profile + generator/parser on-login script |
| 6 | Active / Host / Server | [`docs/mikhmon/06-active-host-server.md`](docs/mikhmon/06-active-host-server.md) | List active, kick, host, server hotspot |
| 7 | Voucher Generator | [`docs/mikhmon/07-voucher-generator.md`](docs/mikhmon/07-voucher-generator.md) | Generate batch (vc/up, char-type), query batch, formatter |
| 8 | Sales Report | [`docs/mikhmon/08-sales-report.md`](docs/mikhmon/08-sales-report.md) | Report per hari/bulan, parser record `-|-` |
| 9 | Expire Monitor | [`docs/mikhmon/09-expire-monitor.md`](docs/mikhmon/09-expire-monitor.md) | Status + install/enable/disable/uninstall scheduler |
| 10 | Template & Print | [`docs/mikhmon/10-template-print.md`](docs/mikhmon/10-template-print.md) | Template editor, render voucher (placeholder + QR) |
| 11 | Resources & Theme | [`docs/mikhmon/11-resources-theme.md`](docs/mikhmon/11-resources-theme.md) | Interface, IP pool, parent queue, NAT, theme |

---

## Inventaris API Legacy (ringkas)

**20 endpoint GET + 12 handler POST** — tabel lengkap (parameter, command RouterOS, format response asli) ada di [`docs/mikhmon/README.md` §3](docs/mikhmon/README.md#3-inventaris-api-surface-asli-requestresponse-aktual).

### Endpoint GET (file `get/*.php`)

`users`, `user`, `profiles`, `profile`, `get_sys_resource`, `get_hotspotinfo`, `get_traffic`, `get_log`, `get_livereport`, `get_report`, `get_hotspot_server`, `get_hotspot_active`, `get_hosts`, `get_expire_mon`, `get_tot_users`, `get_interface`, `get_addr_pool`, `get_parent_queue`, `get_nat`, `connect` (+ `set_theme` admin).

### Handler POST (file `post/*.php`)

`post_a_router` (add/remove/save/saveAdmin), `post_add_user`, `post_update_user`, `post_add_userprofile`, `post_update_userprofile`, `post_generate_voucher`, `post_cache_voucher`, `post_hotspot_remove`, `post_expire_monitor`, `post_template`, `post_logout`.

---

## Pemetaan Legacy → Prosedur ConnectRPC Polyglot

Semua prosedur dipanggil sebagai `POST /<Service>/<Procedure>` pada `net/http.ServeMux` (JSON codec). Service `HotspotService`, `DeviceService`, `AuthService` di-mount oleh `internal/app/app.go`; service selain `AuthService` dibungkus middleware `AuthenticateJWT` + `AuthorizeProcedure` (Casbin).

### Service yang sudah ada (path mount)

| Service | Prefix URL | Mounted di |
| :-- | :-- | :-- |
| `AuthService` | `/polyglot.v1.AuthService/` | public (auth via procedure) |
| `UserService` | `/polyglot.v1.UserService/` | protected (JWT + RBAC) |
| `DeviceService` | `/polyglot.v1.DeviceService/` | protected (JWT + RBAC) |
| `HotspotService` | `/polyglot.v1.HotspotService/` | protected (JWT + RBAC) |
| `RBACService`, `BillingService`, `WhatsAppService`, `BotService`, `KnowledgeService`, `ProbeService` | `/polyglot.v1.*/` | protected (JWT + RBAC) |
| Realtime | `GET /events` (SSE hub), `GET /ws/devices/{id}/terminal` (WebSocket) | `internal/adapter/ws/router.go` |

### Pemetaan per fungsi legacy

| Fungsi legacy | Prosedur ConnectRPC | Status |
| :-- | :-- | :-- |
| `users` / `user` / `get_tot_users` | `HotspotService/ListUsers` (`ListHotspotUsersRequest{device_id, profile, comment, only_unused}`), `GetUser`, `CreateUser`, `UpdateUser`, `ResetUserCounters`, `DeleteUser` | ✅ diekspos (`user_handler.go` + `mapper_user.go`) |
| `profiles` / `profile` | `HotspotService/ListProfiles`, `CreateProfile`, `UpdateProfile`, `DeleteProfile` | ✅ diekspos (`profile_handler.go` + `mapper_profile.go`; parser `ParseOnLoginScript`) |
| `get_sys_resource` + `get_hotspotinfo` | `HotspotService/StreamSystemSnapshot` (`StreamSystemSnapshotRequest{device_id, interval}`) — 5 stream `interval=1s` (clock, resource, routerboard, identity, health) digabung backend jadi 1 frame | ✅ diekspos (`monitor_system_handler.go`); `GetDashboard` lama **dihapus total** |
| `get_traffic` | `HotspotService/StreamTraffic` (`StreamTrafficRequest{device_id, interface}`) | ✅ diekspos (server-streaming, `monitor_interface_handler.go`) |
| `get_sys_resource` (realtime) | `HotspotService/StreamResource` (`StreamResourceRequest{device_id, interval}`) | ✅ diekspos (server-streaming native `interval=1s`, `monitor_system_handler.go`) |
| `get_hotspot_active` (realtime) | `HotspotService/StreamActiveSessions` (`StreamActiveSessionsRequest{device_id, user_filter}`) | ✅ diekspos (server-streaming `follow`, `monitor_session_handler.go`) |
| – (inactive hotspot) | `HotspotService/StreamHotspotInactive` (`StreamHotspotInactiveRequest{device_id, profile_filter, interval}`) — user + active `interval` → selisih `FilterInactiveHotspotUsers` | ✅ diekspos (`monitor_session_handler.go`) |
| – (PPP active/inactive, baru) | `HotspotService/StreamPPPActive` (`follow`) + `StreamPPPInactive` (secret + active `interval` → `FilterInactivePPPoESecrets`) | ✅ diekspos (`monitor_ppp_handler.go`) |
| – (interface ethernet, queue stats) | `HotspotService/StreamInterfaceEthernet` (`/interface/ethernet/print interval=1s`) + `StreamQueueStats` (`/queue/simple/print stats interval=1s`) | ✅ diekspos (`monitor_interface_handler.go`, `monitor_queue_handler.go`) |
| `get_log` | `HotspotService/StreamLogs` (`StreamLogsRequest{device_id, topics_filter}`) — `/log/print follow` | ✅ diekspos (`monitor_log_handler.go`) |
| `get_hotspot_active` / remove active | `HotspotService/ListActiveSessions` / `KickActiveSession` | ✅ diekspos |
| DHCP lease (baru, tidak di legacy) | `HotspotService/ListDHCPLeases` / `BlockDHCPLease` | ✅ diekspos |
| `get_hosts` / remove host | `HotspotService/ListHosts` / `RemoveHost` | ✅ diekspos (`host_server_handler.go` + `mapper_host_server.go`) |
| `get_hotspot_server` | `HotspotService/ListHotspotServers` | ✅ diekspos |
| `post_generate_voucher` | `HotspotService/GenerateVouchers` (`GenerateVouchersRequest{device_id, profile, count, server, time_limit, data_limit, comment, ...}`) | ✅ diekspos; algoritma di `internal/driver/mikrotik/hotspot/voucher.go` |
| `post_cache_voucher` | `HotspotService/GetVoucherBatch` (`GetVoucherBatchRequest{device_id, comment}`) | ✅ diekspos (filter comment + `uptime=0s`) |
| `get_report` / `get_livereport` | `HotspotService/ListReports` (`ListHotspotReportsRequest{device_id, day, month, year, summary_only}`) | ✅ diekspos (filter `day`→`?source=`, `month`→`?owner=`, `year`→suffix; fix bug `?owner=mikhmon-report`) |
| `post_expire_monitor` / `get_expire_mon` | `HotspotService/GetExpireMonitorStatus`/`SetupExpireMonitor`/`DisableExpireMonitor`/`RemoveExpireMonitor` | ✅ diekspos (status kompatibel `Mikhmon-Expire-Monitor` & `mikhmon-expire-scheduler`) |
| `post_template` / print voucher | `HotspotService/ListTemplates`/`GetTemplateSection`/`RenderVouchers` | ✅ diekspos (read-only; QR login URL; metadata scope-down) |
| `get_interface` / `get_addr_pool` / `get_parent_queue` / `get_nat` | `HotspotGateway.ListIPPools/ListParentQueues/ListNATRules` | 🟡 di gateway/usecase, belum diekspos |
| `post_a_router` (add/remove/save) | `DeviceService/UpdateDevice` (create via `UpdateDeviceRequest{device}`) / `DeleteDevice` | ✅ diekspos (inventaris di DB, bukan file `config.php`) |
| `connect` (test) | `DeviceService/TestDeviceConnection` + `StreamDeviceStatus` | ✅ diekspos |
| login/logout | `AuthService/Login`/`Logout` + `RefreshToken` | ✅ diekspos |
| `set_theme` | – | 🔴 belum (perlu settings service) |

---

## Standar Response & Error (ConnectRPC)

Proyek ini **tidak** memakai envelope JSON `{success, message, data}` ala REST. Sebagai gantinya:

- **Sukses:** protobuf message sesuai prosedur (`*Response`), di-serialize sebagai JSON via `internal/adapter/connect/codec.go` (`JSONCodec()`).
- **Error:** `connect.Error` dengan kode standar Connect (setara HTTP): `CodeUnauthenticated` (401), `CodePermissionDenied` (403), `CodeNotFound` (404), `CodeInvalidArgument` (400), `CodeFailedPrecondition`, `CodeInternal` (500), dst.
- **Pemetaan error domain → Connect code:** satu fungsi wajib `response.MapDomainError(err)` di `pkg/response/errors.go` — semua handler ConnectRPC **harus** memakainya (DEVELOPMENT-GUIDELINES.md §6).

Pemetaan kode error yang dulu direncanakan (`ROUTEROS_CONNECTION_FAILED`, `ROUTEROS_AUTH_FAILED`, `ROUTEROS_TRAP`, `UNAUTHORIZED`, dll.) kini diwakili oleh `connect.Code` berikut:

| Kode lama (draft REST) | Kode ConnectRPC | Kapan |
| :--- | :--- | :--- |
| `UNAUTHORIZED` | `CodeUnauthenticated` | Token JWT invalid/expired |
| `FORBIDDEN` | `CodePermissionDenied` | Tidak punya akses (Casbin / policy gate) |
| `ROUTER_NOT_FOUND` | `CodeNotFound` | Device ID tidak terdaftar (`device.ErrNotFound`) |
| `ROUTEROS_CONNECTION_FAILED` | `CodeInternal` (atau `CodeUnavailable`) | Gagal konek / timeout ke MikroTik |
| `ROUTEROS_TRAP` | `CodeInternal` | Command RouterOS mengembalikan `!trap` |
| `VALIDATION_ERROR` | `CodeInvalidArgument` | Body/query tidak valid |
| `NOT_FOUND` | `CodeNotFound` | Resource tidak ditemukan |
| (baru) | `CodeFailedPrecondition` | Approval destructive command / driver tidak support streaming (`ErrApprovalRequired`, `ErrDriverNotStreaming`) |

---

## Arsitektur Implementasi di Polyglot

Clean Architecture dengan boundary layer (domain → port → usecase → adapter/driver):

```text
polyglot/
├── api/proto/v1/
│   ├── hotspot.proto            # HotspotService (dashboard, users, profiles, active, voucher, streaming)
│   ├── device.proto             # DeviceService (inventaris + test + streaming)
│   └── auth.proto / users.proto # AuthService, UserService, RBAC
├── internal/
│   ├── domain/                  # entity murni (device, command, dst) — bebas proto
│   ├── port/
│   │   ├── hotspot_gateway.go   # HotspotGateway (abstraksi operasi hotspot/voucher)
│   │   ├── device_driver.go     # DeviceDriver / StreamingDeviceDriver
│   │   └── ...                  # repository, auth_service, credential_vault, dst
│   ├── usecase/
│   │   ├── hotspot/hotspot_usecase.go   # orkestrasi (gateway + port)
│   │   ├── device/manage_device.go      # CRUD inventaris + vault
│   │   ├── auth/                        # login, refresh token
│   │   └── network/                     # ExecuteCommand (policy gate), ActiveSessions, Terminal
│   ├── driver/
│   │   └── mikrotik/
│   │       ├── hotspot/                 # ★ Logika Mikhmon (gateway, comment, profile, voucher, expire, report)
│   │       └── ...                      # driver RouterOS lain
│   ├── adapter/
│   │   ├── connect/hotspot/             # ★ Handler ConnectRPC HotspotService + mapper + router
│   │   ├── connect/device/              # Handler ConnectRPC DeviceService
│   │   ├── connect/auth/                # Handler AuthService/UserService/RBAC
│   │   ├── ws/                          # SSE hub (/events) + terminal WebSocket
│   │   ├── http/middleware/             # Chain, JWT, RBAC, CORS, Logger, Recovery
│   │   └── postgres/ redis/ auth/ vault/# adapter infra
│   ├── registry/                        # DriverFactory (mikrotik, genieacs, ...)
│   └── app/app.go                       # Composition root — mount semua service
├── internal/template/                   # header/row/footer × default/small/thermal (embed)
└── pkg/response/errors.go               # MapDomainError (Connect)
```

**Poin kunci (berbeda dari draft awal):**

1. **Transport ConnectRPC**, bukan REST/gin/fiber. JSON codec kustom di `internal/adapter/connect/codec.go`.
2. **Tidak ada client pool RouterOS persisten per endpoint.** Koneksi dibuat per-request melalui `port.DeviceDriver` dari `registry.Registry.Get(ctx, deviceID)` (credential di-dekripsi dari vault). Eksekusi command wajib lewat `network.ExecuteCommand` (policy gate: klasifikasi → approval) — lihat ADR `0002-devicedriver-tanpa-session-terpisah.md` dan `0004-generic-cli-driver-scrapligo.md`.
3. **Inventaris device di PostgreSQL + vault AES**, bukan file `config/config.php` (lihat `docs/database-schema.md`).
4. **Streaming realtime sudah ada** via ConnectRPC server-streaming (`StreamTraffic`, `StreamResource`, `StreamActiveSessions`) + SSE hub (`/events`) untuk event WhatsApp/chat + WebSocket terminal (`/ws/devices/{id}/terminal`).
5. **Konteks arsitektur lengkap:** `Polyglot-Architecture.md`, `SYSTEM-STRUCTURE-AND-ARCHITECTURE.md`, `DEVELOPMENT-GUIDELINES.md`.

---

## Gap & Pekerjaan Berikutnya (dari sisi Mikhmon)

Prosedur ConnectRPC yang belum diekspos namun logika gateway/usecase-nya sudah ada:

> ✅ **Selesai batch pertama (modul 04–07):** user CRUD, profile CRUD, hosts/servers, dan voucher batch sudah diekspos (detail di tabel pemetaan §"Pemetaan per fungsi legacy").
>
> ✅ **Selesai batch kedua (modul 08–10):** Sales Report (`ListReports`/`DeleteReport` + fix filter), Expire Monitor (4 prosedur, kompatibel 2 bentuk scheduler), Template & Print (`ListTemplates`/`GetTemplateSection`/`RenderVouchers`, QR login URL, metadata scope-down).

5. **Log:** prosedur baru untuk `get_log` (`/log/print` + pastikan logging prefix `->`).
6. **Resources:** `ListIPPools`, `ListParentQueues`, `ListNATRules`, `ListInterfaces` + theme/settings.
7. **Migrasi legacy:** parser `config/config.php` + `decode`/`enc_rypt`/`dec_rypt` untuk migrasi data lama.
8. **Fase lanjutan template:** `SaveTemplateSection` (override DB), metadata Device (`currency`/`phone`), QR warna, `web/src/gen` (TS) regenerate.

---

*Dokumen ini bersifat living document — setiap endpoint yang diimplementasikan sebaiknya diverifikasi ulang terhadap `get/*.php` dan `post/*.php` untuk menjaga kompatibilitas perilaku, dan terhadap file `.go` aktual di `internal/driver/mikrotik/hotspot/` serta `internal/adapter/connect/hotspot/` untuk menjaga sinkronisasi status.*
