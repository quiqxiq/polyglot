# PLAN v2: Standardisasi Backend Go Polyglot

## 1. Tujuan

Dokumen ini menggantikan `PLAN.md` sebagai rencana kerja teknis untuk membuat backend Polyglot lebih terstruktur, konsisten, mudah diuji, dan aman dikembangkan.

Target akhirnya bukan sekadar memindahkan file atau mengganti nama tipe. Targetnya adalah satu set aturan yang dapat ditegakkan untuk:

- struktur package dan boundary Clean Architecture;
- penamaan package, file, tipe, fungsi, receiver, dan initialism;
- logging berbasis `pkg/logger` yang sudah digunakan project;
- error domain, error wrapping, dan klasifikasi transport;
- request/response ConnectRPC dan plain HTTP;
- error response yang aman dan konsisten;
- validasi request di boundary dengan protovalidate;
- streaming, worker, repository, dan driver;
- test, CI, lint, dan dokumentasi sebagai enforcement.

Perubahan dilakukan bertahap. Setiap fase harus meninggalkan repository dalam keadaan buildable dan testable. Tidak boleh melakukan rewrite besar tanpa checkpoint.

---

## 2. Baseline Aktual

Audit terhadap commit `4e9aa22` menunjukkan fondasi refactor sudah tersedia, tetapi belum konsisten end-to-end.

### 2.0 Status Audit Terbaru

Status berikut diverifikasi terhadap HEAD `77273db` dan hasil command terbaru.

| Fase | Status | Ringkasan |
|---|---|---|
| F0 | **DONE** | Toolchain CI sudah Go 1.26, baseline test/build/vet lulus, E2E router diberi tag eksplisit, `make check` tersedia. |
| F1 | **PARTIAL** | Logrus `pkg/logger`, structured logging, redaction hook, dan test sudah ada; request correlation dan seluruh log repository belum selesai. |
| F2 | **PARTIAL** | `pkg/fault`, Connect mapper, HTTP mapper, dan sentinel utama sudah ada; seluruh error domain belum dimigrasikan secara konsisten. |
| F3 | **PARTIAL** | Protovalidate aktif pada router, inventory request tersedia, test unary/stream tersedia; inventory belum seluruhnya covered dan validasi manual masih tersisa. |
| F4 | **PARTIAL** | Mapper dan HTTP error envelope sudah distandardisasi pada beberapa area; seluruh endpoint belum memiliki contract test dan pagination contract. |
| F5 | **PARTIAL** | ConnectRPC dan layer boundary checker sudah ditegakkan; test usecase/import integration dan lifecycle worker belum seluruhnya selesai. |
| F6 | **PARTIAL** | Beberapa rename/file split selesai; file besar dan naming stutter legacy masih ada. |
| F7 | **PARTIAL** | Model domain mulai dipindahkan dan interface audit diperjelas; seluruh port/model belum selesai dibersihkan. |
| F8 | **PARTIAL** | Modularisasi Mikrotik dan `rosutil` sudah ada; duplicate audit dan facade verification belum lengkap. |
| F9 | **PARTIAL** | 650 test race lulus, tetapi coverage streaming/worker dan beberapa high-risk flow masih rendah atau belum ada. |
| F10 | **PARTIAL** | CI lint, `make check`, `wrapcheck`, boundary check, `buf lint`, dan generated diff check sudah ada; documentation source of truth dan lint legacy cleanup masih berjalan. |

**Catatan verifikasi:** `go build ./...`, `go vet ./...`, `go test ./... -race -count=1`, golangci-lint v2.1.6, `buf lint`, `buf generate --template buf.gen.yaml`, dan boundary checks lulus. Knowledge graph perlu rebuild setelah commit terbaru sebelum dipakai sebagai bukti coverage berikutnya.

### 2.1 Hal yang sudah ada

- `pkg/fault` sudah tersedia dan tidak bergantung pada `internal/*`.
- `pkg/response` sudah memetakan `fault.Kind` ke ConnectRPC code.
- `buf.yaml` sudah memiliki dependency protovalidate.
- Banyak field protobuf sudah diberi anotasi `buf.validate.field`.
- Usecase hotspot, chat, portal, dan worker WhatsApp sudah sebagian direname.
- Model domain sudah mulai dipindahkan dari `internal/port`.
- Driver Mikrotik sudah memiliki subpackage feature dan `rosutil`.
- `.env` sudah tidak ter-track pada HEAD dan `.gitignore` efektif.
- Repository Postgres dan `pkg/fault` sudah mendapat tambahan test.

### 2.2 Gap yang harus dianggap sebagai baseline blocker

- `go test ./...` sudah hijau: 650 test lulus dengan race detector pada verifikasi terakhir.
- Direct `connect.NewError` pada `internal/adapter/connect` sudah nol; pembuatannya dipusatkan di `pkg/response`.
- Validasi manual `req.Msg.X == ""` masih ada di beberapa handler dan belum seluruhnya dipindahkan ke protovalidate/usecase.
- Test usecase masih mengimpor adapter/driver secara langsung.
- `internal/usecase/importer/router_live_e2e_test.go` masih berada di folder usecase.
- `make check` sudah tersedia dan lulus pada environment dengan golangci-lint v2.1.6.
- CI sudah menjalankan lint dan custom ConnectRPC boundary check; `wrapcheck` aktif dengan baseline/exclusion incremental.
- Logging pada area utama dan driver WhatsApp sudah structured; sebagian log legacy di luar area tersebut masih perlu migrasi.
- `pkg/response` sudah menyediakan HTTP envelope; adminapi, gateway, reports, dan portal sudah memakai pola ini, tetapi audit seluruh HTTP endpoint belum selesai.
- Sentinel sudah ditambahkan pada customer, device, billing, dan hotspot; seluruh domain belum memiliki taxonomy/error inventory lengkap.
- File besar masih ada: `internal/app/app.go`, `internal/app/account_manager.go`, `internal/usecase/user/manage_user.go`, `internal/adapter/connect/billing/billing_handler.go`, dan beberapa file test/fake.
- `go.mod` dan CI sekarang sama-sama menggunakan Go `1.26`.

