# Modul 03 — Dashboard & Live Monitoring (Streaming Terpisah)

> Kembali ke [README](README.md) · Kode asli: `get/get_dashboard.php` (meng-handle `get_sys_resource`, `get_hotspotinfo`, `get_traffic`, `get_log`, `get_livereport`), `assets/js/mikhmon.js` (polling 10 detik).
>
> **Status implementasi di Polyglot: ✅ selesai (Modul 03)** — `GetDashboard` (satu endpoint agregasi) **dihapus total** dan diganti **10 prosedur streaming terpisah per-area**, semuanya streaming dari MikroTik → backend → frontend (ConnectRPC server-streaming).

## 1. Prinsip Desain (Berbeda dari Legacy)

Legacy `get_dashboard.php` mengambil **semua data sekaligus** dalam satu request (polling ~10 detik). Di Polyglot:

1. **Tidak ada endpoint agregasi.** Setiap area punya prosedur streaming sendiri (system snapshot, interface, queue, log, traffic, session aktif/inaktif hotspot, session aktif/inaktif PPP). Frontend memanggil hanya yang dibutuhkan, per komponen dashboard.
2. **Streaming dari router.** Semua data realtime memakai mode streaming native RouterOS (`follow` / `interval=1s` / `monitor-traffic`), bukan polling backend. Router mengirim data → backend meneruskan → frontend.
3. **Frame lengkap per interval.** Untuk data yang butuh *full list* (inactive users, queue stats), dipakai `interval=1s` (RouterOS mengirim ulang seluruh list), bukan `follow` (yang hanya mengirim delta per baris).
4. `get_livereport` dipindah ke modul 08 (report).

## 2. Pemetaan Legacy → Prosedur Streaming

| Fungsi legacy | Prosedur Polyglot (server-streaming) | Sumber RouterOS |
| :--- | :--- | :--- |
| `get_sys_resource` (5 print) | `StreamSystemSnapshot` | `/system/clock/print interval=1s` + `/system/resource/print interval=1s` + `/system/routerboard/print interval=1s` + `/system/identity/print interval=1s` + `/system/health/print interval=1s` — 5 stream **native interval** digabung backend jadi 1 frame |
| `get_traffic` (per iface) | `StreamTraffic` | `/interface/monitor-traffic =interface=<nama>` (native streaming, tanpa `once`) |
| – (ethernet list) | `StreamInterfaceEthernet` | `/interface/ethernet/print interval=1s` |
| – (queue simple) | `StreamQueueStats` | `/queue/simple/print stats interval=1s` (filter `dynamic=false`, parent, atau nama) |
| `get_log` | `StreamLogs` | `/log/print follow` (opsional `?topics~=`) |
| `get_hotspotinfo` (active) | `StreamActiveSessions` | `/ip/hotspot/active/print follow` |
| – (inactive hotspot) | `StreamHotspotInactive` | `/ip/hotspot/user/print interval=1s` + `/ip/hotspot/active/print interval=1s` → selisih (`FilterInactiveHotspotUsers`) |
| – (PPP active) | `StreamPPPActive` | `/ppp/active/print follow` |
| – (PPP inactive) | `StreamPPPInactive` | `/ppp/secret/print interval=1s` + `/ppp/active/print interval=1s` → selisih (`FilterInactivePPPoESecrets`) |

Semua prosedur dipanggil `POST /polyglot.v1.HotspotService/<Procedure>` (protected: JWT + RBAC, `hotspot:read`).

## 3. Mapping ke Polyglot (ConnectRPC — `HotspotService`)

Handler: `internal/adapter/connect/hotspot/monitor_*.go` (satu file per area); usecase: `internal/usecase/hotspot/hotspot_usecase.go`; driver: `internal/driver/mikrotik/` (submodul per fitur).

### 3.1 System Snapshot — `StreamSystemSnapshot` (`monitor_system_handler.go`)

- **Prosedur:** `StreamSystemSnapshotRequest{device_id, interval}` → `SystemSnapshotFrame{device_id, clock, resource, routerboard, identity, health, timestamp_unix}`.
- **Handler:** membuka **5 stream native** (`interval=1s`, dapat di-override via `interval`): clock, resource, routerboard, identity, health. Setiap kali 5 frame tersedia dalam satu putaran, digabung menjadi 1 `SystemSnapshotFrame` dan dikirim. Stream putus/tidak tersedia → error field kosong, frame tetap dikirim (best-effort).
- **Parser:** `ParseSystemClock`, `ParseSystemResource`, `ParseSystemRouterboard`, `ParseSystemIdentity`, `ParseSystemHealth`.
- **Catatan penting:** kelima print ini mendukung `interval` di RouterOS (RouterOS mengirim `!re` berulang) — **bukan polling backend**.

