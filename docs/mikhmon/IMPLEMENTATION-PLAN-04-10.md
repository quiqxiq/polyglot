# Rencana Implementasi — Modul 04–10 (Hotspot ConnectRPC Gap Closure)

> Kembali ke [README](README.md)
>
> Dokumen ini adalah **rencana kerja lengkap** untuk menutup seluruh gap implementasi Mikhmon pada proyek Polyglot, khususnya modul **04 s.d. 10** (Hotspot User, Profile, Active/Host/Server, Voucher, Sales Report, Expire Monitor, Template & Print), **plus perluasan proto `Device`** dengan metadata bisnis Mikhmon (yang dibutuhkan modul 10 dan migrasi legacy).
>
> Setiap fase ditulis berurutan sesuai dependensi, menyertakan: perubahan proto, file yang dibuat/diubah, logika wajib yang harus 1:1 dengan legacy, dan acceptance criteria. Semua verifikasi path & nama fungsi mengacu pada kode aktual (terverifikasi terhadap `internal/driver/mikrotik/hotspot/`, `internal/usecase/hotspot/`, `internal/adapter/connect/hotspot/`, `internal/port/`, `api/proto/v1/`).

---

> ✅ **Status batch pertama (Fase 1–4, modul 04–07): SELESAI & terimplementasi** — `user_handler.go`, `profile_handler.go`, `host_server_handler.go`, `voucher_handler.go` + mapper pendamping. Fase 0 (metadata `Device`) **tidak dikerjakan** atas permintaan user.
>
> ✅ **Status batch kedua (Fase 5–7, modul 08–10): SELESAI & terimplementasi** — `report_handler.go` (+ `mapper_report.go`, fix bug filter `ListReports`), `expire_monitor_handler.go` (+ `port/hotspot_expire.go`, status kompatibel dua bentuk scheduler, interval default `00:01:00`), `template_handler.go` (+ `pkg/voucher`: `RenderWithOptions`/`QRContent`/`ListTemplates`, QR mode login URL, metadata scope-down). Keputusan §12: 1=A, 2=00:01:00, 3=return all+summary_only, 4=scope-down, 5=read-only, 6=URL default, 7=skip QR warna. Fase 8 (test & verifikasi) termasuk dalam batch ini — `go build`/`go vet`/`go test ./...` hijau.
>
> ✅ **Status batch ketiga (modul 03, dashboard streaming terpisah): SELESAI & terimplementasi** — `GetDashboard` (endpoint agregasi) **dihapus total** dan diganti **10 prosedur streaming terpisah per-area**: `StreamSystemSnapshot` (5 stream `interval=1s` digabung: clock, resource, routerboard, identity, health), `StreamInterfaceEthernet`, `StreamQueueStats`, `StreamLogs`, `StreamTraffic`, `StreamResource`, `StreamActiveSessions`, `StreamHotspotInactive`, `StreamPPPActive`, `StreamPPPInactive`. Semua streaming native RouterOS (`follow`/`interval`/`monitor-traffic`) — tanpa polling backend. Handler: `monitor_*_handler.go` (6 file); driver baru: `system_health.go`, `system_routerboard.go` + varian `interval` di `iface.go`, `hotspot_user.go`, `hotspot_active.go`, `ppp.go`, `ppp_active.go`. `system_report_handler.go` dihapus. `go build`/`go vet`/`go test ./...` hijau.

## 0. Ringkasan Status Gap

| Modul | Prosedur yang SUDAH diekspos | Gap (yang harus dibuat) |
| :--- | :--- | :--- |
| 04 User | ✅ `ListUsers`, `GetUser`, `CreateUser`, `UpdateUser` (+reset), `ResetUserCounters`, `DeleteUser` | — |
| 05 Profile | ✅ `ListProfiles`, `CreateProfile`, `UpdateProfile`, `DeleteProfile`, parser on-login terstruktur | `GetProfile` (opsional) |
| 06 Active/Host/Server | ✅ `ListActiveSessions`, `KickActiveSession`, `ListDHCPLeases`, `BlockDHCPLease`, `ListHosts`, `RemoveHost`, `ListHotspotServers` | — |
| 07 Voucher | ✅ `GenerateVouchers` (server/time_limit/data_limit/comment), `GetVoucherBatch` (query batch `uptime=0s`) | — |
| 08 Sales Report | ✅ `ListReports` (filter day/month/year + summary_only), `DeleteReport`; **bug filter gateway diperbaiki** | — |
| 09 Expire Monitor | ✅ `GetExpireMonitorStatus`, `SetupExpireMonitor` (idempoten), `DisableExpireMonitor`, `RemoveExpireMonitor` | — |
| 10 Template & Print | ✅ `ListTemplates`, `GetTemplateSection` (read-only), `RenderVouchers` (single/batch/preview, QR login URL) | `SaveTemplateSection` (fase lanjutan) |

**Prasyarat lintas modul:** metadata Mikhmon di proto `Device` (hotspot name, dns name, currency, phone, dst.) — modul 10 butuh field ini untuk render header voucher (Fase 0, **tidak dikerjakan**).

---

## 1. Temuan Kunci yang Mengubah Rencana (diverifikasi terhadap kode)

Sebelum menyusun fase, empat temuan berikut **memperbaiki asumsi di dokumen modul** dan wajib jadi dasar rencana:

### 1.1 Engine render voucher SUDAH ADA di `pkg/voucher` 🔴→🟡
`pkg/voucher/generator.go` sudah mengimplementasikan **seluruh engine render** (dokumen modul 10 menulis "belum ada logika render" — ini tidak akurat):
- `Render(vouchers []VoucherData, layout Layout, templateDir string) (string, error)` — baca `header/row/footer.<layout>.txt`, render per voucher, gabung HTML lengkap.
- Placeholder `%username% %password% %price% %validity% %limitUptime% %limitBytesTotal% %hotspotName% %dnsName% %logo% %comment% %timeStamp% %qrCode% %#%` — semuanya via `strings.ReplaceAll` (aman, tanpa template engine).
- QR: `github.com/skip2/go-qrcode` → data URI base64 (`generateQRBase64`), konten `username\npassword`.
- Layout: `default` / `small` / `thermal`.