### 2.3 Prinsip audit

Temuan di atas adalah baseline repository, bukan asumsi. Setiap pekerjaan di plan ini harus mengubah baseline tersebut menjadi assertion atau lint rule yang dapat diverifikasi.

---

## 3. Keputusan Arsitektur yang Dikunci

### 3.1 Boundary

```text
transport inbound
  -> adapter/connect, adapter/http, adapter/ws, adapter/mcp
  -> usecase
  -> domain + port
  -> adapter outbound / driver
```

Aturan:

1. `internal/domain` hanya berisi model, invariant, value object, dan error domain. Domain tidak mengimpor `port`, `usecase`, `adapter`, protobuf, GORM, ConnectRPC, Logrus, atau framework lain.
2. `internal/usecase` hanya mengimpor domain, port, standard library, dan package generik yang tidak membawa transport atau infrastructure.
3. `internal/port` berisi kontrak yang diperlukan consumer. Model bisnis tidak didefinisikan ulang di port.
4. `internal/adapter/connect` adalah penerjemah protobuf ke input usecase dan domain output ke protobuf.
5. `internal/adapter/http` adalah penerjemah JSON/HTTP ke input usecase dan output usecase ke JSON/HTTP.
6. `internal/driver` tidak boleh diimpor oleh usecase. Usecase berinteraksi melalui port.
7. `internal/app` hanya melakukan composition root, lifecycle, dependency injection, dan server wiring. Business rule tidak boleh ditambahkan ke sana.

### 3.2 Protobuf dan wire compatibility

- Field, number, service, dan wire format existing tidak diubah tanpa ADR dan persetujuan eksplisit.
- Penambahan field harus additive dan backward-compatible.
- Generated code tidak diedit manual.
- Mapper adalah satu-satunya lokasi konversi protobuf/domain untuk ConnectRPC.
- Plain HTTP memiliki DTO boundary sendiri; handler tidak mengekspos struct database atau error internal.

### 3.3 Logging

Project tetap menggunakan `pkg/logger` berbasis Logrus karena itu adalah standar yang sudah dikunci di `AGENTS.md` dan implementasi saat ini.

- Tidak ada migrasi ke `log/slog` dalam plan ini.
- API `pkg/logger` menjadi satu-satunya entry point logging production.
- Logrus tetap digunakan sebagai implementation detail.
- Message log baru harus static; data masuk sebagai fields.
- Field wajib memakai snake_case: `request_id`, `user_id`, `device_id`, `operation`, `duration_ms`, `status`.
- Password, token, cookie, API key, secret, full request body, dan PII sensitif tidak boleh masuk log.
- Error top-level dicatat sekali; layer bawah mengembalikan error dengan context.

### 3.4 Error model

`pkg/fault` adalah taxonomy generik. Domain error dideklarasikan di domain masing-masing.

```go
var ErrNotFound = fault.New(fault.KindNotFound, "device: not found")
```

Aturan:

- Sentinel hanya untuk kondisi yang perlu dikenali caller.
- Error context memakai `%w`.
- Error internal yang tidak boleh diketahui client dipetakan ke pesan publik yang aman.
- `connect.NewError` hanya boleh berada di `pkg/response` atau adapter interceptor yang memang menerjemahkan transport error.
- Handler tidak boleh mengklasifikasikan error memakai `errors.Is` lalu membuat ConnectRPC error sendiri.
- `errors.Is`/`errors.As` dipakai di layer yang memang membutuhkan keputusan bisnis atau mapping.

### 3.5 Request validation

Validasi dibagi menjadi tiga lapis:

1. **Wire validation:** protovalidate untuk required, min/max length, range, pattern, enum, dan nested required.
2. **Usecase validation:** aturan bisnis yang memerlukan repository, actor, state, atau kombinasi field.
3. **Infrastructure validation:** validasi configuration, external response, dan driver payload di boundary masing-masing.

Handler tidak melakukan validasi field wajib yang sudah dapat diekspresikan di proto. Handler tetap boleh melakukan adaptasi transport seperti membaca cookie/header, tetapi hasil invalid harus memakai taxonomy yang sama.

### 3.6 Request/response contract

ConnectRPC:

- Request memakai generated protobuf.
- Response memakai generated protobuf.
- Domain error dipetakan dengan `response.MapDomainError`.
- Validation error berasal dari protovalidate dan menghasilkan `CodeInvalidArgument`.
- Detail validasi harus dipertahankan melalui Connect error detail bila library mendukungnya.

Plain HTTP:

- Success response selalu JSON dengan `Content-Type: application/json`.
- Error response memiliki satu bentuk:

```json
{
  "error": {
    "code": "INVALID_ARGUMENT",
    "message": "request validation failed",
    "details": []
  }
}
```