### 3.2 Interface Traffic — `StreamTraffic` (`monitor_interface_handler.go`)

- **Prosedur:** `StreamTrafficRequest{device_id, interface}` → `TrafficStreamData{device_id, interface, rx_bps, tx_bps, timestamp_unix}`.
- **Handler:** `NewMonitorTrafficStreamCommand(iface)` → per frame parse `ParseInterfaceTrafficStats`. Default interface `ether1` bila kosong.

### 3.3 Interface Ethernet — `StreamInterfaceEthernet` (`monitor_interface_handler.go`)

- **Prosedur:** `StreamInterfaceEthernetRequest{device_id, name_filter, interval}` → `InterfaceEthernetFrame{device_id, interfaces[], timestamp_unix}`.
- **Handler:** `NewStreamInterfacesCommand(name_filter, interval)` → per frame `ParseInterfaces` (full list tiap interval).

### 3.4 Queue Simple Stats — `StreamQueueStats` (`monitor_queue_handler.go`)

- **Prosedur:** `StreamQueueStatsRequest{device_id, name_filter, parent_filter, parents_only, interval}` → `QueueStatsFrame{device_id, queues[], timestamp_unix}`.
- **Handler:** `NewStreamQueueStatsCommand(QueueStreamParams{...})` (`/queue/simple/print stats interval=1s`, `?dynamic=false` + `?parent=` bila parents_only/parent_filter) → per frame `ParseSimpleQueueStats`.

### 3.5 Log — `StreamLogs` (`monitor_log_handler.go`)

- **Prosedur:** `StreamLogsRequest{device_id, topics_filter}` → `LogsStreamFrame{device_id, logs[], timestamp_unix}`.
- **Handler:** `NewStreamLogsCommand(topics_filter)` (`/log/print follow`) → per frame `ParseLogs`.

### 3.6 Active Sessions — `StreamActiveSessions` (`monitor_session_handler.go`)

- **Prosedur:** `StreamActiveSessionsRequest{device_id, user_filter}` → `ActiveSessionsStreamData{device_id, sessions[], timestamp_unix}`.
- **Handler:** `NewStreamHotspotActiveCommand(user_filter)` (`follow`, delta per baris) → per frame `ParseHotspotActiveSessions`.

### 3.7 Inactive Hotspot — `StreamHotspotInactive` (`monitor_session_handler.go`)

- **Prosedur:** `StreamHotspotInactiveRequest{device_id, profile_filter, interval}` → `HotspotInactiveFrame{device_id, users[], timestamp_unix}`.
- **Handler:** gabungan **2 stream interval**: `/ip/hotspot/user/print interval=<n>` (full list user) + `/ip/hotspot/active/print interval=<n>` (full list active). Dalam satu putaran: selisih `FilterInactiveHotspotUsers(users, active)` → frame user yang tidak sedang online. Full-snapshot tiap interval (bukan delta).

### 3.8 PPP Active — `StreamPPPActive` (`monitor_ppp_handler.go`)

- **Prosedur:** `StreamPPPActiveRequest{device_id, name_filter}` → `PPPActiveFrame{device_id, sessions[], timestamp_unix}`.
- **Handler:** `NewStreamPPPActiveCommand(name_filter)` (`follow`, delta per baris) → `ParsePPPActiveSessions`.

### 3.9 PPP Inactive — `StreamPPPInactive` (`monitor_ppp_handler.go`)

- **Prosedur:** `StreamPPPInactiveRequest{device_id, name_filter, interval}` → `PPPInactiveFrame{device_id, secrets[], timestamp_unix}`.
- **Handler:** gabungan **2 stream interval**: `/ppp/secret/print interval=<n>` + `/ppp/active/print interval=<n>` → selisih `FilterInactivePPPoESecrets(secrets, active)` → frame subscriber yang sedang offline.

### 3.10 System Resource — `StreamResource` (`monitor_system_handler.go`)

- **Prosedur:** `StreamResourceRequest{device_id, interval}` → `ResourceStreamData{device_id, cpu_load, free_memory, uptime, timestamp_unix}`.
- **Handler:** `NewStreamSystemResourceCommand(interval)` (`/system/resource/print interval=<n>`, native) → per frame `ParseSystemResource`.

## 4. Tipe Data (proto — `api/proto/v1/hotspot.proto`)

