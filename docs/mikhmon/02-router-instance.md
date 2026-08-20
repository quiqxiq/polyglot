# Modul 02 — Device / Router Instance Management

> Kembali ke [README](README.md) · Kode asli: `post/post_a_router.php` (add/remove/save/saveAdmin), `get/get_connect.php`, `view/admin/settings.php` (upload logo), `config/config.php`, `core/page_route.php` (`decode`), `core/routeros_api.class.php` (`enc_rypt`/`dec_rypt`).
>
> **Status implementasi di Polyglot: ✅ selesai (bentuk berbeda)** — legacy "router instance" (session di file `config.php`) diganti **inventaris device di PostgreSQL** dengan `DeviceService`. Credential disimpan di vault (AES), bukan `enc_rypt()`.

## 1. Pemetaan Legacy

| Fungsi legacy | Trigger | Request asli | Response asli |
| :--- | :--- | :--- | :--- |
| Add router | `router_` + `do=add` | `router_` (diawali `sess_`) | `{"message":"Success","sesname":"sessionXXX<rand>"}` |
| Remove router | `do=remove` | `router_` (session) | `{"message":"Success"}` + hapus `assets/img/logo-<session>.png` |
| Save router | `do=save` | `router_`, `session`, `ipmik`(encode), `usermik`(encode), `passmik`(encode), `hotspotname`, `dnsname`, `currency`, `email`(encode), `infolp`, `idleto`, `phone`(encode), `report` | `{"message":"Success","sess":"<newname>"}` atau pesan duplikat/diperbaiki |
| Save admin | `do=saveAdmin` | `username`(encode), `password`(encode) | `{"message":"Success"}` (lihat modul 01) |
| Test connect | `connect` (GET) | – | Teks: `Connected,` / `Invalid username or password,` / `Error,` |
| Upload logo | `view/admin/settings.php` | multipart `.png` | simpan `assets/img/logo-<session>.png` |

**Catatan `do=add`:** menghasilkan nama session acak `session<rand 100-999>` dengan field default kosong. Di Polyglot, device diidentifikasi oleh **UUID** (`google/uuid`) yang dibuat handler saat create.

## 2. Mapping ke Polyglot (ConnectRPC — `DeviceService`)

Prosedur dipanggil `POST /polyglot.v1.DeviceService/<Procedure>` (protected: JWT + RBAC). Definisi proto: `api/proto/v1/device.proto`; handler: `internal/adapter/connect/device/device_handler.go` + `router.go`; usecase: `internal/usecase/device/manage_device.go`; repositori: `internal/adapter/postgres/device_repository.go`; vault: `internal/adapter/postgres/credential_vault.go`; registry driver: `internal/registry/registry.go`.

### 2.1 List Devices

- **Prosedur:** `DeviceService/ListDevices` — `ListDevicesRequest{}` → `ListDevicesResponse{devices[]}`
- **Handler:** `DeviceConnectHandler.ListDevices` — `ManageDeviceUseCase.ListDevices`.
- **Message `Device`:** `{id, tenant_id, name, vendor, driver_type, host, port, timeout_ms, poll_interval_ms, tags[], enabled, ssh_port}`. `password` **tidak pernah** dikembalikan (hanya field input pada `UpdateDeviceRequest`).

### 2.2 Get / Create / Update / Delete Device

| Operasi | Prosedur | Catatan |
| :--- | :--- | :--- |
| Detail | `DeviceService/GetDevice` — `GetDeviceRequest{id}` | 404 → `CodeNotFound` (`device.ErrNotFound`) |
| Create/Update | `DeviceService/UpdateDevice` — `UpdateDeviceRequest{device, username, password}` | Satu prosedur: bila `device.id` kosong → **create** (UUID baru, `isNew=true`), selain itu → **update**. Menyimpan credential ke vault. |
| Delete | `DeviceService/DeleteDevice` — `DeleteDeviceRequest{id}` | Hapus record + credential vault |

**Response Update:** `UpdateDeviceResponse{device, message}` — pesan berbeda untuk create vs update (`"device created successfully via ConnectRPC"` / `"device updated successfully via ConnectRPC"`).

### 2.3 Test Connection (pengganti `connect`)

- **Prosedur:** `DeviceService/TestDeviceConnection` — `TestDeviceConnectionRequest{id, selected_interface}`
- **Handler:** `DeviceConnectHandler.TestDeviceConnection` — mengambil driver dari registry (`DriverGetter`), lalu `ManageDeviceUseCase.TestConnection`.
- **Response `TestDeviceConnectionResponse`:** `{device_id, status, message, success, latency_ms, uptime, version, board_name, identity, cpu_load, free_memory, total_memory, interfaces[], interface_list[]}` — setara info `connect` + `get_sys_resource` legacy.
- **Gagal konek:** `success=false` + `status="failed"` + pesan error (tidak throw — dipakai UI untuk menampilkan status).

### 2.4 Streaming status / ping / traffic (bonus — tidak ada di legacy)