- `details` opsional dan hanya berisi detail aman, misalnya field validation.
- Error internal tidak mengembalikan `err.Error()` ke client.
- HTTP status dipetakan dari `fault.Kind` melalui satu mapper, bukan `strings.Contains(err.Error(), ...)`.

---

## 4. Target Struktur Backend

Struktur aktual boleh memiliki domain tambahan, tetapi pola tanggung jawab harus mengikuti bentuk berikut:

```text
cmd/
  server/main.go
  probe/main.go
  seed/main.go

internal/
  app/
  config/
  domain/<bounded-context>/
  port/
  usecase/<area>/
  adapter/
    connect/<domain>/
    http/middleware/
    http/<resource>/
    mcp/
    postgres/
    redis/
    auth/
    ws/
  driver/<vendor>/

pkg/
  fault/
  logger/
  response/
  retry/
  ping/
  voucher/

api/proto/v1/
api/gen/v1/
test/integration/
```

Generated files, fixtures, integration tests, and test fakes mendapat pengecualian ukuran file jika memang generated atau terpusat secara sengaja. Pengecualian harus dicatat, bukan dibiarkan implisit.

---

## 5. Rencana Eksekusi Bertahap

### Fase 0: Freeze, inventory, dan baseline reproducible

**Tujuan:** membuat kondisi awal dapat diukur sebelum refactor berikutnya.

**Task 0.1: Selaraskan toolchain**

- Putuskan versi Go yang didukung. Rekomendasi: gunakan versi yang sama di `go.mod`, CI, dokumentasi, dan local check.
- Pastikan `golangci-lint`, `buf`, dan generator tersedia melalui Makefile atau CI.
- Jangan mengubah dependency bisnis pada fase ini.

**Task 0.2: Buat command baseline**

Tambahkan target:

```make
build:
	go build ./...

test:
	go test ./... -race -cover

lint:
	golangci-lint run ./...

check: build lint test
```

**Task 0.3: Perbaiki test baseline**

- Root-cause `TestRouterAccountManager_E2E`, bukan men-disable test.
- Jika test membutuhkan external service, beri build tag atau setup eksplisit sehingga `go test ./...` tetap deterministik.
- Pisahkan test integration yang membutuhkan Docker/network dari unit test.

**Acceptance criteria:**

- `go build ./...` lulus.
- `go test ./... -race -cover` lulus tanpa test disabled tersembunyi.
- `make check` dapat dijalankan dari clean checkout.
- Go version di `go.mod` dan CI sama.

**Checkpoint:** tidak boleh masuk fase berikutnya bila baseline belum hijau.

---

### Fase 1: Fondasi logger yang konsisten

**Tujuan:** menetapkan pemakaian `pkg/logger` tanpa mengubah backend logger dari Logrus.

**Task 1.1: Definisikan logging contract**

- Tambahkan helper untuk request correlation bila belum ada: request ID, user ID, device ID, operation.
- Pastikan `Init` aman dipanggil sekali dan test dapat mengganti output writer.

**Task 1.2: Standardisasi API pemakaian**

- Jalur baru memakai `WithComponent(...).WithFields(...).Info/Warn/Error`.
- Migrasikan `Infof`, `Warnf`, `Errorf`, dan `Debugf` production secara bertahap ke static message + fields.
- Error log selalu menggunakan `WithError(err)`.
- Logger tidak boleh menerima secret atau full payload.

**Task 1.3: Middleware request logging**

- Log method, route, status, duration, request ID, dan ukuran response.
- Jangan log Authorization, Cookie, body, query sensitif, atau token.
- Streaming request dicatat saat start dan close/cancel, bukan setiap frame kecuali debug yang dibatasi.

**Task 1.4: Test logger**

- Uji JSON production dan text development.
- Uji field component, request ID, error, dan redaction.
- Uji bahwa logger tidak membocorkan token/password.

**Acceptance criteria:**

- Tidak ada `log.Printf`, `log.Println`, atau `fmt.Println` di production backend.
- Tidak ada logging secret/credential/token.
- Semua request HTTP/ConnectRPC utama memiliki request ID dan component.
- Log message baru tidak menginterpolasi identifier ke message.

---

### Fase 2: Error taxonomy dan public error mapping

**Tujuan:** membuat error dapat diklasifikasikan secara konsisten dan aman di semua transport.

**Task 2.1: Stabilkan `pkg/fault`**

- Pertahankan `Kind` yang sudah ada.
- Tambahkan test untuk nil, wrapping berlapis, `errors.Is`/`errors.As`, dan unknown error.
- Evaluasi apakah `fault.Error` perlu menyimpan public code terpisah dari message internal. Jika ya, catat ADR sebelum implementasi.

**Task 2.2: Domain error inventory**

Untuk setiap bounded context aktual (`billing`, `bot`, `cashbook`, `command`, `customer`, `device`, `hotspot`, `llm`, `notification`, `plan`, `ppp`, `registration`, `reporting`, `session`, `setting`, `skill`, `subscription`, dan `audit` bila memiliki business failure):

- inventarisasi error yang bisa di-match;
- pindahkan sentinel dari usecase ke domain;
- gunakan nama yang konsisten berdasarkan domain, bukan nama transport;
- gunakan pesan Inggris lowercase dengan prefix domain;
- hapus sentinel duplikat atau buat alias hanya jika ada consumer nyata dan masa migrasi yang jelas.

**Task 2.3: Standardisasi wrapping**

- Tambahkan `%w` pada error propagation.
- Jangan wrap error hanya untuk menambah message yang tidak berguna.
- Jangan gunakan string matching untuk menentukan status.
- Bedakan error business, dependency unavailable, cancellation, dan programmer/configuration error.

