# Modul 01 — Auth & Admin Management

> Kembali ke [README](README.md) · Kode asli: `view/login.php`, `post/post_logout.php`, `post/post_a_router.php` (`do=saveAdmin`), `config/config.php` (baris `mikhmon`).
>
> **Status implementasi di Polyglot: ✅ selesai** — `AuthService` (Login/GetMe/RefreshToken/Logout) + `UserService` (multi-user + RBAC Casbin) sudah aktif.

## 1. Pemetaan Legacy

| Fungsi legacy | File | Request asli | Response asli |
| :--- | :--- | :--- | :--- |
| Login | `view/login.php` (route `?admin/login`) | POST `user`, `pass`, `login` | redirect `?admin/settings`; validasi `user === $useradm && pass === dec_rypt($passadm)` |
| Logout | `post/post_logout.php` | POST `logout` | `session_destroy()`; echo nilai `logout` |
| Ganti credential | `post/post_a_router.php` `do=saveAdmin` | POST `username`(encode), `password`(encode) | `{"message":"Success"}` |
| Sesi | `index.php` | – | cek `$_SESSION["mikhmon"]`, redirect ke `?admin/login` bila kosong |

**Credential default legacy:** username `mikhmon`, password `mikhmon` (password admin disimpan `enc_rypt()` di baris `mikhmon>|>...`). Di Polyglot default admin dibuat lewat migrasi/seeder — tidak ada file config.

## 2. Mapping ke Polyglot (ConnectRPC)

Prosedur dipanggil `POST /polyglot.v1.AuthService/<Procedure>` (JSON codec). Definisi proto: `api/proto/v1/auth.proto` & `api/proto/v1/users.proto`; handler: `internal/adapter/connect/auth/auth_handler.go` + `user_handler.go`; usecase: `internal/usecase/auth/` (login.go, refresh_token.go) & `internal/usecase/user/manage_user.go`.

### 2.1 Login

- **Prosedur:** `AuthService/Login` — `LoginRequest{username, password}`
- **Handler:** `AuthConnectHandler.Login` (`internal/adapter/connect/auth/auth_handler.go`)
- **Usecase:** `AuthUseCase.Login` — validasi kredensial (hash bcrypt), rate-limit login (`ErrTooManyAttempts` → `CodeResourceExhausted`), cek status akun (`ErrAccountInactive` → `CodePermissionDenied`).
- **Response `LoginResponse`:** `{token, user{id, username, email, role, roles, permissions}, expires_at_unix}` + **refresh token diset sebagai cookie httpOnly** (`polyglot_refresh`) — bukan body.
- **Error:** `ErrInvalidCredentials` → `CodeUnauthenticated`.

```json
// POST /polyglot.v1.AuthService/Login
{ "username": "admin", "password": "secret" }
// 200
{ "token": "eyJhbGciOiJIUzI1NiIs...", "expiresAtUnix": 1787000000,
  "user": { "id": "1", "username": "admin", "role": "admin", "permissions": ["user:read", ...] } }
// Set-Cookie: polyglot_refresh=...; HttpOnly; Path=/; Max-Age=...
```

### 2.2 Get Current Admin / Me

- **Prosedur:** `AuthService/GetMe` — `GetMeRequest{}`; token diambil dari header `Authorization: Bearer <token>`.
- **Handler:** `AuthConnectHandler.GetMe` — validasi token → load user + roles + permissions (Casbin) → `GetMeResponse{user}`.

### 2.3 Refresh Token

- **Prosedur:** `AuthService/RefreshToken` — `RefreshTokenRequest{refresh_token}`; token refresh diambil dari cookie `polyglot_refresh` bila field kosong (fallback untuk client non-browser).
- **Handler:** `AuthConnectHandler.RefreshToken` — `RefreshTokenUseCase.Refresh` (rotasi token; `ErrInvalidRefreshToken` → `CodeUnauthenticated`).
- **Response `RefreshTokenResponse`:** `{token, expires_at_unix, refresh_expires_at_unix}` + set ulang cookie refresh.

### 2.4 Logout

- **Prosedur:** `AuthService/Logout` — `LogoutRequest{}`; revoke refresh token (best-effort) + `ClearRefreshTokenCookie`.
- **Response `LogoutResponse`:** `{success: true}`.

### 2.5 Ganti credential / user management (perluasan legacy `saveAdmin`)

Legacy hanya punya satu admin. Polyglot sudah multi-user dengan RBAC:

| Prosedur | Service | Keterangan |
| :-- | :-- | :-- |
| `ListUsers` | `UserService` | Daftar user admin |
| `CreateUser` | `UserService` | Tambah user + role |
| `UpdateUser` | `UserService` | Update profil/role |
| `ResetPassword` | `UserService` | Reset password |
| `ToggleActive` | `UserService` | Aktif/nonaktif akun |
| `DeleteUser` | `UserService` | Hapus user |
| RBAC | `RBACService` | Kelola role/permission (Casbin) |

Handler: `internal/adapter/connect/auth/user_handler.go`, `rbac_handler.go`; enforcer: `internal/adapter/auth/casbin.go`; permission mapping per prosedur: `internal/adapter/auth/procedure_permissions.go`.

## 3. Tipe Data (proto / domain)

```protobuf
// api/proto/v1/auth.proto
message LoginRequest  { string username = 1; string password = 2; }
message LoginResponse { string token = 1; UserProfile user = 2; int64 expires_at_unix = 3; }
message UserProfile  {
  string id = 1; string username = 2; string email = 3; string role = 4;
  repeated string roles = 5;
  repeated string permissions = 6; // flatten "resource:action" dari Casbin
}
message RefreshTokenRequest  { string refresh_token = 1; }
message RefreshTokenResponse { string token = 1; int64 expires_at_unix = 2; int64 refresh_expires_at_unix = 3; }
message LogoutRequest  {}
message LogoutResponse { bool success = 1; }
```

Domain user: `internal/domain/customer/user.go` (atau model repository `internal/adapter/postgres/user_repository.go`).

## 4. Logika Khusus

1. **Hash modern:** password disimpan hash (bcrypt) di tabel `users` — bukan `enc_rypt()` legacy. Bila perlu migrasi dari Mikhmon lama, verifikasi `dec_rypt()` dulu lalu upgrade hash saat login sukses (migrasi bertahap — belum diimplementasikan).
2. **JWT claims:** `sub` (user id), `exp`, dan role; middleware `AuthenticateJWT` + `AuthorizeProcedure` (Casbin) memeriksa permission per prosedur (`internal/adapter/auth/procedure_permissions.go`).
3. **Brute force:** sudah ada rate-limit login (`ErrTooManyAttempts`) + audit trail.
4. **Refresh token** disimpan di Redis (TTL konfigurable) dan dikirim via cookie httpOnly; `secure` flag otomatis di environment `production`.
