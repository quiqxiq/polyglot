# Modul 06 — Hotspot Active, Host & Server

> Kembali ke [README](README.md) · Kode asli: `get/get_hotspot_active.php`, `get/get_hosts.php`, `get/get_hotspot_server.php`, `post/post_hotspot_remove.php` (`where=active_` / `where=host_`).
>
> **Status implementasi di Polyglot: ✅ selesai** — `ListActiveSessions` + `KickActiveSession` + `ListHosts` + `RemoveHost` + `ListHotspotServers` diekspos (plus DHCP lease yang **tidak ada di legacy**).

## 1. Pemetaan Legacy

| Fungsi legacy | Request asli | Command RouterOS | Format Response Asli |
| :--- | :--- | :--- | :--- |
| `get_hotspot_active` | – | `/ip/hotspot/active/print` | jsEncode25 array |
| `get_hosts` | – | `/ip/hotspot/host/print` | jsEncode25 array |
| `get_hotspot_server` | `f` | `/ip/hotspot/print` | jsEncode25 array |
| Remove active | `where=active_`, `id` | `/ip/hotspot/active/remove =.id=<id>` | `{"message":"success"}` / `{"message":"error"}` |
| Remove host | `where=host_`, `id` | `/ip/hotspot/host/remove =.id=<id>` | `{"message":"success"}` / `{"message":"error"}` |

## 2. Mapping ke Polyglot (ConnectRPC — `HotspotService`)

Prosedur dipanggil `POST /polyglot.v1.HotspotService/<Procedure>` (protected: JWT + RBAC). Proto: `api/proto/v1/hotspot.proto`; handler: `internal/adapter/connect/hotspot/session_handler.go`; usecase: `internal/usecase/hotspot/hotspot_usecase.go` + `internal/usecase/network/active_sessions.go`; gateway: `internal/driver/mikrotik/hotspot/gateway.go`.

### 2.1 List Active Sessions — ✅ diekspos

- **Prosedur:** `HotspotService/ListActiveSessions` — `ListHotspotActiveSessionsRequest{device_id}` → `ListHotspotActiveSessionsResponse{sessions[]}`
- **Handler:** `HotspotConnectHandler.ListActiveSessions` — `UseCase.GetActiveSessions` → `ToProtoActiveSessions`.
- **Message `HotspotActiveSession`:** `{id, server, user, address, mac_address, uptime, bytes_in, bytes_out, comment}`.
- Realtime-nya: `StreamActiveSessions` (modul 03).

### 2.2 Kick / Disconnect Active — ✅ diekspos

- **Prosedur:** `HotspotService/KickActiveSession` — `KickHotspotSessionRequest{device_id, ros_id}` → `KickHotspotSessionResponse{message}`
- **Handler:** `HotspotConnectHandler.KickActiveSession` — validasi `ros_id` (wajib), `UseCase.RemoveActiveSession` → `HotspotGateway.RemoveActiveSession` → `mikrotik.NewDisconnectHotspotActiveCommand` (`/ip/hotspot/active/remove`).
- **Catatan:** memutus sesi aktif, tidak menghapus user di `/ip/hotspot/user` — beda dengan delete user (modul 04).

### 2.3 List Hosts — ✅ diekspos

- **Prosedur:** `HotspotService/ListHosts` — `ListHotspotHostsRequest{device_id}` → `ListHotspotHostsResponse{hosts[]}`.
- **Handler:** `HotspotConnectHandler.ListHosts` (`host_server_handler.go`) → `UseCase.GetHosts` → `HotspotGateway.ListHosts` → `/ip/hotspot/host/print` → `res.Rows` → `ToProtoHotspotHosts` (mapper `mapper_host_server.go`; skip baris tanpa `.id`/`mac-address`; `bypassed`/`authorized` dari string `"true"`/`"false"`).
- **Message `HotspotHost`:** `{id, mac_address, address, to_address, server, bypassed, authorized, comment}`.

### 2.4 Delete Host — ✅ diekspos

- **Prosedur:** `HotspotService/RemoveHost` — `RemoveHotspotHostRequest{device_id, ros_id}` → `RemoveHotspotHostResponse{message}`.
- **Handler:** `HotspotConnectHandler.RemoveHost` (`host_server_handler.go`) — validasi `ros_id` (wajib), `UseCase.RemoveHost` → `HotspotGateway.RemoveHost` → `/ip/hotspot/host/remove =.id=<id>`. Hanya menghapus entri host, bukan user.

### 2.5 List Hotspot Servers — ✅ diekspos

- **Prosedur:** `HotspotService/ListHotspotServers` — `ListHotspotServersRequest{device_id}` → `ListHotspotServersResponse{servers[]}`.
- **Handler:** `HotspotConnectHandler.ListHotspotServers` (`host_server_handler.go`) → `UseCase.GetHotspotServers` → `HotspotGateway.ListHotspotServers` → `/ip/hotspot/print` → `res.Rows` → `ToProtoHotspotServers` (skip baris tanpa `.id`/`name`; `disabled` dari string `"true"`/`"false"`).
- **Message `HotspotServerInfo`:** `{id, name, interface, address_pool, disabled, comment}`.

### 2.6 DHCP Lease (baru — tidak ada di legacy) — ✅ diekspos

- **Prosedur:** `HotspotService/ListDHCPLeases` — `ListDHCPLeasesRequest{device_id, mac_filter}` → `ListDHCPLeasesResponse{leases[]}`; dan `HotspotService/BlockDHCPLease` — `BlockDHCPLeaseRequest{device_id, ros_id, blocked, comment}` → `BlockDHCPLeaseResponse{message}`.
- **Usecase:** `network.ActiveSessionsUseCase` (`GetDHCPLeases`, `SetDHCPLeaseBlock`); parsing `mikrotik.ParseDHCPLeases`; mapper `ToProtoDHCPLeases`.

## 3. Tipe Data (proto / port)

```protobuf
// api/proto/v1/hotspot.proto
message HotspotActiveSession {
  string id = 1; string server = 2; string user = 3; string address = 4;
  string mac_address = 5; string uptime = 6; string bytes_in = 7;
  string bytes_out = 8; string comment = 9;
}
message ListHotspotActiveSessionsRequest  { string device_id = 1; }
message ListHotspotActiveSessionsResponse { repeated HotspotActiveSession sessions = 1; }
message KickHotspotSessionRequest  { string device_id = 1; string ros_id = 2; }
message KickHotspotSessionResponse { string message = 1; }

message DHCPLease { string id = 1; string address = 2; string mac_address = 3; string host_name = 4; string status = 5; bool blocked = 6; string comment = 7; }
```

Tipe port: `port.HotspotActiveSession`; hosts/servers berbentuk `[]map[string]string` (raw rows) di gateway — dikonversi ke proto terstruktur di `mapper_host_server.go` (`ToProtoHotspotHosts` / `ToProtoHotspotServers`).

## 4. Logika Khusus

1. **Kick vs hapus:** `/ip/hotspot/active/remove` memutus sesi aktif; tidak menghapus user di `/ip/hotspot/user`. Dokumentasi API harus jelas membedakan dengan delete user (modul 04).
2. **Host `authorized`/`bypassed`:** nilai string `"true"`/`"false"` dari RouterOS → konversi bool di mapper.
3. **Paginasi:** daftar aktif/host bisa besar; tambahkan `limit`/`offset` pada proto bila diperlukan (legacy mengembalikan semua).
4. **Realtime:** frontend legacy memanggil active/hosts tiap beberapa detik — di Polyglot pakai `StreamActiveSessions` (server-streaming), bukan polling.
