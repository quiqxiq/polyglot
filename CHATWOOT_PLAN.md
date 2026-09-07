# Implementasi Chatwoot Docker & Integrasi Omnichannel AI Tool Calling Polyglot

Dokumen ini merancang rencana implementasi untuk menyiapkan **Chatwoot** (menggunakan Docker di `deployments/chatwoot`) dan menghubungkannya secara dua arah dengan **Polyglot** (memanfaatkan engine WhatsApp `whatsmeow` dan AI Agent Engine dengan ISP Tool Calling).

---

## 1. Arsitektur Solusi & Alur Data

```
                         ┌──────────────────────────────────────────────┐
                         │               Kanal Pelanggan                │
                         │   • WhatsApp (via whatsmeow Polyglot)        │
                         │   • Web Live Chat (Portal Pelanggan / Web)   │
                         │   • Telegram / Email (Kanal Bawaan Chatwoot) │
                         └──────────────────────┬───────────────────────┘
                                                │
                                                ▼
┌────────────────────────────────────────────────────────────────────────────────────────┐
│                               CHATWOOT SERVER (Docker)                                 │
│  - Rails (API & Web Server) :3000                                                      │
│  - Sidekiq (Worker Antrean)                                                            │
│  - PostgreSQL & Redis (Data Chatwoot)                                                  │
│  - CS Workspace (Web Inbox, Multi-agent, Team Routing, Mobile Apps)                    │
│  - API Channel: Menampung chat WhatsApp dari Polyglot                                  │
│  - AgentBot Webhook: Menghubungi Polyglot untuk memproses AI                           │
└───────────────────────────────────────┬──────▲─────────────────────────────────────────┘
        1. Webhook (message_created)    │      │  2. REST API (Balasan Bot / Handoff)
                                        ▼      │
┌────────────────────────────────────────────────────────────────────────────────────────┐
│                                    POLYGLOT BACKEND                                    │
│                                                                                        │
│  [HTTP Adapter: /api/v1/webhooks/chatwoot]                                             │
│       │                                                                                │
│       ├── (A) Event dari Agen Manusia (message_type: outgoing)                         │
│       │        └── Teruskan ke whatsmeow (Kirim ke nomor WA pelanggan)                 │
│       │                                                                                │
│       └── (B) Event Chat Masuk (AgentBot Trigger)                                      │
│                └── Kirim ke Bot Engine (Genkit LLM)                                    │
│                         │                                                              │
│                         ├── Tool Calling ISP:                                          │
│                         │     • ping_host (Diagnostik latensi)                         │
│                         │     • get_current_time (Cek waktu & jatuh tempo)             │
│                         │     • notify_technician (Lapor teknisi lapangan)             │
│                         │     • [Baru] check_invoice_bill (Cek tagihan & bayar)        │
│                         │     • [Baru] check_subscription_status (Cek status paket)    │
│                         │     • [Baru] escalate_to_human (Alihkan ke CS manusia)       │
│                         │                                                              │
│                         └── Balas via Chatwoot REST API                                │
│                                                                                        │
│  [WhatsApp Driver: whatsmeow]                                                          │
│       └── Pesan Masuk dari WA ──> REST API Chatwoot (Push ke API Channel)              │
└────────────────────────────────────────────────────────────────────────────────────────┘
```

---

## 2. User Review Required

> [!IMPORTANT]
> **Alokasi Port & Resource Server:**
> 1. Chatwoot Rails berjalan di port internal `3000`. Kita akan mengeksposnya ke port host (misal `3000` atau `3001` jika `3000` sudah dipakai).
> 2. Chatwoot membutuhkan database PostgreSQL dan Redis. Pada konfigurasi standar Docker Chatwoot, disediakan container Postgres & Redis tersendiri agar database Chatwoot terisolasi bersih dari database operasional Polyglot NetOps (`netops`).

> [!NOTE]
> **Identitas Akun WhatsApp di Chatwoot:**
> Nomor WhatsApp pelanggan akan otomatis dibuatkan Contact di Chatwoot dengan format nomor E.164 (`+6281234567890`), sehingga agen CS dapat langsung melihat riwayat chat pelanggan tersebut kapan pun mereka menghubungi kembali.

---

## 3. Open Questions