**Task 2.4: ConnectRPC mapper**

- `pkg/response.MapDomainError` menjadi satu-satunya mapper domain-to-Connect.
- Hapus pass-through `connect.Error` bila itu memungkinkan menyembunyikan error transport yang dibuat sembarangan; jika dipertahankan, dokumentasikan boundary-nya.
- Handler hanya `return nil, response.MapDomainError(err)`.
- Dependency nil/configuration failure harus dibuat sebagai error typed/fault sebelum handler, atau dipetakan melalui helper transport terpusat.

**Task 2.5: Plain HTTP mapper**

- Buat mapper `fault.Kind` ke HTTP status.
- Buat satu `WriteError` yang menghasilkan envelope JSON standar.
- Pesan internal tidak boleh dikirim mentah.
- Error detail validation boleh dikirim; stack trace dan database/driver detail tidak boleh.

**Acceptance criteria:**

- `pkg/response` tidak mengimpor `internal/*`.
- `connect.NewError` tidak muncul di handler atau `internal/app`.
- Tidak ada `strings.Contains(err.Error(), ...)` untuk klasifikasi transport.
- Setiap error client-facing memiliki machine-readable code dan safe message.
- Test mapping mencakup seluruh `fault.Kind`.

---

### Fase 3: Request validation dan kontrak input

**Tujuan:** memindahkan validasi wire ke protovalidate dan menjaga business validation di usecase.

**Task 3.1: Inventarisasi seluruh request protobuf**

- Buat tabel message request, field wajib, tipe, enum, pattern, range, dan rule bisnis.
- Audit seluruh file `api/proto/v1`, bukan hanya 18 file yang disebut di plan lama.
- Tetapkan field ID: format, whitespace policy, max length, dan apakah ID internal atau public.

**Task 3.2: Lengkapi anotasi protovalidate**

- `required` untuk message nested.
- `min_len`/`max_len` untuk string.
- pattern untuk ID atau identifier yang memang memiliki format.
- range untuk angka, pagination, timeout, dan amount.
- enum constraint untuk mode/status yang tertutup.
- repeated item count dan nested validation jika diperlukan.

**Task 3.3: Satu validation interceptor**

- Pilih satu lokasi canonical untuk ConnectRPC validation.
- Pasang pada semua unary dan streaming handler yang didukung library.
- Hapus implementasi wrapper yang tidak dipakai atau jadikan helper canonical.
- Uji bahwa invalid request ditolak sebelum usecase dipanggil.

**Task 3.4: Bersihkan handler**

- Hapus check field kosong yang sudah dicakup proto.
- Pertahankan check kombinasi field yang memang business/semantic rule dan pindahkan ke usecase bila tidak membutuhkan transport.
- Header/cookie authentication validation tetap di transport, tetapi errornya melalui mapper/helper konsisten.

**Task 3.5: HTTP validation**

- Decoder harus menolak malformed JSON dan, bila diperlukan, unknown fields.
- Batasi body size.
- Validasi DTO HTTP sebelum usecase.
- Error field validation menggunakan envelope yang sama.

**Acceptance criteria:**

- Request invalid gagal dengan `CodeInvalidArgument` untuk ConnectRPC.
- Usecase tidak dipanggil pada request yang gagal wire validation.
- Tidak ada duplicate required-field validation di handler.
- HTTP invalid JSON, missing field, wrong type, dan oversized body memiliki response konsisten.
- `buf lint` dan regeneration berhasil.

---

### Fase 4: Request/response contract dan mapper

**Tujuan:** menyamakan pola data keluar tanpa merusak wire contract.

**Task 4.1: Standarkan mapper naming**

- Gunakan `toProto<Entity>` dan `fromProto<Entity>` untuk fungsi package-private.
- Export hanya jika benar-benar dipakai lintas package.
- Mapper tidak melakukan business logic atau database lookup.
- Pisahkan mapper per bounded context bila file terlalu besar.

**Task 4.2: Response semantics**

- Get operation mengembalikan satu resource atau error not found.
- List operation mengembalikan repeated field non-nil sesuai kebutuhan generated protobuf.
- Create/update mengembalikan resource yang sudah tersimpan.
- Delete memakai response eksplisit dan idempotency policy yang terdokumentasi.
- Streaming memakai event message yang jelas: snapshot, update, error, completion.

**Task 4.3: Pagination and list consistency**

- Inventarisasi list endpoint.
- Untuk list yang berpotensi besar, tambahkan pagination secara additive.
- Jangan mengubah field existing tanpa compatibility plan.
- Standarkan `page_size`, `page_token`, `next_page_token` bila contract baru diperlukan.

**Task 4.4: HTTP DTO and response writer**

- Buat response DTO yang tidak mengekspos GORM model.
- Standarkan success envelope hanya jika tidak mematahkan consumer existing; bila consumer existing berbeda, lakukan migrasi endpoint per endpoint.
- Tulis contract tests untuk status, content type, body shape, dan error shape.

**Acceptance criteria:**

- Setiap endpoint memiliki request/response contract yang terdokumentasi.
- Tidak ada GORM model atau internal error yang dikirim langsung.
- Mapper exported/non-exported konsisten.
- Contract tests tersedia untuk setiap transport family.

---

### Fase 5: Boundary dan dependency inversion

**Tujuan:** memastikan usecase benar-benar independen dari adapter dan driver.

**Task 5.1: Static boundary check**

Larangan production:

```text
internal/usecase -> internal/adapter
internal/usecase -> internal/driver
internal/domain -> internal/port
internal/domain -> api/gen
internal/domain -> external framework
```

- Buat script atau analyzer yang gagal jika import tersebut muncul.
- Test integration boleh mengimpor adapter/driver hanya di `test/integration` atau file yang diberi tag/pengecualian terdokumentasi.

**Task 5.2: Pindahkan test integration**

- Pindahkan `router_live_e2e_test.go` ke `test/integration`.
- Bersihkan import driver langsung dari unit test usecase.
- Fake harus mengimplementasikan port yang dibutuhkan consumer.

**Task 5.3: Metrics worker**

- Pastikan `PingStreamWorker` hanya bergantung pada repository dan resolver/port.
- Validasi lifecycle worker: start, cancellation, cleanup, restart, error logging, dan shutdown.
- Semua goroutine harus berhenti melalui context atau wait mechanism.

**Task 5.4: Composition root**

- `internal/app/app.go` hanya wiring.
- Ekstrak registry, router construction, and lifecycle groups bila file tetap di atas 400-500 baris.
- Jangan menaruh authorization/business policy baru di composition root.

**Acceptance criteria:**

- Boundary script lulus.
- Unit test usecase tidak mengimpor adapter/driver.
- Integration test terpisah dari default unit suite atau memiliki dependency setup yang jelas.
- Tidak ada goroutine worker tanpa cancellation/wait path.

---

### Fase 6: Naming dan package structure

**Tujuan:** membuat nama konsisten dan mencerminkan tanggung jawab.

**Aturan:**

- Package lowercase satu kata dan tidak plural.
- File Go memakai snake_case sesuai tanggung jawab.
- Usecase memakai pola `<verb>_<noun>.go`.
- Tipe orchestration memakai `*UseCase`.
- Background process memakai `*Worker`.
- `Engine`, `Guardrail`, dan `ContextManager` dipertahankan hanya bila memang konsepnya engine/guardrail/context builder, lalu dicatat sebagai exception.
- Initialism memakai `ID`, `HTTP`, `URL`, `RPC`, `SSE`, `MCP`, `WA` pada identifier Go.
- Proto-generated `Id` tidak direname manual; dianggap wire compatibility exception.
- Receiver konsisten satu atau dua huruf per package.

**Task 6.1: Rename dengan gopls**

- Lakukan satu logical rename per commit.
- Verifikasi references, tests, generated code boundary, dan public consumer setelah tiap rename.

**Task 6.2: Pecah file besar**

Prioritas:

1. `internal/app/account_manager.go`.
2. `internal/app/app.go`.
3. `internal/usecase/user/manage_user.go`.
4. `internal/adapter/connect/billing/billing_handler.go`.
5. `internal/port/mocktest/fakes_core.go`.
6. `internal/driver/mikrotik/hotspot/gateway.go` bila bertambah.

Pecah berdasarkan responsibility, bukan berdasarkan jumlah fungsi secara mekanis.

**Acceptance criteria:**

- Tidak ada nama lama yang tersisa kecuali compatibility exception terdokumentasi.
- File production non-generated berada di bawah 500 baris atau memiliki exception tertulis.
- `revive` tidak menghasilkan missing exported documentation pada API baru.

---

### Fase 7: Port dan model domain

**Tujuan:** membuat port berisi kontrak, bukan tempat penampungan semua data.

**Task 7.1: Model relocation**

- Audit alias seperti `PPPProfile`, `PPPoESecret`, `VoucherBatch`, `SystemResource`, dan `SubscriberAccount`.
- Domain model berada di bounded context masing-masing.
- Port hanya memiliki interface, parameter command/query, dan result type yang memang merupakan contract.
- Hindari alias jangka panjang tanpa alasan migrasi.

**Task 7.2: Audit interface overlap**

- `AuditWriter` vs `AuditLogWriter`: jelaskan low-level command audit vs business audit atau gabungkan bila sama.
- `PaymentGateway`, `PaymentProcessor`, `PaymentReader`: dokumentasikan direction, ownership, dan transaction semantics.
- Interface didefinisikan oleh consumer dan tidak dibuat hanya untuk memuaskan mocking.

**Task 7.3: Pecah gateway besar**

- Pecah interface besar berdasarkan capability.
- Setiap interface harus dapat dijelaskan dalam satu kalimat.
- Jangan memindahkan transport concern ke port.

**Acceptance criteria:**

- Tidak ada business entity baru yang ditempatkan di `internal/port`.
- Setiap interface memiliki consumer dan implementor yang jelas.
- Tidak ada circular dependency akibat relokasi.

---

### Fase 8: Modularisasi driver Mikrotik

**Tujuan:** menyelesaikan modularisasi tanpa menduplikasi behavior.

**Task 8.1: Root facade audit**

- Root `internal/driver/mikrotik` hanya lifecycle, shared transport, facade, dan error yang benar-benar lintas feature.
- Feature implementation berada di `dhcp`, `firewall`, `hotspot`, `iface`, `ppp`, `queue`, `system`.

**Task 8.2: Duplicate detection**

- Audit root vs subpackage untuk hotspot, PPP, queue, IP, NAT, pool, dan system.
- Hapus duplikasi hanya setelah call sites dan integration tests terpetakan.

**Task 8.3: Driver errors**

- Driver error internal memiliki context vendor.
- Driver tidak membuat ConnectRPC error.
- Dependency failure dipetakan ke `fault.KindUnavailable` hanya pada boundary yang tepat.