> **Dampak:** modul 10 tinggal *mengekspos* engine ini lewat ConnectRPC + menyediakan metadata router. Tidak perlu menulis engine render baru. Gap QR warna (`%qrCodeRed/Green/Blue%`) masih ada — ekstensi kecil (lihat Fase 7).

### 1.2 BUG: filter `ListReports` gateway tidak akan pernah match record nyata 🐛
`internal/driver/mikrotik/hotspot/gateway.go` `ListReports` memakai:
```go
Args: map[string]string{"?owner": "mikhmon-report"}
```
Tetapi on-login script (`profile.go` baris ~74) membuat script laporan dengan:
- `owner = $month$year` → contoh `aug2026`
- `source = $date` → contoh `aug/17/2026`
- `comment = mikhmon`

Legacy `get_report.php` memfilter `?source=<day>`, `get_livereport` memfilter `?owner=<month>`. **Akibatnya `ListReports` (dan `GetTodayIncome` yang bergantung padanya) mengembalikan data kosong di router yang sudah terpasang Mikhmon asli.** Rencana modul 08 wajib memperbaiki ini (lihat Fase 5).

### 1.3 Nama scheduler expire monitor: gateway ≠ legacy
- Legacy: **satu** scheduler `Mikhmon-Expire-Monitor` dengan `on-event` = source script langsung.
- Gateway saat ini: **dua langkah** — script `mikhmon-expire-monitor` + scheduler `mikhmon-expire-scheduler` (on-event `/system script run mikhmon-expire-monitor`), interval default `00:05:00` (legacy `00:01:00`).

Perlu **keputusan kompatibilitas** (Fase 6) — status check harus mengenali kedua bentuk agar tidak "not ready" di router legacy yang sudah terpasang.

### 1.4 `make proto` menghasilkan `*_grpc.pb.go`, handler ConnectRPC ditulis manual
`Makefile` target `proto` hanya menjalankan `protoc-gen-go` + `protoc-gen-go-grpc` (tidak ada `protoc-gen-connect-go`). Service ConnectRPC di-render **manual** via `connect.NewUnaryHandler` / `connect.NewServerStreamHandler` di `router.go` masing-masing adapter. Konsekuensi untuk setiap fase:
1. Edit `api/proto/v1/<x>.proto` → jalankan `make proto` → regenerate `api/gen/v1/<x>.pb.go` + `<x>_grpc.pb.go`.
2. Tulis handler manual di `internal/adapter/connect/<domain>/<modul>_handler.go`.
3. Daftarkan di `router.go` (pattern `mux.Handle("/"+serviceName+"/<Procedure>", ...)`).

---

## 2. Prasyarat & Konvensi Bersama (dipakai semua fase)

- **Driver per request:** semua handler memakai pola `h.getDriver(ctx, device_id)` (sudah ada di `HotspotConnectHandler`).
- **Error mapping:** setiap error dari usecase/gateway wajib lewat `response.MapDomainError(err)`; `device_id` kosong → `CodeInvalidArgument` (lihat `getDriver`).
- **Policy gate:** semua command destruktif (remove, reset, disable) dieksekusi lewat `network.ExecuteCommand` → otomatis kena klasifikasi/approval. **Jangan** panggil `driver.Execute` langsung di handler.
- **Mapper:** tipe port di-*alias* oleh driver (`mikrotik.HotspotUser = port.HotspotUser`, dst.) — mapper bisa menerima tipe port langsung. `mapper.go` saat ini ~90 baris; setelah semua fase bisa melewati 400 baris → **pecah** menjadi file per modul: `mapper_user.go`, `mapper_profile.go`, `mapper_report.go`, `mapper_meta.go` (aturan AGENTS.md: max 400–500 baris/file).
- **RBAC:** setiap prosedur baru harus didaftarkan di policy Casbin (seed/policy loader) — cek pola prosedur lain saat implementasi.
- **File baru:** ikuti tabel penempatan AGENTS.md (handler → `internal/adapter/connect/hotspot/`, mapper → folder sama).
- **Verifikasi setelah tiap fase:** `go build -v ./cmd/server` + `go test ./...` + `make proto` bila ada perubahan proto.

---

## 3. FASE 0 — Metadata Mikhmon di Proto `Device` (prasyarat Modul 10)

> Modul 10 butuh `hotspotName`, `dnsName`, `currency`, `phone`, `email`, `infoLp`, `idleTimeout` untuk render header voucher dan login URL QR. Saat ini **tidak ada** di message `Device` (hanya `id` s.d. `ssh_port`), domain `device.Device`, maupun tabel `devices`.

### 3.1 Proto — `api/proto/v1/device.proto`

```protobuf
// Metadata bisnis Mikhmon yang melekat pada device (setara separator config legacy modul 02).
message DeviceMikhmonMetadata {
  string hotspot_name = 1;   // % (branding voucher)   — mis. "WIFI BERKAH"
  string dns_name = 2;       // ^ (hostname hotspot)   — mis. "wifi.net"
  string currency = 3;       // & (mata uang)          — mis. "Rp"
  string phone = 4;          // * (kontak voucher)
  string email = 5;          // ( (email admin)
  string info_lp = 6;        // ) (info login page, hex-encoded)
  string idle_timeout = 7;   // = (menit / "disable")
  string live_report = 8;    // @!@ (enable / disable)
}

message Device {
  // ... field 1-12 eksisting ...
  DeviceMikhmonMetadata mikhmon = 13; // NEW — metadata bisnis Mikhmon
}
```

