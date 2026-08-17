# Modul 11 — Quick Resources & Theme/Misc

> Kembali ke [README](README.md) · Kode asli: `get/get_interface.php`, `get/get_addr_pool.php`, `get/get_parent_queue.php`, `get/get_nat.php`, `config/settheme.php`.
>
> **Status implementasi di Polyglot: 🟡 sebagian** — `ListIPPools`, `ListParentQueues`, `ListNATRules`, `ListHotspotServers` sudah ada di gateway/usecase; **belum diekspos** via ConnectRPC. Theme/settings belum diimplementasikan.

## 1. Pemetaan Legacy

| Fungsi legacy | Request asli | Command RouterOS | Format Response Asli |
| :--- | :--- | :--- | :--- |
| `get_interface` | – | `/interface/print` | jsEncode25 array |
| `get_addr_pool` | `f` | `/ip/pool/print` | **JSON polos** |
| `get_parent_queue` | `f` | `/queue/simple/print ?dynamic=false` | **JSON polos** |
| `get_nat` | – | `/ip/firewall/nat/print` | **JSON polos** |
| `set_theme` (admin) | `theme` ∈ dark/light/blue/green/pink | – (menulis `config/theme.php`) | HTML/redirect |

> Semua resource read-only — dipakai untuk dropdown form (pilih interface, pool, parent queue) dan pengaturan firewall. Theme di legacy bersifat global per instalasi; di Polyglot bisa jadi preferensi per-user di DB.

## 2. Mapping ke Polyglot (ConnectRPC — `HotspotService`)

Prosedur dipanggil `POST /polyglot.v1.HotspotService/<Procedure>` (protected: JWT + RBAC). Proto: `api/proto/v1/hotspot.proto` (message khusus belum ada); usecase: `internal/usecase/hotspot/hotspot_usecase.go`; gateway: `internal/driver/mikrotik/hotspot/gateway.go`.

### 2.1 Resources — 🟡 logika ada di gateway/usecase, belum diekspos

| Fungsi legacy | Usecase (`UseCase.Get*`) | Gateway (`HotspotGateway.List*`) | RouterOS |
| :--- | :--- | :--- | :--- |
| `get_addr_pool` | `GetIPPools` | `ListIPPools` | `/ip/pool/print` → `[]port.IPPool` (`ParseIPPools`) |
| `get_parent_queue` | `GetParentQueues` | `ListParentQueues` | `/queue/simple/print` → filter parent (`Parent == "none"`) → `[]port.SimpleQueue` |
| `get_nat` | `GetNATRules` | `ListNATRules` | `/ip/firewall/nat/print` → `[]map[string]string` |
| `get_hotspot_server` | `GetHotspotServers` | `ListHotspotServers` | `/ip/hotspot/print` → `[]map[string]string` |
| `get_interface` | – | – | 🔴 belum ada di gateway (perlu command `/interface/print` baru) |

**Rencana prosedur:**
- `HotspotService/ListIPPools` — `{device_id}` → `{pools: [{id, name, ranges}]}` (proto `IPPool`).
- `HotspotService/ListParentQueues` — `{device_id}` → `{queues: [{id, name, target, max_limit, disabled}]}` (proto `SimpleQueue`).
- `HotspotService/ListNATRules` — `{device_id}` → `{rules: [...]}` (proto `NATRule`).
- `HotspotService/ListHotspotServers` — `{device_id}` → `{servers: [...]}`.
- `HotspotService/ListInterfaces` — `{device_id}` → `{interfaces: [...]}` (baru; logika belum ada).

**Tipe port:** `port.IPPool`, `port.SimpleQueue` (`internal/port/`); parsing di `internal/driver/mikrotik/` (`ParseIPPools`, `ParseSimpleQueues`). NAT/host/servers berupa `[]map[string]string` (raw rows) — konversi ke proto terstruktur di mapper.

### 2.2 Theme & Settings — 🔴 belum diimplementasikan

- Legacy `set_theme` menulis `config/theme.php`; di Polyglot belum ada settings service. Rencana: simpan preferensi per-user di tabel users/settings (bukan file global).
- Prosedur usulan di `UserService` atau service baru `SettingsService`: `SetTheme` (`{theme: "dark"|"light"|"blue"|"green"|"pink"}`) + `GetSettings`.

### 2.3 Healthcheck / Metrics

- Proyek sudah punya `ProbeService` (`internal/adapter/connect/device/probe_handler.go`, mount `/polyglot.v1.ProbeService/`) — healthcheck service-level.
- Metrics Prometheus belum ada — opsional.

## 3. Tipe Data (proto usulan)

```protobuf
// tambahan di api/proto/v1/hotspot.proto (usulan)
message IPPool     { string id = 1; string name = 2; string ranges = 3; }
message SimpleQueue { string id = 1; string name = 2; string target = 3; string max_limit = 4; bool disabled = 5; }
message NATRule    { string id = 1; string chain = 2; string action = 3; string to_address = 4; string to_ports = 5; string dst_address = 6; string protocol = 7; bool disabled = 8; string comment = 9; }
```

## 4. Logika Khusus

1. **Filter parent queue:** `ListParentQueues` mengembalikan hanya queue dengan `Parent == "none"` atau kosong (dipakai dropdown parent queue profile, modul 05).
2. **Konversi bool dari string RouterOS** (`"true"`/`"false"`) via helper yang sama dengan modul 06.
3. **NAT rules bisa panjang** — tambahkan `limit`/`offset` pada proto bila perlu; default kembalikan semua sesuai legacy.
4. **Read-only:** semua resource ini tidak memiliki operasi tulis di legacy; jangan tambahkan mutasi kecuali ada kebutuhan baru (tandai sebagai ekstensi).