**Acceptance criteria:**

- Root facade tipis dan terdokumentasi.
- Tidak ada dua implementasi aktif untuk operasi yang sama.
- Driver unit/integration tests lulus.

---

### Fase 9: Test coverage dan contract verification

**Tujuan:** menutup risiko behavior, bukan mengejar angka coverage semata.

Prioritas test:

- `TestRouterAccountManager_E2E` dan lifecycle app.
- `PingStreamWorker` start/cancel/cleanup/error.
- `ChatUseCase` dan `CustomerUseCase`.
- `StreamActiveSessions`, `StreamPing`, `StreamSystemSnapshot`, `StreamLogs`.
- `Guardrail.MarkdownToWhatsApp`.
- `BuildExpireMonitorScript`.
- ConnectRPC validation and error mapping.
- Plain HTTP JSON/error contract.
- Postgres repository critical write paths.
- Auth token, refresh token, cookie, and authorization failure.
- Driver command builder and parser.

Test requirements:

- Test semantics with `errors.Is`, `errors.As`, `fault.KindOf`, and status codes; jangan mengunci full error string kecuali itu contract publik.
- Streaming tests harus memverifikasi cancellation dan tidak ada goroutine leak.
- HTTP tests memakai `httptest` dan real handler chain.
- Integration tests memakai explicit build tag or service setup.

**Acceptance criteria:**

- Semua high-risk flow memiliki test.
- Test failure menyebut function, input, got, dan want.
- Race test lulus.
- Tidak ada test yang disabled tanpa alasan dan mekanisme eksekusi yang terdokumentasi.

---

### Fase 10: Enforcement CI dan dokumentasi

**Tujuan:** mencegah drift setelah refactor.

**Task 10.1: Lint enforcement**

Aktifkan dan konfigurasi sesuai kompatibilitas versi `golangci-lint`:

- `errorlint`.
- `forbidigo` untuk direct `connect.NewError` di luar allowlist.
- `wrapcheck` atau `err113` setelah false-positive audit.
- `revive` exported docs.
- `bodyclose`.
- `exhaustive` hanya pada enum/switch yang memang closed.

**Task 10.2: CI**

CI minimal menjalankan:

```text
go mod verify
```

Proto job menjalankan:

```text
buf lint
buf generate --diff
```

Jika generated output harus committed, CI harus gagal ketika hasil generate berbeda dari Git.

**Task 10.3: Documentation source of truth**

- `DEVELOPMENT-GUIDELINES.md` menjadi aturan implementasi.
- `AGENTS.md` hanya berisi instruksi agent dan referensi ke guidelines, bukan salinan struktur lengkap yang mudah stale.
- `PLAN-BACKEND-STANDARDIZATION-V2.md` menjadi roadmap.
- Buat ADR untuk keputusan yang mengubah public error envelope, pagination, Go version, atau wire compatibility.

**Acceptance criteria:**

- Pull request gagal jika check, lint, test, boundary, atau generated code gagal.
- Tidak ada dokumentasi yang mencantumkan file lama yang sudah dihapus/direname.
- Aturan baru dapat ditemukan dari satu sumber canonical.

---

## 5.1 Status Task Detail

### F0: Freeze, Inventory, dan Baseline

| Task | Status | Bukti atau sisa |
|---|---|---|
| 0.1 Selaraskan toolchain | **DONE** | `go.mod` dan CI memakai Go 1.26. Lint v2.1.6 dapat dijalankan melalui `go run`; instalasi binary lokal belum dijamin. |
| 0.2 Command baseline | **DONE** | `build`, `vet`, `test`, `lint`, `check`, `check-connect-errors`, dan `test-mikrotik-e2e` tersedia. |
| 0.3 Perbaiki test baseline | **DONE** | `TestRouterAccountManager_E2E` diberi build tag `mikrotik_e2e`; default suite tidak bergantung router real. |

### F1: Fondasi Logger

| Task | Status | Bukti atau sisa |
|---|---|---|
| 1.1 Logging contract | **PARTIAL** | `pkg/logger` tetap Logrus dan structured fields sudah digunakan; request correlation ID belum diterapkan menyeluruh. |
| 1.2 API pemakaian | **PARTIAL** | Driver WhatsApp, BotEngine, app lifecycle, probe, user, dan beberapa adapter sudah migrated; audit seluruh repository masih diperlukan. |
| 1.3 Request logging middleware | **PARTIAL** | Middleware logger sudah ada, tetapi field schema, redaction, dan streaming lifecycle belum diverifikasi di semua route. |
| 1.4 Test logger | **DONE** | Test JSON/text, fields, errors, dan redaction token/JID/phone/password/payload sudah ada. |

### F2: Error Taxonomy dan Mapping

| Task | Status | Bukti atau sisa |
|---|---|---|
| 2.1 Stabilkan `pkg/fault` | **DONE** | `Kind`, `New`, `Wrap`, `KindOf`, dan unit test tersedia. |
| 2.2 Domain error inventory | **PARTIAL** | Sentinel tersedia pada domain utama (`billing`, `customer`, `device`, `hotspot`, dan lainnya); masih ada `fmt.Errorf` input/repository di beberapa usecase. |
| 2.3 Standardisasi wrapping | **PARTIAL** | `wrapcheck` aktif dan beberapa jalur sudah dibungkus; masih ada backlog legacy yang memakai direct port return. |
| 2.4 ConnectRPC mapper | **DONE** | Direct `connect.NewError` di `internal/adapter/connect` sudah nol; helper transport dan `MapDomainError` tersedia. |
| 2.5 HTTP mapper | **PARTIAL** | `pkg/response` dan beberapa handler sudah memakai envelope; semua endpoint HTTP belum diaudit/ditest. |