### 3.2 Domain — `internal/domain/device/device.go`
Tambahkan struct typed (bukan menumpuk di `Extra`) agar mapper/DB eksplisit:
```go
// MikhmonMetadata holds Mikhmon v4 business metadata for a device (voucher
// branding, currency, contact, login-page info). Mirrors the legacy config.php
// separator fields (modul 02 §2.2).
type MikhmonMetadata struct {
    HotspotName  string `json:"hotspot_name,omitempty"`
    DNSName      string `json:"dns_name,omitempty"`
    Currency     string `json:"currency,omitempty"`
    Phone        string `json:"phone,omitempty"`
    Email        string `json:"email,omitempty"`
    InfoLP       string `json:"info_lp,omitempty"`
    IdleTimeout  string `json:"idle_timeout,omitempty"`
    LiveReport   string `json:"live_report,omitempty"` // "enable" | "disable"
}
// di struct Device:  Mikhmon *MikhmonMetadata `json:"mikhmon,omitempty"`
```

### 3.3 Migration — `migrations/`
Buat `migrations/NNNNNN_add_device_mikhmon_metadata.up.sql` (+ `.down.sql`):
```sql
ALTER TABLE devices
  ADD COLUMN hotspot_name  TEXT DEFAULT '',
  ADD COLUMN dns_name      TEXT DEFAULT '',
  ADD COLUMN currency      TEXT DEFAULT '',
  ADD COLUMN phone         TEXT DEFAULT '',
  ADD COLUMN email         TEXT DEFAULT '',
  ADD COLUMN info_lp       TEXT DEFAULT '',
  ADD COLUMN idle_timeout  TEXT DEFAULT '',
  ADD COLUMN live_report   TEXT DEFAULT '';
```
> Alternatif tanpa migration: simpan di `Device.Extra` (`map[string]string`). **Rekomendasi: kolom typed** (queryable, dokumentatif). Pilih salah satu dan konsisten di mapper Postgres + `ToProtoDevice`/`ToDomainDevice`.

### 3.4 File yang disentuh
| File | Aksi |
| :--- | :--- |
| `api/proto/v1/device.proto` | tambah message + field `mikhmon = 13` |
| `api/gen/v1/device.pb.go`, `device_grpc.pb.go` | regenerate via `make proto` |
| `internal/domain/device/device.go` | struct `MikhmonMetadata` + field |
| `internal/adapter/postgres/device_repository.go` | scan/insert/update kolom baru |
| `internal/adapter/connect/device/device_handler.go` + `mapper.go` | map `Device.Mikhmon` ⇄ proto |
| `migrations/` | file up/down baru |

### 3.5 Acceptance criteria
- `make proto` sukses; `go build ./...` sukses.
- `UpdateDevice` dapat menyimpan & mengembalikan `mikhmon`; `ListDevices`/`GetDevice` mengembalikannya.
- Kolom baru ada setelah `make migrate-up` (cek `\d devices`).

---

## 4. FASE 1 — Modul 04: Hotspot User CRUD (gap: 5 prosedur)

**Status logika:** gateway (`GetUser`, `AddUser`, `UpdateUser`, `RemoveUser`, `ResetUserCounters`) dan usecase (`GetUser`, `AddUser`, `UpdateUser`, `RemoveUser`, `ResetUserCounters`) **sudah lengkap**. Yang kurang: proto, handler, mapper, dan **dua logika comment legacy** yang belum dipindah ke layer Go.

### 4.1 Proto — `api/proto/v1/hotspot.proto`

```protobuf
message GetHotspotUserRequest { string device_id = 1; string ros_id = 2; }
message GetHotspotUserResponse { HotspotUser user = 1; }

message CreateHotspotUserRequest {
  string device_id  = 1;
  string server     = 2;  // kosong = all
  string name       = 3;
  string password   = 4;
  string profile    = 5;
  string mac_address = 6;
  string time_limit = 7;  // "8h", "30d"
  string data_limit = 8;  // "1000M" → byte via ParseDataLimit
  string comment    = 9;
}
message CreateHotspotUserResponse { HotspotUser user = 1; string message = 2; }

message UpdateHotspotUserRequest {
  string device_id   = 1;
  string ros_id      = 2;
  bool   reset_counter = 3;  // legacy: /ip/hotspot/user/reset-counters dulu
  string server      = 4;
  string name        = 5;
  string password    = 6;
  string profile     = 7;
  string mac_address = 8;
  string time_limit  = 9;
  string data_limit  = 10;
  string comment     = 11;
  string expire_date = 12;  // logika comment legacy (lihat 4.3)
  string user_code   = 13;  // logika comment legacy
}
message UpdateHotspotUserResponse { HotspotUser user = 1; string message = 2; }

message ResetHotspotUserCountersRequest { string device_id = 1; string ros_id = 2; }
message ResetHotspotUserCountersResponse { string message = 1; }

message DeleteHotspotUserRequest { string device_id = 1; string ros_id = 2; }
message DeleteHotspotUserResponse { string message = 1; }
```
`service HotspotService` — tambahkan 5 rpc: `GetUser`, `CreateUser`, `UpdateUser`, `ResetUserCounters`, `DeleteUser`.

### 4.2 File
| File | Aksi |
| :--- | :--- |
| `api/proto/v1/hotspot.proto` + gen | tambah message & rpc, `make proto` |
| `internal/adapter/connect/hotspot/user_handler.go` | **baru** — 5 handler |
| `internal/adapter/connect/hotspot/mapper.go` (atau `mapper_user.go`) | `ToProtoHotspotUser`, `HotspotUserParamsFromProto`, `HotspotUserParamsUpdateFromProto` |
| `internal/adapter/connect/hotspot/router.go` | daftarkan 5 rpc |

