# Modul 04 — Hotspot User Management

> Kembali ke [README](README.md) · Kode asli: `get/get_users.php`, `get/get_user.php`, `post/post_add_user.php`, `post/post_update_user.php`, `post/post_hotspot_remove.php` (`where=user_`), `get/get_tot_users.php`.
>
> **Status implementasi di Polyglot: ✅ selesai** — gateway (`port.HotspotGateway`) + usecase + seluruh prosedur ConnectRPC (List/Get/Create/Update/Reset/Delete) sudah diekspos via `HotspotService`. Logika comment `vc-`/`up-` (create) dan rebuild `expdate`/`ucode` (update) diimplementasikan di `comment.go` (`BuildCreateUserComment`, `BuildUpdatedComment`).

## 1. Pemetaan Legacy

| Fungsi legacy | Request asli | Command RouterOS | Format Response Asli |
| :--- | :--- | :--- | :--- |
| List users (`users`) | `prof` (profile), `f` (force), `c` (cache/count) | `/ip/hotspot/user/print ?profile=<prof>` | jsEncode25 array user; jika `c` ada → **angka** (count) |
| Get user (`user`) | `id` (`.id`) **atau** `name` | `/ip/hotspot/user/print ?.id=<id>` / `?name=<name>` | jsEncode25 array |
| Total users (`get_tot_users`) | – | `/ip/hotspot/user/print =count-only` | **JSON polos** `{"users": count-1}` |
| Add user | `name` + `sessname`, `server`, `password`, `profile`, `macaddr`, `timelimit`, `datalimit`, `comment` | `/ip/hotspot/user/add` (comment otomatis `vc-`/`up-`) | `{"message":"success","data":<user>}` / `{"message":"error","data":{"error":<trap>}}` |
| Update user | `uid`, `reset` (`yes`/`no`), + semua field di atas, `expdate`, `ucode` | `/ip/hotspot/user/reset-counters` (jika reset=yes) → `/ip/hotspot/user/set` → `/ip/hotspot/user/print ?.id` | `{"message":"success","data":<user>}` / error |
| Remove user | `sessname`, `where=user_`, `id` | `/ip/hotspot/user/remove =.id=<id>` | `{"message":"success"}` / `{"message":"error"}` |

## 2. Mapping ke Polyglot (ConnectRPC — `HotspotService`)

Prosedur dipanggil `POST /polyglot.v1.HotspotService/<Procedure>` (protected: JWT + RBAC). Proto: `api/proto/v1/hotspot.proto`; handler: `internal/adapter/connect/hotspot/profile_user_handler.go` (ListUsers, GenerateVouchers); usecase: `internal/usecase/hotspot/hotspot_usecase.go`; gateway: `internal/driver/mikrotik/hotspot/gateway.go`; command user: `internal/driver/mikrotik/hotspot_user.go`; parser comment: `internal/driver/mikrotik/hotspot/comment.go`.

### 2.1 List Users — ✅ diekspos

- **Prosedur:** `HotspotService/ListUsers` — `ListHotspotUsersRequest{device_id, profile, comment, only_unused}` → `ListHotspotUsersResponse{users[]}`
- **Handler:** `HotspotConnectHandler.ListUsers` (`profile_user_handler.go`) — `UseCase.GetUsers(ctx, driver, port.ListUsersFilter{Profile, Comment, OnlyUnused})` → `ToProtoHotspotUsers`.
- **Filter `comment`** (batch tag) & **`only_unused`** (`uptime=0s`, dipakai modul 07) sudah dipetakan ke `?comment=`/`?uptime=0s` di `gateway.ListUsers` (fix: filter `profile` kini benar memakai `?profile=`).
- **Message `HotspotUser`:** `{id, name, password, profile, limit_uptime, limit_bytes, uptime, bytes_in, bytes_out, comment, disabled, server}`.
- **Count-only** (pengganti `get_tot_users`): `len(users)` dari hasil print; legacy mengurangi 1 (user default `default-trial`).

### 2.2 Get Single User — ✅ diekspos

- **Prosedur:** `HotspotService/GetUser` — `GetHotspotUserRequest{device_id, ros_id}` → `GetHotspotUserResponse{user}`.
- **Handler:** `HotspotConnectHandler.GetUser` (`user_handler.go`) → `UseCase.GetUser` → `/ip/hotspot/user/print ?.id=<id>` (error bila kosong). `ros_id` kosong → `CodeInvalidArgument`.

### 2.3 Create User — ✅ diekspos