```protobuf
// System snapshot (5 stream digabung)
message SystemSnapshotFrame {
  string device_id = 1;
  SystemClock clock = 2;
  SystemResourceInfo resource = 3;
  SystemRouterboard routerboard = 4;
  SystemIdentity identity = 5;
  SystemHealth health = 6;
  int64 timestamp_unix = 7;
}
message StreamSystemSnapshotRequest { string device_id = 1; string interval = 2; }

// Interface ethernet
message InterfaceEthernetFrame { string device_id = 1; repeated NetworkInterface interfaces = 2; int64 timestamp_unix = 3; }
message StreamInterfaceEthernetRequest { string device_id = 1; string name_filter = 2; string interval = 3; }

// Queue simple stats
message QueueStatsFrame { string device_id = 1; repeated SimpleQueueStats queues = 2; int64 timestamp_unix = 3; }
message StreamQueueStatsRequest {
  string device_id = 1; string name_filter = 2;
  string parent_filter = 3; bool parents_only = 4; string interval = 5;
}

// Log
message LogsStreamFrame { string device_id = 1; repeated LogEntry logs = 2; int64 timestamp_unix = 3; }
message StreamLogsRequest { string device_id = 1; string topics_filter = 2; }

// Active / inactive hotspot
message HotspotInactiveFrame { string device_id = 1; repeated HotspotUser users = 2; int64 timestamp_unix = 3; }
message StreamHotspotInactiveRequest { string device_id = 1; string profile_filter = 2; string interval = 3; }

// Active / inactive PPP
message PPPActiveFrame { string device_id = 1; repeated PPPActiveSession sessions = 2; int64 timestamp_unix = 3; }
message StreamPPPActiveRequest { string device_id = 1; string name_filter = 2; }
message PPPInactiveFrame { string device_id = 1; repeated PPPSecret secrets = 2; int64 timestamp_unix = 3; }
message StreamPPPInactiveRequest { string device_id = 1; string name_filter = 2; string interval = 3; }

// System resource (native interval)
message StreamResourceRequest { string device_id = 1; string interval = 2; }
```

## 5. File Implementasi

| File | Isi |
| :--- | :--- |
| `internal/adapter/connect/hotspot/monitor_system_handler.go` | `StreamSystemSnapshot` (5-stream gabung) + `StreamResource` |
| `internal/adapter/connect/hotspot/monitor_interface_handler.go` | `StreamTraffic` + `StreamInterfaceEthernet` |
| `internal/adapter/connect/hotspot/monitor_queue_handler.go` | `StreamQueueStats` |
| `internal/adapter/connect/hotspot/monitor_log_handler.go` | `StreamLogs` |
| `internal/adapter/connect/hotspot/monitor_session_handler.go` | `StreamActiveSessions` + `StreamHotspotInactive` |
| `internal/adapter/connect/hotspot/monitor_ppp_handler.go` | `StreamPPPActive` + `StreamPPPInactive` |
| `internal/driver/mikrotik/system_health.go` | `SystemHealth`, `NewStreamSystemHealthCommand`, `ParseSystemHealth` |
| `internal/driver/mikrotik/system_routerboard.go` | `SystemRouterboard`, `NewStreamSystemRouterboardCommand`, `ParseSystemRouterboard` |
| `internal/driver/mikrotik/system_identity.go` | `SystemClock`, `NewStreamSystemClockCommand`, `ParseSystemClock` |
| `internal/driver/mikrotik/system_log.go` | `NewStreamLogsCommand`, `ParseLogs` |
| `internal/driver/mikrotik/iface.go` | `NewStreamInterfacesCommand` |
| `internal/driver/mikrotik/hotspot_user.go` / `hotspot_active.go` | `NewStreamHotspotUsersIntervalCommand`, `NewStreamHotspotActiveIntervalCommand` |
| `internal/driver/mikrotik/ppp.go` / `ppp_active.go` | `NewStreamPPPoESecretsIntervalCommand`, `NewStreamPPPActiveIntervalCommand` |

`GetDashboard`, `GetHotspotDashboardSummary`, dan `system_report_handler.go` **dihapus total** (proto, router, RBAC, handler) — frontend wajib memakai prosedur terpisah di atas.

## 6. Logika Khusus

1. **`follow` vs `interval`:** `follow` mengirim delta per baris (cocok: traffic, log, active). `interval=1s` mengirim **full list** tiap putaran (dibutuhkan untuk inactive user/PPP dan queue stats). Keduanya streaming native — tidak ada polling backend.
2. **Multi-stream gabung:** `StreamSystemSnapshot` dan `StreamHotspotInactive`/`StreamPPPInactive` membuka >1 stream sekaligus; frame dikirim per putaran lengkap, bagian yang gagal dikosongkan (best-effort), stream tetap berjalan.
3. **Koneksi per-request:** tidak ada pool persisten; driver diambil dari `registry` per request (`ConnectDriverProvider`).
4. **RBAC:** semua prosedur modul 03 butuh `hotspot:read`.
