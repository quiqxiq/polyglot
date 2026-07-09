# CLAUDE.md — Polyglot (Go Backend)

**Dokumen ini adalah instruksi operasional untuk AI agent (Claude Code atau agent lain) yang menulis kode di repo ini. Ini bukan panduan gaya untuk kontributor manusia.** Baca ulang dokumen ini di awal setiap task, bukan cuma sekali di awal sesi — kalau ada bagian yang bertentangan dengan pola yang "terlihat masuk akal" dari kode lain di repo, dokumen ini yang menang, bukan pola dominan.

Referensi terkait (dibaca setelah dokumen ini, bukan sebagai pengganti): `Polyglot-Architecture.md` (alasan arsitektur), `TECH-STACK-DAN-PERSIAPAN.md` (versi library).

**Cara memakai dokumen ini:** sebelum membuat file apa pun, cek Bagian 1 dulu untuk menentukan path dan nama filenya. Baru setelah itu tulis isi filenya sesuai Bagian 2-9.

---

## 0. Prinsip Non-Negotiable

1. **Boundary layer adalah hukum.** `domain` tidak pernah impor `adapter`/`driver`/framework eksternal. `usecase` hanya bergantung ke `domain` dan `port`. Kalau sebuah perubahan memaksa boundary ini dilanggar, itu tanda desainnya salah — stop, jangan dipaksakan.
2. **Setiap penyimpangan dari aturan file ini wajib ditandai komentar `// DEVIASI: <alasan>`** tepat di atas kode yang menyimpang. Tanpa komentar itu, kode dianggap melanggar dokumen ini, titik — tidak ada pengecualian diam-diam.
3. **Jangan menebak penempatan file.** Kalau Bagian 1 tidak jelas mencakup kasus yang sedang dikerjakan, ikuti algoritma di §1.2, dan kalau masih ambigu, nyatakan asumsi penempatan secara eksplisit di ringkasan pekerjaan sebelum melanjutkan — jangan diam-diam menaruh file di tempat yang "kelihatannya cocok".
4. **Jangan membuat abstraksi/pola baru yang tidak ada di dokumen ini** (mis. layer baru selain domain/usecase/port/adapter/driver, konvensi penamaan baru) tanpa menyatakannya eksplisit sebagai proposal, bukan langsung dieksekusi sebagai fakta.

---

## 1. Struktur Folder & Penempatan File

### 1.1 Struktur Lengkap — Definitif, Satu-Satunya yang Berlaku

Ini satu-satunya sumber kebenaran struktur folder proyek. Tidak ada versi lain di dokumen mana pun yang menggantikan ini.

