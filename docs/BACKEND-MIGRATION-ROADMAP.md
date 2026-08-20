# 🗺️ Roadmap & Analisis Mendalam Migrasi Backend Go (NetOps Engine)

Dokumen ini adalah peta jalan definitif dan status migrasi komprehensif untuk memindahkan seluruh komunikasi backend Go ke arsitektur transport target.

---

## 📊 Matriks Transport & Status Komunikasi Backend

| Jalur Komunikasi | Transport Target | Alasan Engineering | Status Backend Go |
| :--- | :--- | :--- | :--- |
| **1. Web UI Frontend ↔ Go Backend** | 🚀 **ConnectRPC** | Type-safe, otomatis auto-complete di Vue TS, performa streaming HTTP/2 tinggi. | 🟢 **100% Selesai** (Backend Go ConnectRPC Handlers **Done**: `Device`, `Customer`, `Mikhmon`, `Bot`, `Knowledge`, `WhatsApp`, `Technician`, `LLMConfig`, `Auth`, `RBAC`, `Billing`) |
| **2. AI Agent / MCP Tool ↔ Go Backend** | 🤖 **MCP (JSON-RPC)** | Memanggil Usecase teruji yang sama dengan ConnectRPC. | 🟢 **100% Selesai** (Core `/mcp` Engine & 8 Tools **Done**: `get_device_status`, `run_command`, `push_config`, `mikhmon_get_dashboard`, `mikhmon_generate_voucher`, `mikhmon_kick_session`, `customer_lookup`, `search_knowledge`) |
| **3. Inter-Service / POP Remote Probe** | 📡 **ConnectRPC Binary** | Kompresi binary Protobuf hemat bandwidth antar-POP ISP. | 🟢 **100% Selesai** (Protobuf Schema `probe.proto` & Executable `cmd/probe/` **Done**) |
| **4. Webhook Pihak Ke-3 (WA, Midtrans)** | 🌐 **REST API** | Layanan eksternal publik biasanya hanya mendukung HTTP POST JSON standar. | 🟢 **100% Selesai** (Midtrans, WA Webhook, SSE Stream) |

---

## 🔍 Detail Analisis Mendalam Komponen Sisa

### 🔴 1. Sisa Jalur 1: Web UI Frontend ↔ Backend (ConnectRPC)
Modul backend REST API internal yang belum dimigrasikan ke ConnectRPC:

1. **`AuthService`** (`internal/adapter/http/auth_handler.go`):
   - `Login(username, password)` ➡️ Mengembalikan JWT Token & info user.
   - `GetMe()` ➡️ Profile user login.
   - `RefreshToken()` ➡️ Rotasi token JWT.
2. **`RBACService`** (`internal/adapter/http/rbac_handler.go`):
   - `ListPolicies()`, `AddPolicy()`, `RemovePolicy()`.
   - `ListRoleAssignments()`, `AssignRole()`, `UnassignRole()`.
3. **`BillingService`** (`internal/adapter/http/invoice_handler.go` & `subscription_handler.go`):
   - `ListInvoices()`, `GetInvoice()`, `CreateInvoice()`, `PayInvoice()`.
   - `ListSubscriptions()`, `CreateSubscription()`, `CancelSubscription()`.

---

### 🔴 2. Sisa Jalur 2: Ekspansi Tool MCP untuk AI Agent / GNET Bot
Core engine MCP `/mcp` di `internal/adapter/mcp/` sudah aktif. Tool tambahan yang harus dibuat:

1. **Hotspot & Mikhmon Tools**:
   - `tool_mikhmon_get_dashboard.go`: AI membaca statistik CPU, Uptime, & Income Router.
   - `tool_mikhmon_generate_voucher.go`: Mass generate voucher hotspot via percakapan AI Chatbot.
   - `tool_mikhmon_kick_session.go`: Mengeluarkan user hotspot bermasalah.
2. **Customer & Support Tools**:
   - `tool_customer_lookup.go`: Mencari data langganan customer berdasarkan nomor HP/WA.
   - `tool_search_knowledge.go`: Pencarian artikel solusi RAG untuk menjawab pertanyaan teknis pelanggan.

---

### 🔴 3. Sisa Jalur 3: POP Remote Probe Executable Agent (`cmd/probe/`)
Membuat program executable ringan untuk dipasang di POP ISP cabang:

1. **Binary Agent Lightweight (`cmd/probe/main.go`)**:
   - Program Go kecil (cross-compiled untuk Mikrotik x86, OpenWrt, Debian, ARM).
   - Menjalankan loop background ICMP/Ping telemetry & SNMP poller.
2. **ConnectRPC Binary Telemetry**:
   - Membuka stream bi-directional ConnectRPC HTTP/2 ke Server Pusat.
   - Mengirim kompresi binary Protobuf secara real-time (menghemat >80% bandwidth).

---

## 🗺️ Peta Jalan Eksekusi (Executive Roadmap)

```
[Tahap A.2: Auth, RBAC & Billing ConnectRPC]
                 │
                 ▼
[Tahap B: Ekspansi MCP Tools (Mikhmon & WA)]
                 │
                 ▼
[Tahap C: Executable POP Remote Probe (cmd/probe/)]
```

### 1. **Tahap A.2 (Auth, RBAC & Billing ConnectRPC)**
- Membuat Protobuf schemas `auth.proto`, `rbac.proto`, `billing.proto`.
- Membuat ConnectRPC handlers `auth_handler.go`, `rbac_handler.go`, `billing_handler.go` di `internal/adapter/connect/`.
- Mendaftarkan route di `cmd/server/main.go`.

### 2. **Tahap B (Ekspansi Tools MCP)**
- Membuat file-file tool MCP di `internal/adapter/mcp/`.
- Menghubungkan MCP tools ke `MikhmonUseCase`, `CustomerUseCase`, dan `KnowledgeRetriever`.

### 3. **Tahap C (POP Remote Probe Executable)**
- Membuat folder & file `cmd/probe/main.go`.
- Mengimplementasikan ConnectRPC client runner over HTTP/2 Protobuf binary.
