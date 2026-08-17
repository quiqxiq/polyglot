# Modul 05 — Hotspot User Profile Management

> Kembali ke [README](README.md) · Kode asli: `get/get_profiles.php`, `get/get_profile.php`, `post/post_add_userprofile.php`, `post/post_update_userprofile.php`, `post/post_hotspot_remove.php` (`where=profile_`).
>
> **Status implementasi di Polyglot: ✅ selesai** — CRUD profile lengkap (`ListProfiles`, `CreateProfile`, `UpdateProfile`, `DeleteProfile`) diekspos via `HotspotService`, plus parser on-login terstruktur (`ParseOnLoginScript` di `profile.go`) dan normalisasi nama (`NormalizeProfileName`).

## 1. Pemetaan Legacy

| Fungsi legacy | Request asli | Command RouterOS | Format Response Asli |
| :--- | :--- | :--- | :--- |
| List profiles (`profiles`) | `f` | `/ip/hotspot/user/profile/print` | jsEncode25 array |
| Get profile (`profile`) | `id` **atau** `name` | `/ip/hotspot/user/profile/print ?.id` / `?name` | **JSON polos** |
| Add profile | `name`, `addresspool`, `sharedusers`, `ratelimit`, `parentqueue`, `expmode`, `validity`, `price`, `sellingprice`, `lockuser`, `lockserver` | `/ip/hotspot/user/profile/add` (dengan on-login script) | `{"message":"success","data":<profile>}` / error |
| Update profile | + `upid` (`.id`) | `/ip/hotspot/user/profile/set` → print | `{"message":"success","data":<profile>}` / error |
| Remove profile | `sessname`, `where=profile_`, `id` | `/ip/hotspot/user/profile/remove =.id=<id>` | `{"message":"success"}` / `{"message":"error"}` |

## 2. Mapping ke Polyglot (ConnectRPC — `HotspotService`)

Prosedur dipanggil `POST /polyglot.v1.HotspotService/<Procedure>` (protected: JWT + RBAC). Proto: `api/proto/v1/hotspot.proto`; handler: `internal/adapter/connect/hotspot/profile_user_handler.go` (List) + `profile_handler.go` (Create/Update/Delete); mapper: `mapper_profile.go`; usecase: `internal/usecase/hotspot/hotspot_usecase.go`; gateway: `internal/driver/mikrotik/hotspot/gateway.go`; on-login builder/parser: `internal/driver/mikrotik/hotspot/profile.go`.

### 2.1 List Profiles — ✅ diekspos

- **Prosedur:** `HotspotService/ListProfiles` — `ListHotspotProfilesRequest{device_id}` → `ListHotspotProfilesResponse{profiles[]}`
- **Handler:** `HotspotConnectHandler.ListProfiles` — `UseCase.GetProfiles` → `ToProtoHotspotProfiles`.
- **Message `HotspotProfile`:** `{id, name, shared_users, rate_limit, mode_expire, validity, price, selling_price, lock_user, lock_server, parent_queue, address_pool, comment}` — `mode_expire`/`validity`/`price`/`selling_price`/`lock_user`/`lock_server` kini **terstruktur** via `hotspot.ParseOnLoginScript(p.OnLogin)` (bukan raw script).

### 2.2 Get Single Profile — 🔴 belum diekspos

- Logika: `/ip/hotspot/user/profile/print ?.id=<id>` (via `mikrotik.NewPrintHotspotUserProfilesCommand`). Belum ada prosedur khusus; bisa pakai `ListProfiles` + filter client, atau tambah `GetProfile`.

### 2.3 Create / Update Profile — ✅ diekspos

- **Prosedur:** `HotspotService/CreateProfile` — `CreateHotspotProfileRequest{device_id, profile: HotspotProfileParams}` → `CreateHotspotProfileResponse{profile, message}`; `HotspotService/UpdateProfile` — `UpdateHotspotProfileRequest{device_id, ros_id, profile}` → `UpdateHotspotProfileResponse{profile, message}`.
- **Handler:** `HotspotConnectHandler.CreateProfile`/`UpdateProfile` (`profile_handler.go`) → `ProfileParamsFromProto` (mapper) → `UseCase.CreateProfile`/`UpdateProfile` → `NewAddMikhmonProfileCommand`/`NewSetMikhmonProfileCommand` (`profile.go`) → `mikrotik.NewAdd/SetHotspotUserProfileCommand` dengan `on-login = BuildOnLoginScript(p)`.
- **Create** menormalisasi nama via `hotspot.NormalizeProfileName` (spasi → `-`) lalu re-print & kembalikan profile yang dibuat; **Update** re-print berdasar `ros_id`.
- Parameter: `HotspotProfileParams` proto → `port.MikhmonProfileParams` — `{name, address_pool, shared_users, rate_limit, parent_queue, comment, expire_mode, validity, price, selling_price, lock_user, lock_server, enable_recording}`.

### 2.4 Delete Profile — ✅ diekspos