```
polyglot/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── domain/
│   │   ├── device/
│   │   │   ├── device.go
│   │   │   └── errors.go
│   │   ├── command/
│   │   │   ├── command.go
│   │   │   └── policy.go
│   │   ├── session/
│   │   │   └── session.go
│   │   ├── customer/
│   │   │   └── customer.go
│   │   ├── subscription/
│   │   │   └── subscription.go
│   │   ├── plan/
│   │   │   └── plan.go
│   │   └── billing/
│   │       ├── invoice.go
│   │       └── payment.go
│   │
│   ├── usecase/
│   │   ├── network/
│   │   │   ├── execute_command.go
│   │   │   ├── get_device_status.go
│   │   │   ├── push_config.go
│   │   │   └── stream_output.go
│   │   └── business/
│   │       ├── manage_customer.go
│   │       ├── manage_subscription.go
│   │       ├── manage_plan.go
│   │       └── manage_invoice.go
│   │
│   ├── port/
│   │   ├── device_driver.go
│   │   ├── device_repository.go
│   │   ├── customer_repository.go
│   │   ├── subscription_repository.go
│   │   ├── invoice_repository.go
│   │   ├── credential_vault.go
│   │   └── audit_writer.go
│   │
│   ├── adapter/
│   │   ├── http/
│   │   │   ├── router.go
│   │   │   ├── device_handler.go
│   │   │   ├── customer_handler.go
│   │   │   ├── subscription_handler.go
│   │   │   ├── invoice_handler.go
│   │   │   └── middleware/
│   │   │       ├── auth.go
│   │   │       └── rbac.go
│   │   ├── mcp/
│   │   │   ├── server.go
│   │   │   ├── tool_get_device_status.go
│   │   │   ├── tool_run_command.go
│   │   │   └── tool_push_config.go
│   │   ├── ws/
│   │   │   ├── hub.go
│   │   │   └── device_stream_handler.go
│   │   ├── postgres/
│   │   │   ├── device_repository.go
│   │   │   ├── customer_repository.go
│   │   │   ├── subscription_repository.go
│   │   │   └── invoice_repository.go
│   │   ├── vault/
│   │   │   └── aes_vault.go
│   │   └── auth/
│   │       ├── jwt.go
│   │       └── casbin.go
│   │
│   ├── driver/
│   │   ├── mikrotik/
│   │   │   ├── driver.go
│   │   │   └── commands.go
│   │   ├── cisco/
│   │   │   ├── driver.go
│   │   │   └── commands.go
│   │   ├── genericssh/
│   │   │   ├── driver.go
│   │   │   └── commands.go
│   │   ├── netconf/
│   │   │   ├── driver.go
│   │   │   └── commands.go
│   │   ├── zteolt/
│   │   │   ├── snmp.go
│   │   │   ├── telnet.go
│   │   │   └── commands.go
│   │   ├── huaweiolt/
│   │   │   ├── driver.go
│   │   │   └── commands.go
│   │   └── genieacs/
│   │       ├── client.go
│   │       └── commands.go
│   │
│   ├── platformdef/
│   │   ├── *.yaml
│   │   └── README.md
│   │
│   ├── registry/
│   │   └── registry.go
│   ├── audit/
│   │   └── writer.go
│   └── config/
│       └── config.go
│
├── pkg/
│   └── retry/
│       └── retry.go
│
├── api/
│   ├── openapi.yaml
│   └── mcp-tools.md
│
├── migrations/
│   ├── 000001_create_devices_table.up.sql
│   └── 000001_create_devices_table.down.sql
│
├── deployments/
│   ├── docker/
│   │   └── Dockerfile
│   └── docker-compose.yml
│
├── scripts/
│   └── seed.go
│
├── test/
│   └── integration/
│
├── docs/
│   └── adr/
│       ├── 0001-pilih-gin-daripada-echo.md
│       └── 0002-devicedriver-tanpa-session-terpisah.md
│
├── .golangci.yml
├── .github/
│   └── workflows/
│       └── ci.yml
├── go.mod
├── go.sum
├── Makefile
├── README.md
└── CLAUDE.md
```

> **Catatan penamaan:** domain "paket layanan ISP" diberi nama Go `plan`, bukan `package` — karena `package` adalah **reserved keyword** di Go (dipakai untuk `package` statement itu sendiri) dan tidak bisa dipakai sebagai nama package. Jangan pernah membuat folder/package Go bernama `package` di mana pun dalam proyek ini.

### 1.2 Algoritma Keputusan Penempatan File

Sebelum membuat file baru apa pun, jawab urutan berikut — berhenti di pertanyaan pertama yang jawabannya "ya", ikuti pathnya:

1. **Apakah ini entity/aturan bisnis murni, tidak melakukan I/O, tidak impor library eksternal (selain stdlib)?** → `internal/domain/<nama_domain>/`
2. **Apakah ini orkestrasi satu use case — memanggil satu/lebih `port`, tidak tahu detail implementasi HTTP/DB/device?** → `internal/usecase/network/` atau `internal/usecase/business/`
3. **Apakah ini KONTRAK (interface Go) yang akan diimplementasikan lebih dari satu adapter/driver?** → `internal/port/`
4. **Apakah ini implementasi konkret yang bicara ke sistem luar** (HTTP request/response, protokol MCP, koneksi WebSocket, query Postgres, protokol vendor device tertentu)? → `internal/adapter/<jenis>/` untuk HTTP/MCP/WS/Postgres/Vault/Auth, atau `internal/driver/<vendor>/` khusus untuk komunikasi ke perangkat jaringan.
5. **Apakah ini utilitas generik, tidak bergantung ke domain proyek ini sama sekali, genuinely reusable?** → `pkg/`
6. **Tidak cocok satu pun di atas** → JANGAN buat file. Nyatakan penempatan yang diusulkan secara eksplisit sebelum melanjutkan.

