# Plan: AI Router Ops — Backup, Restore Point, Approval, MCP `/console/inspect`

> **Status implementasi (31 Okt 2026):** draft ini telah di-replan dan
> diimplementasikan sebagai **PLAN-AI-ROUTER-OPS v2**.
> Lihat [`PLAN-AI-ROUTER-OPS-V2-IMPL.md`](PLAN-AI-ROUTER-OPS-V2-IMPL.md) untuk
> record implementasi + deviasi (terutama: validasi `/console/inspect` via
> goros sudah ada, MCP `/mcp` kini JWT-protected, Mikhmon mutating tools tidak
> diekspos ke AI v1, watchdog default OFF).

**Status:** Draft untuk direview
**Tanggal:** 31 Oktober 2026
**Scope:** Polyglot backend + MCP server + driver MikroTik + **UI Polyglot sendiri** (React + shadcn/ui + shadcn AI Elements).
**Target utama:** RouterOS v7 (fitur AI maksimal), v6 degraded mode
**Referensi UI/inspect:** `mikrotik-inspect/main.go`, `mikrotik-inspect/index.html`
**UI stack:** Vite + React 19 + TanStack Router/Query + Tailwind v4 + shadcn `new-york` (Radix, lucide-react) — **bukan LibreChat, bukan AI SDK Elements package**. AI Elements di-copy sebagai pattern ke `web/src/components/ai-elements/`.

---

## 0. Keputusan Desain yang Dikunci

| # | Keputusan | Alasan |
|---|---|---|
| D1 | **Hanya 2 mode command:** `READ` dan `MUTATION` | Tidak perlu klasifikasi risiko bertingkat. Semua yang mengubah state = approval. |
| D2 | **Default unknown = `MUTATION`** (fail-closed) | Kalau classifier tidak yakin, jangan eksekusi otomatis. |
| D3 | **Semua `MUTATION` wajib:** diff + restore point + approval manusia | Backup/restore point adalah syarat mutlak sebelum approval dibuat. |
| D4 | **READ bebas dieksekusi AI** | AI butuh kebebasan inspeksi agar tidak halusinasi dan bisa troubleshooting. |
| D5 | **AI tidak boleh approve request-nya sendiri** | `approved_by` wajib user manusia dengan permission `approval:decide`. |
| D6 | **Target device = router yang sedang aktif**, bukan `.env` | Referensi `mikrotik-inspect` memakai satu router via env; Polyglot harus memakai registry + vault + active-device binding. |
| D7 | **`/console/inspect` hanya RouterOS v7** | v6 tidak punya command ini. AI tetap bisa read/mutation di v6, tapi validasi sintaks & dynamic form disabled (degraded). |
| D8 | **Tidak ada eksekusi langsung dari MCP** | `run_command`/`push_config` lama harus diubah: READ eksekusi, MUTATION hanya membuat proposal. |
| D9 | **Restore point disimpan di Polyglot, bukan hanya di router** | File router bisa terhapus/reboot. Polyglot menyimpan artifact terenkripsi + metadata + hash. |
| D10 | **Rollback utama = inverse script + watchdog scheduler** | RouterOS API tidak punya "commit confirmed". Safe Mode hanya console, bukan API. |
| D11 | **Binary backup setiap mutation + backup terjadwal** | Setiap mutation membuat binary backup; selain itu ada backup harian/mingguan otomatis. Backup lama dihapus otomatis (retensi + keep-last). |
| D12 | **UI = Polyglot sendiri, shadcn + AI Elements** | Tidak memakai LibreChat/UI eksternal. Semua halaman AI Ops dibangun dari komponen shadcn yang sudah ada di `web/`, ditambah pattern AI Elements (`chat`, `confirmation`, `checkpoint`, `queue`, `tool`, `terminal`, `model-selector`, `prompt-input`, dll). |
| D13 | **Confirmation di chat bukan otoritas approval** | Tombol Approve/Deny di UI hanya memanggil `ApprovalService` dengan JWT user. Server tetap memverifikasi permission, revision, expiry, dan restore point. AI/MCP tidak pernah bisa approve. |
| D14 | **Chain of Thought = trace langkah agent, bukan hidden reasoning** | UI menampilkan langkah nyata (inspect → snapshot → backup → diff → approval → execute → verify). Raw internal reasoning tidak ditampilkan kecuali provider mengirimkannya secara eksplisit. |
| D15 | **AI tidak boleh mempropose `reset-configuration`** | Reset menghapus seluruh konfigurasi dan rollback-nya tidak deterministik. Hanya manusia via UI/terminal dengan typed confirmation. |

---

## 1. Baseline Saat Ini (Temuan di Kode)

### 1.1 Yang sudah ada (bisa dipakai)

- `internal/driver/mikrotik/system/backup.go`
  - `NewBackupSaveCommand`, `NewBackupLoadCommand`, `NewExportCommand`, `Settled`, `Sweep`.
- `internal/driver/mikrotik/system/file.go`
  - `NewPrintFilesCommand`, `NewPrintFileCommand`, `NewReadFileChunkCommand`, `NewRemoveFileCommand`, `MaskSecrets`, `ParseFileEntries`.
- `internal/registry/registry.go`
  - `Get(ctx, deviceID)` → resolve device dari DB + vault + factory, cache koneksi. Ini jalur yang benar untuk "router aktif".
- `internal/adapter/mcp/*`
  - MCP streamable HTTP, beberapa tool, `auditBestEffort`.
- `internal/usecase/network/execute_command.go`
  - Sudah ada policy gate sederhana (`Classify` → `Decide` → execute).
- `internal/adapter/ws/device_stream_handler.go`
  - WebSocket PTY terminal (bisa jadi emergency takeover).
- `web/src/components/layout/device-switcher.tsx` + `web/src/stores/device-store.ts`
  - Sudah ada pemilih router global (Zustand + persist) yang dipakai banyak fitur. Ini bisa jadi basis UI "router aktif", tetapi masih client-only dan auto-fallback ke `devices[0]`.
- `web/` sudah memakai shadcn/ui
  - `components.json`: `new-york`, Radix, Tailwind v4, alias `@/components/ui`, ikon `lucide-react`.
  - Sudah terpasang: button, card, dialog, sheet, tabs, select, command, table, badge, alert, collapsible, scroll-area, dll.
- `internal/adapter/connect/llm` + `api/proto/v1/llm.proto`
  - CRUD konfigurasi LLM (provider, model, API key, system prompt, skills). Bisa dipakai untuk Model Selector.
- `internal/adapter/llm/genkit/provider.go`
  - `ChatWithTools` sudah mendukung native tool calling (maks 5 turn).
- `web/src/features/whatsapp/api/use-whatsapp-sse.ts`
  - Pola streaming EventSource + reconnect yang sudah terbukti; bisa jadi referensi, tetapi chat AI Ops sebaiknya memakai ConnectRPC server-streaming.
- `docs/database-schema.md` Lampiran A.3/A.4
  - Draft `command_logs` dan `approval_requests` (belum ada migration).

### 1.2 Gap kritis (harus diperbaiki lebih dulu)

| Gap | Lokasi | Dampak |
|---|---|---|
| Classifier hanya menandai 2 path destructive | `internal/driver/mikrotik/commands.go:16-19` | `/ip/firewall/filter/remove`, `/ip/address/remove`, `/interface/disable` dianggap READ dan dieksekusi tanpa approval. |
| `Class` hanya ReadOnly/Destructive | `internal/domain/command/command.go:47-51` | Tidak ada konsep "mutation" yang konsisten; `add`/`set` tidak terdeteksi. |
| `Decide` hanya cek destructive | `internal/domain/command/policy.go:19-24` | Mutasi non-destructive lolos. |
| MCP `run_command` memakai `ExecuteCommandPreApproved` | `internal/adapter/mcp/tool_run_command.go:44` | AI bisa eksekusi mutation tanpa approval server-side; HITL hanya di client (bisa bug/malicious). |
| MCP `push_config` juga `PreApproved` | `internal/adapter/mcp/tool_push_config.go:40` | Sama. |
| MCP endpoint publik & audit nil | `internal/app/router.go:167-170` (`mcp.New(d.reg, nil)`, `rootMux.Handle("/mcp", ...)`) | Tidak ada autentikasi user, tidak ada audit trail. Sangat berbahaya untuk command execution. |
| Tidak ada tabel approval/restore point | `migrations/` (hanya draft di docs) | Tidak ada state machine, expiry, replay protection. |
| Backup driver belum dipakai usecase/MCP | `system/backup.go` | Belum ada restore point otomatis sebelum mutasi. |
| Active device masih client-only | `web/src/stores/device-store.ts`, `device-switcher.tsx` | AI/MCP tidak tahu router yang dipilih; switcher auto-fallback ke `devices[0]`. |
| Belum ada chat AI Ops untuk operator | `web/src/features/chats/*` | `chats` yang ada adalah inbox WhatsApp pelanggan, bukan chat operator→router. |
| Belum ada streaming chat AI | `internal/adapter/llm/genkit/provider.go` | Hanya `ChatWithTools` non-streaming; UI AI Elements butuh delta text/tool/terminal. |
| Belum ada UI approval/restore point | `web/src/features/` | Approval hanya bisa lewat API/DB; belum ada Confirmation, Queue, diff viewer, restore point timeline. |
| Primitif AI Elements belum terpasang | `web/src/components/ui/` | Belum ada `message-scroller`, `message`, `bubble`, `trace`, `marker`, `code-block`, `input-group`, `item`. |
| Belum ada route/nav AI Ops | `web/src/routes/_authenticated/`, `sidebar-data.ts` | Tidak ada entry menu, permission, atau halaman untuk AI Router Ops. |
| `/console/inspect` belum ada di driver | `internal/driver/mikrotik/` | AI tidak bisa validasi sintaks/argumen sebelum eksekusi. |

### 1.3 Referensi `mikrotik-inspect/`

Yang bisa diadopsi:

- Pola API `/console/inspect`:
  - `request=child`, `path=ip,address`
  - `request=syntax`, `path=ip,address,add,address`
  - `request=completion`, `input="/ip/addr/set "`, `path=ip,address`
  - `.proplist=name,node-type,type,symbol,symbol-type,text`
- Validasi path: regex `^[a-zA-Z0-9_./,-]+$` + tolak `..`.
- Form generation: `request=child path=<menu>,add`, filter `node-type=arg`.
- Menu read: `/<path>/print`.
- Health: `/system/resource/print` (`version`, `board-name`, `cpu-load`, `free-memory`, `uptime`).
- Index HTML: tree navigasi, tabel dinamis, field form, polling monitor interface.

Yang **tidak** boleh diadopsi:

- Kredensial dari `.env` (`ROUTER_ADDRESS`, `ROUTER_USERNAME`, `ROUTER_PASSWORD`).
- Satu router global.
- Tanpa auth, tanpa audit, tanpa approval.
- Hanya `child/syntax/completion`; kita perlu `highlight` dan `error` untuk validasi perintah.
- `connect()` per request tanpa registry/cache (Polyglot sudah punya registry).

---

## 2. Arsitektur Target

```
┌──────────────────────────────────────────────────────────────────────────────┐
│              Polyglot Web UI (React + shadcn + AI Elements)                  │
│  Chat ── MessageScroller/Message/Bubble/PromptInput/ModelSelector            │
│  Activity ── Trace (Tool, Terminal, Plan, Queue, Chain of Thought)           │
│  Approval ── Confirmation/Item/CodeBlock diff        Checkpoint ── Marker    │
│  Device Picker ──► SetActiveDevice   Restore Point list/timeline             │
│  Warning v6 ── Alert                 Emergency console ── Xterm (WS PTY)     │
└───────────────┬───────────────────────────────────────────────┬──────────────┘
                │ ConnectRPC (JWT human)                        │ MCP (OAuth/OBO)
                ▼                                               ▼
┌──────────────────────────────────────────────────────────────────────────────┐
│                          Polyglot AI Router Ops Core                         │
│                                                                              │
│  AIChatService (server-streaming) ──► LLM provider + tool registry           │
│  ActiveDeviceResolver ──► Registry ──► MikroTik Driver                       │
│                                                                              │
│  CommandClassifier (READ | MUTATION)                                         │
│         │ READ                          │ MUTATION                           │
│         ▼                               ▼                                    │
│  ExecuteRead                     ProposeMutation                             │
│  (auto, audit)                   ├─ InspectPreflight (v7)                    │
│                                  ├─ SnapshotBefore (read)                    │
│                                  ├─ BuildDiff                                │
│                                  ├─ CreateRestorePoint                       │
│                                  └─ Approval PENDING                         │
│                                                                              │
│  ApprovalService (approve / reject / edit)                                   │
│         │                                                                    │
│         ▼                                                                    │
│  ExecuteApproved ──► WatchdogScheduler ──► Driver.Execute                    │
│         │                                                                    │
│         ├─ PostVerify (read)                                                 │
│         ├─ CancelWatchdog                                                    │
│         └─ Audit + Result                                                    │
│                                                                              │
│  RestorePointService (list / restore / cleanup)                              │
└───────────────┬───────────────────────────────────────────────┬──────────────┘
                │                                               │
                ▼                                               ▼
       PostgreSQL (approval, restore_points, command_logs, ai_chat)  Encrypted Storage
                                                              data/restorepoints/
```

### 2.1 Alur singkat

