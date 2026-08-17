# Modul 09 — Expire Monitor & Scheduler

> Kembali ke [README](README.md) · Kode asli: `get/get_expire_mon.php`, `post/post_expire_monitor.php`.
>
> **Status implementasi di Polyglot: ✅ selesai (Fase 6)** — 4 prosedur sudah diekspos via `HotspotService`: `GetExpireMonitorStatus`, `SetupExpireMonitor`, `DisableExpireMonitor`, `RemoveExpireMonitor`. Status check **kompatibel dua bentuk scheduler** (legacy `Mikhmon-Expire-Monitor` & gateway `mikhmon-expire-scheduler`); default interval `00:01:00` (legacy).

## 1. Pemetaan Legacy

| Fungsi legacy | Request asli | Command RouterOS | Format Response Asli |
| :--- | :--- | :--- | :--- |
| Status (`get_expire_mon`) | – | `/system/scheduler/print ?name=Mikhmon-Expire-Monitor ?disabled=false` | **JSON polos** `{"expire_monitor":"ok"}` / `{"expire_monitor":"not ready"}` |
| Install/enable (`post_expire_monitor`) | `sessname`, `expmon` (source script) | cek scheduler → `/system/scheduler/add` (interval `00:01:00`, start `00:00:00`, disabled no) atau `/system/scheduler/set` bila disabled | `{"message":"success"}` atau `{"message":"<name>"}` |

**Detail scheduler legacy:**

| Properti | Nilai |
| :--- | :--- |
| `name` | `Mikhmon-Expire-Monitor` |
| `start-time` | `00:00:00` |
| `interval` | `00:01:00` (tiap 1 menit) |
| `on-event` | script source dari UI (field `expmon`) |
| `comment` | `Mikhmon Expire Monitor` |
| `disabled` | `no` |

## 2. Mapping ke Polyglot (ConnectRPC — `HotspotService`)

Prosedur dipanggil `POST /polyglot.v1.HotspotService/<Procedure>` (protected: JWT + RBAC). Proto: `api/proto/v1/hotspot.proto` (message `GetExpireMonitorStatusRequest/Response`, `SetupExpireMonitorRequest/Response`, `DisableExpireMonitorRequest/Response`, `RemoveExpireMonitorRequest/Response`); handler: `internal/adapter/connect/hotspot/expire_monitor_handler.go`; usecase: `internal/usecase/hotspot/hotspot_usecase.go`; gateway: `internal/driver/mikrotik/hotspot/gateway.go`; script builder + klasifikasi status: `internal/driver/mikrotik/hotspot/expire.go`; tipe status: `internal/port/hotspot_expire.go`.

### 2.1 Script builder — ✅ sudah ada (`expire.go`)

- `BuildExpireMonitorScript()` — script RouterOS lengkap yang:
  1. Konversi `currentDate` dan `currentTime` ke integer.
  2. Cari semua user hotspot dengan comment mengandung tahun berjalan/tahun lalu.
  3. Parse tanggal/jam expire dan mode (`N`/`X`) dari comment.
  4. Bila expired: mode `N` → set `limit-uptime=1s` + putus sesi aktif; mode `X` → hapus user + putus sesi aktif.
- `NewSetupMikhmonExpireMonitorCommand(interval)` — command `/system/scheduler/add` (nama `Mikhmon-Expire-Monitor`, start `00:00:00`, interval default `00:01:00`, on-event script, comment "Mikhmon Expire Monitor", disabled no).
- `NewUpdateMikhmonExpireMonitorCommand(rosID, interval)` — command `/system/scheduler/set`.

### 2.2 Setup (install/update) — ✅ diekspos (`SetupExpireMonitor`)

`HotspotService/SetupExpireMonitor` — `{device_id, interval}` → **idempotent & kompatibel dua bentuk** (keputusan Fase 6 §9.1 = A):
1. Cek status via `GetExpireMonitorStatus`.
2. Ada scheduler legacy `Mikhmon-Expire-Monitor` → **update in-place** (`/system/scheduler/set`, interval + on-event script source) — router legacy tidak "not ready".
3. Ada scheduler gateway `mikhmon-expire-scheduler` → re-arm (`set` interval + on-event + `disabled=no`).
4. Tidak ada → buat gaya gateway dua langkah: `/system/script/add name=mikhmon-expire-monitor` + `/system/scheduler/add name=mikhmon-expire-scheduler`.
- **Interval default `00:01:00`** (keputusan #2 = legacy); validasi `HH:MM:SS` di handler (`CodeInvalidArgument` bila format salah).

### 2.3 Status — ✅ diekspos (`GetExpireMonitorStatus`)

`HotspotService/GetExpireMonitorStatus` — `{device_id}` → `{is_installed, is_enabled, status, scheduler_name}`.
- **Logika status (legacy, `classifyExpireMonitorSchedulers` di `expire.go`):**
  - Scheduler (legacy ATAU gateway) ada dan tidak disabled → `is_installed=true, is_enabled=true, status="ok"`.
  - Scheduler ada tapi disabled → `is_installed=true, is_enabled=false, status="not ready"`.
  - Scheduler tidak ada → `is_installed=false, is_enabled=false, status="not ready"`.
- **RouterOS:** `/system/scheduler/print` → `mikrotik.ParseSystemSchedulers` → cocokkan nama legacy & gateway (legacy diutamakan bila keduanya ada).

### 2.4 Enable / Disable / Remove — ✅ diekspos

- `HotspotService/DisableExpireMonitor` — `{device_id}` → cari scheduler terpasang lalu `/system/scheduler/set =.id=<id> =disabled=yes` (tanpa menghapus). Error bila belum terpasang.
- `HotspotService/RemoveExpireMonitor` — `{device_id}` → `/system/scheduler/remove =.id=<id>` (script dibiarkan). Error bila belum terpasang.

## 3. Tipe Data (proto — sudah ada)

```protobuf
// api/proto/v1/hotspot.proto — Modul 09
message GetExpireMonitorStatusRequest { string device_id = 1; }
message ExpireMonitorStatusResponse {
  bool is_installed = 1;
  bool is_enabled = 2;
  string status = 3; // "ok" | "not ready"
  string scheduler_name = 4;
}
message SetupExpireMonitorRequest { string device_id = 1; string interval = 2; } // "00:01:00"
message DisableExpireMonitorRequest { string device_id = 1; }
message RemoveExpireMonitorRequest { string device_id = 1; }
```

Tipe port: `port.ExpireMonitorStatus` — `{IsInstalled, IsEnabled, SchedulerID, SchedulerName}` (`internal/port/hotspot_expire.go`).

## 4. Logika Khusus

1. **Idempotensi install:** cek dulu via `GetExpireMonitorStatus`; `set`/re-arm bila ada (legacy atau gateway), `add` dua-langkah hanya bila kosong.
2. **Validasi `interval`:** wajib format durasi RouterOS (`HH:MM:SS`, regex di handler); default **`00:01:00`** (legacy).
3. **`status` field** dipertahankan `"ok"`/`"not ready"` agar kompatibel dengan logika frontend legacy yang menampilkan indikator hijau/oranye.
4. **Kompatibilitas dua bentuk:** `ExpireMonitorSchedulerNames = ["Mikhmon-Expire-Monitor", "mikhmon-expire-scheduler"]` — legacy diutamakan bila keduanya terpasang.
5. **Scheduler ini berjalan di MikroTik** — endpoint Go cukup memantau keberadaan + enabled state (sesuai legacy); logika expire dijalankan oleh script di router.