**Larangan tegas terkait struktur:**
- **Tidak ada file `.go` langsung di `internal/`** — semua file `.go` harus berada di salah satu subfolder yang sudah didefinisikan di §1.1 (`domain/usecase/port/adapter/driver/registry/audit/config`). Tidak membuat subfolder baru di level ini tanpa menyatakannya eksplisit sebagai perubahan struktur, bukan detail implementasi biasa.
- **Tidak ada file `.go` di root repo** kecuali di dalam `cmd/*/main.go`.
- **Tidak membuat folder `utils/`, `common/`, `helpers/`, `shared/` di mana pun.** Kalau ada logic yang "terasa" perlu ditaruh di sana, itu tanda logic tersebut sebenarnya milik salah satu domain spesifik (`domain/`) atau genuinely milik `pkg/` — cari domain yang tepat, jangan buat folder generik baru.
- **Satu vendor device baru = satu folder baru di `internal/driver/<vendor>/`.** Tidak menambah vendor baru sebagai kode di dalam folder driver vendor lain, dan tidak menambah logic vendor-spesifik di `usecase/` atau `domain/`.
- **Pengetahuan command/protokol milik satu vendor SELALU di `internal/driver/<vendor>/`, tidak pernah di folder terpisah, tidak pernah di `usecase/`/`domain/`.** Dipecah jadi dua file dengan tanggung jawab berbeda: `driver.go` (koneksi + implementasi `port.DeviceDriver`) dan `commands.go` (katalog command native + klasifikasi risiko). `commands.go` **selalu berisi dua fungsi**: `Classify(cmd) command.Class` (command ini destruktif atau tidak, menurut konvensi vendor ini) dan `Translate(op) (command.Command, error)` (operasi abstrak seperti `get_status`/`reboot` diterjemahkan jadi command native vendor ini). Contoh di bawah ini **identik** dengan isi `scaffold.sh` — sudah lolos `go build`+`go vet`+`gofmt`, termasuk compile-time assertion yang membuktikan `*Driver` benar-benar memenuhi `port.DeviceDriver`.

```go
// ✅ BOLEH — internal/driver/mikrotik/driver.go
package mikrotik

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/port"
)

// Driver merepresentasikan satu koneksi yang sudah terbuka ke satu device.
type Driver struct{}

// Bukti compile-time bahwa *Driver benar-benar memenuhi port.DeviceDriver —
// tanpa ini, ketidakcocokan signature bisa tetap lolos `go build` selama
// tidak ada kode lain yang menugaskan *Driver ke variabel interface.
var _ port.DeviceDriver = (*Driver)(nil)

// NewDriver connect ke target dan mengembalikan Driver yang sudah siap.
func NewDriver(ctx context.Context, target device.Target) (*Driver, error) {
	return &Driver{}, nil // TODO: routeros.DialContext(...)
}

func (d *Driver) Execute(ctx context.Context, cmd command.Command) (command.Result, error) {
	return command.Result{}, nil
}

func (d *Driver) Classify(cmd command.Command) command.Class {
	return Classify(cmd) // delegasi ke commands.go
}

func (d *Driver) Translate(op command.Operation) (command.Command, error) {
	return Translate(op) // delegasi ke commands.go
}

func (d *Driver) Close() error {
	return nil
}
```

```go
// ✅ BOLEH — internal/driver/mikrotik/commands.go
// Pengetahuan RouterOS API hidup DI SINI, tidak di usecase/ atau domain/.
package mikrotik

import (
	"fmt"

	"github.com/quixiq/polyglot/internal/domain/command"
)

// Path API yang dianggap destruktif — wajib HITL approval.
var destructivePaths = map[string]bool{
	"/system/reboot":              true,
	"/system/reset-configuration": true,
}

// Katalog: operasi abstrak -> command native RouterOS.
var operationMap = map[command.Operation]command.Command{
	command.OpGetStatus: {Raw: "/system/resource/print"},
	command.OpReboot:    {Raw: "/system/reboot"},
}

func Classify(cmd command.Command) command.Class {
	if destructivePaths[cmd.Raw] {
		return command.ClassDestructive
	}
	return command.ClassReadOnly
}

func Translate(op command.Operation) (command.Command, error) {
	cmd, ok := operationMap[op]
	if !ok {
		return command.Command{}, fmt.Errorf("mikrotik: unsupported operation %q", op)
	}
	return cmd, nil
}
```

Vendor lain (Cisco, dst) mengikuti bentuk identik — cuma isi `destructivePaths`/`operationMap` dan cara `Execute` bicara ke device yang beda. Satu pengecualian sengaja: `internal/driver/genericssh/commands.go` (vendor belum dikurasi) — `Classify` **selalu** mengembalikan `command.ClassDestructive` dan `Translate` **selalu** error, karena risiko vendor yang belum dikenal harus dianggap berbahaya secara default (fail-safe), bukan diam-diam auto-approve.