### 4.3 Logika wajib (dipindah dari legacy `post_add_user` / `post_update_user`)

1. **Comment auto `vc-`/`up-` (CreateUser):** jika `comment` kosong → `name == password` ? prefix `vc-` : `up-`; hasil `FormatPreLoginComment` di `internal/driver/mikrotik/hotspot/comment.go` (reuse, jangan tulis ulang).
2. **Konversi data limit:** `data_limit` string → byte via `ParseDataLimit` (`voucher.go`) → isi `LimitBytesOut` (bukan In — cek command `NewAddHotspotUserCommand` untuk field mana yang dipakai).
3. **Comment rebuild legacy (UpdateUser)** — port aturan `post_update_user.php`:
   - `expire_date == "" && user_code == ""` → comment apa adanya.
   - `expire_date == "" && user_code != ""` → prefix lama (`vc-`/`up-`/`X-`) dipertahankan, comment dibangun ulang.
   - `expire_date != "" && user_code == ""` → `comment = "<expire_date> <comment>"`.
   Implementasikan sebagai helper murni (mis. `BuildUpdatedComment(existing, expireDate, userCode, newComment)`) agar unit-testable.
4. **Alur UpdateUser:** `reset_counter=true` → `ResetUserCounters` dulu, lalu `UpdateUser`, lalu print ulang (`GetUser`) untuk response.
5. **Error:** duplikat name → `!trap` → `response.MapDomainError` → `CodeInternal` dengan pesan asli.

### 4.4 Acceptance criteria
- Unit test helper comment (vc/up + rebuild rules) & `ParseDataLimit`.
- `CreateUser` menghasilkan user dengan comment prefix benar; `UpdateUser` dengan `reset_counter` memanggil reset dulu.
- Semua 5 rpc terdaftar di `router.go` dan lolos `go build` + `go test ./...`.

---

## 5. FASE 2 — Modul 05: Hotspot Profile CRUD (gap: 3–4 prosedur)

**Status logika:** `CreateUserProfile`, `UpdateUserProfile`, `DeleteUserProfile` gateway + `CreateProfile`, `UpdateProfile`, `DeleteProfile` usecase sudah ada. `BuildOnLoginScript` sudah ada (1:1 legacy).

### 5.1 Proto

```protobuf
// Mencerminkan port.MikhmonProfileParams.
message HotspotProfileParams {
  string name           = 1;
  string address_pool   = 2;
  string shared_users   = 3;
  string rate_limit     = 4;   // "5M/5M"
  string parent_queue   = 5;
  string price          = 6;
  string selling_price  = 7;
  string validity       = 8;   // "1d","7d","30d"
  string expire_mode    = 9;   // "0" | "ntf" | "ntfc" | "rem" | "remc"
  bool   lock_user      = 10;
  bool   lock_server    = 11;
  bool   enable_recording = 12;
  string comment        = 13;
}
message CreateHotspotProfileRequest { string device_id = 1; HotspotProfileParams profile = 2; }
message CreateHotspotProfileResponse { HotspotProfile profile = 1; string message = 2; }
message UpdateHotspotProfileRequest { string device_id = 1; string ros_id = 2; HotspotProfileParams profile = 3; }
message UpdateHotspotProfileResponse { HotspotProfile profile = 1; string message = 2; }
message DeleteHotspotProfileRequest { string device_id = 1; string ros_id = 2; }
message DeleteHotspotProfileResponse { string message = 1; }
```
Perluas message `HotspotProfile` (backward-compatible):
```protobuf
string address_pool = 12;  // NEW
string lock_server  = 13;  // NEW
```

### 5.2 File
| File | Aksi |
| :--- | :--- |
| `api/proto/v1/hotspot.proto` + gen | message & rpc baru |
| `internal/adapter/connect/hotspot/profile_handler.go` | **baru** — Create/Update/Delete (+ `GetProfile` opsional) |
| `internal/adapter/connect/hotspot/mapper_profile.go` | **baru** — `ProfileParamsFromProto`, `ProfileMetaFromProto`; perbaiki `ToProtoHotspotProfiles` |
| `internal/adapter/connect/hotspot/router.go` | daftarkan rpc |

### 5.3 Logika wajib
1. **Normalisasi nama:** spasi → `-` saat create (`preg_replace('/\s+/','-')` legacy) karena nama profile disisipkan ke nama script laporan yang dipisah `-|-` (modul 08).
2. **On-login script:** jangan kirim script mentah dari client — bangun dari `MikhmonProfileParams` via `BuildOnLoginScript` (anti-injeksi).
3. **Parser on-login terstruktur (gap mapper):** `ToProtoHotspotProfiles` saat ini mengisi `mode_expire` = raw on-login. Tambahkan di `internal/driver/mikrotik/hotspot/profile.go`:
   ```go
   type ProfileMeta struct {
       ExpireMode   string  // "0","rem","ntf","remc","ntfc"
       Price        float64
       Validity     string
       SellingPrice float64
       LockUser     string  // "Enable"/"Disable"
       LockServer   string  // "Enable"/"Disable"
   }
   func ParseOnLoginScript(onLogin string) (ProfileMeta, error) // parse :put (\",...\")
   ```
   Mapper mengisi `validity`, `price`, `selling_price`, `lock_user`, `lock_server`, `mode_expire` terstruktur dari `ParseOnLoginScript`.
4. **Delete:** RouterOS `!trap` bila profile masih dipakai user — propagasikan pesan asli.

### 5.4 Acceptance criteria
- Unit test `ParseOnLoginScript` terhadap contoh `:put (\",remc,3000,1h,3500,,Enable,Disable,\");`.
- `CreateProfile` menormalisasi nama; `ListProfiles` mengembalikan field metadata terstruktur.
- `go build` + `go test ./...` hijau.

