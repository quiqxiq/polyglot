# PLAN: Standardisasi Backend Golang — Polyglot

## Temuan Kunci yang Mendasari Plan

| # | Masalah | Bukti |
|---|---|---|
| 1 | 🔴 .env berisi secret ter-track git | `git ls-files .env` |
| 2 | 🔴 Arsitektur error terbalik: `pkg/response` mengimpor `internal/usecase` (5 paket) — `pkg` seharusnya generik. Setiap error baru memaksa edit `pkg/response` (tidak scalable) | `pkg/response/errors.go:8-14` |
| 3 | 🔴 Sentinel error tersebar di usecase (`login.go`, `manage_user.go`, `billing/errors.go`), padahal guidelines §6 mewajibkan di `domain/<x>/errors.go`. Hanya 2 dari 16 domain punya `errors.go` | audit direktori domain |
| 4 | 🟠 259 pemakaian `connect.NewError` langsung di handler, bypass `MapDomainError` (122 di antaranya validasi `InvalidArgument` manual) | `grep adapter/connect` |
| 5 | 🟠 Pesan error campur bahasa (EN/ID): `"file kosong atau hanya header"` vs `"device: not found"`; gaya prefix campur: `"device: not found"` vs `ErrNotFoundBilling` | sampel `errors.New` |
| 6 | 🟠 Boundary: `usecase/metrics` → `driver/mikrotik/system` | `ping_stream_manager.go:12` |
| 7 | 🟡 Suffix tipe usecase campur: `*UseCase` (20), `*Service` (2), `*Manager` (2), `Engine`, `Guardrail`; `WaSenderWorker` (akronim salah, harusnya WA) | `grep type` |
| 8 | 🟡 File usecase melanggar pola `<verb>_<noun>.go`: `hotspot_usecase.go`, `chat_service.go`, `portal.go` | listing usecase |
| 9 | 🟡 Mapper campur exported/unexported: `ToProtoDHCPLease` vs `toProtoInvoice` tanpa alasan konsisten | `grep mapper` |
| 10 | 🟡 Port duplikat/ambigu: `AuditWriter` vs `AuditLogWriter`; `PaymentGateway`/`PaymentProcessor`/`PaymentReader`; ~40+ struct data tinggal di `port/` (mis. `PPPProfile`, `VoucherBatch`, `SystemResource`) yang secara konsep milik `domain/` | listing port |
| 11 | 🟡 Transisi mikrotik setengah jalan: file root (`dhcp.go`, `ppp.go`, dst) duplikat dengan subpaket (`dhcp/`, `ppp/`) | listing driver |
| 12 | 🟡 Dokumentasi (`AGENTS.md` §1.1) tertinggal dari struktur aktual | perbandingan |