```go
// ❌ TIDAK BOLEH — jangan taruh pengetahuan vendor di usecase/, karena
// usecase/network/execute_command.go akan jadi tahu detail SEMUA vendor
// dan berubah setiap kali ada vendor baru ditambah — itu sinyal boundary
// yang salah.
package network

func ExecuteCommand(ctx context.Context, cmd Command) error {
	if cmd.Vendor == "mikrotik" && cmd.Raw == "/system/reboot" { // salah — usecase tidak boleh tahu ini
		return ErrNeedsApproval
	}
	// ...
}
```


### 1.3 Tabel Penempatan Cepat (Lookup Reference)

| Jenis perubahan | Path | Pola nama file | Contoh |
|---|---|---|---|
| Entity domain baru | `internal/domain/<domain>/` | `<domain>.go` | `internal/domain/invoice/invoice.go` |
| Error khusus domain | folder sama dengan entity | `errors.go` | `internal/domain/invoice/errors.go` |
| Use case jaringan baru | `internal/usecase/network/` | `<verb>_<noun>.go` | `internal/usecase/network/reboot_device.go` |
| Use case bisnis baru | `internal/usecase/business/` | `<verb>_<noun>.go` | `internal/usecase/business/issue_invoice.go` |
| Interface/port baru | `internal/port/` | `<noun>.go` atau `<noun>_repository.go` | `internal/port/invoice_repository.go` |
| Handler REST baru | `internal/adapter/http/` | `<resource>_handler.go` | `internal/adapter/http/invoice_handler.go` |
| Middleware HTTP baru | `internal/adapter/http/middleware/` | `<nama>.go` | `internal/adapter/http/middleware/rate_limit.go` |
| Tool MCP baru | `internal/adapter/mcp/` | `tool_<nama_tool_snake_case>.go` | `internal/adapter/mcp/tool_get_config.go` |
| Implementasi repository Postgres | `internal/adapter/postgres/` | `<resource>_repository.go` | `internal/adapter/postgres/invoice_repository.go` |
| Driver vendor baru | `internal/driver/<vendor>/` | `driver.go` + `commands.go` | `internal/driver/huawei/driver.go`, `internal/driver/huawei/commands.go` |
| Pengetahuan command/klasifikasi risiko vendor tertentu | folder driver vendor itu sendiri | `commands.go` | `internal/driver/mikrotik/commands.go` |
| Definisi platform YAML custom | `internal/platformdef/` | `<vendor>_<model>.yaml` | `internal/platformdef/zte_c320.yaml` |
| Migrasi DB baru | `migrations/` | `NNNNNN_<deskripsi_snake_case>.up.sql` + `.down.sql` | `migrations/000007_create_invoices_table.up.sql` |
| Test unit | folder sama dengan file diuji | `<nama_file>_test.go` | `internal/domain/invoice/invoice_test.go` |
| Test integrasi | `test/integration/` | `<area>_test.go` | `test/integration/mikrotik_test.go` |
| Script dev/one-off | `scripts/` | `<nama>.go` atau `.sh` | `scripts/seed.go` |
| ADR (keputusan arsitektur) | `docs/adr/` | `NNNN-<slug-kebab-case>.md` | `docs/adr/0003-pilih-gorm-daripada-sqlc.md` |
| README per-package (jarang, lihat §1.5) | folder package itu sendiri | `README.md` | `internal/platformdef/README.md` |

Kalau sebuah task tidak cocok dengan satu baris pun di tabel ini, gunakan algoritma §1.2 — jangan menebak dari analogi tabel ini.

### 1.4 Penamaan File & Folder

- Folder: huruf kecil, satu kata, tanpa underscore/plural (`device`, bukan `devices` atau `device_management`).
- File `.go`: huruf kecil, underscore sebagai pemisah kata (`device_handler.go`, `push_config.go`).
- File `.go` untuk satu tool MCP: **selalu** prefix `tool_`, nama tool dalam snake_case setelahnya, sesuai nama yang dikirim ke `mcp.Tool{Name: ...}` — nama file dan nama tool string harus identik (`tool_get_device_status.go` ↔ `Name: "get_device_status"`).
- File `.go` untuk satu use case: pola `<verb>_<noun>.go`, verb dan noun sama dengan nama fungsi utama di dalamnya (`get_device_status.go` berisi `func GetDeviceStatus(...)`).
- File migrasi: `NNNNNN` enam digit, zero-padded, sequential, tidak pernah reuse angka meskipun migrasi sebelumnya di-revert. `up`/`down` selalu berpasangan, satu perubahan skema per pasang file.
- File test: selalu `_test.go`, **selalu di folder yang sama** dengan file yang diuji (kecuali test integrasi, lihat §1.3) — tidak pernah di folder `test/` terpisah untuk unit test.
- Satu file = satu concern utama. Kalau sebuah file `.go` melebihi ~400 baris, itu sinyal untuk dipecah menjadi beberapa file dalam folder yang sama, bukan alasan untuk pindah folder.