---

## 6. FASE 3 — Modul 06: Active / Host / Server (gap: 3 prosedur)

**Status logika:** `GetHosts`/`RemoveHost`/`GetHotspotServers` usecase + gateway `ListHosts`/`RemoveHost`/`ListHotspotServers` sudah ada; hasilnya masih `[]map[string]string` (raw rows).

### 6.1 Proto

```protobuf
message HotspotHost {
  string id          = 1;
  string mac_address = 2;
  string address     = 3;
  string to_address  = 4;
  string server      = 5;
  bool   bypassed    = 6;
  bool   authorized  = 7;
  string comment     = 8;
}
message ListHotspotHostsRequest { string device_id = 1; }
message ListHotspotHostsResponse { repeated HotspotHost hosts = 1; }

message RemoveHotspotHostRequest { string device_id = 1; string ros_id = 2; }
message RemoveHotspotHostResponse { string message = 1; }

message HotspotServerInfo {
  string id           = 1;
  string name         = 2;
  string interface    = 3;
  string address_pool = 4;
  bool   disabled     = 5;
  string comment      = 6;
}
message ListHotspotServersRequest { string device_id = 1; }
message ListHotspotServersResponse { repeated HotspotServerInfo servers = 1; }
```
`service HotspotService` — tambahkan `ListHosts`, `RemoveHost`, `ListHotspotServers`.

### 6.2 File
| File | Aksi |
| :--- | :--- |
| `api/proto/v1/hotspot.proto` + gen | message & rpc |
| `internal/adapter/connect/hotspot/host_server_handler.go` | **baru** — 3 handler |
| `internal/adapter/connect/hotspot/mapper.go` | konversi `[]map[string]string` → proto (konversi `"true"/"false"` → bool) |
| `internal/adapter/connect/hotspot/router.go` | daftarkan rpc |

### 6.3 Logika wajib
1. **Parser raw rows → struct:** buat helper `parseHotspotHostRow(map[string]string) HotspotHost` dan `parseHotspotServerRow(...)` di mapper; `bypassed`/`authorized`/`disabled` dari string `"true"/"false"`.
2. **Kick vs hapus user:** dokumentasikan di response/error bahwa `RemoveHost` hanya menghapus entri host, bukan user.
3. **Paginasi (opsional):** daftar host bisa besar — tambah `limit`/`offset` bila frontend butuh.

### 6.4 Acceptance criteria
- `ListHosts` mengembalikan host dengan bool terkonversi; `RemoveHost` memanggil `/ip/hotspot/host/remove`.
- `go build` + `go test ./...` hijau.

---

## 7. FASE 4 — Modul 07: Voucher (gap: ekstensi request + query batch)

**Status logika:** generator lengkap; handler `GenerateVouchers` **tidak memetakan** `Server`, `LimitUptime`, `LimitBytes`, `CommentTag` dari proto meskipun `VoucherGenerateParams` mendukungnya. Query batch (`post_cache_voucher`) belum ada.

### 7.1 Proto

```protobuf
message GenerateVouchersRequest {  // diperluas (backward-compatible)
  string device_id     = 1;
  string profile       = 2;
  int32  count         = 3;
  string user_type     = 4;  // vc | up
  int32  user_length   = 5;
  string prefix        = 6;
  string character_set = 7;
  string server        = 8;  // NEW
  string time_limit    = 9;  // NEW — "1d","3h"
  string data_limit    = 10; // NEW — "1000M"
  string comment       = 11; // NEW — tag batch (gcomment)
}

message GetVoucherBatchRequest { string device_id = 1; string comment = 2; }
message GetVoucherBatchResponse { repeated HotspotUser vouchers = 1; int32 count = 2; }
```
Perluas `ListHotspotUsersRequest` (dipakai modul 04 & query batch):
```protobuf
string comment     = 3;  // NEW — filter comment
bool   only_unused = 4;  // NEW — uptime=0s (belum pernah login)
```
`service HotspotService` — tambahkan `GetVoucherBatch`.

### 7.2 File
| File | Aksi |
| :--- | :--- |
| `api/proto/v1/hotspot.proto` + gen | ekstensi + rpc |
| `internal/adapter/connect/hotspot/voucher_handler.go` | **baru** — `GetVoucherBatch`; perbaiki pemetaan `GenerateVouchers` (isi `Server/LimitUptime/LimitBytes/CommentTag`) |
| `internal/driver/mikrotik/hotspot/gateway.go` | `ListUsers` dukung filter `comment` + `only_unused` (print `?comment=<c> ?uptime=0s`) |
| `internal/usecase/hotspot/hotspot_usecase.go` | `GetUsers` terima opsi filter (atau tambah `GetUnusedUsersByTag`) |
| `internal/adapter/connect/hotspot/router.go` | daftarkan rpc |

### 7.3 Logika wajib
1. **`GetVoucherBatch`:** `/ip/hotspot/user/print ?comment=<comment> ?uptime=0s` — hanya voucher yang **belum pernah login**; response count + daftar (pengganti `post_cache_voucher`).
2. **Sanitasi tag/prefix:** tanpa spasi / karakter yang memecah format `vc-<code>-<date>-<tag>`.
3. **Batch besar:** tetap sekuensial seperti sekarang; catat di dokumentasi bahwa qty besar memblokir socket (pertimbangan async = out of scope fase ini).

### 7.4 Acceptance criteria
- `GenerateVouchers` dengan `time_limit`/`data_limit`/`comment` menghasilkan user dengan limit & tag benar.
- `GetVoucherBatch` hanya mengembalikan user `uptime=0s` dengan comment yang cocok.

---

## 8. FASE 5 — Modul 08: Sales Report (gap: 2 prosedur + 1 fix bug)