```
AI/Operator: "blokir port 23 di firewall"
  → READ: /ip/firewall/filter/print (bebas)
  → v7: /console/inspect highlight input="..." (bebas, validasi)
  → classifier: verb "add" = MUTATION
  → READ snapshot: /ip/firewall/filter/print detail
  → build diff: + rule baru
  → create restore point: scoped export + binary backup + inverse script + download
  → approval PENDING (15 menit)
  → UI: [Approve] [Reject] [Edit]
  → Approve:
      → watchdog scheduler (auto-rollback 10 menit)
      → execute
      → post-verify read
      → cancel watchdog
      → audit SUCCESS
```

---

## 3. Model Command: READ vs MUTATION

### 3.1 Domain

Ubah `internal/domain/command/command.go`:

```go
type Mode int

const (
    // ModeRead covers any command that only reads state or introspects
    // syntax. Safe to auto-execute for AI and humans.
    ModeRead Mode = iota
    // ModeMutation covers any command that can create, change, delete,
    // enable/disable, reset, reboot, upload, or otherwise alter router state.
    ModeMutation
)
```

Hapus `ClassReadOnly` / `ClassDestructive`; ganti seluruh pemakaian dengan `Mode`.
`Class` tidak dipakai di protobuf/wire, jadi rename aman.

`internal/domain/command/policy.go`:

```go
func Decide(mode Mode) Decision {
    if mode == ModeMutation {
        return DecisionRequireApproval
    }
    return DecisionAutoApprove
}
```

`DecisionDeny` tetap ada untuk hard-deny list (mis. `reset-configuration` dari AI).

### 3.2 Classifier MikroTik

File baru: `internal/driver/mikrotik/classify.go`

Aturan:

1. Ambil verb dari elemen terakhir path:
   - `/ip/firewall/filter/add` → `add`
   - `/ip/address/print` → `print`
   - `/export` → `export`
   - `/system/reboot` → `reboot`
2. **Read allowlist** (verb):
   ```
   print, get, export, monitor, monitor-traffic, ping, traceroute,
   check, inspect, sniff, torch, find, discover, scan, whois, fetch-check
   ```
   Catatan:
   - `export` = READ (hanya menghasilkan file/teks).
   - `print` dengan `follow`/`interval` = streaming READ.
   - `/console/inspect` = READ.
3. **Mutation verbs** (eksplisit):
   ```
   add, set, remove, unset, enable, disable, move, reset, reboot,
   shutdown, upgrade, downgrade, load, save, import, fetch,
   add-file, set-file, remove-file, format, restore, backup
   ```
4. **Default = MUTATION** jika verb tidak dikenal, path kosong, atau ada argumen yang menandakan penulisan (`contents=`, `file=`, `name=` pada `/system/backup`).
5. **Hard deny untuk AI** (bukan untuk manusia via UI):
   - `/system/reset-configuration`
   - `/system/backup/load` (restore penuh) — hanya lewat alur restore point + typed confirmation
   - `/file/remove` pada file sistem
   - `/tool/fetch` ke URL eksternal (SSRF) — hanya URL yang diizinkan
   - `/system/script/add` + `/system/scheduler/add` yang dibuat AI mentah (watchdog dibuat sistem, bukan AI)
6. **Path mutation yang sering salah**:
   - `/ip/firewall/filter/remove` → MUTATION
   - `/interface/disable` → MUTATION
   - `/ip/route/set` → MUTATION
   - `/system/identity/set` → MUTATION

Test tabel wajib: minimal 50 kasus, termasuk unknown verb → MUTATION.

### 3.3 Kenapa bukan klasifikasi 5 tingkat?

Karena `add` satu baris tetap bisa mengunci router:

```routeros
/ip firewall filter add chain=input action=drop
```

Membuat perbedaan "Additive = aman" adalah ilusi. Model biner READ/MUTATION lebih jujur dan lebih aman.

---

## 4. `/console/inspect` (RouterOS v7)

### 4.1 Driver baru

Direktori: `internal/driver/mikrotik/inspect/`

```go
type Client struct {
    exec port.CommandExecutor // atau *mikrotik.Driver
}

type Node struct {
    Type     string `json:"type"`
    Name     string `json:"name"`
    NodeType string `json:"node-type"`
}

type Syntax struct {
    Type        string `json:"type"`
    Symbol      string `json:"symbol"`
    SymbolType  string `json:"symbol-type"`
    Nested      string `json:"nested"`
    Nonorm      string `json:"nonorm"`
    Text        string `json:"text"`
}

type Completion struct {
    Type       string `json:"type"`
    Completion string `json:"completion"`
    Style      string `json:"style"`
    Offset     int    `json:"offset"`
    Preference int    `json:"preference"`
    Show       bool   `json:"show"`
    Text       string `json:"text"`
}

type Highlight struct {
    Type  string `json:"type"`
    Text  string `json:"text"`
    Start int    `json:"start"`
    End   int    `json:"end"`
}
```

Method:

```go
func (c *Client) Children(ctx context.Context, path string) ([]Node, error)
func (c *Client) Syntax(ctx context.Context, path string) ([]Syntax, error)
func (c *Client) Completion(ctx context.Context, input, path string) ([]Completion, error)
func (c *Client) Highlight(ctx context.Context, input string) ([]Highlight, error)
func (c *Client) Validate(ctx context.Context, script string) (ValidationResult, error)
```

`Validate`:
- Panggil `request=highlight input=<script>`.
- Jika ada token `type=error` → invalid, kembalikan offset + teks.
- Opsional `request=completion` untuk saran perbaikan.
- **Bukan** jaminan runtime; hanya validasi sintaks/argumen.

### 4.2 Command builder

```go
func NewInspectCommand(request, path, input, proplist string) command.Command
```

Contoh args (API):
```
/console/inspect =request=child =path=ip,address =.proplist=name,node-type,type
/console/inspect =request=syntax =path=ip,address,add,address
/console/inspect =request=completion =input="/ip/addr/set " =path=ip,address
/console/inspect =request=highlight =input="/ip/address/add address=1.1.1.1/24"
```

### 4.3 Version gate

Sebelum memanggil `/console/inspect`, cek:

```go
func (c *Client) Capabilities(ctx context.Context) (Capabilities, error)
```

Ambil `/system/resource/print` → `version` (contoh `7.16.2 (stable)`).
- Major ≥ 7 → `InspectSupported=true`, `AIMode="full"`.
- Major < 7 → `InspectSupported=false`, `AIMode="degraded"`.
- Cache per device selama TTL 10 menit (di registry wrapper atau inspect client).
- Jangan panggil `/console/inspect` di v6 (path tidak ada).

### 4.4 Known issue & pengamanan

- `request=syntax`/`completion` pada path scripting `do` bisa membuat REST server deadlock di RouterOS ≤ 7.20.8; fixed 7.21.4.
- Kebijakan konservatif: skip `syntax`/`completion` untuk path yang mengandung `where, do, else, rule, command, on-error` pada versi < 7.21.4.
- `request=child` aman untuk semua versi yang diuji.
- Batasi panjang `input` (mis. 8 KB) dan jumlah request per menit per device.
- Timeout 5 detik per inspect call; jangan block approval flow terlalu lama.

### 4.5 Peringatan v7 (wajib muncul di 4 tempat)

1. **MCP server instructions** (dibaca LLM):
   > "RouterOS AI features are optimized for RouterOS v7. `/console/inspect` is v7-only. On v6, syntax validation is disabled (degraded mode); every mutation still requires human approval and a restore point."
2. **Tool `router_capabilities`** → `ai_mode: "full" | "degraded"`, `warning`.
3. **UI Polyglot** (shadcn):
   - `Alert` di header chat AI Ops: "AI RouterOS Ops optimal di RouterOS v7. Validasi `/console/inspect` tidak tersedia di v6."
   - `Badge variant="secondary"` / `destructive` di device chip: `v6 degraded`.
   - `Alert` di dialog approval: "Sintaks belum diverifikasi router (v6). Periksa manual sebelum approve."
   - `Tooltip` di `ModelSelector`/capability chip yang menjelaskan degraded mode.
4. **Docs**: README, `Polyglot-Architecture.md`, ADR baru.

### 4.6 Degradasi v6

| Fitur | v7 | v6 |
|---|---|---|
| Read/inspect manual | ✅ | ✅ |
| `/console/inspect` child/syntax/completion/highlight | ✅ | ❌ |
| Validasi sintaks sebelum approval | ✅ | ❌ (fallback: static validator) |
| Dynamic form dari inspect | ✅ | ❌ |
| Mutation + diff + restore point + approval | ✅ | ✅ |
| AI mode | `full` | `degraded` |

Fallback v6:
- Gunakan validator statis (allowlist verb/argumen umum).
- Diff tetap dibuat dari snapshot `print`.
- UI menampilkan `Alert` + `Badge` warning bahwa sintaks belum diverifikasi router.
- Admin bisa set `AI_V6_MUTATIONS=false` untuk memblokir mutation AI di v6.

---

## 5. P0-A: Backup Data & Restore Point

Ini prioritas **paling utama**. Approval tidak boleh dibuat jika restore point gagal.

### 5.1 Domain

File baru: `internal/domain/restorepoint/restore_point.go`

```go
type ArtifactKind string

const (
    ArtifactScopedExport  ArtifactKind = "SCOPED_EXPORT"
    ArtifactBinaryBackup  ArtifactKind = "BINARY_BACKUP"
    ArtifactInverseScript ArtifactKind = "INVERSE_SCRIPT"
)

type Status string

const (
    StatusCreating  Status = "CREATING"
    StatusAvailable Status = "AVAILABLE"
    StatusPartial   Status = "PARTIAL"   // sebagian artifact gagal dibuat/diunduh
    StatusUsed      Status = "USED"
    StatusExpired   Status = "EXPIRED"
    StatusDeleted   Status = "DELETED"
)

// Artifact adalah satu file di dalam bundle restore point.
type Artifact struct {
    Kind        ArtifactKind
    Scope       []string          // untuk SCOPED_EXPORT, mis. ["/ip/firewall"]
    StoragePath string            // relative path di RESTORE_POINT_DIR
    SHA256      string
    SizeBytes   int64
}

// RestorePoint adalah bundle: scoped export + inverse script + binary backup.
type RestorePoint struct {
    ID              string
    TenantID        string
    DeviceID        string
    Trigger         string            // MUTATION | SCHEDULED | MANUAL
    Scope           []string          // gabungan menu yang di-backup
    RouterOSVersion string
    ConfigHash      string            // untuk dedup binary backup
    Artifacts       []Artifact
    RollbackScript  []command.Command // inverse commands
    Manifest        Manifest
    Status          Status
    TotalSizeBytes  int64
    CreatedBy       string            // user/session
    CreatedAt       time.Time
    ExpiresAt       time.Time
    UsedAt          *time.Time
}
```

Error sentinel:
```go
var (
    ErrNotFound          = fault.New(fault.KindNotFound, "restorepoint: not found")
    ErrNotAvailable      = fault.New(fault.KindFailedPrecondition, "restorepoint: not available")
    ErrExpired           = fault.New(fault.KindFailedPrecondition, "restorepoint: expired")
    ErrCreationFailed    = fault.New(fault.KindInternal, "restorepoint: creation failed")
)
```

### 5.2 Apa yang di-backup

**Default (D11): setiap mutation membuat scoped export + inverse script + binary backup.**
Binary backup juga dijadwalkan harian/mingguan. Untuk menghindari penumpukan:

- **Dedup:** hitung hash konfigurasi. Jika sama dengan binary backup terakhir yang sukses, **reuse** artifact itu (hardlink lokal / object reference S3) dan tetap catat entry `BINARY_BACKUP` di bundle; scoped export + inverse tetap dibuat. Jangan menghilangkan entry binary dari bundle.
- **Keep-last + age:** simpan `N` backup terakhir (default 30) dan/atau `RESTORE_POINT_RETENTION_DAYS` (default 14 hari).
- **Auto-delete:** job harian menghapus file + row lama (status `DELETED`, metadata audit tetap).
- **Ukuran:** batas `RESTORE_POINT_MAX_BYTES`; jika binary backup melebihi batas/free space, mutation diblokir sampai admin membersihkan storage atau menaikkan batas.

| Situasi | Backup |
|---|---|
| Mutation apa pun (default) | Scoped export menu terkait + inverse script + **binary backup** |
| Mutation menyentuh >1 menu | Scoped export tiap menu + inverse script + binary backup |
| Mutation pada path kritis (`/system`, `/ip/service`, `/interface`, `/ip/firewall`, `/ip/route`) | Scoped export + binary backup (wajib, tidak boleh di-skip dedup) |
| `reset-configuration`, `backup/load`, upgrade/downgrade | Binary backup full + typed confirmation |
| Backup terjadwal | Binary backup harian + full export; retensi otomatis |
| AI minta `reboot` | Binary backup + scoped export tetap dibuat (audit/recovery); inverse tidak relevan; wajib approval + cek uptime + post-verify |

Scoped export:
```routeros
/ip/firewall/export file=polyglot-rp-<id> compact terse hide-sensitive
```
Catatan: `hide-sensitive` untuk artifact yang disimpan; jika perlu restore credential, gunakan export terpisah dengan `show-sensitive` yang dienkripsi ketat—default jangan.