### 1.5 Dokumentasi — Penempatan dan Penamaan

- **ADR baru** → `docs/adr/NNNN-<slug-kebab-case>.md`. `NNNN` empat digit zero-padded dimulai `0001`, sequential, tidak pernah reuse angka (kalau sebuah ADR digantikan, buat ADR baru dengan angka baru yang menyebut "supersedes 000X", jangan edit/hapus yang lama).
- **README per-package** hanya dibuat kalau logic package tersebut genuinely tidak jelas dari nama file saja (jarang — kebanyakan package tidak butuh ini). Letak: langsung di folder package itu (`internal/platformdef/README.md`), bukan dikumpulkan di `docs/`.
- **Dokumen level-proyek** (arsitektur, tech stack, dokumen ini sendiri) → root repo, penamaan `UPPERCASE-KEBAB-CASE.md` untuk dokumen multi-kata (`TECH-STACK-DAN-PERSIAPAN.md`) atau `UPPERCASE.md` untuk nama tunggal yang sudah konvensi (`README.md`, `CLAUDE.md`).
- **Tidak membuat file dokumen baru di root repo tanpa menautkannya dari `README.md`.** Root yang berisi banyak `.md` yang tidak saling terhubung dilarang — kalau membuat dokumen baru di root, tambahkan link ke dokumen itu di `README.md` pada commit yang sama.
- **Tidak menaruh dokumentasi desain di dalam komentar kode yang panjang.** Kalau penjelasan butuh lebih dari ~5 baris komentar, itu seharusnya jadi ADR atau bagian di `Polyglot-Architecture.md`, dengan komentar kode hanya merujuk ke sana.

---

## 2. Penamaan Identifier (Package, Variabel, Fungsi, Konstanta, Interface)

### 2.1 Package

Huruf kecil, satu kata, tanpa underscore, tanpa plural — sudah dicontohkan di §1.1/§1.4. Nama package **tidak diulang** di nama fungsi/tipe di dalamnya:

```go
// ✅ BOLEH
package device
func New(...) *Device { ... }
// caller: device.New(...)

// ❌ TIDAK BOLEH
package device
func NewDevice(...) *Device { ... }
// caller: device.NewDevice(...) — redundan, "device" disebut dua kali
```

### 2.2 Variabel & Fungsi

- **Exported:** `PascalCase`. **Unexported:** `camelCase`.
- Variabel scope pendek (≤~10 baris, mis. di dalam loop): nama singkat boleh (`i`, `d`). Variabel yang hidup lebih lama atau lintas fungsi: nama deskriptif penuh (`deviceID`, bukan `did`).
- **Akronim konsisten huruf besar/kecil sepenuhnya** — ini aturan Go, ditegakkan `staticcheck`:

```go
// ✅ BOLEH
var deviceID string
var httpClient *http.Client
func FetchURLConfig(apiKey string) (*ConfigAPI, error)

// ❌ TIDAK BOLEH
var deviceId string
var HttpClient *http.Client
func FetchUrlConfig(apiKey string) (*ConfigApi, error)
```

### 2.3 Konstanta

`PascalCase` untuk exported, **tidak pernah** `ALL_CAPS` (gaya C, bukan Go):

```go
// ✅ BOLEH
const MaxRetryAttempts = 3

// ❌ TIDAK BOLEH
const MAX_RETRY_ATTEMPTS = 3
```

### 2.4 Interface

- Interface satu method: nama method + `-er` (`Reader`, `Closer`).
- Interface multi-method (`port.DeviceDriver`): nama benda yang menjelaskan peran.
- **Selalu didefinisikan di `internal/port/`, tidak pernah di `internal/adapter/` atau `internal/driver/`.** Implementasi ada di adapter/driver, kontraknya ada di port.

### 2.5 Receiver

1-2 huruf, konsisten untuk tipe yang sama di seluruh file/package (`func (d *Device) Connect(...)`, selalu `d`, tidak berganti jadi `dev`/`device` di fungsi lain dalam file yang sama). Tidak pernah `self`/`this`.