**Status logika:** parser `ParseMikhmonTransactions`, usecase `GetReports`, `GetReportsByFilter`, `DeleteReport`, `GetTodayIncome` sudah ada. **Bug §1.2 harus diperbaiki dulu** agar data benar-benar terbaca.

### 8.1 Perbaikan gateway `ListReports` (PRASYARAT)

Ubah `internal/driver/mikrotik/hotspot/gateway.go` — `ListReports` menerima filter dan memetakan ke filter RouterOS sesuai legacy:
```go
// Legacy: get_report → /system/script/print ?source=<day>
//         get_livereport → /system/script/print ?owner=<month>
// Tanpa filter → /system/script/print ?comment=mikhmon (hanya record Mikhmon)
func (g *Gateway) ListReports(ctx context.Context, driver port.DeviceDriver, day, month string) ([]port.MikhmonTransaction, error)
```
- `port.HotspotGateway` ikut diubah signature-nya; sesuaikan usecase `GetReports`/`GetReportsByFilter`/`GetTodayIncome`.
- `GetTodayIncome` tetap `time.Now()` lokal — dokumentasikan timezone router.

### 8.2 Proto

```protobuf
message ListHotspotReportsRequest {
  string device_id   = 1;
  string day         = 2;  // legacy: "aug/17/2026" (filter source)
  string month       = 3;  // legacy: "aug2026"     (filter owner)
  string year        = 4;  // filter suffix tanggal
  bool   summary_only = 5; // tanpa daftar record
}
message ListHotspotReportsResponse {
  repeated HotspotReport reports = 1;
  double total_income = 2;
  int32  total        = 3;
}
message DeleteHotspotReportRequest { string device_id = 1; string ros_id = 2; }
message DeleteHotspotReportResponse { string message = 1; }
```
`service HotspotService` — tambahkan `ListReports`, `DeleteReport`. Message `HotspotReport` sudah ada.

### 8.3 File
| File | Aksi |
| :--- | :--- |
| `internal/port/hotspot_gateway.go` | ubah signature `ListReports` |
| `internal/driver/mikrotik/hotspot/gateway.go` | implementasi filter baru |
| `internal/usecase/hotspot/hotspot_usecase.go` | sesuaikan pemanggilan |
| `internal/adapter/connect/hotspot/report_handler.go` | **baru** — `ListReports`, `DeleteReport` |
| `internal/adapter/connect/hotspot/mapper_report.go` | **baru** — `ToProtoHotspotReports` (konversi `Price` string → `double`) |
| `internal/adapter/connect/hotspot/router.go` | daftarkan rpc |

### 8.4 Logika wajib
1. **Filter ganda:** bila `day`+`month`+`year` kosong → semua record (atau `CodeInvalidArgument` — putuskan di handler; rekomendasi: kembalikan semua, `summary_only` untuk efisiensi).
2. **Record rusak:** `ParseMikhmonTransactions` melewati script tanpa `-|-`; `total_income` dihitung dari record valid.
3. **`summary_only`:** tanpa daftar record, hanya `total_income` + `total` (hemat bandwidth dashboard).
4. **DeleteReport** = `/system/script/remove` (baru — tidak ada di legacy).

### 8.5 Acceptance criteria
- Unit test `ParseMikhmonTransactions` dengan contoh record `aug/17/2026-|-14:20:00-|-VIP123-|-3000-|-...` (sudah ada sebagian; tambah kasus rusak).
- `ListReports` dengan `day="aug/17/2026"` hanya mengembalikan record hari itu; `GetTodayIncome` > 0 di router dengan data nyata.

---

## 9. FASE 6 — Modul 09: Expire Monitor (gap: 4 prosedur)

**Status logika:** `BuildExpireMonitorScript`, `NewSetupMikhmonExpireMonitorCommand`, `NewUpdateMikhmonExpireMonitorCommand` (expire.go) + `HotspotGateway.SetupExpireMonitor` ada. **Tidak ada** gateway method untuk status/enable/disable/remove scheduler.

### 9.1 Keputusan kompatibilitas (putuskan SEBELUM implementasi)

| Opsi | Deskripsi | Konsekuensi |
| :--- | :--- | :--- |
| A (rekomendasi) | Status check mengenali **kedua** bentuk: legacy `Mikhmon-Expire-Monitor` **dan** gateway `mikhmon-expire-scheduler`. `SetupExpireMonitor` idempotent: ada legacy → update legacy; tidak ada → buat gaya gateway. | Kompatibel dengan router yang sudah terpasang Mikhmon asli. |
| B | Standarisasi penuh ke gaya gateway (2 langkah). | Sederhana, tapi router legacy terlihat "not ready" sampai di-setup ulang. |

### 9.2 Tambahan gateway — `internal/driver/mikrotik/hotspot/gateway.go` (+ `port.HotspotGateway`)

```go
// GetExpireMonitorStatus returns install/enabled state of the expire monitor
// scheduler (mengenali nama legacy & gateway, lihat keputusan Fase 6 §9.1).
GetExpireMonitorStatus(ctx context.Context, driver DeviceDriver) (ExpireMonitorStatus, error)

// SetExpireMonitorDisabled toggles scheduler disabled flag by RouterOS .id.
SetExpireMonitorDisabled(ctx context.Context, driver DeviceDriver, rosID string, disabled bool) (command.Result, error)

// RemoveExpireMonitor deletes the scheduler by RouterOS .id.
RemoveExpireMonitor(ctx context.Context, driver DeviceDriver, rosID string) (command.Result, error)
```
Tipe baru di `internal/port/`:
```go
type ExpireMonitorStatus struct {
    IsInstalled   bool
    IsEnabled     bool
    SchedulerID   string // .id untuk disable/remove
    SchedulerName string // nama yang ditemukan
}
```
Usecase: wrapper `GetExpireMonitorStatus`, `DisableExpireMonitor`, `RemoveExpireMonitor` di `hotspot_usecase.go`.