Binary backup:
```routeros
/system/backup/save name=polyglot-rp-<id> dont-encrypt=yes
```
Alasan `dont-encrypt`: password backup harus dikelola; file hasilnya dienkripsi at-rest oleh Polyglot. Jika tim ingin backup berpassword, simpan password di vault dan referensikan di manifest.

### 5.3 Alur create restore point

```
1. Preflight (READ):
   - /system/resource/print  → version, free-hdd-space
   - /system/clock/print     → waktu router (untuk watchdog)
   - /file/print             → pastikan nama file belum ada
2. Scoped export untuk tiap menu:
   - kirim /<scope>/export file=polyglot-rp-<rp-id>-<scope-slug> compact terse hide-sensitive
   - tunggu file settle (system.Settled, timeout 15s)
3. Binary backup (default setiap mutation):
   - /system/backup/save name=polyglot-rp-<rp-id>-bin dont-encrypt=yes
   - tunggu file settle
   - skip jika hash konfigurasi sama dengan backup terakhir (kecuali path kritis)
4. Download:
   - /file/print ?name=<file> → size
   - /file/read file=<name> offset=... chunk-size=60000 (loop)
   - hitung SHA256, tulis ke storage terenkripsi
   - hapus file di router (best-effort) setelah download sukses
5. Generate inverse script dari diff (lihat 5.5)
6. Simpan DB row + artifacts (status AVAILABLE, trigger MUTATION/SCHEDULED/MANUAL)
7. Return RestorePoint{ID, Artifacts, TotalSizeBytes, ExpiresAt}
```

Jika langkah 1–6 gagal → **jangan** buat approval; return error `ErrCreationFailed`. Fail-closed.

### 5.4 Storage

Konfigurasi baru:

```env
RESTORE_POINT_DIR=data/restorepoints
RESTORE_POINT_RETENTION_DAYS=14
RESTORE_POINT_ENCRYPTION_KEY=<base64 32 bytes>
RESTORE_POINT_MAX_BYTES=52428800   # 50 MB
```

Struktur:

```
data/restorepoints/
  <tenant_id>/
    <device_id>/
      <restore_point_id>/
        manifest.json
        artifacts/
          scoped-<scope-slug>.rsc.enc
          binary.backup.enc
          inverse.rsc.enc
```

- Satu restore point adalah **bundle**; `manifest.json` mendaftar semua artifact (kind, scope, path, SHA256, size).
- Enkripsi AES-256-GCM; nonce disimpan di header file.
- File permission `0600`, directory `0700`.
- `manifest.json` **tidak** berisi secret; hanya metadata + hash.
- Path dibentuk dari UUID/ID tervalidasi, bukan input user.
- Di production, refuse start jika key kosong.
- Interface `port.RestorePointStorage` agar nanti bisa S3/MinIO tanpa ubah usecase.

### 5.5 Inverse script (rollback recipe)

Generate dari diff:

| Operasi | Inverse |
|---|---|
| `add` | `remove` berdasarkan `.id` yang baru dibuat, atau `comment` unik yang ditempel otomatis |
| `set` | `set` kembali nilai lama per field |
| `remove` | `add` kembali object lengkap dari snapshot `print detail` |
| `enable/disable` | lawannya |
| `move` | `move` kembali ke posisi lama |
| `reboot` | tidak ada inverse; gunakan post-verify + health check |
| `import` | restore dari scoped export / binary backup |

Untuk `add`, Polyglot menambahkan komentar penanda:
```
comment="polyglot:<approval_id>"
```
Kecuali command sudah punya `comment`; jika ada, gunakan itu + suffix marker? Jangan mengubah semantics; lebih aman gunakan `.id` hasil response `!done` (RouterOS mengembalikan `ret`/`.id` pada add). Jika tidak tersedia, fallback ke selector unik dari field yang dikirim.

Batasan:
- Inverse script adalah best-effort, bukan jaminan sempurna (mis. object dengan dependency).
- Scoped export tetap disimpan sebagai fallback manual.
- Binary backup adalah fallback terakhir (full restore + reboot).

### 5.6 Restore flow

```go
func (uc *RestorePointUseCase) Restore(ctx, restorePointID, actor, confirmation string) (Result, error)
```

Langkah:
1. Load restore point; cek `AVAILABLE`/`PARTIAL` dan belum expired.
2. Cek permission `restore:apply` + akses device.
3. Untuk restore kritis: wajib ketik nama device (`confirmation == device.Name`).
4. Upload file ke router:
   - `.rsc`/script kecil: `/file/add name=...` + `/file/set contents=` (≤60 KB per call).
   - File besar / binary: `/tool/fetch url=<signed one-time URL> dst-path=...` jika router bisa reach Polyglot; fallback chunked upload.
5. Terapkan:
   - Scoped export: `/import file=<name>` (merge; bukan snapshot sempurna).
   - Binary: `/system/backup/load name=<name>` lalu `/system/reboot`.
6. Post-verify read; mark `USED`; audit.
7. Hapus file sementara di router.

Catatan: `import` bersifat merge/aditif. Untuk rollback yang lebih tepat, gunakan inverse script; scoped export adalah bukti + manual fallback.

### 5.7 Retention & cleanup

- **Keep-last:** simpan minimal `RESTORE_POINT_KEEP_LAST` backup terakhir per device (default 30).
- **Age:** hapus backup lebih tua dari `RESTORE_POINT_RETENTION_DAYS` (default 14 hari).
- **Dedup:** binary backup dengan `ConfigHash` sama tidak disimpan ulang; artifact di-reuse (hardlink/object reference) dan tetap tercatat di bundle. Jangan hapus file yang masih direferensikan restore point lain.
- Job harian: hapus file + row lama (status `DELETED`, metadata audit tetap).
- Jangan hapus restore point yang terhubung ke approval `PENDING`/`EXECUTING`.
- Jangan hapus backup terakhir yang sukses untuk setiap device.
- Metrics: total bytes, jumlah restore point per device, cleanup errors, free space.

### 5.8 Backup terjadwal (harian/mingguan)

Worker baru: `internal/usecase/device/backup_scheduler.go`

- **Harian:** binary backup full + full export pada jam yang dikonfigurasi (`BACKUP_DAILY_AT=02:00`).
- **Mingguan:** binary backup + full export pada hari/jam (`BACKUP_WEEKLY_DAY=sun`, `BACKUP_WEEKLY_AT=03:00`).
- **Per device:** hanya device enabled dan `driver_type` MikroTik; lewati device yang sedang ada approval `EXECUTING`.
- **Dedup:** jika `ConfigHash` sama dengan backup terjadwal terakhir, reuse binary artifact (hardlink/object reference) dan catat audit "reused_unchanged"; jangan hilangkan artifact dari bundle.
- **Multi-instance:** gunakan Postgres advisory lock per device (`pg_try_advisory_lock`) agar tidak dobel saat server >1 replika.
- **Retry:** maksimal 2x dengan backoff; jika gagal, alert + audit `FAILED`.
- **Tidak butuh approval:** actor `SYSTEM`; tetapi tetap audit dan tidak boleh mengubah konfigurasi (hanya export/backup).
- **Batasan:** backup terjadwal tidak boleh berjalan bersamaan dengan mutation execution pada device yang sama (mutex per device).

Konfigurasi:

```env
BACKUP_DAILY_ENABLED=true
BACKUP_DAILY_AT=02:00
BACKUP_WEEKLY_ENABLED=true
BACKUP_WEEKLY_DAY=sun
BACKUP_WEEKLY_AT=03:00
RESTORE_POINT_KEEP_LAST=30
RESTORE_POINT_RETENTION_DAYS=14
```

### 5.9 API ConnectRPC

Proto baru: `api/proto/v1/restore_point.proto`

```proto
service RestorePointService {
  rpc ListRestorePoints(ListRestorePointsRequest) returns (ListRestorePointsResponse);
  rpc GetRestorePoint(GetRestorePointRequest) returns (GetRestorePointResponse);
  rpc CreateRestorePoint(CreateRestorePointRequest) returns (CreateRestorePointResponse);
  rpc RestoreFromPoint(RestoreFromPointRequest) returns (RestoreFromPointResponse);
  rpc DeleteRestorePoint(DeleteRestorePointRequest) returns (DeleteRestorePointResponse);
}
```

Aturan:
- List/Get: `restore:read` + akses device.
- Create manual: `restore:create` + akses device; trigger `MANUAL`; tidak mengubah konfigurasi.
- Restore: `restore:apply` + akses device + typed confirmation (`confirmation == device.name`), hanya user JWT; tidak diekspos ke MCP.
- Delete: `restore:delete`; soft-delete (`DELETED`) jika masih direferensikan approval/audit.
- Response memuat `artifacts[]` (kind, scope, sha256, size, download URL terautentikasi).
- Semua aksi menulis audit + `command_logs` bila mengeksekusi command.

---

## 6. P0-B: Approval — Approve, Reject, Edit

### 6.1 Domain

File baru: `internal/domain/approval/request.go`

```go
type Status string

const (
    StatusPending   Status = "PENDING"
    StatusApproved  Status = "APPROVED"
    StatusExecuting Status = "EXECUTING"
    StatusExecuted  Status = "EXECUTED"
    StatusFailed    Status = "FAILED"
    StatusRejected  Status = "REJECTED"
    StatusExpired   Status = "EXPIRED"
    StatusCancelled Status = "CANCELLED"
)

type Request struct {
    ID              string
    TenantID        string
    DeviceID        string
    SessionID       string
    RequestedBy     string        // actor id (AI session/user)
    ActorType       string        // "AGENT" | "USER" | "SYSTEM"
    Command         command.Command
    Revision        int
    Diff            Diff
    Validation      Validation    // hasil /console/inspect
    RestorePointID  string
    Status          Status
    Reason          string        // alasan reject / catatan
    EditedCommand   *command.Command
    ApprovedBy      string
    CreatedAt       time.Time
    ExpiresAt       time.Time
    DecidedAt       *time.Time
    ExecutedAt      *time.Time
    ExecutionResult *ExecutionResult
}
```

State machine:

```
PENDING ──approve──► APPROVED ──execute──► EXECUTING ──► EXECUTED
   │                    │                        └──────► FAILED
   ├──reject──────────► REJECTED
   ├──edit────────────► PENDING (revision+1, diff/validation ulang)
   ├──expire──────────► EXPIRED
   └──cancel──────────► CANCELLED
APPROVED ──timeout───► EXPIRED (belum dieksekusi)
```

Aturan:
- `PENDING` hanya bisa diedit/di-approve/di-reject.
- `APPROVED` belum dieksekusi; execution mengambil row lock dan set `EXECUTING`.
- `EXECUTING` tidak bisa diedit/di-approve ulang (anti double-execute).
- Semua transisi menulis audit.
- Expiry default 15 menit (`APPROVAL_TTL_MINUTES`).

### 6.2 Diff engine

File baru: `internal/usecase/network/diff.go`

```go
type Diff struct {
    DeviceID string
    Menu     string
    Verb     string
    Summary  string
    Lines    []DiffLine
    Before   []map[string]string
    After    map[string]string
}

type DiffLine struct {
    Op     string // "context" | "add" | "remove" | "change"
    Path   string // mis. "action", "src-address"
    Before string
    After  string
    Text   string
}
```

Cara membangun diff:

1. Parse path → `menu` + `verb`.
2. Tentukan selector:
   - `.id=*A` → langsung.
   - `numbers=*A` → langsung.
   - `?name=foo` → filter.
   - `numbers=[find comment="x"]` → resolve dengan READ `/menu/print detail where comment="x"` (bounded; maks 50 baris).
   - tanpa selector pada `set/remove` → **tidak boleh**; proposal ditolak dengan pesan "selector wajib".
3. READ snapshot:
   ```
   /<menu>/print detail .id=<id>
   ```
   atau `?<field>=<value>`.
4. Bangun diff:
   - `add`: semua args → `+` lines.
   - `set`: bandingkan field yang dikirim vs snapshot → `~` lines (`before` → `after`).
   - `remove`: semua field snapshot → `-` lines.
   - `move`: posisi lama vs baru.
   - `reboot/reset`: summary + warning; tidak ada field diff.
5. Untuk multi-command (`propose_config`): satu `CommandDiff` per baris; gabungkan.

UI hanya menampilkan diff yang sudah ter-render + opsi "raw JSON".

### 6.3 API ConnectRPC

Proto baru: `api/proto/v1/approval.proto`

```proto
service ApprovalService {
  rpc ListApprovals(ListApprovalsRequest) returns (ListApprovalsResponse);
  rpc GetApproval(GetApprovalRequest) returns (GetApprovalResponse);
  rpc DecideApproval(DecideApprovalRequest) returns (DecideApprovalResponse);
  rpc CancelApproval(CancelApprovalRequest) returns (CancelApprovalResponse);
}

enum ApprovalDecision {
  APPROVAL_DECISION_UNSPECIFIED = 0;
  APPROVAL_DECISION_APPROVE = 1;
  APPROVAL_DECISION_REJECT = 2;
  APPROVAL_DECISION_EDIT = 3;              // update proposal, tetap PENDING
  APPROVAL_DECISION_EDIT_AND_APPROVE = 4;  // update + approve dalam satu langkah
}

message DecideApprovalRequest {
  string id = 1;
  ApprovalDecision decision = 2;
  string edited_command = 3;   // hanya untuk EDIT / EDIT_AND_APPROVE
  map<string,string> edited_args = 4;
  string reason = 5;           // wajib untuk reject
  int32 revision = 6;          // optimistic concurrency
}
```