---

## 3. Struktur Kontrol

### 3.1 For Loop

```go
// ✅ BOLEH — range sebagai default
for _, device := range devices {
    process(device)
}

// ✅ BOLEH — Go 1.22+, iterasi angka murni
for i := range 10 {
    fmt.Println(i)
}

// ❌ TIDAK BOLEH — C-style padahal range sudah cukup
for i := 0; i < len(devices); i++ {
    process(devices[i])
}

// ✅ BOLEH — C-style HANYA kalau genuinely butuh step/arah non-linear
for i := len(buffer) - 1; i >= 0; i -= 2 {
    process(buffer[i])
}
```

Goroutine di dalam loop — **selalu** lewatkan variabel loop sebagai parameter closure secara eksplisit, meskipun sejak Go 1.22 capture per-iterasi sudah otomatis (eksplisit tetap wajib untuk kejelasan):

```go
// ✅ BOLEH
for _, device := range devices {
    go func(d Device) {
        connect(d)
    }(device)
}
```

### 3.2 If / Else / Else-If

**Guard clause / early return wajib, bukan nesting:**

```go
// ✅ BOLEH
func (u *ExecuteCommandUseCase) Run(ctx context.Context, cmd Command) (Result, error) {
    if cmd.DeviceID == "" {
        return Result{}, ErrMissingDeviceID
    }
    device, err := u.repo.FindByID(ctx, cmd.DeviceID)
    if err != nil {
        return Result{}, fmt.Errorf("find device: %w", err)
    }
    if !device.IsReachable() {
        return Result{}, ErrDeviceUnreachable
    }
    return u.driver.Execute(ctx, device, cmd)
}

// ❌ TIDAK BOLEH — nesting berlapis, happy path terkubur
func (u *ExecuteCommandUseCase) Run(ctx context.Context, cmd Command) (Result, error) {
    if cmd.DeviceID != "" {
        device, err := u.repo.FindByID(ctx, cmd.DeviceID)
        if err == nil {
            if device.IsReachable() {
                return u.driver.Execute(ctx, device, cmd)
            } else {
                return Result{}, ErrDeviceUnreachable
            }
        } else {
            return Result{}, err
        }
    } else {
        return Result{}, ErrMissingDeviceID
    }
}
```

**Tidak ada `else` setelah cabang `if` yang sudah `return`/`continue`/`break`/`panic`:**

```go
// ✅ BOLEH
if err != nil {
    return err
}
doNext()

// ❌ TIDAK BOLEH
if err != nil {
    return err
} else {
    doNext()
}
```

**Maksimal 2 `else if` berturutan (3 cabang total). Lebih dari itu, wajib `switch`** — lihat §3.3 untuk contoh konversinya.

### 3.3 Switch

```go
// ✅ BOLEH — tag-less switch mengganti if-else-if boolean panjang
switch {
case cmd.IsReadOnly():
    return policy.AutoApprove
case cmd.RequiresApproval():
    return policy.PendingApproval
default:
    return policy.Deny
}

// ✅ BOLEH — switch pada enum/tipe tertutup, default eksplisit menolak kasus tak dikenal
switch cmd.Type {
case CommandTypeRead:
    return handleRead(cmd)
case CommandTypeWrite:
    return handleWrite(cmd)
default:
    return fmt.Errorf("unknown command type: %v", cmd.Type)
}
```

- Tidak menulis `break` di akhir setiap `case` — Go tidak fallthrough secara default.
- `fallthrough` eksplisit hanya dengan komentar yang menjelaskan alasannya di baris itu.
- Switch pada enum/tipe tertutup **selalu** punya `default` yang menangani kasus tak terduga secara eksplisit — tidak pernah dibiarkan diam-diam tidak menangani apa pun.

### 3.4 Function

**Error selalu di posisi return terakhir:**

```go
// ✅ BOLEH
func NewDriver(ctx context.Context, target device.Target) (*Driver, error)

// ❌ TIDAK BOLEH
func NewDriver(ctx context.Context, target device.Target) (error, *Driver)
```

**Lebih dari 4 parameter → wajib struct:**

> Contoh di bawah pakai `Invoice`/package `billing` (bukan `Device`/package
> `device`) supaya tidak bentrok dengan aturan anti-stutter §2.1 — kalau
> contoh ini ditulis sebagai `device.NewDevice(...)`, itu justru melanggar
> aturan `device.New(...)` yang sudah ditegaskan di atas.