### 9.3 Proto

```protobuf
message GetExpireMonitorStatusRequest { string device_id = 1; }
message ExpireMonitorStatusResponse {
  bool   is_installed   = 1;
  bool   is_enabled     = 2;
  string status         = 3;  // "ok" | "not ready" (kompatibel frontend legacy)
  string scheduler_name = 4;
}
message SetupExpireMonitorRequest { string device_id = 1; string interval = 2; } // "00:01:00"
message SetupExpireMonitorResponse { string message = 1; }
message DisableExpireMonitorRequest { string device_id = 1; }
message DisableExpireMonitorResponse { string message = 1; }
message RemoveExpireMonitorRequest { string device_id = 1; }
message RemoveExpireMonitorResponse { string message = 1; }
```
`service HotspotService` — tambahkan `GetExpireMonitorStatus`, `SetupExpireMonitor`, `DisableExpireMonitor`, `RemoveExpireMonitor`.

### 9.4 File
| File | Aksi |
| :--- | :--- |
| `api/proto/v1/hotspot.proto` + gen | message & rpc |
| `internal/port/hotspot_gateway.go` + `hotspot_expire.go` (baru) | interface + tipe status |
| `internal/driver/mikrotik/hotspot/gateway.go` | 3 method baru (cek scheduler via `/system/scheduler/print ?name=...`) |
| `internal/usecase/hotspot/hotspot_usecase.go` | 3 wrapper |
| `internal/adapter/connect/hotspot/expire_monitor_handler.go` | **baru** — 4 handler |
| `internal/adapter/connect/hotspot/router.go` | daftarkan rpc |

### 9.5 Logika wajib
1. **Status mapping (legacy):** scheduler ada & tidak disabled → `ok`; ada tapi disabled → `not ready` (is_installed=true); tidak ada → `not ready` (is_installed=false).
2. **Idempotensi install:** print dulu → ada: `set` (atau skip bila aktif); tidak ada: `add`.
3. **Validasi `interval`:** format durasi RouterOS (`HH:MM:SS`); default `00:01:00` (legacy) — samakan konstanta gateway.
4. **`Disable` ≠ `Remove`:** disable hanya set flag, remove hapus scheduler (script bisa dibiarkan).

### 9.6 Acceptance criteria
- Unit test status mapping (3 skenario).
- `SetupExpireMonitor` idempotent (2× panggil → tidak duplikat scheduler).
- `go build` + `go test ./...` hijau.

---

## 10. FASE 7 — Modul 10: Template & Print (gap: ekspos engine + metadata)

**Temuan §1.1:** engine render **sudah ada** di `pkg/voucher` (`Render`, QR base64, layout default/small/thermal). Yang kurang: prosedur ConnectRPC, template edit (opsional), QR warna, dan metadata router (Fase 0).

### 10.1 Proto

```protobuf
message TemplateInfo { string name = 1; repeated string sections = 2; } // name: default|small|thermal; sections: header,row,footer
message ListTemplatesRequest {}
message ListTemplatesResponse { repeated TemplateInfo templates = 1; }

message GetTemplateSectionRequest { string template_name = 1; string section = 2; } // section ∈ header|row|footer
message GetTemplateSectionResponse { string content = 1; }

message SaveTemplateSectionRequest { string template_name = 1; string section = 2; string content = 3; }
message SaveTemplateSectionResponse { string message = 1; }

message RenderVouchersRequest {
  string device_id    = 1;
  string template_name = 2; // default | small | thermal
  string comment      = 3;  // batch (eksklusif dengan user_id)
  string user_id      = 4;  // single (eksklusif dengan comment)
  bool   preview      = 5;  // data dummy (mikhmon/1234)
}
message RenderVouchersResponse { string html = 1; int32 total_vouchers = 2; }
```
`service HotspotService` — tambahkan `ListTemplates`, `GetTemplateSection`, `SaveTemplateSection`, `RenderVouchers`.

### 10.2 File
| File | Aksi |
| :--- | :--- |
| `api/proto/v1/hotspot.proto` + gen | message & rpc |
| `internal/adapter/connect/hotspot/template_handler.go` | **baru** — 4 handler |
| `internal/usecase/hotspot/template_usecase.go` (atau method di `hotspot_usecase.go`) | baca template dari `embed.go` / DB; orkestrasi data voucher → `pkg/voucher.Render` |
| `pkg/voucher/generator.go` | (opsional) tambah `%qrCodeRed/Green/Blue%` + `%currency%` `%phone%` |
| `internal/adapter/connect/hotspot/router.go` | daftarkan rpc |
| **Fase 0** (metadata Device) | sumber `hotspotName`/`dnsName`/`currency`/`phone` |

### 10.3 Logika wajib
1. **Sumber data voucher:**
   - `user_id` → `UseCase.GetUser(ctx, driver, rosID)` (single).
   - `comment` → `GetVoucherBatch` / `GetUsersByTag` + filter `uptime=0s` (batch) — reuse Fase 4.
   - keduanya kosong & `preview=false` → `CodeInvalidArgument`.
   - `preview=true` → data dummy (`mikhmon`/`1234`), tanpa akses router (device_id boleh kosong).