Aturan:
- `DecideApproval` hanya untuk user manusia (JWT), bukan MCP.
- Permission: `approval:decide` + akses device.
- `reason` wajib untuk reject.
- `revision` mencegah approve diff yang sudah diedit orang lain.
- Edit (`APPROVAL_DECISION_EDIT`):
  - Hanya `PENDING`; permission `approval:edit` + akses device.
  - `device_id` immutable.
  - Re-run classifier → harus tetap `MUTATION`.
  - Re-run `/console/inspect` (v7) → validasi.
  - Rebuild diff + restore point scope.
  - Revision +1; status tetap `PENDING`; audit old/new command.
  - Restriction: edited command harus berada di menu yang sama dengan proposal awal, kecuali user punya `approval:edit:any`.
- Edit-and-approve (`APPROVAL_DECISION_EDIT_AND_APPROVE`): edit di atas + `approval:decide`, lalu lanjut ke langkah approve.
- Approve (`APPROVAL_DECISION_APPROVE`):
  - Cek restore point `AVAILABLE`.
  - Cek device enabled + reachable.
  - Set `APPROVED`, lalu panggil execution pipeline.

### 6.4 UI Approval — shadcn AI Elements

Halaman: `/ai-ops/approvals` (Approval Inbox) + detail `Sheet`/`Dialog`. Semua komponen memakai shadcn yang sudah ada; tidak ada UI eksternal.

| Bagian | Komponen shadcn/AI Elements | Catatan |
|---|---|---|
| Daftar approval | `Queue` (pattern `Trace`) + `Tabs` (`Pending`, `Executed`, `Rejected`) + `Badge` status | Queue membagi "To do" dan "Completed"; filter device/status/waktu. |
| Baris approval | `Item` + `ItemMedia`/`ItemContent`/`ItemActions` | Ringkas: device, command, mode, expiry countdown, status. |
| Detail | `Sheet` (mobile) / `Dialog` (desktop) + `Tabs` | Tab: **Diff**, **Command**, **Validation**, **Restore Point**. |
| Diff | `CodeBlock` + `CodeBlockLine` (`variant="added"`/`"removed"`/`"highlighted"`) | Jangan render `<pre>` manual; jangan pakai warna mentah. |
| Command editor | `InputGroup` + `Textarea` + `Button` **Validate** / **Save edit** | `EDIT` menyimpan tanpa eksekusi; revision naik; validasi ulang via inspect/static. |
| Approve / Reject | `Confirmation` pattern (`Item` + `Button`) + `AlertDialog` untuk reject reason | Reject wajib alasan. Approve memanggil `ApprovalService.DecideApproval` dengan JWT. |
| Critical action | `AlertDialog` + input nama device | Untuk upgrade yang dipropos AI. Restore/reset adalah alur manusia terpisah dengan typed confirmation. |
| Restore point | `Card`/`Item` + `Badge` + `Button` download | Tampilkan artifact kinds, total size, SHA256 per artifact, expiry, status. |
| Empty/loading | `Empty`, `Skeleton`, `Spinner` | Jangan bikin empty state manual. |
| Warning v6 | `Alert` | "Sintaks belum diverifikasi router." |

Aturan UI:
- Tombol **Approve** tidak pernah mengeksekusi command langsung; ia memanggil `DecideApproval` dan menunggu status server (`APPROVED` → `EXECUTING` → `EXECUTED`).
- Tombol di dalam pesan chat (Confirmation) hanya memanggil API yang sama; tidak ada approval lewat MCP.
- `revision` dikirim dari server; jika sudah berubah, tampilkan konflik dan minta refresh.
- Setelah approve, tampilkan progress `Trace`/`Task`: execute → post-verify → watchdog cancel.
- Semua status memakai `Badge` semantic (`secondary`, `destructive`, `outline`), bukan warna hardcoded.
- Countdown expiry memakai `date-fns`; saat `EXPIRED`, tombol approve disabled dan muncul `Alert`.
- Aksi approve/reject/edit masuk audit dan `command_logs`.

### 6.5 Anti-replay & concurrency

- Execution memakai `SELECT ... FOR UPDATE` pada row approval.
- Cek `status=APPROVED`, `revision` cocok, `expires_at > now()`.
- Set `EXECUTING` sebelum mengirim command.
- Simpan `execution_nonce` (UUID) di audit.
- Jika timeout jaringan: status `FAILED`/`UNKNOWN`; **jangan retry otomatis**. Lakukan READ verifikasi dulu, lalu minta user memutuskan.
- Idempotency key untuk `DecideApproval`: `approval_id + revision + decision`.

### 6.6 Audit

Tabel `command_logs`:

```sql
CREATE TABLE command_logs (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id       TEXT NOT NULL DEFAULT 'tenant-default',
    device_id       UUID REFERENCES devices(id) ON DELETE SET NULL,
    actor_type      TEXT NOT NULL,          -- USER | AGENT | SYSTEM
    actor_id        TEXT,
    session_id      TEXT,
    mode            TEXT NOT NULL,          -- READ | MUTATION
    command         TEXT NOT NULL,
    command_args    JSONB,
    approval_id     TEXT REFERENCES approval_requests(id) ON DELETE SET NULL,
    restore_point_id TEXT REFERENCES restore_points(id) ON DELETE SET NULL,
    status          TEXT NOT NULL,          -- SUCCESS | FAILED | DENIED | PENDING_APPROVAL | REJECTED | EXPIRED
    output_summary  TEXT,
    duration_ms     BIGINT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

Semua percobaan dicatat, termasuk yang ditolak/expired. `output_summary` di-mask (`system.MaskSecrets`) dan dipotong (mis. 4 KB).

---

## 7. Execution Pipeline & Watchdog Rollback

### 7.1 Execute approved

File baru: `internal/usecase/network/execute_approved.go`

```go
func ExecuteApproved(
    ctx context.Context,
    deps Deps,
    approvalID string,
    revision int,
) (command.Result, error)
```

Langkah:
1. Load approval + restore point, validasi status/revision/expiry.
2. Ambil driver via registry (device aktif).
3. Cek capabilities (v7/v6).
4. (Opsional) pasang watchdog scheduler.
5. `driver.Execute` command (bukan `ExecuteCommandPreApproved` mentah; tetap lewat policy `DecisionDeny`).
6. Post-verify READ (health/status + menu terkait).
7. Cancel watchdog jika sukses.
8. Update status + audit.
9. Return result ke UI; MCP session bisa poll `approval_get`.

### 7.2 Watchdog rollback (pengganti "commit confirmed")

RouterOS API tidak punya Safe Mode. Yang tersedia:
- Safe Mode hanya di console/terminal (Ctrl+X), tidak via API.
- Scheduler dapat menjalankan script saat koneksi putus.

Implementasi:

```
1. Upload inverse script ke router: /file/add name=polyglot-wd-<approval>.rsc
2. Baca /system/clock/print → hitung start-time + WATCHDOG_SECONDS
3. /system/scheduler/add \
     name=polyglot-wd-<approval> \
     start-date=<router date> start-time=<router time> \
     interval=0 on-event="/import file=polyglot-wd-<approval>.rsc" \
     policy=ftp,reboot,read,write,policy,test,password,sniff,sensitive
4. Jalankan mutation.
5. Post-verify sukses → /system/scheduler/remove [find name=...]
6. Jika post-verify gagal / user klik rollback → biarkan scheduler jalan, atau trigger manual.
7. Cleanup file script setelah watchdog dibatalkan/dijalankan.
```

Catatan penting:
- Waktu memakai clock router, bukan server.
- Jika perubahan menyebabkan router reboot, scheduler tetap ada dan akan jalan setelah boot.
- Watchdog adalah mutasi juga; dibuat oleh `actor_type=SYSTEM` dengan `approval_id` asal, dan dicatat.
- Jangan pakai watchdog untuk `reboot`/`reset-configuration` (tidak relevan).
- Batasi 1 watchdog aktif per device; jika ada yang belum selesai, tolak approval baru.
- Default `WATCHDOG_SECONDS=600`; bisa 300–1800.
- Risiko: jika inverse script salah, watchdog bisa memperburuk. Karena itu inverse script diuji di CHR/staging dan ditampilkan di UI sebelum approve.

### 7.3 Post-verify

Minimal:
- `/system/resource/print` (device alive, uptime tidak reset tak terduga).
- Menu yang diubah: `print` ulang, bandingkan dengan diff `after`.
- Untuk firewall/route: cek rule/route ada dan `active`.
- Untuk interface: cek `running`.
- Untuk hotspot/PPP: cek session count tidak anomali.

Jika post-verify gagal:
- Status `FAILED` + `rollback_recommended=true`.
- UI menampilkan tombol **Rollback Sekarang**.
- Jika watchdog aktif, biarkan sampai waktu habis atau trigger manual.

---

## 8. Active Router Resolution

### 8.1 Tujuan (penjelasan sederhana)

Analoginya: saat operator membuka **Router-CabangA** di UI Polyglot, lalu bertanya ke AI *"cek firewall"*, AI harus membaca **Router-CabangA**, bukan router lain.

"Active device binding" = catatan kecil di backend: **"user X sedang memilih router Y"**. Catatan ini dipakai AI kalau tool tidak menyebut `device_id`.

Default yang dipakai:
- **Per user** (utama): satu router aktif per user, mengikuti router terakhir yang dipilih di UI.
- **Per session** (opsional, internal): kalau nanti ada multi-tab/multi-chat, tiap sesi bisa punya router sendiri.
- **Explicit `device_id`** selalu menang.
- Tidak ada fallback acak ke router lain.

AI harus membaca/mengeksekusi ke **router yang sedang aktif di UI/session operator**, bukan router dari `.env` seperti referensi `mikrotik-inspect`.

### 8.2 Komponen

- `internal/usecase/device/active_device.go`
- `port.ActiveDeviceStore`:

```go
type ActiveDevice struct {
    TenantID  string
    UserID    string
    SessionID string
    DeviceID  string
    UpdatedAt time.Time
}

