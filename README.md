# Polyglot — NetOps & ISP Management Engine

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![ConnectRPC](https://img.shields.io/badge/ConnectRPC-v1.18-2B3252?style=flat)](https://connectrpc.com/)
[![React](https://img.shields.io/badge/React-19.x-61DAFB?style=flat&logo=react)](https://react.dev/)
[![TailwindCSS](https://img.shields.io/badge/TailwindCSS-v4-38B2AC?style=flat&logo=tailwind-css)](https://tailwindcss.com/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

**Polyglot** adalah platform engine backend NetOps dan manajemen ISP modern yang dibangun dengan Go murni (`net/http.ServeMux` Go 1.22+). Engine ini mengekspos **ConnectRPC**, **Model Context Protocol (MCP)**, dan **Server-Sent Events (SSE) / WebSocket** untuk otomasi jaringan multi-vendor, billing ISP, hotspot voucher (Mikhmon parity), dan customer service bot AI terintegrasi WhatsApp.

---

## 📑 Daftar Isi Dokumen (Documentation Index)

Berikut adalah daftar dokumen acuan teknis, arsitektur, dan pedoman pengembangan di proyek ini:

### 🌟 Dokumen Utama & Pedoman Arsitektur
- **[DEVELOPMENT-GUIDELINES.md](DEVELOPMENT-GUIDELINES.md)** — **Panduan & Standar Pengembangan Definitif Proyek** (Arsitektur Clean Hexagonal, Standar Naming, Go Interfaces, Structs, Logging Terpusat, Error Handling, Layer Boundaries, dan Checklist).
- **[ROADMAP-AND-ISSUES.md](ROADMAP-AND-ISSUES.md)** — **Peta Jalan Pengembangan (Roadmap) & Pelacakan Isu Kritis (Known Issues)** seperti normalisasi nomor WhatsApp/whitelist, transformer format AI markdown-to-WhatsApp, visual template designer, dan integrasi OLT.
- **[SYSTEM-STRUCTURE-AND-ARCHITECTURE.md](SYSTEM-STRUCTURE-AND-ARCHITECTURE.md)** — Dokumentasi struktur folder definitif, relasi antar-komponen, dan diagram arsitektur sistem.
- **[Polyglot-Architecture.md](Polyglot-Architecture.md)** — Alur kerja arsitektur sistem, konsep vendor-agnostic command gateway, dan state orchestration.
- **[TECH-STACK-DAN-PERSIAPAN.md](TECH-STACK-DAN-PERSIAPAN.md)** — Pemilihan stack teknologi, dependensi, dan versi library resmi.
- **[AGENTS.md](AGENTS.md)** / **[CLAUDE.md](CLAUDE.md)** — Instruksi operasional dan aturan coding standar untuk AI agent/developer.

### 📚 Spesifikasi Fitur & Basis Data
- **[docs/database-schema.md](docs/database-schema.md)** — **Skema Basis Data ISP Management Lengkap** (Pelanggan, Subscriptions PPPoE/Hotspot, Plans, Invoices, Transaksi Pembayaran, Kas, dan Laporan).
- **[docs/mikhmon/README.md](docs/mikhmon/README.md)** — **Spesifikasi Modul Hotspot Mikhmon Parity (01–11)**:
  - [01-auth-admin.md](docs/mikhmon/01-auth-admin.md) — Multi-Router Auth & Session
  - [02-router-instance.md](docs/mikhmon/02-router-instance.md) — Manajemen Router MikroTik
  - [03-dashboard-monitoring.md](docs/mikhmon/03-dashboard-monitoring.md) — Dashboard & Streaming Realtime
  - [04-hotspot-user.md](docs/mikhmon/04-hotspot-user.md) — Hotspot User CRUD & Bulk Cleaner
  - [05-hotspot-profile.md](docs/mikhmon/05-hotspot-profile.md) — User Profile & Expire Mode
  - [06-active-host-server.md](docs/mikhmon/06-active-host-server.md) — Active Sessions, Hosts & IP Bindings
  - [07-voucher-generator.md](docs/mikhmon/07-voucher-generator.md) — Batch Voucher Generator & Prefix Rules
  - [08-sales-report.md](docs/mikhmon/08-sales-report.md) — Rekap Penjualan & Pembukuan Kas
  - [09-expire-monitor.md](docs/mikhmon/09-expire-monitor.md) — Expire Monitor Daemon & Script
  - [10-template-print.md](docs/mikhmon/10-template-print.md) — Layout Template Cetak & Thermal
  - [11-resources-theme.md](docs/mikhmon/11-resources-theme.md) — Resource Monitor & Tema
  - [IMPLEMENTATION-PLAN-04-10.md](docs/mikhmon/IMPLEMENTATION-PLAN-04-10.md) — Detail Teknis Implementasi Hotspot

### 🏛️ ADR (Architectural Decision Records)
- [0001: Pilih Gin daripada Echo](docs/adr/0001-pilih-gin-daripada-echo.md) *(Superseded oleh ADR 0005)*
- [0002: DeviceDriver Tanpa Session Terpisah](docs/adr/0002-devicedriver-tanpa-session-terpisah.md) — Deviasi terkontrol tanpa session terpisah
- [0003: MikroTik Dual-Connection Streaming](docs/adr/0003-mikrotik-dual-connection-streaming.md) — Pemisahan koneksi exec vs streaming
- [0004: Generic CLI Driver Scrapligo](docs/adr/0004-generic-cli-driver-scrapligo.md) — Mesin CLI terpadu via Scrapligo (SSH/Telnet)
- [0005: Migrasi Total ke net/http.ServeMux](docs/adr/0005-migrasi-dari-gin-ke-net-http-servemux.md) — Penggantian Gin penuh dengan HTTP router standar Go 1.22+

---

## 🚀 Fitur Utama

```
                      ┌─────────────────────────────────────────┐
                      │          Polyglot Web Frontend          │
                      │       (React 19 + Vite + Tailwind)      │
                      └────────────────────┬────────────────────┘
                                           │ ConnectRPC / SSE
                                           ▼
┌────────────────────────────────────────────────────────────────────────────────┐
│                          Polyglot NetOps Core Engine                           │
│  ┌───────────────────────┬───────────────────────┬──────────────────────────┐  │
│  │   ConnectRPC Layer    │       MCP Server      │     SSE / WebSocket      │  │
│  │ (Auth, Users, Devices,│ (AI Agent Tools: Ping,│ (Realtime Live Traffic,  │  │
│  │  Hotspot, PPP, Bot)   │ Status, Push Config)  │  Resource, Active Users) │  │
│  └───────────┬───────────┴───────────┬───────────┴────────────┬─────────────┘  │
│              │                       │                        │                │
│              ▼                       ▼                        ▼                │
│  ┌──────────────────────────────────────────────────────────────────────────┐  │
│  │                       Clean Architecture Use Cases                       │  │
│  │  • Hotspot & Voucher  • PPP / PPPoE Secrets  • AI WhatsApp Customer Care │  │
│  │  • Billing & Invoices • Device Management    • Knowledge Base (RAG)      │  │
│  └───────────────────────────────────┬──────────────────────────────────────┘  │
│                                      │                                         │
│                                      ▼                                         │
│  ┌──────────────────────────────────────────────────────────────────────────┐  │
│  │                     Hardware & External Drivers Layer                    │  │
│  │  • MikroTik RouterOS  • Cisco IOS (CLI)      • Generic SSH/Telnet        │  │
│  │  • WhatsApp (whatsmeow)• ZTE / Huawei OLT    • GenieACS (TR-069)         │  │
│  └──────────────────────────────────────────────────────────────────────────┘  │
└────────────────────────────────────────────────────────────────────────────────┘
```

1. **Hotspot & Voucher Engine (Mikhmon Parity)**:
   - Dashboard live traffic & CPU/RAM streaming.
   - Hotspot User CRUD & Pembersih Massal (Bulk Delete by Profile, by Comment/Batch, dan Expired Users).
   - Hotspot IP Bindings (`/ip/hotspot/ip-binding`) dengan fitur 1-klik Bypass langsung dari tabel Host.
   - Hotspot Cookies (`/ip/hotspot/cookie`) session management.
   - Quick Voucher Status Checker & live inspector (Uptime, Data, MAC Lock, Status Online).
   - Multi-layout voucher generator & direct printing (Standard, Small, Thermal roll).
2. **PPPoE & Broadband Service**:
   - PPPoE Secrets & Profiles management.
   - Pemantauan sesi aktif secara realtime (Stream & Kick active PPPoE sessions).
   - Inactive secrets detector untuk mendeteksi pelanggan offline/isolir.
3. **WhatsApp Bot & AI Customer Service Agent**:
   - Integrasi WhatsApp multi-session via `whatsmeow` dengan pairing QR code.
   - LLM Provider fleksibel (OpenAI, Gemini, Ollama, DeepSeek).
   - Dynamic Tools: Ping Host, Get Current Time, dan Notifikasi Otomatis ke WhatsApp Teknisi Lapangan.
   - Multi-tier rate limiting dan anti-spam guardrails.
4. **Multi-Vendor Network Automation**:
   - MikroTik (RouterOS API v6/v7).
   - Cisco IOS / Scrapligo generic SSH & Telnet engine.
   - ZTE OLT & Huawei OLT optical diagnostics.
   - GenieACS TR-069 CPE provisioning.

---

## 🛠️ Menjalankan Aplikasi

### 1. Prasyarat
- **Go**: 1.22 atau lebih baru
- **Node.js**: 20+ & npm / pnpm
- **PostgreSQL**: 15+
- **Protoc & Plugins**: `protoc-gen-go`, `protoc-gen-connect-go`, `protoc-gen-es`, `protoc-gen-connect-es` (opsional jika mengedit `.proto`)

### 2. Konfigurasi Lingkungan (.env)
Salin berkas konfigurasi template:
```bash
cp .env.example .env
cp web/.env.example web/.env
```
Sesuaikan kredensial database PostgreSQL, JWT secret, dan port server pada berkas `.env`.

### 3. Menjalankan Backend
```bash
# Build binary server
make build

# Jalankan server backend (default port: 8080)
make run
```

### 4. Menjalankan Frontend Web UI
```bash
cd web
npm install
npm run dev
```
Akses web dashboard di `http://localhost:5173`.

---

## 🧪 Pengujian & Quality Assurance

```bash
# Menjalankan seluruh Unit Tests
make test

# Menjalankan Linter Go
make lint

# Menjalankan Test Integrasi ke Perangkat MikroTik Fisik
MIKROTIK_TEST_HOST=192.168.88.1 \
MIKROTIK_TEST_USER=admin \
MIKROTIK_TEST_PASS=secret \
make test-integration
```

---

## 📦 Perintah Makefile yang Sering Digunakan

| Perintah | Deskripsi |
|---|---|
| `make build` | Mengompilasi binary server ke `bin/server` |
| `make run` | Menjalankan server Go secara langsung |
| `make proto` | Meng-generate protobuf stubs untuk Go (`api/gen/v1/`) |
| `make proto-web` | Meng-generate protobuf stubs untuk Web TypeScript (`web/src/gen/v1/`) |
| `make test` | Menjalankan seluruh test suite Go |
| `make lint` | Menjalankan static analysis dan linter Go |
| `make migrate-up` | Menjalankan database migrations ke PostgreSQL |
| `make migrate-down` | Me-rollback database migrations |

---

*Dikembangkan dengan standar Clean Hexagonal Architecture untuk keandalan dan skalabilitas operasional ISP.*
