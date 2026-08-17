# Modul 07 — Voucher Generator & Batch Engine

> Kembali ke [README](README.md) · Kode asli: `post/post_generate_voucher.php`, `post/post_cache_voucher.php`, `core/generator_functions.php`.
>
> **Status implementasi di Polyglot: ✅ selesai** — algoritma generator (`internal/driver/mikrotik/hotspot/voucher.go`), `GenerateVouchers` (dengan `server`/`time_limit`/`data_limit`/`comment`), dan query batch `GetVoucherBatch` (filter comment + `uptime=0s`, pengganti `post_cache_voucher`) sudah diekspos.

## 1. Pemetaan Legacy

| Fungsi legacy | Request asli | Command RouterOS | Format Response Asli |
| :--- | :--- | :--- | :--- |
| Generate batch | `qty`, `sessname`, `server`, `user` (`vc`/`up`), `userl`, `prefix`, `char`, `profile`, `timelimit`, `datalimit`, `gcomment`, `gencode` | `/ip/hotspot/user/add` × qty → `/ip/hotspot/user/print ?comment=<commt>` | `{"message":"success","data":{"count":N,"comment":<commt>,"profile":...}}` / error |
| Cache/query batch | `sessname`, `user`, `gcomment`, `gencode` | `/ip/hotspot/user/print ?comment=<commt> ?uptime=0s` | `{"message":"success","data":{"count":N,"comment":<commt>}}` |

## 2. Mapping ke Polyglot (ConnectRPC — `HotspotService`)

Prosedur dipanggil `POST /polyglot.v1.HotspotService/<Procedure>` (protected: JWT + RBAC). Proto: `api/proto/v1/hotspot.proto`; handler: `internal/adapter/connect/hotspot/profile_user_handler.go` (GenerateVouchers); usecase: `internal/usecase/hotspot/hotspot_usecase.go`; generator: `internal/driver/mikrotik/hotspot/voucher.go`; comment: `internal/driver/mikrotik/hotspot/comment.go`.

### 2.1 Generate Batch Vouchers — ✅ diekspos

- **Prosedur:** `HotspotService/GenerateVouchers` — `GenerateVouchersRequest{device_id, profile, count, user_type, user_length, prefix, character_set, server, time_limit, data_limit, comment}` → `GenerateVouchersResponse{vouchers[], message}`
- **Handler:** `HotspotConnectHandler.GenerateVouchers` (`voucher_handler.go`) — `UseCase.GenerateVouchers(ctx, driver, params, count)` → `HotspotGateway.GenerateVouchers` → `NewGenerateVoucherBatchCommands` (eksekusi `/ip/hotspot/user/add` per voucher) → kembalikan daftar `{username, password, comment}`.
- **Validasi:** `count <= 0` → default 1 (proto `int32`).
- **Pemetaan lengkap:** `server` → `VoucherGenerateParams.Server`, `time_limit` → `LimitUptime`, `data_limit` → `LimitBytes` (dikonversi ke byte via `ParseDataLimit`), `comment` → `CommentTag` (tag batch). `prefix`/`comment` di-sanitize (spasi dihilangkan) agar tidak memecah format `vc-<code>-<date>-<tag>`.
- **Response:** `GenerateVouchersResponse{vouchers (dengan profile/limit_uptime/limit_bytes/comment), message: "successfully generated N vouchers"}` — detail voucher langsung di response (stateless; tidak disimpan di session PHP seperti legacy).

### 2.2 Query Generated Batch (cache / belum pernah login) — ✅ diekspos

- **Prosedur:** `HotspotService/GetVoucherBatch` — `GetVoucherBatchRequest{device_id, comment}` → `GetVoucherBatchResponse{vouchers[], count}`.
- **Handler:** `HotspotConnectHandler.GetVoucherBatch` (`voucher_handler.go`) — `UseCase.GetUsers(ctx, driver, port.ListUsersFilter{Comment, OnlyUnused: true})` → `HotspotGateway.ListUsers` → `/ip/hotspot/user/print ?comment=<c> ?uptime=0s` — hanya voucher yang **belum pernah login**.
- `comment` kosong → `CodeInvalidArgument`. `ListHotspotUsersRequest` juga diperluas (`comment`, `only_unused`) sehingga `ListUsers` bisa dipakai untuk query yang sama.