### F3: Request Validation

| Task | Status | Bukti atau sisa |
|---|---|---|
| 3.1 Inventory request protobuf | **TODO** | Belum ada tabel canonical seluruh request, field rule, ID policy, dan business rule. |
| 3.2 Lengkapi anotasi protovalidate | **PARTIAL** | Banyak anotasi sudah ada; coverage semua request, max length, repeated count, dan nested rules belum dibuktikan. |
| 3.3 Satu validation interceptor | **PARTIAL** | `DefaultHandlerOptions` memakai protovalidate; contract test unary/stream ada, tetapi semua route dan streaming variants belum diaudit. |
| 3.4 Bersihkan handler | **PARTIAL** | Banyak direct transport errors sudah dibersihkan; validasi kombinasi dan field manual masih ada di beberapa handler. |
| 3.5 HTTP validation | **PARTIAL** | Malformed JSON dan error envelope sudah dites pada beberapa route; body limit, unknown fields, dan DTO validation belum seragam. |

### F4: Request/Response Contract

| Task | Status | Bukti atau sisa |
|---|---|---|
| 4.1 Mapper naming | **PARTIAL** | Mapper banyak sudah package-private dan konsisten; beberapa unused mapper/naming legacy masih ada. |
| 4.2 Response semantics | **PARTIAL** | Contract ConnectRPC dan streaming validation sudah dites; seluruh Get/List/Create/Update/Delete belum memiliki contract suite lengkap. |
| 4.3 Pagination | **PARTIAL** | User list sudah memiliki page/page_size; inventory dan standardisasi pagination semua list endpoint belum selesai. |
| 4.4 HTTP DTO/writer | **PARTIAL** | Envelope error adminapi/gateway/reports/portal sudah ditambahkan; success envelope dan DTO seluruh HTTP endpoint belum disatukan. |

### F5: Boundary dan Dependency Inversion

| Task | Status | Bukti atau sisa |
|---|---|---|
| 5.1 Static boundary check | **PARTIAL** | `check-connect-errors.sh` tersedia; static check domain/usecase imports adapter/driver belum dibuat menyeluruh. |
| 5.2 Pindahkan integration test | **TODO** | `router_live_e2e_test.go` dan beberapa test usecase masih perlu dipindahkan/diisolasi sesuai plan. |
| 5.3 Metrics worker | **PARTIAL** | Resolver melalui port/function dan lifecycle worker sudah ada; coverage cancel/restart/leak masih perlu ditambah. |
| 5.4 Composition root | **PARTIAL** | `app.go` tetap composition root dan logging lifecycle sudah dirapikan; file masih besar dan perlu pemisahan wiring bila bertambah. |

### F6: Naming dan Package Structure

| Task | Status | Bukti atau sisa |
|---|---|---|
| 6.1 Rename dengan gopls | **PARTIAL** | Chat, portal, hotspot, ping worker, dan WA worker sudah direname; naming stutter legacy pada domain/driver masih ada. |
| 6.2 Pecah file besar | **PARTIAL** | Hotspot usecase sudah dipecah; `app.go`, `account_manager.go`, `manage_user.go`, handler billing, dan fake besar masih ada. |

### F7: Port dan Model Domain

| Task | Status | Bukti atau sisa |
|---|---|---|
| 7.1 Model relocation | **PARTIAL** | Banyak model sudah berada di domain dan port memakai alias/parameter; audit penghapusan alias jangka panjang belum selesai. |
| 7.2 Interface overlap | **PARTIAL** | Audit writer dan payment interfaces sudah didokumentasikan; consumer/ownership semua interface belum diverifikasi. |
| 7.3 Pecah gateway besar | **PARTIAL** | Capability gateway sudah terbagi pada beberapa file; ukuran dan cohesion semua gateway belum selesai diaudit. |

### F8: Modularisasi Mikrotik

| Task | Status | Bukti atau sisa |
|---|---|---|
| 8.1 Root facade audit | **PARTIAL** | Subpackage `dhcp`, `firewall`, `hotspot`, `iface`, `ppp`, `queue`, `system`, dan `rosutil` tersedia; audit facade formal belum selesai. |
| 8.2 Duplicate detection | **PARTIAL** | Beberapa duplikasi sudah dihapus; root/subpackage duplicate audit lengkap belum dibuat. |
| 8.3 Driver errors | **PARTIAL** | Driver tidak membuat ConnectRPC error; taxonomy `Unavailable` untuk semua dependency failure belum seragam. |

### F9: Test Coverage dan Contract Verification

| Task | Status | Bukti atau sisa |
|---|---|---|
| High-risk flow tests | **PARTIAL** | 650 test lulus dengan race detector; beberapa streaming worker, SSE, dan driver masih coverage rendah. |
| Error/validation contracts | **PARTIAL** | `pkg/fault`, ConnectRPC unary/stream, HTTP adminapi/gateway/reports, dan logger sudah dites; seluruh endpoint belum. |
| Integration isolation | **PARTIAL** | Real-router E2E sudah explicit tag; seluruh integration test belum memiliki satu convention yang sama. |

### F10: Enforcement dan Dokumentasi

