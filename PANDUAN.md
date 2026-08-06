# Roadmap Strategis: NetOps & Hotspot Engine MVP (Polyglot)

Dokumen ini adalah **Roadmap Resmi & Panduan Eksekusi** untuk pengembangan fondasi utama **Polyglot NetOps Engine**. Fokus utama tahap ini adalah menghadirkan monitoring jaringan riil, manajemen Hotspot ala Mikhmon, fondasi Multi-Tenant, pengiriman file via WhatsApp, dan eksekusi tool MikroTik via Chatbot Admin.

---

## 🏗️ Ringkasan Pilar Arsitektur & Teknologi

* **Backend**: Go (Clean Architecture: `domain`, `usecase`, `port`, `adapter`, `driver`)
* **Frontend**: Vue 3 + Vite + Pinia + Tailwind / Vanilla CSS (Dark/Light Theme)
* **Realtime**: WebSocket & Server-Sent Events (SSE)
* **Keamanan & Access Control**: Casbin RBAC + Multi-Tenant (`tenant_id`)
* **Database**: PostgreSQL (GORM) + Redis (Session & State)
* **WA Gateway**: Whatsmeow (Multi-Session + Media Sender)

---

## 📌 Milestone Roadmap & Urutan Eksekusi

```mermaid
timeline
    title Peta Jalan Pengembangan Polyglot NetOps MVP
    section Tahap 1 : Multi-Tenancy & Active Sessions
        Database Tenant Isolation : Skema tenant_id pada Device, WASession, Voucher
        Monitoring Engine : PPPoE Active, Hotspot Active, DHCP Leases API
    section Tahap 2 : Mikhmon Engine & PDF Generator
        Hotspot Profiles & Users : CRUD User & Profile RouterOS
        Voucher PDF Engine : Generator Voucher Massal & Layout PDF/PNG
    section Tahap 3 : WA Gateway File Sender
        WhatsApp Media Adapter : Support SendDocument (PDF) & SendImage (QR/Graphs)
        Media Storage Handler : Temp PDF generator & cleanup worker
    section Tahap 4 : Admin Chatbot Tool Execution
        LLM Tool Calling : Function Calling untuk Cek Mikrotik & Buat Voucher
        Admin Auth & HITL : Verifikasi Admin via WA & 2-Step Confirmation
    section Tahap 5 : UI Dashboard & Realtime Polish
        Live Monitoring Cards : Kartu status realtime PPPoE/Hotspot/DHCP
        Voucher Print & Share : Preview & Tombol Kirim PDF WA langsung dari UI
```

---

## 📋 Rincian Implementasi per Tahap

### TAHAP 1: Multi-Tenancy Foundation & Active Sessions Monitoring
> **Goal**: Mengisolasi data per-ISP (`tenant_id`) dan menghadirkan API & Driver monitoring session aktif MikroTik secara realtime.

#### 1.1 Multi-Tenant Schema & Isolation
- [ ] Tambahkan `tenant_id` pada GORM models: `Device`, `WASession`, `HotspotUser`, `HotspotVoucher`, `KnowledgeEntry`.
- [ ] Perbarui Casbin RBAC middleware `internal/adapter/http/middleware/rbac.go` untuk memvalidasi `tenant_id` pada setiap request API.

#### 1.2 Active Sessions Driver & Usecase (`internal/driver/mikrotik/`)
- [ ] **PPPoE Active Sessions**: Impl `/ppp/active/print` (Username, IP, Caller-ID, Uptime, Bytes).
- [ ] **Hotspot Active Users**: Impl `/ip/hotspot/active/print` (User, IP, MAC, Uptime, Bytes In/Out).
- [ ] **DHCP Leases**: Impl `/ip/dhcp-server/lease/print` (IP Address, MAC, Hostname, Status, Expires).
- [ ] Buat usecase `internal/usecase/network/get_active_sessions.go`.
- [ ] Daftarkan REST API di `internal/adapter/http/mikhmon_handler.go`.

---

### TAHAP 2: Mikhmon Hotspot Engine & Voucher PDF Generator
> **Goal**: Memungkinkan pembuatan voucher massal di MikroTik RouterOS dan mencetaknya dalam bentuk PDF/PNG.

#### 2.1 Hotspot User & Profile Management
- [ ] API & Usecase CRUD User Profile Hotspot (`/ip/hotspot/user/profile`).
- [ ] API & Usecase CRUD User Hotspot (`/ip/hotspot/user`).

#### 2.2 Voucher Generation Engine (`internal/usecase/business/manage_voucher.go`)
- [ ] Generator string unik kode voucher (Prefix, Panjan, Karakter Acak).
- [ ] Batch Insert User ke MikroTik RouterOS secara paralel/pipeline.
- [ ] Standar statistik penjualan voucher (Income report harian/bulanan).

#### 2.3 Voucher Layout PDF/PNG Generator (`pkg/voucher/`)
- [ ] Generator Layout PDF Voucher siap cetak (kisi-kisi 3x10 per lembar A4, QR Code per voucher, Logo ISP, Harga, Durasi).
- [ ] Render PDF ke `[]byte` buffer untuk siap diunduh atau dikirim via WhatsApp.