2. **Metadata harga/validity:** baca dari on-login profile via `ParseOnLoginScript` (Fase 2) — index price = selling price, validity. Metadata router dari `Device.Mikhmon` (Fase 0).
3. **Login URL QR:** `http://<dns_name>/login?username=<urlencode(user)>&password=<urlencode(pass)>` — `pkg/voucher` saat ini meng-encode `username\npassword`; **ubah/dukung keduanya** (mode URL) agar sama dengan legacy.
4. **Template edit:** `embed.go` read-only. Untuk `SaveTemplateSection` → simpan override di DB (tabel `template_overrides`) atau lewati fase ini (read-only). Rekomendasi: fase ini hanya `List` + `Get` (read-only); `Save` ditandai opsional lanjutan.
5. **Response:** `html` siap `window.print()`; `Content-Type: text/html; charset=utf-8` di sisi client/render.
6. **Cache-buster logo:** `%logo%` + `?<YYYYmmddHHMMSS>`.

### 10.4 Acceptance criteria
- `RenderVouchers(preview=true)` menghasilkan HTML lengkap dengan data dummy tanpa koneksi router.
- `RenderVouchers(user_id=*)` mengambil user nyata + metadata profile/device dan render N row.
- Unit test `pkg/voucher` (sudah ada? — tambah bila belum) untuk tiap layout & placeholder.

---

## 11. FASE 8 — Test & Verifikasi Akhir

### 11.1 Unit test yang wajib ada (folder sama dengan kode)
| Area | File test |
| :--- | :--- |
| Comment vc/up + rebuild rules | `internal/driver/mikrotik/hotspot/comment_test.go` (sebagian ada) |
| `ParseOnLoginScript` | `internal/driver/mikrotik/hotspot/profile_test.go` |
| `ParseDataLimit` / charset | `internal/driver/mikrotik/hotspot/voucher_test.go` |
| `ParseMikhmonTransactions` (+ record rusak) | `internal/driver/mikrotik/hotspot/report_test.go` |
| Expire status mapping | `internal/driver/mikrotik/hotspot/expire_test.go` |
| Mapper proto (user/profile/report/host) | `internal/adapter/connect/hotspot/mapper_*_test.go` |
| `pkg/voucher` render (3 layout, QR, preview) | `pkg/voucher/generator_test.go` |

### 11.2 Checklist wajib (AGENTS.md §4)
- [ ] `make proto` (bila proto berubah) + `go build -v ./cmd/server`
- [ ] `go test ./...` 100% hijau
- [ ] Ukuran file ≤ 400–500 baris (pecah handler/mapper per modul)
- [ ] Tidak ada pelanggaran boundary (`domain` bebas proto/adapter; usecase hanya `domain`+`port`)
- [ ] Logging via `pkg/logger`, error via `fmt.Errorf("%w")` + `response.MapDomainError`
- [ ] `go mod tidy` bila menambah dependency (mis. tidak perlu — `skip2/go-qrcode` sudah ada)

### 11.3 Urutan commit yang disarankan
1. `feat(proto): device mikhmon metadata + hotspot gap messages` (Fase 0 + semua proto, satu `make proto`)
2. `feat(hotspot): user CRUD procedures` (Fase 1)
3. `feat(hotspot): profile CRUD + on-login parser` (Fase 2)
4. `feat(hotspot): hosts & servers procedures` (Fase 3)
5. `feat(hotspot): voucher batch query + request fields` (Fase 4)
6. `fix(hotspot): sales report ListReports filter` + `feat(hotspot): report procedures` (Fase 5)
7. `feat(hotspot): expire monitor status & control` (Fase 6)
8. `feat(hotspot): template & voucher render procedures` (Fase 7)
9. `test(hotspot): unit tests fase 1–8` + update status modul di `docs/mikhmon/*.md` + `ANALISIS_REST_API_MIKHMON_GOLANG.md`

---

## 12. Keputusan Terbuka (perlu konfirmasi sebelum implementasi)

| # | Keputusan | Opsi | Rekomendasi |
| :--- | :--- | :--- | :--- |
| 1 | Nama scheduler expire monitor | A: kompatibel dua bentuk; B: standar gateway | **A** (router legacy yang sudah terpasang tidak boleh "not ready") |
| 2 | Default interval expire monitor | `00:01:00` (legacy) vs `00:05:00` (gateway) | **`00:01:00`** (samakan legacy) |
| 3 | `ListReports` tanpa filter | kembalikan semua vs `CodeInvalidArgument` | kembalikan semua + `summary_only` |
| 4 | Metadata Device: kolom DB vs `Extra` map | typed columns vs map | **typed columns** + migration |
| 5 | `SaveTemplateSection` | read-only (embed) vs simpan override di DB | read-only dulu; Save = fase lanjutan |
| 6 | QR mode | `username\npassword` (saat ini) vs login URL (legacy) | dukung keduanya (mode URL default) |

---

## 13. Ringkasan Berkas (semua fase)

**Proto:** `api/proto/v1/hotspot.proto`, `api/proto/v1/device.proto` (+ `make proto`).

**Baru:** `user_handler.go`, `profile_handler.go`, `host_server_handler.go`, `voucher_handler.go`, `report_handler.go`, `expire_monitor_handler.go`, `template_handler.go` (semua di `internal/adapter/connect/hotspot/`); `mapper_user.go`, `mapper_profile.go`, `mapper_report.go` (folder sama); `internal/port/hotspot_expire.go`; `internal/driver/mikrotik/hotspot/gateway.go` (method baru); migration baru.

**Diubah:** `router.go` (daftar rpc), `mapper.go` (perbaikan `ToProtoHotspotProfiles`/`ToProtoHotspotUsers`), `internal/usecase/hotspot/hotspot_usecase.go`, `internal/port/hotspot_gateway.go`, `internal/domain/device/device.go`, `internal/adapter/postgres/device_repository.go`, `internal/adapter/connect/device/*`, `pkg/voucher/generator.go` (opsional QR warna/URL), `docs/mikhmon/*.md` (update status setelah selesai).

---

*Dokumen ini living plan — setelah setiap fase selesai, tandai di bagian "Ringkasan Status Gap" (§0) dan perbarui badge status di `README.md` + modul terkait.*