type ActiveDeviceStore interface {
    Set(ctx context.Context, a ActiveDevice, ttl time.Duration) error
    Get(ctx context.Context, tenantID, userID, sessionID string) (ActiveDevice, error)
    Clear(ctx context.Context, tenantID, userID, sessionID string) error
}
```

Implementasi:
- Redis (utama): key `polyglot:active_device:{tenant}:{user}:{session}` + TTL 24 jam.
- Postgres fallback/audit: tabel `active_devices`.
- In-memory fallback untuk dev/test.

### 8.3 Session ID

- Ambil dari header MCP `Mcp-Session-Id` (streamable HTTP) atau `X-Polyglot-Session-ID` (opsional multi-tab).
- Bungkus `HTTPHandler`:

```go
func (s *Server) HTTPHandler() http.Handler {
    h := mcp.NewStreamableHTTPHandler(...)
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        sid := r.Header.Get("Mcp-Session-Id")
        if sid == "" { sid = r.Header.Get("X-Polyglot-Session-ID") }
        ctx := context.WithValue(r.Context(), ctxKeySessionID, sid)
        h.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

- Jika SDK menyediakan session info di `CallToolRequest`, pakai itu; header hanya fallback.

### 8.4 Resolution order

```
1. explicit device_id di argumen tool
   → validasi user punya akses (user_devices / role)
2. active device untuk (tenant, user, session)
3. active device untuk (tenant, user)
4. jika hanya ada 1 device MikroTik enabled di tenant dan AI_SINGLE_DEVICE_FALLBACK=true
   → pakai + log warning
5. error ErrNoActiveDevice:
   "No active router selected. Open Devices and select a router, or pass device_id."
```

Jangan pernah fallback ke device pertama secara acak.

### 8.5 UI Active Device — shadcn

- `web/src/components/layout/device-switcher.tsx` tetap jadi pemilih router global, tetapi saat user memilih:
  - update `device-store` (UI state), dan
  - panggil `DeviceService.SetActiveDevice` (server binding untuk AI).
- `device-store` hanya untuk tampilan; **source of truth AI adalah server**.
- Jangan biarkan auto-select `devices[0]` membuat binding AI secara implisit. Jika belum ada binding, AI Ops menampilkan `Alert` + tombol **Pilih Router**.
- Badge di device chip: **"AI target"**; jika v6: tambah `Badge variant="secondary"` **"v6 degraded"**.
- AI Ops header memakai `DropdownMenu`/`Command` untuk ganti router; setelah ganti, conversation baru dimulai untuk device itu (history lama tetap), dan UI menampilkan `Marker` konteks device.
- Jika device dihapus/disabled → `ClearActiveDevice`, tampilkan `Alert`, dan hentikan tool call.
- Tambah item sidebar **AI Router Ops** (ikon `Bot`/`Sparkles`) dengan permission `ai:ops`, di grup `AI Omnichannel`.

### 8.6 Kenapa bukan `.env`

- `.env` tidak multi-tenant, tidak audit, tidak RBAC, tidak mengikuti UI.
- Kredensial harus dari `port.CredentialVault` via registry.
- AI tidak boleh tahu kredensial mentah (sudah menjadi prinsip arsitektur).

---

## 9. UI Polyglot — shadcn + AI Elements

UI **termasuk scope plan ini** dan **bukan LibreChat**. Semua dibangun di `web/` dengan shadcn/ui yang sudah dipakai proyek.

### 9.1 Prinsip

- shadcn tetap `new-york`, Radix, lucide, Tailwind v4; jangan ganti preset.
- **AI Elements bukan package/dependency.** Install primitif shadcn, lalu copy pattern dari docs Radix ke `web/src/components/ai-elements/`.
- Jangan hand-roll bubble/scroll/collapsible/code block. Pakai primitif shadcn.
- Jangan tampilkan raw hidden chain-of-thought. `ChainOfThought`/`Reasoning` hanya merender langkah/ringkasan yang memang dikirim backend.
- **Confirmation di chat bukan otoritas approval.** Approval tetap server-side (JWT + permission + revision + restore point).
- Semua warna semantic (`bg-background`, `text-muted-foreground`, `Badge` variant), tanpa `bg-blue-500`/`dark:` manual.
- Spacing `gap-*`, bukan `space-y-*`; `size-*` untuk kotak; `cn()` untuk conditional.

### 9.2 Instalasi primitif

```bash
cd web
# Primitif yang dipakai pattern AI Elements (sesuai docs Radix).
pnpm dlx shadcn@latest add message-scroller message bubble attachment input-group item marker badge
pnpm dlx shadcn@latest add empty spinner progress hover-card alert card tabs sheet dialog alert-dialog button
# trace/code-block: cek registry; jika tersedia:
# pnpm dlx shadcn@latest add trace code-block
```

Catatan:
- `message-scroller` otomatis menambah `@shadcn/react` (hooks scroller).
- `trace`/`code-block` adalah komponen AI Elements. Jika registry CLI belum menyediakannya (`pnpm dlx shadcn@latest view @shadcn/trace`), copy source dari docs ke `web/src/components/ai-elements/`; jangan bikin dari nol.
- `prompt-input`, `model-selector`, `confirmation`, `queue`, `checkpoint`, `chain-of-thought`, `reasoning`, `plan`, `task`, `tool`, `terminal`, `sources`, `context`, `suggestion`, `file-tree`, `stack-trace`, dan `chat` adalah **pattern**, bukan item registry. Compose dari primitif di atas.
- Cek dulu `pnpm dlx shadcn@latest info`; jangan re-add komponen yang sudah ada (button, card, dialog, sheet, tabs, badge, alert, dropdown-menu, tooltip, command sudah terpasang).
- Copy pattern Radix dari `https://ui.shadcn.com/docs/ai-elements/radix/*`; simpan di `web/src/components/ai-elements/` dan pertahankan nama ekspor agar mudah update.
- Package manager proyek adalah `pnpm` (`packageManager` di `web/package.json`).

### 9.3 Peta komponen → fitur Polyglot

| AI Element / shadcn | Dipakai untuk | Catatan |
|---|---|---|
| `Chat` (`MessageScroller` + `Message` + `Bubble`) | Thread AI Ops | `scrollAnchor` di pesan user; `autoScroll`; `MessageScrollerButton`. |
| `PromptInput` (`InputGroup`) | Composer | Textarea + device chip + send; attachment opsional (export/log). |
| `ModelSelector` (`DropdownMenu` radio) | Pilih LLM config | Dari `LLMConfigService.ListLLMConfigs`; menu effort disembunyikan sampai provider mendukung. |
| `ChainOfThought` (`Trace` collapsible) | Ringkasan langkah agent | Bukan hidden reasoning; summary + expand ke Trace. |
| `Reasoning` (`Trace`) | Delta reasoning kalau provider kirim | Kalau tidak ada, jangan tampilkan. |
| `Tool` (`Trace`) | Tool call MCP/agent | Input/output, status running/error; `CodeBlock` untuk output. |
| `Terminal` (`Trace` + `CodeBlock`) | Output command RouterOS | Streaming output; status exit; bukan pengganti Xterm interaktif. |
| `Plan` / `Task` (`Trace`) | Pipeline mutation | inspect → snapshot → backup → diff → approval → execute → verify. |
| `Queue` (`Trace`) | Approval inbox / todo | To do vs Completed. |
| `Confirmation` (`Item` + `Button`) | Approval prompt di chat | Approve/Deny memanggil `ApprovalService`, bukan MCP. |
| `Checkpoint` (`Marker` + `Button`) | Checkpoint percakapan | Restore messages sampai turn itu. |
| `Marker` | Restore point / ganti device | Timeline restore point dan penanda konteks. |
| `CodeBlock` (`CodeBlockLine`) | Diff, command, raw JSON, terminal | `added`/`removed`/`highlighted`; copy button. |
| `Sources` / `InlineCitation` | Referensi docs/command | Opsional; tautan ke docs RouterOS/audit. |
| `Context` (`HoverCard` + `Progress`) | Token usage | Dari metadata conversation. |
| `FileTree` | Artifact restore point | manifest, scoped export, binary backup, inverse script. |
| `Suggestion` | Quick prompts | "Cek firewall", "Buat backup", "Lihat interface down". |
| `StackTrace` | Command gagal | Output error + langkah. |
| `Empty`/`Skeleton`/`Spinner` | Empty/loading | Standar shadcn. |
| `Alert` | Warning v6 / error | Bukan custom callout. |

### 9.4 Halaman & route

```
/ai-ops                    Chat + right rail konteks
/ai-ops/approvals          Approval Inbox (Queue + detail Sheet)
/ai-ops/restore-points     Restore point list/timeline
/devices/$id               Tab "AI & Restore Points" + "Set as AI target"
```

- Sidebar: tambah **AI Router Ops** di grup `AI Omnichannel`, permission `ai:ops`.
- Chat page layout: `MessageScroller` (kiri) + `Sheet`/rail (kanan) untuk device, capabilities, pending approvals, restore points.
- Mobile: approval detail `Sheet`; chat full width; right rail jadi `Sheet`.
- `device-switcher` existing dipertahankan; tambah badge AI target.

### 9.5 Streaming contract

Backend: `api/proto/v1/ai_chat.proto`

```proto
// Connect-Web hanya mendukung server-streaming; request unary, response stream.
service AIChatService {
  rpc Chat(ChatRequest) returns (stream ChatEvent);
}

message ChatRequest {
  string conversation_id = 1;   // kosong = conversation baru
  string message = 2;           // diabaikan saat resume (since_event_id > 0)
  string device_id = 3;         // opsional; explicit menang atas active binding
  string model_config_id = 4;   // opsional; default LLM config aktif
  int64  since_event_id = 5;    // >0 = resume stream
}

message ChatEvent {
  oneof event {
    TextDelta text = 1;
    ReasoningDelta reasoning = 2;
    ToolCall tool_call = 3;
    ToolResult tool_result = 4;
    PlanStep plan_step = 5;
    TerminalOutput terminal = 6;
    ApprovalRequested approval = 7;
    Checkpoint checkpoint = 8;
    Usage usage = 9;
    Done done = 10;
    Error error = 11;
  }
}
```

- Client mengirim pesan + `conversation_id` + `device_id` (opsional) + `model_config_id`.
- Server resolve active device, jalankan LLM + tools, dan stream event.
- `ToolCall`/`ToolResult` untuk `router_read`, `inspect_*`, `propose_command`, dll.
- `ApprovalRequested` membawa `approval_id`, diff summary, restore point status.
- `TerminalOutput` untuk output command (bukan Xterm).
- Reconnect: conversation + event persisted; client resume dengan `ChatRequest.since_event_id` (tanpa mengirim ulang pesan).
- Fallback jika provider tidak streaming: kirim `TextDelta` besar sekali + event selesai; UI tetap sama.

### 9.6 Restore point & checkpoint

Dua konsep berbeda, jangan dicampur:

| Konsep | Komponen | Fungsi |
|---|---|---|
| Conversation checkpoint | `Checkpoint` (`Marker`) | Kembali ke turn percakapan; hapus pesan setelahnya. |
| Router restore point | `Card`/`Item` + `Marker` timeline | Artifact backup router; restore ke device. |

- Restore point list: `Table`/`Card`, `Badge` status, total size, artifact kinds, SHA256 per artifact (truncate + copy), expiry, trigger (mutation/scheduled/manual).
- Tombol **Backup now** memanggil `RestorePointService.CreateRestorePoint` (trigger `MANUAL`); progress via `Trace`/`Task`; tidak membuat approval karena tidak mengubah konfigurasi.
- Download artifact memakai signed URL terautentikasi; jangan expose path storage langsung.
- Detail: `Tabs` (Manifest, Config, Backup, Inverse script) + `CodeBlock` + `FileTree`.
- Aksi restore: `AlertDialog` + typed device name; hanya user dengan `restore:apply`.
- Timeline: `Marker` per backup; `Badge` "scheduled" / "before mutation".
- Cleanup status (`DELETED`) tetap tampil di audit, bukan hilang.

### 9.7 Terminal vs Xterm

- **AI Elements Terminal** = output command yang dijalankan agent, read-only, streaming.
- **Xterm** (`@xterm/xterm` + WS PTY) tetap untuk console interaktif/emergency takeover.
- Jangan menaruh Xterm di dalam `Message`/`Bubble`; buka di `Sheet`/route terpisah.
- Terminal agent harus menampilkan command, exit status, durasi, dan tombol copy; output di-mask.

### 9.8 Approval di chat vs inbox

- Chat menampilkan `Confirmation` ringkas setelah `approval-requested`.
- Approve/Deny di chat memanggil `ApprovalService.DecideApproval` (JWT, permission, revision).
- Untuk critical action, chat hanya menampilkan tombol **Review** yang membuka Approval Inbox; tidak ada approve inline.
- Jangan pernah menaruh tombol approve di dalam output LLM yang bisa dipalsukan; tombol dirender oleh aplikasi dari event terstruktur.
- Setelah diputuskan, `Confirmation` berubah menjadi `Tool` line (`approved`/`denied`).

### 9.9 Design & accessibility

- Gunakan `Field`/`FieldGroup` untuk form edit command; `data-invalid` + `aria-invalid`.
- `Dialog`/`Sheet` wajib punya `Title` (bisa `sr-only`).
- `TabsTrigger` di dalam `TabsList`; `SelectItem` di dalam `SelectGroup`; `CommandItem` di dalam `CommandGroup`.
- Icon di `Button` pakai `data-icon`; jangan `size-4` manual.
- `CodeBlock` untuk diff/terminal; jangan `<pre>`.
- `Badge` untuk status; jangan span berwarna.
- `Alert` untuk warning; jangan div custom.
- `Empty` untuk kosong; `Skeleton` untuk loading.
- Dark mode otomatis via token; jangan `dark:` manual.
- Keyboard: approval actions focusable; `AlertDialog` untuk destructive; ESC menutup Sheet.
- Reduced motion: hormati `prefers-reduced-motion` untuk shimmer/streaming.

### 9.10 Test UI

- Vitest + browser (sudah ada) untuk komponen: Confirmation, Queue, Checkpoint, diff rendering.
- Playwright: alur chat → propose → approve → executed; reject dengan alasan; edit + revision; v6 warning.
- Contract test event streaming: urutan event, reconnect, error.
- A11y: axe/role checks pada dialog/sheet.
- Visual: screenshot dark/light untuk Approval Inbox dan Chat.

### 9.11 File UI

```
web/src/components/ui/                     # primitif shadcn hasil `add`
  message-scroller.tsx  message.tsx  bubble.tsx  input-group.tsx
  item.tsx  marker.tsx  empty.tsx  spinner.tsx  progress.tsx  hover-card.tsx

web/src/components/ai-elements/            # pattern copy dari docs Radix
  trace.tsx  code-block.tsx                # copy bila registry belum menyediakan
  chat.tsx  message.tsx  prompt-input.tsx  model-selector.tsx
  chain-of-thought.tsx  reasoning.tsx  plan.tsx  task.tsx
  queue.tsx  tool.tsx  terminal.tsx  confirmation.tsx
  checkpoint.tsx  sources.tsx  context.tsx  suggestion.tsx
  file-tree.tsx  stack-trace.tsx

web/src/features/ai-ops/
  index.tsx
  api/{keys.ts,use-ai-chat.ts,use-approvals.ts,use-restore-points.ts,use-capabilities.ts}
  components/{chat-thread.tsx,prompt-composer.tsx,device-context.tsx,
              approval-card.tsx,approval-inbox.tsx,approval-detail.tsx,
              diff-view.tsx,restore-point-list.tsx,restore-point-detail.tsx,
              v6-warning.tsx,plan-trace.tsx,tool-trace.tsx,terminal-trace.tsx}
  stores/{chat-store.ts,approval-ui-store.ts}

web/src/routes/_authenticated/ai-ops/{index.tsx,approvals.tsx,restore-points.tsx}
web/src/routes/_authenticated/devices/$id.tsx   # tambah tab AI
```

- `web/src/components/layout/data/sidebar-data.ts`: item AI Ops.
- `web/src/components/layout/device-switcher.tsx`: badge + SetActiveDevice.
- `web/src/styles/index.css`: hanya jika butuh token tambahan; jangan bikin file CSS baru.

### 9.12 Yang tidak dipakai

- LibreChat / UI eksternal.
- Package `ai-elements` / AI SDK Elements; pattern di-copy.
- `ScrollArea` untuk chat (pakai `MessageScroller`).
- `Collapsible` manual untuk reasoning/tool (pakai `Trace`).
- `<pre>`/div bubble manual.
- Menampilkan raw hidden reasoning.
- Tombol approve yang tidak memanggil server.

---

## 10. MCP Server Baru

### 10.1 Prasyarat keamanan (P0)

- `/mcp` **wajib** di belakang auth middleware (OAuth/OBO atau API key + user context).
- `mcp.New(d.reg, auditWriter)` — audit tidak boleh `nil`.
- Rate limit per user/session/device.
- Tenant scoping.
- Tool yang mengubah state tidak boleh dieksekusi tanpa approval server-side.

### 10.2 Tool catalog

#### Read / Inspect (auto-approve)

| Tool | Fungsi | Anotasi |
|---|---|---|
| `get_active_device` | Router yang sedang aktif | ReadOnly |
| `router_capabilities` | versi, `inspect_supported`, `ai_mode`, warning v7 | ReadOnly |
| `get_device_status` | status ringkas | ReadOnly |
| `router_read` | scoped read: `print/get/export/monitor` (path + filter + proplist) | ReadOnly |
| `inspect_children` | `/console/inspect request=child` | ReadOnly |
| `inspect_syntax` | `request=syntax` | ReadOnly |
| `inspect_completion` | `request=completion` | ReadOnly |
| `inspect_validate` | `request=highlight` → valid/invalid + error tokens | ReadOnly |

#### Mutation (proposal only)

| Tool | Fungsi |
|---|---|
| `propose_command` | Satu command mutation → approval + diff + restore point |
| `propose_config` | Multi-line script → satu approval (maks 50 baris) |
| `get_approval` | Status + hasil eksekusi |
| `list_approvals` | Daftar pending/recent |
| `cancel_approval` | Batalkan proposal milik sendiri |

#### Restore point

| Tool | Fungsi |
|---|---|
| `list_restore_points` | Daftar restore point device |
| `create_restore_point` | Backup manual (read-only terhadap konfigurasi, tapi menulis file di router → tetap butuh approval? Lihat catatan) |
| `get_restore_point` | Metadata + hash |

Catatan: `create_restore_point` menulis file ke router (`/export`/`/system/backup/save`), tetapi **tidak mengubah konfigurasi**. Operasi ini diizinkan sebagai operasi sistem (bukan command konfigurasi; tidak menambah mode baru), tetap audit, dan dibatasi rate + size + free space. Jika `RESTORE_POINT_MAX_BYTES` terlampaui, tool gagal dan tidak membuat approval.

#### Tidak diekspos ke MCP

- `approve_approval` / `reject_approval` — hanya ConnectRPC UI dengan user manusia.
- `restore_from_point` — hanya UI dengan typed confirmation.
- `reset_configuration` — hard deny untuk AI.

Approval di chat UI memakai ConnectRPC `ApprovalService` (JWT user), **bukan MCP elicitation**. MCP elicitation tidak dipakai di v1 (lihat §19).

### 10.3 Perubahan tool lama

`run_command`:
- READ → eksekusi seperti biasa, audit.
- MUTATION → **tidak eksekusi**; panggil `propose_command`, return:

```json
{
  "status": "pending_approval",
  "executed": false,
  "approval_id": "apr_...",
  "device_id": "...",
  "command": "...",
  "diff": {...},
  "restore_point": {"id": "...", "artifacts": [{"kind": "...", "sha256": "..."}], "expires_at": "..."},
  "validation": {"supported": true, "valid": true},
  "message": "Mutation was NOT executed. Ask the operator to approve it in the Approval Inbox."
}
```

`push_config`: sama, multi-line → `propose_config`.

Tool description harus tegas:
> "This tool NEVER executes mutations. It creates an approval request. Do not tell the user the change was applied until approval status is EXECUTED."

### 10.4 Tool output wajib

Setiap output menyertakan:
- `device_id`, `device_name`, `active` (apakah device ini active binding)
- `routeros_version`, `ai_mode`
- `executed: true|false`
- `approval_id` (jika pending)
- `warning` (jika v6/degraded)

### 10.5 Server instructions

Tambahkan di `mcp.NewServer(..., &mcp.ServerOptions{Instructions: ...})`:

```
Polyglot RouterOS AI rules:
1. READ and INSPECT are free. Always inspect before proposing changes.
2. Every MUTATION (add/set/remove/enable/disable/reset/reboot/import/...) requires
   human approval and a restore point. Never claim a mutation was executed before
   the approval status is EXECUTED.
3. Use propose_command/propose_config. Do not bypass with run_command.
4. RouterOS v7 is required for /console/inspect. On v6, ai_mode=degraded and
   syntax validation is unavailable.
5. Never request or reveal credentials. Credentials stay in the Polyglot vault.
6. If no active router is selected, ask the operator to select one.
```

---

## 11. Database

Migration baru: `migrations/000041_ai_router_ops.up.sql` / `.down.sql`

> Urutan create di migration (FK): `restore_points` → `restore_point_artifacts` → `approval_requests` → `command_logs` → `active_devices` → `ai_conversations` → `ai_chat_events`. Jangan membuat `approval_requests` sebelum `restore_points` karena FK `restore_point_id`.

### 11.1 `approval_requests`

```sql
CREATE TABLE approval_requests (
    id                TEXT PRIMARY KEY,
    tenant_id         TEXT NOT NULL DEFAULT 'tenant-default',
    device_id         UUID REFERENCES devices(id) ON DELETE SET NULL,
    session_id        TEXT,
    actor_type        TEXT NOT NULL,             -- AGENT | USER | SYSTEM
    actor_id          TEXT NOT NULL,
    requested_by      TEXT NOT NULL,
    command           TEXT NOT NULL,
    command_args      JSONB NOT NULL DEFAULT '{}',
    revision          INT NOT NULL DEFAULT 1,
    edited_command    TEXT,
    edited_args       JSONB,
    diff              JSONB NOT NULL,
    validation        JSONB,
    restore_point_id  TEXT REFERENCES restore_points(id) ON DELETE SET NULL,
    status            TEXT NOT NULL DEFAULT 'PENDING',
    reason            TEXT,
    approved_by       BIGINT REFERENCES users(id) ON DELETE SET NULL,
    execution_result  JSONB,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    decided_at        TIMESTAMPTZ,
    executed_at       TIMESTAMPTZ,
    expires_at        TIMESTAMPTZ NOT NULL,
    CHECK (status IN ('PENDING','APPROVED','EXECUTING','EXECUTED','FAILED',
                      'REJECTED','EXPIRED','CANCELLED'))
);

CREATE INDEX idx_approval_pending ON approval_requests (tenant_id, status, created_at DESC);
CREATE INDEX idx_approval_device  ON approval_requests (device_id, created_at DESC);
```

### 11.2 `restore_points` & `restore_point_artifacts`

```sql
CREATE TABLE restore_points (
    id                TEXT PRIMARY KEY,
    tenant_id         TEXT NOT NULL DEFAULT 'tenant-default',
    device_id         UUID REFERENCES devices(id) ON DELETE SET NULL,
    trigger           TEXT NOT NULL,             -- MUTATION | SCHEDULED | MANUAL
    scope             JSONB NOT NULL,            -- gabungan ["/ip/firewall", ...]
    routeros_version  TEXT,
    config_hash       TEXT,                      -- dedup binary backup
    rollback_script   JSONB,
    manifest          JSONB NOT NULL,
    total_size_bytes  BIGINT NOT NULL DEFAULT 0,
    status            TEXT NOT NULL DEFAULT 'CREATING',
    created_by        TEXT NOT NULL,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at        TIMESTAMPTZ NOT NULL,
    used_at           TIMESTAMPTZ,
    CHECK (status IN ('CREATING','AVAILABLE','PARTIAL','USED','EXPIRED','DELETED'))
);

CREATE TABLE restore_point_artifacts (
    id                BIGSERIAL PRIMARY KEY,
    restore_point_id  TEXT NOT NULL REFERENCES restore_points(id) ON DELETE CASCADE,
    kind              TEXT NOT NULL,             -- SCOPED_EXPORT | BINARY_BACKUP | INVERSE_SCRIPT
    scope             JSONB NOT NULL DEFAULT '[]',
    storage_path      TEXT NOT NULL,
    sha256            TEXT NOT NULL,
    size_bytes        BIGINT NOT NULL,
    UNIQUE (restore_point_id, kind, storage_path)
);

CREATE INDEX idx_restore_device ON restore_points (device_id, created_at DESC);
CREATE INDEX idx_restore_status ON restore_points (status, expires_at);
CREATE INDEX idx_restore_artifacts ON restore_point_artifacts (restore_point_id);
```

### 11.3 `command_logs`

Lihat §6.6.

### 11.4 `active_devices` (fallback/audit)

```sql
CREATE TABLE active_devices (
    tenant_id  TEXT NOT NULL,
    user_id    TEXT NOT NULL,
    session_id TEXT NOT NULL DEFAULT '',
    device_id  UUID REFERENCES devices(id) ON DELETE CASCADE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (tenant_id, user_id, session_id)
);
```

### 11.5 `ai_conversations` & `ai_chat_events`

Dibutuhkan untuk streaming reconnect (`since_event_id`) dan checkpoint percakapan.

```sql
CREATE TABLE ai_conversations (
    id              TEXT PRIMARY KEY,
    tenant_id       TEXT NOT NULL DEFAULT 'tenant-default',
    user_id         BIGINT REFERENCES users(id) ON DELETE CASCADE,
    device_id       UUID REFERENCES devices(id) ON DELETE SET NULL,
    model_config_id TEXT,
    title           TEXT,
    status          TEXT NOT NULL DEFAULT 'ACTIVE',  -- ACTIVE | ARCHIVED
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE ai_chat_events (
    id              BIGSERIAL PRIMARY KEY,
    conversation_id TEXT NOT NULL REFERENCES ai_conversations(id) ON DELETE CASCADE,
    seq             BIGINT NOT NULL,
    event_type      TEXT NOT NULL,   -- text | reasoning | tool_call | tool_result | plan_step
                                     -- terminal | approval | checkpoint | usage | done | error
    payload         JSONB NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (conversation_id, seq)
);

CREATE INDEX idx_ai_events_conversation ON ai_chat_events (conversation_id, seq);
```

- `device_id` di conversation bersifat immutable; ganti device membuat conversation baru + `Marker` di UI.
- Checkpoint menghapus event setelah `seq` checkpoint dan mencatat audit UI (bukan audit keamanan).
- Event tidak boleh memuat credential, full config, atau raw provider payload.

### 11.6 Retention job

- `approval_requests`: retain 90 hari (audit), hapus `EXPIRED`/`CANCELLED` > 90 hari.
- `restore_points`: hapus file > `RESTORE_POINT_RETENTION_DAYS`, row → `DELETED`.
- `restore_point_artifacts`: hapus file + row setelah restore point `DELETED`, kecuali masih direferensikan bundle lain (hardlink/refcount).
- `command_logs`: retain 180 hari (configurable).
- `ai_chat_events`: retain `AI_CHAT_EVENT_RETENTION_DAYS` (default 30 hari); `ai_conversations` yang `ARCHIVED` > 90 hari dihapus.

---

## 12. File Impact Map

### 12.1 Baru

```
internal/domain/approval/
  request.go
  errors.go
  request_test.go

internal/domain/restorepoint/
  restore_point.go
  errors.go
  restore_point_test.go

internal/driver/mikrotik/inspect/
  client.go
  commands.go
  parser.go
  types.go
  capabilities.go
  client_test.go

internal/driver/mikrotik/classify.go
internal/driver/mikrotik/classify_test.go

internal/driver/mikrotik/system/restorepoint.go
internal/driver/mikrotik/system/restorepoint_test.go

internal/usecase/network/
  inspect_router.go
  propose_command.go
  propose_config.go
  decide_approval.go
  execute_approved.go
  restore_point.go
  diff.go
  validate_command.go
  *_test.go

internal/usecase/device/active_device.go
internal/usecase/device/active_device_test.go
internal/usecase/device/backup_scheduler.go
internal/usecase/device/backup_scheduler_test.go

internal/domain/aichat/
  conversation.go
  event.go
  errors.go

internal/usecase/aichat/
  chat.go
  stream.go
  conversation.go
  chat_test.go

internal/adapter/postgres/
  approval_repository.go
  restore_point_repository.go
  command_log_repository.go
  active_device_repository.go
  ai_chat_repository.go

internal/adapter/storage/
  restore_point_storage.go
  restore_point_storage_test.go

internal/adapter/connect/approval/
  router.go
  handler.go
  mapper.go

internal/adapter/connect/restorepoint/
  router.go
  handler.go
  mapper.go

internal/adapter/mcp/
  tool_router_capabilities.go
  tool_router_read.go
  tool_inspect_children.go
  tool_inspect_syntax.go
  tool_inspect_completion.go
  tool_inspect_validate.go
  tool_propose_command.go
  tool_propose_config.go
  tool_get_approval.go
  tool_list_approvals.go
  tool_cancel_approval.go
  tool_list_restore_points.go
  tool_create_restore_point.go
  tool_get_active_device.go

api/proto/v1/approval.proto
api/proto/v1/restore_point.proto
api/proto/v1/ai_chat.proto

internal/adapter/connect/aichat/
  router.go
  handler.go
  stream.go
  mapper.go

migrations/000041_ai_router_ops.up.sql
migrations/000041_ai_router_ops.down.sql
docs/adr/0011-ai-router-ops-approval-restore.md

web/src/components/ai-elements/           # pattern AI Elements, lihat §9.11
web/src/features/ai-ops/                  # chat, approval, restore point, lihat §9.11
web/src/routes/_authenticated/ai-ops/     # route /ai-ops, /ai-ops/approvals, /ai-ops/restore-points
web/src/routes/_authenticated/devices/$id.tsx  # device detail + tab AI & restore point
web/src/features/devices/components/device-ai-tab.tsx
```

Catatan: detail file UI ada di §9.11. Semua bagian UI **termasuk scope plan** dan dikerjakan di Fase 7–9 (§14).

### 12.2 Diubah

```
internal/domain/command/command.go        # Class → Mode
internal/domain/command/policy.go         # Decide(mode)
internal/domain/audit/log.go              # Event + ApprovalID/RestorePointID/Mode/Status
internal/port/audit.go                    # event fields (alias)
internal/port/device_repository.go        # (jika perlu active device)
internal/driver/mikrotik/commands.go      # classifier baru, hapus destructivePaths
internal/usecase/network/execute_command.go
internal/usecase/network/push_config.go
internal/adapter/mcp/server.go            # tool list, instructions, auth
internal/adapter/mcp/tool_run_command.go  # mutation → propose
internal/adapter/mcp/tool_push_config.go  # mutation → propose
internal/adapter/connect/device/router.go # SetActiveDevice/GetActiveDevice
internal/app/router.go                    # mount approval + restorepoint + AIChatService, protect /mcp, audit
internal/app/app.go                       # wire repos/usecases/storage/scheduler/AIChatService
internal/config/config.go                 # env baru
api/proto/v1/device.proto                 # Set/Get/ClearActiveDevice
web/src/components/layout/data/sidebar-data.ts  # menu AI Router Ops
web/src/components/layout/device-switcher.tsx   # SetActiveDevice + badge AI target
web/src/gen/v1/*                          # regenerated; jangan edit manual
web/src/routeTree.gen.ts                  # generated route tree
docs/database-schema.md                   # promosikan Lampiran A
Polyglot-Architecture.md                  # §6.1 alur baru
README.md                                 # fitur + warning v7
AGENTS.md                                 # aturan baru
data/system-prompt.md                     # instruksi AI
```

Catatan: perubahan `web/*` (chat AI Ops, approval inbox, restore point, device picker) **termasuk scope**; detail komponen di §9.11.

---

## 13. Konfigurasi Baru

```env
# AI Router Ops
AI_ROUTER_OPS_ENABLED=true
AI_V6_MUTATIONS=true               # izinkan mutation di v6 dengan warning degraded
AI_MAX_PROPOSALS_PER_MINUTE=10
AI_MAX_SCRIPT_LINES=50

# AI Chat / UI
AI_CHAT_STREAM_ENABLED=true
AI_CHAT_MAX_CONTEXT_MESSAGES=50
AI_CHAT_EVENT_RETENTION_DAYS=30

# Approval
APPROVAL_TTL_MINUTES=15
APPROVAL_REQUIRE_REASON_ON_REJECT=true

# Restore Point
RESTORE_POINT_DIR=data/restorepoints
RESTORE_POINT_RETENTION_DAYS=14
RESTORE_POINT_KEEP_LAST=30
RESTORE_POINT_ENCRYPTION_KEY=<base64-32-bytes>
RESTORE_POINT_MAX_BYTES=52428800    # 50 MB
RESTORE_POINT_DEDUPE=true          # reuse artifact, bukan menghilangkan binary
# Binary backup setiap mutation selalu aktif (D11); tidak ada toggle disable.

# Backup terjadwal
BACKUP_DAILY_ENABLED=true
BACKUP_DAILY_AT=02:00
BACKUP_WEEKLY_ENABLED=true
BACKUP_WEEKLY_DAY=sun
BACKUP_WEEKLY_AT=03:00

# Watchdog
WATCHDOG_ROLLBACK_SECONDS=600
WATCHDOG_ENABLED=true

# Active Device
ACTIVE_DEVICE_TTL_HOURS=24
AI_SINGLE_DEVICE_FALLBACK=false

# Inspect
INSPECT_TIMEOUT_SECONDS=5
INSPECT_CACHE_TTL_MINUTES=10
```

Validasi config saat start:
- Production: `RESTORE_POINT_ENCRYPTION_KEY` wajib.
- `RESTORE_POINT_DIR` harus writable.
- `APPROVAL_TTL_MINUTES` > 0.
- `RESTORE_POINT_KEEP_LAST` > 0; binary backup per mutation tidak bisa dimatikan.
- `AI_V6_MUTATIONS=false` hanya memblokir mutation AI di v6, bukan mutation manusia.

---

## 14. Fase Implementasi & Estimasi

> Setiap fase harus buildable, `make test` hijau, dan tidak meninggalkan regression.

### Fase 0 — Fondasi & Keamanan (P0, 2–3 hari)

- [ ] Ganti `Class` → `Mode` (READ/MUTATION) + test.
- [ ] Classifier MikroTik baru (read allowlist, default mutation) + 50 test case.
- [ ] Hapus `ExecuteCommandPreApproved` dari MCP; mutation → error/pending sementara.
- [ ] Lindungi `/mcp` dengan auth; wire audit writer.
- [ ] Tambah config + validasi.
- **Exit:** AI tidak bisa eksekusi mutation tanpa approval; semua percobaan teraudit.

### Fase 1 — Backup & Restore Point (P0, 5–7 hari)

- [ ] Domain `restorepoint` + error (`Trigger`, `ConfigHash`).
- [ ] Driver: scoped export, binary backup, file download chunked, cleanup.
- [ ] Storage bundle terenkripsi + manifest + SHA256 per artifact.
- [ ] Repository Postgres + migration `restore_points`.
- [ ] Usecase `CreateRestorePoint`, `ListRestorePoints`, `GetRestorePoint`.
- [ ] ConnectRPC `RestorePointService` (list/get/create) + protovalidate.
- [ ] Binary backup setiap mutation (default) + dedup `ConfigHash`.
- [ ] Backup terjadwal harian/mingguan + advisory lock.
- [ ] Retention/cleanup job (`KEEP_LAST` + age).
- [ ] Test dengan fake driver + CHR v7.
- **Exit:** Setiap mutation punya restore point `AVAILABLE` sebelum approval; gagal backup = tidak ada approval; backup lama terhapus otomatis.

### Fase 2 — Approval Backend (P0, 4–6 hari)

- [ ] Domain `approval` + state machine + test.
- [ ] Migration `approval_requests`, `command_logs`.
- [ ] Diff engine (add/set/remove/reboot/multi).
- [ ] Usecase `ProposeCommand`, `ProposeConfig`, `DecideApproval`, `CancelApproval`.
- [ ] ConnectRPC `ApprovalService` + protovalidate.
- [ ] Audit + metrics.
- [ ] Response contract lengkap untuk UI (§6.3–6.4).
- **Exit:** Approve/Reject/Edit berfungsi via API; tidak ada eksekusi sebelum approve; edit memvalidasi ulang.

### Fase 3 — Eksekusi Aman & Rollback (P0, 4–5 hari)

- [ ] `ExecuteApproved` dengan row lock + idempotency.
- [ ] Watchdog scheduler + cancel + cleanup.
- [ ] Post-verify + status `FAILED`/`rollback_recommended`.
- [ ] Restore flow (inverse script, scoped export, binary).
- [ ] `RestorePointService.RestoreFromPoint` + `DeleteRestorePoint` + typed confirmation.
- [ ] Test lockout scenario di CHR/staging.
- **Exit:** Perubahan berhasil diverifikasi; jika gagal, rollback tersedia dan teruji.

### Fase 4 — `/console/inspect` & v7 (P1, 3–4 hari)

- [ ] Driver inspect (child/syntax/completion/highlight/error).
- [ ] Parser + typed structs + test.
- [ ] Capability gate v7/v6 + cache.
- [ ] Validasi pra-approval (API) + payload untuk dynamic form.
- [ ] MCP tools inspect + `router_capabilities`.
- [ ] Warning v7 di MCP/docs/UI contract.
- **Exit:** v7 memvalidasi sintaks sebelum approval; v6 degraded tanpa crash.

### Fase 5 — Active Device (P1, 2–3 hari)

- [ ] `ActiveDeviceStore` (Redis + Postgres fallback).
- [ ] Session ID extraction dari MCP header/context.
- [ ] Resolver (explicit → session → user → single fallback).
- [ ] RPC Set/Get/Clear.
- [ ] RBAC `user_devices`.
- **Exit:** AI tanpa `device_id` menargetkan router aktif; explicit device tetap didukung; akses tidak sah ditolak.

### Fase 6 — Hardening Backend & Dokumentasi (P1, 3–4 hari)

- [ ] E2E CHR v7 (inspect, export, approval, restore, scheduled backup) + v6 degraded.
- [ ] Metrics/alerting.
- [ ] Docs + ADR + README + AGENTS.
- [ ] Security review (SSRF, path traversal, secret masking, replay).
- [ ] Load test proposal/approval.
- **Exit:** `make check`, `make test`, `make test-integration` hijau; dokumentasi backend lengkap.

### Fase 7 — UI Approval & Restore Point (P0, 4–5 hari)

- [ ] Install primitif shadcn + copy pattern AI Elements yang dipakai approval/restore (§9.2).
- [ ] Route `/ai-ops/approvals` + item sidebar **AI Router Ops** (permission `ai:ops`).
- [ ] `Queue` + `Tabs` + `Item`, filter device/status/waktu.
- [ ] Detail approval `Sheet`/`Dialog`: tab Diff, Command, Validation, Restore Point.
- [ ] Diff `CodeBlock`; editor `InputGroup` + tombol Validate; revision conflict + expiry.
- [ ] Approve/Reject (`AlertDialog` + alasan); critical action typed device name.
- [ ] `Confirmation` component siap dipakai chat (wiring di Fase 8).
- [ ] `/ai-ops/restore-points`: list/timeline/detail/download/restore + typed confirmation.
- [ ] Warning v6 `Alert`/`Badge` di approval dan restore point.
- **Exit:** Approve/Reject/Edit dan restore point berfungsi end-to-end dari UI; tidak ada tombol yang mengeksekusi command langsung.

### Fase 8 — AI Chat Backend + Chat UI (P1, 5–7 hari)

- [ ] `api/proto/v1/ai_chat.proto` + generate (`Chat` server-streaming).
- [ ] ConnectRPC `AIChatService` + auth + permission `ai:ops`.
- [ ] Stream adapter Genkit (`GenerateStream`) → `TextDelta`/`ToolCall`/`ToolResult`/`PlanStep`/`TerminalOutput`/`ApprovalRequested`/`Done`/`Error`.
- [ ] Migration `ai_conversations`/`ai_chat_events` + repository.
- [ ] Conversation + event persistence untuk reconnect (`since_event_id`); fallback non-streaming.
- [ ] Route `/ai-ops` chat (sidebar sudah ada di Fase 7), `MessageScroller`/`Message`/`Bubble`, `PromptInput`, `ModelSelector`.
- [ ] `Trace` untuk Plan/Tool/Terminal/Chain of Thought; `Marker`/`Checkpoint`; `Sources`/`Context`/`Suggestion` opsional.
- [ ] Wire `Confirmation` di chat ke `ApprovalService.DecideApproval`.
- **Exit:** Operator bisa chat dengan AI, melihat langkah tool, dan approve/reject dari chat tanpa bypass server.

### Fase 9 — Active Device UI, Polish & E2E (P1, 3–4 hari)

- [ ] `device-switcher` → `SetActiveDevice`, badge AI target, v6 badge, `Marker` saat ganti device.
- [ ] Header AI Ops: `DropdownMenu`/`Command` pilih router; tidak ada fallback `devices[0]`.
- [ ] Responsive (mobile `Sheet`), dark/light, reduced motion.
- [ ] A11y: dialog title, focus, keyboard, axe.
- [ ] Playwright: chat → propose → approve → executed; reject; edit + revision; restore typed confirm; v6 warning.
- [ ] Visual regression Approval Inbox + Chat (dark/light); `Empty`/`Skeleton`/`Spinner`/`Alert`.
- [ ] Docs UI + screenshot + update README/AGENTS.
- **Exit:** UI lolos `pnpm build`, `pnpm test`, Playwright, dan axe; tidak ada komponen chat hand-rolled.

**Total estimasi:** backend 23–32 hari + UI/chat 12–16 hari = **35–48 hari kerja** (1 developer). Bisa lebih cepat jika Fase 7–9 dikerjakan paralel oleh frontend engineer setelah kontrak API stabil.

---

## 15. Acceptance Criteria

### Backup & Restore Point
- [ ] Mutation tidak bisa diproposisikan jika restore point gagal dibuat.
- [ ] Setiap mutation membuat scoped export + inverse script + binary backup (kecuali dedup reuse artifact).
- [ ] Backup terjadwal harian/mingguan berjalan dan tidak dobel saat multi-instance.
- [ ] Dedup `ConfigHash` bekerja; backup identik tidak menumpuk.
- [ ] Setiap artifact (scoped export, binary, inverse) tersimpan, terenkripsi, punya SHA256.
- [ ] Binary backup dibuat untuk semua mutation (kecuali dedup) dan path kritis selalu wajib.
- [ ] File sementara di router dibersihkan setelah download.
- [ ] Restore point expired tidak bisa dipakai.
- [ ] Cleanup job menjaga `KEEP_LAST` + age dan tidak menghapus backup terakhir yang sukses.
- [ ] Cleanup job tidak menghapus restore point yang masih terkait approval aktif.

### Approval
- [ ] Semua verb mutation (`add/set/remove/enable/disable/reset/reboot/import/...`) → `PENDING`, tidak dieksekusi.
- [ ] Unknown verb → `MUTATION`.
- [ ] READ (`print/get/export/monitor/ping/inspect`) → eksekusi langsung + audit.
- [ ] Approve mengeksekusi tepat sekali (anti double-execute).
- [ ] Reject wajib alasan; tidak ada eksekusi.
- [ ] Edit mengubah command, memvalidasi ulang, memperbarui diff, dan menaikkan revision.
- [ ] Approval expired → `EXPIRED`, tidak bisa dieksekusi.
- [ ] AI tidak bisa approve request sendiri.
- [ ] Semua transisi tercatat di `command_logs`.
- [ ] `ApprovalService` dipakai UI approval inbox dan `Confirmation` chat; tidak ada approval lewat MCP.

### `/console/inspect`
- [ ] v7: `child`, `syntax`, `completion`, `highlight` berfungsi.
- [ ] v7: perintah invalid ditolak sebelum approval dengan pesan offset/error.
- [ ] v6: `inspect_supported=false`, `ai_mode=degraded`, tidak ada panic/error path.
- [ ] Path berbahaya (`do`, dll.) di-skip pada versi rentan.
- [ ] Timeout/rate limit melindungi router.

### Active Device
- [ ] Set active device (via UI dan API) → MCP tool tanpa `device_id` memakai device itu.
- [ ] Explicit `device_id` menang atas active binding.
- [ ] User tanpa akses device ditolak.
- [ ] Device disabled/deleted → binding invalid + error jelas.
- [ ] Tidak ada fallback acak ke device lain.

### UI (shadcn + AI Elements)
- [ ] `/ai-ops`, `/ai-ops/approvals`, dan `/ai-ops/restore-points` dapat diakses sesuai permission `ai:ops`.
- [ ] Chat merender `TextDelta`, `ToolCall`, `ToolResult`, `PlanStep`, `TerminalOutput`, dan `ApprovalRequested` dari event terstruktur.
- [ ] `Confirmation` di chat memanggil `ApprovalService.DecideApproval` dengan JWT; tidak pernah mengeksekusi command langsung.
- [ ] Approval Inbox menampilkan diff (`CodeBlock`), command editor, validation, restore point, expiry, dan revision.
- [ ] Approve/Reject/Edit menangani revision conflict, expiry, dan reject reason.
- [ ] Restore point list/detail menampilkan artifact kinds, total size, SHA256 per artifact, expiry, status; restore butuh typed confirmation.
- [ ] Device picker memanggil `SetActiveDevice`; tidak ada auto-fallback `devices[0]` untuk AI.
- [ ] v6 menampilkan `Alert`/`Badge` degraded di chat, approval, dan device.
- [ ] Tidak ada dependency/package `ai-elements`; hanya primitif shadcn + pattern yang di-copy.
- [ ] Dark/light, mobile `Sheet`, keyboard-only, dialog title, dan axe lolos.
- [ ] Tidak ada tombol approval yang dirender dari teks LLM.

### Keamanan
- [ ] `/mcp` memerlukan auth; tidak ada akses anonim.
- [ ] Audit writer tidak `nil` di production.
- [ ] Secret masking pada export/diff/log.
- [ ] Path traversal & SSRF dicegah.
- [ ] Rate limit proposal per device/user.

---

## 16. Test Plan

### Unit
- Classifier table-driven (≥50 kasus, termasuk unknown).
- Diff engine (add/set/remove/multi, selector resolution).
- Approval state machine (semua transisi + invalid transition).
- Restore point manifest/hash/path/encryption.
- Inspect parser (fixtures respons RouterOS v7).
- Active device resolver (prioritas + RBAC).
- Watchdog time computation (clock router vs server).

### Integration
- Migration up/down di Postgres.
- Repository approval/restore point/command log.
- ConnectRPC `ApprovalService` (auth, validation, conflict revision).
- ConnectRPC `RestorePointService` (auth, list/get, create manual, typed confirmation, soft delete).
- ConnectRPC `AIChatService`: auth, urutan event, `since_event_id`, reconnect, checkpoint.
- MCP tool calls via in-memory transport: propose → pending; approve → executed (fake driver).
- Restore flow dengan fake storage.
- Backup scheduler: daily/weekly, advisory lock, dedup, retention.

### E2E (CHR)
- RouterOS v7 CHR:
  - inspect child/syntax/completion/highlight.
  - scoped export + binary backup + download + hash.
  - mutation add firewall → approval → execute → verify → rollback.
  - watchdog: buat rule yang memutus API, tunggu scheduler, verifikasi rollback.
  - scheduled backup + auto-delete.
- RouterOS v6 CHR:
  - `ai_mode=degraded`, inspect tool mengembalikan error yang jelas.
  - mutation tetap bisa dengan approval + warning.

### UI
- Vitest browser: `Confirmation`, `Queue`, `Checkpoint`, diff rendering, v6 warning, `ModelSelector`.
- Contract test stream: urutan event, `since_event_id`, reconnect, fallback non-streaming, error.
- Playwright: chat → propose → approve → executed; reject dengan alasan; edit + revision conflict; restore typed confirm; ganti device.
- A11y: axe pada dialog/sheet/approval; keyboard-only; focus trap.
- Visual regression: Approval Inbox + Chat, dark/light.
- Negative: tombol approval dari output LLM tidak dirender; event `ApprovalRequested` palsu ditolak server.

### Failure injection
- Restore point creation gagal → proposal ditolak.
- Download putus di tengah → status `PARTIAL`, approval tidak boleh dibuat.
- Execute timeout → status `FAILED`/`UNKNOWN`, tidak retry otomatis.
- Approval expired saat eksekusi → ditolak.
- Dua approve bersamaan → hanya satu yang menang.

---

## 17. Risiko & Mitigasi

| Risiko | Mitigasi |
|---|---|
| AI membuat perubahan yang mengunci router | Watchdog scheduler + inverse script + binary backup **setiap mutation** (D11); typed confirmation untuk reset/restore. |
| `/console/inspect` tidak lengkap / beda versi | Gunakan sebagai validasi awal, bukan jaminan. Selalu post-verify. Cache capability per device. |
| `import` bukan rollback sempurna | Prioritaskan inverse script; scoped export sebagai fallback manual; binary backup terakhir. |
| File backup besar memenuhi flash/router | Download lalu hapus file router; batas ukuran; cleanup job; monitoring free space. |
| Binary backup setiap mutation membebani router (flash wear/lambat) | Dedup `ConfigHash`, reuse artifact jika tidak berubah, backup terjadwal, retensi `KEEP_LAST`, monitoring durasi/ukuran. |
| Backup terjadwal bertabrakan dengan mutation | Mutex per device + advisory lock; skip jika ada approval `EXECUTING`. |
| Approval fatigue | Hanya mutation yang butuh approval; multi-command digabung jadi satu approval; diff ringkas. |
| MCP stateless / session hilang | Active device di Redis dengan TTL + fallback user; explicit device_id selalu bisa. |
| AI disalahkan karena mengklaim sudah eksekusi | Tool description + status `executed:false`; UI menampilkan status approval. |
| Secret bocor ke LLM | `hide-sensitive`, `MaskSecrets`, batasi output, jangan kirim credential ke model. |
| Approval di-bypass lewat tool lama | Hapus `ExecuteCommandPreApproved` dari MCP; classifier default mutation; audit semua. |
| Watchdog salah rollback | Uji inverse script di CHR; tampilkan di UI; batasi 1 watchdog/device; manual override. |
| `Confirmation` di chat terlalu mudah diklik | Critical action pakai typed confirmation; tampilkan diff + restore point; audit setiap keputusan; Queue dengan status jelas. |
| Tombol approval dipalsukan dari output LLM | Tombol hanya dirender dari event terstruktur `ApprovalRequested`; tidak ada HTML/LLM-generated button; server tetap otoritas. |
| Stream chat terputus | Event persisted + `since_event_id`; reconnect; fallback non-streaming. |
| Pattern AI Elements di-copy lalu drift | Pin dokumentasi/commit; komponen lokal punya test; update manual terencana. |
| UI menampilkan hidden reasoning | Hanya `ReasoningDelta` terstruktur; default collapse; jangan dump raw provider payload. |
| Device target salah di UI | Server active binding + badge + `Marker` saat ganti device; explicit `device_id` menang. |

---

## 18. Yang Secara Sengaja TIDAK Dilakukan

- ❌ Klasifikasi risiko 5 tingkat (Additive/Mutating/Destructive/Critical). Cukup READ vs MUTATION.
- ❌ Menyimpan kredensial router di `.env` atau mengirimnya ke LLM.
- ❌ Mengeksekusi mutation dari `run_command` tanpa approval.
- ❌ Mengandalkan "Safe Mode" via API (tidak tersedia).
- ❌ Mengirim seluruh `/export` ke prompt LLM (token mahal + secret risk).
- ❌ Membiarkan AI approve request sendiri.
- ❌ Restore penuh otomatis tanpa konfirmasi manusia.
- ❌ Fallback ke device pertama jika active device tidak ada.
- ❌ Memakai LibreChat atau UI eksternal.
- ❌ Menambah dependency/package `ai-elements`.
- ❌ Merender tombol approve dari output teks LLM.
- ❌ Menampilkan raw hidden chain-of-thought.

---

## 19. Pertanyaan Terbuka (Perlu Keputusan)

1. **Storage restore point:**
   - Rekomendasi v1: local encrypted dir (`data/restorepoints`).
   - Alternatif: S3/MinIO (interface sudah disiapkan).

2. **Angka retensi final:**
   - Rekomendasi: `KEEP_LAST=30`, `RETENTION_DAYS=14`, backup harian 02:00, mingguan Minggu 03:00.
   - Bisa disesuaikan setelah melihat ukuran backup rata-rata.

3. **Approval via MCP elicitation?**
   - Rekomendasi: tidak untuk v1; approval hanya via ConnectRPC (user session). Bisa ditambahkan nanti dengan verifikasi human session.

4. **Lingkup conversation AI Ops:**
   - Rekomendasi: conversation terikat `device_id` saat dibuat. Ganti device membuat `Marker` dan konteks baru; history tetap tersimpan.

5. **Apakah chat boleh upload file/export?**
   - Rekomendasi: v1 tidak. Attachment hanya untuk artifact yang dihasilkan server (manifest, diff, log export).

6. **Scope Model Selector:**
   - Rekomendasi: per conversation, default dari LLM config aktif user; jangan mengubah config global dari chat.

---

## 20. Definition of Done

- [ ] READ bebas, MUTATION selalu `PENDING` + diff + restore point.
- [ ] Approve/Reject/Edit berfungsi end-to-end via UI dan API; chat `Confirmation` memakai `ApprovalService`.
- [ ] Setiap mutation memiliki binary backup (baru atau reuse dedup `ConfigHash`); backup terjadwal harian/mingguan berjalan.
- [ ] Restore point tersimpan, terenkripsi, bisa di-restore, dan dibersihkan otomatis (`KEEP_LAST` + age).
- [ ] Watchdog rollback teruji di CHR v7.
- [ ] `/console/inspect` terintegrasi di MCP + validasi pra-approval.
- [ ] v6 mutation diizinkan dengan warning degraded dan tidak crash.
- [ ] Active device binding berfungsi; tidak ada `.env` router.
- [ ] `/mcp` terautentikasi; audit tidak nil.
- [ ] Semua test + `make check` + `make test-integration` hijau.
- [ ] Dokumentasi, ADR, README, AGENTS, dan system prompt AI diperbarui.
- [ ] Chat AI Ops streaming, approval inbox, restore point UI, dan active device UI berfungsi.
- [ ] Tidak ada dependency `ai-elements`; primitif shadcn + pattern copy.
- [ ] UI lolos `pnpm build`, `pnpm test`, Playwright, dan axe (dark/light, mobile, keyboard).
- [ ] Peringatan v7 muncul di MCP instructions, tool capabilities, UI, dan docs.

---

## 21. Ringkasan Prioritas

```
P0 (wajib sebelum AI menyentuh router):
  1. Auth MCP + audit
  2. Classifier READ/MUTATION (default mutation)
  3. Backup + restore point (setiap mutation + terjadwal + auto-delete)
  4. Approval backend (approve/reject/edit + diff + expiry)
  5. Execute approved + watchdog rollback
  6. UI approval minimal (Confirmation + Inbox + diff + restore point)

P1 (operator experience):
  7. AI chat backend streaming + UI (MessageScroller, PromptInput,
     ModelSelector, Trace, Terminal, Plan, Checkpoint)
  8. /console/inspect v7 + degraded v6
  9. Active device binding + UI (badge AI target, tanpa fallback devices[0])
  10. Restore point timeline + download/restore UI
  11. Queue/Checkpoint/Sources/Context/Suggestion polish
  12. Metrics + E2E + docs + a11y + visual regression
```

**Prinsip akhir:** AI boleh membaca sebanyak yang ia mau, tetapi tidak boleh mengubah apa pun tanpa manusia melihat diff, menyetujui, dan memiliki jalan pulang (restore point).