| Task | Status | Bukti atau sisa |
|---|---|---|
| 10.1 Lint enforcement | **DONE** | `errorlint`, `wrapcheck`, `revive`, `bodyclose`, `exhaustive`, dan custom Connect boundary check aktif; full lint terakhir `0 issues`. |
| 10.2 CI | **PARTIAL** | CI menjalankan build/vet/race test/lint/boundary, `go mod verify`, `buf lint`, dan generated diff check; workflow matrix/security/integration terpisah belum lengkap. |
| 10.3 Documentation source of truth | **PARTIAL** | Plan v2 dan guidelines sudah tersedia; `AGENTS.md` masih memuat struktur duplikat yang dapat stale. |

---

## 6. Urutan Dependency

Urutan wajib:

```text
F0 baseline
  -> F1 logger contract
  -> F2 error taxonomy and transport mapping
  -> F3 request validation
  -> F4 request/response contract
  -> F5 boundary
  -> F6 naming and file structure
  -> F7 port/domain cleanup
  -> F8 driver modularization
  -> F9 test gaps
  -> F10 enforcement and documentation
```

F9 boleh berjalan paralel setelah F2 untuk area yang tidak mengubah contract. Rename, model relocation, dan proto changes tidak boleh berjalan paralel pada file/package yang sama.

Setiap fase menjadi branch/PR terpisah atau logical commit group. Setiap PR wajib memiliki:

- scope yang jelas;
- daftar file utama;
- test yang ditambahkan/diubah;
- command verification;
- risk dan rollback note;
- bukti bahwa boundary dan wire format tidak berubah, bila relevan.

---

## 7. Definition of Done

Perubahan backend dianggap selesai hanya jika:

- `go build ./...` lulus.
- `go vet ./...` lulus.
- `go test ./... -race -cover` lulus.
- `golangci-lint run ./...` lulus.
- `buf lint` lulus.
- Generated code tidak memiliki diff tak terduga.
- Tidak ada direct `connect.NewError` di luar allowlist.
- Tidak ada direct logging API selain `pkg/logger` pada production.
- Tidak ada secret atau PII sensitif pada log.
- Error dapat diklasifikasikan tanpa string matching.
- HTTP error envelope konsisten.
- ConnectRPC error mapping konsisten.
- Request invalid ditolak di boundary.
- Usecase tidak mengimpor adapter/driver.
- Semua goroutine memiliki cancellation atau wait path.
- Test integration dipisahkan dan dapat dijalankan secara eksplisit.
- Dokumentasi dan naming sesuai struktur aktual.
- Tidak ada file production non-generated di atas 500 baris tanpa exception tertulis.

---

## 8. Risiko dan Mitigasi

| Risiko | Dampak | Mitigasi |
|---|---|---|
| Mengubah error message publik | Client bergantung pada text lama | Stabilkan machine code; migrasikan message secara terkontrol |
| Memindahkan validasi dari handler | Perubahan status atau detail error | Contract test sebelum dan sesudah |
| Menyatukan response envelope HTTP | Client lama gagal decode | Migrasi endpoint bertahap atau versioned response |
| Rename tipe lintas package | Compile break dan generated reference | Pakai gopls rename, satu rename per commit |
| Relokasi model port ke domain | Import cycle | Pindahkan satu bounded context per checkpoint |
| Lint terlalu ketat | False positive dan bypass rule | Audit false positive, allowlist sempit, jangan menonaktifkan global |
| Streaming worker refactor | Goroutine leak atau lost event | Test cancellation, race, goleak bila sesuai |
| Logging migration | Hilangnya observability | Test output dan field schema sebelum menghapus pola lama |
| Go version mismatch | CI gagal meski lokal lulus | Satu versi canonical di `go.mod`, CI, Makefile, dan docs |

---

## 9. Checklist Implementasi Pertama

Implementasi sebaiknya dimulai dengan urutan kecil berikut, bukan langsung melakukan seluruh refactor:

1. Selaraskan Go version antara `go.mod` dan CI.
2. Perbaiki `TestRouterAccountManager_E2E` sampai baseline hijau.
3. Tambahkan `make check` dan installable lint step di CI.
4. Tambahkan `forbidigo` untuk direct `connect.NewError` dalam mode report-only terlebih dahulu.
5. Selesaikan satu vertical slice error end-to-end, direkomendasikan billing atau auth:
   - domain sentinel;
   - usecase wrapping;
   - ConnectRPC mapper;
   - HTTP mapper bila endpoint tersedia;
   - contract tests.
6. Jadikan slice tersebut template untuk domain lain.
7. Setelah pola terbukti, migrasikan handler lain secara bertahap.

Jangan memulai dengan rename besar atau memindahkan semua model sekaligus. Error mapping, validation, dan response contract adalah fondasi yang harus stabil terlebih dahulu.

---

## 10. Status terhadap `PLAN.md` Lama

`PLAN.md` lama tetap berguna sebagai riwayat temuan, tetapi kurang lengkap untuk target baru karena belum mendefinisikan:

- logging contract berbasis Logrus yang sudah dipakai project;
- HTTP response envelope dan safe error message;
- request/response contract lintas transport;
- validasi malformed JSON, body limit, dan external response;
- status error yang tidak boleh ditentukan dengan string matching;
- compatibility policy untuk error text dan wire format;
- concurrency lifecycle dan streaming contract;
- Go version mismatch;
- CI enforcement untuk proto generation dan boundary.

Karena itu, status planning baru tidak lagi memakai klaim “F0-F7 selesai” hanya berdasarkan perubahan file. Fase dianggap selesai berdasarkan acceptance criteria dan command verification yang tercantum di dokumen ini.
