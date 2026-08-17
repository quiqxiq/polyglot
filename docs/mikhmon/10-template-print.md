# Modul 10 — Voucher Template & Print Engine

> Kembali ke [README](README.md) · Kode asli: `post/post_template.php` (`do=saveTemplate`), `view/admin/editor.php`, `view/print_voucher.php`, `template/{header,row,footer}.{default,small,thermal}.txt`.
>
> **Status implementasi di Polyglot: ✅ selesai (Fase 7)** — `ListTemplates`, `GetTemplateSection` (read-only) dan `RenderVouchers` (single/batch/preview) sudah diekspos via `HotspotService`. Engine render sudah ada di `pkg/voucher` (dengan QR mode **login URL**); metadata header di-scope-down (identity router + server + on-login profile) karena Fase 0 (proto `Device`) tidak dikerjakan. `SaveTemplateSection` = fase lanjutan (keputusan #5: read-only dulu).

## 1. Pemetaan Legacy

| Fungsi legacy | Request asli | Aksi | Format Response Asli |
| :--- | :--- | :--- | :--- |
| Simpan template | `do=saveTemplate`, `_template` (isi), `file_` (path template) | menulis `template/<file_>` | `{"message":"Saved"}` / pesan error |
| Editor admin | `?admin/editor` | baca + tulis template | HTML |
| Print voucher | `?<m_user>/print` dan `?admin/vpreview` dengan query `c` (comment batch), `uid` (id user), `prev` (preview), `d`/`s`/`t` (template default/small/thermal) | render HTML → `window.print()` | HTML lengkap |

**Alur render legacy:**
1. Ambil data user: by `.id` (single) atau by comment + `?uptime=0s` (batch).
2. Baca metadata harga/validity dari **on-login profile** (`explode(",", $ponlogin)` → index 2 = price, 3 = validity, 4 = selling price).
3. Render `template/{header,row,footer}.<variant>.txt` dengan placeholder.
4. Preview (`prev`) memakai data dummy (`mikhmon` / `1234`); hasil akhir HTML lengkap yang memicu `window.print()`.

## 2. Kondisi Aktual di Polyglot

### 2.1 File template — ✅ sudah ada (embedded)

`internal/template/` berisi 9 file + `embed.go`:

```text
internal/template/
├── header.default.txt  header.small.txt  header.thermal.txt
├── row.default.txt     row.small.txt     row.thermal.txt
├── footer.default.txt  footer.small.txt  footer.thermal.txt
└── embed.go            // go:embed header.*.txt row.*.txt footer.*.txt
```

Usecase hotspot sudah menerima `TemplateDir` (`hotspotUC.New("internal/template", hotGateway)`) namun **belum ada logika render** yang memakainya.

### 2.2 Editor & render — ✅ diekspos (`template_handler.go` + `pkg/voucher`)

Prosedur ConnectRPC sudah ada:
- `HotspotService/ListTemplates` — `{...}` → `{templates: [{name: "default|small|thermal", sections: ["header","row","footer"]}]}`.
- `HotspotService/GetTemplateSection` — `{template_name, section}` → `{content}` (baca dari embedded FS `internal/template.FS`).
- `HotspotService/RenderVouchers` — `{device_id, template_name, comment|user_id, preview}` → `{html, total_vouchers}` (QR mode login URL).
- `SaveTemplateSection` **tidak** diimplementasikan (read-only, keputusan #5).

## 3. Rencana Mapping ke Polyglot (ConnectRPC — usulan)

Prosedur dipanggil `POST /polyglot.v1.HotspotService/<Procedure>` (protected: JWT + RBAC). Template bisa di-embed (read-only via `embed.go`) atau disimpan di DB bila ingin edit di runtime.

### 3.1 List Templates / Get Section — ✅ diekspos

| Operasi | Prosedur | Keterangan |
| :--- | :--- | :--- |
| List | `HotspotService/ListTemplates` | `{...}` → `{templates: [{name: "default|small|thermal", sections: ["header","row","footer"]}]}` (`voucher.ListTemplates`) |
| Get section | `HotspotService/GetTemplateSection` | `{template_name, section}` → `{content}` (baca dari `internal/template.FS` embed) |
| Save section | — | **fase lanjutan** (read-only dulu — keputusan #5) |

### 3.2 Render Voucher (HTML) — ✅ diekspos

- **Prosedur:** `HotspotService/RenderVouchers` — `RenderVouchersRequest{device_id, template_name, comment, user_id, preview}` → `RenderVouchersResponse{html, total_vouchers}`
- **Aturan:** `user_id` dan `comment` saling eksklusif (single vs batch); bila keduanya kosong dan `preview=false` → `CodeInvalidArgument`. `preview=true` memakai data dummy (`mikhmon`/`1234`) tanpa koneksi router.
- **Alur render (`template_handler.go`):**
  1. Ambil user: by `.id` (via `UseCase.GetUser`) atau by comment + `uptime=0s` (via `ListUsers` + filter).
  2. Metadata scope-down (keputusan #4): hotspot name = identity router; dns name = `address`/`name` server hotspot pertama; harga/validity dari on-login profile (`ParseOnLoginScript` — modul 05).
  3. Render `header` + `row`×N + `footer` via `pkg/voucher.RenderWithOptions(..., QRMode=QRModeLoginURL)` (strings.ReplaceAll — tanpa template engine yang bisa eksekusi code).
  4. Response `html` siap `window.print()` (Content-Type `text/html; charset=utf-8` di sisi client).

## 4. Logika Khusus — Placeholder & QR

### 4.1 Placeholder (harus identik di Go)

| Placeholder | Diganti dengan |
| :--- | :--- |
| `%username%` | Nama user |
| `%password%` | Password user |
| `%profile%` | Nama profile |
| `%limitBytesTotal%` | Data limit (formatBytes 2 desimal) |
| `%limitUptime%` | Time limit (contoh `1h`) |
| `%validity%` | Validity dari on-login profile |
| `%price%` | Selling price (dari on-login, index ke-4) |
| `%comment%` | Comment user |
| `%#%` | Nomor urut (1-based) |
| `%dnsName%` | DNS name router |
| `%hotspotName%` | Hotspot name router |
| `%currency%` | Currency router |
| `%qrCode%` / `%qrCodeRed%` / `%qrCodeGreen%` / `%qrCodeBlue%` | Canvas QR (warna default/merah/hijau/biru) |
| `%phone%` | Phone router |
| `%logo%` | URL logo (dengan cache-buster `?YYYYmmddHHMMSS`) |
| `%timeStamp%` | Timestamp cetak `Y-m-d H:i:s` |

> **Catatan metadata router:** di legacy, `dnsName`/`hotspotName`/`currency`/`phone` berasal dari config router (modul 02 / Fase 0 proto `Device` — **tidak dikerjakan**). Implementasi saat ini memakai **scope-down**: `%hotspotName%` = identity router, `%dnsName%` = `address`/`name` server hotspot pertama, `%price%`/`%validity%` = on-login profile. `%currency%`/`%phone%` belum terisi (placeholder tetap dibiarkan).

### 4.2 QR login URL — ✅ (default di `RenderVouchers`)

```
http://<dns_name>/login?username=<urlencode(user)>&password=<urlencode(pass)>
```

Konten QR dipilih via `pkg/voucher.QRMode` (`QRModeLoginURL` = default prosedur; `QRModeCredentials` = perilaku lama `username\npassword`, masih dipertahankan `Render`).

### 4.3 Struktur template

`header` dicetak sekali, `row` per user, `footer` di akhir. Engine render di Go:
1. Baca 3 section sesuai `template_name`.
2. Header di-render dengan metadata router (`hotspotName`, `dnsName`, `currency`, `phone`, `logo`, `timeStamp`).
3. Untuk setiap user (dari `.id` atau comment+`uptime=0s`), render row + QR.
4. Gabungkan dengan footer → HTML lengkap.

## 5. Catatan Implementasi

1. **Keamanan template:** template adalah input admin; render hanya placeholder terbatas (cukup `strings.ReplaceAll`), jangan gunakan engine template arbitrary yang bisa eksekusi code.
2. **QR:** `github.com/skip2/go-qrcode` sudah dipakai di `pkg/voucher` (`QRContent` + `generateQRBase64`, data URI base64). QR warna (`%qrCodeRed/Green/Blue%`) **tidak** diimplementasikan (keputusan #7 — QR hitam default).
3. **Cache-buster logo:** `%logo%` sebaiknya menyertakan `?<timestamp>` agar browser tidak memakai cache logo lama (belum ada sumber logo router — scope-down).
4. **HTML output:** set `Content-Type: text/html; charset=utf-8` pada respons render; dokumen berisi `<script>window.print()</script>` (atau trigger manual) sesuai perilaku legacy.
5. **Fase lanjutan:** `SaveTemplateSection` (override DB), `%currency%`/`%phone%` (metadata Device), QR warna, `web/src/gen` (TS) regenerate bila frontend memakai prosedur baru.