| Prosedur | Deskripsi |
| :--- | :--- |
| `DeviceService/StreamDeviceStatus` | Streaming frame `{device, test_metrics}` berkala |
| `DeviceService/StreamPing` | Streaming `StreamDevicePingFrame{device_id, address, latency_ms, status}` |
| `DeviceService/StreamInterfaceTraffic` | Streaming `StreamDeviceTrafficFrame{device_id, interface_name, rx_bps, tx_bps}` |
| `DeviceService/StreamTerminal` | Bidirectional stream `TerminalFrame` (PTY SSH) |

Terminal juga tersedia via WebSocket: `GET /ws/devices/{id}/terminal` (`internal/adapter/ws/device_stream_handler.go`, `network.OpenTerminalUseCase` + `genericssh.DialSSHPty`).

### 2.5 Upload Logo

- 🔴 **Belum diimplementasikan.** Legacy menyimpan `assets/img/logo-<session>.png`. Rencana: prosedur baru di `DeviceService` (atau HTTP endpoint) dengan multipart upload + penyimpanan di storage/DB, URL dipakai placeholder `%logo%` modul 10.

## 3. Tipe Data (proto / domain)

```protobuf
// api/proto/v1/device.proto
message Device {
  string id = 1; string tenant_id = 2; string name = 3; string vendor = 4;
  string driver_type = 5; string host = 6; int32 port = 7; int32 timeout_ms = 8;
  int32 poll_interval_ms = 9; repeated string tags = 10; bool enabled = 11;
  int32 ssh_port = 12;
}
message UpdateDeviceRequest { Device device = 1; string username = 2; string password = 3; }
message UpdateDeviceResponse { Device device = 1; string message = 2; }
message TestDeviceConnectionRequest  { string id = 1; string selected_interface = 2; }
message TestDeviceConnectionResponse { /* lihat §2.3 */ }
```

Domain: `internal/domain/device/device.go` (termasuk `Target` — host/port/credential dari vault) + `credentials.go` + `errors.go`.

## 4. Logika Khusus

### 4.1 Penyimpanan & keamanan

- Device disimpan di tabel `devices` (Postgres); `driver_type` menentukan pabrik driver (`mikrotik`, `genieacs`, dst.) di `registry.DriverFactory`.
- Credential (username/password) dienkripsi AES di vault — tidak pernah dikembalikan di response (field input-only).
- `host` = IP/domain, `port` = port API (8728 untuk RouterOS), `ssh_port` terpisah untuk terminal.

### 4.2 Enkripsi legacy & parser config (hanya untuk migrasi)

Logika `decode()` / `enc_rypt()` / `dec_rypt()` dari legacy **belum di-port** ke Polyglot. Bila diperlukan migrasi `config/config.php` → tabel `devices`:

```go
package legacy

import (
    "encoding/base64"
    "strings"
)

// decode(): base64 -> XOR 10 -> double base64  (input form: ipmik, usermik, passmik, email, phone)
func Decode(data string) string {
    once, err := base64.StdEncoding.DecodeString(data)
    if err != nil { return "" }
    xored := make([]byte, len(once))
    for i, b := range once { xored[i] = b ^ 10 }
    onceMore, err := base64.StdEncoding.DecodeString(string(xored))
    if err != nil { return "" }
    twice, err := base64.StdEncoding.DecodeString(string(onceMore))
    if err != nil { return "" }
    return string(twice)
}

var legacyKey = "128" // key berulang: '1','2','8'

func Encrypt(s string) string {
    out := make([]byte, len(s))
    for i := 0; i < len(s); i++ { out[i] = s[i] + legacyKey[i%len(legacyKey)] }
    return base64.StdEncoding.EncodeToString(out)
}

func Decrypt(s string) string {
    raw, err := base64.StdEncoding.DecodeString(s)
    if err != nil { return "" }
    out := make([]byte, len(raw))
    for i := 0; i < len(raw); i++ { out[i] = raw[i] - legacyKey[i%len(legacyKey)] }
    return string(out)
}

// GetConfig: membaca satu nilai dari baris config legacy.
// contoh: 'router01!192.168.88.1:8728' dengan start "router01!" end "'"
func GetConfig(line, start, end string) string {
    idx := strings.Index(line, start)
    if idx == -1 { return "" }
    idx += len(start)
    endIdx := strings.Index(line[idx:], end)
    if endIdx == -1 { endIdx = len(line) - idx }
    return line[idx : idx+endIdx]
}
```

### 4.3 Endpoint migrasi (rekomendasi — belum ada)

- Prosedur admin baru: terima upload `config/config.php` lama, parse per baris dengan tabel separator (README §2.2), tulis ke tabel `devices`; password di-decode (`Decode` + `Decrypt`) lalu di-encrypt ulang AES ke vault. Field Mikhmon yang tidak ada di `Device` (hotspot name, dns name, currency, phone, email, info_lp, idle_timeout, live_report) memerlukan kolom/attr tambahan di domain device (contoh: map `attributes`) — sesuaikan proto bila dibutuhkan.