## 3. Tipe Data (proto / port)

```protobuf
// api/proto/v1/hotspot.proto
message GenerateVouchersRequest {
  string device_id = 1; string profile = 2; int32 count = 3; string user_type = 4;
  int32 user_length = 5; string prefix = 6; string character_set = 7;
  string server = 8; string time_limit = 9; string data_limit = 10; string comment = 11;
}
message GenerateVouchersResponse { repeated HotspotUser vouchers = 1; string message = 2; }
message GetVoucherBatchRequest  { string device_id = 1; string comment = 2; }
message GetVoucherBatchResponse { repeated HotspotUser vouchers = 1; int32 count = 2; }
```

Tipe port: `port.VoucherGenerateParams`, `port.VoucherBatch`, `port.GeneratedVoucher` (`internal/port/`); charset & generator: `internal/driver/mikrotik/hotspot/voucher.go`.

## 4. Logika Khusus — Algoritma Generator (implementasi aktual)

**Implementasi: `internal/driver/mikrotik/hotspot/voucher.go`** — memakai `crypto/rand` (bukan `str_shuffle` PHP yang tidak crypto-secure). Kompatibel secara perilaku: format comment `vc-<code>-<MM.dd.yy>[-tag]` via `FormatPreLoginComment`, dan konversi data limit via `ParseDataLimit`.

**Charset per tipe (`CharSet`):**

| CharSet (proto `character_set`) | Charset |
| :--- | :--- |
| `numeric` | `1234567890` |
| `lower` | `a-z` |
| `upper` | `A-Z` |
| `lowernum` | `a-z0-9` |
| `uppernum` | `A-Z0-9` |
| `mixed` | `a-zA-Z0-9` |

**Aturan generasi (`NewGenerateVoucherBatchCommands`):**
- `userLen` default 6; `count` default 1.
- `uname = prefix + GenerateVoucherCode(userLen, charSet)`; `pass = uname` bila `passLength == 0` (mode `vc`), selain itu `pass = GenerateVoucherCode(passLength, charSet)` (mode `up`).
- `comment = FormatPreLoginComment("vc", code3, tag, now)` — contoh `vc-A3X-08.17.26-MyTag`.
- `limit-bytes-out` = `ParseDataLimit(p.LimitBytes)` dalam byte bila > 0.

**Konversi data limit (byte):**
```go
// ParseDataLimit: "1000M" -> 1048576000; "1G" -> 1073741824; "500" -> 500; "" -> 0
func ParseDataLimit(s string) int64 { /* lihat voucher.go */ }
```

## 5. Formatter & Helper

Draft awal menaruh formatter di `pkg/utils`. Di Polyglot, format byte/bits untuk UI belum terpusat — `pkg/voucher/` ada untuk generate (lihat `pkg/voucher/generator.go`). Usulan bila diperlukan (modul 03, 04, 07, 10):

```go
// ParseDataLimitToBytes: "1000M" -> 1048576000; "1G" -> 1073741824; "500" -> 500
// FormatBytes: 1048576 -> "1.00 MiB" (desimal 2 utk print voucher)
// FormatBits: 12503200 -> "12.50 Mbps" (dipakai untuk traffic & UI)
```
*(Implementasi identik dengan versi di draft awal — salin ke `pkg/utils/` atau tempatkan di driver bila hanya dipakai internal.)*

## 6. Catatan Implementasi

1. **Batch besar:** qty besar memblokir socket RouterOS cukup lama (legacy `set_time_limit(0)`). `GenerateVouchers` gateway mengeksekusi command secara sekuensial; untuk qty besar pertimbangkan async (job queue) + status endpoint.
2. **Duplikat:** generator bisa menghasilkan username duplikat saat qty besar; RouterOS akan `!trap` "duplicate name" — propagasikan pesan asli (legacy menyerahkan pada `!trap`).
3. **Keamanan:** `prefix`/`comment`/`tag` di-sanitize di handler `voucher_handler.go` (`sanitizeBatchTag`: spasi dihilangkan) agar tidak memecah format batch `vc-<code>-<date>-<tag>`.