- **Prosedur:** `HotspotService/DeleteProfile` — `DeleteHotspotProfileRequest{device_id, ros_id}` → `DeleteHotspotProfileResponse{message}`.
- **Handler:** `HotspotConnectHandler.DeleteProfile` (`profile_handler.go`) → `UseCase.DeleteProfile` → `mikrotik.NewRemoveHotspotUserProfileCommand`. RouterOS menolak menghapus profile yang masih dipakai user (`!trap`) — pesan asli dipropagasikan.

## 3. Tipe Data (proto / port)

```protobuf
// api/proto/v1/hotspot.proto
message HotspotProfile {
  string id = 1; string name = 2; string shared_users = 3; string rate_limit = 4;
  string mode_expire = 5; string validity = 6; double price = 7; double selling_price = 8;
  string lock_user = 9; string parent_queue = 10; string comment = 11;
  string address_pool = 12; string lock_server = 13;
}
message ListHotspotProfilesRequest  { string device_id = 1; }
message ListHotspotProfilesResponse { repeated HotspotProfile profiles = 1; }
message HotspotProfileParams { /* lihat hotspot.proto — name, address_pool, shared_users, rate_limit, parent_queue, price, selling_price, validity, expire_mode, lock_user, lock_server, enable_recording, comment */ }
```

Tipe port: `port.MikhmonProfileParams` (`internal/port/`). Expire mode: `port.ExpireMode` — `ntf`/`ntfc` (Notice), `rem`/`remc` (Remove), `0` (none) — konstanta di `internal/driver/mikrotik/hotspot/profile.go`.

## 4. Logika Khusus — On-Login Script (WAJIB 1:1 dengan legacy)

Mikhmon menyematkan metadata bisnis ke properti `on-login` User Profile. **Implementasi aktual: `BuildOnLoginScript(p MikhmonProfileParams) string` di `internal/driver/mikrotik/hotspot/profile.go`** — port langsung dari Mikhmon v4. Format header metadata (string di dalam `:put`):

```
:put (\",<expmode>,<price>,<validity>,<sellingprice>,,<lockuser>,<lockserver>,\");
```

Contoh: `:put (\",remc,3000,1h,3500,,Enable,Disable,\");`

### 4.1 Struktur script yang dihasilkan

- **Record transaksi** (`enable_recording` / mode `*c`): `; :local mac $mac-address; ... /system script add name="$date-|-$time-|-$user-|-<price>-|-$address-|-$mac-|-<validity>-|-<profile>-|-$comment" owner="$month$year" source=$date comment=mikhmon` — dipakai modul 08.
- **Lock user** (`lock_user`): set `mac-address` user dari `$mac-address` saat login.
- **Lock server** (`lock_server`): set `server` user dari hotspot host.
- **Scheduling expire** (mode bukan `0`): buat scheduler per-user → ambil `next-run` → tulis comment `"<date> <time> N/X"` → hapus scheduler. Mode `N` (ntf/ntfc) vs `X` (rem/remc).

**Tabel mode expire:**

| Mode | Arti | Perilaku update profile |
| :--- | :--- | :--- |
| `0` | Tanpa expire (harga > 0 → `noexp`) | – |
| `rem` | Remove setelah expire | `remove` |
| `remc` | Remove & Record (tulis laporan) | `remove` |
| `ntf` | Notice (set `limit-uptime=1s`) | `set limit-uptime=1s` |
| `ntfc` | Notice & Record | `set limit-uptime=1s` |

### 4.2 Parser on-login (membaca profil) — ✅ diimplementasikan

Parser terstruktur sudah ada di `internal/driver/mikrotik/hotspot/profile.go` dan dipakai `mapper_profile.go`:

```go
// di internal/driver/mikrotik/hotspot/profile.go
type ProfileMeta struct {
    ExpireMode   string  // "0","rem","ntf","remc","ntfc"
    Price        float64 // index [2] — selling price
    Validity     string  // index [3]
    SellingPrice float64 // index [4] — cost price
    LockUser     string  // index [6] — "Enable"/"Disable"
    LockServer   string  // index [7] — "Enable"/"Disable"
}
// ParseOnLoginScript(onLogin string) (ProfileMeta, error) — parse :put (",...")
// Mengenali dua layout: standar (mode di index 1) dan noexp (marker "noexp" di index 5).
```

### 4.3 Normalisasi nama profile

Legacy menghapus spasi pada nama profile saat create (`preg_replace('/\s+/', '-', name)`) — **karena nama profile ikut disisipkan ke nama script laporan yang dipisah `-|-`** (modul 08). Aturan ini diimplementasikan di `hotspot.NormalizeProfileName` dan dipakai `CreateProfile` (handler `profile_handler.go`).

## 5. Catatan Implementasi

- **Jangan simpan script mentah dari client:** saat create/update, selalu regenerasi on-login dari `MikhmonProfileParams` terstruktur (`BuildOnLoginScript`) — mencegah injeksi script RouterOS.
- `MikhmonProfileParams` sudah dipakai oleh `GenerateVouchers`/`CreateUserProfile` di gateway — reuse, jangan buat tipe baru.