1. **Domain/URL Akses Chatwoot:**
   Apakah untuk awal Chatwoot akan diakses via `http://localhost:3000` / IP lokal, atau Anda sudah menyiapkan subdomain khusus (misal: `cs.domainanda.com` / `chat.localhost`)?
2. **Kanal Awal di Chatwoot:**
   Apakah selain WhatsApp dari Polyglot, kita juga ingin langsung membuatkan kode embed *Web Live Chat Widget* untuk dipasang di Portal Pelanggan Polyglot?

---

## 4. Proposed Changes

### Komponen 1: Deployment Chatwoot (`deployments/chatwoot/`)

Konfigurasi Docker resmi sesuai panduan Chatwoot Production Docker Deployment:

#### [NEW] [docker-compose.yaml](file:///home/quixiq/projects/polyground/polyglot/deployments/chatwoot/docker-compose.yaml)
- Menjalankan services:
  - `base`: Image `chatwoot/chatwoot:v3.16.0` (atau latest stable).
  - `rails`: Web server aplikasi Chatwoot (port `3000:3000`).
  - `sidekiq`: Background job processor untuk pengiriman webhook, email, dan antrean event.
  - `postgres`: PostgreSQL 16 untuk Chatwoot.
  - `redis`: Redis 7 Alpine untuk background job dan cache.
- Menghubungkan network dengan Polyglot atau network internal Docker.

#### [NEW] [.env.example](file:///home/quixiq/projects/polyground/polyglot/deployments/chatwoot/.env.example) & [.env](file:///home/quixiq/projects/polyground/polyglot/deployments/chatwoot/.env)
- Konfigurasi variabel environment:
  - `SECRET_KEY_BASE` (generate via hex/rake).
  - `FRONTEND_URL` (default `http://localhost:3000`).
  - `REDIS_URL`, `POSTGRES_DATABASE`, `POSTGRES_PASSWORD`.
  - `ACTIVE_STORAGE_SERVICE=local`.
  - `DEFAULT_LOCALE=id` (Bahasa Indonesia).

#### [NEW] [setup.sh](file:///home/quixiq/projects/polyground/polyglot/deployments/chatwoot/setup.sh)
- Script automasi shell untuk:
  1. Generate `.env` dari `.env.example` beserta random `SECRET_KEY_BASE` dan password database.
  2. Menjalankan `docker compose run --rm rails bundle exec rails db:chatwoot_prepare` untuk inisialisasi schema dan seed.
  3. Memulai service (`docker compose up -d`).

---

### Komponen 2: Konfigurasi Polyglot (`internal/config/`)

#### [MODIFY] [config.go](file:///home/quixiq/projects/polyground/polyglot/internal/config/config.go)
- Menambahkan parameter integrasi Chatwoot:
  - `ChatwootEnabled` (`CHATWOOT_ENABLED`, default `false`).
  - `ChatwootBaseURL` (`CHATWOOT_BASE_URL`, contoh `http://localhost:3000`).
  - `ChatwootAPIToken` (`CHATWOOT_API_TOKEN`, User atau Bot Access Token).
  - `ChatwootAccountID` (`CHATWOOT_ACCOUNT_ID`, default `1`).
  - `ChatwootInboxID` (`CHATWOOT_INBOX_ID`, ID inbox API Channel WhatsApp).
  - `ChatwootWebhookSecret` (`CHATWOOT_WEBHOOK_SECRET`).

---

### Komponen 3: Domain & Port Chatwoot (`internal/domain/` & `internal/port/`)

#### [NEW] [chatwoot.go](file:///home/quixiq/projects/polyground/polyglot/internal/domain/chatwoot/chatwoot.go)
- Model representasi entity Chatwoot:
  - `WebhookEvent`: payload event `message_created`, `conversation_status_changed`, dsb.
  - `Contact`: data kontak (`ID`, `Name`, `PhoneNumber`, `Email`).
  - `Conversation`: data percakapan (`ID`, `InboxID`, `Status`, `ContactID`).
  - `Message`: data pesan (`Content`, `MessageType`: `incoming`/`outgoing`, `Private`).

#### [NEW] [chatwoot.go](file:///home/quixiq/projects/polyground/polyglot/internal/port/chatwoot.go)
- Interface `ChatwootClient`:
  - `GetOrCreateContact(ctx, phone, name) (*domain.Contact, error)`
  - `GetOrCreateConversation(ctx, contactID, inboxID) (*domain.Conversation, error)`
  - `SendTextMessage(ctx, conversationID, text, isPrivate) (*domain.Message, error)`
  - `ToggleConversationStatus(ctx, conversationID, status) error`