---

### TAHAP 3: WhatsApp Gateway Media & Document Sender
> **Goal**: Meng-upgrade WhatsApp Gateway agar bisa mengirim file PDF, Gambar, dan Laporan.

#### 3.1 Kontrak Port & Driver WhatsApp (`internal/port/whatsapp_gateway.go`)
- [ ] Tambahkan method `SendDocument(sessionID uint, to string, fileBytes []byte, fileName string, caption string) error`.
- [ ] Tambahkan method `SendImage(sessionID uint, to string, imageBytes []byte, caption string) error`.
- [ ] Implementasikan di `internal/driver/whatsapp/` menggunakan API Whatsmeow Upload/Send.

#### 3.2 Media Cleaning Worker
- [ ] Buat worker pembersih file temporary PDF/PNG di `internal/usecase/bot/media_cleaner.go`.

---

### TAHAP 4: Admin Chatbot Tool Execution & HITL Security
> **Goal**: Admin / Super Admin dapat mengontrol MikroTik dan mencetak voucher langsung via obrolan WhatsApp / WebChat.

#### 4.1 Identifikasi User Admin via WA
- [ ] Query pemetaan `customerNumber` -> Tabel `users` (Role: Admin / SuperAdmin).
- [ ] Percabangan di `Engine.HandleIncomingMessage`: Pelanggan biasa (FAQ RAG) vs Admin (Tool Execution Engine).

#### 4.2 LLM Tool Calling Integration (`internal/adapter/llm/`)
- [ ] Integrasikan Function Calling ke Provider LLM (OpenAI/Gemini/Claude):
  - Tool 1: `check_active_sessions(type)` (PPPoE / Hotspot / DHCP)
  - Tool 2: `generate_vouchers(profile, qty, format)`
  - Tool 3: `get_device_health(device_name)`
  - Tool 4: `reboot_router(device_name)`

#### 4.3 Human-In-The-Loop (HITL) untuk Perintah Destruktif
- [ ] Integrasi `command.ClassDestructive` dengan Redis Token.
- [ ] Alur 2-Step Confirmation via WA untuk tindakan berbahaya (Reboot/Reset Config).

---

### TAHAP 5: Modern UI Dashboard & Live Realtime Updates
> **Goal**: Menampilkan monitoring session, manajemen voucher, dan RBAC secara visual di Dashboard Web Vue 3.

#### 5.1 Active Sessions & Network Monitoring View (`web/src/views/ActiveSessionsView.vue`)
- [ ] Tab **PPPoE Active Sessions**: Tabel realtime, filter, tombol kick user (`DELETE /api/v1/mikrotik/ppp/active/:id`).
- [ ] Tab **Hotspot Active Sessions**: Tabel realtime (MAC, IP, Uptime, Bytes), tombol kick user (`DELETE /api/v1/mikrotik/hotspot/active/:id`).
- [ ] Tab **DHCP Leases**: Tabel leases (IP, MAC Address, Server, Dynamic/Static Status).

#### 5.2 Hotspot Voucher Engine & Printing View (`web/src/views/VouchersView.vue`)
- [ ] Tab **Voucher Generator**: Form pembuatan voucher massal (Profile, Qty, Mode Server, Prefix, Character Set).
- [ ] Modal **Preview & Cetak Voucher**: Render HTML layout voucher (A4 grid 3x10 / Thermal receipt) lengkap dengan QR Code base64.
- [ ] Integrasi **Kirim Voucher ke WA Admin**: Tombol instant untuk mengirim voucher via WhatsApp Gateway (`POST /api/v1/sessions/:id/send-document`).
- [ ] Tab **Laporan Penjualan (Income Report)**: Filter laporan harian, bulanan, dan tahunan dari backend.

#### 5.3 User & RBAC Management View (`web/src/views/RbacManagementView.vue`)
- [ ] Tab **User Management**: Tabel Pengguna, Form Tambah/Edit User, Reset Password per Tenant.
- [ ] Tab **RBAC Policy & Role Assignment**: Matriks Centang Permission per Role, Form Penugasan Role ke User via Casbin API.

#### 5.4 Realtime SSE & Pinia Integration
- [ ] `stores/network.ts`, `stores/mikhmon.ts`, dan `stores/rbac.ts` untuk manajemen state.
- [ ] Integration `stores/realtime.ts` (SSE Event `active_sessions_update` & status update).

---

## 🎯 Rencana Verifikasi & Testing (Quality Assurance)

1. **Unit Testing**:
   - Test generator voucher di `pkg/voucher/`
   - Test klasifikasi perintah MikroTik di `internal/driver/mikrotik/commands_test.go`
2. **Integration Testing**:
   - Simulated RouterOS API test untuk `/ppp/active` dan `/ip/hotspot/user`
   - Real WhatsApp Media Upload test
3. **Build & Lint Verification**:
   - `go test ./...`
   - `golangci-lint run`
   - `cd web && npm run build`

---

> **Dokumen ini siap dijadikan acuan utama dalam perencanaan dan implementasi langkah demi langkah.**