```go
// ❌ TIDAK BOLEH
func NewInvoice(id, customerID, subscriptionID string, amount int, dueDate time.Time, status string) *Invoice

// ✅ BOLEH
type NewInvoiceParams struct {
    ID             string
    CustomerID     string
    SubscriptionID string
    Amount         int
    DueDate        time.Time
    Status         string
}
func NewInvoice(p NewInvoiceParams) *Invoice
```

**Naked return hanya untuk fungsi ≤5 baris.** Di luar itu, tulis eksplisit — fungsi panjang dengan naked return memaksa pembaca scroll ke atas untuk tahu apa yang di-return.

**Nama fungsi pola `VerbNoun`, konsisten dengan nama file (§1.4).** Satu fungsi, satu tanggung jawab — kalau nama butuh kata "And" (`FetchAndValidateAndSave`), pecah jadi beberapa fungsi.

---

## 4. Error Handling

```go
// ✅ BOLEH — cek langsung, wrap dengan %w dan konteks operasi
result, err := repo.FindByID(ctx, id)
if err != nil {
    return fmt.Errorf("find device %s: %w", id, err)
}

// ❌ TIDAK BOLEH — error dibuang diam-diam
result, _ := repo.FindByID(ctx, id)

// ✅ BOLEH — kalau genuinely diabaikan, WAJIB komentar alasan
_ = conn.Close() // best-effort cleanup, kegagalan close tidak mengubah hasil operasi

// ❌ TIDAK BOLEH — %v memutus error chain, errors.Is di caller tidak akan bekerja
return fmt.Errorf("connect to device %s: %v", deviceID, err)
```

**`errors.Is` untuk sentinel error, `errors.As` untuk custom error type. Tidak pernah bandingkan string error:**

```go
// ✅ BOLEH
if errors.Is(err, sql.ErrNoRows) {
    return nil, ErrDeviceNotFound
}
var validationErr *ValidationError
if errors.As(err, &validationErr) {
    return validationErr.Fields
}

// ❌ TIDAK BOLEH
if err.Error() == "sql: no rows in result set" {
    return nil, ErrDeviceNotFound
}
```

- **Sentinel error** (`var ErrDeviceNotFound = errors.New(...)`) untuk kondisi tanpa data tambahan. **Custom error type** kalau caller butuh data tambahan dari error tersebut.
- **`panic` hanya di `main()`/startup init** untuk kegagalan yang genuinely tidak bisa dilanjutkan (config wajib tidak ada). Kode di `usecase/`, `adapter/`, `driver/` **selalu** return `error`.

---

## 5. Concurrency

```go
// ✅ BOLEH — context.Context selalu parameter pertama, tidak pernah field struct
func (d *Driver) Execute(ctx context.Context, cmd command.Command) (command.Result, error)

// ❌ TIDAK BOLEH
type Driver struct {
    ctx context.Context // JANGAN — context punya lifecycle per-request
}
```

```go
// ✅ BOLEH — errgroup, semua goroutine ditunggu, error pertama di-propagate
g, ctx := errgroup.WithContext(ctx)
for _, device := range devices {
    device := device
    g.Go(func() error {
        return poll(ctx, device)
    })
}
if err := g.Wait(); err != nil {
    return err
}

// ❌ TIDAK BOLEH — fire-and-forget, tidak ditunggu, error hilang, context.Background() alih-alih ctx yang diteruskan
for _, device := range devices {
    go poll(context.Background(), device)
}
```

- Channel: arah tipe eksplisit di signature (`chan<- OutputChunk`, bukan `chan OutputChunk` kalau cuma dipakai satu arah).
- Mutex sebagai field (`mu sync.RWMutex`), **tidak pernah embed** kecuali struct itu sendiri adalah "sebuah lock" — embed membocorkan `Lock()`/`Unlock()` sebagai method exported yang tidak seharusnya diakses caller.

---

## 6. Tipe & Struktur Data

- **Tidak boleh `interface{}`/`any`** kecuali generic code (type parameter) atau boundary deserialisasi JSON yang genuinely dinamis. `map[string]interface{}` adalah sinyal ada tipe yang belum didefinisikan — definisikan structnya di `domain/` yang sesuai.
- Struct tag diformat `gofumpt` otomatis, tidak manual align.
- Zero value harus valid sebisa mungkin. Kalau field butuh inisialisasi non-trivial (map, channel), sediakan `New()` — jangan biarkan caller menebak field mana yang wajib di-set.