> **Keputusan Anda yang sudah dikunci:** error message Inggris, protovalidate, hybrid semantik (*UseCase/*Worker), refactor menyeluruh bertahap, wire/proto tidak berubah.

---

## FASE 0 — Keamanan & Hygiene (prasyarat, ~30 menit)

1. `git rm --cached .env` + commit; verifikasi `*.env` di `.gitignore` efektif.
2. Rotasi semua kredensial yang pernah ter-commit (manual oleh Anda).
3. Hapus artefak: `.understand-anything/.trash-*`.
4. Baseline hijau: `go build ./... && go test ./... && golangci-lint run` — catat hasil sebagai acuan setiap fase.

---

## FASE 1 — Arsitektur Error Baru (fondasi semua fase lain)

Desain `pkg/fault` — paket generik (nol dependensi internal) berisi error kind:

```go
// pkg/fault/fault.go
type Kind int
const (
    KindUnknown Kind = iota
    KindNotFound; KindInvalidInput; KindAlreadyExists
    KindPermissionDenied; KindUnauthenticated
    KindFailedPrecondition; KindConflict; KindUnavailable; KindResourceExhausted
)
func New(kind Kind, msg string) error      // sentinel constructor
func Wrap(kind Kind, err error) error      // wrap dengan kind
func KindOf(err error) Kind                // errors.As traversal
```

**Lalu:**
1. Buat `errors.go` di seluruh 16 domain (billing, customer, bot, cashbook, plan, subscription, registration, setting, notification, session, llm, command, audit, reporting + rapikan device, skill). Format: `var ErrNotFound = fault.New(fault.KindNotFound, "billing: not found")` — prefix domain, pesan Inggris.
2. Migrasi sentinel dari usecase → domain: `authuc.ErrInvalidCredentials`, `useruc.ErrUserAlreadyExists`, `billing.ErrNotFoundBilling` → `domain/billing.ErrNotFound`, dst. Error murni-orkestrasi (mis. `networkuc.ErrApprovalRequired`) boleh tetap di usecase tapi wajib dibungkus `fault.Wrap`.
3. Tulis ulang `pkg/response/errors.go`: `MapDomainError` cukup switch atas `fault.KindOf(err)` → `connect.Code`. Hapus semua import `internal/*`. Error baru di masa depan tidak perlu menyentuh file ini lagi.
4. Standarkan wrapping: `fmt.Errorf("find user %d: %w", id, err)` — sweep `errors.New`/`fmt.Errorf` tanpa `%w` di jalur propagasi (temuan: ~15 file).
5. Terjemahkan semua pesan error ID → EN (`"file kosong atau hanya header"` → `"importer: file empty or header only"`).
6. Update `DEVELOPMENT-GUIDELINES.md` §6 dengan pola baru.

**Verifikasi:** build + test hijau; `rg 'internal/(usecase|adapter|driver)' pkg/` kosong; test unit `pkg/fault`.

---

## FASE 2 — Validasi Request via protovalidate

1. Tambah `buf.build/bufbuild/protovalidate` ke `buf.gen.yaml`/deps; anotasi `buf.validate.field` di 18 file proto (required, min_len, pattern untuk ID) — hanya anotasi, tidak mengubah field/wire format.
2. Buat interceptor ConnectRPC `internal/adapter/http/middleware/validate.go` (atau connect interceptor) yang menjalankan protovalidate untuk semua unary+stream.
3. Hapus 122 validasi manual `req.Msg.X == ""` di handler — handler jadi tipis: map → usecase → `MapDomainError`.
4. Validasi aturan bisnis tetap di usecase, kembalikan `fault.KindInvalidInput`.
5. Regenerasi kode: `make proto && make proto-web` (TS ikut regenerate, tidak mengubah perilaku FE).

**Verifikasi:** build + test; uji manual satu endpoint dengan field kosong → `CodeInvalidArgument` dari interceptor.

---

## FASE 3 — Boundary & Penempatan

1. `usecase/metrics/ping_stream_manager.go`: definisikan interface/tipe yang dibutuhkan di `internal/port` (atau pindah tipe data ke domain), driver mengimplementasi. Hapus import driver.
2. Pindah `usecase/importer/router_live_e2e_test.go` → `test/integration/` (konsisten dengan test integrasi lain).
3. Test usecase yang import adapter/driver (4 file lain): ganti ke mock `port/mocktest` bila memungkinkan; sisanya beri `// DEVIASI`.
4. Bersihkan sisa referensi gin di `ws/sse_hub_test.go`.

**Verifikasi:** `rg 'internal/(adapter|driver)' internal/usecase --glob '!*e2e*'` hanya menyisakan yang ber-DEVIASI; build + test.

---

## FASE 4 — Penamaan (tipe, file, fungsi) — via gopls rename, per-commit kecil

**Tipe usecase (hybrid semantik):**

| Sekarang | Menjadi |
|---|---|
| `ChatService` | `ChatUseCase` |
| `ConversationService` | `ConversationUseCase` |
| `ContextManager` (bot) | tetap (komponen internal engine) — atau `contextBuilder` jika unexported memungkinkan |
| `PingStreamManager` | `PingStreamWorker` (long-running background) |
| `WaSenderWorker` | `WASenderWorker` |
| `Engine`, `Guardrail` (bot) | tetap — nama domain yang deskriptif, dicatat sebagai pengecualian sah di guidelines |

**File usecase (`<verb>_<noun>.go`):**
- `hotspot/hotspot_usecase.go` → pecah `manage_user.go` + `manage_voucher.go` + `manage_profile.go` (sekalian turunkan 421 baris)
- `chat/chat_service.go` → `chat/manage_chat.go`
- `portal/portal.go` → `portal/manage_portal.go`

**Mapper (aturan baru: unexported kecuali dipakai lintas paket):**
- Audit `ToProto*` exported di `connect/hotspot` — jika hanya dipakai dalam paket, turunkan ke `toProto*`. Arah konversi konsisten: `toProto<Entity>` / `fromProto<Entity>`.

**Konvensi yang sudah benar & dipertahankan (dicatat di guidelines):** `*ConnectHandler`, `New<Type>`, receiver 1–2 huruf, `Id` dari proto-generated = pengecualian sah.

**Verifikasi:** build + test per rename; `golangci-lint run` (revive exported docs ikut menjaga).

---

## FASE 5 — Perapian Layer Port

1. Gabung/perjelas audit: `AuditWriter` + `AuditLogWriter` → satu file `audit.go` berisi keduanya dengan doc pembeda, atau rename `AuditWriter` → `CommandAuditWriter` agar self-explanatory.
2. Payment trio: dokumentasikan pembagian `PaymentGateway` (provider eksternal) / `PaymentProcessor` (persist) / `PaymentReader` (query) di satu file `payment.go`, atau merge jika overlap.
3. Relokasi struct data dari `port` → `domain`: `PPPProfile`, `PPPoESecret`, `VoucherBatch`, `SystemResource`, `SubscriberAccount`, dll → `internal/domain/<domain sesuai>/`. Port hanya berisi interface + tipe param tipis. (Ini perubahan terbesar fase ini; dikerjakan per kelompok domain: ppp → hotspot → device → billing.)
4. Pecah `hotspot_gateway.go` (6.5K) jika >1 interface besar.

**Verifikasi:** build + test per kelompok; boundary check tetap bersih.

---

## FASE 6 — Tuntaskan Modularisasi Mikrotik (lanjutan branch aktif)

1. Migrasi sisa file root ke subpaket: `hotspot_active.go`, `hotspot_profile.go`, `hotspot_user.go` → `hotspot/`; `ppp_active.go`, `ppp_profile.go` → `ppp/`; `ip.go`, `nat.go`, `pool.go` → subpaket sesuai (`firewall/`/`iface/` atau paket baru `ip/`).
2. Hapus duplikasi `queue_gateway.go` root vs `queue/gateway.go`.
3. File root yang tersisa = facade tipis `driver.go` + `gateway.go` saja; dokumentasikan polanya.
4. Ekstrak helper duplikat `setIfNonEmpty` (firewall, ppp, +lainnya) → satu paket internal `mikrotik/internal/rosutil`.

**Verifikasi:** build + test driver + `test/integration` mikrotik.

---

## FASE 7 — Penegakan & Dokumentasi (agar tidak drift lagi)

1. `golangci-lint`: tambah `errorlint` (salah pakai `==` vs `errors.Is`), `wrapcheck` atau `err113` (paksa `%w`), `forbidigo` dengan pattern `connect\.NewError` di luar `pkg/response` & interceptor (penegak fase 1–2 permanen).
2. `Makefile`: target `make check` = build+lint+test; pastikan dipakai CI (`.github/`).
3. Update `AGENTS.md` §1.1 + `DEVELOPMENT-GUIDELINES.md`: struktur aktual lengkap, pola error fault, aturan validasi protovalidate, tabel suffix penamaan, pengecualian sah (Id proto, Engine/Guardrail).
4. Jadikan satu sumber kebenaran struktur folder (file lain merujuk, tidak menduplikasi).

---

## FASE 8 (paralel/opsional) — Gap Test Prioritas

`usecase/metrics` (baru saja di-refactor fase 3), `usecase/chat`, `usecase/customer`, handler streaming SSE (`StreamActiveSessions`, `StreamPing`, `StreamSystemSnapshot`), `Guardrail.MarkdownToWhatsApp`, `BuildExpireMonitorScript`, repo postgres (6/29 → bertahap).

---

## Urutan, Estimasi & Aturan Main

```text
F0 (0.5h) → F1 (besar, inti) → F2 (sedang) → F3 (kecil) → F4 (sedang) → F5 (sedang-besar) → F6 (sedang) → F7 (kecil)
                                                                       F8 berjalan paralel sejak F3
```

- Setiap fase = 1 branch/PR sendiri, selesai dalam keadaan hijau (`go build ./... && go test ./... && golangci-lint run`) sebelum lanjut.
- Proto wire format tidak berubah di semua fase → frontend tidak terdampak (hanya regenerate di F2).
- Fase 1–2 adalah pengungkit terbesar untuk tujuan "konsisten & mudah dikembangkan": setelahnya, menambah fitur baru = tambah sentinel di domain + anotasi proto, tanpa menyentuh `pkg/response` atau menulis validasi manual.
- Catatan: F6 sebaiknya menunggu branch `refactor/mikrotik-driver-modularization` yang sedang berjalan selesai/merge dulu agar tidak konflik.

---

**Apakah plan ini sesuai? Jika ya, saya mulai dari Fase 0 + Fase 1 (arsitektur error pkg/fault + migrasi sentinel ke domain).**