- **Prosedur:** `HotspotService/CreateUser` — `CreateHotspotUserRequest{device_id, server, name, password, profile, mac_address, time_limit, data_limit, comment}` → `CreateHotspotUserResponse{user, message}`.
- **Handler:** `HotspotConnectHandler.CreateUser` (`user_handler.go`) → `UseCase.AddUser` → `mikrotik.NewAddHotspotUserCommand` (`/ip/hotspot/user/add`).
- **Logika comment:** `comment` kosong → `name == password` ? prefix `vc-` : `up-`, dibangun via `hotspot.BuildCreateUserComment` (reuse `FormatPreLoginComment`); data limit `"1000M"` → byte via `ParseDataLimit` (isi `limit-bytes-out`). Response berisi user dari print ulang (`?name=<name>`).
- **Error:** duplikat name → `!trap` RouterOS → `CodeInternal` dengan pesan asli.

### 2.4 Update User — ✅ diekspos

- **Prosedur:** `HotspotService/UpdateUser` — `UpdateHotspotUserRequest{device_id, ros_id, reset_counter, server, name, password, profile, mac_address, time_limit, data_limit, comment, expire_date, user_code}` → `UpdateHotspotUserResponse{user, message}`.
- **Handler:** `HotspotConnectHandler.UpdateUser` (`user_handler.go`) — `reset_counter=true` → `UseCase.ResetUserCounters` dulu, lalu `UseCase.UpdateUser` → `mikrotik.NewSetHotspotUserCommand` (kini mendukung `name` & `server`), lalu print ulang (`GetUser`) untuk response.
- **Logika comment (legacy `post_update_user`) via `hotspot.BuildUpdatedComment`:**
  - `expire_date == "" && user_code == ""` → comment apa adanya.
  - `expire_date == "" && user_code != ""` → prefix lama (`vc-`/`up-`) atau `X-` dipertahankan; comment dibangun ulang.
  - `expire_date != "" && user_code == ""` → `comment = "<expire_date> <comment>"` (menyimpan tanggal expire).

### 2.5 Reset Counter — ✅ diekspos

- **Prosedur:** `HotspotService/ResetUserCounters` — `ResetHotspotUserCountersRequest{device_id, ros_id}` → `ResetHotspotUserCountersResponse{message}`.
- **Handler:** `HotspotConnectHandler.ResetUserCounters` (`user_handler.go`) → `UseCase.ResetUserCounters` → `/ip/hotspot/user/reset-counters =.id=<id>`.

### 2.6 Delete User — ✅ diekspos

- **Prosedur:** `HotspotService/DeleteUser` — `DeleteHotspotUserRequest{device_id, ros_id}` → `DeleteHotspotUserResponse{message}`.
- **Handler:** `HotspotConnectHandler.DeleteUser` (`user_handler.go`) → `UseCase.RemoveUser` → `/ip/hotspot/user/remove =.id=<id>`.

## 3. Tipe Data (proto / port)

```protobuf
// api/proto/v1/hotspot.proto
message HotspotUser {
  string id = 1; string name = 2; string password = 3; string profile = 4;
  string limit_uptime = 5; string limit_bytes = 6; string uptime = 7;
  string bytes_in = 8; string bytes_out = 9; string comment = 10; bool disabled = 11;
}
message ListHotspotUsersRequest  { string device_id = 1; string profile = 2; }
message ListHotspotUsersResponse { repeated HotspotUser users = 1; }
```

Tipe port: `port.HotspotUser`, `port.HotspotUserParams`, `port.MikhmonComment` (`internal/port/`); parser: `mikrotik.ParseHotspotUsers` (`internal/driver/mikrotik/hotspot_user.go`); mapper: `ToProtoHotspotUsers` (`internal/adapter/connect/hotspot/mapper.go`).

## 4. Logika Khusus

1. **Comment builder** (`vc-`/`up-`) — dipakai juga oleh generator voucher (modul 07). Implementasi: `FormatPreLoginComment` + `ParseMikhmonComment` (`internal/driver/mikrotik/hotspot/comment.go`), termasuk format post-login `"DD/MM/YYYY HH:MM:SS <mode> <old-comment>"` dan `IsExpired`.
2. **`valid_until` parser** — comment user di-update oleh on-login script menjadi `<date> <time> N/X`; ekstrak via `ParseMikhmonComment` (`ExpireDate`/`ExpireTime`).
3. **Nilai byte:** `limit-bytes-total` dikirim sebagai byte (konversi `"1000M"` → `1048576000` via `ParseDataLimit`, modul 07).
4. **Idempotensi:** cek duplikat `name` sebelum add (RouterOS akan `!trap` "duplicate name"); propagasikan pesan asli.
5. **Policy gate:** semua command (termasuk remove/reset yang destruktif) dieksekusi lewat `network.ExecuteCommand` — destructive command butuh approval.