---

## 7. Komentar & GoDoc

```go
// ✅ BOLEH — dimulai dengan nama identifier, menjelaskan mengapa/peran, bukan menarasikan baris kode
// Registry manages device connections and dispatches commands
// to the correct DeviceDriver implementation based on vendor type.
type Registry struct { ... }

// ❌ TIDAK BOLEH — tidak dimulai dengan nama identifier
// This struct manages devices.
type Registry struct { ... }
```

- Setiap identifier exported wajib punya doc comment dimulai dengan nama identifier itu sendiri — ditegakkan `staticcheck`.
- Tidak ada komentar yang menarasikan kode baris-per-baris (`// loop through devices` di atas `for _, d := range devices`).
- Komentar menjelaskan "mengapa" (keputusan non-obvious), bukan "apa" (kode sudah bilang apa).
- Penjelasan lebih dari ~5 baris → pindah ke ADR (`docs/adr/`), komentar kode cukup merujuk nomor ADR-nya.

---

## 8. Import

```go
import (
    // 1. Standard library
    "context"
    "fmt"

    // 2. Third-party
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"

    // 3. Internal project
    "github.com/quixiq/polyglot/internal/domain/device"
    "github.com/quixiq/polyglot/internal/port"
)
```

Tiga grup dipisah baris kosong, di-enforce `goimports`/`gofumpt`. Tidak ada `import .` kecuali di file test dengan alasan eksplisit.

---

## 9. Testing

```go
// ✅ BOLEH — table-driven, default untuk fungsi dengan banyak skenario
func TestValidateCommand(t *testing.T) {
    tests := []struct {
        name    string
        cmd     Command
        wantErr error
    }{
        {"empty device id", Command{DeviceID: ""}, ErrMissingDeviceID},
        {"valid read command", Command{DeviceID: "mtk-1", Type: "read"}, nil},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateCommand(tt.cmd)
            if tt.wantErr != nil {
                require.ErrorIs(t, err, tt.wantErr)
                return
            }
            require.NoError(t, err)
        })
    }
}
```

- Nama test: `TestFungsi` + subtest `t.Run` deskriptif seperti contoh di atas.
- `require` menghentikan test saat gagal (precondition/setup). `assert` melanjutkan dan mengumpulkan kegagalan (beberapa pengecekan independen dalam satu test).
- Mocking hanya untuk yang genuinely di luar kendali (external HTTP API, device fisik tak tersedia di CI). Postgres/Redis: `testcontainers-go` dengan instance asli. Driver Mikrotik: uji ke Mikrotik CHR di GNS3, bukan mock `DeviceDriver`.

---

## 10. Ringkasan — Larangan & Kewajiban Mutlak

**Tidak boleh, tanpa pengecualian:**
- File `.go` langsung di `internal/` atau di root repo (di luar `cmd/*/main.go`)
- Folder `utils/`, `common/`, `helpers/`, `shared/`
- `interface{}`/`any` di luar generic code atau boundary JSON dinamis
- `panic` di luar `main()`/startup init
- Membandingkan error dengan `err.Error() == "..."`
- `%v` untuk wrap error (harus `%w`)
- Mengabaikan error dengan `_` tanpa komentar alasan
- `else` setelah cabang `if` yang sudah `return`/`continue`/`break`/`panic`
- Lebih dari 2 `else if` berturutan
- Fungsi dengan lebih dari 4 parameter posisional
- `context.Context` sebagai field struct
- Goroutine tanpa mekanisme tunggu/pembatalan yang jelas
- Import dari `domain/` ke `adapter/`/`driver/`/framework eksternal
- Akronim tidak konsisten kapitalisasi (`Id`, `Http`, `Url`, `Api`)
- Naked return di fungsi lebih dari ~5 baris
- Membuat file/folder baru tanpa mengikuti §1.2/§1.3

**Wajib, tanpa pengecualian:**
- Cek §1 (path & nama file) sebelum membuat file apa pun
- Setiap identifier exported punya doc comment dimulai dengan namanya sendiri
- Error selalu jadi return value terakhir
- `context.Context` sebagai parameter pertama untuk fungsi yang melakukan I/O
- Wrap error dengan `%w` dan konteks operasi yang gagal
- Table-driven test untuk fungsi dengan banyak skenario
- Komentar `// DEVIASI: <alasan>` untuk setiap penyimpangan dari dokumen ini
- Dokumen baru (ADR/README) ditautkan dari `README.md` root pada commit yang sama