---

### Komponen 4: Adapter HTTP & Client Chatwoot (`internal/adapter/`)

#### [NEW] [client.go](file:///home/quixiq/projects/polyground/polyglot/internal/adapter/chatwoot/client.go)
- Implementasi `port.ChatwootClient` via HTTP REST ke API Chatwoot v1.

#### [NEW] [handler.go](file:///home/quixiq/projects/polyground/polyglot/internal/adapter/http/chatwoot/handler.go)
- Handler HTTP untuk endpoint: `POST /api/v1/webhooks/chatwoot`
- Menangani 2 skenario event:
  1. **Balasan Agen Manusia (`message_type: outgoing` dan bukan dari bot):**
     Meneruskan pesan langsung ke WhatsApp pelanggan via `waGateway.SendMessage(...)`.
  2. **Pesan Masuk untuk AI AgentBot:**
     Meneruskan pesan ke `BotEngine` untuk diproses LLM + Tool Calling, lalu mengirim balasan AI kembali ke Chatwoot.

---

### Komponen 5: Integrasi WhatsApp Driver & AI Tool Calling (`internal/usecase/`)

#### [NEW] [bridge.go](file:///home/quixiq/projects/polyground/polyglot/internal/usecase/chatwoot/bridge.go)
- Sinkronisasi dua arah:
  - Menerima event pesan WhatsApp masuk dari `whatsmeow` -> auto-create contact & conversation di Chatwoot -> push pesan masuk ke Chatwoot inbox.

#### [MODIFY] [tools.go](file:///home/quixiq/projects/polyground/polyglot/internal/usecase/bot/tools.go)
- Menambahkan tool calling baru untuk AI:
  - `NewCheckCustomerInvoiceTool(invRepo, customerRepo)`: AI dapat mengecek status tagihan belum dibayar pelanggan secara real-time.
  - `NewEscalateToHumanTool(chatwootClient)`: AI dapat mengubah status percakapan menjadi `open` di Chatwoot dan menandai tiket untuk diambil alih staf CS manusia.

#### [MODIFY] [app.go](file:///home/quixiq/projects/polyground/polyglot/internal/app/app.go) & [router.go](file:///home/quixiq/projects/polyground/polyglot/internal/app/router.go)
- Inisialisasi `ChatwootClient` dan webhook handler di `app.go`.
- Registrasi route `POST /api/v1/webhooks/chatwoot` di public mux `router.go`.
- Mendaftarkan forwarder WhatsApp ke Chatwoot pada callback `waManager`.

---

## 5. Verification Plan

### Automated Tests
- Unit test adapter Chatwoot Client: `go test -v ./internal/adapter/chatwoot/...`
- Unit test Chatwoot Webhook Handler: `go test -v ./internal/adapter/http/chatwoot/...`
- Unit test Tool Calling baru (Invoice & Escalation): `go test -v ./internal/usecase/bot/...`
- Validasi build dan layer boundaries:
  ```bash
  make check-layer-boundaries
  make vet
  make test
  ```

### Manual Verification
1. **Verifikasi Docker Chatwoot:**
   - Jalankan `setup.sh` di `deployments/chatwoot`.
   - Buka `http://localhost:3000`, pastikan halaman onboarding superadmin Chatwoot terbuka.
   - Buat Akun & Inbox tipe **API Channel** di Chatwoot, dapatkan `inbox_id` dan `api_token`.
2. **Verifikasi End-to-End Chat:**
   - Kirim chat dari HP ke nomor WhatsApp Polyglot (`whatsmeow`).
   - Pastikan pesan muncul secara real-time di inbox Chatwoot.
   - Balas pesan dari Chatwoot sebagai Agen CS -> pastikan balasan diterima di HP WhatsApp.
   - Uji AI Bot: kirim pesan tanya tagihan / minta cek koneksi -> pastikan AI membalas dengan akurat memanfaatkan *tool calling* (ping / invoice / lapor teknisi).
   - Uji Eskalasi: kirim pesan "saya mau bicara dengan manusia" -> pastikan bot memanggil `escalate_to_human` dan status di Chatwoot berubah menjadi `Open`.
